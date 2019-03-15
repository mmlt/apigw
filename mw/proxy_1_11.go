// +build go1.11

package mw

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/http/httputil"
	"strings"
)

// This is a copy of echo/middleware/proxy.go with modifications to provided our own
// httputil.ReverseProxy Director function.

// ProxyHTTP returns a ReverseProxy that adds a Host header (https://tools.ietf.org/html/rfc7230#page-44)
// Based on net.http.httputil.reverseproxy.go NewSingleHostReverseProxy()
func proxyHTTP(tgt *ProxyTarget, c echo.Context, config ProxyConfig) http.Handler {
	target := tgt.URL
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		req.Host = target.Host // https://github.com/golang/go/issues/5692
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	errorHandler := func(resp http.ResponseWriter, req *http.Request, err error) {
		desc := tgt.URL.String()
		if tgt.Name != "" {
			desc = fmt.Sprintf("%s(%s)", tgt.Name, tgt.URL.String())
		}
		c.Logger().Errorf("remote %s unreachable, could not forward: %v", desc, err)
		c.Error(echo.NewHTTPError(http.StatusServiceUnavailable))
	}

	proxy := &httputil.ReverseProxy{
		Director: director,
		ErrorHandler: errorHandler,
		Transport: config.Transport,
	}

	return proxy
}

// SingleJoiningSlash is copied from net.http.httputil.reverseproxy.go
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}



// 4.0.0 proxyHTTP implementation
func v400proxyHTTP(tgt *ProxyTarget, c echo.Context, config ProxyConfig) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(tgt.URL)
	proxy.ErrorHandler = func(resp http.ResponseWriter, req *http.Request, err error) {
		desc := tgt.URL.String()
		if tgt.Name != "" {
			desc = fmt.Sprintf("%s(%s)", tgt.Name, tgt.URL.String())
		}
		c.Logger().Errorf("remote %s unreachable, could not forward: %v", desc, err)
		c.Error(echo.NewHTTPError(http.StatusServiceUnavailable))
	}
	proxy.Transport = config.Transport
	return proxy
}

// 3.3.4 proxyHTTP implementation with fix.
func v334proxyHTTP(t *ProxyTarget, c echo.Context, config ProxyConfig) http.Handler {
	director := func(req *http.Request) {
		req.URL.Scheme = t.URL.Scheme
		req.URL.Host = t.URL.Host
		req.Host = t.URL.Host // https://github.com/golang/go/issues/5692

		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable user agent so it's not to setthe default value
			req.Header.Set("User-Agent", "")
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

