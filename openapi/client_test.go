package openapi

import (
	"context"
	"fmt"
	"github.com/mmlt/apigw/backoff"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"github.com/mmlt/apigw/path"
)

// TestGet shows that we read from a HTTP endpoint even if this requires retries.
func TestGet(t *testing.T) {
	// Mock TimeSleep to speed-up this test.
	bu := backoff.TimeSleep
	defer func() {
		backoff.TimeSleep = bu
	}()
	backoff.TimeSleep = func(d time.Duration) {}

	// Prepare a HTTP server that counts the number of requests.
	var requests int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests > 0 {
			fmt.Fprint(w, "response")
		} else {
			w.WriteHeader(500)
		}
		requests++
	}))
	defer ts.Close()

	c := NewClient(context.Background(), ts.URL)
	b, err := c.Get()

	assert.NoError(t, err)
	assert.EqualValues(t, 2, requests)
	assert.EqualValues(t, "response", string(b))
}

// TestPoll shows that we can detect changes in the OpenAPI spec.
func TestPoll(t *testing.T) {
	var swagger = `{
		"swagger": "2.0",
  		"info": {
			"version": "v%d",
    		"title": "MyBank.OpenApi"
		},
		"paths": {
    		"/version": { }
		}
	}`

	// Prepare a HTTP server that returns a different spec every 2 requests.
	var requests int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, swagger, requests/2)
		requests++
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	client := NewClient(ctx, ts.URL)
	// Run Poll until 3 changes are detected.
	indices := make(chan *path.Index, 3)
	done := make(chan interface{})
	go client.Poll(time.Millisecond, func(idx *path.Index) {
		indices <- idx
		if len(indices) == 3 {
			cancel()
			close(done)
		}
	})
	// Give Poll() time to do its thing.
	<-done
	// For 3 changes to take place 5 requests are needed (request 1 returns v0; 2 v0; 3 v1; 4 v1; 5 v2)
	assert.EqualValues(t, 5, requests)
}
