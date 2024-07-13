package qstash

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

type mockClient struct {
	r *http.Request
}

func (c *mockClient) Do(r *http.Request) (*http.Response, error) {
	c.r = r
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("{ \"messageId\":\"mock-id\" }")),
	}, nil
}

type mockUUID struct {
	uuid string
	err  error
}

func (u *mockUUID) NewV4() (string, error) {
	return u.uuid, u.err
}

func TestPublisher_Publish(t *testing.T) {
	type fields struct {
		token  string
		url    string
		topic  string
		client *mockClient
		uuid   *mockUUID
	}
	type args struct {
		message Message
		opts    []PublishOption
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantHeader http.Header
		wantURL    string
		wantBody   []byte
	}{{
		name: "Publish with no options",
		fields: fields{
			token:  "token",
			url:    "url",
			topic:  "topic",
			client: &mockClient{},
			uuid: &mockUUID{
				uuid: "uuid",
			},
		},
		args: args{
			message: Message{
				Body: []byte("message"),
			},
		},
		wantErr: false,
		wantHeader: http.Header{
			"Authorization":            []string{"Bearer token"},
			"Content-Type":             []string{"application/json"},
			"Upstash-Deduplication-ID": []string{"uuid"},
		},
		wantURL:  "url/topic",
		wantBody: []byte("message"),
	}, {
		name: "Publish with delay",
		fields: fields{
			token:  "token",
			url:    "url",
			topic:  "topic",
			client: &mockClient{},
			uuid: &mockUUID{
				uuid: "uuid",
			},
		},
		args: args{
			message: Message{
				Body: []byte("message"),
			},
			opts: []PublishOption{
				WithDelay(time.Second),
			},
		},
		wantErr: false,
		wantHeader: http.Header{
			"Authorization":            []string{"Bearer token"},
			"Content-Type":             []string{"application/json"},
			"Upstash-Deduplication-ID": []string{"uuid"},
			"Upstash-Delay":            []string{"1s"},
		},
		wantURL:  "url/topic",
		wantBody: []byte("message"),
	}, {
		name: "Publish with custom headers",
		fields: fields{
			token:  "token",
			url:    "url",
			topic:  "topic",
			client: &mockClient{},
			uuid: &mockUUID{
				uuid: "uuid",
			},
		},
		args: args{
			message: Message{
				Headers: http.Header{
					"Upstash-Forward-Key": []string{"value"},
				},
				Body: []byte("message"),
			},
		},
		wantErr: false,
		wantHeader: http.Header{
			"Authorization":            []string{"Bearer token"},
			"Content-Type":             []string{"application/json"},
			"Upstash-Deduplication-ID": []string{"uuid"},
			"Upstash-Forward-Key":      []string{"value"},
		},
		wantURL:  "url/topic",
		wantBody: []byte("message"),
	}, {
		name: "Publish with headers with bad prefix fails",
		fields: fields{
			token:  "token",
			url:    "url",
			topic:  "topic",
			client: &mockClient{},
			uuid: &mockUUID{
				uuid: "uuid",
			},
		},
		args: args{
			message: Message{
				Headers: http.Header{
					"key": []string{"value"},
				},
				Body: []byte("message"),
			},
		},
		wantErr: true,
	}, {
		name: "Publish with custom id",
		fields: fields{
			token:  "token",
			url:    "url",
			topic:  "topic",
			client: &mockClient{},
			uuid: &mockUUID{
				uuid: "uuid",
			},
		},
		args: args{
			message: Message{
				ID:   "custom-deduplication-id",
				Body: []byte("message"),
			},
		},
		wantErr: false,
		wantHeader: http.Header{
			"Authorization":            []string{"Bearer token"},
			"Content-Type":             []string{"application/json"},
			"Upstash-Deduplication-ID": []string{"custom-deduplication-id"},
		},
		wantURL:  "url/topic",
		wantBody: []byte("message"),
	}, {
		name: "Publish with a content based deduplication id",
		fields: fields{
			token:  "token",
			url:    "url",
			topic:  "topic",
			client: &mockClient{},
			uuid: &mockUUID{
				uuid: "uuid",
			},
		},
		args: args{
			message: Message{
				Body: []byte("message"),
			},
			opts: []PublishOption{
				WithContentBasedDeduplication(),
			},
		},
		wantErr: false,
		wantHeader: http.Header{
			"Authorization":                       []string{"Bearer token"},
			"Content-Type":                        []string{"application/json"},
			"Upstash-Content-Based-Deduplication": []string{"true"},
		},
		wantURL:  "url/topic",
		wantBody: []byte("message"),
	}, {
		name: "Publish with custom id and content based deduplication fails",
		fields: fields{
			token:  "token",
			url:    "url",
			topic:  "topic",
			client: &mockClient{},
			uuid: &mockUUID{
				uuid: "uuid",
			},
		},
		args: args{
			message: Message{
				ID:   "custom-deduplication-id",
				Body: []byte("message"),
			},
			opts: []PublishOption{
				WithContentBasedDeduplication(),
			},
		},
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Publisher{
				token:  tt.fields.token,
				url:    tt.fields.url,
				topic:  tt.fields.topic,
				client: tt.fields.client,
				uuid:   tt.fields.uuid,
			}
			if err := q.Publish(context.TODO(), &tt.args.message, tt.args.opts...); err != nil {
				if !tt.wantErr {
					t.Fatalf("Publisher.Publish() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			// Verify the url
			if tt.wantURL != tt.fields.client.r.URL.String() {
				t.Fatalf("Publisher.Publish() url = %v, want %v", tt.fields.client.r.URL.String(), tt.wantURL)
				return
			}
			// Verify the headers
			if len(tt.wantHeader) != len(tt.fields.client.r.Header) {
				t.Errorf("Publisher.Publish() header length = %v, want %v", len(tt.fields.client.r.Header), len(tt.wantHeader))
				t.Fatalf("Publisher.Publish() header = %v, want %v", tt.fields.client.r.Header, tt.wantHeader)
			}
			for k, v := range tt.wantHeader {
				if tt.fields.client.r.Header.Get(k) != v[0] {
					t.Fatalf("Publisher.Publish() header %v = %v, want %v", k, tt.fields.client.r.Header.Get(k), v[0])
				}
			}
			// Verify the body
			if bs, err := io.ReadAll(tt.fields.client.r.Body); err != nil {
				t.Fatalf("Publisher.Publish() error reading body = %v", err)
			} else if string(tt.wantBody) != string(bs) {
				t.Fatalf("Publisher.Publish() body = %v, want %v", tt.fields.client.r.Body, tt.wantBody)
			}

		})
	}
}
