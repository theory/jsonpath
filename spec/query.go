package spec

import "strings"

// PathQuery represents a JSONPath query. Interfaces implemented:
//   - [Selector]
//   - [FuncExprArg]
//   - [fmt.Stringer]
type PathQuery struct {
	segments []*Segment
	root     bool
}

// Query returns a new [PathQuery] consisting of segments. When root is true
// it indicates a query from the root of a value. Set to false for filter
// subqueries.
func Query(root bool, segments ...*Segment) *PathQuery {
	return &PathQuery{root: root, segments: segments}
}

// Segments returns q's [Segment] values.
func (q *PathQuery) Segments() []*Segment {
	return q.segments
}

// String returns a string representation of q.
func (q *PathQuery) String() string {
	var buf strings.Builder
	q.writeTo(&buf)
	return buf.String()
}

// writeTo writes a string representation of q to buf. Defined by
// [stringWriter].
func (q *PathQuery) writeTo(buf *strings.Builder) {
	if q.root {
		buf.WriteRune('$')
	} else {
		buf.WriteRune('@')
	}
	for _, s := range q.segments {
		buf.WriteString(s.String())
	}
}

// Select selects the values from current or root and returns the results.
// Returns just current if q has no segments. Defined by the [Selector]
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

// SelectLocated values from current or root into [LocatedNode] values and
// returns the results. Returns just current if q has no segments. Defined by
// the [Selector] interface.
func (q *PathQuery) SelectLocated(current, root any, parent NormalizedPath) []*LocatedNode {
	res := []*LocatedNode{nil}
	if q.root {
		res[0] = newLocatedNode(nil, root)
	} else {
		res[0] = newLocatedNode(parent, current)
	}
	for _, seg := range q.segments {
		segRes := []*LocatedNode{}
		for _, v := range res {
			segRes = append(segRes, seg.SelectLocated(v.Node, root, v.Path)...)
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

// Singular returns the [SingularQueryExpr] variant of q if q is a singular
// query. Otherwise it returns nil.
func (q *PathQuery) Singular() *SingularQueryExpr {
	if q.isSingular() {
		return singular(q)
	}

	return nil
}

// Expression returns a [SingularQueryExpr] variant of q if q is a singular
// query, and otherwise returns q.
func (q *PathQuery) Expression() FuncExprArg {
	if q.isSingular() {
		return singular(q)
	}

	return q
}

// evaluate returns a [NodesType] containing the result of executing q.
// Defined by the [FuncExprArg] interface.
func (q *PathQuery) evaluate(current, root any) PathValue {
	return NodesType(q.Select(current, root))
}

// ResultType returns [FuncValue] if q is a singular query, and [FuncNodes]
// if it is not. Defined by the [FuncExprArg] interface.
func (q *PathQuery) ResultType() FuncType {
	if q.isSingular() {
		return FuncValue
	}
	return FuncNodes
}

// ConvertsTo returns true if q's result can be converted to ft. A singular
// query can be converted to either [FuncValue] or [FuncNodes]. All other
// queries can only be converted to FuncNodes.
func (q *PathQuery) ConvertsTo(ft FuncType) bool {
	if q.isSingular() {
		return ft == FuncValue || ft == FuncNodes
	}
	return ft == FuncNodes
}

// singular is a utility function that converts q to a singularQuery.
func singular(q *PathQuery) *SingularQueryExpr {
	selectors := make([]Selector, len(q.segments))
	for i, s := range q.segments {
		selectors[i] = s.selectors[0]
	}
	return &SingularQueryExpr{selectors: selectors, relative: !q.root}
}
