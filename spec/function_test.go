package spec

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				a.True(tc.fType.convertsTo(pv))
			}
			for _, pv := range tc.npv {
				a.False(tc.fType.convertsTo(pv))
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
				a.PanicsWithValue(tc.err, func() { newNodesTypeFrom(tc.from) })
				return
			}
			nt := newNodesTypeFrom(tc.from)
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
				a.PanicsWithValue(tc.err, func() { newLogicalTypeFrom(tc.from) })
				return
			}
			lt := newLogicalTypeFrom(tc.from)
			a.Equal(tc.exp, lt)
			a.Equal(PathLogical, lt.PathType())
			a.Equal(FuncLogical, lt.FuncType())
			a.Equal(tc.str, lt.String())
			a.Equal(tc.str, bufString(lt))
			a.Equal(tc.boolean, lt.Bool())
			if tc.boolean {
				a.Equal(LogicalTrue, logicalFrom(tc.boolean))
			} else {
				a.Equal(LogicalFalse, logicalFrom(tc.boolean))
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
				a.PanicsWithValue(tc.err, func() { newValueTypeFrom(tc.val) })
				return
			}
			val := newValueTypeFrom(tc.val)
			a.Equal(tc.exp, val)
		})
	}
}

// Mock up a valid JSONPathValue that returns a new type.
type newValueType struct{}

func (newValueType) PathType() PathType         { return PathType(15) }
func (newValueType) FuncType() FuncType         { return FuncType(16) }
func (newValueType) writeTo(b *strings.Builder) { b.WriteString("FuncType(16)") }

func TestMain(m *testing.M) {
	// Set up a nodes-type returning function and some other type function.
	// Do it before running any tests so that the global list of functions
	// will remain static for the duration of the tests.
	Register(&Function{
		Name:       "__mk_nodes",
		ResultType: FuncNodeList,
		Validate:   func([]FunctionExprArg) error { return nil },
		Evaluate: func(args []JSONPathValue) JSONPathValue {
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
	})

	// Set up a function that returns a logical result with no args
	Register(&Function{
		Name:       "__true",
		ResultType: FuncLogical,
		Validate:   func([]FunctionExprArg) error { return nil },
		Evaluate:   func([]JSONPathValue) JSONPathValue { return LogicalTrue },
	})

	// Set up a function that returns an unknown return type.
	Register(&Function{
		Name:       "__new_type",
		ResultType: FuncType(16),
		Validate:   func([]FunctionExprArg) error { return nil },
		Evaluate:   func([]JSONPathValue) JSONPathValue { return newValueType{} },
	})

	m.Run()
}

func TestRegistry(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)
	a.Len(registry, 8)

	for _, tc := range []struct {
		name  string
		rType FuncType
		expr  []FunctionExprArg
		args  []JSONPathValue
		exp   any
	}{
		// RFC 9535-defined functions.
		{
			name:  "length",
			rType: FuncValue,
			expr:  []FunctionExprArg{Literal("foo")},
			args:  []JSONPathValue{Value("foo")},
			exp:   Value(3),
		},
		{
			name:  "count",
			rType: FuncValue,
			expr:  []FunctionExprArg{&SingularQueryExpr{}},
			args:  []JSONPathValue{NodesType([]any{1, 2})},
			exp:   Value(2),
		},
		{
			name:  "value",
			rType: FuncValue,
			expr:  []FunctionExprArg{&SingularQueryExpr{}},
			args:  []JSONPathValue{NodesType([]any{42})},
			exp:   Value(42),
		},
		{
			name:  "match",
			rType: FuncLogical,
			expr:  []FunctionExprArg{Literal("foo"), Literal(".*")},
			args:  []JSONPathValue{Value("foo"), Value(".*")},
			exp:   LogicalTrue,
		},
		{
			name:  "search",
			rType: FuncLogical,
			expr:  []FunctionExprArg{Literal("foo"), Literal(".")},
			args:  []JSONPathValue{Value("foo"), Value(".")},
			exp:   LogicalTrue,
		},
		// Test functions set up by TestMain()
		{
			name:  "__mk_nodes",
			rType: FuncNodeList,
			expr:  []FunctionExprArg{Literal("foo"), Literal(".")},
			args:  []JSONPathValue{Value("foo"), Value(".")},
			exp:   NodesType{"foo", "."},
		},
		{
			name:  "__true",
			rType: FuncLogical,
			expr:  []FunctionExprArg{},
			args:  []JSONPathValue{},
			exp:   LogicalTrue,
		},
		{
			name:  "__new_type",
			rType: FuncType(16),
			expr:  []FunctionExprArg{},
			args:  []JSONPathValue{},
			exp:   newValueType{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ft := GetFunction(tc.name)
			a.NotNil(ft)
			a.Equal(tc.rType, ft.ResultType)
			r.NoError(ft.Validate(tc.expr))
			a.Equal(tc.exp, ft.Evaluate(tc.args))
		})
	}
}

func TestRegisterErr(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		fn   *Function
		err  string
	}{
		{
			name: "nil_func",
			fn:   nil,
			err:  "jsonpath: Register function is nil",
		},
		{
			name: "existing_func",
			fn:   &Function{Name: "length"},
			err:  "jsonpath: Register called twice for function length",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.PanicsWithValue(tc.err, func() { Register(tc.fn) })
		})
	}
}

func TestLengthFunc(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		vals []JSONPathValue
		exp  int
		err  string
	}{
		{
			name: "empty_string",
			vals: []JSONPathValue{Value("")},
			exp:  0,
		},
		{
			name: "ascii_string",
			vals: []JSONPathValue{Value("abc def")},
			exp:  7,
		},
		{
			name: "unicode_string",
			vals: []JSONPathValue{Value("foö")},
			exp:  3,
		},
		{
			name: "emoji_string",
			vals: []JSONPathValue{Value("Hi 👋🏻")},
			exp:  5,
		},
		{
			name: "empty_array",
			vals: []JSONPathValue{Value([]any{})},
			exp:  0,
		},
		{
			name: "array",
			vals: []JSONPathValue{Value([]any{1, 2, 3, 4, 5})},
			exp:  5,
		},
		{
			name: "nested_array",
			vals: []JSONPathValue{Value([]any{1, 2, 3, "x", []any{456, 67}, true})},
			exp:  6,
		},
		{
			name: "empty_object",
			vals: []JSONPathValue{Value(map[string]any{})},
			exp:  0,
		},
		{
			name: "object",
			vals: []JSONPathValue{Value(map[string]any{"x": 1, "y": 0, "z": 2})},
			exp:  3,
		},
		{
			name: "nested_object",
			vals: []JSONPathValue{Value(map[string]any{
				"x": 1,
				"y": 0,
				"z": []any{1, 2},
				"a": map[string]any{"b": 9},
			})},
			exp: 4,
		},
		{
			name: "integer",
			vals: []JSONPathValue{Value(42)},
			exp:  -1,
		},
		{
			name: "bool",
			vals: []JSONPathValue{Value(true)},
			exp:  -1,
		},
		{
			name: "null",
			vals: []JSONPathValue{Value(nil)},
			exp:  -1,
		},
		{
			name: "nil",
			vals: []JSONPathValue{nil},
			exp:  -1,
		},
		{
			name: "not_value",
			vals: []JSONPathValue{LogicalFalse},
			err:  "unexpected argument of type spec.LogicalType",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.err != "" {
				a.PanicsWithValue(tc.err, func() {
					lengthFunc(tc.vals)
				})
				return
			}
			res := lengthFunc(tc.vals)
			if tc.exp < 0 {
				a.Nil(res)
			} else {
				a.Equal(Value(tc.exp), res)
			}
		})
	}
}

func TestCheckSingularFuncArgs(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	for _, tc := range []struct {
		name      string
		expr      []FunctionExprArg
		err       string
		lengthErr string
		countErr  string
		valueErr  string
	}{
		{
			name: "no_args",
			expr: []FunctionExprArg{},
			err:  "expected 1 argument but found 0",
		},
		{
			name: "two_args",
			expr: []FunctionExprArg{Literal(nil), Literal(nil)},
			err:  "expected 1 argument but found 2",
		},
		{
			name:     "literal_string",
			expr:     []FunctionExprArg{Literal(nil)},
			countErr: "cannot convert argument to PathNodes",
			valueErr: "cannot convert argument to PathNodes",
		},
		{
			name: "singular_query",
			expr: []FunctionExprArg{SingularQuery(false, nil)},
		},
		{
			name: "filter_query",
			expr: []FunctionExprArg{&filterQuery{
				Query(true, []*Segment{Child(Name("x"))}),
			}},
		},
		{
			name: "logical_function_expr",
			expr: []FunctionExprArg{&FunctionExpr{
				fn: registry["match"],
				args: []FunctionExprArg{&filterQuery{
					Query(true, []*Segment{Child(Name("x"))}),
				}},
			}},
			lengthErr: "cannot convert argument to ValueType",
			countErr:  "cannot convert argument to PathNodes",
			valueErr:  "cannot convert argument to PathNodes",
		},
		{
			name:      "logical_or",
			expr:      []FunctionExprArg{&LogicalOr{}},
			lengthErr: "cannot convert argument to ValueType",
			countErr:  "cannot convert argument to PathNodes",
			valueErr:  "cannot convert argument to PathNodes",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Test length args
			err := checkLengthArgs(tc.expr)
			switch {
			case tc.err != "":
				r.EqualError(err, tc.err)
			case tc.lengthErr != "":
				r.EqualError(err, tc.lengthErr)
			default:
				r.NoError(err)
			}

			// Test count args
			err = checkCountArgs(tc.expr)
			switch {
			case tc.err != "":
				r.EqualError(err, tc.err)
			case tc.countErr != "":
				r.EqualError(err, tc.countErr)
			default:
				r.NoError(err)
			}

			// Test value args
			err = checkValueArgs(tc.expr)
			switch {
			case tc.err != "":
				r.EqualError(err, tc.err)
			case tc.valueErr != "":
				r.EqualError(err, tc.valueErr)
			default:
				r.NoError(err)
			}
		})
	}
}

func TestCheckRegexFuncArgs(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	for _, tc := range []struct {
		name string
		expr []FunctionExprArg
		err  string
	}{
		{
			name: "no_args",
			expr: []FunctionExprArg{},
			err:  "expected 2 arguments but found 0",
		},
		{
			name: "one_arg",
			expr: []FunctionExprArg{Literal("hi")},
			err:  "expected 2 arguments but found 1",
		},
		{
			name: "three_args",
			expr: []FunctionExprArg{Literal("hi"), Literal("hi"), Literal("hi")},
			err:  "expected 2 arguments but found 3",
		},
		{
			name: "logical_or_1",
			expr: []FunctionExprArg{&LogicalOr{}, Literal("hi")},
			err:  "cannot convert argument 1 to PathNodes",
		},
		{
			name: "logical_or_2",
			expr: []FunctionExprArg{Literal("hi"), LogicalOr{}},
			err:  "cannot convert argument 2 to PathNodes",
		},
		{
			name: "singular_query_literal",
			expr: []FunctionExprArg{&SingularQueryExpr{}, Literal("hi")},
		},
		{
			name: "literal_singular_query",
			expr: []FunctionExprArg{Literal("hi"), &SingularQueryExpr{}},
		},
		{
			name: "filter_query_1",
			expr: []FunctionExprArg{
				&filterQuery{Query(true, []*Segment{Child(Name("x"))})},
				Literal("hi"),
			},
		},
		{
			name: "filter_query_2",
			expr: []FunctionExprArg{
				Literal("hi"),
				&filterQuery{Query(true, []*Segment{Child(Name("x"))})},
			},
		},
		{
			name: "function_expr_1",
			expr: []FunctionExprArg{&FunctionExpr{
				fn: registry["match"],
				args: []FunctionExprArg{&filterQuery{
					Query(true, []*Segment{Child(Name("x"))}),
				}},
			}, Literal("hi")},
			err: "cannot convert argument 1 to PathNodes",
		},
		{
			name: "function_expr_2",
			expr: []FunctionExprArg{Literal("hi"), &FunctionExpr{
				fn: registry["match"],
				args: []FunctionExprArg{&filterQuery{
					Query(true, []*Segment{Child(Name("x"))}),
				}},
			}},
			err: "cannot convert argument 2 to PathNodes",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Test match args
			err := checkMatchArgs(tc.expr)
			if tc.err == "" {
				r.NoError(err)
			} else {
				r.EqualError(err, strings.Replace(tc.err, "%v", "match", 1))
			}

			// Test search args
			err = checkSearchArgs(tc.expr)
			if tc.err == "" {
				r.NoError(err)
			} else {
				r.EqualError(err, strings.Replace(tc.err, "%v", "search", 1))
			}
		})
	}
}

func TestCountFunc(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		vals []JSONPathValue
		exp  int
		err  string
	}{
		{"empty", []JSONPathValue{NodesType([]any{})}, 0, ""},
		{"one", []JSONPathValue{NodesType([]any{1})}, 1, ""},
		{"three", []JSONPathValue{NodesType([]any{1, true, nil})}, 3, ""},
		{"not_nodes", []JSONPathValue{LogicalTrue}, 0, "unexpected argument of type spec.LogicalType"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.err != "" {
				a.PanicsWithValue(tc.err, func() { countFunc(tc.vals) })
				return
			}
			res := countFunc(tc.vals)
			if tc.exp < 0 {
				a.Nil(res)
			} else {
				a.Equal(Value(tc.exp), res)
			}
		})
	}
}

func TestValueFunc(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		vals []JSONPathValue
		exp  JSONPathValue
		err  string
	}{
		{"empty", []JSONPathValue{NodesType([]any{})}, nil, ""},
		{"one_int", []JSONPathValue{NodesType([]any{1})}, Value(1), ""},
		{"one_null", []JSONPathValue{NodesType([]any{nil})}, Value(nil), ""},
		{"one_string", []JSONPathValue{NodesType([]any{"x"})}, Value("x"), ""},
		{"three", []JSONPathValue{NodesType([]any{1, true, nil})}, nil, ""},
		{"not_nodes", []JSONPathValue{LogicalFalse}, nil, "unexpected argument of type spec.LogicalType"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.err != "" {
				a.PanicsWithValue(tc.err, func() { valueFunc(tc.vals) })
				return
			}
			a.Equal(tc.exp, valueFunc(tc.vals))
		})
	}
}

func TestRegexFuncs(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name   string
		input  *ValueType
		regex  *ValueType
		match  bool
		search bool
	}{
		{
			name:   "dot",
			input:  Value("x"),
			regex:  Value("."),
			match:  true,
			search: true,
		},
		{
			name:   "two_chars",
			input:  Value("xx"),
			regex:  Value("."),
			match:  false,
			search: true,
		},
		{
			name:   "multi_line_newline",
			input:  Value("xx\nyz"),
			regex:  Value(".*"),
			match:  false,
			search: true,
		},
		{
			name:   "multi_line_crlf",
			input:  Value("xx\r\nyz"),
			regex:  Value(".*"),
			match:  false,
			search: true,
		},
		{
			name:   "not_string_input",
			input:  Value(1),
			regex:  Value("."),
			match:  false,
			search: false,
		},
		{
			name:   "not_string_regex",
			input:  Value("x"),
			regex:  Value(1),
			match:  false,
			search: false,
		},
		{
			name:   "invalid_regex",
			input:  Value("x"),
			regex:  Value(".["),
			match:  false,
			search: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(logicalFrom(tc.match), matchFunc([]JSONPathValue{tc.input, tc.regex}))
			a.Equal(logicalFrom(tc.search), searchFunc([]JSONPathValue{tc.input, tc.regex}))
		})
	}
}

func TestExecRegexFuncs(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name   string
		vals   []JSONPathValue
		match  bool
		search bool
		err    string
	}{
		{
			name:   "dot",
			vals:   []JSONPathValue{Value("x"), Value("x")},
			match:  true,
			search: true,
		},
		{
			name: "first_not_value",
			vals: []JSONPathValue{NodesType{}, Value("x")},
			err:  "unexpected argument of type spec.NodesType",
		},
		{
			name: "second_not_value",
			vals: []JSONPathValue{Value("x"), LogicalFalse},
			err:  "unexpected argument of type spec.LogicalType",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.err == "" {
				a.Equal(matchFunc(tc.vals), logicalFrom(tc.match))
				a.Equal(searchFunc(tc.vals), logicalFrom(tc.search))
			} else {
				a.PanicsWithValue(tc.err, func() { matchFunc(tc.vals) })
				a.PanicsWithValue(tc.err, func() { searchFunc(tc.vals) })
			}
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
		{"filter_query", &filterQuery{}},
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
			a.Equal(Value(tc.literal), lit.execute(nil, nil))
			a.Equal(Value(tc.literal), lit.asValue(nil, nil))
			a.Equal(tc.literal, lit.Value())
			a.Equal(FuncLiteral, lit.asTypeKind())
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
			a.Equal(FuncSingularQuery, sq.asTypeKind())

			// Start with absolute query.
			a.False(sq.relative)
			a.Equal(tc.exp, sq.execute(nil, tc.input))
			a.Equal(tc.exp, sq.asValue(nil, tc.input))
			a.Equal("$"+tc.str, bufString(sq))

			// Try a relative query.
			sq.relative = true
			a.Equal(tc.exp, sq.execute(tc.input, nil))
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
			fq := &filterQuery{tc.query}
			a.Equal(tc.typeKind, fq.asTypeKind())
			a.Equal(NodesType(tc.exp), fq.execute(tc.current, tc.root))
			a.Equal(tc.query.String(), bufString(fq))
		})
	}
}

func TestFunctionExpr(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)
	rootX := Query(true, []*Segment{Descendant(Name("x"))})
	rootX.root = true

	for _, tc := range []struct {
		name    string
		fName   string
		args    []FunctionExprArg
		current any
		root    any
		exp     JSONPathValue
		logical bool
		str     string
		err     string
	}{
		{
			name:    "length_string",
			fName:   "length",
			args:    []FunctionExprArg{SingularQuery(true, []Selector{Name("x")})},
			root:    map[string]any{"x": "xyz"},
			exp:     Value(3),
			logical: true,
			str:     `length($["x"])`,
		},
		{
			name:    "length_slice",
			fName:   "length",
			args:    []FunctionExprArg{SingularQuery(true, []Selector{Name("x")})},
			root:    map[string]any{"x": []any{1, 2, 3, 4, 5}},
			exp:     Value(5),
			logical: true,
			str:     `length($["x"])`,
		},
		{
			name:    "count",
			fName:   "count",
			args:    []FunctionExprArg{&filterQuery{rootX}},
			root:    map[string]any{"x": map[string]any{"x": 1}},
			exp:     Value(2),
			logical: true,
			str:     `count($..["x"])`,
		},
		{
			name:    "value",
			fName:   "value",
			args:    []FunctionExprArg{SingularQuery(true, []Selector{Name("x")})},
			root:    map[string]any{"x": "xyz"},
			exp:     Value("xyz"),
			logical: true,
			str:     `value($["x"])`,
		},
		{
			name:  "match",
			fName: "match",
			args: []FunctionExprArg{
				SingularQuery(true, []Selector{Name("x")}),
				Literal("hi"),
			},
			root:    map[string]any{"x": "hi"},
			exp:     LogicalTrue,
			logical: true,
			str:     `match($["x"], "hi")`,
		},
		{
			name:  "search",
			fName: "search",
			args: []FunctionExprArg{
				SingularQuery(true, []Selector{Name("x")}),
				Literal("i"),
			},
			root:    map[string]any{"x": "hi"},
			exp:     LogicalTrue,
			logical: true,
			str:     `search($["x"], "i")`,
		},
		{
			name:  "invalid_args",
			fName: "count",
			args:  []FunctionExprArg{Literal("hi")},
			err:   "function count() cannot convert argument to PathNodes",
		},
		{
			name:  "unknown_func",
			fName: "nonesuch",
			args:  []FunctionExprArg{Literal("hi")},
			err:   "unknown function nonesuch()",
		},
		{
			name:    "__mk_nodes",
			fName:   "__mk_nodes",
			args:    []FunctionExprArg{Literal(42), Literal(99)},
			exp:     NodesType{42, 99},
			logical: true,
			str:     `__mk_nodes(42, 99)`,
		},
		{
			name:    "__mk_nodes_empty",
			fName:   "__mk_nodes",
			args:    []FunctionExprArg{},
			exp:     NodesType{},
			logical: false,
			str:     `__mk_nodes()`,
		},
		{
			name:    "__true",
			fName:   "__true",
			args:    []FunctionExprArg{},
			exp:     LogicalTrue,
			logical: true,
			str:     `__true()`,
		},
		{
			name:    "__new_type",
			fName:   "__new_type",
			args:    []FunctionExprArg{},
			exp:     newValueType{},
			logical: false,
			str:     `__new_type()`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fe, err := NewFunctionExpr(tc.fName, tc.args)
			if tc.err != "" {
				r.EqualError(err, tc.err)
				a.Nil(fe)
				return
			}
			a.Equal(registry[tc.fName].ResultType, fe.asTypeKind())
			a.Equal(tc.exp, fe.execute(tc.current, tc.root))
			a.Equal(tc.exp, fe.asValue(tc.current, tc.root))
			a.Equal(tc.logical, fe.testFilter(tc.current, tc.root))
			a.Equal(fe.fn.ResultType, fe.ResultType())
			a.Equal(!tc.logical, NotFuncExpr{fe}.testFilter(tc.current, tc.root))
			a.Equal(tc.str, bufString(fe))
		})
	}
}
