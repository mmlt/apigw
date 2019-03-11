// Run an ingress e2e test with multiple in-process servers.
package main

import (
	"fmt"

	"flag"
	"github.com/mmlt/apigw/gateway"
	"github.com/mmlt/testsvr/testsvr"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	ingressPort   = "localhost:13230"
	upstreamPort1 = "localhost:13231"
	upstreamPort2 = "localhost:13232"
	oauth2idpPort = "localhost:13233"
)

// configYaml is a yaml for testing the Apigw.
var configYaml = fmt.Sprintf(`
# ingress handles incomming API traffic.
ingress:
  # ip:port to bind.
  bind: %s
  #tls:
  #  key:
  #  cert:
  middleware:
    # path rewriting
    path:
      requirePrefix: /api/v1
      trimPrefix: /api/v1
    # CORS headers
    cors:
      allowOrigins: ["*"]
      allowMethods: ["GET", "POST", "DELETE"]
    #
    proxy:
      # url(s) of upstream servers
      targets: ["http://%s", "http://%s"]
  errorResponse: |
    { "developerMessage":"{{.Message}}", "endUserMessage":"", "errorCode":"{{.Message}}", "errorId":{{.Status}} }
   
# management handles incomming /metrics /health etc. traffic.
management:
  # ip:port to bind.
  bind: ":9102"

# openapi gets the API definition.
openapi:
  # url that serves the swagger.json API definition.
  url: http://%s/v1/doc/swagger.json

# oauth2idp gets the tokeninfo data.
oauth2idp:
  # url of tokeninfo
  tokeninfoUrl: http://%s/oauth2/tokeninfo
`, ingressPort, upstreamPort1, upstreamPort2, upstreamPort1, oauth2idpPort)

var (
	gw        *gateway.gateway
	upstream1 *testsvr.Testsvr
	upstream2 *testsvr.Testsvr
	idp       *testsvr.Testsvr
)

func Setup() {
	// Create upstream servers.
	upstream1 = testsvr.Multipurpose("upstream1", upstreamPort1)
	go func() {
		if err := upstream1.Run(); err != nil {
			fmt.Println(upstream1.Name, "Run", err)
		}
	}()

	upstream2 = testsvr.Multipurpose("upstream2", upstreamPort2)
	go func() {
		if err := upstream2.Run(); err != nil {
			fmt.Println(upstream2.Name, "Run", err)
		}
	}()

	// Create an oauth2 tokeninfo server.
	idp = testsvr.IDP("idp", oauth2idpPort)
	go func() {
		if err := idp.Run(); err != nil {
			fmt.Println(idp.Name, "Run", err)
		}
	}()

	// wait for servers to start
	// TODO change to http get check?
	time.Sleep(100 * time.Millisecond)

	// Create Gateway.
	// parse config
	cfg := &gateway.Config{}
	err := yaml.Unmarshal([]byte(configYaml), cfg)
	if err != nil {
		fmt.Println("Yaml", err)
	}
	// create server
	gw = gateway.NewWithConfig(cfg)
	go func() {
		if err := gw.Run(); err != nil {
			fmt.Println("gateway Run", err)
		}
	}()

	// wait for servers to start
	// TODO change to http get check as start time varies a lot (linux <100mS, Windows <500mS)?
	time.Sleep(500 * time.Millisecond)
}

func Teardown() {
	if upstream1 != nil {
		upstream1.ShutdownWithTimeout(10 * time.Second)
	}

	if upstream2 != nil {
		upstream2.ShutdownWithTimeout(10 * time.Second)
	}

	if idp != nil {
		idp.ShutdownWithTimeout(10 * time.Second)
	}

	if gw != nil {
		gw.ShutdownWithTimeout(10 * time.Second)
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	Setup()
	code := m.Run()
	Teardown()
	os.Exit(code)
}

// TestUpstreamHttpGet shows that all upstream API servers are responding to HTTP GET.
func TestUpstreamHttpGet(t *testing.T) {
	tests := []struct {
		port string
	}{
		{port: upstreamPort1},
		{port: upstreamPort2},
	}

	for _, v := range tests {
		resp, err := http.Get("http://" + v.port)
		if err != nil {
			t.Error(v.port, "http get", err)
		}
		want := http.StatusOK
		if resp.StatusCode != want {
			t.Error(v.port, "http get expected status %d, got", want, resp.StatusCode)
		}
		resp.Body.Close()
	}
}

// TestOAuth2IDPHttpGet shows that IDP server is responding to HTTP GET.
func TestOAuth2IDPHttpGet(t *testing.T) {
	resp, err := http.Get("http://" + oauth2idpPort + "/oauth2/tokeninfo")
	if err != nil {
		t.Error(oauth2idpPort, "http get", err)
	}
	want := http.StatusBadRequest // because we didn't provide the access_token parameter.
	if resp.StatusCode != want {
		t.Error(oauth2idpPort, "http get expected status %d, got", want, resp.StatusCode)
	}
	resp.Body.Close()
}

// TestGwHttpGet shows that:
// - proper status codes are returned for paths that exist, do not exist, require authentication.
// - multiple HTTP Get's on the gateway load balances over the upstream servers.
func TestGwHttpGet(t *testing.T) {
	tests := []struct {
		url     string
		want    int
		comment string
	}{
		{"http://" + ingressPort + "/api/v1/public", http.StatusOK, "response by upstream server"},
		{"http://" + ingressPort + "/doesnotexist", http.StatusNotFound, "404 by gateway"},
		{"http://" + ingressPort + "/api/v1/doesnotexist", http.StatusNotFound, "404 by upstream server"},
		{"http://" + ingressPort + "/read", http.StatusNotFound, "404 by gateway for /read"},
		{"http://" + ingressPort + "/api/v1/read", http.StatusBadRequest, "requires Authorization header"},
	}

	serverNames := []string{"upstream1", "upstream2"}

	for _, tst := range tests {
		// do 4 GET's so each upstream server gets 2 requests...
		for i := 0; i < 4; i++ {
			resp, err := http.Get(tst.url)
			if err != nil {
				t.Error("TestGwHttpGet", tst.comment, err)
			}
			assert.Equal(t, tst.want, resp.StatusCode, tst.comment)
			body, _ := ioutil.ReadAll(resp.Body)
			//fmt.Println(string(body)) //!!!TODO
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				// the body text ends with the server name
				w := strings.Split(string(body), " ")
				name := w[len(w)-1]
				// because we do round robin loadbalancing we know what name to expect
				assert.Equal(t, serverNames[i%len(serverNames)], name, tst.comment)
			}
		}
	}
}

// TestErrorResponse show that proper JSON error messages are returned.
func TestErrorResponse(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"http://" + ingressPort + "/doesnotexist", `{ "developerMessage":"Not found", "endUserMessage":"", "errorCode":"Not found", "errorId":404 }`},
		{"http://" + ingressPort + "/api/v1/read", `{ "developerMessage":"Missing or malformed token", "endUserMessage":"", "errorCode":"Missing or malformed token", "errorId":400 }`},
	}

	for _, tst := range tests {
		resp, err := http.Get(tst.url)
		if err != nil {
			t.Error("TestGwHttpGet", err)
		}
		body, _ := ioutil.ReadAll(resp.Body)
		got := strings.TrimSuffix(string(body), "\n")
		assert.Equal(t, tst.want, got)
		resp.Body.Close()
	}
}

func TestErrorResponseInvalidBearer(t *testing.T) {
	// call with invalid bearer token
	url := "http://" + ingressPort + "/api/v1/read"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("Authorization", "Bearer invalidtoken")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}
	//want := `{"Code":401,"DeveloperMessage":"Not allowed","UserMessage":""}`
	want := `{ "developerMessage":"Not allowed", "endUserMessage":"", "errorCode":"Not allowed", "errorId":401 }`
	body, _ := ioutil.ReadAll(resp.Body)
	got := strings.TrimSuffix(string(body), "\n")
	assert.Equal(t, want, got)
	resp.Body.Close()
}

// TestReadScope shows that a read scope is required for /read endpoint.
// Various bearer token problems are tested.
func TestReadScope(t *testing.T) {
	url := "http://" + ingressPort + "/api/v1/read"
	// call without bearer token returns Unauthorized
	resp, err := http.Get(url)
	if err != nil {
		t.Error(err)
	}
	want := http.StatusBadRequest
	if resp.StatusCode != want {
		t.Errorf(" without bearer token; expected %d, got %d", want, resp.StatusCode)
	}
	resp.Body.Close()

	// call with invalid bearer token should return Unauthorized
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("Authorization", "Bearer invalidtoken")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}
	want = http.StatusUnauthorized
	if resp.StatusCode != want {
		t.Errorf(" with invalid bearer token; expected %d, got %d", want, resp.StatusCode)
	}
	resp.Body.Close()

	// call with valid bearer token returns OK
	req, _ = http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("Authorization", "Bearer readabcdef") // defined in testoauth2idp
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}
	want = http.StatusOK
	if resp.StatusCode != want {
		t.Errorf(" with valid bearer token; expected %d, got %d", want, resp.StatusCode)
	}
	resp.Body.Close()

	// call with expired bearer token should return Unauthorized
	req, _ = http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("Authorization", "Bearer expiredabcdef")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}
	want = http.StatusUnauthorized
	if resp.StatusCode != want {
		t.Errorf(" with expired bearer token; expected %d, got %d", want, resp.StatusCode)
	}
	resp.Body.Close()
}

func TestHeaders(t *testing.T) {
	tests := []struct {
		id			 string // some random string to match test data with test error messages.
		method       string
		path         string
		headers      map[string]string
		wantStatus   int
		wantHeaders  map[string]string // required headers
		extraHeaders map[string]string // optional headers
	}{
		// Access /api/v1/read with Bearer token.
		// CORS preflight request
		{"th01", http.MethodOptions, "/api/v1/read",
			map[string]string{
				"Origin": "localhost",
				"Authorization": "Bearer readabcdef",
				"Access-Control-Request-Method": "GET",
				"Access-Control-Request-Headers": "authorization",
			},
			http.StatusOK,
			map[string]string{
				"Content-Length": "0",
				"Access-Control-Allow-Origin": "localhost",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Allow-Methods": "GET",
				"Access-Control-Allow-Headers": "authorization,Content-Type",
				"Vary": "Origin",
			},
			map[string]string{
				"Content-Type": "",
				"Date": "",
				"Access-Control-Max-Age": "",
			},
		},
		// CORS request
		{"th02",http.MethodGet, "/api/v1/read",
			map[string]string{
				"Origin": "localhost",
				"Authorization": "Bearer readabcdef",
			},
			http.StatusOK,
			map[string]string{
				"Access-Control-Allow-Origin": "localhost",
				"Access-Control-Allow-Credentials": "true",
				"Vary": "Origin",
			},
			map[string]string{
				"Content-Length": "",
				"Content-Type": "",
				"Date": "",
			},
		},

		// Access /api/v1/write with Bearer token.
		// CORS preflight request
		{"th03",http.MethodOptions, "/api/v1/write",
			map[string]string{
				"Origin": "localhost",
				"Authorization": "Bearer writeabcdef",
				"Access-Control-Request-Method": "PUT",
				"Access-Control-Request-Headers": "authorization",
			},
			http.StatusOK,
			map[string]string{
				"Content-Length": "0",
				"Access-Control-Allow-Origin": "localhost",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Allow-Methods": "PUT",
				"Access-Control-Allow-Headers": "authorization,Content-Type",
				"Vary": "Origin",
			},
			map[string]string{
				"Content-Type": "",
				"Date": "",
				"Access-Control-Max-Age": "",
			},
		},
		// CORS request
		{"th04",http.MethodGet, "/api/v1/write",
			map[string]string{
				"Origin": "localhost",
				"Authorization": "Bearer writeabcdef",
			},
			http.StatusOK,
			map[string]string{
				"Access-Control-Allow-Origin": "localhost",
				"Access-Control-Allow-Credentials": "true",
				"Vary": "Origin",
			},
			map[string]string{
				"Content-Length": "",
				"Content-Type": "",
				"Date": "",
			},
		},

		// Access a non-existing resource with Bearer token.
		// CORS preflight request
		{"th05",http.MethodOptions, "/api/v1/doesnotexist",
			map[string]string{
				"Origin": "localhost",
				"Authorization": "Bearer readabcdef",
				"Access-Control-Request-Method": "GET",
				"Access-Control-Request-Headers": "authorization",
			},
			http.StatusNotFound,
			map[string]string{
			},
			map[string]string{
				"Content-Length": "",
				"Content-Type": "",
				"Date": "",
			},
		},
		// CORS request
		{"th06",http.MethodGet, "/api/v1/doesnotexist",
			map[string]string{
				"Origin": "localhost",
				"Authorization": "Bearer readabcdef",
			},
			http.StatusNotFound,
			map[string]string{
				"Access-Control-Allow-Origin": "localhost",
				"Access-Control-Allow-Credentials": "true",
				"Vary": "Origin",
			},
			map[string]string{
				"Content-Length": "",
				"Content-Type": "",
				"Date": "",
			},
		},

		// Access a public resource without Bearer token.
		// CORS preflight request
		{"th07",http.MethodOptions, "/api/v1/public",
			map[string]string{
				"Origin": "localhost",
				"Access-Control-Request-Method": "GET",
			},
			http.StatusOK,
			map[string]string{
				"Content-Length": "0",
				"Access-Control-Allow-Origin": "localhost",
				"Access-Control-Allow-Methods": "GET",
				"Access-Control-Allow-Headers": "Content-Type",
				"Vary": "Origin",
			},
			map[string]string{
				"Content-Type": "",
				"Date": "",
				"Access-Control-Max-Age": "",
			},
		},
		// CORS request
		{"th08",http.MethodGet, "/api/v1/public",
			map[string]string{
				"Origin": "localhost",
			},
			http.StatusOK,
			map[string]string{
				"Access-Control-Allow-Origin": "localhost",
				"Vary": "Origin",
			},
			map[string]string{
				"Content-Length": "",
				"Content-Type": "",
				"Date": "",
			},
		},

		//TODO Check OPTION without Origin header
		//TODO Do a preflight with a method that's not allowed, for example PUT on api/v1/read resource
	}

	for _, tst := range tests {
		// prepare request
		req, _ := http.NewRequest(tst.method, "http://"+ingressPort+tst.path, nil)
		for k, v := range tst.headers {
			req.Header.Add(k, v)
		}
		// do request
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, tst.wantStatus, resp.StatusCode, "%s) status code", tst.id)

		// assert wanted headers are present
		for k, v := range tst.wantHeaders {
			if assert.NotEmpty(t, resp.Header[k], "%s) response %s header missing.", tst.id, k) {
				got := resp.Header[k][0]
				assert.Equalf(t, v, got, "%s) response %s header wrong values.", tst.id, k)
			}
		}

		// assert no other headers than wantedHeaders and extraHeaders are in the response
		for k, v := range resp.Header {
			if _, ok := tst.wantHeaders[k]; ok {
				continue
			}
			if _, ok := tst.extraHeaders[k]; ok {
				continue
			}
			t.Errorf("%s) response %s header unexpected (value: %s)", tst.id, k, v)
		}

		resp.Body.Close()
	}
}