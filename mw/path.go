package mw

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

/*
	Path middleware allows one to check and/or change the request URL path.
*/

type (
	// PathConfig defines the config for Path middleware.
	// Checks like RequirePrefix are done prior to mutations like TrimPrefix.
	PathConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper
		// RequirePrefix defines how to path should start, non compliant paths result in 404 NotFound.
		RequirePrefix string
		// TrimPrefix defines what to remove from the front of the path.
		TrimPrefix string
	}
)

var (
	// DefaultPathConfig is the default Path middleware config.
	DefaultPathConfig = PathConfig{
		Skipper:    middleware.DefaultSkipper,
		TrimPrefix: "",
	}
)

// Path returns a Path middleware with default config.
func Path() echo.MiddlewareFunc {
	c := DefaultPathConfig
	return PathWithConfig(c)
}

// PathWithConfig returns a Path middleware with config.
func PathWithConfig(config PathConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultPathConfig.Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}

			url := c.Request().URL
			// Check
			if !strings.HasPrefix(url.Path, config.RequirePrefix) {
				return echo.NewHTTPError(http.StatusNotFound, "Not found")
			}
			// Mutate
			url.Path = strings.TrimPrefix(url.Path, config.TrimPrefix)

			return next(c)
		}
	}
}
