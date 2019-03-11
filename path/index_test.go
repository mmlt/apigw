package path

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
	"github.com/stretchr/testify/assert"
)

// TestAddPath
func TestAddPath(t *testing.T) {
	idx := NewIndex()
	idx.AddPath("version")
	idx.AddPath("accounts")
	idx.AddPath("accounts/{number}/positions/{id}")

	spew.Config.Indent = "  "
	spew.Config.DisablePointerAddresses = true
	spew.Config.DisableCapacities = true
	//TODO spew.Dump(idx)
}

// TestMatch show that we can do basic Find() operations.
func TestMatch(t *testing.T) {
	idx := NewIndex()
	idx.AddPath("version")
	idx.AddPath("accounts")
	idx.AddPath("accounts/{number}/positions/{id}")

	var err error

	// Basic queries
	_, err = idx.Find("version")
	assert.NoError(t, err)

	_, err = idx.Find("accounts/123/positions/45")
	assert.NoError(t, err)

	_, err = idx.Find("accounts/678")
	assert.NoError(t, err)

	// Queries that should error
	_, err = idx.Find("xxx")
	assert.Error(t, err)

	_, err = idx.Find("version/xxx")
	assert.Error(t, err)
}

// TestMatchAndUpdate shows that we can create an index from path and attach method/scope and retrieve it again..
func TestMatchAndUpdate(t *testing.T) {
	// create index
	idx := NewIndex()
	idx.AddPath("version")
	idx.AddPath("accounts")
	idx.AddPath("accounts/{number}/positions/{id}")

	testPaths := []string{
		"version","accounts/123/positions/45","accounts","accounts/123/positions","accounts/123",
	}
	// attach data at various places in index...
	for _, path := range testPaths {
		n, err := idx.Find(path)
		assert.NoError(t, err)
		n.Methods["GET"] = []string{path} // use path as data
	}
	// ...and check
	for _, path := range testPaths {
		n, err := idx.Find(path)
		assert.NoError(t, err)
		assert.ElementsMatch(t, n.Methods["GET"], []string{path})
	}
}


// TestIndexNew shows that we can create and index, add a static path with scopes and find it again.
func TestIndexNew(t *testing.T) {
	path := "/folders/a/files/echo.gif"
	idx := NewIndex()
	wantScopes := []string{"xyz"}
	idx.AddMethodPathScopes("GET", path, wantScopes)
	got, err := idx.FindScopes("GET", path)
	assert.NoError(t, err)
	assert.ElementsMatch(t, wantScopes, got)
}

var tests = []struct {
	path string
	method string
	scopes Scopes
}{
	{"/session", "DELETE", Scopes{"read"}},
	{"/version",  "GET", Scopes{}},
	{"/instruments","GET", Scopes{"read"}},
	{"/instruments/{id}","GET", Scopes{"read"}},
	{"/instruments/lists",  "GET", Scopes{"read"}},
	{"/instruments/lists/{id}",  "GET", Scopes{"read"}},
	{"/accounts", "GET", Scopes{"read"}},
	{"/accounts/{accountNumber}",  "GET", Scopes{"read"}},
	{"/accounts/{accountNumber}/balances", "GET", Scopes{"read"}},
	{"/accounts/{accountNumber}/transactions",  "GET", Scopes{"read"}},
	{"/accounts/{accountNumber}/positions", "GET", Scopes{"read"}},
	{"/accounts/{accountNumber}/positions/{id}",  "GET", Scopes{"read"}},
	{"/accounts/{accountNumber}/orders",  "GET", Scopes{"read"}},
	{"/accounts/{accountNumber}/orders", "POST", Scopes{"write"}},
	{"/accounts/{accountNumber}/performances",  "GET", Scopes{"read"}},
	{"/accounts/{accountNumber}/orders", "POST", Scopes{"write"}},
	{"/accounts/{accountNumber}/orders",  "GET", Scopes{"read"}},
	{"/accounts/{accountNumber}/orders/{number}", "GET", Scopes{"read"}},
	{"/accounts/{accountNumber}/orders/{number}",  "DELETE", Scopes{"write"}},
}

var methodQueries = []struct {
	path string
	methods []string
}{
	{"/session", []string{"DELETE"}},
	{"/version", []string{"GET"}},
	{"/instruments", []string{"GET"}},
	{"/instruments/00", []string{"GET"}},
	{"/instruments/lists/01",   []string{"GET"}},
	{"/instruments/lists",   []string{"GET"}},
	{"/accounts",  []string{"GET"}},
	{"/accounts/23",   []string{"GET"}},
	{"/accounts/45/balances",  []string{"GET"}},
	{"/accounts/67/transactions",   []string{"GET"}},
	{"/accounts/89/positions",  []string{"GET"}},
	{"/accounts/1011/positions/20",   []string{"GET"}},
	{"/accounts/1213/orders",   []string{"GET", "POST"}},
	{"/accounts/1415/performances",   []string{"GET"}},
	{"/accounts/1617/orders",  []string{"POST","GET"}},
	{"/accounts/1819/orders/21",  []string{"GET","DELETE"}},
}

func newTestIndex(t *testing.T) *Index {
	idx := NewIndex()
	for i, tst := range tests {
		_, err := idx.AddMethodPathScopes(tst.method, tst.path, tst.scopes)
		assert.NoError(t, err, "%d)", i)
	}
	return idx
}

// TestFindScopes shows that we can find scopes.
func TestFindScopes(t *testing.T) {
	idx := newTestIndex(t)

	// find (using the path with {} which is a bit of hack)
	for i, tst := range tests {
		gotScopes, err := idx.FindScopes(tst.method, tst.path)
		assert.NoError(t, err, "%d)", i)
		assert.ElementsMatch(t, tst.scopes, gotScopes, "%d)", i)
	}
}

// TestFindMethods shows that we can find methods.
func TestFindMethods(t *testing.T) {
	idx := newTestIndex(t)

	// find using methods queries
	for i, tst := range methodQueries {
		gotMethods, err := idx.FindMethods(tst.path)
		assert.NoError(t, err, "%d)", i)
		assert.ElementsMatch(t, tst.methods, gotMethods, "%d) want: %v got: %v for %s", i, tst.methods, gotMethods, tst.path)
	}
}

// TestNoMethods shows that querying for non-exsiting methods results in errors.
func TestNoMethods(t *testing.T) {
	idx := newTestIndex(t)
	var methodQueries = []struct {
		path string
		method string
	}{
		{"/session", "PUT"},
		{"/version", "PUT"},
		{"/instruments", "PUT"},
		{"/instruments/00", "PUT"},
		{"/instruments/lists/01", "PUT"},
		{"/instruments/lists", "PUT"},
	}

	for _, tst := range methodQueries {
		_, err := idx.FindScopes(tst.method, tst.path)
		assert.Error(t, err, "%s)", tst.path)
	}
}
