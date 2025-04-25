package spec

//go:generate stringer -linecomment -output function_string.go -type LogicalType,PathType,FuncType

import (
	"fmt"
	"strings"
)

// PathType represents the types of filter expression values.
type PathType uint8

//revive:disable:exported
const (
	// A type containing a single value.
	PathValue PathType = iota + 1 // ValueType

	// A logical (boolean) type.
	PathLogical // LogicalType

	// A type containing a list of nodes.
	PathNodes // NodesType
)

// FuncType defines the function argument expressions and return types defined
// by [RFC 9535]. Function extensions check that these types can be converted
// to [spec.PathType] values for evaluation.
//
// [RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535.html
type FuncType uint8

const (
	// FuncLiteral represents a literal JSON value.
	FuncLiteral FuncType = iota + 1 // FuncLiteral

	// FuncSingularQuery represents a value from a singular query.
	FuncSingularQuery // FuncSingularQuery

	// FuncValue represents a JSON value, used to represent functions that
	// return [ValueType].
	FuncValue // FuncValue

	// FuncNodeList represents a node list, either from a filter query argument, or a function that
	// returns [NodesType].
	FuncNodeList // FuncNodeList

	// FuncLogical represents a logical, either from a logical expression, or
	// from a function that returns [LogicalType].
	FuncLogical // FuncLogical
)

// ConvertsTo returns true if a function argument of type ft can be converted
// to pv.
func (ft FuncType) ConvertsTo(pv PathType) bool {
	switch ft {
	case FuncLiteral, FuncValue:
		return pv == PathValue
	case FuncSingularQuery:
		return true
	case FuncNodeList:
		return pv != PathValue
	case FuncLogical:
		return pv == PathLogical
	default:
		return false
	}
}

// JSONPathValue defines the interface for JSON path values.
type JSONPathValue interface {
	stringWriter
	// PathType returns the JSONPathValue's PathType.
	PathType() PathType
	// FuncType returns the JSONPathValue's FuncType.
	FuncType() FuncType
}

// NodesType defines the JSONPath type representing a node list; in other
// words, a list of JSON values.
type NodesType []any

// PathType returns PathNodes. Defined by the JSONPathValue interface.
func (NodesType) PathType() PathType { return PathNodes }

// FuncType returns FuncNodeList. Defined by the JSONPathValue interface.
func (NodesType) FuncType() FuncType { return FuncNodeList }

// NodesFrom attempts to convert value to a NodesType and panics if it cannot.
func NodesFrom(value JSONPathValue) NodesType {
	switch v := value.(type) {
	case NodesType:
		return v
	case *ValueType:
		return NodesType([]any{v.any})
	case nil:
		return NodesType([]any{})
	default:
		panic(fmt.Sprintf("unexpected argument of type %T", v))
	}
}

// writeTo writes a string representation of the NodesType to buf.
func (NodesType) writeTo(buf *strings.Builder) {
	buf.WriteString("NodesType")
}

// LogicalType is a JSONPath type that represents true or false.
type LogicalType uint8

const (
	LogicalFalse LogicalType = iota // false
	LogicalTrue                     // true
)

// Bool returns the boolean equivalent to lt.
func (lt LogicalType) Bool() bool { return lt == LogicalTrue }

// PathType returns PathLogical. Defined by the JSONPathValue interface.
func (LogicalType) PathType() PathType { return PathLogical }

// FuncType returns FuncLogical. Defined by the JSONPathValue interface.
func (LogicalType) FuncType() FuncType { return FuncLogical }

// LogicalFrom attempts to convert value to a LogicalType and panics if it
// cannot.
func LogicalFrom(value any) LogicalType {
	switch v := value.(type) {
	case LogicalType:
		return v
	case NodesType:
		return LogicalFrom(len(v) > 0)
	case bool:
		if v {
			return LogicalTrue
		}
		return LogicalFalse
	case nil:
		return LogicalFalse
	default:
		panic(fmt.Sprintf("unexpected argument of type %T", v))
	}
}

// writeTo writes a string representation of lt to buf.
func (lt LogicalType) writeTo(buf *strings.Builder) {
	buf.WriteString(lt.String())
}

// ValueType encapsulates a JSON value, which should be a string, integer,
// float, nil, true, false, []any, or map[string]any. A nil ValueType pointer
// indicates no value.
type ValueType struct {
	any
}

// Value returns a new ValueType.
func Value(val any) *ValueType {
	return &ValueType{val}
}

// Value returns the underlying value of vt.
func (vt *ValueType) Value() any { return vt.any }

// PathType returns PathValue. Defined by the JSONPathValue interface.
func (*ValueType) PathType() PathType { return PathValue }

// FuncType returns FuncValue. Defined by the JSONPathValue interface.
func (*ValueType) FuncType() FuncType { return FuncValue }

// ValueFrom attempts to convert value to a ValueType and panics if it cannot.
func ValueFrom(value JSONPathValue) *ValueType {
	switch v := value.(type) {
	case *ValueType:
		return v
	case nil:
		return nil
	}
	panic(fmt.Sprintf("unexpected argument of type %T", value))
}

// Returns true if vt.any is truthy. Defined by the BasicExpr interface.
func (vt *ValueType) testFilter(_, _ any) bool {
	switch v := vt.any.(type) {
	case nil:
		return false
	case bool:
		return v
	case int:
		return v != 0
	case int8:
		return v != int8(0)
	case int16:
		return v != int16(0)
	case int32:
		return v != int32(0)
	case int64:
		return v != int64(0)
	case uint:
		return v != 0
	case uint8:
		return v != uint8(0)
	case uint16:
		return v != uint16(0)
	case uint32:
		return v != uint32(0)
	case uint64:
		return v != uint64(0)
	case float32:
		return v != float32(0)
	case float64:
		return v != float64(0)
	default:
		return true
	}
}

// writeTo writes a string representation of vt to buf.
func (vt *ValueType) writeTo(buf *strings.Builder) {
	buf.WriteString("ValueType")
}

// FunctionExprArg defines the interface for function argument expressions.
type FunctionExprArg interface {
	stringWriter
	// evaluate evaluates the function expression against current and root and
	// returns the resulting JSONPathValue.
	evaluate(current, root any) JSONPathValue
	// ResultType returns the FuncType that defines the type of the return
	// value of JSONPathValue.
	ResultType() FuncType
}

// LiteralArg represents a literal JSON value, excluding objects and arrays.
type LiteralArg struct {
	// Number, string, bool, or null
	literal any
}

// Literal creates and returns a new LiteralArg.
func Literal(lit any) *LiteralArg {
	return &LiteralArg{lit}
}

// Value returns the underlying value of la.
func (la *LiteralArg) Value() any { return la.literal }

// evaluate returns a [ValueType] containing the literal value. Defined by the
// [FunctionExprArg] interface.
func (la *LiteralArg) evaluate(_, _ any) JSONPathValue {
	return &ValueType{la.literal}
}

// ResultType returns FuncLiteral. Defined by the [FunctionExprArg] interface.
func (la *LiteralArg) ResultType() FuncType {
	return FuncLiteral
}

// writeTo writes a string representation of la to buf.
func (la *LiteralArg) writeTo(buf *strings.Builder) {
	if la.literal == nil {
		buf.WriteString("null")
	} else {
		fmt.Fprintf(buf, "%#v", la.literal)
	}
}

// asValue returns la.literal as a [ValueType]. Defined by the [comparableVal]
// interface.
func (la *LiteralArg) asValue(_, _ any) JSONPathValue {
	return &ValueType{la.literal}
}

// SingularQueryExpr represents a query that produces a single node (JSON value),
// or nothing.
type SingularQueryExpr struct {
	// The kind of singular query, relative (from the current node) or
	// absolute (from the root node).
	relative bool
	// The query Name and/or Index selectors.
	selectors []Selector
}

// SingularQuery creates and returns a SingularQueryExpr.
func SingularQuery(root bool, selectors []Selector) *SingularQueryExpr {
	return &SingularQueryExpr{relative: !root, selectors: selectors}
}

// evaluate returns a [ValueType] containing the return value of executing sq.
// Defined by the [FunctionExprArg] interface.
func (sq *SingularQueryExpr) evaluate(current, root any) JSONPathValue {
	target := root
	if sq.relative {
		target = current
	}

	for _, seg := range sq.selectors {
		res := seg.Select(target, nil)
		if len(res) == 0 {
			return nil
		}
		target = res[0]
	}

	return &ValueType{target}
}

// ResultType returns FuncSingularQuery. Defined by the [FunctionExprArg]
// interface.
func (*SingularQueryExpr) ResultType() FuncType {
	return FuncSingularQuery
}

// asValue returns the result of executing sq.execute against current and root.
// Defined by the [comparableVal] interface.
func (sq *SingularQueryExpr) asValue(current, root any) JSONPathValue {
	return sq.evaluate(current, root)
}

// writeTo writes a string representation of sq to buf.
func (sq *SingularQueryExpr) writeTo(buf *strings.Builder) {
	if sq.relative {
		buf.WriteRune('@')
	} else {
		buf.WriteRune('$')
	}

	for _, seg := range sq.selectors {
		buf.WriteRune('[')
		seg.writeTo(buf)
		buf.WriteRune(']')
	}
}

// FilterQueryExpr represents a JSONPath Query used in a filter expression.
type FilterQueryExpr struct {
	*PathQuery
}

// FilterQuery creates and returns a new FilterQueryExpr.
func FilterQuery(q *PathQuery) *FilterQueryExpr {
	return &FilterQueryExpr{q}
}

// evaluate returns a [NodesType] containing the result of executing fq.
// Defined by the [FunctionExprArg] interface.
func (fq *FilterQueryExpr) evaluate(current, root any) JSONPathValue {
	return NodesType(fq.Select(current, root))
}

// ResultType returns FuncSingularQuery if fq is a singular query, and
// FuncNodeList if it is not. Defined by the [FunctionExprArg] interface.
func (fq *FilterQueryExpr) ResultType() FuncType {
	if fq.isSingular() {
		return FuncSingularQuery
	}
	return FuncNodeList
}

// writeTo writes a string representation of fq to buf.
func (fq *FilterQueryExpr) writeTo(buf *strings.Builder) {
	buf.WriteString(fq.String())
}

// FunctionExpr represents a function expression, consisting of a named
// function and its arguments.
type FunctionExpr struct {
	args []FunctionExprArg
	fn   PathFunction
}

// PathFunction represents a JSONPath function. See
// [github.com/theory/jsonpath/registry] for the implementation.
type PathFunction interface {
	Name() string
	ResultType() FuncType
	Evaluate(args []JSONPathValue) JSONPathValue
}

// Function creates an returns a new function expression that will execute fn
// against the return values of args.
func Function(fn PathFunction, args []FunctionExprArg) *FunctionExpr {
	return &FunctionExpr{args: args, fn: fn}
}

// writeTo writes the string representation of fe to buf.
func (fe *FunctionExpr) writeTo(buf *strings.Builder) {
	buf.WriteString(fe.fn.Name() + "(")
	for i, arg := range fe.args {
		arg.writeTo(buf)
		if i < len(fe.args)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteRune(')')
}

// evaluate returns a [NodesType] containing the results of executing each
// argument in fe.args. Defined by the [FunctionExprArg] interface.
func (fe *FunctionExpr) evaluate(current, root any) JSONPathValue {
	res := []JSONPathValue{}
	for _, a := range fe.args {
		res = append(res, a.evaluate(current, root))
	}

	return fe.fn.Evaluate(res)
}

// ResultType returns the result type of the registered function named
// fe.name. Defined by the [FunctionExprArg] interface.
func (fe *FunctionExpr) ResultType() FuncType {
	return fe.fn.ResultType()
}

// asValue returns the result of executing fe.execute against current and root.
// Defined by the [comparableVal] interface.
func (fe *FunctionExpr) asValue(current, root any) JSONPathValue {
	return fe.evaluate(current, root)
}

// testFilter executes fe and returns true if the function returns a truthy
// value:
//
//   - If the result is [NodesType], returns true if it is not empty.
//   - If the result is [*ValueType], returns true if its underlying value
//     is truthy.
//   - If the result is [LogicalType], returns the underlying boolean.
//
// Returns false in all other cases.
func (fe *FunctionExpr) testFilter(current, root any) bool {
	switch res := fe.evaluate(current, root).(type) {
	case NodesType:
		return len(res) > 0
	case *ValueType:
		return res.testFilter(current, root)
	case LogicalType:
		return res.Bool()
	default:
		return false
	}
}

// NotFuncExpr represents a "!func()" expression. It reverses the result of
// the return value of a function expression.
type NotFuncExpr struct {
	*FunctionExpr
}

// NotFunction creates an returns a new NotFuncExpr that will execute fn
// against the return values of args and return the inverses of its return
// value.
func NotFunction(fn *FunctionExpr) NotFuncExpr {
	return NotFuncExpr{fn}
}

// testFilter returns the inverse of nf.FunctionExpr.testFilter().
func (nf NotFuncExpr) testFilter(current, root any) bool {
	return !nf.FunctionExpr.testFilter(current, root)
}
