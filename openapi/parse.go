// Package openapi is used to access Swagger specs.
package openapi

import (
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

// OAuth2ScopeIterFunc functions are used to collect path, action and oauth2 scopes from a swagger spec.
type OAuth2ScopeIterFunc func(path string, action string, scopes []string)

// SpecFromRaw returns a swagger spec from a json blob.
func SpecFromRaw(json []byte) (*spec.Swagger, error) {
	doc, err := loads.Analyzed(json, "")
	if err != nil {
		return nil, err
	}

	return doc.Spec(), nil
}

// SpecOAuth2ScopeIter iterates a Swagger spec and calls a function with url path, http action and oauth2 scopes.
// Prerequisite: specification.Path != nil
func SpecOAuth2ScopeIter(specification *spec.Swagger, fn OAuth2ScopeIterFunc) {
	op := func(path string, method string, prop *spec.Operation, fn OAuth2ScopeIterFunc) {
		if prop == nil {
			// no properties so ignore path
			return
		}

		if len(prop.Security) == 0 {
			// no Security properties so no scopes
			fn(path, method, nil)
			return
		}

		scopes, ok := prop.Security[0]["oauth2"]
		if !ok {
			// no oauth2 property so no scopes
			fn(path, method, nil)
			return
		}

		fn(path, method, scopes)
	}

	for path, prop := range specification.Paths.Paths {
		op(path, "GET", prop.Get, fn)
		op(path, "PUT", prop.Put, fn)
		op(path, "POST", prop.Post, fn)
		op(path, "DELETE", prop.Delete, fn)
		op(path, "OPTIONS", prop.Options, fn)
		op(path, "HEAD", prop.Head, fn)
		op(path, "PATCH", prop.Patch, fn)
	}
}
