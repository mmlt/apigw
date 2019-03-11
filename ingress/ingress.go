// Package ingress implements a reverse proxy with oauth2 authorization.
package ingress

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/mmlt/apigw/mw"
	"net/http"
	"net/url"
	"text/template"
)

type (
	// Config defines the configuration for an Ingress.
	// TODO Config.Middleware.Path struct matches PathConfig struct except the former has yaml struct tags. Consider removing
	// this duplication (and struct-to-struct copying) by added tag to PathConfig or converting to PathConfig (see spec. struct_types)
	Config struct {
		// Bind port.
		Bind string `yaml:"bind"`
		// TLS
		TLS struct {
			// Key is path of key.pem file.
			Key string `yaml:"key"`
			// Key is path of cert.pem file.
			Cert string `yaml:"cert"`
		} `yaml:"tls"`
		// Middleware
		Middleware struct {
			Path struct {
				RequirePrefix string `yaml:"requirePrefix"`
				TrimPrefix    string `yaml:"trimPrefix"`
			} `yaml:"path"`
			// CORS headers
			Cors struct {
				AllowOrigins []string `yaml:"allowOrigins"`
				AllowMethods []string `yaml:"allowMethods"`
			} `yaml:"cors"`
			// Reverse proxy
			Proxy struct {
				// Targets are the url(s) of upstream servers.
				Targets []string `yaml:"targets"`
			} `yaml:"proxy"`
		} `yaml:"middleware"`
		// Error response template (expanded with Status and Message parameters).
		ErrorResponse string `yaml:"errorResponse"`
	}

	// Ingress holds the state for a reverse proxy with oauth2 authorization.
	Ingress struct {
		Port string
		Echo *echo.Echo
		// CertFile contains the path of the TLS cert file.
		certFile string
		// KeyFile contains the path of the TLS key file.
		keyFile string
	}
)

// NewWithConfig creates an Ingress instance.
func NewWithConfig(cfg *Config, scopesFn mw.ScopesFunc, tokeninfoFn mw.TokeninfoFunc, allowMethodsFn mw.AllowMethodsFunc) *Ingress {
	e := echo.New()
	e.HideBanner = true

	if cfg.ErrorResponse == "" {
		glog.Warning("config: errorResponse is not set.")
	}

	// Add error handler that shows custom messages.
	var err error
	e.HTTPErrorHandler, err = customHTTPErrorHandler(cfg.ErrorResponse)
	if err != nil {
		glog.Fatal(err)
	}

	e.Use(mw.Logger())

	// Path rewriting
	e.Use(mw.PathWithConfig(mw.PathConfig{
		RequirePrefix: cfg.Middleware.Path.RequirePrefix,
		TrimPrefix:    cfg.Middleware.Path.TrimPrefix,
	}))

	// CORS headers
	e.Use(mw.CORSWithConfig(mw.CORSConfig{
		AllowOrigins: cfg.Middleware.Cors.AllowOrigins,
		//TODO AllowMethods: cfg.Middleware.Cors.AllowMethods,
		AllowMethodsFn: allowMethodsFn,
	}))

	// Setup OAuth2 authorization
	e.Use(mw.OAuth2WithConfig(mw.OAuth2Config{
		RequiredScopes: scopesFn,
		Tokeninfo:      tokeninfoFn,
	}))

	// Setup reverse proxy with load balancer.
	targets := []*middleware.ProxyTarget{}
	for _, t := range cfg.Middleware.Proxy.Targets {
		u, err := url.Parse(t)
		if err != nil {
			glog.Fatal(err)
		}
		targets = append(targets, &middleware.ProxyTarget{URL: u})
	}
	lb := middleware.NewRoundRobinBalancer(targets)
	e.Use(middleware.Proxy(lb))

	return &Ingress{Port: cfg.Bind, Echo: e, certFile: cfg.TLS.Cert, keyFile: cfg.TLS.Key}
}

// Run the ingress.
func (in *Ingress) Run() error {
	if in.certFile == "" {
		return in.Echo.Start(in.Port)
	} else {
		return in.Echo.StartTLS(in.Port, in.certFile, in.keyFile)
	}
}

// CustomHTTPErrorHandler returns a func of type echo.HTTPErrorHandler that writes error messages to the HTTP response stream.
// Messages are generated with a golang template and Status and Message parameters.
func customHTTPErrorHandler(tmpl string) (echo.HTTPErrorHandler, error) {
	t, err := template.New("error").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	return func(in error, c echo.Context) {
		e := struct {
			Status  int
			Message string
		}{
			Status:  http.StatusInternalServerError,
			Message: in.Error(),
		}
		if he, ok := in.(*echo.HTTPError); ok {
			e.Status = he.Code
			e.Message = fmt.Sprintf("%v", he.Message)
		}
		c.Response().WriteHeader(e.Status)
		err := t.Execute(c.Response().Writer, e)
		if err != nil {
			glog.Warning("customHTTPErrorHandler: %v", err)
			// respond with plain json
			c.JSON(e.Status, e)
		}
	}, nil
}
