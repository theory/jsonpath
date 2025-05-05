package spec

import (
	"strings"
)

// BasicExpr defines the basic interface for filter expressions.
// Implementations:
//
//   - [CompExpr]
//   - [ExistExpr]
//   - [FuncExpr]
//   - [LogicalAnd]
//   - [LogicalOr]
//   - [NonExistExpr]
//   - [NotFuncExpr]
//   - [NotParenExpr]
//   - [ParenExpr]
//   - [ValueType]
type BasicExpr interface {
	stringWriter
	// testFilter executes the filter expression on current and root and
	// returns true or false depending on the truthiness of its result.
	testFilter(current, root any) bool
}

// LogicalAnd represents a list of one or more expressions ANDed together by
// the && operator. Evaluates to true if all of its expressions evaluate to
// true. Short-circuits and returns false for the first expression that
// returns false. Interfaces implemented:
//
//   - [BasicExpr]
//   - [fmt.Stringer]
type LogicalAnd []BasicExpr

// And creates a LogicalAnd of all expr.
func And(expr ...BasicExpr) LogicalAnd {
	return LogicalAnd(expr)
}

// String returns the string representation of la.
func (la LogicalAnd) String() string {
	var buf strings.Builder
	la.writeTo(&buf)
	return buf.String()
}

// testFilter returns true if all of la's expressions return true.
// Short-circuits and returns false for the first expression that returns
// false. Defined by [BasicExpr].
func (la LogicalAnd) testFilter(current, root any) bool {
	for _, e := range la {
		if !e.testFilter(current, root) {
			return false
		}
	}
	return true
}

// writeTo writes the string representation of la to buf. Defined by
// [stringWriter].
func (la LogicalAnd) writeTo(buf *strings.Builder) {
	for i, e := range la {
		e.writeTo(buf)
		if i < len(la)-1 {
			buf.WriteString(" && ")
		}
	}
}

// LogicalOr represents a list of one or more expressions ORed together by the
// || operator. Evaluates to true if any of its expressions evaluates to true.
// Short-circuits and returns true for the first expression that returns true.
//
// Interfaces implemented:
//   - [BasicExpr]
//   - [FuncExprArg]
//   - [fmt.Stringer]
type LogicalOr []LogicalAnd

// Or returns a LogicalOr of all expr.
func Or(expr ...LogicalAnd) LogicalOr {
	return LogicalOr(expr)
}

// String returns the string representation of lo.
func (lo LogicalOr) String() string {
	var buf strings.Builder
	lo.writeTo(&buf)
	return buf.String()
}

// testFilter returns true if one of lo's expressions return true.
// Short-circuits and returns true for the first expression that returns true.
// Defined by [BasicExpr].
func (lo LogicalOr) testFilter(current, root any) bool {
	for _, e := range lo {
		if e.testFilter(current, root) {
			return true
		}
	}
	return false
}

// writeTo writes the string representation of lo to buf. Defined by
// [stringWriter].
func (lo LogicalOr) writeTo(buf *strings.Builder) {
	for i, e := range lo {
		e.writeTo(buf)
		if i < len(lo)-1 {
			buf.WriteString(" || ")
		}
	}
}

// evaluate evaluates lo and returns LogicalTrue when it returns true and
// LogicalFalse when it returns false. Defined by the [FuncExprArg]
// interface.
func (lo LogicalOr) evaluate(current, root any) PathValue {
	return Logical(lo.testFilter(current, root))
}

// ResultType returns [FuncLogical]. Defined by the [FuncExprArg] interface.
func (lo LogicalOr) ResultType() FuncType {
	return FuncLogical
}

// ConvertsTo returns true if the result of the [LogicalOr] can be converted
// to ft.
func (LogicalOr) ConvertsTo(ft FuncType) bool { return ft == FuncLogical }

// ParenExpr represents a parenthesized expression that groups the elements of
// a [LogicalOr]. Interfaces implemented (via the underlying [LogicalOr]):
//   - [BasicExpr]
//   - [FuncExprArg]
//   - [fmt.Stringer]
type ParenExpr struct {
	LogicalOr
}

// Paren returns a new ParenExpr that ORs the results of each expr.
func Paren(expr ...LogicalAnd) *ParenExpr {
	return &ParenExpr{LogicalOr: LogicalOr(expr)}
}

// writeTo writes a string representation of p to buf. Defined by
// [stringWriter].
func (p *ParenExpr) writeTo(buf *strings.Builder) {
	buf.WriteRune('(')
	p.LogicalOr.writeTo(buf)
	buf.WriteRune(')')
}

// String returns the string representation of p.
func (p *ParenExpr) String() string {
	var buf strings.Builder
	p.writeTo(&buf)
	return buf.String()
}

// NotParenExpr represents a negated parenthesized expression that groups the
// elements of a [LogicalOr]. Interfaces implemented (via the underlying
// [LogicalOr]):
//   - [BasicExpr]
//   - [FuncExprArg]
//   - [fmt.Stringer]
type NotParenExpr struct {
	LogicalOr
}

// NotParen returns a new NotParenExpr that ORs each expr.
func NotParen(expr ...LogicalAnd) *NotParenExpr {
	return &NotParenExpr{LogicalOr: LogicalOr(expr)}
}

// writeTo writes a string representation of p to buf. Defined by
// [stringWriter].
func (np *NotParenExpr) writeTo(buf *strings.Builder) {
	buf.WriteString("!(")
	np.LogicalOr.writeTo(buf)
	buf.WriteRune(')')
}

// String returns the string representation of np.
func (np *NotParenExpr) String() string {
	var buf strings.Builder
	np.writeTo(&buf)
	return buf.String()
}

// testFilter returns false if the np.LogicalOrExpression returns true and
// true if it returns false. Defined by [BasicExpr].
func (np *NotParenExpr) testFilter(current, root any) bool {
	return !np.LogicalOr.testFilter(current, root)
}

// ExistExpr represents a [PathQuery] used as a filter expression, in which
// context it returns true if the [PathQuery] selects at least one node.
// Interfaces implemented:
//   - [BasicExpr]
//   - [Selector] (via the underlying [PathQuery])
//   - [fmt.Stringer] (via the underlying [PathQuery])
type ExistExpr struct {
	*PathQuery
}

// Existence creates a new [ExistExpr] for q.
func Existence(q *PathQuery) *ExistExpr {
	return &ExistExpr{PathQuery: q}
}

// testFilter returns true if e.Query selects any results from current or
// root. Defined by [BasicExpr].
func (e *ExistExpr) testFilter(current, root any) bool {
	return len(e.Select(current, root)) > 0
}

// writeTo writes a string representation of e to buf. Defined by
// [stringWriter].
func (e *ExistExpr) writeTo(buf *strings.Builder) {
	buf.WriteString(e.String())
}

// NonExistExpr represents a negated [PathQuery] used as a filter expression,
// in which context it returns true if the [PathQuery] selects no nodes.
// Interfaces implemented:
//   - [BasicExpr]
//   - [Selector] (via the underlying [PathQuery])
//   - [fmt.Stringer] (via the underlying [PathQuery])
type NonExistExpr struct {
	*PathQuery
}

// Nonexistence creates a new [NonExistExpr] for q.
func Nonexistence(q *PathQuery) *NonExistExpr {
	return &NonExistExpr{PathQuery: q}
}

// writeTo writes a string representation of ne to buf. Defined by
// [stringWriter].
func (ne NonExistExpr) writeTo(buf *strings.Builder) {
	buf.WriteRune('!')
	buf.WriteString(ne.String())
}

// testFilter returns true if ne.Query selects no results from current or
// root. Defined by [BasicExpr].
func (ne NonExistExpr) testFilter(current, root any) bool {
	return len(ne.Select(current, root)) == 0
}
