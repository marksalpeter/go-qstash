package qstash

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"testing"
	"time"

	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

func TestRoundTrip(t *testing.T) {
	ctx, cancelNotify := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	ctx, cancelTimeout := context.WithTimeout(ctx, time.Second*60)
	cancel := func() {
		cancelNotify()
		cancelTimeout()
	}
	defer cancel()

	send := Message{
		Body: []byte("message"),
	}

	// Check that the received message matches the one we sent
	topicURL, done, err := testReceive(t, ctx, func(_ context.Context, m *Message) {
		if m.ID != send.ID {
			t.Errorf("expected message id %s, got %s", send.ID, m.ID)
		} else if string(m.Body) != string(send.Body) {
			t.Errorf("expected message body '%s', got '%s'", string(send.Body), string(m.Body))
		}
		m.Ack()
		cancel()
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("listening on ", topicURL)

	// Publish the message to the topic
	t.Log("publish message...")
	if err := testPublish(ctx, topicURL, &send); err != nil {
		t.Fatal(err)
	}

	t.Log("waiting to receive message...")
	<-done

}

// testPublish publishes a message to the topic url
func testPublish(ctx context.Context, topicURL string, m *Message, opts ...PublishOption) error {
	p, err := NewPublisher(topicURL)
	if err != nil {
		return err
	}
	return p.Publish(ctx, m, opts...)
}

// testReceive uses ngrok to connect a public reverse proxy to the receiver
func testReceive(t *testing.T, ctx context.Context, onReceive func(ctx context.Context, m *Message)) (string, <-chan struct{}, error) {
	// Create a receiver
	r, err := NewReceiver()
	if err != nil {
		return "", nil, err
	}
	// Create a public reverse proxy
	tun, err := ngrok.Listen(
		ctx,
		config.HTTPEndpoint(),
		ngrok.WithAuthtokenFromEnv(),
	)
	if err != nil {
		return "", nil, err
	}

	// This will stop when the context is canceled
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := http.Serve(tun, r.Receive(onReceive)); err != nil {
			t.Log(err)
		}
	}()
	return tun.URL(), done, nil
}
