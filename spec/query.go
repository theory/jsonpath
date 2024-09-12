package spec

import "strings"

// PathQuery represents a JSONPath expression.
type PathQuery struct {
	segments []*Segment
	root     bool
}

// Query returns a new query consisting of segments.
func Query(root bool, segments []*Segment) *PathQuery {
	return &PathQuery{root: root, segments: segments}
}

// Segments returns q's Segments.
func (q *PathQuery) Segments() []*Segment {
	return q.segments
}

// String returns a string representation of q.
func (q *PathQuery) String() string {
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
func (q *PathQuery) Select(current, root any) []any {
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
func (q *PathQuery) isSingular() bool {
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

// Singular returns a singularQuery variant of q if q [isSingular] returns true.
func (q *PathQuery) Singular() *SingularQueryExpr {
	if q.isSingular() {
		return singular(q)
	}

	return nil
}

// Expression returns a singularQuery variant of q if q [isSingular] returns
// true, and otherwise returns a filterQuery.
func (q *PathQuery) Expression() FunctionExprArg {
	if q.isSingular() {
		return singular(q)
	}

	return FilterQuery(q)
}

// singular is a utility function that converts q to a singularQuery.
func singular(q *PathQuery) *SingularQueryExpr {
	selectors := make([]Selector, len(q.segments))
	for i, s := range q.segments {
		selectors[i] = s.selectors[0]
	}
	return &SingularQueryExpr{selectors: selectors, relative: !q.root}
}
