package parser

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theory/jsonpath/registry"
	"github.com/theory/jsonpath/spec"
)

func TestParseRoot(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)

	q, err := Parse(registry.New(), "$")
	r.NoError(err)
	a.Equal("$", q.String())
	a.Empty(q.Segments())
}

func TestParseSimple(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)
	reg := registry.New()

	for _, tc := range []struct {
		name string
		path string
		exp  *spec.PathQuery
		err  string
	}{
		{
			name: "root",
			path: "$",
			exp:  spec.Query(true, []*spec.Segment{}...),
		},
		{
			name: "name",
			path: "$.x",
			exp:  spec.Query(true, spec.Child(spec.Name("x"))),
		},
		{
			name: "trim_leading_space",
			path: "   $.x",
			err:  `jsonpath: unexpected blank space at position 1`,
		},
		{
			name: "trim_trailing_space",
			path: "$.x    ",
			err:  `jsonpath: unexpected blank space at position 4`,
		},
		{
			name: "no_interim_space",
			path: "$.x   .y",
			exp: spec.Query(
				true,
				spec.Child(spec.Name("x")),
				spec.Child(spec.Name("y")),
			),
		},
		{
			name: "unexpected_integer",
			path: "$.62",
			err:  `jsonpath: unexpected integer at position 3`,
		},
		{
			name: "unexpected_token",
			path: "$.==12",
			err:  `jsonpath: unexpected '=' at position 3`,
		},
		{
			name: "name_name",
			path: "$.x.y",
			exp: spec.Query(
				true,
				spec.Child(spec.Name("x")),
				spec.Child(spec.Name("y")),
			),
		},
		{
			name: "wildcard",
			path: "$.*",
			exp:  spec.Query(true, spec.Child(spec.Wildcard)),
		},
		{
			name: "wildcard_wildcard",
			path: "$.*.*",
			exp: spec.Query(
				true,
				spec.Child(spec.Wildcard),
				spec.Child(spec.Wildcard),
			),
		},
		{
			name: "name_wildcard",
			path: "$.x.*",
			exp: spec.Query(
				true,
				spec.Child(spec.Name("x")),
				spec.Child(spec.Wildcard),
			),
		},
		{
			name: "desc_name",
			path: "$..x",
			exp:  spec.Query(true, spec.Descendant(spec.Name("x"))),
		},
		{
			name: "desc_name_2x",
			path: "$..x..y",
			exp: spec.Query(
				true,
				spec.Descendant(spec.Name("x")),
				spec.Descendant(spec.Name("y")),
			),
		},
		{
			name: "desc_wildcard",
			path: "$..*",
			exp:  spec.Query(true, spec.Descendant(spec.Wildcard)),
		},
		{
			name: "desc_wildcard_2x",
			path: "$..*..*",
			exp: spec.Query(
				true,
				spec.Descendant(spec.Wildcard),
				spec.Descendant(spec.Wildcard),
			),
		},
		{
			name: "desc_wildcard_name",
			path: "$..*.xyz",
			exp: spec.Query(
				true,
				spec.Descendant(spec.Wildcard),
				spec.Child(spec.Name("xyz")),
			),
		},
		{
			name: "wildcard_desc_name",
			path: "$.*..xyz",
			exp: spec.Query(
				true,
				spec.Child(spec.Wildcard),
				spec.Descendant(spec.Name("xyz")),
			),
		},
		{
			name: "empty_string",
			path: "",
			err:  "jsonpath: unexpected end of input",
		},
		{
			name: "bad_start",
			path: ".x",
			err:  `jsonpath: unexpected '.' at position 1`,
		},
		{
			name: "not_a_segment",
			path: "$foo",
			err:  "jsonpath: unexpected identifier at position 1",
		},
		{
			name: "not_a_dot_segment",
			path: "$.{x}",
			err:  `jsonpath: unexpected '{' at position 3`,
		},
		{
			name: "not_a_descendant",
			path: "$..{x}",
			err:  `jsonpath: unexpected '{' at position 4`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			q, err := Parse(reg, tc.path)
			if tc.err == "" {
				r.NoError(err)
				a.Equal(tc.exp, q)
			} else {
				a.Nil(q)
				r.EqualError(err, tc.err)
				r.ErrorIs(err, ErrPathParse)
			}
		})
	}
}

func TestParseFilter(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)
	reg := registry.New()
	_ = reg.Register(
		"__true",
		spec.FuncLogical,
		func([]spec.FuncExprArg) error { return nil },
		func([]spec.JSONPathValue) spec.JSONPathValue {
			return spec.LogicalTrue
		},
	)
	trueFunc := reg.Get("__true")

	for _, tc := range []struct {
		name   string
		query  string
		filter *spec.FilterSelector
		err    string
	}{
		// ExistExpr
		{
			name:  "current_exists",
			query: "@",
			filter: spec.Filter(spec.And(
				spec.Existence(spec.Query(false, []*spec.Segment{}...)),
			)),
		},
		{
			name:  "root_exists",
			query: "$",
			filter: spec.Filter(spec.And(
				spec.Existence(spec.Query(true, []*spec.Segment{}...)),
			)),
		},
		{
			name:  "current_name_exists",
			query: "@.x",
			filter: spec.Filter(spec.And(
				spec.Existence(spec.Query(false, spec.Child(spec.Name("x")))),
			)),
		},
		{
			name:  "root_name_exists",
			query: "$.x",
			filter: spec.Filter(spec.And(
				spec.Existence(spec.Query(true, spec.Child(spec.Name("x")))),
			)),
		},
		{
			name:  "current_two_segment_exists",
			query: "@.x[1]",
			filter: spec.Filter(spec.And(
				spec.Existence(spec.Query(
					false,
					spec.Child(spec.Name("x")),
					spec.Child(spec.Index(1)),
				)),
			)),
		},
		{
			name:  "root_two_selector_exists",
			query: `$["x", 1]`,
			filter: spec.Filter(spec.And(
				spec.Existence(spec.Query(
					true,
					spec.Child(spec.Name("x"), spec.Index(1)),
				)),
			)),
		},
		// NonExistExpr
		{
			name:  "current_not_exists",
			query: "!@",
			filter: spec.Filter(spec.And(
				spec.Nonexistence(spec.Query(false, []*spec.Segment{}...)),
			)),
		},
		{
			name:  "root_not_exists",
			query: "!$",
			filter: spec.Filter(spec.And(spec.Nonexistence(
				spec.Query(true, []*spec.Segment{}...),
			))),
		},
		{
			name:  "current_name_not_exists",
			query: "!@.x",
			filter: spec.Filter(spec.And(
				spec.Nonexistence(spec.Query(false, spec.Child(spec.Name("x")))),
			)),
		},
		{
			name:  "root_name_not_exists",
			query: "!$.x",
			filter: spec.Filter(spec.And(
				spec.Nonexistence(spec.Query(true, spec.Child(spec.Name("x")))),
			)),
		},
		{
			name:  "current_two_segment_not_exists",
			query: "!@.x[1]",
			filter: spec.Filter(spec.And(
				spec.Nonexistence(spec.Query(
					false,
					spec.Child(spec.Name("x")),
					spec.Child(spec.Index(1)),
				)),
			)),
		},
		{
			name:  "root_two_selector_not_exists",
			query: `!$["x", 1]`,
			filter: spec.Filter(spec.And(
				spec.Nonexistence(spec.Query(
					true,
					spec.Child(spec.Name("x"), spec.Index(1)),
				)),
			)),
		},
		// ParenExistExpr
		{
			name:  "paren_current_exists",
			query: "(@)",
			filter: spec.Filter(spec.And(
				spec.Paren(spec.And(
					spec.Existence(spec.Query(false, []*spec.Segment{}...)),
				)),
			)),
		},
		{
			name:  "paren_root_exists",
			query: "($)",
			filter: spec.Filter(spec.And(
				spec.Paren(spec.And(
					spec.Existence(spec.Query(true, []*spec.Segment{}...)),
				)),
			)),
		},
		{
			name:  "paren_current_exists_name_index",
			query: `(@["x", 1])`,
			filter: spec.Filter(spec.And(
				spec.Paren(spec.And(
					spec.Existence(spec.Query(
						false,
						spec.Child(spec.Name("x"), spec.Index(1)),
					)),
				)),
			)),
		},
		{
			name:  "paren_logical_and",
			query: `(  @["x", 1] && $["y"]  )`,
			filter: spec.Filter(spec.And(
				spec.Paren(spec.And(
					spec.Existence(spec.Query(
						false,
						spec.Child(spec.Name("x"), spec.Index(1)),
					)),
					spec.Existence(spec.Query(
						true,
						spec.Child(spec.Name("y")),
					)),
				)),
			)),
		},
		{
			name:  "paren_logical_or",
			query: `(@["x", 1] || $["y"])`,
			filter: spec.Filter(spec.And(
				spec.Paren(
					spec.And(spec.Existence(spec.Query(
						false,
						spec.Child(spec.Name("x"), spec.Index(1)),
					))),
					spec.And(spec.Existence(spec.Query(
						true,
						spec.Child(spec.Name("y")),
					))),
				)),
			),
		},
		// NotParenExistExpr
		{
			name:  "not_paren_current_exists",
			query: "!(@)",
			filter: spec.Filter(spec.And(
				spec.NotParen(spec.And(
					spec.Existence(spec.Query(false, []*spec.Segment{}...)),
				)),
			)),
		},
		{
			name:  "not_paren_root_exists",
			query: "!($)",
			filter: spec.Filter(spec.And(
				spec.NotParen(spec.And(
					spec.Existence(spec.Query(true, []*spec.Segment{}...)),
				)),
			)),
		},
		{
			name:  "not_paren_current_exists_name_index",
			query: `!(@["x", 1])`,
			filter: spec.Filter(spec.And(
				spec.NotParen(spec.And(
					spec.Existence(spec.Query(
						false,
						spec.Child(spec.Name("x"), spec.Index(1)),
					)),
				)),
			)),
		},
		{
			name:  "not_paren_logical_and",
			query: `!(  @["x", 1] && $["y"]  )`,
			filter: spec.Filter(spec.And(
				spec.NotParen(spec.And(
					spec.Existence(spec.Query(
						false,
						spec.Child(spec.Name("x"), spec.Index(1)),
					)),
					spec.Existence(spec.Query(
						true,
						spec.Child(spec.Name("y")),
					)),
				)),
			)),
		},
		{
			name:  "not_paren_logical_or",
			query: `!(@["x", 1] || $["y"])`,
			filter: spec.Filter(spec.And(
				spec.NotParen(
					spec.And(spec.Existence(spec.Query(
						false,
						spec.Child(spec.Name("x"), spec.Index(1)),
					))),
					spec.And(spec.Existence(spec.Query(
						true,
						spec.Child(spec.Name("y")),
					))),
				),
			)),
		},
		// FunExpr
		{
			name:  "function_current",
			query: "__true(@)",
			filter: spec.Filter(spec.And(
				spec.Function(
					trueFunc,
					spec.SingularQuery(false, []spec.Selector{}...),
				),
			)),
		},
		{
			name:  "function_match_current_integer",
			query: "match( @,  42  )",
			filter: spec.Filter(spec.And(
				spec.Function(
					reg.Get("match"),
					spec.SingularQuery(false, []spec.Selector{}...),
					spec.Literal(int64(42)),
				),
			)),
		},
		{
			name:  "function_search_two_queries",
			query: "search( $.x,  @[0]  )",
			filter: spec.Filter(spec.And(
				spec.Function(
					reg.Get("search"),
					spec.SingularQuery(true, spec.Name("x")),
					spec.SingularQuery(false, spec.Index(0)),
				),
			)),
		},
		{
			name:  "function_length_string",
			query: `length("hi") == 2`,
			filter: spec.Filter(spec.And(
				spec.Comparison(
					spec.Function(reg.Get("length"), spec.Literal("hi")),
					spec.EqualTo,
					spec.Literal(int64(2)),
				),
			)),
		},
		{
			name:  "function_length_true",
			query: `length(true) == 1`,
			filter: spec.Filter(spec.And(
				spec.Comparison(
					spec.Function(reg.Get("length"), spec.Literal(true)),
					spec.EqualTo,
					spec.Literal(int64(1)),
				),
			)),
		},
		{
			name:  "function_length_false",
			query: `length(false)==1`,
			filter: spec.Filter(spec.And(
				spec.Comparison(
					spec.Function(reg.Get("length"), spec.Literal(false)),
					spec.EqualTo,
					spec.Literal(int64(1)),
				),
			)),
		},
		{
			name:  "function_value_null",
			query: `__true(null)`, // defined in function_test.go
			filter: spec.Filter(spec.And(
				spec.Function(trueFunc, spec.Literal(nil)),
			)),
		},
		{
			name:  "nested_function",
			query: `__true(count(@))`, // defined in function_test.go
			filter: spec.Filter(spec.And(
				spec.Function(
					trueFunc,
					spec.Function(
						reg.Get("count"),
						spec.SingularQuery(false, []spec.Selector{}...),
					),
				),
			)),
		},
		{
			name:  "function_paren_logical_expr",
			query: `__true((@.x))`, // defined in function_test.go
			filter: spec.Filter(spec.And(
				spec.Function(trueFunc, spec.Or(spec.And(
					spec.Existence(spec.Query(false, spec.Child(spec.Name("x")))),
				))),
			)),
		},
		{
			name:  "function_paren_logical_not_expr",
			query: `__true((!@.x))`, // defined in function_test.go
			filter: spec.Filter(spec.And(
				spec.Function(trueFunc, spec.Or(spec.And(
					spec.Nonexistence(spec.Query(false, spec.Child(spec.Name("x")))),
				))),
			)),
		},
		{
			name:  "function_lots_of_literals",
			query: `__true("hi", 42, true, false, null, 98.6)`,
			filter: spec.Filter(spec.And(
				spec.Function(
					trueFunc,
					spec.Literal("hi"),
					spec.Literal(int64(42)),
					spec.Literal(true),
					spec.Literal(false),
					spec.Literal(nil),
					spec.Literal(float64(98.6)),
				),
			)),
		},
		{
			name:  "function_no_args",
			query: `__true()`,
			filter: spec.Filter(spec.And(
				spec.Function(trueFunc, []spec.FuncExprArg{}...),
			)),
		},
		// ComparisonExpr
		{
			name:  "literal_comparison",
			query: `42 == 42`,
			filter: spec.Filter(spec.And(
				spec.Comparison(spec.Literal(int64(42)), spec.EqualTo, spec.Literal(int64(42))),
			)),
		},
		{
			name:  "literal_singular_comparison",
			query: `42 != $.a.b`,
			filter: spec.Filter(spec.And(
				spec.Comparison(
					spec.Literal(int64(42)),
					spec.NotEqualTo,
					spec.SingularQuery(true, spec.Name("a"), spec.Name("b")),
				),
			)),
		},
		{
			name:  "literal_comparison_function",
			query: `42 > length("hi")`,
			filter: spec.Filter(spec.And(
				spec.Comparison(
					spec.Literal(int64(42)),
					spec.GreaterThan,
					spec.Function(reg.Get("length"), spec.Literal("hi")),
				),
			)),
		},
		{
			name:  "function_cmp_singular",
			query: `length("hi") <=   @[0]["a"]`,
			filter: spec.Filter(spec.And(
				spec.Comparison(
					spec.Function(reg.Get("length"), spec.Literal("hi")),
					spec.LessThanEqualTo,
					spec.SingularQuery(false, spec.Index(0), spec.Name("a")),
				),
			)),
		},
		{
			name:  "singular_cmp_literal",
			query: `$.a.b >= 98.6`,
			filter: spec.Filter(spec.And(
				spec.Comparison(
					spec.SingularQuery(true, spec.Name("a"), spec.Name("b")),
					spec.GreaterThanEqualTo,
					spec.Literal(float64(98.6)),
				),
			)),
		},
		{
			name:  "function_cmp_literal",
			query: `length("hi") <   42 `,
			filter: spec.Filter(spec.And(
				spec.Comparison(
					spec.Function(reg.Get("length"), spec.Literal("hi")),
					spec.LessThan,
					spec.Literal(int64(42)),
				),
			)),
		},
		{
			name:  "not_function",
			query: `!__true()`,
			filter: spec.Filter(spec.And(
				spec.NotFunction(spec.Function(trueFunc, []spec.FuncExprArg{}...)),
			)),
		},
		{
			name:  "singular_cmp_literal_no_space",
			query: `@.x<42`,
			filter: spec.Filter(spec.And(
				spec.Comparison(
					spec.SingularQuery(false, spec.Name("x")),
					spec.LessThan,
					spec.Literal(int64(42)),
				),
			)),
		},
		{
			name:  "invalid_logical_or",
			query: `(@["x", 1] || hi)`,
			err:   `jsonpath: unexpected identifier at position 15`,
		},
		{
			name:  "invalid_logical_or",
			query: `(@["x", 1] || hi)`,
			err:   `jsonpath: unexpected identifier at position 15`,
		},
		{
			name:  "incomplete_logical_or",
			query: `(@["x", 1] | $["y"])`,
			err:   `jsonpath: expected '|' but found blank space at position 13`,
		},
		{
			name:  "incomplete_logical_and",
			query: `(@["x", 1] &? $["y"])`,
			err:   `jsonpath: expected '&' but found '?' at position 13`,
		},
		{
			name:  "invalid_and_expression",
			query: `(@["x", 1] && nope(@))`,
			err:   `jsonpath: unknown function nope() at position 15`,
		},
		{
			name:  "nonexistent_function",
			query: `nonesuch(@)`,
			err:   `jsonpath: unknown function nonesuch() at position 1`,
		},
		{
			name:  "not_nonexistent_function",
			query: `!nonesuch(@)`,
			err:   `jsonpath: unknown function nonesuch() at position 2`,
		},
		{
			name:  "invalid_literal",
			query: `99e+1234`,
			err:   `jsonpath: cannot parse "99e+1234", value out of range at position 1`,
		},
		{
			name:  "invalid_query",
			query: `@["x", hi]`,
			err:   `jsonpath: unexpected identifier at position 8`,
		},
		{
			name:  "invalid_function_comparison",
			query: `length(@.x) == hi`,
			err:   `jsonpath: unexpected identifier at position 16`,
		},
		{
			name:  "function_without_comparison",
			query: `length(@.x)`,
			err:   `jsonpath: missing comparison to function result at position 12`,
		},
		{
			name:  "invalid_not_exists_query",
			query: `!@.0`,
			err:   `jsonpath: unexpected integer at position 4`,
		},
		{
			name:  "unclosed_paren_expr",
			query: `(@["x", 1]`,
			err:   `jsonpath: expected ')' but found eof at position 11`,
		},
		{
			name:  "unclosed_not_paren_expr",
			query: `!(@["x", 1]`,
			err:   `jsonpath: expected ')' but found eof at position 12`,
		},
		{
			name:  "bad_function_arg",
			query: `length(xyz)`,
			err:   `jsonpath: unexpected identifier at position 8`,
		},
		{
			name:  "invalid_function_arg",
			query: `length(@[1, 2])`,
			err:   `jsonpath: function length() cannot convert argument to Value at position 7`,
		},
		{
			name:  "too_many_function_args",
			query: `length(@, $)`,
			err:   `jsonpath: function length() expected 1 argument but found 2 at position 7`,
		},
		{
			name:  "function_literal_parse_error",
			query: `length(99e+1234)`,
			err:   `jsonpath: cannot parse "99e+1234", value out of range at position 8`,
		},
		{
			name:  "function_query_parse_error",
			query: `length(@[foo])`,
			err:   `jsonpath: unexpected identifier at position 10`,
		},
		{
			name:  "unknown_function_in_function_arg",
			query: `length(nonesuch())`,
			err:   `jsonpath: unknown function nonesuch() at position 8`,
		},
		{
			name:  "invalid_not_function_arg",
			query: `length(!@[foo])`,
			err:   `jsonpath: unexpected identifier at position 11`,
		},
		{
			name:  "invalid_second_arg",
			query: `length("foo" == "bar")`,
			err:   `jsonpath: unexpected '=' at position 14`,
		},
		{
			name:  "invalid_comparable_expression",
			query: `"foo" => "bar"`,
			err:   `jsonpath: invalid comparison operator at position 7`,
		},
		{
			name:  "invalid_comparable_function",
			query: `42 == nonesuch()`,
			err:   `jsonpath: unknown function nonesuch() at position 7`,
		},
		{
			name:  "cannot_compare_logical_func",
			query: `42 == __true()`,
			err:   `jsonpath: cannot compare result of logical function at position 7`,
		},
		{
			name:  "function_wrong_arg_count",
			query: `match("foo")`,
			err:   `jsonpath: function match() expected 2 arguments but found 1 at position 6`,
		},
		{
			name:  "function_second_arg_parse_error",
			query: `search("foo", @[foo])`,
			err:   `jsonpath: unexpected identifier at position 17`,
		},
		{
			name:  "cmp_query_invalid_index",
			query: `42 == $[-0]`,
			err:   `jsonpath: invalid integer path value "-0" at position 9`,
		},
		{
			name:  "cmp_val_not_valid",
			query: `42 == {}`,
			err:   `jsonpath: unexpected '{' at position 7`,
		},
		{
			name:  "cmp_query_unclosed_bracket",
			query: `42 == $[0`,
			err:   `jsonpath: unexpected eof at position 10`,
		},
		{
			name:  "cmp_query_invalid_selector",
			query: `42 == $[foo]`,
			err:   `jsonpath: unexpected identifier at position 9`,
		},
		{
			name:  "cmp_query_invalid_ident",
			query: `42 == $.42`,
			err:   `jsonpath: unexpected integer at position 9`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			parser := &parser{lex: newLexer(tc.query), reg: reg}
			filter, err := parser.parseFilter()
			if tc.err == "" {
				r.NoError(err, tc.name)
				a.Equal(tc.filter, filter, tc.name)
			} else {
				a.Nil(filter, tc.name)
				r.EqualError(err, tc.err, tc.name)
				r.ErrorIs(err, ErrPathParse, tc.name)
			}
		})
	}
}

func TestParseSelectors(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)
	reg := registry.New()

	for _, tc := range []struct {
		name string
		path string
		exp  *spec.PathQuery
		err  string
	}{
		{
			name: "index",
			path: "$[0]",
			exp:  spec.Query(true, spec.Child(spec.Index(0))),
		},
		{
			name: "two_indexes",
			path: "$[0, 1]",
			exp:  spec.Query(true, spec.Child(spec.Index(0), spec.Index(1))),
		},
		{
			name: "name",
			path: `$["foo"]`,
			exp:  spec.Query(true, spec.Child(spec.Name("foo"))),
		},
		{
			name: "sq_name",
			path: `$['foo']`,
			exp:  spec.Query(true, spec.Child(spec.Name("foo"))),
		},
		{
			name: "two_names",
			path: `$["foo", "🐦‍🔥"]`,
			exp:  spec.Query(true, spec.Child(spec.Name("foo"), spec.Name("🐦‍🔥"))),
		},
		{
			name: "json_escapes",
			path: `$["abx_xyx\ryup", "\b\f\n\r\t\/\\"]`,
			exp: spec.Query(true, spec.Child(
				spec.Name("abx_xyx\ryup"),
				spec.Name("\b\f\n\r\t/\\"),
			)),
		},
		{
			name: "unicode_escapes",
			path: `$["fo\u00f8", "tune \uD834\uDD1E"]`,
			exp:  spec.Query(true, spec.Child(spec.Name("foø"), spec.Name("tune 𝄞"))),
		},
		{
			name: "slice_start",
			path: `$[1:]`,
			exp:  spec.Query(true, spec.Child(spec.Slice(1))),
		},
		{
			name: "slice_start_2",
			path: `$[2:]`,
			exp:  spec.Query(true, spec.Child(spec.Slice(2))),
		},
		{
			name: "slice_start_end",
			path: `$[2:6]`,
			exp:  spec.Query(true, spec.Child(spec.Slice(2, 6))),
		},
		{
			name: "slice_end",
			path: `$[:6]`,
			exp:  spec.Query(true, spec.Child(spec.Slice(nil, 6))),
		},
		{
			name: "slice_start_end_step",
			path: `$[2:6:2]`,
			exp:  spec.Query(true, spec.Child(spec.Slice(2, 6, 2))),
		},
		{
			name: "slice_start_step",
			path: `$[2::2]`,
			exp:  spec.Query(true, spec.Child(spec.Slice(2, nil, 2))),
		},
		{
			name: "slice_step",
			path: `$[::2]`,
			exp:  spec.Query(true, spec.Child(spec.Slice(nil, nil, 2))),
		},
		{
			name: "slice_defaults",
			path: `$[:]`,
			exp:  spec.Query(true, spec.Child(spec.Slice())),
		},
		{
			name: "slice_spacing",
			path: `$[   1:  2  : 2   ]`,
			exp:  spec.Query(true, spec.Child(spec.Slice(1, 2, 2))),
		},
		{
			name: "slice_slice",
			path: `$[:,:]`,
			exp:  spec.Query(true, spec.Child(spec.Slice(), spec.Slice())),
		},
		{
			name: "slice_slice_slice",
			path: `$[2:,:4,7:9]`,
			exp: spec.Query(true, spec.Child(
				spec.Slice(2),
				spec.Slice(nil, 4),
				spec.Slice(7, 9),
			)),
		},
		{
			name: "slice_name",
			path: `$[:,"hi"]`,
			exp: spec.Query(
				true,
				spec.Child(spec.Slice(), spec.Name("hi")),
			),
		},
		{
			name: "name_slice",
			path: `$["hi",2:]`,
			exp: spec.Query(
				true,
				spec.Child(spec.Name("hi"), spec.Slice(2)),
			),
		},
		{
			name: "slice_index",
			path: `$[:,42]`,
			exp: spec.Query(
				true,
				spec.Child(spec.Slice(), spec.Index(42)),
			),
		},
		{
			name: "index_slice",
			path: `$[42,:3]`,
			exp: spec.Query(
				true,
				spec.Child(spec.Index(42), spec.Slice(nil, 3)),
			),
		},
		{
			name: "slice_wildcard",
			path: `$[:,   *]`,
			exp:  spec.Query(true, spec.Child(spec.Slice(), spec.Wildcard)),
		},
		{
			name: "wildcard_slice",
			path: `$[  *,  :   ]`,
			exp:  spec.Query(true, spec.Child(spec.Wildcard, spec.Slice())),
		},
		{
			name: "slice_neg_start",
			path: `$[-3:]`,
			exp:  spec.Query(true, spec.Child(spec.Slice(-3))),
		},
		{
			name: "slice_neg_end",
			path: `$[:-3:]`,
			exp:  spec.Query(true, spec.Child(spec.Slice(nil, -3))),
		},
		{
			name: "slice_neg_step",
			path: `$[::-2]`,
			exp:  spec.Query(true, spec.Child(spec.Slice(nil, nil, -2))),
		},
		{
			name: "index_name_slice, wildcard",
			path: `$[3, "🦀", :3,*]`,
			exp: spec.Query(true, spec.Child(
				spec.Index(3),
				spec.Name("🦀"),
				spec.Slice(nil, 3),
				spec.Wildcard,
			)),
		},
		{
			name: "filter_eq",
			path: `$[?@.x == 'y']`,
			exp: spec.Query(true, spec.Child(
				spec.Filter(spec.And(
					spec.Comparison(
						spec.SingularQuery(false, spec.Name("x")),
						spec.EqualTo,
						spec.Literal("y"),
					),
				))),
			),
		},
		{
			name: "filter_exists",
			path: `$[?@.x]`,
			exp: spec.Query(true, spec.Child(
				spec.Filter(spec.And(
					spec.Existence(
						spec.Query(false, spec.Child(spec.Name("x"))),
					),
				)),
			)),
		},
		{
			name: "filter_not_exists",
			path: `$[?!@[0]]`,
			exp: spec.Query(true, spec.Child(
				spec.Filter(spec.And(
					spec.Nonexistence(
						spec.Query(false, spec.Child(spec.Index(0))),
					),
				)),
			)),
		},
		{
			name: "filter_current_and_root",
			path: `$[? @ && $[0]]`,
			exp: spec.Query(true, spec.Child(
				spec.Filter(spec.And(
					spec.Existence(spec.Query(false, []*spec.Segment{}...)),
					spec.Existence(spec.Query(true, spec.Child(spec.Index(0)))),
				)),
			)),
		},
		{
			name: "filter_err",
			path: `$[?`,
			err:  `jsonpath: unexpected eof at position 4`,
		},
		{
			name: "slice_bad_start",
			path: `$[:d]`,
			err:  `jsonpath: unexpected identifier at position 4`,
		},
		{
			name: "slice_four_parts",
			path: `$[0:0:0:0]`,
			err:  `jsonpath: unexpected integer at position 9`,
		},
		{
			name: "invalid_selector",
			path: `$[{}]`,
			err:  `jsonpath: unexpected '{' at position 3`,
		},
		{
			name: "invalid_second_selector",
			path: `$[1, hi]`,
			err:  `jsonpath: unexpected identifier at position 6`,
		},
		{
			name: "missing_segment_comma",
			path: `$[1 "hi"]`,
			err:  `jsonpath: unexpected string at position 5`,
		},
		{
			name: "space_index",
			path: "$[   0]",
			exp:  spec.Query(true, spec.Child(spec.Index(0))),
		},
		{
			name: "index_space_comma_index",
			path: "$[0    , 12]",
			exp: spec.Query(
				true,
				spec.Child(spec.Index(0), spec.Index(12)),
			),
		},
		{
			name: "index_comma_space_name",
			path: `$[0, "xyz"]`,
			exp: spec.Query(
				true,
				spec.Child(spec.Index(0), spec.Name("xyz")),
			),
		},
		{
			name: "tab_index",
			path: "$[\t0]",
			exp:  spec.Query(true, spec.Child(spec.Index(0))),
		},
		{
			name: "newline_index",
			path: "$[\n0]",
			exp:  spec.Query(true, spec.Child(spec.Index(0))),
		},
		{
			name: "return_index",
			path: "$[\r0]",
			exp:  spec.Query(true, spec.Child(spec.Index(0))),
		},
		{
			name: "name_space",
			path: `$["hi"   ]`,
			exp:  spec.Query(true, spec.Child(spec.Name("hi"))),
		},
		{
			name: "wildcard_tab",
			path: "$[*\t]",
			exp:  spec.Query(true, spec.Child(spec.Wildcard)),
		},
		{
			name: "slice_newline",
			path: "$[2:\t]",
			exp:  spec.Query(true, spec.Child(spec.Slice(2))),
		},
		{
			name: "index_return",
			path: "$[0\r]",
			exp:  spec.Query(true, spec.Child(spec.Index(0))),
		},
		{
			name: "descendant_index",
			path: "$..[0]",
			exp:  spec.Query(true, spec.Descendant(spec.Index(0))),
		},
		{
			name: "descendant_name",
			path: `$..["hi"]`,
			exp:  spec.Query(true, spec.Descendant(spec.Name("hi"))),
		},
		{
			name: "descendant_multi",
			path: `$..[  "hi", 2, *, 4:5  ]`,
			exp: spec.Query(true, spec.Descendant(
				spec.Name("hi"),
				spec.Index(2),
				spec.Wildcard,
				spec.Slice(4, 5),
			)),
		},
		{
			name: "invalid_descendant",
			path: "$..[oops]",
			err:  `jsonpath: unexpected identifier at position 5`,
		},
		{
			name: "invalid_unicode_escape",
			path: `$["fo\uu0f8"]`,
			err:  `jsonpath: invalid escape after backslash at position 8`,
		},
		{
			name: "invalid_integer",
			path: `$[170141183460469231731687303715884105727]`, // too large
			err:  `jsonpath: cannot parse "170141183460469231731687303715884105727", value out of range at position 3`,
		},
		{
			name: "invalid_slice_float",
			path: `$[:170141183460469231731687303715884105727]`, // too large
			err:  `jsonpath: cannot parse "170141183460469231731687303715884105727", value out of range at position 4`,
		},
		{
			name: `name_sq_name_desc_wild`,
			path: `$.names['first_name']..*`,
			exp: spec.Query(
				true,
				spec.Child(spec.Name("names")),
				spec.Child(spec.Name("first_name")),
				spec.Descendant(spec.Wildcard),
			),
		},
		{
			name: "no_tail",
			path: `$.a['b']tail`,
			err:  `jsonpath: unexpected identifier at position 9`,
		},
		{
			name: "dq_name",
			path: `$["name"]`,
			exp:  spec.Query(true, spec.Child(spec.Name("name"))),
		},
		{
			name: "sq_name",
			path: `$['name']`,
			exp:  spec.Query(true, spec.Child(spec.Name("name"))),
		},
		{
			name: "two_name_segment",
			path: `$["name","test"]`,
			exp: spec.Query(
				true,
				spec.Child(spec.Name("name"), spec.Name("test")),
			),
		},
		{
			name: "name_index_slice_segment",
			path: `$['name',10,0:3]`,
			exp: spec.Query(
				true,
				spec.Child(spec.Name("name"), spec.Index(10), spec.Slice(0, 3)),
			),
		},
		{
			name: "default_slice_wildcard_segment",
			path: `$[::,*]`,
			exp:  spec.Query(true, spec.Child(spec.Slice(), spec.Wildcard)),
		},
		{
			name: "leading_zero_index",
			path: `$[010]`,
			err:  `jsonpath: invalid number literal at position 3`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			q, err := Parse(reg, tc.path)
			if tc.err == "" {
				r.NoError(err)
				a.Equal(tc.exp, q)
			} else {
				a.Nil(q)
				r.EqualError(err, tc.err)
				r.ErrorIs(err, ErrPathParse)
			}
		})
	}
}

func TestMakeNumErr(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	t.Run("parse_int", func(t *testing.T) {
		t.Parallel()
		_, numErr := strconv.ParseInt("170141183460469231731687303715884105727", 10, 64)
		r.Error(numErr)
		tok := token{invalid, "", 6}
		err := makeNumErr(tok, numErr)
		r.EqualError(
			err,
			`jsonpath: cannot parse "170141183460469231731687303715884105727", value out of range at position 7`,
		)
		r.ErrorIs(err, ErrPathParse)
	})

	t.Run("parse_float", func(t *testing.T) {
		t.Parallel()
		_, numErr := strconv.ParseFloat("99e+1234", 64)
		r.Error(numErr)
		tok := token{invalid, "", 12}
		err := makeNumErr(tok, numErr)
		r.EqualError(
			err,
			`jsonpath: cannot parse "99e+1234", value out of range at position 13`,
		)
		r.ErrorIs(err, ErrPathParse)
	})

	t.Run("other error", func(t *testing.T) {
		t.Parallel()
		myErr := errors.New("oops")
		tok := token{invalid, "", 19}
		err := makeNumErr(tok, myErr)
		r.EqualError(err, `jsonpath: oops at position 20`)
		r.ErrorIs(err, ErrPathParse)
	})
}

func TestParsePathInt(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)

	for _, tc := range []struct {
		name  string
		input string
		exp   int64
		err   string
	}{
		{"zero", "0", 0, ""},
		{"1000", "1000", 1000, ""},
		{"neg_1000", "-1000", -1000, ""},
		{
			name:  "neg_zero",
			input: "-0",
			err:   `jsonpath: invalid integer path value "-0" at position 4`,
		},
		{
			name:  "too_big",
			input: "9007199254740992",
			err:   `jsonpath: cannot parse "9007199254740992", value out of range at position 4`,
		},
		{
			name:  "too_small",
			input: "-9007199254740992",
			err:   `jsonpath: cannot parse "-9007199254740992", value out of range at position 4`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			num, err := parsePathInt(token{integer, tc.input, 3})
			if tc.err == "" {
				r.NoError(err)
				a.Equal(tc.exp, num)
			} else {
				a.Zero(num)
				r.EqualError(err, tc.err)
				r.ErrorIs(err, ErrPathParse)
			}
		})
	}
}

func TestParseLiteral(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)

	for _, tc := range []struct {
		name string
		tok  token
		exp  any
		err  string
	}{
		{
			name: "string",
			tok:  token{goString, "hello", 0},
			exp:  "hello",
		},
		{
			name: "integer",
			tok:  token{integer, "42", 0},
			exp:  int64(42),
		},
		{
			name: "float",
			tok:  token{number, "98.6", 0},
			exp:  float64(98.6),
		},
		{
			name: "true",
			tok:  token{boolTrue, "", 0},
			exp:  true,
		},
		{
			name: "false",
			tok:  token{boolFalse, "", 0},
			exp:  false,
		},
		{
			name: "null",
			tok:  token{jsonNull, "", 0},
			exp:  nil,
		},
		{
			name: "invalid_int",
			tok:  token{integer, "170141183460469231731687303715884105727", 5},
			err:  `jsonpath: cannot parse "170141183460469231731687303715884105727", value out of range at position 6`,
		},
		{
			name: "invalid_float",
			tok:  token{number, "99e+1234", 3},
			err:  `jsonpath: cannot parse "99e+1234", value out of range at position 4`,
		},
		{
			name: "non_literal_token",
			tok:  token{eof, "", 3},
			err:  `jsonpath: unexpected eof at position 4`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			lit, err := parseLiteral(tc.tok)
			if tc.err == "" {
				r.NoError(err)
				a.Equal(spec.Literal(tc.exp), lit)
			} else {
				a.Nil(lit)
				r.EqualError(err, tc.err)
				r.ErrorIs(err, ErrPathParse)
			}
		})
	}
}
