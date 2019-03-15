package mw

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestScopeValidation shows that the correct HTTP Status code is returned for a given combination of required scopes and
// allowed scopes.
// Note: a Bearer token is provided but its value isn't checked in the Tokeninfo function.
func TestScopeValidation(t *testing.T) {
	var tests = []struct {
		required []string
		allowed  map[string]struct{}
		want     int
		info     string
	}{
		{required: []string{}, allowed: map[string]struct{}{}, want: http.StatusOK, info: "public access"},
		{required: []string{}, allowed: map[string]struct{}{"read": {}}, want: http.StatusOK, info: "more allowed than required"},
		{required: []string{"read"}, allowed: map[string]struct{}{"read": {}}, want: http.StatusOK, info: "allowed == required (1 scope)"},
		{required: []string{"read", "write"}, allowed: map[string]struct{}{"write": {}, "read": {}}, want: http.StatusOK, info: "allowed == required (2 scopes)"},
		{required: []string{"read"}, allowed: map[string]struct{}{"write": {}}, want: http.StatusUnauthorized, info: "allowed != required (1 scope)"},
		{required: []string{"read"}, allowed: map[string]struct{}{}, want: http.StatusUnauthorized, info: "more required than allowed (1 scope)"},
	}

	for _, tst := range tests {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Add("Authorization", "Bearer value-not-important")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Configure middleware
		oauth2 := OAuth2WithConfig(OAuth2Config{
			RequiredScopes: func(method string, url *url.URL) ([]string, error) {
				return tst.required, nil
			},
			Tokeninfo: func(token string) (*TokeninfoResponse, error) {
				return &TokeninfoResponse{
					Scopes:    tst.allowed,
					ExpiresIn: 10,
				}, nil
			},
		})
		// Create chain of handlers
		h := oauth2(func(c echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		// Invoke handler and check result
		err := h(c)
		if err != nil {
			assert.Equal(t, tst.want, err.(*echo.HTTPError).Code, tst.info)
		} else {
			assert.Equal(t, tst.want, c.Response().Status, tst.info)
		}
	}
}

// TestTokenMissing shows that without a bearer token, only method/paths that have no required scopes are permitted.
// Note: no Bearer token is provided.
func TestTokenMissing(t *testing.T) {
	var tests = []struct {
		required []string
		allowed  map[string]struct{}
		want     int
		info     string
	}{
		{required: []string{}, allowed: map[string]struct{}{}, want: http.StatusOK, info: "public access"},
		{required: []string{}, allowed: map[string]struct{}{"read": {}}, want: http.StatusOK, info: "more allowed than required"},
		{required: []string{"read"}, allowed: map[string]struct{}{"read": {}}, want: http.StatusBadRequest, info: "allowed == required (1 scope)"},
		{required: []string{"read", "write"}, allowed: map[string]struct{}{"write": {}, "read": {}}, want: http.StatusBadRequest, info: "allowed == required (2 scopes)"},
		{required: []string{"read"}, allowed: map[string]struct{}{"write": {}}, want: http.StatusBadRequest, info: "allowed != required (1 scope)"},
		{required: []string{"read"}, allowed: map[string]struct{}{}, want: http.StatusBadRequest, info: "more required than allowed (1 scope)"},
	}

	for _, tst := range tests {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Configure middleware
		oauth2 := OAuth2WithConfig(OAuth2Config{
			RequiredScopes: func(method string, url *url.URL) ([]string, error) {
				return tst.required, nil
			},
			Tokeninfo: func(token string) (*TokeninfoResponse, error) {
				return &TokeninfoResponse{
					Scopes:    tst.allowed,
					ExpiresIn: 10,
				}, nil
			},
		})
		// Create chain of handlers
		h := oauth2(func(c echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		// Invoke handler and check result
		err := h(c)
		if err != nil {
			assert.Equal(t, tst.want, err.(*echo.HTTPError).Code, tst.info)
		} else {
			assert.Equal(t, tst.want, c.Response().Status, tst.info)
		}
	}
}

// TestTokenExpired shows that when token is expired only method/paths that have no required scopes are permitted.
// Note: a Bearer token is provided.
func TestTokenExpired(t *testing.T) {
	var tests = []struct {
		required []string
		allowed  map[string]struct{}
		want     int
		info     string
	}{
		{required: []string{}, allowed: map[string]struct{}{}, want: http.StatusOK, info: "public access"},
		{required: []string{"read"}, allowed: map[string]struct{}{"read": {}}, want: http.StatusUnauthorized, info: "allowed == required (1 scope)"},
	}

	for _, tst := range tests {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Add("Authorization", "Bearer value-not-important")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Configure middleware
		oauth2 := OAuth2WithConfig(OAuth2Config{
			RequiredScopes: func(method string, url *url.URL) ([]string, error) {
				return tst.required, nil
			},
			Tokeninfo: func(token string) (*TokeninfoResponse, error) {
				return &TokeninfoResponse{
					Scopes:    tst.allowed,
					ExpiresIn: 0, // it's expired
				}, nil
			},
		})
		// Create chain of handlers
		h := oauth2(func(c echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		// Invoke handler and check result
		err := h(c)
		if err != nil {
			assert.Equal(t, tst.want, err.(*echo.HTTPError).Code, tst.info)
		} else {
			assert.Equal(t, tst.want, c.Response().Status, tst.info)
		}
	}
}

// TestTokeninfoFail shows that when tokeninfo fails only method/paths that have no required scopes are permitted.
// Note: a Bearer token is provided.
func TestTokeninfoFail(t *testing.T) {
	var tests = []struct {
		required []string
		allowed  map[string]struct{}
		want     int
		info     string
	}{
		{required: []string{}, allowed: map[string]struct{}{}, want: http.StatusOK, info: "public access"},
		{required: []string{}, allowed: map[string]struct{}{"read": {}}, want: http.StatusOK, info: "more allowed than required"},
		{required: []string{"read"}, allowed: map[string]struct{}{"read": {}}, want: http.StatusUnauthorized, info: "allowed == required (1 scope)"},
		{required: []string{"read", "write"}, allowed: map[string]struct{}{"write": {}, "read": {}}, want: http.StatusUnauthorized, info: "allowed == required (2 scopes)"},
		{required: []string{"read"}, allowed: map[string]struct{}{"write": {}}, want: http.StatusUnauthorized, info: "allowed != required (1 scope)"},
		{required: []string{"read"}, allowed: map[string]struct{}{}, want: http.StatusUnauthorized, info: "more required than allowed (1 scope)"},
	}

	for _, tst := range tests {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Add("Authorization", "Bearer value-not-important")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Configure middleware
		oauth2 := OAuth2WithConfig(OAuth2Config{
			RequiredScopes: func(method string, url *url.URL) ([]string, error) {
				return tst.required, nil
			},
			Tokeninfo: func(token string) (*TokeninfoResponse, error) {
				return &TokeninfoResponse{
					Scopes:    tst.allowed,
					ExpiresIn: 10,
				}, echo.NewHTTPError(http.StatusInternalServerError, "Unknown error")
			},
		})
		// Create chain of handlers
		h := oauth2(func(c echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		// Invoke handler and check result
		err := h(c)
		if err != nil {
			assert.Equal(t, tst.want, err.(*echo.HTTPError).Code, tst.info)
		} else {
			assert.Equal(t, tst.want, c.Response().Status, tst.info)
		}
	}
}

// TestScopesFail shows that when Scopes fails no access is granted.
// Note: a Bearer token is provided.
func TestScopesFail(t *testing.T) {
	var tests = []struct {
		required []string
		allowed  map[string]struct{}
		want     int
		info     string
	}{
		{required: []string{}, allowed: map[string]struct{}{}, want: http.StatusNotFound, info: "public access"},
		{required: []string{}, allowed: map[string]struct{}{"read": {}}, want: http.StatusNotFound, info: "more allowed than required"},
		{required: []string{"read"}, allowed: map[string]struct{}{"read": {}}, want: http.StatusNotFound, info: "allowed == required (1 scope)"},
		{required: []string{"read", "write"}, allowed: map[string]struct{}{"write": {}, "read": {}}, want: http.StatusNotFound, info: "allowed == required (2 scopes)"},
		{required: []string{"read"}, allowed: map[string]struct{}{"write": {}}, want: http.StatusNotFound, info: "allowed != required (1 scope)"},
		{required: []string{"read"}, allowed: map[string]struct{}{}, want: http.StatusNotFound, info: "more required than allowed (1 scope)"},
	}

	for _, tst := range tests {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Add("Authorization", "Bearer value-not-important")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Configure middleware
		oauth2 := OAuth2WithConfig(OAuth2Config{
			RequiredScopes: func(method string, url *url.URL) ([]string, error) {
				return tst.required, echo.NewHTTPError(http.StatusNotFound, "No API")
			},
			Tokeninfo: func(token string) (*TokeninfoResponse, error) {
				return &TokeninfoResponse{
					Scopes:    tst.allowed,
					ExpiresIn: 10,
				}, nil
			},
		})
		// Create chain of handlers
		h := oauth2(func(c echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		// Invoke handler and check result
		err := h(c)
		if err != nil {
			assert.Equal(t, tst.want, err.(*echo.HTTPError).Code, tst.info)
		} else {
			assert.Equal(t, tst.want, c.Response().Status, tst.info)
		}
	}
}
