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

// Select selects q.segments from current or root and returns the result.
// Returns just input if q has no segments. Defined by the [Selector]
// interface.
func (q *Query) Select(current, root any) []any {
	res := []any{current}
	if q.root {
		res[0] = root
	}
	for _, seg := range q.segments {
		segRes := []any{}
		for _, v := range res {
			segRes = append(segRes, seg.Select(v, root)...)
		}
		res = segRes
	}

	return res
}

// isSingular returns true if q always returns a singular value. Defined by
// the [Selector] interface.
func (q *Query) isSingular() bool {
	for _, s := range q.segments {
		if s.descendant {
			return false
		}
		if !s.isSingular() {
			return false
		}
	}
	return true
}

// singular returns a singularQuery variant of q if q [isSingular] returns true.
func (q *Query) singular() *singularQuery {
	if q.isSingular() {
		return singular(q)
	}

	return nil
}

// expression returns a singularQuery variant of q if q [isSingular] returns
// true, and otherwise returns a filterQuery.
//
//nolint:ireturn
func (q *Query) expression() FunctionExprArg {
	if q.isSingular() {
		return singular(q)
	}

	return &filterQuery{q}
}

// singular is a utility function that converts q to a singularQuery.
func singular(q *Query) *singularQuery {
	selectors := make([]Selector, len(q.segments))
	for i, s := range q.segments {
		selectors[i] = s.selectors[0]
	}
	return &singularQuery{selectors: selectors, relative: !q.root}
}
