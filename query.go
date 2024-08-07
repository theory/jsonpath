package jsonpath

import "strings"

// Query represents a JSONPath expression.
type Query struct {
	segments []*Segment
	root     bool
}

// NewQuery returns a new query consisting of segments.
func NewQuery(segments []*Segment) *Query {
	return &Query{segments: segments}
}

// Segments returns q's Segments.
func (q *Query) Segments() []*Segment {
	return q.segments
}

// String returns a string representation of q.
func (q *Query) String() string {
	buf := new(strings.Builder)
	if q.root {
		buf.WriteRune('$')
	} else {
		buf.WriteRune('@')
	}
	for _, s := range q.segments {
		buf.WriteString(s.String())
	}
	return buf.String()
}

// Select selects q.segments from input and returns the result. Returns just
// input if q has no segments.
func (q *Query) Select(input any) []any {
	res := []any{input}
	for _, seg := range q.segments {
		segRes := []any{}
		for _, v := range res {
			segRes = append(segRes, seg.Select(v)...)
		}
		res = segRes
	}

	return res
}
