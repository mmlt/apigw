# TODO

- OpenAPI definition contains texts for 401 etc responses, use those text to override default (hardcoded) error texts. 
      
- create TLS testcases
- create TLS certs on each docker build (so they never expire)
- send logs to ELK
- update config.yaml 
  - with :9102 and remove prom-addrs flag
- dashboard with top clientid's by number of requests

- Move project to github.com/mmlt
- Review TODO FIXME
- Increase test coverage
  - mw/oauth2 garbarge collection
  - move part of main.go to apigw_test?
  - openapi/index.go contains some special cases that aren't tested
  - add -race
  - go-fuzz

- add stats on idp endpoint? (response times, errors, cache hits/misses)
- Add? Access-Control-Allow-Headers: authorization,Access-Control-Allow-Origin,Content-Type,SOAPAction
- PR for echo middleware/cors.go
			// Simple request
			if req.Method != echo.OPTIONS {
				res.Header().Add(echo.HeaderVary, echo.HeaderOrigin)
				res.Header().Set(echo.HeaderAccessControlAllowOrigin, allowOrigin)
				res.Header().Set(echo.HeaderAccessControlAllowMethods, allowMethods) // BUGFIX
				if config.AllowCredentials {
- PR for echo middleware/proxy.go
	func proxyHTTP(t *ProxyTarget) http.Handler {
    	//targetQuery := t.URL.RawQuery
    	director := func(req *http.Request) {
    		req.URL.Scheme = t.URL.Scheme
    		req.URL.Host = t.URL.Host
    		req.Host = t.URL.Host // https://github.com/golang/go/issues/5692
	(consider reimplementing as it also allows path rewriting that we don't need)		

- PR for echo echo.go
	// Execute chain
	if err := h(c); err != nil {
		//TODO fix double error responses.
		// e.HTTPErrorHandler(err, c)
		// The above line is removed to fix double error responses when LoggerWithConfig middleware is used.
		// To debug the issue define a custom echo.HTTPErrorHandler and set a breakpoint.
		//
		// Removing this line makes errors returned by 	e.Pre(mw.PathWithConfig(mw.PathConfig{ (ingress.go:62) disapear.
		//
		// Conclusion:
		// When Logger(WithConfig) is used the errors reponses are send by the logger. The above e.HTTPErrorHandler invokation
		// will repeat the same error message a 2nd time.
		// When e.Pre() handlers are used the logger won't handle those errors and the above HTTPErrorHandler is needed.
		// Work-a-round is to disable the above e.HTTPErrorHandler and move e.Pre() handler to e.Use() (=main pipeline)
		//
	}

- dep ensure update doesn't work (for now manually copied files to vendor)

- Add side car filebeat to send weblog to ELK
  - log file rotation?
  - mixed glog and web log output (glog rejected by logstash but visible in prom stats)
- What happens when upstream server goes down (under load)? 				
- Add testsvr check for correct Host header (Host header is used icw virtual servers)


# Versions
0.0.7 Rewrite CORS middleware according to https://www.w3.org/TR/cors


# JWT token validation

Example Python code to validate JWT token

```
import requests
import json
from jose import jwt
from amapi import AM

AM_URL = 'http://am.example.com:7080/am'

am = AM(AM_URL)

# fetch oauth2/.well-known/openid-configuration
well_known = am.well_known()
print(json.dumps(well_known.json(), indent=2))

jwks_uri = well_known.json()['jwks_uri']

keys = requests.get(jwks_uri)
print(json.dumps(keys.json(), indent=2))

resp = am.resource_owner('test1', 'test1', 'demo', 'changeit', ['email', 'openid'])

access_token = resp.json()['access_token']
id_token = resp.json()['id_token']
print("id_token=", id_token)

headers = jwt.get_unverified_headers(id_token)
print(headers)
kid = headers['kid']

signing_key = None
for key in keys.json()['keys']:
#    print(key)
    if key['kid'] == kid:
        signing_key = key
        break

# Verify the id_token JWT (what to verify: options)
options = {
    'verify_signature': True,
    'verify_aud': True,
    'verify_iat': True,
    'verify_exp': True,
    'verify_nbf': True,
    'verify_iss': True,
    'verify_sub': True,
    'verify_jti': True,
    'verify_at_hash': True,
    'leeway': 0
}

# algorithms should be checked to be in the union of
#   - supported algo's (see response)
#   - our minimal requirements ('none' should no be allowed)
claims = jwt.decode(id_token, signing_key, algorithms=['RS256'], audience='test1', access_token=access_token)
print("Claims: ", claims)

```

Library code

```
import json
import requests
import urllib

from requests.auth import HTTPBasicAuth

class AM:
    def __init__(self, amUrl, cookieName='iPlanetDirectoryPro'):
        self.amUrl = amUrl
        self.cookieName = cookieName

    def serverinfo(self, resource='*'):
        headers = {
            'Content-type' : 'application/json',
            'Accept-API-Version': 'protocol=1.0,resource=1.1'}
        return requests.get(self.amUrl+'/json/serverinfo/' + resource, headers=headers)

    def authenticate(self, username, password):
        headers = {
            'X-OpenAM-Username': username, 'X-OpenAM-Password' : password,
            'Accept-API-Version': 'resource=2.0, protocol=1.0'
        }
        return requests.post(self.amUrl+'/json/authenticate', data={}, headers=headers)

    def users(self, token, userId):
        headers = {
            self.cookieName: token
        }
        return requests.get(self.amUrl+'/json/users/'+userId, headers=headers)

    def queryUsers(self, token, queryFilter):
        headers = {
            'Accept-API-Version': 'resource=2.0, protocol=1.0',
            self.cookieName: token
        }
        # print (urllib.parse.quote(queryFilter))
        url = self.amUrl+'/json/users?_queryFilter=' + urllib.parse.quote(queryFilter)
        print(url)
        return requests.get(url, headers=headers)

    def queryPolicies(self, token, queryFilter, fields='', sortKeys=''):
        headers = {
            'Accept-API-Version': 'resource=2.1',
            self.cookieName: token
        }
        url = self.amUrl+'/json/policies?_queryFilter={}'.format(urllib.parse.quote(queryFilter))

        if fields:
            url = url + "&_fields={}".format(fields)
        if sortKeys:
            url = url + "&_sortKeys={}".format(sortKeys)
        
        return requests.get(url, headers=headers)
    
    def logout(self, token):
        headers = {
            'Accept-API-Version': 'resource=2.1',
            self.cookieName: token
        }
        return requests.post(self.amUrl + '/json/sessions?_action=logout', headers=headers)

# OAuth2
    def well_known(self):
        return requests.get(self.amUrl + '/oauth2/.well-known/openid-configuration')

# curl \
# --request POST \
# --user "client_id:client_secret" \
# --data "grant_type=password&username=amadmin&password=cangetinam&scope=profile" \
# http://am.example.com:7080/am/oauth2/access_token
    def resource_owner(self, clientId, clientSecret, username, password, scopes):
        headers = {
            'Content-Type' : 'application/x-www-form-urlencoded'
        }
        data = {
            'grant_type' : 'password',
            'username' : username,
            'password' : password,
            'scope' : ' '.join(scopes)
        }
        return requests.post(self.amUrl + '/oauth2/access_token',
            auth=HTTPBasicAuth(clientId, clientSecret),
            headers=headers, data=data)

    # FR Common REST way
    def crUsers(self, token, userId):
        headers = {
            'Accept-API-Version': 'resource=2.0, protocol=1.0',
            self.cookieName: token
        }
        # return requests.get(amUrl+'/json/users/'+userId, headers=headers)
        return requests.post(self.amUrl+'/json/users/'+userId+'?_action=read', headers=headers)4

```
