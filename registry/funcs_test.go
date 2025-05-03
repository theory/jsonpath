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
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		vals []spec.JSONPathValue
		exp  int
		err  string
	}{
		{
			name: "empty_string",
			vals: []spec.JSONPathValue{spec.Value("")},
			exp:  0,
		},
		{
			name: "ascii_string",
			vals: []spec.JSONPathValue{spec.Value("abc def")},
			exp:  7,
		},
		{
			name: "unicode_string",
			vals: []spec.JSONPathValue{spec.Value("foö")},
			exp:  3,
		},
		{
			name: "emoji_string",
			vals: []spec.JSONPathValue{spec.Value("Hi 👋🏻")},
			exp:  5,
		},
		{
			name: "empty_array",
			vals: []spec.JSONPathValue{spec.Value([]any{})},
			exp:  0,
		},
		{
			name: "array",
			vals: []spec.JSONPathValue{spec.Value([]any{1, 2, 3, 4, 5})},
			exp:  5,
		},
		{
			name: "nested_array",
			vals: []spec.JSONPathValue{spec.Value([]any{1, 2, 3, "x", []any{456, 67}, true})},
			exp:  6,
		},
		{
			name: "empty_object",
			vals: []spec.JSONPathValue{spec.Value(map[string]any{})},
			exp:  0,
		},
		{
			name: "object",
			vals: []spec.JSONPathValue{spec.Value(map[string]any{"x": 1, "y": 0, "z": 2})},
			exp:  3,
		},
		{
			name: "nested_object",
			vals: []spec.JSONPathValue{spec.Value(map[string]any{
				"x": 1,
				"y": 0,
				"z": []any{1, 2},
				"a": map[string]any{"b": 9},
			})},
			exp: 4,
		},
		{
			name: "integer",
			vals: []spec.JSONPathValue{spec.Value(42)},
			exp:  -1,
		},
		{
			name: "bool",
			vals: []spec.JSONPathValue{spec.Value(true)},
			exp:  -1,
		},
		{
			name: "null",
			vals: []spec.JSONPathValue{spec.Value(nil)},
			exp:  -1,
		},
		{
			name: "nil",
			vals: []spec.JSONPathValue{nil},
			exp:  -1,
		},
		{
			name: "not_value",
			vals: []spec.JSONPathValue{spec.LogicalFalse},
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
				a.Equal(spec.Value(tc.exp), res)
			}
		})
	}
}

func TestCheckSingularFuncArgs(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	reg := New()

	for _, tc := range []struct {
		name      string
		expr      []spec.FuncExprArg
		err       string
		lengthErr string
		countErr  string
		valueErr  string
	}{
		{
			name: "no_args",
			expr: []spec.FuncExprArg{},
			err:  "expected 1 argument but found 0",
		},
		{
			name: "two_args",
			expr: []spec.FuncExprArg{spec.Literal(nil), spec.Literal(nil)},
			err:  "expected 1 argument but found 2",
		},
		{
			name:     "literal_string",
			expr:     []spec.FuncExprArg{spec.Literal(nil)},
			countErr: "cannot convert argument to Nodes",
			valueErr: "cannot convert argument to Nodes",
		},
		{
			name: "singular_query",
			expr: []spec.FuncExprArg{spec.SingularQuery(false, nil)},
		},
		{
			name: "nodes_query",
			expr: []spec.FuncExprArg{spec.NodesQuery(
				spec.Query(true, spec.Child(spec.Name("x"))),
			)},
		},
		{
			name: "logical_func_expr",
			expr: []spec.FuncExprArg{spec.Function(reg.Get("match"),
				spec.NodesQuery(
					spec.Query(true, spec.Child(spec.Name("x"))),
				),
				spec.Literal("hi"),
			)},
			lengthErr: "cannot convert argument to Value",
			countErr:  "cannot convert argument to Nodes",
			valueErr:  "cannot convert argument to Nodes",
		},
		{
			name:      "logical_or",
			expr:      []spec.FuncExprArg{spec.LogicalOr{}},
			lengthErr: "cannot convert argument to Value",
			countErr:  "cannot convert argument to Nodes",
			valueErr:  "cannot convert argument to Nodes",
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
	reg := New()

	for _, tc := range []struct {
		name string
		expr []spec.FuncExprArg
		err  string
	}{
		{
			name: "no_args",
			expr: []spec.FuncExprArg{},
			err:  "expected 2 arguments but found 0",
		},
		{
			name: "one_arg",
			expr: []spec.FuncExprArg{spec.Literal("hi")},
			err:  "expected 2 arguments but found 1",
		},
		{
			name: "three_args",
			expr: []spec.FuncExprArg{spec.Literal("hi"), spec.Literal("hi"), spec.Literal("hi")},
			err:  "expected 2 arguments but found 3",
		},
		{
			name: "logical_or_1",
			expr: []spec.FuncExprArg{&spec.LogicalOr{}, spec.Literal("hi")},
			err:  "cannot convert argument 1 to Value",
		},
		{
			name: "logical_or_2",
			expr: []spec.FuncExprArg{spec.Literal("hi"), spec.LogicalOr{}},
			err:  "cannot convert argument 2 to Value",
		},
		{
			name: "singular_query_literal",
			expr: []spec.FuncExprArg{&spec.SingularQueryExpr{}, spec.Literal("hi")},
		},
		{
			name: "literal_singular_query",
			expr: []spec.FuncExprArg{spec.Literal("hi"), &spec.SingularQueryExpr{}},
		},
		{
			name: "nodes_query_1",
			expr: []spec.FuncExprArg{
				spec.NodesQuery(spec.Query(true, spec.Child(spec.Name("x")))),
				spec.Literal("hi"),
			},
		},
		{
			name: "nodes_query_2",
			expr: []spec.FuncExprArg{
				spec.Literal("hi"),
				spec.NodesQuery(spec.Query(true, spec.Child(spec.Name("x")))),
			},
		},
		{
			name: "func_expr_1",
			expr: []spec.FuncExprArg{
				spec.Function(
					reg.Get("match"),
					spec.NodesQuery(spec.Query(true, spec.Child(spec.Name("x")))),
					spec.Literal("hi"),
				),
				spec.Literal("hi"),
			},
			err: "cannot convert argument 1 to Value",
		},
		{
			name: "func_expr_2",
			expr: []spec.FuncExprArg{
				spec.Literal("hi"),
				spec.Function(
					reg.Get("match"),
					spec.NodesQuery(spec.Query(true, spec.Child(spec.Name("x")))),
					spec.Literal("hi"),
				),
			},
			err: "cannot convert argument 2 to Value",
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
		vals []spec.JSONPathValue
		exp  int
		err  string
	}{
		{"empty", []spec.JSONPathValue{spec.Nodes()}, 0, ""},
		{"one", []spec.JSONPathValue{spec.Nodes(1)}, 1, ""},
		{"three", []spec.JSONPathValue{spec.Nodes(1, true, nil)}, 3, ""},
		{"not_nodes", []spec.JSONPathValue{spec.LogicalTrue}, 0, "unexpected argument of type spec.LogicalType"},
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
				a.Equal(spec.Value(tc.exp), res)
			}
		})
	}
}

func TestValueFunc(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		vals []spec.JSONPathValue
		exp  spec.JSONPathValue
		err  string
	}{
		{"empty", []spec.JSONPathValue{spec.Nodes()}, nil, ""},
		{"one_int", []spec.JSONPathValue{spec.Nodes(1)}, spec.Value(1), ""},
		{"one_null", []spec.JSONPathValue{spec.Nodes(nil)}, spec.Value(nil), ""},
		{"one_string", []spec.JSONPathValue{spec.Nodes("x")}, spec.Value("x"), ""},
		{"three", []spec.JSONPathValue{spec.Nodes(1, true, nil)}, nil, ""},
		{"not_nodes", []spec.JSONPathValue{spec.LogicalFalse}, nil, "unexpected argument of type spec.LogicalType"},
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
		input  *spec.ValueType
		regex  *spec.ValueType
		match  bool
		search bool
	}{
		{
			name:   "dot",
			input:  spec.Value("x"),
			regex:  spec.Value("."),
			match:  true,
			search: true,
		},
		{
			name:   "two_chars",
			input:  spec.Value("xx"),
			regex:  spec.Value("."),
			match:  false,
			search: true,
		},
		{
			name:   "multi_line_newline",
			input:  spec.Value("xx\nyz"),
			regex:  spec.Value(".*"),
			match:  false,
			search: true,
		},
		{
			name:   "multi_line_crlf",
			input:  spec.Value("xx\r\nyz"),
			regex:  spec.Value(".*"),
			match:  false,
			search: true,
		},
		{
			name:   "not_string_input",
			input:  spec.Value(1),
			regex:  spec.Value("."),
			match:  false,
			search: false,
		},
		{
			name:   "not_string_regex",
			input:  spec.Value("x"),
			regex:  spec.Value(1),
			match:  false,
			search: false,
		},
		{
			name:   "invalid_regex",
			input:  spec.Value("x"),
			regex:  spec.Value(".["),
			match:  false,
			search: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(spec.LogicalFrom(tc.match), matchFunc([]spec.JSONPathValue{tc.input, tc.regex}))
			a.Equal(spec.LogicalFrom(tc.search), searchFunc([]spec.JSONPathValue{tc.input, tc.regex}))
		})
	}
}

func TestExecRegexFuncs(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name   string
		vals   []spec.JSONPathValue
		match  bool
		search bool
		err    string
	}{
		{
			name:   "dot",
			vals:   []spec.JSONPathValue{spec.Value("x"), spec.Value("x")},
			match:  true,
			search: true,
		},
		{
			name: "first_not_value",
			vals: []spec.JSONPathValue{spec.Nodes(), spec.Value("x")},
			err:  "unexpected argument of type spec.NodesType",
		},
		{
			name: "second_not_value",
			vals: []spec.JSONPathValue{spec.Value("x"), spec.LogicalFalse},
			err:  "unexpected argument of type spec.LogicalType",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.err == "" {
				a.Equal(matchFunc(tc.vals), spec.LogicalFrom(tc.match))
				a.Equal(searchFunc(tc.vals), spec.LogicalFrom(tc.search))
			} else {
				a.PanicsWithValue(tc.err, func() { matchFunc(tc.vals) })
				a.PanicsWithValue(tc.err, func() { searchFunc(tc.vals) })
			}
		})
	}
}
