// Package gateway implements a reverse proxy for openapis.org endpoints with OAuth2 authorization.
package gateway

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.com/mmlt/apigw/ingress"
	"github.com/mmlt/apigw/mw"
	"github.com/mmlt/apigw/openapi"
	"net/http"
	"net/url"
	"sync"
	"time"
	"github.com/mmlt/apigw/path"
	"github.com/labstack/echo/v4"
)

type (
	// Config for an gateway.
	Config struct {
		// Ingress handles the api traffic.
		Ingress ingress.Config `yaml:"ingress"`
		// Management handles the /metrics traffic.
		Management struct {
			Bind string `yaml:"bind"`
		} `yaml:"management"`
		// Openapi defines how to ingest the API definition.
		Openapi struct {
			URL string `yaml:"url"`
		} `yaml:"openapi"`
		// Oauth2Idp defines how to connect to the OAuth2 IDP.
		Oauth2Idp struct {
			TokeninfoURL string `yaml:"tokeninfoUrl"`
		} `yaml:"oauth2idp"`
	}

	// Gateway contains the state of an Apigw instance.
	Gateway struct {
		// Context for apigw lifecycle.
		ctx context.Context
		// Cancel the operation of this apigw.
		cancel context.CancelFunc
		// Config of the gateway.
		cfg *Config
		// Ingress handles incoming API traffic.
		in *ingress.Ingress
		// Tokeninfo client gets scopes based on access token.
		tic *mw.TokeninfoClient
		// Swagger client gets an openapi definition.
		openapiClient *openapi.Client
	}
)

// NewWithConfig returns an initialized Apigw.
func NewWithConfig(c *Config) *Gateway {
	ctx, fn := context.WithCancel(context.Background())
	return &Gateway{
		ctx:    ctx,
		cancel: fn,
		cfg:    c,
	}
}

// Run an Gateway.
func (gw *Gateway) Run() error {
	// Check if tokeninfo is reachable.
	err := pingURL(gw.cfg.Oauth2Idp.TokeninfoURL)
	if err != nil {
		return err
	}
	glog.Infof("ping idp at %s successful.", gw.cfg.Oauth2Idp.TokeninfoURL)

	// Get Swagger definition via HTTP
	var index *path.Index
	var m sync.RWMutex
	gw.openapiClient = openapi.NewClient(gw.ctx, gw.cfg.Openapi.URL)
	go gw.openapiClient.Poll(time.Minute, func(idx *path.Index) {
		glog.Info("switch to new OpenAPI definition")
		m.Lock()
		index = idx
		m.Unlock()
	})

	// scopesFn looks-up scopes in the index.
	// Note that:
	// - the index may be swapped anytime (when a new swagger.json is read and parsed successfully)
	// - the lookup may fail because no OpenAPI definition read (yet)
	scopesFn := func(method string, url *url.URL) ([]string, error) {
		m.RLock()
		idx := index
		m.RUnlock()
		if idx == nil {
			return []string{}, fmt.Errorf("No OpenAPI definition read (yet).") //TODO use error const
		}
		ss, err := idx.FindScopes(method, url.Path)
		if err != nil {
			// any error while looking for method/path is a 404
			err = echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return ss, err
	}

	allowMethodsFn := func(path string) ([]string, error) {
		m.RLock()
		idx := index
		m.RUnlock()
		if idx == nil {
			return []string{}, fmt.Errorf("No OpenAPI definition read (yet).") //TODO use error const
		}
		ss, err := idx.FindMethods(path)
		if err != nil {
			// any error while looking for path is a 404
			err = echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return ss, err
	}

	/*	TODO consider merging BasicTokeninfo into TokeninfoClient
		gw.in = ingress.NewWithConfig(
			&gw.cfg.Ingress,
			scopesFn,
			mw.BasicTokeninfo(gw.cfg.Oauth2Idp.TokeninfoURL))*/

	gw.tic = mw.NewTokeninfoClient(gw.cfg.Oauth2Idp.TokeninfoURL)
	gw.in = ingress.NewWithConfig(
		&gw.cfg.Ingress,
		scopesFn,
		gw.tic.Call,
		allowMethodsFn)

	gw.tic.EnableGC(true)
	return gw.in.Run()
}

// Shutdown stops server the gracefully.
func (gw *Gateway) Shutdown(ctx context.Context) error {
	gw.cancel()
	// TODO remove gw.openapiClient.Shutdown(ctx)
	gw.tic.EnableGC(false)
	return gw.in.Echo.Shutdown(ctx)
}

// ShutdownWithTimeout attempts to stop server the gracefully but waits no more then the specified time for connections to close.
func (gw *Gateway) ShutdownWithTimeout(to time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()
	return gw.Shutdown(ctx)
}

// PingURL returns nil if an url is reachable and an error otherwise.
func pingURL(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusBadRequest, http.StatusUnauthorized:
		// we didn't provide a token so we should expect an error
		return nil
	default:
		return fmt.Errorf("GET %s failed with %d", url, resp.StatusCode)
	}
}
