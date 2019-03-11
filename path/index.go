// Package path provides and index that stores http method/path and associated values like scopes.
package path

import (
	"strings"
	"fmt"
)

// Index stores http method/path and associated values.
type Index struct {
	// Root of the tree.
	root *node
}

// Node of a tree.
type node struct {
	// Name of the node. For example path /api has a node with named 'api'.
	name string
	// Param indicates if the node represents a parameter.
	param bool
	// Children nodes.
	children nodes
	// Methods are the http methods available for a path with their associated scopes.
	Methods map[string]Scopes
}

type nodes []*node

// Scopes are a collection of OAuth2 scope names.
type Scopes []string


// NewIndex returns an empty index.
func NewIndex() *Index {
	return &Index{root: &node{}}
}

// AddMethodPathScopes adds http method/path with scopes to the receiver.
func (idx *Index) AddMethodPathScopes(method, path string, scopes []string) (*node, error) {
	n, err := idx.AddPath(path)
	if err != nil {
		return nil, err
	}
	n.Methods[method] = scopes

	return n, nil
}

// AddPath adds a http path to the receiver and returns a node.
func (idx *Index) AddPath(path string) (*node, error) {
	n := idx.root
	for _, e := range split(path) {
		// is this element a parameter (see http://www.ietf.org/rfc/rfc1738.txt)?
		isParam := string(e[0]) == "{"

		// do we have a node for e?
		nn := n.children.find(e, isParam)
		if nn == nil {
			// add new child node
			nn = &node{
				name: e,
				param: isParam,
				Methods: make(map[string]Scopes),
			}
			n.children = append(n.children, nn)
		}
		n = nn
	}

	return n, nil
}

// FindScopes returns scopes for a http method/path.
// Return error if method/path isn't found.
func (idx *Index) FindScopes(method, path string) (Scopes, error) {
	n, err := idx.Find(path)
	if err != nil {
		return nil, err
	}
	s, ok := n.Methods[method]
	if !ok {
		err = fmt.Errorf("no %s %s in index", method, path)
	}
	return s, err
}

// FindMethods returns http methods for a path.
// Return error if path isn't found.
func (idx *Index) FindMethods(path string) ([]string, error) {
	n, err := idx.Find(path)
	if err != nil {
		return nil, err
	}

	var m []string
	for k,_ := range n.Methods {
		m = append(m, k)
	}

	return m, nil
}

// Find returns a Node for a given path.
// Return error if path isn't found.
func (idx *Index) Find(path string) (*node, error) {
	elems := split(path)
	// Use elems to traverse the tree.
	n := idx.root
	var lastI int
	for i, e := range elems {
		lastI = i
		n = n.children.match(e) // TODO optimize
		if n == nil {
			return nil, fmt.Errorf("no match for %s", e)
		}
	}
	// We matched a series of path elements.
	// Now ensure all path elements are matched.
	if lastI+1 < len(elems) {
		return nil, fmt.Errorf("no match for %s in path %s", elems[lastI+1], path)
	}

	return n, nil
}

// Find node with name OR node with param true. Typically used during index building.
// Return node or nil.
func (ns nodes) find(name string, param bool) *node {
	for _,n := range ns {
		if n.param && param {
			return n
		}
		if n.name == name {
			return n
		}
	}
	return nil
}

// Find node with name or else a parameter node. Used during runtime path lookup.
// Return node or nil.
func (ns nodes) match(name string) *node {
	var r *node
	for _,n := range ns {
		if n.name == name {
			return n
		}
		if n.param {
			// param node will match any name for now. This will change when parameters are validated by type.
			r = n
		}
	}
	return r
}

// Split http path.
func split(path string) []string {
	// optionally remove leading /
	if len(path) > 0 && string(path[0]) == "/" {
		path = path[1:]
	}

	return strings.Split(path, "/")
}