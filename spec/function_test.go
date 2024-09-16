package spec

import (
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

func TestPathType(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		pType PathType
		str   string
	}{
		{PathValue, "ValueType"},
		{PathLogical, "LogicalType"},
		{PathNodes, "NodesType"},
		{PathType(16), "PathType(16)"},
	} {
		a.Equal(tc.str, tc.pType.String())
	}
}

func TestFuncType(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name  string
		fType FuncType
		ypv   []PathType
		npv   []PathType
	}{
		{
			name:  "FuncLiteral",
			fType: FuncLiteral,
			ypv:   []PathType{PathValue},
			npv:   []PathType{PathLogical, PathNodes},
		},
		{
			name:  "FuncSingularQuery",
			fType: FuncSingularQuery,
			ypv:   []PathType{PathValue, PathLogical, PathNodes},
			npv:   []PathType{},
		},
		{
			name:  "FuncValue",
			fType: FuncValue,
			ypv:   []PathType{PathValue},
			npv:   []PathType{PathLogical, PathNodes},
		},
		{
			name:  "FuncNodeList",
			fType: FuncNodeList,
			ypv:   []PathType{PathLogical, PathNodes},
			npv:   []PathType{PathValue},
		},
		{
			name:  "FuncLogical",
			fType: FuncLogical,
			ypv:   []PathType{PathLogical},
			npv:   []PathType{PathValue, PathNodes},
		},
		{
			name:  "FuncType(16)",
			fType: FuncType(16),
			ypv:   []PathType{},
			npv:   []PathType{PathLogical, PathValue, PathNodes},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.name, tc.fType.String())
			for _, pv := range tc.ypv {
				a.True(tc.fType.ConvertsTo(pv))
			}
			for _, pv := range tc.npv {
				a.False(tc.fType.ConvertsTo(pv))
			}
		})
	}
}

func TestNodesType(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		from JSONPathValue
		exp  NodesType
		err  string
	}{
		{"nodes", NodesType([]any{1, 2}), NodesType([]any{1, 2}), ""},
		{"value", Value(1), NodesType([]any{1}), ""},
		{"nil", nil, NodesType([]any{}), ""},
		{"logical", LogicalTrue, nil, "unexpected argument of type spec.LogicalType"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.err != "" {
				a.PanicsWithValue(tc.err, func() { NodesFrom(tc.from) })
				return
			}
			nt := NodesFrom(tc.from)
			a.Equal(tc.exp, nt)
			a.Equal(PathNodes, nt.PathType())
			a.Equal(FuncNodeList, nt.FuncType())
			a.Equal("NodesType", bufString(nt))
		})
	}
}

func TestLogicalType(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name    string
		from    JSONPathValue
		exp     LogicalType
		boolean bool
		err     string
		str     string
	}{
		{"true", LogicalTrue, LogicalTrue, true, "", "true"},
		{"false", LogicalFalse, LogicalFalse, false, "", "false"},
		{"unknown", LogicalType(16), LogicalType(16), false, "", "LogicalType(16)"},
		{"empty_nodes", NodesType([]any{}), LogicalFalse, false, "", "false"},
		{"nodes", NodesType([]any{1}), LogicalTrue, true, "", "true"},
		{"null", nil, LogicalFalse, false, "", "false"},
		{"value", Value(1), LogicalFalse, false, "unexpected argument of type *spec.ValueType", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.err != "" {
				a.PanicsWithValue(tc.err, func() { LogicalFrom(tc.from) })
				return
			}
			lt := LogicalFrom(tc.from)
			a.Equal(tc.exp, lt)
			a.Equal(PathLogical, lt.PathType())
			a.Equal(FuncLogical, lt.FuncType())
			a.Equal(tc.str, lt.String())
			a.Equal(tc.str, bufString(lt))
			a.Equal(tc.boolean, lt.Bool())
			if tc.boolean {
				a.Equal(LogicalTrue, LogicalFrom(tc.boolean))
			} else {
				a.Equal(LogicalFalse, LogicalFrom(tc.boolean))
			}
		})
	}
}

func TestValueType(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

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
			val := Value(tc.val)
			a.Equal(PathValue, val.PathType())
			a.Equal(FuncValue, val.FuncType())
			a.Equal(tc.val, val.Value())
			a.Equal("ValueType", bufString(val))
			a.Equal(tc.exp, val.testFilter(nil, nil))
		})
	}
}

func TestValueTypeFrom(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		val  JSONPathValue
		exp  *ValueType
		err  string
	}{
		{"valueType", Value(42), Value(42), ""},
		{"nil", nil, nil, ""},
		{"logical", LogicalFalse, nil, "unexpected argument of type spec.LogicalType"},
		{"nodes", NodesType([]any{1}), nil, "unexpected argument of type spec.NodesType"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.err != "" {
				a.PanicsWithValue(tc.err, func() { ValueFrom(tc.val) })
				return
			}
			val := ValueFrom(tc.val)
			a.Equal(tc.exp, val)
		})
	}
}

func TestJSONPathValueInterface(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name     string
		pathVal  any
		pathType PathType
		funcType FuncType
		str      string
	}{
		{"nodes", &NodesType{}, PathNodes, FuncNodeList, "NodesType"},
		{"logical", LogicalType(1), PathLogical, FuncLogical, "true"},
		{"value", &ValueType{}, PathValue, FuncValue, "ValueType"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Implements((*JSONPathValue)(nil), tc.pathVal)
			pv, _ := tc.pathVal.(JSONPathValue)
			a.Equal(tc.pathType, pv.PathType())
			a.Equal(tc.funcType, pv.FuncType())
			a.Equal(tc.str, bufString(pv))
		})
	}
}

func TestJsonFunctionExprArgInterface(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		expr any
	}{
		{"literal", &LiteralArg{}},
		{"filter_query", &FilterQueryExpr{}},
		{"singular_query", &SingularQueryExpr{}},
		{"logical_or", &LogicalOr{}},
		{"function_expr", &FunctionExpr{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Implements((*FunctionExprArg)(nil), tc.expr)
		})
	}
}

func TestJSONComparableValInterface(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		expr any
	}{
		{"literal", &LiteralArg{}},
		{"singular_query", &SingularQueryExpr{}},
		{"function_expr", &FunctionExpr{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Implements((*CompVal)(nil), tc.expr)
		})
	}
}

func TestLiteralArg(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

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
			lit := Literal(tc.literal)
			a.Equal(Value(tc.literal), lit.evaluate(nil, nil))
			a.Equal(Value(tc.literal), lit.asValue(nil, nil))
			a.Equal(tc.literal, lit.Value())
			a.Equal(FuncLiteral, lit.ResultType())
			a.Equal(tc.str, bufString(lit))
		})
	}
}

func TestSingularQuery(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name      string
		selectors []Selector
		input     any
		exp       JSONPathValue
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
			sq := &SingularQueryExpr{selectors: tc.selectors, relative: false}
			a.Equal(FuncSingularQuery, sq.ResultType())

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
	a := assert.New(t)

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
			query:    Query(true, []*Segment{Child(Name("x"))}),
			root:     map[string]any{"x": 42},
			exp:      []any{42},
			typeKind: FuncSingularQuery,
		},
		{
			name:     "current_name",
			query:    Query(false, []*Segment{Child(Name("x"))}),
			current:  map[string]any{"x": 42},
			exp:      []any{42},
			typeKind: FuncSingularQuery,
		},
		{
			name:     "root_name_index",
			query:    Query(true, []*Segment{Child(Name("x")), Child(Index(1))}),
			root:     map[string]any{"x": []any{19, 234}},
			exp:      []any{234},
			typeKind: FuncSingularQuery,
		},
		{
			name:     "root_slice",
			query:    Query(true, []*Segment{Child(Slice(0, 2))}),
			root:     []any{13, 2, 5},
			exp:      []any{13, 2},
			typeKind: FuncNodeList,
		},
		{
			name:     "current_wildcard",
			query:    Query(false, []*Segment{Child(Wildcard)}),
			current:  []any{13, 2, []any{4}},
			exp:      []any{13, 2, []any{4}},
			typeKind: FuncNodeList,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fq := &FilterQueryExpr{tc.query}
			a.Equal(tc.typeKind, fq.ResultType())
			a.Equal(NodesType(tc.exp), fq.evaluate(tc.current, tc.root))
			a.Equal(tc.query.String(), bufString(fq))
		})
	}
}

// Mock up a function.
type testFunc struct {
	name   string
	result FuncType
	eval   func(args []JSONPathValue) JSONPathValue
}

func (tf *testFunc) Name() string         { return tf.name }
func (tf *testFunc) ResultType() FuncType { return tf.result }
func (tf *testFunc) Evaluate(args []JSONPathValue) JSONPathValue {
	return tf.eval(args)
}

func newTrueFunc() *testFunc {
	return &testFunc{
		name:   "__true",
		result: FuncLogical,
		eval:   func([]JSONPathValue) JSONPathValue { return LogicalTrue },
	}
}

func newValueFunc(val any) *testFunc {
	return &testFunc{
		name:   "__val",
		result: FuncValue,
		eval:   func([]JSONPathValue) JSONPathValue { return Value(val) },
	}
}

func newNodesFunc() *testFunc {
	return &testFunc{
		name:   "__mk_nodes",
		result: FuncNodeList,
		eval: func(args []JSONPathValue) JSONPathValue {
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
	}
}

// Mock up a valid JSONPathValue that returns a new type.
type newValueType struct{}

func (newValueType) PathType() PathType         { return PathType(15) }
func (newValueType) FuncType() FuncType         { return FuncType(16) }
func (newValueType) writeTo(b *strings.Builder) { b.WriteString("FuncType(16)") }

func newTypeFunc() *testFunc {
	return &testFunc{
		name:   "__new_type",
		result: FuncType(16),
		eval:   func([]JSONPathValue) JSONPathValue { return newValueType{} },
	}
}

func TestFunctionExpr(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name    string
		fn      *testFunc
		args    []FunctionExprArg
		current any
		root    any
		exp     JSONPathValue
		logical bool
		str     string
	}{
		{
			name:    "val",
			fn:      newValueFunc(42),
			args:    []FunctionExprArg{SingularQuery(true, []Selector{Name("x")})},
			root:    map[string]any{"x": "xyz"},
			exp:     Value(42),
			logical: true,
			str:     `__val($["x"])`,
		},
		{
			name:    "__mk_nodes",
			fn:      newNodesFunc(),
			args:    []FunctionExprArg{Literal(42), Literal(99)},
			exp:     NodesType{42, 99},
			logical: true,
			str:     `__mk_nodes(42, 99)`,
		},
		{
			name:    "__mk_nodes_empty",
			fn:      newNodesFunc(),
			args:    []FunctionExprArg{},
			exp:     NodesType{},
			logical: false,
			str:     `__mk_nodes()`,
		},
		{
			name:    "__true",
			fn:      newTrueFunc(),
			args:    []FunctionExprArg{},
			exp:     LogicalTrue,
			logical: true,
			str:     `__true()`,
		},
		{
			name:    "__new_type",
			fn:      newTypeFunc(),
			args:    []FunctionExprArg{},
			exp:     newValueType{},
			logical: false,
			str:     `__new_type()`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fe := NewFunctionExpr(tc.fn, tc.args)
			a.Equal(tc.fn.result, fe.ResultType())
			a.Equal(tc.exp, fe.evaluate(tc.current, tc.root))
			a.Equal(tc.exp, fe.asValue(tc.current, tc.root))
			a.Equal(tc.logical, fe.testFilter(tc.current, tc.root))
			a.Equal(!tc.logical, NotFuncExpr{fe}.testFilter(tc.current, tc.root))
			a.Equal(tc.str, bufString(fe))
		})
	}
}
