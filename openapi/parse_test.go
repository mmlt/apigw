package openapi

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestSpecFromRaw shows we can read a swagger.json and access elements.
func TestSpecFromRaw(t *testing.T) {
	spec, err := SpecFromRaw([]byte(swagger))
	if err != nil {
		t.Error(err)
	}

	expTitle := "MyBank.OpenApi"
	gotTitle := spec.Info.Title
	if expTitle != gotTitle {
		t.Errorf("title expected %s, got %s", expTitle, gotTitle)
	}

	p := spec.Paths.Paths
	expPaths := 17
	gotPaths := len(p)
	if expPaths != gotPaths {
		t.Errorf("title expected %d, got %d", expPaths, gotPaths)
	}
}

// TestSpecOAuth2ScopeIter show that we can collect path, operation and oauth2 scopes from a swagger spec.
func TestSpecOAuth2ScopeIter(t *testing.T) {
	spec, err := SpecFromRaw([]byte(swagger))
	if err != nil {
		t.Error(err)
	}

	got := make(map[string]string, 20)
	SpecOAuth2ScopeIter(spec, func(path string, action string, scopes []string) {
		//fmt.Printf("expect[\"%s\"] = \"%s%v\"\n", path, action, scopes)
		got[path] = fmt.Sprint(action, scopes)
	})

	expect := make(map[string]string, 20)
	expect["/session"] = "DELETE[read]"
	expect["/accounts/{accountNumber}"] = "GET[read]"
	expect["/version"] = "GET[]"
	expect["/accounts/{accountNumber}/balances"] = "GET[read]"
	expect["/accounts/{accountNumber}/positions"] = "GET[read]"
	expect["/instruments"] = "GET[read]"
	expect["/instruments/{id}"] = "GET[read]"
	expect["/accounts/{accountNumber}/transactions"] = "GET[read]"
	expect["/instruments/lists/{id}"] = "GET[read]"
	expect["/accounts/{accountNumber}/positions/{id}"] = "GET[read]"
	expect["/accounts/{accountNumber}/orders"] = "GET[read]"
	expect["/accounts/{accountNumber}/orders"] = "POST[write]"
	expect["/accounts/{accountNumber}/performances"] = "GET[read]"
	expect["/instruments/lists"] = "GET[read]"
	expect["/accounts/{accountNumber}/orders/{number}"] = "GET[read]"
	expect["/accounts/{accountNumber}/orders/{number}"] = "DELETE[write]"
	expect["/accounts"] = "GET[read]"
	expect["/accounts/{accountNumber}/orders/preview"] = "POST[write]"
	expect["/instruments/derivatives"] = "GET[read]"

	if len(expect) != len(got) {
		t.Errorf("expected len %d, got %d", len(expect), len(got))
	}

	for path, g := range got {
		e, ok := expect[path]
		if !ok {
			t.Errorf("unexpected %s", path)
		}
		if e != g {
			t.Errorf("expected %s %s, got %s", path, e, g)
		}
	}
}

// TestNewIndexFromSpec shows that we can retrieve scopes from a swagger spec via an index for fast lookup.
func TestNewIndexFromSpec(t *testing.T) {
	// Read spec from Swagger text.
	spec, err := SpecFromRaw([]byte(swagger))
	if err != nil {
		t.Error(err)
	}
	// Build an index.
	idx, err := newIndexFromSpec(spec)
	if err != nil {
		t.Error(err)
	}
	// Test
	tests := []struct {
		path   string
		method string
		scopes []string
	}{
		{"/version", "GET", []string{}},
		{"/session", "DELETE", []string{"read"}},
		{"/accounts/123", "GET", []string{"read"}},
		{"/accounts/123/balances", "GET", []string{"read"}},
		{"/accounts/123/orders/456", "DELETE", []string{"write"}},
		{"/accounts/123/orders/preview", "POST", []string{"write"}},
	}
	for _, tst := range tests {
		got, err := idx.FindScopes(tst.method, tst.path)
		assert.NoError(t, err)
		assert.ElementsMatch(t, tst.scopes, got, "%s %s", tst.method, tst.path)
	}
}

// Swagger is a string with a swagger json API specification.
var swagger = `
{
  "swagger": "2.0",
  "info": {
    "version": "v1",
    "title": "MyBank.OpenApi",
    "description": "MyBank Open API is a Restful API Platform to access MyBank's trading services."
  },
  "host": "api.mybank.com",
  "schemes": [
    "http"
  ],
  "paths": {
    "/accounts": {
      "get": {
        "tags": [
          "Accounts"
        ],
        "summary": "Gets all the accounts of a MyBank customer.",
        "description": "If there is no account, the collection will be empty.",
        "operationId": "Accounts_GetAccounts",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "responses": {
          "200": {
            "description": "[OK] A list of the accounts of a MyBank customer.",
            "schema": {
              "$ref": "#/definitions/AccountsResponse"
            }
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "read"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "read"
      }
    },
    "/accounts/{accountNumber}": {
      "get": {
        "tags": [
          "Accounts"
        ],
        "summary": "Gets the specific account of a MyBank customer.",
        "operationId": "Accounts_GetAccount",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "parameters": [
          {
            "name": "accountNumber",
            "in": "path",
            "description": "Requested MyBank account number.",
            "required": true,
            "type": "string"
          }
        ],
        "responses": {
          "200": {
            "description": "[OK] The account is successfully retrieved.",
            "schema": {
              "$ref": "#/definitions/AccountsResponse"
            }
          },
          "400": {
            "description": "[Bad Request] The request is not valid. See response body for more detail."
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          },
          "404": {
            "description": "[NotFound] The account is not found."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "read"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "read"
      }
    },
    "/accounts/{accountNumber}/balances": {
      "get": {
        "tags": [
          "Balances"
        ],
        "summary": "Gets the balance for a specific account of a MyBank customer.",
        "operationId": "Balances_GetAccountBalances",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "parameters": [
          {
            "name": "accountNumber",
            "in": "path",
            "description": "A MyBank account number.",
            "required": true,
            "type": "string"
          }
        ],
        "responses": {
          "200": {
            "description": "[OK] The balance is successfully retrieved.",
            "schema": {
              "$ref": "#/definitions/BalancesResponse"
            }
          },
          "400": {
            "description": "[Bad Request] The request is not valid. See response body for more detail."
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          },
          "404": {
            "description": "[NotFound] The account is not found."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "read"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "read"
      }
    },
    "/instruments/lists": {
      "get": {
        "tags": [
          "Instruments"
        ],
        "summary": "Gets the list of all predefined instrument lists and their description.",
        "operationId": "Instruments_GetListOfLists",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "responses": {
          "200": {
            "description": "[OK] A list of all instrument lists.",
            "schema": {
              "$ref": "#/definitions/InstrumentListsResponse"
            }
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "read"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "read"
      }
    },
    "/instruments/lists/{id}": {
      "get": {
        "tags": [
          "Instruments"
        ],
        "summary": "Gets all instruments in a predefined list",
        "description": "Use instruments/lists to get all predefined lists.",
        "operationId": "Instruments_GetListContents",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "description": "id of the list",
            "required": true,
            "type": "string"
          },
          {
            "name": "accountNumber",
            "in": "query",
            "description": "Mandatory Account Number",
            "required": false,
            "type": "string"
          },
          {
            "name": "range",
            "in": "query",
            "description": "Paging parameter to retrieve a subset of the complete collection. Format is &lt;offset&gt;-&lt;limit&gt;. \r\nBoth values are an offset from the first entry of the complete collection. The first entry has offset '0'.\r\n(e.g. 12-21)",
            "required": false,
            "type": "string",
            "pattern": "[0-9]+-[0-9]*"
          }
        ],
        "responses": {
          "200": {
            "description": "[OK] A list of 1 or more instruments.",
            "schema": {
              "$ref": "#/definitions/InstrumentsResponseModel"
            }
          },
          "400": {
            "description": "[Bad Request] The request is not valid. See response body for more detail."
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          },
          "404": {
            "description": "Not found"
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "read"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "read"
      }
    },
    "/instruments": {
      "get": {
        "tags": [
          "Instruments"
        ],
        "summary": "Gets instrument information.",
        "description": "Parameter 'SearchText' or 'Isin' is required. 'Type' is optional, 'Mic' can only be used together with 'Isin'.",
        "operationId": "Instruments_GetInstruments",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "parameters": [
          {
            "name": "accountNumber",
            "in": "query",
            "description": "Mandatory Account Number",
            "required": false,
            "type": "string"
          },
          {
            "name": "instrumentType",
            "in": "query",
            "description": "Additional optional filter on instrument type. Cannot be used alone.",
            "required": false,
            "type": "string",
            "enum": [
              "index",
              "equity",
              "sprinter",
              "turbo",
              "speeder",
              "otherLeveragedProduct",
              "discounter",
              "optionClass",
              "option",
              "investmentFund",
              "tracker",
              "futureClass",
              "future",
              "bond",
              "warrant",
              "certificate",
              "structuredProduct",
              "srdClass",
              "srd",
              "ipo",
              "claim",
              "choiceDividend",
              "stockDividend",
              "cashDividend",
              "coupon",
              "unclassified"
            ]
          },
          {
            "name": "searchText",
            "in": "query",
            "description": "Case insensitive search text, minimum length 2. Cannot be used in combination with 'Isin'.",
            "required": false,
            "type": "string"
          },
          {
            "name": "isin",
            "in": "query",
            "description": "Selection on isincode. Cannot be used in combination with 'SearchText'.",
            "required": false,
            "type": "string"
          },
          {
            "name": "mic",
            "in": "query",
            "description": "Additional optional selection on Market Identification Code, to be used only in combination with 'Isin'",
            "required": false,
            "type": "string"
          },
          {
            "name": "range",
            "in": "query",
            "description": "Paging parameter to retrieve a subset of the complete collection. Format is &lt;offset&gt;-&lt;limit&gt;. \r\nBoth values are an offset from the first entry of the complete collection. The first entry has offset '0'.\r\n(e.g. 12-21)",
            "required": false,
            "type": "string",
            "pattern": "[0-9]+-[0-9]*"
          }
        ],
        "responses": {
          "200": {
            "description": "[OK] A list of 1 or more instruments.",
            "schema": {
              "$ref": "#/definitions/InstrumentsResponseModel"
            }
          },
          "400": {
            "description": "[Bad Request] The request is not valid. See response body for more detail."
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "read"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "read"
      }
    },
    "/instruments/{id}": {
      "get": {
        "tags": [
          "Instruments"
        ],
        "summary": "Gets instrument information for the specific id.",
        "operationId": "Instruments_GetInstrument",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "description": "Ids of the equity to retrieve. If there are multiple ids, seperate them by comma's.",
            "required": true,
            "type": "string"
          },
          {
            "name": "accountNumber",
            "in": "query",
            "description": "Mandatory Account Number",
            "required": false,
            "type": "string"
          }
        ],
        "responses": {
          "200": {
            "description": "[OK] The instrument information is successfully retrieved.",
            "schema": {
              "$ref": "#/definitions/InstrumentsResponseModel"
            }
          },
          "400": {
            "description": "[Bad Request] The request is not valid. See response body for more detail."
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          },
          "404": {
            "description": "Not found"
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "read"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "read"
      }
    },
    "/instruments/derivatives": {
      "get": {
        "tags": [
          "Instruments"
        ],
        "summary": "Gets the series for a derivatives class (options/futures).",
        "operationId": "Instruments_GetDerivatives",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "parameters": [
          {
            "name": "accountNumber",
            "in": "query",
            "description": "Mandatory Account Number",
            "required": false,
            "type": "string"
          },
          {
            "name": "symbol",
            "in": "query",
            "description": "Selection on symbol. \r\nCannot be used in combination with 'UnderlyingInstrumentId'.",
            "required": false,
            "type": "string"
          },
          {
            "name": "underlyingInstrumentId",
            "in": "query",
            "description": "Selection on the ID of the underlying equity.\r\nCannot be used in combination with 'symbol'.",
            "required": false,
            "type": "string"
          },
          {
            "name": "range",
            "in": "query",
            "description": "Paging parameter to retrieve a subset of the complete collection. Format is &lt;offset&gt;-&lt;limit&gt;. \r\nBoth values are an offset from the first entry of the complete collection. The first entry has offset '0'.\r\n(e.g. 12-21)",
            "required": false,
            "type": "string",
            "pattern": "[0-9]+-[0-9]*"
          }
        ],
        "responses": {
          "200": {
            "description": "[OK] A list of derivative series information.",
            "schema": {
              "$ref": "#/definitions/DerivativesResponseModel"
            }
          },
          "400": {
            "description": "[Bad Request] The request is not valid. See response body for more detail."
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "read"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "read"
      }
    },
    "/accounts/{accountNumber}/orders": {
      "get": {
        "tags": [
          "Orders"
        ],
        "summary": "Gets all the orders of a specific MyBank account.",
        "description": "If there is no order, the collection will be empty.",
        "operationId": "Orders_GetOrders",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "parameters": [
          {
            "name": "accountNumber",
            "in": "path",
            "description": "A MyBank account number.",
            "required": true,
            "type": "string"
          },
          {
            "name": "range",
            "in": "query",
            "description": "Paging parameter to retrieve a subset of the complete collection. Format is &lt;offset&gt;-&lt;limit&gt;. \r\nBoth values are an offset from the first entry of the complete collection. The first entry has offset '0'.\r\n(e.g. 12-21)",
            "required": false,
            "type": "string",
            "pattern": "[0-9]+-[0-9]*"
          },
          {
            "name": "status",
            "in": "query",
            "description": "'all' will select all the orders. Other possible values are 'open', 'executed' and 'canceled'.",
            "required": false,
            "type": "string",
            "pattern": "^all$|^open$|^executed$|^canceled$"
          }
        ],
        "responses": {
          "200": {
            "description": "[OK] A list of the orders of a BinkBank customer.",
            "schema": {
              "$ref": "#/definitions/OrdersResponse"
            }
          },
          "400": {
            "description": "[Bad Request] The request is not valid. See response body for more detail."
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "read"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "read"
      },
      "post": {
        "tags": [
          "Orders"
        ],
        "summary": "Registers an order to sent to the market.",
        "description": "This order will be sent to the market and executed according to the specifications.",
        "operationId": "Orders_RegisterOrder",
        "consumes": [
          "application/json",
          "text/json",
          "application/xml",
          "text/xml",
          "application/x-www-form-urlencoded"
        ],
        "produces": [
          "application/json",
          "text/json"
        ],
        "parameters": [
          {
            "name": "accountNumber",
            "in": "path",
            "description": "A MyBank account number to register the order.",
            "required": true,
            "type": "string"
          },
          {
            "name": "newOrder",
            "in": "body",
            "description": "Specifications to be used for the order.",
            "required": true,
            "schema": {
              "$ref": "#/definitions/NewOrderModel"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "[Created] Order is succesfully registered. The order number in MyBank system \r\n            is returned in the response.",
            "schema": {
              "$ref": "#/definitions/OrderNumberResponse"
            }
          },
          "400": {
            "description": "[Bad Request] The request is not valid. See response body for more detail."
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          },
          "422": {
            "description": "[Unprocessable Entity] The request is valid but cannot be executed. See response body for more detail."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "write"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "write"
      }
    },
    "/accounts/{accountNumber}/orders/{number}": {
      "get": {
        "tags": [
          "Orders"
        ],
        "summary": "Gets the specific order for a MyBank account.",
        "operationId": "Orders_GetOrder",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "parameters": [
          {
            "name": "accountNumber",
            "in": "path",
            "description": "A MyBank account number.",
            "required": true,
            "type": "string"
          },
          {
            "name": "number",
            "in": "path",
            "description": "The number of the requested order.",
            "required": true,
            "type": "integer",
            "format": "int64"
          }
        ],
        "responses": {
          "200": {
            "description": "[OK] The order is successfully retrieved.",
            "schema": {
              "$ref": "#/definitions/OrderResponse"
            }
          },
          "400": {
            "description": "[Bad Request] The request is not valid. See response body for more detail."
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          },
          "404": {
            "description": "[NotFound] The order is not found."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "read"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "read"
      },
      "delete": {
        "tags": [
          "Orders"
        ],
        "summary": "Requests to cancel an order.",
        "description": "Cancelation is possible only if the order has not already been executed.",
        "operationId": "Orders_CancelOrder",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "parameters": [
          {
            "name": "accountNumber",
            "in": "path",
            "description": "The MyBank account number used to register the order.",
            "required": true,
            "type": "string"
          },
          {
            "name": "number",
            "in": "path",
            "description": "The order number for this account.",
            "required": true,
            "type": "integer",
            "format": "int64"
          }
        ],
        "responses": {
          "200": {
            "description": "[OK] The cancelation request is succesfully created.",
            "schema": {
              "$ref": "#/definitions/OrderNumberResponse"
            }
          },
          "400": {
            "description": "[Bad Request] The request is not valid. See response body for more detail."
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          },
          "422": {
            "description": "[Unprocessable Entity] The request is valid but cannot be executed. See response body for more detail."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "write"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "write"
      }
    },
    "/accounts/{accountNumber}/orders/preview": {
      "post": {
        "tags": [
          "Orders"
        ],
        "summary": "Previews an order.",
        "description": "This allows you to validate an order without sending it to the market. The order will not be created in\r\nthe MyBank system. The response will contain useful information.",
        "operationId": "Orders_PreviewOrder",
        "consumes": [
          "application/json",
          "text/json",
          "application/xml",
          "text/xml",
          "application/x-www-form-urlencoded"
        ],
        "produces": [
          "application/json",
          "text/json"
        ],
        "parameters": [
          {
            "name": "accountNumber",
            "in": "path",
            "description": "A MyBank account number to register the order.",
            "required": true,
            "type": "string"
          },
          {
            "name": "newOrder",
            "in": "body",
            "description": "Specifications to be used for the order.",
            "required": true,
            "schema": {
              "$ref": "#/definitions/NewOrderModel"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "[Succes] The preview is successful. For more information about the order, check the response.",
            "schema": {
              "$ref": "#/definitions/PreviewOrderResponse"
            }
          },
          "400": {
            "description": "[Bad Request] The request is not valid. See response body for more detail."
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "write"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "write"
      }
    },
    "/accounts/{accountNumber}/performances": {
      "get": {
        "tags": [
          "Performances"
        ],
        "summary": "Gets the financial performance information for a MyBank account.",
        "operationId": "Performances_GetPerformances",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "parameters": [
          {
            "name": "accountNumber",
            "in": "path",
            "description": "A MyBank account number.",
            "required": true,
            "type": "string"
          },
          {
            "name": "onPosition",
            "in": "query",
            "description": "Performances can be calculated on position level or on instrument level. When 'onPosition' set to true,\r\nthe performance of all individual instruments will be reported. If set to false, the performance of \r\nderivative instruments is included in the performance of the underlying instrument.",
            "required": false,
            "type": "boolean"
          },
          {
            "name": "year",
            "in": "query",
            "description": "Performances can be calculated in detail for the requested year. When ommited a performances summary for\r\neach year is given. When used a 4-digit year is expected.",
            "required": false,
            "type": "integer",
            "format": "int32",
            "pattern": "^\\d{4}$"
          }
        ],
        "responses": {
          "200": {
            "description": "[OK] The performance information is successfully retrieved.",
            "schema": {
              "$ref": "#/definitions/PerformancesResponse"
            }
          },
          "400": {
            "description": "[Bad Request] The request is not valid. See response body for more detail."
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          },
          "404": {
            "description": "[NotFound] The account is not found."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "read"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "read"
      }
    },
    "/accounts/{accountNumber}/positions": {
      "get": {
        "tags": [
          "Positions"
        ],
        "summary": "Gets all the positions of a specific MyBank account.",
        "description": "If there is no position, the collection will be empty.",
        "operationId": "Positions_GetPositions",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "parameters": [
          {
            "name": "accountNumber",
            "in": "path",
            "description": "A MyBank account number.",
            "required": true,
            "type": "string"
          },
          {
            "name": "range",
            "in": "query",
            "description": "Paging parameter to retrieve a subset of the complete collection. Format is &lt;offset&gt;-&lt;limit&gt;. \r\nBoth values are an offset from the first entry of the complete collection. The first entry has offset '0'.\r\n(e.g. 12-21)",
            "required": false,
            "type": "string",
            "pattern": "[0-9]+-[0-9]*"
          }
        ],
        "responses": {
          "200": {
            "description": "[OK] A list of the positions of a BinkBank customer.",
            "schema": {
              "$ref": "#/definitions/PositionsResponse"
            }
          },
          "400": {
            "description": "[Bad Request] The request is not valid. See response body for more detail."
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "read"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "read"
      }
    },
    "/accounts/{accountNumber}/positions/{id}": {
      "get": {
        "tags": [
          "Positions"
        ],
        "summary": "Gets the specific position for a MyBank account.",
        "operationId": "Positions_GetPosition",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "parameters": [
          {
            "name": "accountNumber",
            "in": "path",
            "description": "A MyBank account number.",
            "required": true,
            "type": "string"
          },
          {
            "name": "id",
            "in": "path",
            "description": "The Id of the requested position.",
            "required": true,
            "type": "string"
          }
        ],
        "responses": {
          "200": {
            "description": "[OK] The position is successfully retrieved.",
            "schema": {
              "$ref": "#/definitions/PositionResponse"
            }
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          },
          "404": {
            "description": "[NotFound] The position is not found."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "read"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "read"
      }
    },
    "/session": {
      "delete": {
        "tags": [
          "Session"
        ],
        "summary": "Ends the current active session.",
        "operationId": "Session_EndSession",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "responses": {
          "200": {
            "description": "[OK] The session is ended.",
            "schema": {
              "$ref": "#/definitions/LogoffResponse"
            }
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "read"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "read"
      }
    },
    "/accounts/{accountNumber}/transactions": {
      "get": {
        "tags": [
          "Transactions"
        ],
        "summary": "Gets all the transactions specific MyBank account.",
        "operationId": "Transactions_GetTransactions",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "parameters": [
          {
            "name": "accountNumber",
            "in": "path",
            "description": "A My account number",
            "required": true,
            "type": "string"
          },
          {
            "name": "range",
            "in": "query",
            "description": "Paging parameter to retrieve a subset of the complete collection. Format is &lt;offset&gt;-&lt;limit&gt;. \r\nBoth values are an offset from the first entry of the complete collection. The first entry has offset '0'.\r\n(e.g. 12-21)",
            "required": false,
            "type": "string",
            "pattern": "[0-9]+-[0-9]*"
          },
          {
            "name": "fromDate",
            "in": "query",
            "description": "Date from which to filter. Format YYYY-MM-DD",
            "required": false,
            "type": "string",
            "format": "date-time"
          },
          {
            "name": "toDate",
            "in": "query",
            "description": "Date to which to filter. Format YYYY-MM-DD",
            "required": false,
            "type": "string",
            "format": "date-time"
          },
          {
            "name": "mutationGroup",
            "in": "query",
            "description": "Valid values are 'moneyTransfer', 'dividendPayments', 'couponPayments', 'interestPayments', \r\n'buyAndSell', 'positionMutations' and 'costs'.",
            "required": false,
            "type": "string",
            "pattern": "^moneyTransfer$|^dividendPayments$|^couponPayments$|^interestPayments$|^buyAndSell$|^costs$|^positionMutations$"
          },
          {
            "name": "currency",
            "in": "query",
            "description": "3-letter currency code (ISO 4217)",
            "required": false,
            "type": "string",
            "pattern": "^[a-zA-Z]{3}$"
          }
        ],
        "responses": {
          "200": {
            "description": "[OK] A list of the transactions from an account of a BinkBank customer.",
            "schema": {
              "$ref": "#/definitions/TransactionsResponse"
            }
          },
          "400": {
            "description": "[Bad Request] The request is not valid. See response body for more detail."
          },
          "401": {
            "description": "[Unauthorized] Authorization has been denied for this request."
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              "read"
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "Application & Application User",
        "x-scope": "read"
      }
    },
    "/version": {
      "get": {
        "tags": [
          "Version"
        ],
        "summary": "Gets the current version and the last date of build.",
        "operationId": "Version_GetVersion",
        "consumes": [],
        "produces": [
          "application/json",
          "text/json"
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "$ref": "#/definitions/VersionModel"
            }
          }
        },
        "deprecated": false,
        "security": [
          {
            "oauth2": [
              ""
            ]
          }
        ],
        "x-throttling-tier": "Unlimited",
        "x-auth-type": "None",
        "x-scope": ""
      }
    }
  },
  "definitions": {
    "AccountsResponse": {
      "description": "Accounts API response",
      "required": [
        "metadata"
      ],
      "type": "object",
      "properties": {
        "accountsCollection": {
          "$ref": "#/definitions/AccountsCollectionModel",
          "description": "Collection of zero, one or more accounts"
        },
        "metadata": {
          "$ref": "#/definitions/MetadataModel",
          "description": "API response meta data"
        }
      }
    },
    "AccountsCollectionModel": {
      "description": "Collection of zero, one or more accounts",
      "type": "object",
      "properties": {
        "accounts": {
          "description": "Collection of zero, one or more accounts",
          "type": "array",
          "items": {
            "$ref": "#/definitions/AccountModel"
          }
        }
      }
    },
    "MetadataModel": {
      "description": "API response meta data",
      "required": [
        "version",
        "timestamp"
      ],
      "type": "object",
      "properties": {
        "version": {
          "description": "Version information",
          "type": "string"
        },
        "timestamp": {
          "format": "date-time",
          "description": "Date and time the response is created",
          "type": "string"
        }
      }
    },
    "AccountModel": {
      "description": "Account model",
      "type": "object",
      "properties": {
        "name": {
          "description": "The name of the account",
          "type": "string"
        },
        "iban": {
          "description": "The IBAN of the account",
          "type": "string"
        },
        "currency": {
          "description": "The account's currency",
          "type": "string"
        },
        "number": {
          "description": "Accountnumber",
          "type": "string"
        },
        "type": {
          "description": "Type of account",
          "enum": [
            "savings",
            "xxxx",
            "yyyy",
            "zzzz"
          ],
          "type": "string"
        },
        "status": {
          "description": "Status of the account",
          "enum": [
            "active",
            "inactive",
            "pendingApplication"
          ],
          "type": "string"
        }
      }
    },
    "BalancesResponse": {
      "description": "Balances API response",
      "required": [
        "metadata"
      ],
      "type": "object",
      "properties": {
        "balancesCollection": {
          "$ref": "#/definitions/BalancesCollectionModel",
          "description": "Collection of zero, one or more balances"
        },
        "metadata": {
          "$ref": "#/definitions/MetadataModel",
          "description": "API response meta data"
        }
      }
    },
    "BalancesCollectionModel": {
      "description": "Collection of zero, one or more balances",
      "type": "object",
      "properties": {
        "balances": {
          "description": "Collection of zero, one or more balances",
          "type": "array",
          "items": {
            "$ref": "#/definitions/BalanceModel"
          }
        },
        "currency": {
          "description": "Currency",
          "type": "string"
        },
        "totals": {
          "$ref": "#/definitions/BalanceModelTotals",
          "description": "Totals of all balances"
        }
      }
    },
    "BalanceModel": {
      "description": "Balance model",
      "type": "object",
      "properties": {
        "assetsTotalValue": {
          "format": "double",
          "description": "Assets total value",
          "type": "number"
        },
        "cashBalance": {
          "format": "double",
          "description": "Cash balance",
          "type": "number"
        },
        "portfolioValue": {
          "format": "double",
          "description": "Portfolio value",
          "type": "number"
        },
        "spendingPower": {
          "format": "double",
          "description": "Spending power",
          "type": "number"
        },
        "spendingLimit": {
          "format": "double",
          "description": "Spending limit",
          "type": "number"
        }
      }
    },
    "BalanceModelTotals": {
      "description": "Balance totals",
      "type": "object",
      "properties": {
        "assetsTotalValue": {
          "format": "double",
          "type": "number"
        },
        "cashBalance": {
          "format": "double",
          "type": "number"
        },
        "portfolioValue": {
          "format": "double",
          "type": "number"
        }
      }
    },
    "InstrumentListsResponse": {
      "description": "Response type for list of all instrument lists",
      "required": [
        "metadata"
      ],
      "type": "object",
      "properties": {
        "instrumentListsCollection": {
          "$ref": "#/definitions/InstrumentListsCollectionModel",
          "description": ""
        },
        "metadata": {
          "$ref": "#/definitions/MetadataModel",
          "description": "API response meta data"
        }
      }
    },
    "InstrumentListsCollectionModel": {
      "description": "Collection model",
      "type": "object",
      "properties": {
        "instrumentLists": {
          "description": "",
          "type": "array",
          "items": {
            "$ref": "#/definitions/InstrumentListModel"
          }
        }
      }
    },
    "InstrumentListModel": {
      "description": "One FundList entry",
      "type": "object",
      "properties": {
        "name": {
          "description": "Name of the list as can be used in the instruments calls",
          "type": "string"
        },
        "description": {
          "description": "Short description of the list",
          "type": "string"
        }
      }
    },
    "AccountNumberQueryParam": {
      "description": "Parameters class to validate query parameters.",
      "type": "object",
      "properties": {
        "accountNumber": {
          "description": "Mandatory Account Number",
          "type": "string"
        }
      }
    },
    "PaginationQueryParam": {
      "description": "Pagination parameters",
      "type": "object",
      "properties": {
        "range": {
          "description": "Paging parameter to retrieve a subset of the complete collection. Format is &lt;offset&gt;-&lt;limit&gt;. \r\nBoth values are an offset from the first entry of the complete collection. The first entry has offset '0'.\r\n(e.g. 12-21)",
          "pattern": "[0-9]+-[0-9]*",
          "type": "string"
        }
      }
    },
    "InstrumentsResponseModel": {
      "description": "Instruments API response (includes paging)",
      "required": [
        "metadata"
      ],
      "type": "object",
      "properties": {
        "instrumentsCollection": {
          "$ref": "#/definitions/InstrumentsCollectionModel",
          "description": "Collection of zero, one or more instruments"
        },
        "paging": {
          "$ref": "#/definitions/PagingModel",
          "description": "Paging information"
        },
        "metadata": {
          "$ref": "#/definitions/MetadataModel",
          "description": "API response meta data"
        }
      }
    },
    "InstrumentsCollectionModel": {
      "description": "Collection of instruments",
      "type": "object",
      "properties": {
        "instruments": {
          "description": "Collection of instruments",
          "type": "array",
          "items": {
            "$ref": "#/definitions/InstrumentModel"
          }
        }
      }
    },
    "PagingModel": {
      "description": "Paging model for offset base pagination",
      "type": "object",
      "properties": {
        "count": {
          "format": "int64",
          "description": "Number of entries in the complete collection",
          "type": "integer"
        },
        "limit": {
          "format": "int64",
          "description": "Offset of the last entry in this subset",
          "type": "integer"
        },
        "max": {
          "format": "int64",
          "description": "Maximum number of entries per subset",
          "type": "integer"
        },
        "offset": {
          "format": "int64",
          "description": "Offset of the first entry in this subset",
          "type": "integer"
        },
        "next": {
          "description": "parameter and value to add to the request to retrieve the next subset",
          "type": "string"
        },
        "previous": {
          "description": "parameter and value to add to the request to retrieve the previous subset",
          "type": "string"
        },
        "refresh": {
          "description": "parameter and value to add to the request to retrieve the same subset",
          "type": "string"
        }
      }
    },
    "InstrumentModel": {
      "description": "Instrument model",
      "type": "object",
      "properties": {
        "id": {
          "description": "Identification of the instrument",
          "type": "string"
        },
        "name": {
          "description": "Name of the instrument",
          "type": "string"
        },
        "symbol": {
          "description": "Symbol of the instrument",
          "type": "string"
        },
        "isincode": {
          "description": "ISIN-code of the instrument",
          "type": "string"
        },
        "type": {
          "description": "OptionType of the instrument",
          "enum": [
            "index",
            "equity",
            "sprinter",
            "turbo",
            "speeder",
            "otherLeveragedProduct",
            "discounter",
            "optionClass",
            "option",
            "investmentFund",
            "tracker",
            "futureClass",
            "future",
            "bond",
            "warrant",
            "certificate",
            "structuredProduct",
            "srdClass",
            "srd",
            "ipo",
            "claim",
            "choiceDividend",
            "stockDividend",
            "cashDividend",
            "coupon",
            "unclassified"
          ],
          "type": "string"
        },
        "marketIdentificationCode": {
          "description": "Market Identification Code of the instrument",
          "type": "string"
        },
        "derivativesInfo": {
          "$ref": "#/definitions/DerivativesInfoModel",
          "description": "Derivative serie information"
        },
        "bondInfo": {
          "$ref": "#/definitions/BondInfoModel",
          "description": "Bond only information"
        },
        "currency": {
          "description": "Currency of the instrument",
          "type": "string"
        },
        "tickerSymbol": {
          "description": "Ticker symbol of the instrument",
          "type": "string"
        }
      }
    },
    "DerivativesInfoModel": {
      "description": "Derivatives series model",
      "type": "object",
      "properties": {
        "underlyingInstrumentId": {
          "description": "Instrument Id",
          "type": "string"
        },
        "strike": {
          "format": "double",
          "description": "Strike price",
          "type": "number"
        },
        "strikeDecimals": {
          "format": "int32",
          "description": "Maximum number of decimals in strike price",
          "type": "integer"
        },
        "optionType": {
          "description": "Option type (put or call)",
          "enum": [
            "put",
            "call"
          ],
          "type": "string"
        },
        "contractSize": {
          "format": "double",
          "description": "Contract size",
          "type": "number"
        },
        "expirationDate": {
          "format": "date-time",
          "description": "Expiration date",
          "type": "string"
        }
      }
    },
    "BondInfoModel": {
      "description": "Bond information\r\nIntentionally left empty",
      "type": "object",
      "properties": {}
    },
    "InstrumentsQueryParams": {
      "description": "GetInstruments query parameters model",
      "type": "object",
      "properties": {
        "accountNumber": {
          "description": "Mandatory Account Number",
          "type": "string"
        },
        "instrumentType": {
          "description": "Additional optional filter on instrument type. Cannot be used alone.",
          "enum": [
            "index",
            "equity",
            "sprinter",
            "turbo",
            "speeder",
            "otherLeveragedProduct",
            "discounter",
            "optionClass",
            "option",
            "investmentFund",
            "tracker",
            "futureClass",
            "future",
            "bond",
            "warrant",
            "certificate",
            "structuredProduct",
            "srdClass",
            "srd",
            "ipo",
            "claim",
            "choiceDividend",
            "stockDividend",
            "cashDividend",
            "coupon",
            "unclassified"
          ],
          "type": "string"
        },
        "searchText": {
          "description": "Case insensitive search text, minimum length 2. Cannot be used in combination with 'Isin'.",
          "type": "string"
        },
        "isin": {
          "description": "Selection on isincode. Cannot be used in combination with 'SearchText'.",
          "type": "string"
        },
        "mic": {
          "description": "Additional optional selection on Market Identification Code, to be used only in combination with 'Isin'",
          "type": "string"
        }
      }
    },
    "InstrumentDerivativesQueryParams": {
      "description": "GetDerivatives query parameters model",
      "type": "object",
      "properties": {
        "accountNumber": {
          "description": "Mandatory Account Number",
          "type": "string"
        },
        "symbol": {
          "description": "Selection on symbol. \r\nCannot be used in combination with 'UnderlyingInstrumentId'.",
          "type": "string"
        },
        "underlyingInstrumentId": {
          "description": "Selection on the ID of the underlying equity.\r\nCannot be used in combination with 'symbol'.",
          "type": "string"
        }
      }
    },
    "DerivativesResponseModel": {
      "description": "Derivatives API response",
      "required": [
        "metadata"
      ],
      "type": "object",
      "properties": {
        "derivativesCollection": {
          "$ref": "#/definitions/DerivativesCollectionModel",
          "description": "Derivative classes information"
        },
        "paging": {
          "$ref": "#/definitions/PagingModel",
          "description": "Paging information"
        },
        "metadata": {
          "$ref": "#/definitions/MetadataModel",
          "description": "API response meta data"
        }
      }
    },
    "DerivativesCollectionModel": {
      "description": "Collection of instruments",
      "type": "object",
      "properties": {
        "classes": {
          "description": "Derivative classes collection",
          "type": "array",
          "items": {
            "$ref": "#/definitions/DerivativeClassInfoModel"
          }
        }
      }
    },
    "DerivativeClassInfoModel": {
      "type": "object",
      "properties": {
        "underlyingInstrumentId": {
          "description": "Identification of the underlying instrument",
          "type": "string"
        },
        "name": {
          "description": "Name of the class",
          "type": "string"
        },
        "symbol": {
          "description": "Symbol of the class",
          "type": "string"
        },
        "isincode": {
          "description": "ISIN Code of the class",
          "type": "string"
        },
        "marketIdentificationCode": {
          "description": "ISO MIC of the class",
          "type": "string"
        },
        "currency": {
          "description": "Currency of the class",
          "type": "string"
        },
        "type": {
          "description": "Type of the class (option or future class)",
          "enum": [
            "index",
            "equity",
            "sprinter",
            "turbo",
            "speeder",
            "otherLeveragedProduct",
            "discounter",
            "optionClass",
            "option",
            "investmentFund",
            "tracker",
            "futureClass",
            "future",
            "bond",
            "warrant",
            "certificate",
            "structuredProduct",
            "srdClass",
            "srd",
            "ipo",
            "claim",
            "choiceDividend",
            "stockDividend",
            "cashDividend",
            "coupon",
            "unclassified"
          ],
          "type": "string"
        },
        "contractSize": {
          "format": "double",
          "description": "Contract Size of the class",
          "type": "number"
        },
        "series": {
          "description": "Collection of series for this class",
          "type": "array",
          "items": {
            "$ref": "#/definitions/DerivativeSeriesInfoModel"
          }
        }
      }
    },
    "DerivativeSeriesInfoModel": {
      "description": "Derivatives series information",
      "type": "object",
      "properties": {
        "instrumentId": {
          "description": "Instrument Id of the serie",
          "type": "string"
        },
        "strike": {
          "format": "double",
          "description": "Strike price (options only)",
          "type": "number"
        },
        "strikeDecimals": {
          "format": "int32",
          "description": "Number of decimals in strike price (options only)",
          "type": "integer"
        },
        "optionType": {
          "description": "Option type (put or call) (options only)",
          "enum": [
            "put",
            "call"
          ],
          "type": "string"
        },
        "contractSize": {
          "format": "double",
          "description": "Contract size",
          "type": "number"
        },
        "expirationDate": {
          "format": "date-time",
          "description": "Expiration date",
          "type": "string"
        }
      }
    },
    "OrderStatusQueryParams": {
      "description": "Order status parameters",
      "type": "object",
      "properties": {
        "status": {
          "description": "'all' will select all the orders. Other possible values are 'open', 'executed' and 'canceled'.",
          "pattern": "^all$|^open$|^executed$|^canceled$",
          "type": "string"
        }
      }
    },
    "OrdersResponse": {
      "description": "Orders API response (includes paging)",
      "required": [
        "metadata"
      ],
      "type": "object",
      "properties": {
        "ordersCollection": {
          "$ref": "#/definitions/OrdersCollectionModel",
          "description": "Collection of zero, one or more orders"
        },
        "paging": {
          "$ref": "#/definitions/PagingModel",
          "description": "Paging information"
        },
        "metadata": {
          "$ref": "#/definitions/MetadataModel",
          "description": "API response meta data"
        }
      }
    },
    "OrdersCollectionModel": {
      "description": "Collection of orders",
      "type": "object",
      "properties": {
        "orders": {
          "description": "Collection of orders",
          "type": "array",
          "items": {
            "$ref": "#/definitions/OrderModel"
          }
        }
      }
    },
    "OrderModel": {
      "description": "Order model",
      "type": "object",
      "properties": {
        "number": {
          "format": "int64",
          "description": "The identification of the order",
          "type": "integer"
        },
        "instrument": {
          "$ref": "#/definitions/InstrumentBriefModel",
          "description": "Attributes of the financial instrument ordered"
        },
        "miscellaneousType": {
          "description": "Miscellaneous characteristic of the order",
          "enum": [
            "normal",
            "wholesale",
            "eveningTrade",
            "batchOrder"
          ],
          "type": "string"
        },
        "type": {
          "description": "The price type of the order",
          "enum": [
            "limit",
            "market",
            "stop",
            "stopLimit",
            "allOrNone"
          ],
          "type": "string"
        },
        "status": {
          "description": "The status of the order",
          "enum": [
            "executed",
            "canceled",
            "expired",
            "executing",
            "canceling",
            "executingRemainder",
            "cancelingRemainder",
            "remainderCanceled",
            "remainderExpired",
            "historicized",
            "refused",
            "remainderRefused"
          ],
          "type": "string"
        },
        "duration": {
          "description": "Specifies the term for which the order is in effect",
          "enum": [
            "fillOrKill",
            "immediateOrCancel",
            "day",
            "goodTillCancelled",
            "goodTillDate",
            "atTheOpening",
            "goodTillCrossing"
          ],
          "type": "string"
        },
        "line": {
          "format": "int64",
          "description": "Line number of this order in case of a multi line order",
          "type": "integer"
        },
        "transactionKind": {
          "description": "Kind of transaction as text",
          "type": "string"
        },
        "executionStatus": {
          "description": "Execution status of the order as text",
          "type": "string"
        },
        "executed": {
          "format": "double",
          "description": "Number of instruments (equities), nominal value (odds) or number of contracts (options and futures) to be executed",
          "type": "number"
        },
        "limitValue": {
          "format": "double",
          "description": "Value of the order's limit",
          "type": "number"
        },
        "averagePrice": {
          "format": "double",
          "description": "Average price of all fills on this order",
          "type": "number"
        },
        "quantity": {
          "format": "double",
          "description": "Quantity ordered",
          "type": "number"
        },
        "expirationDate": {
          "format": "date-time",
          "description": "Expiration date and time for a good till date order",
          "type": "string"
        },
        "stopPrice": {
          "format": "double",
          "description": "Stop price for a stop or stop limit order",
          "type": "number"
        },
        "rejectionReason": {
          "description": "The reason of rejection in case of a rejected order",
          "type": "string"
        },
        "statusDescription": {
          "description": "Explanation of the order's status",
          "type": "string"
        },
        "fixingPrice": {
          "format": "double",
          "description": "Fixing price of the order",
          "type": "number"
        },
        "rejectionReasonDetail": {
          "description": "Detail explanation of the reason for rejection of the order",
          "type": "string"
        }
      }
    },
    "InstrumentBriefModel": {
      "description": "Brief instrument information",
      "type": "object",
      "properties": {
        "id": {
          "description": "Identification of the instrument",
          "type": "string"
        },
        "name": {
          "description": "Instrument's name",
          "type": "string"
        }
      }
    },
    "NewOrderModel": {
      "description": "New order model with model validations.",
      "required": [
        "type",
        "duration",
        "side",
        "quantity",
        "instrumentId"
      ],
      "type": "object",
      "properties": {
        "type": {
          "description": "The kind of order to be placed",
          "enum": [
            "limit",
            "market",
            "stop",
            "stopLimit"
          ],
          "type": "string"
        },
        "duration": {
          "description": "Specifies the term for which the order is in effect",
          "enum": [
            "day",
            "goodTillCancelled",
            "goodTillDate"
          ],
          "type": "string"
        },
        "side": {
          "description": "The action that the broker is requested to perform",
          "enum": [
            "buy",
            "sell"
          ],
          "type": "string"
        },
        "quantity": {
          "format": "double",
          "description": "The number of financial instruments to buy or sell",
          "type": "number"
        },
        "instrumentId": {
          "description": "Financial instrument Id in MyBank system",
          "type": "string"
        },
        "limitValue": {
          "format": "double",
          "description": "The highest price at which to buy or the lowest price at which to sell if specified in a limit order",
          "type": "number"
        },
        "expirationDate": {
          "format": "date-time",
          "description": "The date when the order will be expired",
          "type": "string"
        },
        "stopPrice": {
          "format": "double",
          "description": "The trigger price to initiate a buy or sell order",
          "type": "number"
        }
      }
    },
    "OrderNumberResponse": {
      "description": "OrderNumber response",
      "required": [
        "metadata"
      ],
      "type": "object",
      "properties": {
        "order": {
          "$ref": "#/definitions/OrderNumberModel",
          "description": "OrderNumber response"
        },
        "metadata": {
          "$ref": "#/definitions/MetadataModel",
          "description": "API response meta data"
        }
      }
    },
    "OrderNumberModel": {
      "description": "OrderNumber model",
      "type": "object",
      "properties": {
        "number": {
          "format": "int64",
          "description": "Identification of the order",
          "type": "integer"
        }
      }
    },
    "OrderResponse": {
      "description": "Order API response",
      "required": [
        "metadata"
      ],
      "type": "object",
      "properties": {
        "ordersCollection": {
          "$ref": "#/definitions/OrdersCollectionModel",
          "description": "Collection of zero, one or more orders"
        },
        "metadata": {
          "$ref": "#/definitions/MetadataModel",
          "description": "API response meta data"
        }
      }
    },
    "PreviewOrderResponse": {
      "description": "Preview Order API response",
      "required": [
        "metadata"
      ],
      "type": "object",
      "properties": {
        "previewOrder": {
          "$ref": "#/definitions/PreviewOrderModel",
          "description": "Information related to the order validation such as if the order is processable, spending power, etc."
        },
        "metadata": {
          "$ref": "#/definitions/MetadataModel",
          "description": "API response meta data"
        }
      }
    },
    "PreviewOrderModel": {
      "description": "This object is used in the response of the order validation call.",
      "type": "object",
      "properties": {
        "orderCanBeRegistered": {
          "description": "True if the order can be placed",
          "type": "boolean"
        },
        "expectedExpirationDate": {
          "format": "date-time",
          "description": "For GTD en GTC orders the enddate will be limited to about 2 weeks maximum",
          "type": "string"
        },
        "side": {
          "description": "Contains the side information (buy, sell)",
          "type": "string"
        },
        "positionEffect": {
          "description": "Contains the position effect information (open, close)",
          "type": "string"
        },
        "effectOnSpendingLimit": {
          "format": "double",
          "description": "Effect of a succesfully placed order on the spending limit of the account",
          "type": "number"
        },
        "currentSpendingLimit": {
          "format": "double",
          "description": "The current spending limit of the account (before placing the order)",
          "type": "number"
        },
        "newSpendingLimit": {
          "format": "double",
          "description": "The new spending limit of the account (after placing the order)",
          "type": "number"
        },
        "currency": {
          "description": "The currency of the spending limit",
          "type": "string"
        },
        "oldRiskNumber": {
          "format": "int32",
          "description": "Risk number before placing the order",
          "type": "integer"
        },
        "newRiskNumber": {
          "format": "int32",
          "description": "Risk number after successfully placing the order",
          "type": "integer"
        },
        "recommendedRiskNumber": {
          "format": "int32",
          "description": "Recommended risk number",
          "type": "integer"
        },
        "warnings": {
          "description": "Warnings or error messages about the requested order",
          "type": "array",
          "items": {
            "type": "string"
          }
        }
      }
    },
    "PerformancesQueryParams": {
      "description": "Performances query parameters model",
      "type": "object",
      "properties": {
        "onPosition": {
          "description": "Performances can be calculated on position level or on instrument level. When 'onPosition' set to true,\r\nthe performance of all individual instruments will be reported. If set to false, the performance of \r\nderivative instruments is included in the performance of the underlying instrument.",
          "type": "boolean"
        },
        "year": {
          "format": "int32",
          "description": "Performances can be calculated in detail for the requested year. When ommited a performances summary for\r\neach year is given. When used a 4-digit year is expected.",
          "pattern": "^\\d{4}$",
          "type": "integer"
        }
      }
    },
    "PerformancesResponse": {
      "description": "Performances API response",
      "required": [
        "metadata"
      ],
      "type": "object",
      "properties": {
        "performancesCollection": {
          "$ref": "#/definitions/PerformanceCollectionModel"
        },
        "summary": {
          "$ref": "#/definitions/PerformanceSummaryModel"
        },
        "metadata": {
          "$ref": "#/definitions/MetadataModel",
          "description": "API response meta data"
        }
      }
    },
    "PerformanceCollectionModel": {
      "description": "Collection of zero, one or more performances for positions",
      "type": "object",
      "properties": {
        "performances": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/PerformanceDetailModel"
          }
        }
      }
    },
    "PerformanceSummaryModel": {
      "description": "Summary for a year or a total summary (in case of year results requested)",
      "type": "object",
      "properties": {
        "year": {
          "description": "Year for this summary",
          "type": "string"
        },
        "currency": {
          "description": "Currency for this summary",
          "type": "string"
        },
        "realized": {
          "format": "double",
          "description": "Realized profit/loss",
          "type": "number"
        },
        "unrealized": {
          "format": "double",
          "description": "Unrealized profit/loss",
          "type": "number"
        },
        "annual": {
          "format": "double",
          "description": "Total this year",
          "type": "number"
        },
        "previousYearsTotal": {
          "format": "double",
          "description": "Including previous years",
          "type": "number"
        },
        "total": {
          "format": "double",
          "description": "Total",
          "type": "number"
        }
      }
    },
    "PerformanceDetailModel": {
      "description": "Performance for one instrument. Due to rounding errors some calculation have difference of one cent.",
      "type": "object",
      "properties": {
        "currency": {
          "type": "string"
        },
        "instrument": {
          "$ref": "#/definitions/InstrumentBriefModel",
          "description": "Instrument information"
        },
        "year": {
          "description": "Year in case of summary info",
          "type": "string"
        },
        "realized": {
          "format": "double",
          "description": "Realized profit/loss",
          "type": "number"
        },
        "unrealized": {
          "format": "double",
          "description": "Unrealized profit/loss",
          "type": "number"
        },
        "annual": {
          "format": "double",
          "description": "This year",
          "type": "number"
        },
        "previousYearsTotal": {
          "format": "double",
          "description": "Including previous years",
          "type": "number"
        },
        "total": {
          "format": "double",
          "description": "Total",
          "type": "number"
        }
      }
    },
    "PositionsResponse": {
      "description": "Positions API response (includes paging)",
      "required": [
        "metadata"
      ],
      "type": "object",
      "properties": {
        "positionsCollection": {
          "$ref": "#/definitions/PositionsCollectionModel",
          "description": "Collection of zero, one or more positions"
        },
        "paging": {
          "$ref": "#/definitions/PagingModel",
          "description": "Paging information"
        },
        "metadata": {
          "$ref": "#/definitions/MetadataModel",
          "description": "API response meta data"
        }
      }
    },
    "PositionsCollectionModel": {
      "description": "Collection of positions",
      "type": "object",
      "properties": {
        "positions": {
          "description": "Collection of positions",
          "type": "array",
          "items": {
            "$ref": "#/definitions/PositionModel"
          }
        }
      }
    },
    "PositionModel": {
      "description": "Position model",
      "type": "object",
      "properties": {
        "instrument": {
          "$ref": "#/definitions/InstrumentBriefModel"
        },
        "id": {
          "description": "Identification of the position",
          "type": "string"
        },
        "quantity": {
          "format": "int64",
          "description": "Number of securities or contracts or nominal value",
          "type": "integer"
        },
        "currency": {
          "description": "Currency",
          "type": "string"
        },
        "accruedInterest": {
          "$ref": "#/definitions/PositionAccruedInterest",
          "description": "Accrued interest in case of a debt instrument"
        },
        "historicPrice": {
          "format": "double",
          "description": "Historic price",
          "type": "number"
        },
        "midPrice": {
          "format": "double",
          "description": "Mid price",
          "type": "number"
        },
        "valueInEuro": {
          "format": "double",
          "description": "Value of the position expressed in the EURO currency",
          "type": "number"
        },
        "averagePurchasePrice": {
          "format": "double",
          "description": "Average purchase price",
          "type": "number"
        },
        "averagePrice": {
          "format": "double",
          "description": "Average price",
          "type": "number"
        },
        "coverage": {
          "$ref": "#/definitions/PositionCoverage",
          "description": "Coverage"
        },
        "margin": {
          "$ref": "#/definitions/PositionMargin",
          "description": "Margin"
        },
        "result": {
          "$ref": "#/definitions/PositionResult",
          "description": "Result"
        },
        "resultInEuro": {
          "$ref": "#/definitions/PositionResult",
          "description": "Result expressed in the EURO currency"
        },
        "value": {
          "format": "double",
          "description": "Value",
          "type": "number"
        }
      }
    },
    "PositionAccruedInterest": {
      "description": "Position accrued interest",
      "type": "object",
      "properties": {
        "value": {
          "format": "double",
          "description": "Value",
          "type": "number"
        },
        "rate": {
          "format": "double",
          "description": "Rate",
          "type": "number"
        }
      }
    },
    "PositionCoverage": {
      "description": "Position coverage",
      "type": "object",
      "properties": {
        "value": {
          "format": "double",
          "description": "Value",
          "type": "number"
        },
        "dynamicRate": {
          "format": "int64",
          "description": "Dynamic rate",
          "type": "integer"
        },
        "staticRate": {
          "format": "int64",
          "description": "Static rate",
          "type": "integer"
        }
      }
    },
    "PositionMargin": {
      "description": "Position margin",
      "type": "object",
      "properties": {
        "value": {
          "format": "double",
          "description": "Value",
          "type": "number"
        },
        "factor": {
          "format": "double",
          "description": "Factor",
          "type": "number"
        }
      }
    },
    "PositionResult": {
      "description": "Position result",
      "type": "object",
      "properties": {
        "currency": {
          "description": "Currency the result is expressed in",
          "type": "string"
        },
        "unrealized": {
          "format": "double",
          "description": "Unrealized result",
          "type": "number"
        },
        "realized": {
          "format": "double",
          "description": "Realized result",
          "type": "number"
        },
        "total": {
          "format": "double",
          "description": "Total result",
          "type": "number"
        },
        "unrealizedPercentage": {
          "format": "double",
          "description": "Unrealized result in a percentage of ... (?)",
          "type": "number"
        },
        "purchaseValue": {
          "format": "double",
          "description": "Purchase value",
          "type": "number"
        }
      }
    },
    "PositionResponse": {
      "description": "Position API response",
      "required": [
        "metadata"
      ],
      "type": "object",
      "properties": {
        "positionsCollection": {
          "$ref": "#/definitions/PositionsCollectionModel"
        },
        "metadata": {
          "$ref": "#/definitions/MetadataModel",
          "description": "API response meta data"
        }
      }
    },
    "LogoffResponse": {
      "description": "Response of the logoff request.",
      "required": [
        "metadata"
      ],
      "type": "object",
      "properties": {
        "message": {
          "description": "A message confirming the sign out is complete",
          "type": "string"
        },
        "metadata": {
          "$ref": "#/definitions/MetadataModel",
          "description": "API response meta data"
        }
      }
    },
    "DateRangeQueryParams": {
      "description": "From - To date parameters",
      "type": "object",
      "properties": {
        "fromDate": {
          "format": "date-time",
          "description": "Date from which to filter. Format YYYY-MM-DD",
          "type": "string"
        },
        "toDate": {
          "format": "date-time",
          "description": "Date to which to filter. Format YYYY-MM-DD",
          "type": "string"
        }
      }
    },
    "MutationGroupQueryParams": {
      "description": "Mutation group parameter",
      "type": "object",
      "properties": {
        "mutationGroup": {
          "description": "Valid values are 'moneyTransfer', 'dividendPayments', 'couponPayments', 'interestPayments', \r\n'buyAndSell', 'positionMutations' and 'costs'.",
          "pattern": "^moneyTransfer$|^dividendPayments$|^couponPayments$|^interestPayments$|^buyAndSell$|^costs$|^positionMutations$",
          "type": "string"
        },
        "mutationGroupType": {
          "description": "Return the mutation group as a models mutation group type\r\nReturns null if nothing is filled in.",
          "enum": [
            "moneyTransfer",
            "dividendPayments",
            "couponPayments",
            "interestPayments",
            "positionMutations",
            "costs",
            "buyAndSell",
            "secLenPayments"
          ],
          "type": "string",
          "readOnly": true
        }
      }
    },
    "CurrencyQueryParams": {
      "description": "Currency parameter",
      "type": "object",
      "properties": {
        "currency": {
          "description": "3-letter currency code (ISO 4217)",
          "pattern": "^[a-zA-Z]{3}$",
          "type": "string"
        }
      }
    },
    "TransactionsResponse": {
      "description": "Transaction API response",
      "required": [
        "metadata"
      ],
      "type": "object",
      "properties": {
        "transactionsCollection": {
          "$ref": "#/definitions/TransactionsCollectionModel",
          "description": "Collection of zero, one or more Transactions"
        },
        "paging": {
          "$ref": "#/definitions/PagingModel",
          "description": "Paging information"
        },
        "metadata": {
          "$ref": "#/definitions/MetadataModel",
          "description": "API response meta data"
        }
      }
    },
    "TransactionsCollectionModel": {
      "type": "object",
      "properties": {
        "transactions": {
          "description": "Collection of transactions",
          "type": "array",
          "items": {
            "$ref": "#/definitions/TransactionModel"
          }
        }
      }
    },
    "TransactionModel": {
      "description": "Transaction model",
      "type": "object",
      "properties": {
        "transactionCostComponents": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/TransactionCostComponentModel"
          }
        },
        "accountCurrency": {
          "description": "Currency code for an account",
          "type": "string"
        },
        "sequenceId": {
          "format": "int64",
          "description": "Secuence id of the transaction for an account currency",
          "type": "integer"
        },
        "id": {
          "description": "Unique id for this transaction",
          "type": "string"
        },
        "date": {
          "format": "date-time",
          "description": "Date of the transaction",
          "type": "string"
        },
        "mutationType": {
          "description": "Enumerated value of the mutation type",
          "enum": [
            "unknown",
            "repayment",
            "conversion",
            "excerciseCall",
            "excercisePut",
            "assignmentCall",
            "assignmentPut",
            "settlementDividend",
            "deposit",
            "emissionAllocation",
            "internalBooking",
            "transferOutOfSecurities",
            "notificationOfRedemption",
            "outstandingBooking",
            "securityTransfer",
            "conversionClaims",
            "conversionDividend",
            "cashSettlement",
            "adjustment",
            "awardCoupon",
            "internalTransfer",
            "externalTransfer",
            "cashTransferToPartnerBank",
            "cashTransferFromPartnerBank",
            "openingBuy",
            "openingBuyFutures",
            "openingSale",
            "openingSaleFuture",
            "regulation",
            "creditInterest",
            "debitInterest",
            "closingBuy",
            "closingBuyFuture",
            "closingSale",
            "closingSaleFuture",
            "awardClaim",
            "awardDividend",
            "couponPayment",
            "dividendPayment",
            "settlementCosts",
            "sell",
            "buy",
            "liquidationTransfer",
            "extensionOpenDeposit",
            "extensionOpenTransferOut",
            "extensionCloseDeposit",
            "extensionCloseTransferOut",
            "settlementBuy",
            "settlementSell",
            "securitiesLendingDividendPayment",
            "securitiesLendingCouponPayment",
            "securitiesLendingInterestPayment",
            "onlineMoneyTransfer"
          ],
          "type": "string"
        },
        "description": {
          "description": "Human readable description of the mutation details",
          "type": "string"
        },
        "balanceMutation": {
          "format": "double",
          "description": "Total amount when the transaction is completed",
          "type": "number"
        },
        "mutatedBalance": {
          "format": "double",
          "description": "Total amount when the transaction is completed",
          "type": "number"
        },
        "instrument": {
          "$ref": "#/definitions/InstrumentBriefModel",
          "description": "The instrument object"
        },
        "price": {
          "format": "double",
          "description": "The price of one instrument",
          "type": "number"
        },
        "quantity": {
          "format": "double",
          "description": "The number of financial instruments to buy or sell",
          "type": "number"
        },
        "exchange": {
          "description": "Name of the exchange where this instrument was handled",
          "type": "string"
        },
        "totalCosts": {
          "format": "double",
          "description": "All costs for this transaction",
          "type": "number"
        },
        "currency": {
          "description": "Transaction currency. This currency was used to complete this transaction",
          "type": "string"
        },
        "netAmount": {
          "format": "double",
          "description": "The total amount for this transaction without the costs",
          "type": "number"
        },
        "currencyRate": {
          "format": "double",
          "description": "The exchange rate used for the transaction currency",
          "type": "number"
        }
      }
    },
    "TransactionCostComponentModel": {
      "description": "Transaction costs components",
      "type": "object",
      "properties": {
        "cost": {
          "format": "double",
          "description": "Amount for the costs",
          "type": "number"
        },
        "currency": {
          "description": "Currency of these costs",
          "type": "string"
        },
        "costTypeCategory": {
          "description": "Costs category",
          "enum": [
            "commission",
            "stampDuty",
            "capitalGainTax",
            "capitalIncomeTax",
            "exchangeTax",
            "withHoldingTax",
            "stampDutyBE",
            "vAT",
            "securitiesFee",
            "socialTax",
            "incomeTax",
            "sRDExtensionCommission",
            "sRDDifference",
            "sRDSettlement",
            "sRDCompensationPayment",
            "sRDCommission",
            "speculationTax",
            "assetManagementFee"
          ],
          "type": "string"
        },
        "costTypeSubCategory": {
          "description": "Costs sub category",
          "enum": [
            "fixedcosts",
            "variableStorage",
            "feerebate",
            "aEBandASASCourtage",
            "oBOcosts",
            "orderprocessingcosts",
            "copystatementofaccount",
            "exerciseCosts",
            "aboutbooking",
            "nYcommissionFid",
            "fixedcostsUSA",
            "fTAFinsettlFUSFTI",
            "fTAFinsettlFTOFT5etc",
            "clearFloorBrokerfee",
            "fTAClearingFees",
            "fTACourtageFTIL",
            "fTAFinsettlFFAFIAetc",
            "eOEBrokerageindexoptions2",
            "eOEBrokeragepressindexoptions",
            "eOEBrokeragesharesopt2",
            "eOEBrokeragebondoptions",
            "eOEBrokeragecurrencyoptions",
            "eOEBrokerageE100OPT2",
            "eOEBrokerageE100opt3",
            "exercEOEoptpushAss",
            "exercEOEoptAssobl",
            "exercEOEAssOtheroptions",
            "exercEOEoptdirectoryAss",
            "eOEFloorBrokerfee",
            "eOEClearingFees",
            "eOEespAEBclearingexass",
            "exercEOEoptAssE100",
            "eOEBrokerageE100opt4",
            "euronextexerciseassignment",
            "fTACourtage",
            "fTACourtageFTO",
            "aEBandASASCourtage2",
            "aEBClearingFees",
            "aSASClearingFees",
            "aEBASASandCommCTCI",
            "aEBASASandCommTSA",
            "bondCosts1",
            "bondCosts2",
            "bondCosts3",
            "nYcommissionFid2",
            "nYClearingFees",
            "nYCommLiberty",
            "nYExerciseCosts",
            "uSFixedCosts",
            "uScharges",
            "londonFixedcosts",
            "fixedcostsZurich",
            "londonstampduty",
            "stampdutyZurich",
            "fixedcostsFutChicago",
            "fixedcostsFutEurex",
            "fixedkosten2USA",
            "parisoverheads",
            "brusselsfixedcosts",
            "fixedcostsCopenhagen",
            "fixedcostsStockholm",
            "fixedcostsOslo",
            "euronextcharges",
            "futlight",
            "futindex",
            "options",
            "optionsgt065",
            "options000to005",
            "clearingexassow100",
            "optlight000to005",
            "optlightgt006",
            "clearingexassIndexow10",
            "clearingexassIndexow100",
            "exasscurrencyoptionsNL",
            "currencyoptions",
            "clearingexassstockoption",
            "fixedcostsUSExercised",
            "currencyoptionslt005",
            "divdailystatement",
            "divOverheads",
            "divOrderbehoral",
            "marketDataFees",
            "divdailystatementexass",
            "divexOverheadsass",
            "divOrderbehoralexass",
            "copystatementofaccount2",
            "exerciseCosts2",
            "provisieGeldboekingen",
            "vAT",
            "koersabonnementen",
            "vermogensbeheervergoeding",
            "bewaarloon",
            "algemenedienstverlening",
            "productabonnementen",
            "retourprovisie",
            "aftedragenvergoeding",
            "dienstverleningBTWvrij",
            "aandelenItalie",
            "futuresItalie",
            "optiesItalie",
            "obligatiesEuroTLX",
            "capitalincometax",
            "capitalincometaxBNP",
            "capitalgaintax",
            "capitalincometax2",
            "capitalincometax3",
            "securitiesfee",
            "socialtax",
            "socialtax2",
            "beurstax",
            "roerendeVoorheffing",
            "zegelrechtBelgium",
            "rVtaxXrekalingehouden",
            "rVtaxXNrekalingehouden",
            "speculatietaks"
          ],
          "type": "string"
        },
        "amountPercentage": {
          "format": "double",
          "description": "???",
          "type": "number"
        }
      }
    },
    "VersionModel": {
      "description": "Version model",
      "required": [
        "currentVersion",
        "buildDate",
        "metadata"
      ],
      "type": "object",
      "properties": {
        "currentVersion": {
          "description": "Current version",
          "type": "string"
        },
        "buildDate": {
          "description": "Build date",
          "type": "string"
        },
        "metadata": {
          "$ref": "#/definitions/MetadataModel",
          "description": "API response meta data"
        }
      }
    }
  },
  "securityDefinitions": {
    "oauth2": {
      "type": "oauth2",
      "description": "OAuth2 Authorization Code Grant",
      "flow": "accessCode",
      "authorizationUrl": "https://oauth2.mybank.com/openam/oauth2/authorize",
      "tokenUrl": "https://oauth2.mybank.com/openam/oauth2/access_token?realm=myapi",
      "scopes": {
        "read": "Read access to protected resources",
        "write": "Write access to protected resources"
      }
    }
  },
  "security": [
    {
      "oauth2": [
        "read",
        "write"
      ]
    }
  ],
  "x-wso2-security": {
    "apim": {
      "x-wso2-scopes": [
        {
          "name": "Read",
          "description": "",
          "key": "read",
          "roles": ""
        },
        {
          "name": "Write",
          "description": "",
          "key": "write",
          "roles": ""
        }
      ]
    }
  }
}
`
