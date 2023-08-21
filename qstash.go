// Package qstash is a go library for the (QStash) https://upstash.com/docs/qstash/overall/getstarted service.
//
// ## Getting Started
// To get started, you need to create a qstash instance from the (Upstash Console) https://console.upstash.com/.
// You can follow the (Getting Started Guide) https://upstash.com/docs/qstash/overall/getstarted to create your first qstash instance.
//
// Once you have created a qstash instance, you will need to add the following environment variables to your project from the Upstash console:
//
// - `QSTASH_TOKEN` - The api token is used to publish messages to your qstash instance
// - `QSTASH_SIGNING_KEY` - The signing key of your qstash instance
// - `QSTASH_NEXT_SIGNING_KEY` - The next signing key of your qstash instance
//
// You must set these environment variables or pass them manually as options to the `NewReceiver` and `NewPublisher` functions.
package qstash
