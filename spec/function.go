package spec

//go:generate stringer -output function_string.go -linecomment -trimprefix Func -type FuncType,LogicalType

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FuncType describes the types of function parameters and results as defined
// by [RFC 9535 Section 2.4.1].
//
// Implements [fmt.Stringer].
//
// [RFC 9535 Section 2.4.1]: https://www.rfc-editor.org/rfc/rfc9535.html#section-2.4.1
type FuncType uint8

const (
	// FuncValue represents a JSON value as provided by [ValueType].
	FuncValue FuncType = iota + 1

	// FuncNodes represents a list of nodes as provided by [NodesType].
	FuncNodes

	// FuncLogical represents a logical value as provided by [LogicalType].
	FuncLogical
)

// PathValue defines the interface for JSONPath values used as comparison
// operands, filter expression results, and function parameters & return
// values.
//
// Implemented by the function types defined by [RFC 9535 Section 2.4.1]:
//   - [ValueType]
//   - [LogicalType]
//   - [NodesType]
//
// [RFC 9535 Section 2.4.1]: https://www.rfc-editor.org/rfc/rfc9535.html#section-2.4.1
type PathValue interface {
	stringWriter
	// FuncType returns the PathValue's [FuncType].
	FuncType() FuncType
}

// NodesType defines a node list (a list of JSON values) for a function
// expression parameters or results, as defined by [RFC 9535 Section 2.4.1].
// It can also be used in filter expressions. The underlying types should be
// string, integer, float, [json.Number], nil, true, false, []any, or
// map[string]any. Interfaces implemented:
//
// - [PathValue]
// - [fmt.Stringer]
//
// [RFC 9535 Section 2.4.1]: https://www.rfc-editor.org/rfc/rfc9535.html#section-2.4.1
type NodesType []any

// Nodes creates a NodesType that contains val, all of which should be the Go
// equivalent of the JSON data types: string, integer, float, [json.Number],
// nil, true, false, []any, or map[string]any.
func Nodes(val ...any) NodesType {
	return NodesType(val)
}

// FuncType returns [FuncNodes]. Defined by the [PathValue] interface.
func (NodesType) FuncType() FuncType { return FuncNodes }

// NodesFrom converts value to a [NodesType] and panics if it cannot. Use in
// [github.com/theory/jsonpath/registry.Registry.Register] [Evaluator]. Avoid
// the panic by returning an error from the accompanying [Validator] function
// when [FuncExprArg.ConvertsToNodes] returns false for the [FuncExprArg] that
// returns value.
//
// Converts each implementation of [PathValue] as follows:
//   - [NodesType]: returns value
//   - [ValueType]: Returns a [NodesType] containing that single value
//   - [LogicalType]: Panics
//   - nil: Returns an empty [NodesType]
func NodesFrom(value PathValue) NodesType {
	switch v := value.(type) {
	case NodesType:
		return v
	case *ValueType:
		return NodesType([]any{v.any})
	case nil:
		return NodesType([]any{})
	case LogicalType:
		panic("cannot convert LogicalType to NodesType")
	default:
		panic(fmt.Sprintf("unexpected argument of type %T", v))
	}
}

// writeTo writes the string representation of nt to buf. Defined by
// [stringWriter].
func (nt NodesType) writeTo(buf *strings.Builder) {
	buf.WriteString(nt.String())
}

// String returns the string representation of nt.
func (nt NodesType) String() string {
	return fmt.Sprintf("%v", []any(nt))
}

// LogicalType encapsulates a true or false value for a function expression
// parameters or results, as defined by [RFC 9535 Section 2.4.1]. Interfaces
// implemented:
//
//   - [PathValue]
//   - [fmt.Stringer]
//
// [RFC 9535 Section 2.4.1]: https://www.rfc-editor.org/rfc/rfc9535.html#section-2.4.1
type LogicalType uint8

const (
	// LogicalFalse represents a true [LogicalType].
	LogicalFalse LogicalType = iota // false

	// LogicalTrue represents a true [LogicalType].
	LogicalTrue // true
)

// Logical returns the LogicalType equivalent to boolean.
func Logical(boolean bool) LogicalType {
	if boolean {
		return LogicalTrue
	}
	return LogicalFalse
}

// Bool returns the boolean equivalent to lt.
func (lt LogicalType) Bool() bool { return lt == LogicalTrue }

// FuncType returns [FuncLogical]. Defined by the [PathValue] interface.
func (LogicalType) FuncType() FuncType { return FuncLogical }

// LogicalFrom converts value to a [LogicalType] and panics if it cannot. Use
// in [github.com/theory/jsonpath/registry.Registry.Register] [Evaluator]
// functions. Avoid the panic by returning an error from the accompanying
// [Validator] function when [FuncExprArg.ConvertsToLogical] returns false for
// the [FuncExprArg] that returns value.
//
// Converts each implementation of [PathValue] as follows:
//   - [LogicalType]: returns value
//   - [NodesType]: Returns [LogicalFalse] if value is empty and [LogicalTrue]
//     if it is not
//   - [ValueType]: Panics
//   - nil: Returns [LogicalFalse]
func LogicalFrom(value any) LogicalType {
	switch v := value.(type) {
	case LogicalType:
		return v
	case NodesType:
		return Logical(len(v) > 0)
	case nil:
		return LogicalFalse
	case *ValueType:
		panic("cannot convert ValueType to LogicalType")
	default:
		panic(fmt.Sprintf("unexpected argument of type %T", v))
	}
}

// writeTo writes a string representation of lt to buf. Defined by
// [stringWriter].
func (lt LogicalType) writeTo(buf *strings.Builder) {
	buf.WriteString(lt.String())
}

// ValueType encapsulates a JSON value for a function expression parameter or
// result, as defined by [RFC 9535 Section 2.4.1]. It can also be used as in
// filter expression. The underlying value should be a string, integer,
// [json.Number], float, nil, true, false, []any, or map[string]any. A nil
// ValueType pointer indicates no value. Interfaces implemented:
//
//   - [PathValue]
//   - [BasicExpr]
//   - [fmt.Stringer]
//
// [RFC 9535 Section 2.4.1]: https://www.rfc-editor.org/rfc/rfc9535.html#section-2.4.1
type ValueType struct {
	any
}

// Value returns a new [ValueType] for val, which must be the Go equivalent of
// a JSON data type: string, integer, float, [json.Number], nil, true, false,
// []any, or map[string]any.
func Value(val any) *ValueType {
	return &ValueType{val}
}

// Value returns the underlying value of vt.
func (vt *ValueType) Value() any { return vt.any }

// String returns the string representation of vt.
func (vt *ValueType) String() string { return fmt.Sprintf("%v", vt.any) }

// FuncType returns [FuncValue]. Defined by the [PathValue] interface.
func (*ValueType) FuncType() FuncType { return FuncValue }

// ValueFrom converts value to a [ValueType] and panics if it cannot. Use in
// [github.com/theory/jsonpath/registry.Registry.Register] [Evaluator]
// functions. Avoid the panic by returning an error from the accompanying
// [Validator] function when [FuncExprArg.ConvertsToValue] returns false for
// the [FuncExprArg] that returns value.
//
// Converts each implementation of [PathValue] as follows:
//   - [ValueType]: returns value
//   - [NodesType]: Panics
//   - [ValueType]: Panics
//   - nil: Returns nil
func ValueFrom(value PathValue) *ValueType {
	switch v := value.(type) {
	case *ValueType:
		return v
	case nil:
		return nil
	case LogicalType:
		panic("cannot convert LogicalType to ValueType")
	case NodesType:
		panic("cannot convert NodesType to ValueType")
	}
	panic(fmt.Sprintf("unexpected argument of type %T", value))
}

// Returns true if vt.any is truthy. Defined by the BasicExpr interface.
// Defined by [BasicExpr].
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
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f != float64(0)
		}
		return true
	default:
		return true
	}
}

// writeTo writes a string representation of vt to buf. Defined by
// [stringWriter].
func (vt *ValueType) writeTo(buf *strings.Builder) {
	buf.WriteString(vt.String())
}

// FuncExprArg defines the interface for function argument expressions.
// Implementations:
//
//   - [LogicalOr]
//   - [LiteralArg]
//   - [SingularQueryExpr]
//   - [PathQuery]
//   - [FuncExpr]
type FuncExprArg interface {
	stringWriter
	// evaluate evaluates the function expression against current and root and
	// returns the resulting PathValue.
	evaluate(current, root any) PathValue
	// ResultType returns the [FuncType] that defines the type of the return
	// value of the [FuncExprArg].
	ResultType() FuncType

	// ConvertsTo returns true if the function expression's result can be
	// converted to ft.
	ConvertsTo(ft FuncType) bool
}

// LiteralArg represents a literal JSON value, excluding objects and arrays.
// Its underlying value there must be one of string, integer, float,
// [json.Number], nil, true, or false.
//
// Interfaces implemented:
//   - [FuncExprArg]
//   - [CompVal]
//   - [fmt.Stringer]
type LiteralArg struct {
	// Number, string, bool, or null
	literal any
}

// Literal creates and returns a new [LiteralArg] consisting of lit, which
// must ge one of string, integer, float, [json.Number], nil, true, or false.
func Literal(lit any) *LiteralArg {
	return &LiteralArg{lit}
}

// Value returns the underlying value of la.
func (la *LiteralArg) Value() any { return la.literal }

// String returns the JSON string representation of la.
func (la *LiteralArg) String() string {
	if la.literal == nil {
		return "null"
	}
	return fmt.Sprintf("%#v", la.literal)
}

// evaluate returns a [ValueType] containing the literal value. Defined by the
// [FuncExprArg] interface.
func (la *LiteralArg) evaluate(_, _ any) PathValue {
	return &ValueType{la.literal}
}

// ResultType returns [FuncValue]. Defined by the [FuncExprArg] interface.
func (la *LiteralArg) ResultType() FuncType {
	return FuncValue
}

// ConvertsTo returns true if the result of the [LiteralArg] can be converted
// to ft.
func (*LiteralArg) ConvertsTo(ft FuncType) bool { return ft == FuncValue }

// writeTo writes a JSON string representation of la to buf. Defined by
// [stringWriter].
func (la *LiteralArg) writeTo(buf *strings.Builder) {
	if la.literal == nil {
		buf.WriteString("null")
	} else {
		fmt.Fprintf(buf, "%#v", la.literal)
	}
}

// asValue returns la.literal as a [ValueType]. Defined by the [CompVal]
// interface.
func (la *LiteralArg) asValue(_, _ any) PathValue {
	return &ValueType{la.literal}
}

// SingularQueryExpr represents a query that produces a single [ValueType]
// (JSON value) or nothing. Used in contexts that require a singular value,
// such as comparison operations and function arguments. Interfaces
// implemented:
//
//   - [CompVal]
//   - [FuncExprArg]
//   - [fmt.Stringer]
type SingularQueryExpr struct {
	// The kind of singular query, relative (from the current node) or
	// absolute (from the root node).
	relative bool
	// The query Name and/or Index selectors.
	selectors []Selector
}

// SingularQuery creates and returns a [SingularQueryExpr] that selects a
// single value at the path defined by selectors.
func SingularQuery(root bool, selectors ...Selector) *SingularQueryExpr {
	return &SingularQueryExpr{relative: !root, selectors: selectors}
}

// evaluate returns a [ValueType] containing the return value of executing sq.
// Defined by the [FuncExprArg] interface.
func (sq *SingularQueryExpr) evaluate(current, root any) PathValue {
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

// ResultType returns [FuncValue]. Defined by the [FuncExprArg] interface.
func (sq *SingularQueryExpr) ResultType() FuncType {
	return FuncValue
}

// ConvertsTo returns true if the result of the [SingularQueryExpr] can be
// converted to ft.
func (*SingularQueryExpr) ConvertsTo(ft FuncType) bool {
	return ft == FuncValue || ft == FuncNodes
}

// asValue returns the result of executing sq.execute against current and
// root. Defined by the [CompVal] interface.
func (sq *SingularQueryExpr) asValue(current, root any) PathValue {
	return sq.evaluate(current, root)
}

// writeTo writes a string representation of sq to buf. Defined by
// [stringWriter].
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

// String returns the string representation of sq.
func (sq *SingularQueryExpr) String() string {
	var buf strings.Builder
	sq.writeTo(&buf)
	return buf.String()
}

// Validator functions validate that the args expressions to a [FuncExtension]
// can be processed by the function.
type Validator func(args []FuncExprArg) error

// Evaluator functions execute a [FuncExtension] against the values returned
// by args and returns a result.
type Evaluator func(args []PathValue) PathValue

// FuncExtension defines a JSONPath function extension as defined in [RFC 9535
// Section 2.4]. Use [github.com/theory/jsonpath/registry.Registry.Register]
// to construct and register a new [FuncExtension].
//
// [RFC 9535 Section 2.4]: https://www.rfc-editor.org/rfc/rfc9535#name-function-extensions
type FuncExtension struct {
	// name is the name of the function. Must be unique among all functions in
	// a registry.
	name string

	// returnType defines the type of the function return value.
	returnType FuncType

	// validator executes at parse time to validate that all the args to
	// the function are compatible with the function.
	validator Validator

	// evaluator executes the function and returns the result of type
	// resultType.
	evaluator Evaluator
}

// Extension creates a new JSONPath function extension. Created by
// [github.com/theory/jsonpath/registry.Registry.Register]. The parameters
// are:
//
//   - name: the name of the function extension as used in JSONPath queries.
//   - returnType: The data type of the function return value.
//   - validator: A validation function that will be called at parse time
//     to validate that all the function args are compatible with the function.
//   - evaluator: The implementation of the function itself that executes
//     against args and returns the result of the type defined by resultType.
func Extension(name string, returnType FuncType, validator Validator, evaluator Evaluator) *FuncExtension {
	return &FuncExtension{name, returnType, validator, evaluator}
}

// Name returns the name of the [FuncExtension].
func (f *FuncExtension) Name() string { return f.name }

// ReturnType returns the data type of the [FuncExtension] return value.
func (f *FuncExtension) ReturnType() FuncType { return f.returnType }

// Evaluate executes the [FuncExtension] against args and returns a result of
// type [ResultType].
func (f *FuncExtension) Evaluate(args []PathValue) PathValue {
	return f.evaluator(args)
}

// Validate executes at parse time to validate that all the args to the are
// compatible with the [FuncExtension].
func (f *FuncExtension) Validate(args []FuncExprArg) error {
	return f.validator(args)
}

// FuncExpr represents a function expression, consisting of a named
// [FuncExtension] and its arguments. See
// [github.com/theory/jsonpath/registry] for an example defining a custom
// function. Interfaces Implemented:
//   - [FuncExprArg]
//   - [BasicExpr]
//   - [fmt.Stringer]
//   - [CompVal]
type FuncExpr struct {
	args []FuncExprArg
	fn   *FuncExtension
}

// Function creates an returns a new [FuncExpr] that will execute fn against
// the return values of args.
func Function(fn *FuncExtension, args ...FuncExprArg) *FuncExpr {
	return &FuncExpr{args: args, fn: fn}
}

// writeTo writes the string representation of fe to buf. Defined by
// [stringWriter].
func (fe *FuncExpr) writeTo(buf *strings.Builder) {
	buf.WriteString(fe.fn.Name() + "(")
	for i, arg := range fe.args {
		arg.writeTo(buf)
		if i < len(fe.args)-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteRune(')')
}

// String returns a string representation of fe.
func (fe *FuncExpr) String() string {
	var buf strings.Builder
	fe.writeTo(&buf)
	return buf.String()
}

// evaluate returns a [PathValue] containing the result of executing each
// [FuncExprArg] in fe (as passed to [Function]) and passing them to fe's
// [FuncExtension].
func (fe *FuncExpr) evaluate(current, root any) PathValue {
	res := []PathValue{}
	for _, a := range fe.args {
		res = append(res, a.evaluate(current, root))
	}

	return fe.fn.Evaluate(res)
}

// ResultType returns the result type of fe's [FuncExtension]. Defined by the
// [FuncExprArg] interface.
func (fe *FuncExpr) ResultType() FuncType {
	return fe.fn.ReturnType()
}

// ConvertsTo returns true if fe's result can be converted to ft.
func (fe *FuncExpr) ConvertsTo(ft FuncType) bool {
	return ft == fe.fn.ReturnType()
}

// asValue returns the result of executing fe.evaluate against current and
// root. Defined by the [CompVal] interface.
func (fe *FuncExpr) asValue(current, root any) PathValue {
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
// Returns false in all other cases. Defined by [BasicExpr].
func (fe *FuncExpr) testFilter(current, root any) bool {
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

// NotFuncExpr represents a negated function expression. It reverses the
// result of the return value of the underlying [FuncExpr]. Interfaces
// implemented:
//   - [BasicExpr]
//   - [fmt.Stringer]
type NotFuncExpr struct {
	*FuncExpr
}

// NotFunction creates and returns a new [NotFuncExpr] that will execute fn
// against the return values of args and return the inverse of its return
// value.
func NotFunction(fn *FuncExpr) NotFuncExpr {
	return NotFuncExpr{fn}
}

// String returns the string representation of nf.
func (nf NotFuncExpr) String() string {
	return "!" + nf.FuncExpr.String()
}

// testFilter returns the inverse of [FuncExpr.testFilter]. Defined by
// [BasicExpr].
func (nf NotFuncExpr) testFilter(current, root any) bool {
	return !nf.FuncExpr.testFilter(current, root)
}
