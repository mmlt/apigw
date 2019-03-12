// Svr is used when testing Apigw.
package server

import (
	"context"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net"

	"fmt"
	"golang.org/x/net/websocket"
	"net/http"
	"time"
)

// Testsvr
type Testsvr struct {
	Name string
	Port string
	Echo *echo.Echo
}

// IDP creates an IDP mock server.
// Paths:
//	/oauth2/tokeninfo - OAuth2 tokeninfo endpoint with hardcoded token-response pairs.
func IDP(name, port string) *Testsvr {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.GET("/oauth2/tokeninfo", func(c echo.Context) error {
		token := c.QueryParam("access_token")
		if token == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Missing access_token")
		}

		// Hardcoded tokeninfo responses
		var scope string
		var expiresIn int = 20
		switch token {
		case "readabcdef":
			scope = `"read"`
		case "writeabcdef":
			scope = `"read", "write"`
		case "expiredabcdef":
			scope = `"read", "write"`
			expiresIn = 0
		case "00ef62b1-5dac-4725-bf10-ab6ebfa8ffc0":
			// standard response
		default:
			return c.JSONBlob(http.StatusUnauthorized, []byte(`
				"error": "invalid_token",
				"error_description": "The access token provided is expired, revoked, malformed or invalid for other reasons."
			`))
		}
		s := fmt.Sprintf(`{
		  "auth_level": 1,
		  "session_id": "bQLjd6hHfaTTdbv3lBST0vMLslg.*AAJTSQACMDIAAlNLABxGYVFUNlc0K2ZnWERLVnpHR1JGR3cyYXh2eXc9AAJTMQACMDE.*",
		  "token_type": "Bearer",
		  "client_id": "JJqMvO_m5zJ9odxE0iCXeOVGW2oa",
		  "access_token": "%s",
		  "grant_type": "authorization_code",
		  "scope": [ %s ],
		  "expires_in": %d
		}`, token, scope, expiresIn)

		return c.JSONBlob(http.StatusOK, []byte(s))
	})

	return &Testsvr{Name: name, Port: port, Echo: e}
}

// Multipurpose creates a multipurpose testsvr.
// Paths:
//	/ - a html page with websocket event demo
//	/public, /read, /write - echo path and server name
//	/v1/doc/swagger.json - api definition for /read and /write paths
//	/ws - websocket server that sends 1 msg/sec
func Multipurpose(name, port string) *Testsvr {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, fmt.Sprintf(index, name, port))
	})
	e.GET("/public", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "public "+name)
	})
	e.GET("/read", func(c echo.Context) error {
		fmt.Println()
		fmt.Println("/read")

		return c.HTML(http.StatusOK, "read "+name)
	})
	e.GET("/write", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "write "+name)
	})
	e.GET("/v1/doc/swagger.json", func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, []byte(swagger))
	})

	// WebSocket handler
	e.GET("/ws", func(c echo.Context) error {
		websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()
			for {
				// Write
				err := websocket.Message.Send(ws, fmt.Sprintf("Hello from upstream server %s!", name))
				if err != nil {
					e.Logger.Error(err)
				}
				time.Sleep(1 * time.Second)
			}
		}).ServeHTTP(c.Response(), c.Request())
		return nil
	})

	return &Testsvr{Name: name, Port: port, Echo: e}
}

// Run the server.
func (t *Testsvr) Run() error {
	return t.Echo.Start(t.Port)
}

// WaitForReadiness waits for the server to respond.
func (t *Testsvr) WaitForReadiness() error {
	conn, err := net.DialTimeout("tcp", t.Port, time.Second)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

// Shutdown stops server the gracefully.
func (t *Testsvr) Shutdown(ctx context.Context) error {
	return t.Echo.Shutdown(ctx)
}

// Shutdown attempts to stop server the gracefully but waits no more then the specified time for connections to close.
func (t *Testsvr) ShutdownWithTimeout(to time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()
	t.Shutdown(ctx)
}

// Index page.
var index = `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<meta http-equiv="X-UA-Compatible" content="ie=edge">
		<title>Upstream Server</title>
		<style>
			h1, p {
				font-weight: 300;
			}
		</style>
	</head>
	<body>
		<h1>HTTP</h1>
		<p>
			Hello from upstream server %s
		</p>
		<h1>WebSocket</h1>
		<p id="output"></p>
		<script>
			var ws = new WebSocket('ws://%s/ws')

			ws.onmessage = function(evt) {
				var out = document.getElementById('output');
				out.innerHTML += evt.data + '<br>';
			}
		</script>
	</body>
	</html
`

// Swagger definition for this server.
//TODO remove x-scope??
var swagger = `
{
  "swagger": "2.0",
  "info": { "version": "v1", "title": "MyBank.OpenApi", "description": "MyBank Open API test." },
  "host": "api.mybank.com",
  "schemes": ["http"],
  "paths": {
    "/read": {
      "get": {
        "tags": ["read"],
        "summary": "get a test page that requires 'read' scope.",
        "description": "",
        "operationId": "RD",
        "consumes": [],
        "produces": ["application/json","text/json"],
        "deprecated": false,
        "security": [{"oauth2": ["read"]}],
        "x-scope": "read"
      }
    },
    "/write": {
      "get": {
        "tags": ["write"],
        "summary": "get a test page that requires 'write' scope.",
        "description": "",
        "operationId": "WR",
        "consumes": [],
        "produces": ["application/json", "text/json"],
        "deprecated": false,
        "security": [{"oauth2": ["write"]}],
        "x-scope": "read"
      },
      "put": {
        "tags": ["write"],
        "summary": "put a test page that requires 'write' scope.",
        "description": "",
        "operationId": "WR",
        "consumes": [],
        "produces": ["application/json", "text/json"],
        "deprecated": false,
        "security": [{"oauth2": ["write"]}]
      }
    },
    "/public": {
      "get": {
        "tags": ["public"],
        "summary": "get a public test page.",
        "description": "",
        "operationId": "WR",
        "consumes": [],
        "produces": ["application/json", "text/json"],
        "deprecated": false,
        "security": [{"oauth2": [""]}],
        "x-scope": ""
      }
    }
  }
}
`
/*

*/