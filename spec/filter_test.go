package spec

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpressionInterface(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
		tok  any
	}{
		{"paren", Paren(nil)},
		{"not_paren", NotParen(nil)},
		{"comparison", Comparison(nil, EqualTo, nil)},
		{"exist", Existence(nil)},
		{"not_exist", Nonexistence(nil)},
		{"func_expr", &FuncExpr{}},
		{"not_func_expr", &NotFuncExpr{}},
		{"logical_and", LogicalOr{}},
		{"logical_or", LogicalAnd{}},
		{"value", Value(nil)},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			assert.Implements(t, (*BasicExpr)(nil), tc.tok)
		})
	}
}

func TestLogicalAndExpr(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test    string
		expr    []BasicExpr
		root    any
		current any
		exp     bool
		str     string
	}{
		{
			test:    "no_expr",
			expr:    []BasicExpr{},
			current: map[string]any{"x": 0},
			exp:     true,
			str:     "",
		},
		{
			test: "one_true_expr",
			expr: []BasicExpr{
				Existence(Query(false, Child(Name("x")))),
			},
			current: map[string]any{"x": 0},
			exp:     true,
			str:     `@["x"]`,
		},
		{
			test: "one_false_expr",
			expr: []BasicExpr{
				Existence(Query(true, Child(Name("y")))),
			},
			root: map[string]any{"x": 0},
			exp:  false,
			str:  `$["y"]`,
		},
		{
			test: "two_true_expr",
			expr: []BasicExpr{
				Existence(Query(false, Child(Name("x")))),
				Existence(Query(false, Child(Name("y")))),
			},
			current: map[string]any{"x": 0, "y": 1},
			exp:     true,
			str:     `@["x"] && @["y"]`,
		},
		{
			test: "one_true_one_false",
			expr: []BasicExpr{
				Existence(Query(false, Child(Name("x")))),
				Existence(Query(false, Child(Name("y")))),
			},
			current: map[string]any{"x": 0, "z": 1},
			exp:     false,
			str:     `@["x"] && @["y"]`,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			andExpr := LogicalAnd(tc.expr)
			a.Equal(tc.exp, andExpr.testFilter(tc.current, tc.root, nil))
			a.Equal(tc.str, bufString(andExpr))
		})
	}
}

func TestLogicalOrExpr(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test    string
		expr    []LogicalAnd
		root    any
		current any
		exp     bool
		str     string
	}{
		{
			test:    "no_expr",
			expr:    []LogicalAnd{{}},
			current: map[string]any{"x": 0},
			exp:     true,
			str:     "",
		},
		{
			test: "one_expr",
			expr: []LogicalAnd{{Existence(
				Query(true, Child(Name("x"))),
			)}},
			root: map[string]any{"x": 0},
			exp:  true,
			str:  `$["x"]`,
		},
		{
			test: "one_false_expr",
			expr: []LogicalAnd{{Existence(
				Query(false, Child(Name("x"))),
			)}},
			current: map[string]any{"y": 0},
			exp:     false,
			str:     `@["x"]`,
		},
		{
			test: "two_true_expr",
			expr: []LogicalAnd{
				{Existence(Query(false, Child(Name("x"))))},
				{Existence(Query(false, Child(Name("y"))))},
			},
			current: map[string]any{"x": 0, "y": "hi"},
			exp:     true,
			str:     `@["x"] || @["y"]`,
		},
		{
			test: "one_true_one_false",
			expr: []LogicalAnd{
				{Existence(Query(false, Child(Name("x"))))},
				{Existence(Query(false, Child(Name("y"))))},
			},
			current: map[string]any{"x": 0, "z": "hi"},
			exp:     true,
			str:     `@["x"] || @["y"]`,
		},
		{
			test: "nested_ands",
			expr: []LogicalAnd{
				{
					Existence(Query(false, Child(Name("x")))),
					Existence(Query(false, Child(Name("y")))),
				},
				{
					Existence(Query(false, Child(Name("y")))),
					Existence(Query(false, Child(Name("x")))),
				},
			},
			current: map[string]any{"x": 0, "y": "hi"},
			exp:     true,
			str:     `@["x"] && @["y"] || @["y"] && @["x"]`,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			orExpr := LogicalOr(tc.expr)
			a.Equal(FuncLogical, orExpr.ResultType())
			a.Equal(tc.exp, orExpr.testFilter(tc.current, tc.root, nil))
			a.Equal(Logical(tc.exp), orExpr.evaluate(tc.current, tc.root, nil))
			a.Equal(tc.str, bufString(orExpr))
			a.True(orExpr.ConvertsTo(FuncLogical))
			a.False(orExpr.ConvertsTo(FuncValue))
			a.False(orExpr.ConvertsTo(FuncNodes))

			// Test ParenExpr.
			pExpr := Paren(orExpr...)
			a.Equal(tc.exp, pExpr.testFilter(tc.current, tc.root, nil))
			a.Equal("("+tc.str+")", bufString(pExpr))

			// Test NotParenExpr.
			npExpr := NotParen(orExpr...)
			a.Equal(!tc.exp, npExpr.testFilter(tc.current, tc.root, nil))
			a.Equal("!("+tc.str+")", bufString(npExpr))
		})
	}
}

func TestExistExpr(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test    string
		query   *PathQuery
		root    any
		current any
		exp     bool
	}{
		{
			test:    "current_name",
			query:   Query(false, Child(Name("x"))),
			current: map[string]any{"x": 0},
			exp:     true,
		},
		{
			test:  "root_name",
			query: Query(true, Child(Name("x"))),
			root:  map[string]any{"x": 0},
			exp:   true,
		},
		{
			test:    "current_false",
			query:   Query(false, Child(Name("x"))),
			current: map[string]any{"y": 0},
			exp:     false,
		},
		{
			test:  "root_false",
			query: Query(true, Child(Name("x"))),
			root:  map[string]any{"y": 0},
			exp:   false,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			// Test existExpr.
			exist := ExistExpr{tc.query}
			a.Equal(tc.exp, exist.testFilter(tc.current, tc.root, nil))
			buf := new(strings.Builder)
			exist.writeTo(buf)
			a.Equal(tc.query.String(), buf.String())

			// Test NonExistExpr.
			ne := NonExistExpr{tc.query}
			a.Equal(!tc.exp, ne.testFilter(tc.current, tc.root, nil))
			buf.Reset()
			ne.writeTo(buf)
			a.Equal("!"+tc.query.String(), buf.String())
		})
	}
}
