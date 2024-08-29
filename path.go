package jsonpath

// Path represents a [RFC 9535] JSONPath query.
//
// [RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535.html
type Path struct {
	q *Query
}

// New creates and returns a new Path consisting of q.
func New(q *Query) *Path {
	q.root = true
	return &Path{q: q}
}

// String returns a string representation of p.
func (p *Path) String() string {
	return p.q.String()
}

// Query returns p's root Query.
func (p *Path) Query() *Query {
	return p.q
}

// Select executes the p query against input and returns the results.
func (p *Path) Select(input any) []any {
	return p.q.Select(nil, input)
}
