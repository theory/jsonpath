package jsonpath

import (
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
		{"value", &ValueType{1}, NodesType([]any{1}), ""},
		{"nil", nil, NodesType([]any{}), ""},
		{"logical", LogicalTrue, nil, "unexpected argument of type jsonpath.LogicalType"},
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
		{"value", &ValueType{1}, LogicalFalse, false, "unexpected argument of type *jsonpath.ValueType", ""},
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
			val := &ValueType{tc.val}
			a.Equal(PathValue, val.PathType())
			a.Equal(FuncValue, val.FuncType())
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
		{"valueType", &ValueType{42}, &ValueType{42}, ""},
		{"nil", nil, nil, ""},
		{"logical", LogicalFalse, nil, "unexpected argument of type jsonpath.LogicalType"},
		{"nodes", NodesType([]any{1}), nil, "unexpected argument of type jsonpath.NodesType"},
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

//nolint:paralleltest
func TestRegistry(t *testing.T) {
	// Testing a global variable changed by other tests, so don't run in parallel.
	a := assert.New(t)
	r := require.New(t)
	a.Len(registry, 5)

	for _, tc := range []struct {
		name  string
		rType FuncType
		expr  []FunctionExprArg
		args  []JSONPathValue
		exp   any
	}{
		{
			name:  "length",
			rType: FuncValue,
			expr:  []FunctionExprArg{&literalArg{"foo"}},
			args:  []JSONPathValue{&ValueType{"foo"}},
			exp:   &ValueType{3},
		},
		{
			name:  "count",
			rType: FuncValue,
			expr:  []FunctionExprArg{&singularQuery{}},
			args:  []JSONPathValue{NodesType([]any{1, 2})},
			exp:   &ValueType{2},
		},
		{
			name:  "value",
			rType: FuncValue,
			expr:  []FunctionExprArg{&singularQuery{}},
			args:  []JSONPathValue{NodesType([]any{42})},
			exp:   &ValueType{42},
		},
		{
			name:  "match",
			rType: FuncLogical,
			expr:  []FunctionExprArg{&literalArg{"foo"}, &literalArg{".*"}},
			args:  []JSONPathValue{&ValueType{"foo"}, &ValueType{".*"}},
			exp:   LogicalTrue,
		},
		{
			name:  "search",
			rType: FuncLogical,
			expr:  []FunctionExprArg{&literalArg{"foo"}, &literalArg{"."}},
			args:  []JSONPathValue{&ValueType{"foo"}, &ValueType{"."}},
			exp:   LogicalTrue,
		},
	} {
		t.Run(tc.name, func(*testing.T) {
			ft := registry[tc.name]
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
			vals: []JSONPathValue{&ValueType{""}},
			exp:  0,
		},
		{
			name: "ascii_string",
			vals: []JSONPathValue{&ValueType{"abc def"}},
			exp:  7,
		},
		{
			name: "unicode_string",
			vals: []JSONPathValue{&ValueType{"foö"}},
			exp:  3,
		},
		{
			name: "emoji_string",
			vals: []JSONPathValue{&ValueType{"Hi 👋🏻"}},
			exp:  5,
		},
		{
			name: "empty_array",
			vals: []JSONPathValue{&ValueType{[]any{}}},
			exp:  0,
		},
		{
			name: "array",
			vals: []JSONPathValue{&ValueType{[]any{1, 2, 3, 4, 5}}},
			exp:  5,
		},
		{
			name: "nested_array",
			vals: []JSONPathValue{&ValueType{[]any{1, 2, 3, "x", []any{456, 67}, true}}},
			exp:  6,
		},
		{
			name: "empty_object",
			vals: []JSONPathValue{&ValueType{map[string]any{}}},
			exp:  0,
		},
		{
			name: "object",
			vals: []JSONPathValue{&ValueType{map[string]any{"x": 1, "y": 0, "z": 2}}},
			exp:  3,
		},
		{
			name: "nested_object",
			vals: []JSONPathValue{&ValueType{map[string]any{
				"x": 1,
				"y": 0,
				"z": []any{1, 2},
				"a": map[string]any{"b": 9},
			}}},
			exp: 4,
		},
		{
			name: "integer",
			vals: []JSONPathValue{&ValueType{42}},
			exp:  -1,
		},
		{
			name: "bool",
			vals: []JSONPathValue{&ValueType{true}},
			exp:  -1,
		},
		{
			name: "null",
			vals: []JSONPathValue{&ValueType{nil}},
			exp:  -1,
		},
		{
			name: "not_value",
			vals: []JSONPathValue{LogicalFalse},
			err:  "unexpected argument of type jsonpath.LogicalType",
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
				a.Equal(&ValueType{tc.exp}, res)
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
			err:  "jsonpath: expected 1 argument but received 0",
		},
		{
			name: "two_args",
			expr: []FunctionExprArg{&literalArg{}, &literalArg{}},
			err:  "jsonpath: expected 1 argument but received 2",
		},
		{
			name:     "literal_string",
			expr:     []FunctionExprArg{&literalArg{}},
			countErr: "jsonpath: expected argument to count() to be convertible to PathNodes but received FuncLiteral",
			valueErr: "jsonpath: expected argument to value() to be convertible to PathNodes but received FuncLiteral",
		},
		{
			name: "singular_query",
			expr: []FunctionExprArg{&singularQuery{}},
		},
		{
			name: "filter_query",
			expr: []FunctionExprArg{&filterQuery{NewQuery([]*Segment{Child(Name("x"))})}},
		},
		{
			name: "logical_function_expr",
			expr: []FunctionExprArg{&FunctionExpr{
				fn: registry["match"],
				args: []FunctionExprArg{&filterQuery{
					NewQuery([]*Segment{Child(Name("x"))}),
				}},
			}},
			lengthErr: "jsonpath: expected argument to length() to be convertible to PathValue but received FuncLogical",
			countErr:  "jsonpath: expected argument to count() to be convertible to PathNodes but received FuncLogical",
			valueErr:  "jsonpath: expected argument to value() to be convertible to PathNodes but received FuncLogical",
		},
		{
			name:      "logical_or",
			expr:      []FunctionExprArg{&LogicalOrExpr{}},
			lengthErr: "jsonpath: expected argument to length() to be convertible to PathValue but received FuncLogical",
			countErr:  "jsonpath: expected argument to count() to be convertible to PathNodes but received FuncLogical",
			valueErr:  "jsonpath: expected argument to value() to be convertible to PathNodes but received FuncLogical",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Test length args
			err := checkLengthArgs(tc.expr)
			switch {
			case tc.err != "":
				r.EqualError(err, tc.err)
				r.ErrorIs(err, ErrPathParse)
			case tc.lengthErr != "":
				r.EqualError(err, tc.lengthErr)
				r.ErrorIs(err, ErrPathParse)
			default:
				r.NoError(err)
			}

			// Test count args
			err = checkCountArgs(tc.expr)
			switch {
			case tc.err != "":
				r.EqualError(err, tc.err)
				r.ErrorIs(err, ErrPathParse)
			case tc.countErr != "":
				r.EqualError(err, tc.countErr)
				r.ErrorIs(err, ErrPathParse)
			default:
				r.NoError(err)
			}

			// Test value args
			err = checkValueArgs(tc.expr)
			switch {
			case tc.err != "":
				r.EqualError(err, tc.err)
				r.ErrorIs(err, ErrPathParse)
			case tc.valueErr != "":
				r.EqualError(err, tc.valueErr)
				r.ErrorIs(err, ErrPathParse)
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
			err:  "jsonpath: expected 2 arguments but received 0",
		},
		{
			name: "one_arg",
			expr: []FunctionExprArg{&literalArg{"hi"}},
			err:  "jsonpath: expected 2 arguments but received 1",
		},
		{
			name: "three_args",
			expr: []FunctionExprArg{&literalArg{"hi"}, &literalArg{"hi"}, &literalArg{"hi"}},
			err:  "jsonpath: expected 2 arguments but received 3",
		},
		{
			name: "logical_or_1",
			expr: []FunctionExprArg{&LogicalOrExpr{}, &literalArg{"hi"}},
			err:  "jsonpath: expected argument 1 to %v() to be convertible to PathValue but received FuncLogical",
		},
		{
			name: "logical_or_2",
			expr: []FunctionExprArg{&literalArg{"hi"}, &LogicalOrExpr{}},
			err:  "jsonpath: expected argument 2 to %v() to be convertible to PathValue but received FuncLogical",
		},
		{
			name: "singular_query_literal",
			expr: []FunctionExprArg{&singularQuery{}, &literalArg{"hi"}},
		},
		{
			name: "literal_singular_query",
			expr: []FunctionExprArg{&literalArg{"hi"}, &singularQuery{}},
		},
		{
			name: "filter_query_1",
			expr: []FunctionExprArg{&filterQuery{NewQuery([]*Segment{Child(Name("x"))})}, &literalArg{"hi"}},
		},
		{
			name: "filter_query_2",
			expr: []FunctionExprArg{&literalArg{"hi"}, &filterQuery{NewQuery([]*Segment{Child(Name("x"))})}},
		},
		{
			name: "function_expr_1",
			expr: []FunctionExprArg{&FunctionExpr{
				fn: registry["match"],
				args: []FunctionExprArg{&filterQuery{
					NewQuery([]*Segment{Child(Name("x"))}),
				}},
			}, &literalArg{"hi"}},
			err: "jsonpath: expected argument 1 to %v() to be convertible to PathValue but received FuncLogical",
		},
		{
			name: "function_expr_2",
			expr: []FunctionExprArg{&literalArg{"hi"}, &FunctionExpr{
				fn: registry["match"],
				args: []FunctionExprArg{&filterQuery{
					NewQuery([]*Segment{Child(Name("x"))}),
				}},
			}},
			err: "jsonpath: expected argument 2 to %v() to be convertible to PathValue but received FuncLogical",
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
				r.ErrorIs(err, ErrPathParse)
			}

			// Test search args
			err = checkSearchArgs(tc.expr)
			if tc.err == "" {
				r.NoError(err)
			} else {
				r.EqualError(err, strings.Replace(tc.err, "%v", "search", 1))
				r.ErrorIs(err, ErrPathParse)
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
		{"not_nodes", []JSONPathValue{LogicalTrue}, 0, "unexpected argument of type jsonpath.LogicalType"},
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
				a.Equal(&ValueType{tc.exp}, res)
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
		{"one_int", []JSONPathValue{NodesType([]any{1})}, &ValueType{1}, ""},
		{"one_null", []JSONPathValue{NodesType([]any{nil})}, &ValueType{nil}, ""},
		{"one_string", []JSONPathValue{NodesType([]any{"x"})}, &ValueType{"x"}, ""},
		{"three", []JSONPathValue{NodesType([]any{1, true, nil})}, nil, ""},
		{"not_nodes", []JSONPathValue{LogicalFalse}, nil, "unexpected argument of type jsonpath.LogicalType"},
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
			input:  &ValueType{"x"},
			regex:  &ValueType{"."},
			match:  true,
			search: true,
		},
		{
			name:   "two_chars",
			input:  &ValueType{"xx"},
			regex:  &ValueType{"."},
			match:  false,
			search: true,
		},
		{
			name:   "multi_line_newline",
			input:  &ValueType{"xx\nyz"},
			regex:  &ValueType{".*"},
			match:  true,
			search: true,
		},
		{
			name:   "multi_line_crlf",
			input:  &ValueType{"xx\r\nyz"},
			regex:  &ValueType{".*"},
			match:  true,
			search: true,
		},
		{
			name:   "not_string_input",
			input:  &ValueType{1},
			regex:  &ValueType{"."},
			match:  false,
			search: false,
		},
		{
			name:   "not_string_regex",
			input:  &ValueType{"x"},
			regex:  &ValueType{1},
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
			vals:   []JSONPathValue{&ValueType{"x"}, &ValueType{"x"}},
			match:  true,
			search: true,
		},
		{
			name: "first_not_value",
			vals: []JSONPathValue{NodesType{}, &ValueType{"x"}},
			err:  "unexpected argument of type jsonpath.NodesType",
		},
		{
			name: "second_not_value",
			vals: []JSONPathValue{&ValueType{"x"}, LogicalFalse},
			err:  "unexpected argument of type jsonpath.LogicalType",
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
		{"literal", &literalArg{}},
		{"filter_query", &filterQuery{}},
		{"singular_query", &singularQuery{}},
		{"logical_or", &LogicalOrExpr{}},
		{"function_expr", &FunctionExpr{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Implements((*FunctionExprArg)(nil), tc.expr)
		})
	}
}

func TestJsoncomparableValInterface(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		expr any
	}{
		{"literal", &literalArg{}},
		{"singular_query", &singularQuery{}},
		{"function_expr", &FunctionExpr{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Implements((*comparableVal)(nil), tc.expr)
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
			lit := &literalArg{tc.literal}
			a.Equal(&ValueType{tc.literal}, lit.execute(nil, nil))
			a.Equal(&ValueType{tc.literal}, lit.asValue(nil, nil))
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
			exp:       &ValueType{42},
			str:       `["x"]`,
		},
		{
			name:      "two_names",
			selectors: []Selector{Name("x"), Name("y")},
			input:     map[string]any{"x": map[string]any{"y": 98.6}},
			exp:       &ValueType{98.6},
			str:       `["x"]["y"]`,
		},
		{
			name:      "one_index",
			selectors: []Selector{Index(1)},
			input:     []any{"x", 42},
			exp:       &ValueType{42},
			str:       `[1]`,
		},
		{
			name:      "two_indexes",
			selectors: []Selector{Index(1), Index(0)},
			input:     []any{"x", []any{true}},
			exp:       &ValueType{true},
			str:       `[1][0]`,
		},
		{
			name:      "one_of_each",
			selectors: []Selector{Index(1), Name("x")},
			input:     []any{"x", map[string]any{"x": 12}},
			exp:       &ValueType{12},
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
			sq := &singularQuery{selectors: tc.selectors, relative: false}
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
		query    *Query
		current  any
		root     any
		exp      []any
		typeKind FuncType
	}{
		{
			name:     "root_name",
			query:    &Query{segments: []*Segment{Child(Name("x"))}, root: true},
			root:     map[string]any{"x": 42},
			exp:      []any{42},
			typeKind: FuncSingularQuery,
		},
		{
			name:     "current_name",
			query:    &Query{segments: []*Segment{Child(Name("x"))}, root: false},
			current:  map[string]any{"x": 42},
			exp:      []any{42},
			typeKind: FuncSingularQuery,
		},
		{
			name:     "root_name_index",
			query:    &Query{segments: []*Segment{Child(Name("x")), Child(Index(1))}, root: true},
			root:     map[string]any{"x": []any{19, 234}},
			exp:      []any{234},
			typeKind: FuncSingularQuery,
		},
		{
			name:     "root_slice",
			query:    &Query{segments: []*Segment{Child(Slice(0, 2))}, root: true},
			root:     []any{13, 2, 5},
			exp:      []any{13, 2},
			typeKind: FuncNodeList,
		},
		{
			name:     "current_wildcard",
			query:    &Query{segments: []*Segment{Child(Wildcard)}, root: false},
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
	rootX := NewQuery([]*Segment{Descendant(Name("x"))})
	rootX.root = true

	// Set up a nodes-type returning function and some other type function.
	Register(&Function{
		Name:       "mk_nodes",
		ResultType: FuncNodeList,
		Validate:   func([]FunctionExprArg) error { return nil },
		Evaluate: func(args []JSONPathValue) JSONPathValue {
			ret := NodesType{}
			for _, x := range args {
				v, ok := x.(*ValueType)
				if !ok {
					t.Fatalf("unexpected argument of type %T", x)
				}
				ret = append(ret, v.any)
			}
			return ret
		},
	})

	// Set up a function that returns an unknown return type.
	Register(&Function{
		Name:       "new_type",
		ResultType: FuncType(16),
		Validate:   func([]FunctionExprArg) error { return nil },
		Evaluate:   func([]JSONPathValue) JSONPathValue { return newValueType{} },
	})

	// Remove the test functions from the registry
	t.Cleanup(func() {
		registryMu.Lock()
		defer registryMu.Unlock()
		delete(registry, "mk_nodes")
		delete(registry, "new_type")
	})

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
			args:    []FunctionExprArg{&singularQuery{selectors: []Selector{Name("x")}}},
			root:    map[string]any{"x": "xyz"},
			exp:     &ValueType{3},
			logical: true,
			str:     `length($["x"])`,
		},
		{
			name:    "length_slice",
			fName:   "length",
			args:    []FunctionExprArg{&singularQuery{selectors: []Selector{Name("x")}}},
			root:    map[string]any{"x": []any{1, 2, 3, 4, 5}},
			exp:     &ValueType{5},
			logical: true,
			str:     `length($["x"])`,
		},
		{
			name:    "count",
			fName:   "count",
			args:    []FunctionExprArg{&filterQuery{rootX}},
			root:    map[string]any{"x": map[string]any{"x": 1}},
			exp:     &ValueType{2},
			logical: true,
			str:     `count($..["x"])`,
		},
		{
			name:    "value",
			fName:   "value",
			args:    []FunctionExprArg{&singularQuery{selectors: []Selector{Name("x")}}},
			root:    map[string]any{"x": "xyz"},
			exp:     &ValueType{"xyz"},
			logical: true,
			str:     `value($["x"])`,
		},
		{
			name:  "match",
			fName: "match",
			args: []FunctionExprArg{
				&singularQuery{selectors: []Selector{Name("x")}},
				&literalArg{"hi"},
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
				&singularQuery{selectors: []Selector{Name("x")}},
				&literalArg{"i"},
			},
			root:    map[string]any{"x": "hi"},
			exp:     LogicalTrue,
			logical: true,
			str:     `search($["x"], "i")`,
		},
		{
			name:  "invalid_args",
			fName: "count",
			args:  []FunctionExprArg{&literalArg{"hi"}},
			err:   "jsonpath: expected argument to count() to be convertible to PathNodes but received FuncLiteral",
		},
		{
			name:  "unknown_func",
			fName: "nonesuch",
			args:  []FunctionExprArg{&literalArg{"hi"}},
			err:   "jsonpath: unknown jsonpath function nonesuch()",
		},
		{
			name:    "mk_nodes",
			fName:   "mk_nodes",
			args:    []FunctionExprArg{&literalArg{42}, &literalArg{99}},
			exp:     NodesType{42, 99},
			logical: true,
			str:     `mk_nodes(42, 99)`,
		},
		{
			name:    "mk_nodes_empty",
			fName:   "mk_nodes",
			args:    []FunctionExprArg{},
			exp:     NodesType{},
			logical: false,
			str:     `mk_nodes()`,
		},
		{
			name:    "new_type",
			fName:   "new_type",
			args:    []FunctionExprArg{},
			exp:     newValueType{},
			logical: false,
			str:     `new_type()`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fe, err := NewFunctionExpr(tc.fName, tc.args)
			if tc.err != "" {
				r.EqualError(err, tc.err)
				r.ErrorIs(err, ErrPathParse)
				a.Nil(fe)
				return
			}
			a.Equal(registry[tc.fName].ResultType, fe.asTypeKind())
			a.Equal(tc.exp, fe.execute(tc.current, tc.root))
			a.Equal(tc.exp, fe.asValue(tc.current, tc.root))
			a.Equal(tc.logical, fe.testFilter(tc.current, tc.root))
			a.Equal(!tc.logical, NotFuncExpr{fe}.testFilter(tc.current, tc.root))
			a.Equal(tc.str, bufString(fe))
		})
	}
}

type newValueType struct{}

func (newValueType) PathType() PathType       { return PathType(15) }
func (newValueType) FuncType() FuncType       { return FuncType(16) }
func (newValueType) writeTo(*strings.Builder) {}
