package spec

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func bufString(sw stringWriter) string {
	buf := new(strings.Builder)
	sw.writeTo(buf)
	return buf.String()
}

func TestFuncType(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		fType FuncType
	}{
		{
			name:  "Value",
			fType: FuncValue,
		},
		{
			name:  "Nodes",
			fType: FuncNodes,
		},
		{
			name:  "Logical",
			fType: FuncLogical,
		},
		{
			name:  "FuncType(16)",
			fType: FuncType(16),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.name, tc.fType.String())
		})
	}
}

func TestNodesType(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		from PathValue
		exp  NodesType
		str  string
		err  string
	}{
		{"nodes", NodesType([]any{1, 2}), Nodes(1, 2), "[1 2]", ""},
		{"value", Value(1), Nodes(1), "[1]", ""},
		{"nil", nil, Nodes([]any{}...), "[]", ""},
		{"logical", LogicalTrue, nil, "", "cannot convert LogicalType to NodesType"},
		{"unknown", newValueType{}, nil, "", "unexpected argument of type spec.newValueType"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			if tc.err != "" {
				a.PanicsWithValue(tc.err, func() { NodesFrom(tc.from) })
				return
			}
			nt := NodesFrom(tc.from)
			a.Equal(tc.exp, nt)
			a.Equal(FuncNodes, nt.FuncType())
			a.Equal(tc.str, bufString(nt))
		})
	}
}

func TestLogicalType(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		from    PathValue
		exp     LogicalType
		boolean bool
		err     string
		str     string
	}{
		{"true", LogicalTrue, LogicalTrue, true, "", "true"},
		{"false", LogicalFalse, LogicalFalse, false, "", "false"},
		{"unknown", LogicalType(16), LogicalType(16), false, "", "LogicalType(16)"},
		{"empty_nodes", Nodes(), LogicalFalse, false, "", "false"},
		{"nodes", Nodes(1), LogicalTrue, true, "", "true"},
		{"null", nil, LogicalFalse, false, "", "false"},
		{"value", Value(1), LogicalFalse, false, "cannot convert ValueType to LogicalType", ""},
		{"unknown", newValueType{}, LogicalFalse, false, "unexpected argument of type spec.newValueType", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			if tc.err != "" {
				a.PanicsWithValue(tc.err, func() { LogicalFrom(tc.from) })
				return
			}
			lt := LogicalFrom(tc.from)
			a.Equal(tc.exp, lt)
			a.Equal(FuncLogical, lt.FuncType())
			a.Equal(tc.str, lt.String())
			a.Equal(tc.str, bufString(lt))
			a.Equal(tc.boolean, lt.Bool())
			if tc.boolean {
				a.Equal(LogicalTrue, Logical(tc.boolean))
			} else {
				a.Equal(LogicalFalse, Logical(tc.boolean))
			}
		})
	}
}

func TestValueType(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		val  any
		exp  bool
	}{
		{"nil", nil, false},
		{"true", true, true},
		{"false", false, false},
		{"int", 42, true},
		{"int_zero", 0, false},
		{"int8", int8(1), true},
		{"int8_zero", int8(0), false},
		{"int16", int16(1), true},
		{"int16_zero", int16(0), false},
		{"int32", int32(1), true},
		{"int32_zero", int32(0), false},
		{"int64", int64(1), true},
		{"int64_zero", int64(0), false},
		{"uint", uint(42), true},
		{"uint_zero", 0, false},
		{"uint8", uint8(1), true},
		{"uint8_zero", uint8(0), false},
		{"uint16", uint16(1), true},
		{"uint16_zero", uint16(0), false},
		{"uint32", uint32(1), true},
		{"uint32_zero", uint32(0), false},
		{"json_number_int", json.Number("42"), true},
		{"json_number_zero", json.Number("0"), false},
		{"json_number_float", json.Number("98.6"), true},
		{"json_number_float_zero", json.Number("0.0"), false},
		{"json_number_invalid", json.Number("not a number"), true},
		{"uint64", uint64(1), true},
		{"uint64_zero", uint64(0), false},
		{"float32", float32(1), true},
		{"float32_zero", float32(0), false},
		{"float64", float64(1), true},
		{"float64_zero", float64(0), false},
		{"object", map[string]any{}, true},
		{"array", []any{}, true},
		{"struct", struct{}{}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			val := Value(tc.val)
			a.Equal(FuncValue, val.FuncType())
			a.Equal(tc.val, val.Value())
			a.Equal(fmt.Sprintf("%v", tc.val), bufString(val))
			a.Equal(fmt.Sprintf("%v", tc.val), val.String())
			a.Equal(tc.exp, val.testFilter(nil, nil))
		})
	}
}

func TestValueFrom(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		val  PathValue
		exp  *ValueType
		err  string
	}{
		{"valueType", Value(42), Value(42), ""},
		{"nil", nil, nil, ""},
		{"logical", LogicalFalse, nil, "cannot convert LogicalType to ValueType"},
		{"nodes", Nodes(1), nil, "cannot convert NodesType to ValueType"},
		{"unknown", newValueType{}, nil, "unexpected argument of type spec.newValueType"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			if tc.err != "" {
				a.PanicsWithValue(tc.err, func() { ValueFrom(tc.val) })
				return
			}
			val := ValueFrom(tc.val)
			a.Equal(tc.exp, val)
		})
	}
}

func TestPathValueInterface(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		pathVal  any
		funcType FuncType
		str      string
	}{
		{"nodes", Nodes(), FuncNodes, "[]"},
		{"logical", LogicalType(1), FuncLogical, "true"},
		{"value", &ValueType{}, FuncValue, "<nil>"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.Implements((*PathValue)(nil), tc.pathVal)
			pv, _ := tc.pathVal.(PathValue)
			a.Equal(tc.funcType, pv.FuncType())
			a.Equal(tc.str, bufString(pv))
			a.Equal(tc.str, pv.String())
		})
	}
}

func TestJsonFuncExprArgInterface(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		expr any
	}{
		{"literal", &LiteralArg{}},
		{"path_query", &PathQuery{}},
		{"singular_query", &SingularQueryExpr{}},
		{"logical_or", &LogicalOr{}},
		{"func_expr", &FuncExpr{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Implements(t, (*FuncExprArg)(nil), tc.expr)
		})
	}
}

func TestJSONComparableValInterface(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		expr any
	}{
		{"literal", &LiteralArg{}},
		{"singular_query", &SingularQueryExpr{}},
		{"func_expr", &FuncExpr{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Implements(t, (*CompVal)(nil), tc.expr)
		})
	}
}

func TestLiteralArg(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		literal any
		str     string
	}{
		{"string", "hi", `"hi"`},
		{"number", 42, "42"},
		{"true", true, "true"},
		{"false", false, "false"},
		{"null", nil, "null"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			lit := Literal(tc.literal)
			a.Equal(Value(tc.literal), lit.evaluate(nil, nil))
			a.Equal(Value(tc.literal), lit.asValue(nil, nil))
			a.Equal(tc.literal, lit.Value())
			a.Equal(FuncValue, lit.ResultType())
			a.Equal(tc.str, bufString(lit))
			a.Equal(tc.str, lit.String())
			a.True(lit.ConvertsTo(FuncValue))
			a.False(lit.ConvertsTo(FuncNodes))
			a.False(lit.ConvertsTo(FuncLogical))
		})
	}
}

func TestSingularQuery(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		selectors []Selector
		input     any
		exp       PathValue
		str       string
	}{
		{
			name:      "one_name",
			selectors: []Selector{Name("x")},
			input:     map[string]any{"x": 42},
			exp:       Value(42),
			str:       `["x"]`,
		},
		{
			name:      "two_names",
			selectors: []Selector{Name("x"), Name("y")},
			input:     map[string]any{"x": map[string]any{"y": 98.6}},
			exp:       Value(98.6),
			str:       `["x"]["y"]`,
		},
		{
			name:      "one_index",
			selectors: []Selector{Index(1)},
			input:     []any{"x", 42},
			exp:       Value(42),
			str:       `[1]`,
		},
		{
			name:      "two_indexes",
			selectors: []Selector{Index(1), Index(0)},
			input:     []any{"x", []any{true}},
			exp:       Value(true),
			str:       `[1][0]`,
		},
		{
			name:      "one_of_each",
			selectors: []Selector{Index(1), Name("x")},
			input:     []any{"x", map[string]any{"x": 12}},
			exp:       Value(12),
			str:       `[1]["x"]`,
		},
		{
			name:      "nonexistent",
			selectors: []Selector{Name("x")},
			input:     map[string]any{"y": 42},
			str:       `["x"]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			sq := &SingularQueryExpr{selectors: tc.selectors, relative: false}
			a.Equal(FuncValue, sq.ResultType())
			a.True(sq.ConvertsTo(FuncValue))
			a.True(sq.ConvertsTo(FuncNodes))
			a.False(sq.ConvertsTo(FuncLogical))

			// Start with absolute query.
			a.False(sq.relative)
			a.Equal(tc.exp, sq.evaluate(nil, tc.input))
			a.Equal(tc.exp, sq.asValue(nil, tc.input))
			a.Equal("$"+tc.str, bufString(sq))

			// Try a relative query.
			sq.relative = true
			a.Equal(tc.exp, sq.evaluate(tc.input, nil))
			a.Equal(tc.exp, sq.asValue(tc.input, nil))
			a.Equal("@"+tc.str, bufString(sq))
		})
	}
}

func TestFilterQuery(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		query    *PathQuery
		current  any
		root     any
		exp      []any
		typeKind FuncType
	}{
		{
			name:     "root_name",
			query:    Query(true, Child(Name("x"))),
			root:     map[string]any{"x": 42},
			exp:      []any{42},
			typeKind: FuncValue,
		},
		{
			name:     "current_name",
			query:    Query(false, Child(Name("x"))),
			current:  map[string]any{"x": 42},
			exp:      []any{42},
			typeKind: FuncValue,
		},
		{
			name:     "root_name_index",
			query:    Query(true, Child(Name("x")), Child(Index(1))),
			root:     map[string]any{"x": []any{19, 234}},
			exp:      []any{234},
			typeKind: FuncValue,
		},
		{
			name:     "root_slice",
			query:    Query(true, Child(Slice(0, 2))),
			root:     []any{13, 2, 5},
			exp:      []any{13, 2},
			typeKind: FuncNodes,
		},
		{
			name:     "current_wildcard",
			query:    Query(false, Child(Wildcard())),
			current:  []any{13, 2, []any{4}},
			exp:      []any{13, 2, []any{4}},
			typeKind: FuncNodes,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			fq := tc.query
			a.Equal(tc.typeKind, fq.ResultType())
			a.Equal(NodesType(tc.exp), fq.evaluate(tc.current, tc.root))
			a.Equal(tc.query.String(), bufString(fq))
			a.Equal(tc.typeKind == FuncValue, fq.ConvertsTo(FuncValue))
			a.True(fq.ConvertsTo(FuncNodes))
			a.False(fq.ConvertsTo(FuncLogical))
		})
	}
}

func newTrueFunc() *FuncExtension {
	return Extension(
		"__true",
		FuncLogical,
		func([]FuncExprArg) error { return nil },
		func([]PathValue) PathValue { return LogicalTrue },
	)
}

func newValueFunc(val any) *FuncExtension {
	return Extension(
		"__val",
		FuncValue,
		func([]FuncExprArg) error { return nil },
		func([]PathValue) PathValue { return Value(val) },
	)
}

func newNodesFunc() *FuncExtension {
	return Extension(
		"__mk_nodes",
		FuncNodes,
		func(args []FuncExprArg) error {
			for _, arg := range args {
				if !arg.ConvertsTo(FuncValue) {
					return fmt.Errorf("unexpected argument of type %v", arg.ResultType())
				}
			}
			return nil
		},
		func(args []PathValue) PathValue {
			ret := NodesType{}
			for _, x := range args {
				v, ok := x.(*ValueType)
				if !ok {
					panic(fmt.Sprintf("unexpected argument of type %T", x))
				}
				ret = append(ret, v.any)
			}
			return ret
		},
	)
}

// Mock up a valid PathValue that returns a new type.
type newValueType struct{}

func (newValueType) FuncType() FuncType         { return FuncType(16) }
func (newValueType) writeTo(b *strings.Builder) { b.WriteString("FuncType(16)") }
func (newValueType) String() string             { return "FuncType(16)" }

func newTypeFunc() *FuncExtension {
	return Extension(
		"__new_type",
		FuncType(16),
		func([]FuncExprArg) error { return nil },
		func([]PathValue) PathValue { return newValueType{} },
	)
}

func TestFunc(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		fn   *FuncExtension
		args []PathValue
		err  error
		exp  PathValue
	}{
		{
			name: "valid_err_value",
			fn: Extension(
				"xyz", FuncValue,
				func([]FuncExprArg) error { return errors.New("oops") },
				func([]PathValue) PathValue { return Value(42) },
			),
			args: []PathValue{},
			exp:  Value(42),
			err:  errors.New("oops"),
		},
		{
			name: "no_valid_err_nodes",
			fn: Extension(
				"abc", FuncNodes,
				func([]FuncExprArg) error { return nil },
				func([]PathValue) PathValue { return Nodes("hi") },
			),
			args: []PathValue{},
			exp:  Nodes("hi"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.Equal(tc.fn.name, tc.fn.Name())
			a.Equal(tc.err, tc.fn.Validate(nil))
			a.Equal(tc.exp, tc.fn.Evaluate(tc.args))
		})
	}
}

func TestFuncExpr(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		fn      *FuncExtension
		args    []FuncExprArg
		current any
		root    any
		exp     PathValue
		logical bool
		str     string
	}{
		{
			name:    "val",
			fn:      newValueFunc(42),
			args:    []FuncExprArg{SingularQuery(true, Name("x"))},
			root:    map[string]any{"x": "xyz"},
			exp:     Value(42),
			logical: true,
			str:     `__val($["x"])`,
		},
		{
			name:    "__mk_nodes",
			fn:      newNodesFunc(),
			args:    []FuncExprArg{Literal(42), Literal(99)},
			exp:     Nodes(42, 99),
			logical: true,
			str:     `__mk_nodes(42, 99)`,
		},
		{
			name:    "__mk_nodes_empty",
			fn:      newNodesFunc(),
			args:    []FuncExprArg{},
			exp:     Nodes([]any{}...),
			logical: false,
			str:     `__mk_nodes()`,
		},
		{
			name:    "__true",
			fn:      newTrueFunc(),
			args:    []FuncExprArg{},
			exp:     LogicalTrue,
			logical: true,
			str:     `__true()`,
		},
		{
			name:    "__new_type",
			fn:      newTypeFunc(),
			args:    []FuncExprArg{},
			exp:     newValueType{},
			logical: false,
			str:     `__new_type()`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			fe := Function(tc.fn, tc.args...)
			a.Equal(tc.fn.ReturnType(), fe.ResultType())
			a.Equal(tc.exp, fe.evaluate(tc.current, tc.root))
			a.Equal(tc.exp, fe.asValue(tc.current, tc.root))
			a.Equal(tc.logical, fe.testFilter(tc.current, tc.root))
			a.Equal(!tc.logical, NotFunction(fe).testFilter(tc.current, tc.root))
			a.Equal(tc.str, fe.String())
			a.Equal(tc.fn.ReturnType() == FuncValue, fe.ConvertsTo(FuncValue))
			a.Equal(tc.fn.ReturnType() == FuncNodes, fe.ConvertsTo(FuncNodes))
			a.Equal(tc.fn.ReturnType() == FuncLogical, fe.ConvertsTo(FuncLogical))
		})
	}
}
