// Package jsonpath implements RFC 9535 JSONPath query expressions.
package jsonpath

import (
	"github.com/theory/jsonpath/parser"
	"github.com/theory/jsonpath/spec"
)

// Path represents a [RFC 9535] JSONPath query.
//
// [RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535.html
type Path struct {
	q *spec.PathQuery
}

// New creates and returns a new Path consisting of q.
func New(q *spec.PathQuery) *Path {
	return &Path{q: q}
}

// Parse parses path, a JSON Path query string, into a Path. Returns a
// PathParseError on parse failure.
//
//nolint:wrapcheck
func Parse(path string) (*Path, error) {
	q, err := parser.Parse(path)
	if err != nil {
		return nil, err
	}
	return New(q), nil
}

// String returns a string representation of p.
func (p *Path) String() string {
	return p.q.String()
}

// Query returns p's root Query.
func (p *Path) Query() *spec.PathQuery {
	return p.q
}

// Select executes the p query against input and returns the results.
func (p *Path) Select(input any) []any {
	return p.q.Select(nil, input)
}
