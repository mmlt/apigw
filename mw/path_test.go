package mw

import (
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
)

// TestPath validates RequirePrefix and TrimPrefix functionality.
func TestPath(t *testing.T) {
	var tests = []struct {
		requirePrefix string
		trimPrefix    string
		path          string
		wantCode      int
		want          string
		comment       string
	}{
		{"", "/prefix", "/prefix/users", 200, "/users", "1"},
		{"", "/api/v1", "/api/v1/doc/swagger.json", 200, "/doc/swagger.json", "2"},
		{"", "/nomatch", "/prefix/users", 200, "/prefix/users", "3"},
		{"", "/nomatch", "/prefix/users", 200, "/prefix/users", "4"},
		{"/api/v1", "", "/api/v1/doc/swagger.json", 200, "/api/v1/doc/swagger.json", "5"},
		{"/api/v1", "", "/doc/swagger.json", 404, "/api/v1/doc/swagger.json", "6"},
		{"/api/v1", "/api/v1", "/api/v1/doc/swagger.json", 200, "/doc/swagger.json", "7"},
	}
	for _, tst := range tests {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, tst.path, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Configure middleware
		path := PathWithConfig(PathConfig{
			RequirePrefix: tst.requirePrefix,
			TrimPrefix:    tst.trimPrefix,
		})
		// Create chain of handlers
		var rewrittenPath string
		h := path(func(c echo.Context) error {
			rewrittenPath = c.Request().URL.Path
			return c.String(http.StatusOK, "test")
		})
		// Invoke handler and check result
		err := h(c)
		if err != nil {
			assert.Equal(t, tst.wantCode, err.(*echo.HTTPError).Code, tst.comment)
		} else {
			assert.Equal(t, tst.wantCode, c.Response().Status, tst.comment)
			assert.Equal(t, tst.want, rewrittenPath, tst.comment)
		}
	}
}
