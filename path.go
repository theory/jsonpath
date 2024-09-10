// Package jsonpath implements RFC 9535 JSONPath query expressions.
package jsonpath

import "github.com/theory/jsonpath/spec"

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
