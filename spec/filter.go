package spec

import (
	"strings"
)

// BasicExpr defines the interface for filter expressions.
type BasicExpr interface {
	stringWriter
	// testFilter executes the filter expression on current and root and
	// returns true or false depending on the truthiness of its result.
	testFilter(current, root any) bool
}

// LogicalAnd represents a list of one or more expressions ANDed together
// by the && operator.
type LogicalAnd []BasicExpr

// testFilter returns true if all of la's expressions return true.
// Short-circuits and returns false for the first expression that returns
// false.
func (la LogicalAnd) testFilter(current, root any) bool {
	for _, e := range la {
		if !e.testFilter(current, root) {
			return false
		}
	}
	return true
}

// writeTo writes the string representation of la to buf.
func (la LogicalAnd) writeTo(buf *strings.Builder) {
	for i, e := range la {
		e.writeTo(buf)
		if i < len(la)-1 {
			buf.WriteString(" && ")
		}
	}
}

// LogicalOr represents a list of one or more expressions ORed together by
// the || operator.
type LogicalOr []LogicalAnd

func (lo LogicalOr) testFilter(current, root any) bool {
	for _, e := range lo {
		if e.testFilter(current, root) {
			return true
		}
	}
	return false
}

// writeTo writes the string representation of lo to buf.
func (lo LogicalOr) writeTo(buf *strings.Builder) {
	for i, e := range lo {
		e.writeTo(buf)
		if i < len(lo)-1 {
			buf.WriteString(" || ")
		}
	}
}

// evaluate evaluates lo and returns LogicalTrue when it returns true and
// LogicalFalse when it returns false. Defined by the [FunctionExprArg]
// interface.
func (lo LogicalOr) evaluate(current, root any) JSONPathValue {
	return LogicalFrom(lo.testFilter(current, root))
}

// ResultType returns FuncLogical. Defined by the [FunctionExprArg] interface.
func (lo LogicalOr) ResultType() FuncType {
	return FuncLogical
}

// ParenExpr represents a parenthesized expression.
type ParenExpr struct {
	LogicalOr
}

// Paren returns a new ParenExpr.
func Paren(or LogicalOr) *ParenExpr {
	return &ParenExpr{LogicalOr: or}
}

// writeTo writes a string representation of p to buf.
func (p *ParenExpr) writeTo(buf *strings.Builder) {
	buf.WriteRune('(')
	p.LogicalOr.writeTo(buf)
	buf.WriteRune(')')
}

// NotParenExpr represents a parenthesized expression preceded with a !.
type NotParenExpr struct {
	LogicalOr
}

// NotParen returns a new NotParenExpr.
func NotParen(or LogicalOr) *NotParenExpr {
	return &NotParenExpr{LogicalOr: or}
}

// writeTo writes a string representation of p to buf.
func (np *NotParenExpr) writeTo(buf *strings.Builder) {
	buf.WriteString("!(")
	np.LogicalOr.writeTo(buf)
	buf.WriteRune(')')
}

// testFilter returns false if the np.LogicalOrExpression returns true and
// true if it returns false.
func (np *NotParenExpr) testFilter(current, root any) bool {
	return !np.LogicalOr.testFilter(current, root)
}

// ExistExpr represents an existence expression.
type ExistExpr struct {
	*PathQuery
}

// Existence returns a new ExistExpr.
func Existence(q *PathQuery) *ExistExpr {
	return &ExistExpr{PathQuery: q}
}

// testFilter returns true if e.Query selects any results from current or
// root.
func (e *ExistExpr) testFilter(current, root any) bool {
	return len(e.Select(current, root)) > 0
}

// writeTo writes a string representation of e to buf.
func (e *ExistExpr) writeTo(buf *strings.Builder) {
	buf.WriteString(e.PathQuery.String())
}

// NotExistsExpr represents a nonexistence expression.
type NotExistsExpr struct {
	*PathQuery
}

// Nonexistence returns a new NotExistsExpr.
func Nonexistence(q *PathQuery) *NotExistsExpr {
	return &NotExistsExpr{PathQuery: q}
}

// writeTo writes a string representation of ne to buf.
func (ne NotExistsExpr) writeTo(buf *strings.Builder) {
	buf.WriteRune('!')
	buf.WriteString(ne.PathQuery.String())
}

// testFilter returns true if ne.Query selects no results from current or
// root.
func (ne NotExistsExpr) testFilter(current, root any) bool {
	return len(ne.Select(current, root)) == 0
}
