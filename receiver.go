package qstash

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)

// Receiver generates [http.Handler]s that receive and verify qstash messages from a lambda function
type Receiver struct {
	signingKey     string
	nextSigningKey string
}

// NewReceiver returns a new QStash Receiver
func NewReceiver(opts ...ReceiverOption) (*Receiver, error) {
	// Apply the options
	var os ReceiverOptions
	if err := os.apply(opts...); err != nil {
		return nil, fmt.Errorf("receiver is missing config: %w", err)
	}
	return &Receiver{
		signingKey:     os.SigningKey,
		nextSigningKey: os.NextSigningKey,
	}, nil
}

// Receive receives a message from the QStash
// Note: you must call ack or nack on the message for the request to complete
func (q *Receiver) Receive(onReceive func(ctx context.Context, m *Message)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the body
		body, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Verify the signature
		tokenString := r.Header.Get("Upstash-Signature")
		if err := q.verify(body, tokenString, q.signingKey); err != nil {
			// Try the next signing key
			if err := q.verify(body, tokenString, q.nextSigningKey); err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		// Parse the message
		var m Message
		m.ID = r.Header.Get("Upstash-Message-Id")
		m.Headers = r.Header
		m.Body = body
		m.Retried, _ = strconv.Atoi(r.Header.Get("Upstash-Retried"))
		m.w = w
		// Call the receiver
		if onReceive != nil {
			onReceive(r.Context(), &m)
		}
		// Retry unacknowledged messages
		if !m.isAcknowledged {
			http.Error(w, "message was not acknowledged by the receiver", http.StatusUnprocessableEntity)
			return
		}
	})
}

// verify verifies the body of a signed qstash request
func (q *Receiver) verify(body []byte, tokenString, signingKey string) error {
	// Parse the JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(signingKey), nil
	})
	if err != nil {
		return fmt.Errorf("could not parse jwt: %w", err)
	}
	// Validate the claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return fmt.Errorf("could not jwt process token claims")
	} else if !claims.VerifyIssuer("Upstash", true) {
		return fmt.Errorf("invalid issuer")
	} else if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return fmt.Errorf("token has expired")
	} else if !claims.VerifyNotBefore(time.Now().Unix(), true) {
		return fmt.Errorf("token is not valid yet")
	}
	bodyHash := sha256.Sum256(body)
	if claims["body"] != base64.URLEncoding.EncodeToString(bodyHash[:]) {
		return fmt.Errorf("body hash does not match")
	}
	return nil
}
