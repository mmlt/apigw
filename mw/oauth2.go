package mw

/*
	OAuth2 middleware allows or denies traffic to pass based on Authorization header.

	The OAuth2 ClientID is added to context to enable other middleware to do client specific things like throttling.

	See OAuth2 RFC at https://tools.ietf.org/html/rfc6749
*/

import (
	"net/http"
	"strings"

	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io/ioutil"
	"net/url"
	"sync"
	"time"
)

type (
	// OAuth2Config defines the config for checking bearer tokens against an oauth2 idp.
	OAuth2Config struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		// Tokeninfo is a function that gets a TokeninfoResponse for a given access_token.
		Tokeninfo TokeninfoFunc

		// RequiredScopes is an optional function that gets the scopes that are required to access a path.
		//
		// If RequiredScopes is not set:
		//  missing token -> BadRequest
		//  invalid or expired tokens -> Unauthorized
		//
		// If RequiredScopes is set and the resulting set of scopes is empty no error is returned and processing
		// continues with the next middleware.
		//
		// If RequiredScopes is set and it returns a non empty set of scopes for a path:
		//  missing token -> BadRequest
		//  invalid or expired tokens -> Unauthorized
		//  scopes don't match -> Unauthorized
		RequiredScopes ScopesFunc
	}

	// ScopesFunc gets the scopes to access a path.
	ScopesFunc func(method string, url *url.URL) ([]string, error)

	// TokeninfoFunc gets a TokeninfoResponse for a given access_token.
	TokeninfoFunc func(token string) (*TokeninfoResponse, error)

	// TokeninfoResponse is the result of a tokeninfo call.
	// See https://tools.ietf.org/html/rfc7662#page-8
	TokeninfoResponse struct {
		// ClientID identifies the application for which the token created.
		ClientID string `json:"client_id"`
		// Scope contains zero or more oauth2 scopes.
		Scope []string `json:"scope"`
		// ExpiresIn is the number of seconds this session is still valid.
		ExpiresIn int `json:"expires_in"`
		// Error is empty on success or contains an error name when something went wrong.
		Error string `json:"error"`
		//TODO Expires is calculated from ExpiresIn at time of reception.
		//TODOExpires time.Time
		// Scopes is a map representation of Scope.
		Scopes map[string]struct{}
		// Timestamp is when the tokeninfo response is produced.
		// Timestamp + ExpiresIn is the absolute expiry time.
		Timestamp time.Time
	}
)

// Errors
var (
	ErrTokenMissing = echo.NewHTTPError(http.StatusBadRequest, "Missing or malformed token")
	ErrTokenInvalid = echo.NewHTTPError(http.StatusUnauthorized, "Not allowed")
)

var (
	// DefaultOauth2Config is the default config.
	DefaultOauth2Config = OAuth2Config{
		Skipper: middleware.DefaultSkipper,
	}
)

// OAuth2 returns an OAuth2 auth middleware.
func OAuth2() echo.MiddlewareFunc {
	c := DefaultOauth2Config
	return OAuth2WithConfig(c)
}

// OAuth2WithConfig returns an OAuth2 auth middleware with config.
// See: `OAuth2Config`.
func OAuth2WithConfig(config OAuth2Config) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultOauth2Config.Skipper
	}
	if config.Tokeninfo == nil {
		panic("echo: OAuth2 middleware requires a Tokeninfo function.")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			var err error

			// get required scopes
			var required []string
			if config.RequiredScopes != nil {
				required, err = config.RequiredScopes(c.Request().Method, c.Request().URL)
				if err != nil {
					return err
				}
			}

			if len(required) == 0 {
				// no required scopes, proceed
				return next(c)
			}

			// get token
			token, err := extractToken(c)
			if err != nil {
				return err
			}
			// get tokeninfo
			ti, err := config.Tokeninfo(token)
			if err != nil || ti == nil {
				// log internal error but don't let the caller know.
				glog.Info(err)
				return ErrTokenInvalid
			}

			// check if token still valid.
			if ti.ExpiresIn <= 0 {
				return ErrTokenInvalid
			}

			// check allowed scopes (allowed is a superset of required scopes)
			allowed := ti.Scopes
			// all required scopes must be allowed
			for _, r := range required {
				if _, ok := allowed[r]; !ok {
					glog.V(2).Infof("%s %s requires scopes %v (allowed=%v)", c.Request().Method, c.Request().URL, required, allowed)
					return ErrTokenInvalid
				}
			}

			// make it possible for other middleware to use ClientID.
			c.Set("ClientID", ti.ClientID)

			return next(c)
		}
	}
}

// ExtractToken gets an OAuth2 token from a header of the form: `Authorization : Bearer cn389ncoiwuencr`.
// See https://tools.ietf.org/html/rfc6750
func extractToken(c echo.Context) (string, error) {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return "", ErrTokenMissing
	}
	ss := strings.Split(auth, "Bearer")
	if len(ss) != 2 {
		return "", ErrTokenMissing
	}

	return strings.TrimSpace(ss[1]), nil
}

// BasicTokeninfo returns a function that calls the tokeninfo endpoint of an OAuth2 IDP.
// This implementation performs a HTTP GET for each invocation.
// See https://tools.ietf.org/html/rfc7662
func BasicTokeninfo(url string) TokeninfoFunc {
	return func(token string) (*TokeninfoResponse, error) {
		// clock the timeInserted now so we're always on the save side when adding ExpiresIn
		now := time.Now()

		resp, err := http.Get(url + "?access_token=" + token)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GET %s?access_token=xxxx status %d", url, resp.StatusCode)
		}
		// Read body.
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		// Unmarshal.
		var r TokeninfoResponse
		err = json.Unmarshal(body, &r)
		if err != nil {
			return nil, fmt.Errorf("tokeninfo response unmarshall error %v", err)
		}
		// Check json for errors.
		if r.Error != "" {
			return nil, fmt.Errorf("tokeninfo response error %v", r.Error)
		}

		// Do post processing of fields...
		// Store scope as a map.
		r.Scopes = make(map[string]struct{}, len(r.Scope))
		for _, s := range r.Scope {
			r.Scopes[s] = struct{}{}
		}
		//TODOr.Expires = time.Now().Add(time.Duration(r.ExpiresIn) * time.Second)
		r.Timestamp = now

		return &r, err
	}
}

// TokeninfoClient is used to call and OAuth2 tokeninfo endpoint with caching to reduce the load on the OAuth2 IDP.
// A side effect of caching is that a token may be invalid due to logout before it is expired. To limit this effect we
// do another tokeninfo call if the cached value is older then N seconds.
type TokeninfoClient struct {
	cache     *tokeninfoCache
	url       string
	basicCall TokeninfoFunc
}

// NewTokeninfoClient returns an initialized TokeninfoClient.
func NewTokeninfoClient(url string) *TokeninfoClient {
	return &TokeninfoClient{
		cache:     newTokeninfoCache(),
		url:       url,
		basicCall: BasicTokeninfo(url),
	}
}

// Call a tokeninfo endpoint with caching.
func (c *TokeninfoClient) Call(token string) (*TokeninfoResponse, error) {
	// Check if there is a cached entry (not older than 10 seconds) for token.
	ti, ok := c.cache.get(token)
	if ok && ti.Timestamp.Add(10*time.Second).After(time.Now()) {
		return ti, nil
	}
	// Call tokeninfo.
	ti, err := c.basicCall(token)
	if err != nil {
		return nil, err
	}
	// Store the result for future use.
	c.cache.set(token, ti)

	return ti, nil
}

// EnableGC of cached tokeninfo responses that are expired. TODO use context to shutdown GC
func (c *TokeninfoClient) EnableGC(b bool) {
	if b {
		c.cache.runGC()
	} else {
		c.cache.stopGC()
	}
}

// TokeninfoCache is a cache for tokeninfo responses.
// The cache has garbage collection to prevent ever increasing memory consumption.
type tokeninfoCache struct {
	// RWMutex
	sync.RWMutex
	// data holds cached TokeninfoResponses.
	data map[string]*TokeninfoResponse
	// ticker sets the garbage collect interval.
	ticker *time.Ticker
	// stop the garbage collector.
	stop chan struct{}
}

func newTokeninfoCache() *tokeninfoCache {
	return &tokeninfoCache{
		data: map[string]*TokeninfoResponse{},
	}
}

func (c *tokeninfoCache) get(key string) (*TokeninfoResponse, bool) {
	c.RLock()
	defer c.RUnlock()
	v, ok := c.data[key]
	return v, ok
}

func (c *tokeninfoCache) set(key string, value *TokeninfoResponse) {
	c.Lock()
	defer c.Unlock()
	c.data[key] = value
}

func (c *tokeninfoCache) delete(key string, value *TokeninfoResponse) {
	c.Lock()
	defer c.Unlock()
	delete(c.data, key)
}

func (c *tokeninfoCache) runGC() {
	if c.ticker != nil {
		// already started
		return
	}

	c.Lock()
	c.ticker = time.NewTicker(time.Minute)
	c.stop = make(chan struct{})
	c.Unlock()

	go func() {
		for {
			select {
			case <-c.ticker.C:
				now := time.Now()
				c.Lock()
				for k, v := range c.data {
					if v.Timestamp.Add(time.Duration(v.ExpiresIn) * time.Second).Before(now) {
						delete(c.data, k)
					}
				}
				c.Unlock()
			case <-c.stop:
				return
			}
		}
	}()
}

func (c *tokeninfoCache) stopGC() {
	if c.ticker == nil {
		// already stopped
		return
	}

	c.Lock()
	defer c.Unlock()

	c.ticker.Stop()
	c.ticker = nil
	close(c.stop)
	c.stop = nil
}
