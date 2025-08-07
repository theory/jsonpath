package registry

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theory/jsonpath/spec"
)

func TestLengthFunc(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
		vals []spec.PathValue
		exp  int
		err  string
	}{
		{
			test: "empty_string",
			vals: []spec.PathValue{spec.Value("")},
			exp:  0,
		},
		{
			test: "ascii_string",
			vals: []spec.PathValue{spec.Value("abc def")},
			exp:  7,
		},
		{
			test: "unicode_string",
			vals: []spec.PathValue{spec.Value("foö")},
			exp:  3,
		},
		{
			test: "emoji_string",
			vals: []spec.PathValue{spec.Value("Hi 👋🏻")},
			exp:  5,
		},
		{
			test: "empty_array",
			vals: []spec.PathValue{spec.Value([]any{})},
			exp:  0,
		},
		{
			test: "array",
			vals: []spec.PathValue{spec.Value([]any{1, 2, 3, 4, 5})},
			exp:  5,
		},
		{
			test: "nested_array",
			vals: []spec.PathValue{spec.Value([]any{1, 2, 3, "x", []any{456, 67}, true})},
			exp:  6,
		},
		{
			test: "empty_object",
			vals: []spec.PathValue{spec.Value(map[string]any{})},
			exp:  0,
		},
		{
			test: "object",
			vals: []spec.PathValue{spec.Value(map[string]any{"x": 1, "y": 0, "z": 2})},
			exp:  3,
		},
		{
			test: "nested_object",
			vals: []spec.PathValue{spec.Value(map[string]any{
				"x": 1,
				"y": 0,
				"z": []any{1, 2},
				"a": map[string]any{"b": 9},
			})},
			exp: 4,
		},
		{
			test: "integer",
			vals: []spec.PathValue{spec.Value(42)},
			exp:  -1,
		},
		{
			test: "bool",
			vals: []spec.PathValue{spec.Value(true)},
			exp:  -1,
		},
		{
			test: "null",
			vals: []spec.PathValue{spec.Value(nil)},
			exp:  -1,
		},
		{
			test: "nil",
			vals: []spec.PathValue{nil},
			exp:  -1,
		},
		{
			test: "not_value",
			vals: []spec.PathValue{spec.LogicalFalse},
			err:  "cannot convert LogicalType to ValueType",
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

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
				a.Equal(spec.Value(tc.exp), res)
			}
		})
	}
}

func TestCheckSingularFuncArgs(t *testing.T) {
	t.Parallel()
	reg := New()

	for _, tc := range []struct {
		test      string
		expr      []spec.FuncExprArg
		err       string
		lengthErr string
		countErr  string
		valueErr  string
	}{
		{
			test: "no_args",
			expr: []spec.FuncExprArg{},
			err:  "expected 1 argument but found 0",
		},
		{
			test: "two_args",
			expr: []spec.FuncExprArg{spec.Literal(nil), spec.Literal(nil)},
			err:  "expected 1 argument but found 2",
		},
		{
			test:     "literal_string",
			expr:     []spec.FuncExprArg{spec.Literal(nil)},
			countErr: "cannot convert argument to Nodes",
			valueErr: "cannot convert argument to Nodes",
		},
		{
			test: "singular_query",
			expr: []spec.FuncExprArg{spec.SingularQuery(false, nil)},
		},
		{
			test: "nodes_query",
			expr: []spec.FuncExprArg{
				spec.Query(true, spec.Child(spec.Name("x"))),
			},
		},
		{
			test: "logical_func_expr",
			expr: []spec.FuncExprArg{spec.Function(reg.Get("match"),
				spec.Query(true, spec.Child(spec.Name("x"))),
				spec.Literal("hi"),
			)},
			lengthErr: "cannot convert argument to Value",
			countErr:  "cannot convert argument to Nodes",
			valueErr:  "cannot convert argument to Nodes",
		},
		{
			test:      "logical_or",
			expr:      []spec.FuncExprArg{spec.LogicalOr{}},
			lengthErr: "cannot convert argument to Value",
			countErr:  "cannot convert argument to Nodes",
			valueErr:  "cannot convert argument to Nodes",
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

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
	reg := New()

	for _, tc := range []struct {
		test string
		expr []spec.FuncExprArg
		err  string
	}{
		{
			test: "no_args",
			expr: []spec.FuncExprArg{},
			err:  "expected 2 arguments but found 0",
		},
		{
			test: "one_arg",
			expr: []spec.FuncExprArg{spec.Literal("hi")},
			err:  "expected 2 arguments but found 1",
		},
		{
			test: "three_args",
			expr: []spec.FuncExprArg{spec.Literal("hi"), spec.Literal("hi"), spec.Literal("hi")},
			err:  "expected 2 arguments but found 3",
		},
		{
			test: "logical_or_1",
			expr: []spec.FuncExprArg{&spec.LogicalOr{}, spec.Literal("hi")},
			err:  "cannot convert argument 1 to Value",
		},
		{
			test: "logical_or_2",
			expr: []spec.FuncExprArg{spec.Literal("hi"), spec.LogicalOr{}},
			err:  "cannot convert argument 2 to Value",
		},
		{
			test: "singular_query_literal",
			expr: []spec.FuncExprArg{&spec.SingularQueryExpr{}, spec.Literal("hi")},
		},
		{
			test: "literal_singular_query",
			expr: []spec.FuncExprArg{spec.Literal("hi"), &spec.SingularQueryExpr{}},
		},
		{
			test: "nodes_query_1",
			expr: []spec.FuncExprArg{
				spec.Query(true, spec.Child(spec.Name("x"))),
				spec.Literal("hi"),
			},
		},
		{
			test: "nodes_query_2",
			expr: []spec.FuncExprArg{
				spec.Literal("hi"),
				spec.Query(true, spec.Child(spec.Name("x"))),
			},
		},
		{
			test: "func_expr_1",
			expr: []spec.FuncExprArg{
				spec.Function(
					reg.Get("match"),
					spec.Query(true, spec.Child(spec.Name("x"))),
					spec.Literal("hi"),
				),
				spec.Literal("hi"),
			},
			err: "cannot convert argument 1 to Value",
		},
		{
			test: "func_expr_2",
			expr: []spec.FuncExprArg{
				spec.Literal("hi"),
				spec.Function(
					reg.Get("match"),
					spec.Query(true, spec.Child(spec.Name("x"))),
					spec.Literal("hi"),
				),
			},
			err: "cannot convert argument 2 to Value",
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

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

	for _, tc := range []struct {
		test string
		vals []spec.PathValue
		exp  int
		err  string
	}{
		{"empty", []spec.PathValue{spec.Nodes()}, 0, ""},
		{"one", []spec.PathValue{spec.Nodes(1)}, 1, ""},
		{"three", []spec.PathValue{spec.Nodes(1, true, nil)}, 3, ""},
		{"not_nodes", []spec.PathValue{spec.LogicalTrue}, 0, "cannot convert LogicalType to NodesType"},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			if tc.err != "" {
				a.PanicsWithValue(tc.err, func() { countFunc(tc.vals) })
				return
			}
			res := countFunc(tc.vals)
			if tc.exp < 0 {
				a.Nil(res)
			} else {
				a.Equal(spec.Value(tc.exp), res)
			}
		})
	}
}

func TestValueFunc(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
		vals []spec.PathValue
		exp  spec.PathValue
		err  string
	}{
		{"empty", []spec.PathValue{spec.Nodes()}, nil, ""},
		{"one_int", []spec.PathValue{spec.Nodes(1)}, spec.Value(1), ""},
		{"one_null", []spec.PathValue{spec.Nodes(nil)}, spec.Value(nil), ""},
		{"one_string", []spec.PathValue{spec.Nodes("x")}, spec.Value("x"), ""},
		{"three", []spec.PathValue{spec.Nodes(1, true, nil)}, nil, ""},
		{"not_nodes", []spec.PathValue{spec.LogicalFalse}, nil, "cannot convert LogicalType to NodesType"},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

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

	for _, tc := range []struct {
		test   string
		input  *spec.ValueType
		regex  *spec.ValueType
		match  bool
		search bool
	}{
		{
			test:   "dot",
			input:  spec.Value("x"),
			regex:  spec.Value("."),
			match:  true,
			search: true,
		},
		{
			test:   "two_chars",
			input:  spec.Value("xx"),
			regex:  spec.Value("."),
			match:  false,
			search: true,
		},
		{
			test:   "multi_line_newline",
			input:  spec.Value("xx\nyz"),
			regex:  spec.Value(".*"),
			match:  false,
			search: true,
		},
		{
			test:   "multi_line_crlf",
			input:  spec.Value("xx\r\nyz"),
			regex:  spec.Value(".*"),
			match:  false,
			search: true,
		},
		{
			test:   "not_string_input",
			input:  spec.Value(1),
			regex:  spec.Value("."),
			match:  false,
			search: false,
		},
		{
			test:   "not_string_regex",
			input:  spec.Value("x"),
			regex:  spec.Value(1),
			match:  false,
			search: false,
		},
		{
			test:   "invalid_regex",
			input:  spec.Value("x"),
			regex:  spec.Value(".["),
			match:  false,
			search: false,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.Equal(spec.Logical(tc.match), matchFunc([]spec.PathValue{tc.input, tc.regex}))
			a.Equal(spec.Logical(tc.search), searchFunc([]spec.PathValue{tc.input, tc.regex}))
		})
	}
}

func TestExecRegexFuncs(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test   string
		vals   []spec.PathValue
		match  bool
		search bool
		err    string
	}{
		{
			test:   "dot",
			vals:   []spec.PathValue{spec.Value("x"), spec.Value("x")},
			match:  true,
			search: true,
		},
		{
			test: "first_not_value",
			vals: []spec.PathValue{spec.Nodes(), spec.Value("x")},
			err:  "cannot convert NodesType to ValueType",
		},
		{
			test: "second_not_value",
			vals: []spec.PathValue{spec.Value("x"), spec.LogicalFalse},
			err:  "cannot convert LogicalType to ValueType",
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			if tc.err == "" {
				a.Equal(matchFunc(tc.vals), spec.Logical(tc.match))
				a.Equal(searchFunc(tc.vals), spec.Logical(tc.search))
			} else {
				a.PanicsWithValue(tc.err, func() { matchFunc(tc.vals) })
				a.PanicsWithValue(tc.err, func() { searchFunc(tc.vals) })
			}
		})
	}
}
