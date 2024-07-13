package qstash_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/marksalpeter/go-qstash"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

// This example demonstrates how to receive messages from a qstash queue in a lambda function
func ExampleReceiver() {
	// Create a new qstash receiver
	r, err := qstash.NewReceiver()
	if err != nil {
		log.Fatal(err)
	}
	// Create a handler that verifies and processes QStash messages
	handler := r.Receive(func(ctx context.Context, msg *qstash.Message) {
		// Print the message body
		fmt.Println(msg)
		// Acknowledge the message or it will be retried
		msg.Ack()
	})
	// Usually you would host this endpoint in a serverless function
	if err := http.ListenAndServe(":80", handler); err != nil {
		log.Println(err)
	}
}

// You can publish a message to a qstash queue like this
func ExamplePublisher_publish() {
	// Create a new qstash sender
	p, err := qstash.NewPublisher("https://my-serverless-project.com/api/receive_message")
	if err != nil {
		log.Fatal(err)
	}
	// Publish a message
	if err := p.Publish(context.Background(), &qstash.Message{
		Body: []byte("Hello World!"),
	}); err != nil {
		log.Fatal(err)
	}
}

// Its also possible to add delays to a message in the queue
// Note: the delays happen on the server side, not the client side
func ExamplePublisher_publishWithDelay() {
	// Create a new qstash sender
	p, err := qstash.NewPublisher("https://my-serverless-project.com/api/receive_message")
	if err != nil {
		log.Fatal(err)
	}
	// Send a message
	if err := p.PublishWithDelay(context.Background(), &qstash.Message{
		Body: []byte("Hello In 5 Seconds!"),
	}, 5*time.Second); err != nil {
		log.Fatal(err)
	}
}

// The following example demonstrates how to use qstash to send and receive messages.
// This example uses ngrok to expose our message receiver to the internet (otherwise upstash
// will not be able to reach our endpoint).
//
// To run this example you'll need a free (upstash) https://upstash.com and (ngrok) https://ngrok.com account and the
// following environment variables will need to be set:
//
// - `QSTASH_SIGNING_KEY` - The signing key verifies the body of received messages from qstash
// - `QSTASH_NEXT_SIGNING_KEY` - The next signing key is used for key rotation during message verification
// - `QSTASH_TOKEN` - The api token is used to publish messages to your qstash instance
// - `NGROK_AUTHTOKEN` - The auth token for your ngrok account
func Example_ngrokDemoApp() {
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Create a new qstash receiver lambda handler
	// Note: This requires setting up an QStash account at https://upstash.com/
	// and setting the 'QSTASH_SIGNING_KEY' and 'QSTASH_NEXT_SIGNING_KEY'
	// environment variables
	r, err := qstash.NewReceiver()
	if err != nil {
		log.Fatal(err)
	}
	h := r.Receive(func(ctx context.Context, msg *qstash.Message) {
		// Print the message body
		log.Println("Received: ", string(msg.Body))
		// Acknowledge the message or it will be retried
		msg.Ack()
	})

	// Serve the handler with ngrok
	// NOTE: this requires setting up an NGrok account at https://ngrok.com/
	// and setting the 'NGROK_AUTHTOKEN' environment variable
	tun, err := ngrok.Listen(
		ctx,
		config.HTTPEndpoint(),
		ngrok.WithAuthtokenFromEnv(),
	)
	if err != nil {
		log.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := http.Serve(tun, h); err != nil {
			log.Print(err)
		}
	}()
	log.Println("Server is running...")

	// Publish some messages to qstash
	// Note: this requires setting the 'QSTASH_TOKEN' environment variable
	p, err := qstash.NewPublisher(tun.URL())
	if err != nil {
		log.Fatal(err)
	}
	// ...now
	if err := p.Publish(ctx, &qstash.Message{
		Body: []byte("Hello World!"),
	}); err != nil {
		log.Fatal(err)
	}
	// ... in 1 second
	if err := p.PublishWithDelay(ctx, &qstash.Message{
		Body: []byte("Hello 1 Second Later!"),
	}, 1*time.Second); err != nil {
		log.Fatal(err)
	}

	// Wait for the ngrok tunnel to shut down
	<-done
	log.Println("Server shutdown")

	// Output
	// Server is running...
	// Received:  Hello World!
	// Received:  Hello 1 Second Later!
	// Server shutdown
}
