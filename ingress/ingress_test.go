package ingress

import (
	"errors"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

func TestCustomHTTPErrorHandler(t *testing.T) {
	var tests = []struct {
		err        error
		template   string
		wantStatus int
		wantBody   string
	}{
		{errors.New("Plain text"), "{{.Status}} {{.Message}}", 500, "500 Plain text"},
		{echo.NewHTTPError(401, "Not allowed"), "{{.Status}} {{.Message}}", 401, "401 Not allowed"},
		{echo.NewHTTPError(404, "Not found"), "", 404, ""},
		{echo.NewHTTPError(404, "Not found"), "{{.ThisNameIsNotDefined}}", 404, "{\"Status\":404,\"Message\":\"Not found\"}"},
	}

	e := echo.New()

	for i, test := range tests {
		f, err := customHTTPErrorHandler(test.template)
		if assert.NoError(t, err) {
			w := httptest.NewRecorder()
			c := e.NewContext(httptest.NewRequest("GET", "http://example.com/foo", nil), w)
			f(test.err, c)
			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			assert.EqualValues(t, test.wantStatus, resp.StatusCode, "%d) status", i)
			assert.EqualValues(t, test.wantBody, string(body), "%d) body", i)
			//TODO test resp.Header.Get("Content-Type")
		}
	}
}
