package qstash

import (
	"net/http"
	"time"
)

// HTTPClient is a wrapper around http.Client that implements retry logic
type httpClient struct {
	client     *http.Client
	MaxBackOff time.Duration
	MinBackOff time.Duration
	Retries    int
}

// Do executes the http request with retry logic
func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	// Execute the request
	var resp *http.Response
	var err error
	for i := 1; i <= c.Retries+1; i++ {
		// Execute the request
		resp, err = c.client.Do(req)
		// If there is an error or the status code is not in the 200's, wait and try again
		if err != nil || !c.isStatusOK(resp.StatusCode) {
			time.Sleep(c.getExponentialBackOffDuration(i))
			continue
		}
		// Return the successful response
		break
	}
	return resp, err
}

// isStatusOK returns true if the status code is between 200 and 299
func (c *httpClient) isStatusOK(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// getExponentialBackOffDuration returns a the exponential back off duration between
// the min and max values based on the number of attempted requests
func (c *httpClient) getExponentialBackOffDuration(attempt int) time.Duration {
	exp := c.MinBackOff
	for i := 0; i < attempt; i++ {
		exp *= 2
		if exp > c.MaxBackOff {
			exp = c.MaxBackOff
			break
		}
	}
	return exp
}
