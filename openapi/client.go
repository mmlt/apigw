package openapi

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/golang/glog"
	"github.com/mmlt/apigw/backoff"
	"io/ioutil"
	"net/http"
	"time"
	"github.com/mmlt/apigw/path"
	"github.com/go-openapi/spec"
)

var ErrNoPathInSpec = fmt.Errorf("openapi definition doesn't contain paths")

type (
	// Client gets an OpenAPI definition from a HTTP endpoint.
	Client struct {
		url     string
		backoff *backoff.Backoff
		ctx     context.Context
	}

	// PollResultFunc is the type of function that is called when a new OpenAPI definition is detected.
	PollResultFunc func(index *path.Index)
)

// NewClient returns a client.
func NewClient(ctx context.Context, url string) *Client {
	return &Client{
		url:     url,
		backoff: backoff.New(ctx, 8, time.Second),
		ctx:     ctx,
	}
}

// Poll starts a loop that checks for changes in the OpenAPI definition.
// The fn is called when a new definition is detected.
// Use Shutwdown() to stop polling.
func (c *Client) Poll(interval time.Duration, fn PollResultFunc) {
	var checksum [16]byte

	for ; c.ctx.Err() == nil; c.sleep(interval) {
		// Get OpenAPI definition from endpoint.
		b, err := c.Get()
		if err != nil {
			// try again
			glog.Error(err)
			continue
		}

		// Check if definition has changed
		cs := md5.Sum(b)
		if cs == checksum {
			continue
		}

		// Create index
		idx, err := parse(b)
		if err != nil {
			glog.Error("parse OpenAPI json: ", err)
			continue
		}

		fn(idx)
		checksum = cs
	}
}

// Get performs a HTTP GET with backoff.
func (c *Client) Get() (b []byte, err error) {
	bo := *c.backoff //copy
	bo.Do(func() error {
		b, err = httpGetReadAll(c.url)
		return err
	})
	return
}

// HTTPGetReadAll gets a []byte from an URL.
func httpGetReadAll(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// Parse translates an OpenAPI spec into an Index.
func parse(json []byte) (*path.Index, error) {
	// Parse openapiClient.
	spec, err := SpecFromRaw(json)
	if err != nil {
		return nil, err
	}

	// Check if spec contains paths
	if spec.Paths == nil {
		return nil, ErrNoPathInSpec
	}

	// Check number of paths in openapi definition.
	i := len(spec.Paths.Paths)
	if i == 0 {
		return nil, ErrNoPathInSpec
	}

	glog.Infof("openapi definition fetch successful (contains %d paths)", i)

	// Build index for quick lookups.
	idx, err := newIndexFromSpec(spec)

	return idx, err
}

// NewIndexFromSpec returns a path.Index instance for lookup of scopes by method/path.
// The required scopes for a method/path are read from swagger 'security' 'oauth2' sections.
func newIndexFromSpec(spec *spec.Swagger) (*path.Index, error) {
	idx := path.NewIndex()
	SpecOAuth2ScopeIter(spec, func(path string, method string, scopes []string) {
		// Swagger spec scopes can contain empty strings, remove them.
		var s []string
		for _, str := range scopes {
			if str != "" {
				s = append(s, str)
			}
		}
		idx.AddMethodPathScopes(method, path, s)
	})
	return idx, nil
}

// Sleep with cancel.
func (c *Client) sleep(d time.Duration) {
	select {
	case <-time.After(d):
	case <-c.ctx.Done():
	}
}
