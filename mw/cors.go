package mw

/*
	CORS middleware sets response headers to support the CORS protocol.
	Header values can be statically configured or looked-up on each request.

	See https://www.w3.org/TR/cors
		https://fetch.spec.whatwg.org/#http-cors-protocol
        https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/OPTIONS
*/

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	// CORSConfig defines the config for checking bearer tokens against an oauth2 idp.
	CORSConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		// AllowOriginsFn is an optional function that gets CORS allowed origins.
		AllowOriginsFn AllowOriginsFunc

		// AllowMethodsFn is an optional function that gets CORS allowed methods for a path.
		AllowMethodsFn AllowMethodsFunc

		// AllowOrigin defines CORS allowed origins when no AllowOriginsFn is defined.
		// Defaults to "*"
		AllowOrigins []string

		// AllowMethods defines CORS allowed methods when no AllowMethodsFn is defined.
		// Defaults to GET, HEAD, PUT, PATCH, POST, DELETE
		AllowMethods []string
	}

	// AllowOriginsFunc gets a TokeninfoResponse for a given access_token.
	AllowOriginsFunc func() ([]string, error)

	// AllowMethodsFunc gets the allowed methods for a path.
	AllowMethodsFunc func(path string) ([]string, error)

)

// Errors
var (
	//ErrTokenInvalid = echo.NewHTTPError(http.StatusUnauthorized, "Not allowed") TODO
)

var (
	// DefaultCORSConfig is a default config that allows anything.
	DefaultCORSConfig = CORSConfig{
		Skipper: middleware.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}
)

// OAuth2 returns an OAuth2 auth middleware.
func CORS() echo.MiddlewareFunc {
	c := DefaultCORSConfig
	return CORSWithConfig(c)
}

// CORSWithConfig returns a CORS middleware with config.
// See: `CORSConfig`.
func CORSWithConfig(config CORSConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultOauth2Config.Skipper
	}

	if config.AllowOriginsFn == nil {
		config.AllowOriginsFn = func() ([]string, error) {
			return config.AllowOrigins, nil
		}
	}

	if config.AllowMethodsFn == nil {
		config.AllowMethodsFn = func(path string) ([]string, error) {
			return config.AllowMethods, nil
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()

			// https://www.w3.org/TR/cors/#user-credentials
			supportsCredentials := req.Header.Get(echo.HeaderAuthorization) != ""
			// TODO replace with: supportsCredentials := if this resource requires scopes

			// https://www.w3.org/TR/cors/ section 6.1 and 6.2

			// 1. If the Origin header is not present terminate this set of steps. The request is outside the scope of
			// this specification.
			origin := req.Header.Get(echo.HeaderOrigin)
			if origin == "" {
				return next(c)
			}

			// 2. If the value of the Origin header is not a case-sensitive match for any of the values in list of origins,
			// do not set any additional headers and terminate this set of steps.
			ao, err := config.AllowOriginsFn()
			if err != nil {
				return err
			}
			var allowOrigin string
			for _, o := range ao {
				if o == "*" || o == origin {
					allowOrigin = o
					break
				}
			}
			if allowOrigin == "" {
				return next(c)
			}


			if req.Method == echo.OPTIONS {
				// 6.2 Preflight Request

				// 3. Let method be the value as result of parsing the Access-Control-Request-Method header.
				// If there is no Access-Control-Request-Method header or if parsing failed, do not set any additional
				// headers and terminate this set of steps. The request is outside the scope of this specification.
				method := req.Header.Get(echo.HeaderAccessControlRequestMethod)
				if method == "" {
					return c.NoContent(http.StatusNoContent) //TODO next(c) let upstream server handle OPTIONS?
				}

				// 4. Let header field-names be the values as result of parsing the Access-Control-Request-Headers headers.
				// If there are no Access-Control-Request-Headers headers let header field-names be the empty list.
				headerFieldNames := req.Header.Get(echo.HeaderAccessControlRequestHeaders)

				// 5. If method is not a case-sensitive match for any of the values in list of methods do not set any additional
				// headers and terminate this set of steps. Always matching is acceptable since the list of methods can be unbounded.
				allowMethods, err := config.AllowMethodsFn(req.URL.Path)
				if err != nil {
					return err
				}
				if !stringInSlice(method, allowMethods) {
					return c.NoContent(http.StatusNoContent) //TODO next(c) let upstream server handle OPTIONS?
				}

				// 6. If any of the header field-names is not a ASCII case-insensitive match for any of the values in list
				// of headers do not set any additional headers and terminate this set of steps.
				// Always matching is acceptable since the list of headers can be unbounded.
				// -> not implemented

				// 7. If the resource supports credentials add a single Access-Control-Allow-Origin header, with the value
				// of the Origin header as value, and add a single Access-Control-Allow-Credentials header with the case-sensitive
				// string "true" as value.
				// Otherwise, add a single Access-Control-Allow-Origin header, with either the value of the Origin header
				// or the string "*" as value. The string "*" cannot be used for a resource that supports credentials.
				res.Header().Set(echo.HeaderAccessControlAllowOrigin, origin)
				if supportsCredentials {
					res.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
				}

				// 8. Optionally add a single Access-Control-Max-Age header with as value the amount of seconds the user
				// agent is allowed to cache the result of the request.
				// -> not implemented

				// 9. If method is a simple method this step may be skipped.
				// Add one or more Access-Control-Allow-Methods headers consisting of (a subset of) the list of methods.
				// If a method is a simple method it does not need to be listed, but this is not prohibited.
				// Since the list of methods can be unbounded, simply returning the method indicated by Access-Control-Request-Method (if supported) can be enough.
				res.Header().Set(echo.HeaderAccessControlAllowMethods, method)

				// 10. If each of the header field-names is a simple header and none is Content-Type, this step may be skipped.
				// Add one or more Access-Control-Allow-Headers headers consisting of (a subset of) the list of headers.
				// If a header field name is a simple header and is not Content-Type, it is not required to be listed.
				// Content-Type is to be listed as only a subset of its values makes it qualify as simple header.
				// Since the list of headers can be unbounded, simply returning supported headers from Access-Control-Allow-Headers can be enough.
				b := headerFieldNames
				if !(b == "") {
					b = b + ","
				}
				b = b + "Content-Type" //TODO add more headers / make configurable
				res.Header().Set(echo.HeaderAccessControlAllowHeaders, b)

				// Tell caches to add Origin header to key.
				res.Header().Add(echo.HeaderVary, echo.HeaderOrigin)

				// Return status 200 (not 204) https://www.w3.org/Protocols/rfc2616/rfc2616-sec9.html
				return c.NoContent(http.StatusOK)
			}

			// 6.1 Simple Cross-Origin Request, Actual Request, and Redirects

			// 3. If the resource supports credentials add a single Access-Control-Allow-Origin header, with the value of the
			// Origin header as value, and add a single Access-Control-Allow-Credentials header with the case-sensitive string "true" as value.
			// Otherwise, add a single Access-Control-Allow-Origin header, with either the value of the Origin header or the string "*" as value.
			// The string "*" cannot be used for a resource that supports credentials.
			res.Header().Set(echo.HeaderAccessControlAllowOrigin, origin)
			if supportsCredentials {
				res.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
			}

			// 4. If the list of exposed headers is not empty add one or more Access-Control-Expose-Headers headers, with
			// as values the header field names given in the list of exposed headers.
			// By not adding the appropriate headers resource can also clear the preflight result cache of all entries where
			// origin is a case-sensitive match for the value of the Origin header and url is a case-sensitive match for th
			// URL of the resource.
			// -> not implemented

			// Tell caches to add Origin header to key.
			res.Header().Add(echo.HeaderVary, echo.HeaderOrigin)

			return next(c)
		}
	}
}

// A method is said to be a simple method if it is a case-sensitive match for one of the following: GET, HEAD, POST
func simpleMethod(m string) bool {
	// TODO
	return true
}

// A header is said to be a simple header if the header field name is an ASCII case-insensitive match for Accept, Accept-Language,
// or Content-Language or if it is an ASCII case-insensitive match for Content-Type and the header field value media type
// (excluding parameters) is an ASCII case-insensitive match for application/x-www-form-urlencoded, multipart/form-data, or text/plain.
func simpleHeader(k, v string) bool {
	// TODO
	return true
}

// StringInSlice return true if 'a' is in 'list'.
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}


