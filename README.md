# API Gateway
API Gateway (APIGW) is a reverse proxy that implements authorization based on OAuth2 scopes and a Swagger API definition.

It's typical use case is to provide authorization for a REST service.
```
                +--------------+
                |              |
                |  OAuth2 IDP  |
                |              |
                +------+-------+
                       ^
                       | tokeninfo
                       |
                +------+-------+        +----------------+
                |              |        |                |
  Request ----->+  APIGW       +------->+  REST service  |
                |              |        |                |
                +--------------+        +----------------+
```

The APIGW will only pass a request to the upstream service when all of the following conditions are met:

- the requested path is in the swagger api definition
- if the swagger definition does not contain oauth2 scopes for the requested path:
  - the request is passed 
- if the swagger definition does contain oauth2 scopes for the requested path:
  - the request contains a valid OAuth2 access_token
  - the tokeninfo returned scopes are a superset of the swagger defined scopes 


## Features
- TSL (or not) 
- Load balancing over upstream services. 
- Swagger definitions are read from upstream server(s) on start-up (and periodically checked for updates).
- Configurable CORS headers (by default Access-Control-Allow-Methods are read from OpenAPI endpoint definitions).
- Configurable error responses.
- Caching of tokeninfo responses to reduces load on tokeninfo endpoint.

- IIS Compatible log file
  - user field contains ClientID
- Prometheus stats
  - Histogram of handling time of successful requests - by Method
  - Counter of fully handled request - by ClientID, Status

- Simplicity; APIGW protects one Swagger defined API (for multiple API's use multiple instances icw L7 path routing).
- Unit and e2e tests to validate behavior (see coverage report)


## Roadmap
- Add API level throttling.
- Add client specific throttling 
- Fuzzing of request urls.
- Add Swagger parameter validation (currently a {parameter} matches any type). Swagger definition contains [parameters specs](https://swagger.io/docs/specification/2-0/describing-parameters/) 
  - unit test that checks parameter types
  - Use go-swagger instead of openapi package, see func (c *Context) RouteInfo(request *http.Request) (*MatchedRoute, *http.Request, bool)
  - [how-to-serve-two-or-more-swagger-specs-from-one-server](https://github.com/go-swagger/go-swagger/blob/master/docs/faq/faq_server.md#how-to-serve-two-or-more-swagger-specs-from-one-server)
- Create /health endpoint (on management port) that checks internal state
- Improve configuration
  - Dynamic configuration of CORS Access-Control-Allow-Origin header. For example a ConfigMap with allowed origins that is easy to reload.
  - Create a /reload endpoint (on management port) to dynamically reload config when the ConfigMap is updated. 

## Architecture
[echo](https://echo.labstack.com/) middleware:
- error handler with custom messages
- path rewriting
- CORS header handling
- oauth access token + scope validation see OAuth2 RFC at https://tools.ietf.org/html/rfc6749
- reverse proxy with load balancing


## Run
Apigw is configured by yaml file, see `configYaml` in apigw_test.go for an example.

Assuming an config_http.yaml in the CWD do `apigw -v=2 --logtostderr --config=config_http.yaml` to get basic logging at stderr.



## Development
Prereqs: GO 1.12 or later, GOBIN in path for tooling.


### Run tests
See Makefile on how to run tests.

#### e2e tests (in process)
`apigw_test.go` runs e2e tests with all servers in a single process. 
This is great for testing new functionality but not so great for performance testing/tuning.
TODO write e2e tests were all servers have their own golang runtime environment:
```
	ctx, cancel := context.Background()
	defer cancel()
	if err := exec.CommandContext(ctx, "go", "run", "upstreamsvr.go t1 localhost:1323").Run(); err != nil {
        //handle error
	}
```

#### Manual tests

Get Prometheus metrics:
`curl localhost:9102/metrics`

Get API version (no auth required):
`curl -v localhost:8080/api/v1/version`

Get accounts must fail because auth is required:
`curl -v localhost:8080/api/v1/accounts` -> 400

Login via Examplesite to obtain a bearer token, then get accounts:
`curl -v -H "Authorization: Bearer 2dd46d2e-5711-4608-8722-79b36b60f851" localhost:8080/api/v1/accounts` 


### Known issues




