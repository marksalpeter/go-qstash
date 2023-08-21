package qstash

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Publisher for the qstash queue
type Publisher struct {
	token  string
	url    string
	topic  string
	client interface {
		Do(*http.Request) (*http.Response, error)
	}
	uuid interface {
		NewV4() (string, error)
	}
	verbose bool
}

// NewPublisher creates a new qstash publisher
func NewPublisher(topic string, opts ...PublisherOption) (*Publisher, error) {
	// Apply the options
	var os PublisherOptions
	if err := os.apply(append(opts, withTopic(topic))...); err != nil {
		return nil, err
	}
	return &Publisher{
		token: os.QStashToken,
		url:   os.QStashURL,
		topic: os.topic,
		uuid:  new(uuid),
		client: &httpClient{
			client: &http.Client{
				Timeout: os.Client.Timeout,
			},
			MaxBackOff: os.Client.MaxBackOff,
			MinBackOff: os.Client.MinBackOff,
			Retries:    os.Client.Retries,
		},
		verbose: os.Verbose,
	}, nil
}

// Publish publishes a message to the QStash
func (q *Publisher) Publish(ctx context.Context, m *Message, opts ...PublishOption) error {
	// Parse the publish options
	var os PublishOptions
	if opts != nil {
		if err := os.apply(opts...); err != nil {
			return fmt.Errorf("bad options: %w", err)
		}
	}
	// Create the request
	r, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/%s", q.url, q.topic),
		bytes.NewBuffer(m.Body),
	)
	if err != nil {
		return fmt.Errorf("could not create request %w", err)
	}

	// Validate and add the optional message headers
	if m.Headers != nil {
		for k := range m.Headers {
			if !strings.HasPrefix(strings.ToLower(k), "upstash-forward-") {
				return fmt.Errorf("headers must start with 'Upstash-Forward-'")
			}
		}
		r.Header = m.Headers
	}

	// Determine the deduplication id
	if hasID := len(m.ID) > 0; hasID && os.ContentBasedDeduplication {
		return fmt.Errorf("you cannot set 'content based deduplication' and pass a custom deduplication id")
	} else if os.ContentBasedDeduplication {
		r.Header.Set("Upstash-Content-Based-Deduplication", "true")
	} else if hasID {
		r.Header.Set("Upstash-Deduplication-ID", m.ID)
	} else if deduplicationID, err := q.uuid.NewV4(); err != nil {
		return fmt.Errorf("could not generate uuid %w", err)
	} else {
		// By default, generate a uuid to allow for retries on publish
		r.Header.Set("Upstash-Deduplication-ID", deduplicationID)
	}

	// Set the standard request headers
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", q.token))
	r.Header.Set("Content-Type", "application/json")

	// Configure scheduling and retry functionality
	if os.Delay > 0 {
		r.Header.Set("Upstash-Delay", os.Delay.String())
	}
	if len(os.Schedule) > 0 {
		r.Header.Set("Upstash-Schedule", os.Schedule)
	}
	if os.Retries > 0 {
		r.Header.Set("Upstash-Retries", strconv.Itoa(os.Retries))
	}

	// Publish the message
	rsp, err := q.client.Do(r.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("could not complete request %w", err)
	} else if rsp.StatusCode < 200 || rsp.StatusCode > 299 {
		bs, _ := io.ReadAll(rsp.Body)
		rsp.Body.Close()
		return fmt.Errorf("bad request status %d: %s", rsp.StatusCode, string(bs))
	}

	// Return the message id
	var body struct {
		MessageID string `json:"messageId"`
	}
	defer rsp.Body.Close()
	if err := json.NewDecoder(rsp.Body).Decode(&body); err != nil {
		return fmt.Errorf("could not decode response %w", err)
	}
	m.ID = body.MessageID

	// Success
	return nil
}

// PublishWithSchedule publishes a message to the QStash with a chron schedule
// Note: see https://crontab.guru/ for help with the schedule format
func (q *Publisher) PublishWithSchedule(ctx context.Context, message *Message, schedule string, opts ...PublishOption) error {
	return q.Publish(ctx, message, append(opts, WithSchedule(schedule))...)
}

// PublishWithDelay publishes a message to the QStash with a delay
func (q *Publisher) PublishWithDelay(ctx context.Context, message *Message, delay time.Duration, opts ...PublishOption) error {
	return q.Publish(ctx, message, append(opts, WithDelay(delay))...)
}
