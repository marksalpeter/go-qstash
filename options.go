package qstash

import (
	"fmt"
	"os"
	"time"
)

// ReceiverOptions come from the environment or they can be overridden
type ReceiverOptions struct {
	SigningKey     string
	NextSigningKey string
}

func (o *ReceiverOptions) apply(opts ...ReceiverOption) error {
	// Apply the receiver options
	for _, opt := range append(defaultReceiverOptions, opts...) {
		opt(o)
	}
	// Validate the options
	if o.SigningKey == "" {
		return fmt.Errorf("'QSTASH_SIGNING_KEY' is required")
	}
	if o.NextSigningKey == "" {
		return fmt.Errorf("'QSTASH_NEXT_SIGNING_KEY' is required")
	}
	return nil
}

// ReceiverOption overrides on of the default options
type ReceiverOption func(*ReceiverOptions)

// WithQStashSigningKey overrides the default QStash signing key
func WithSigningKey(signingKey string) ReceiverOption {
	return func(o *ReceiverOptions) {
		o.SigningKey = signingKey
	}
}

// WithNextSigningKey overrides the default QStash next signing key
func WithNextSigningKey(signingKey string) ReceiverOption {
	return func(o *ReceiverOptions) {
		o.NextSigningKey = signingKey
	}
}

// defaultOptions are the default options
var defaultReceiverOptions = []ReceiverOption{
	WithSigningKey(os.Getenv("QSTASH_SIGNING_KEY")),
	WithNextSigningKey(os.Getenv("QSTASH_NEXT_SIGNING_KEY")),
}

// PublisherOptions represents the options for a qstash.Publisher
type PublisherOptions struct {
	QStashURL   string
	QStashToken string
	Client      struct {
		Timeout    time.Duration
		MaxBackOff time.Duration
		MinBackOff time.Duration
		Retries    int
	}
	Verbose bool
	topic   string
}

// apply applies the publisher options and validates them
func (o *PublisherOptions) apply(opts ...PublisherOption) error {
	// Apply the publisher options
	for _, opt := range append(defaultPublisherOptions, opts...) {
		opt(o)
	}
	// Validate the options
	if o.QStashToken == "" {
		return fmt.Errorf("'QSTASH_TOKEN' is required")
	}
	if o.QStashURL == "" {
		return fmt.Errorf("qstash url is required")
	}
	if o.topic == "" {
		return fmt.Errorf("topic is required")
	}
	if o.Client.Timeout < time.Millisecond {
		return fmt.Errorf("http client timeout must at least 1 millisecond")
	}
	if o.Client.Retries < 0 {
		return fmt.Errorf("http client retries must be at least 0")
	}
	if o.Client.MinBackOff < time.Millisecond {
		return fmt.Errorf("http client min back off must at least 1 millisecond")
	}
	if o.Client.MaxBackOff < time.Millisecond {
		return fmt.Errorf("http client max back off must at least 1 millisecond")
	}
	if o.Client.MinBackOff > o.Client.MaxBackOff {
		return fmt.Errorf("http client min back off must be less than or equal to max back off")
	}
	return nil
}

// PublisherOption overrides one of the default publisher options
type PublisherOption func(*PublisherOptions)

// WithClientMaxBackOff overrides the default http client max back off
func WithClientMaxBackOff(maxBackOff time.Duration) PublisherOption {
	return func(o *PublisherOptions) {
		o.Client.MaxBackOff = maxBackOff
	}
}

// WithClientMinBackOff overrides the default http client min back off
func WithClientMinBackOff(minBackOff time.Duration) PublisherOption {
	return func(o *PublisherOptions) {
		o.Client.MinBackOff = minBackOff
	}
}

// WithClientRetries overrides the default http client retries
func WithClientRetries(retries int) PublisherOption {
	return func(o *PublisherOptions) {
		o.Client.Retries = retries
	}
}

// WithClientTimeout overrides the default http client timeout
func WithClientTimeout(timeout time.Duration) PublisherOption {
	return func(o *PublisherOptions) {
		o.Client.Timeout = timeout
	}
}

// WithQStashURL sets the url for the qstash publisher
// The default url is https://qstash.upstash.io/v1/publish
func WithQStashURL(url string) PublisherOption {
	return func(o *PublisherOptions) {
		o.QStashURL = url
	}
}

// WithQStashToken sets the token for the qstash publisher
// The default token is the QSTASH_TOKEN environment variable
func WithQStashToken(token string) PublisherOption {
	return func(o *PublisherOptions) {
		o.QStashToken = token
	}
}

// WithVerbose will make the publisher log the http responses of the publish requests
// for debugging purposes
func WithVerbose() PublisherOption {
	return func(o *PublisherOptions) {
		o.Verbose = true
	}
}

// withTopic sets the topic for the qstash publisher
func withTopic(topic string) PublisherOption {
	return func(o *PublisherOptions) {
		o.topic = topic
	}
}

// defaultPublisherOptions are the default publisher options
var defaultPublisherOptions = []PublisherOption{
	WithQStashURL("https://qstash.upstash.io/v2/publish"),
	WithQStashToken(os.Getenv("QSTASH_TOKEN")),
	WithClientTimeout(time.Second),
	WithClientMaxBackOff(time.Second),
	WithClientMinBackOff(200 * time.Millisecond),
	WithClientRetries(5),
}

// PublishOptions represents the options for an individual publish request
type PublishOptions struct {
	Delay                     time.Duration
	Retries                   int
	ContentBasedDeduplication bool
}

// apply applies the publish options and validates them
func (o *PublishOptions) apply(opts ...PublishOption) error {
	// Apply the publish options
	for _, opt := range opts {
		opt(o)
	}
	return nil
}

// PublishOption overrides one of the default publish options
type PublishOption func(*PublishOptions)

// WithDelay sets the delay for the message
func WithDelay(delay time.Duration) PublishOption {
	return func(o *PublishOptions) {
		o.Delay = delay
	}
}

// WithContentBasedDeduplication sets the content base deduplication header
// WARNING: this will override the unique message ids generated by the qstash publisher
//
//	and can cause dropped messages
func WithContentBasedDeduplication() PublishOption {
	return func(o *PublishOptions) {
		o.ContentBasedDeduplication = true
	}
}

// WithRetries overrides the number of retries for the message
func WithRetries(retries int) PublishOption {
	return func(o *PublishOptions) {
		o.Retries = retries
	}
}
