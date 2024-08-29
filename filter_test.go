package jsonpath

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpressionInterface(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		tok  any
	}{
		{"paren", &ParenExpr{}},
		{"not_paren", &NotParenExpr{}},
		{"comparison", &ComparisonExpr{}},
		{"exist", &ExistExpr{}},
		{"not_exist", &NotExistsExpr{}},
		{"func_expr", &FunctionExpr{}},
		{"not_func_expr", &NotFuncExpr{}},
		{"logical_and", LogicalOrExpr{}},
		{"logical_or", LogicalAndExpr{}},
		{"value", &ValueType{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Implements((*basicExpr)(nil), tc.tok)
		})
	}
}

func TestLogicalAndExpr(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name    string
		expr    []basicExpr
		root    any
		current any
		exp     bool
		str     string
	}{
		{
			name:    "no_expr",
			expr:    []basicExpr{},
			current: map[string]any{"x": 0},
			exp:     true,
			str:     "",
		},
		{
			name: "one_true_expr",
			expr: []basicExpr{&ExistExpr{
				&Query{segments: []*Segment{Child(Name("x"))}, root: false},
			}},
			current: map[string]any{"x": 0},
			exp:     true,
			str:     `@["x"]`,
		},
		{
			name: "one_false_expr",
			expr: []basicExpr{&ExistExpr{
				&Query{segments: []*Segment{Child(Name("y"))}, root: true},
			}},
			root: map[string]any{"x": 0},
			exp:  false,
			str:  `$["y"]`,
		},
		{
			name: "two_true_expr",
			expr: []basicExpr{
				&ExistExpr{&Query{segments: []*Segment{Child(Name("x"))}}},
				&ExistExpr{&Query{segments: []*Segment{Child(Name("y"))}}},
			},
			current: map[string]any{"x": 0, "y": 1},
			exp:     true,
			str:     `@["x"] && @["y"]`,
		},
		{
			name: "one_true_one_false",
			expr: []basicExpr{
				&ExistExpr{&Query{segments: []*Segment{Child(Name("x"))}}},
				&ExistExpr{&Query{segments: []*Segment{Child(Name("y"))}}},
			},
			current: map[string]any{"x": 0, "z": 1},
			exp:     false,
			str:     `@["x"] && @["y"]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			andExpr := LogicalAndExpr(tc.expr)
			a.Equal(tc.exp, andExpr.testFilter(tc.current, tc.root))
			a.Equal(tc.str, bufString(andExpr))
		})
	}
}

func TestLogicalOrExpr(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name    string
		expr    []LogicalAndExpr
		root    any
		current any
		exp     bool
		str     string
	}{
		{
			name:    "no_expr",
			expr:    []LogicalAndExpr{LogicalAndExpr([]basicExpr{})},
			current: map[string]any{"x": 0},
			exp:     true,
			str:     "",
		},
		{
			name: "one_expr",
			expr: []LogicalAndExpr{LogicalAndExpr([]basicExpr{&ExistExpr{
				&Query{segments: []*Segment{Child(Name("x"))}, root: true},
			}})},
			root: map[string]any{"x": 0},
			exp:  true,
			str:  `$["x"]`,
		},
		{
			name: "one_false_expr",
			expr: []LogicalAndExpr{LogicalAndExpr([]basicExpr{&ExistExpr{
				&Query{segments: []*Segment{Child(Name("x"))}},
			}})},
			current: map[string]any{"y": 0},
			exp:     false,
			str:     `@["x"]`,
		},
		{
			name: "two_true_expr",
			expr: []LogicalAndExpr{
				LogicalAndExpr([]basicExpr{
					&ExistExpr{&Query{segments: []*Segment{Child(Name("x"))}}},
				}),
				LogicalAndExpr([]basicExpr{
					&ExistExpr{&Query{segments: []*Segment{Child(Name("y"))}}},
				}),
			},
			current: map[string]any{"x": 0, "y": "hi"},
			exp:     true,
			str:     `@["x"] || @["y"]`,
		},
		{
			name: "one_true_one_false",
			expr: []LogicalAndExpr{
				LogicalAndExpr([]basicExpr{
					&ExistExpr{&Query{segments: []*Segment{Child(Name("x"))}}},
				}),
				LogicalAndExpr([]basicExpr{
					&ExistExpr{&Query{segments: []*Segment{Child(Name("y"))}}},
				}),
			},
			current: map[string]any{"x": 0, "z": "hi"},
			exp:     true,
			str:     `@["x"] || @["y"]`,
		},
		{
			name: "nested_ands",
			expr: []LogicalAndExpr{
				LogicalAndExpr([]basicExpr{
					&ExistExpr{&Query{segments: []*Segment{Child(Name("x"))}}},
					&ExistExpr{&Query{segments: []*Segment{Child(Name("y"))}}},
				}),
				LogicalAndExpr([]basicExpr{
					&ExistExpr{&Query{segments: []*Segment{Child(Name("y"))}}},
					&ExistExpr{&Query{segments: []*Segment{Child(Name("x"))}}},
				}),
			},
			current: map[string]any{"x": 0, "y": "hi"},
			exp:     true,
			str:     `@["x"] && @["y"] || @["y"] && @["x"]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			orExpr := LogicalOrExpr(tc.expr)
			a.Equal(FuncLogical, orExpr.asTypeKind())
			a.Equal(tc.exp, orExpr.testFilter(tc.current, tc.root))
			a.Equal(logicalFrom(tc.exp), orExpr.execute(tc.current, tc.root))
			a.Equal(tc.str, bufString(orExpr))

			// Test ParenExpr.
			pExpr := &ParenExpr{orExpr}
			a.Equal(tc.exp, pExpr.testFilter(tc.current, tc.root))
			a.Equal("("+tc.str+")", bufString(pExpr))

			// Test NotParenExpr.
			npExpr := &NotParenExpr{orExpr}
			a.Equal(!tc.exp, npExpr.testFilter(tc.current, tc.root))
			a.Equal("!("+tc.str+")", bufString(npExpr))
		})
	}
}

func TestExistExpr(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name    string
		query   *Query
		root    any
		current any
		exp     bool
	}{
		{
			name:    "current_name",
			query:   &Query{segments: []*Segment{Child(Name("x"))}, root: false},
			current: map[string]any{"x": 0},
			exp:     true,
		},
		{
			name:  "root_name",
			query: &Query{segments: []*Segment{Child(Name("x"))}, root: true},
			root:  map[string]any{"x": 0},
			exp:   true,
		},
		{
			name:    "current_false",
			query:   &Query{segments: []*Segment{Child(Name("x"))}, root: false},
			current: map[string]any{"y": 0},
			exp:     false,
		},
		{
			name:  "root_false",
			query: &Query{segments: []*Segment{Child(Name("x"))}, root: true},
			root:  map[string]any{"y": 0},
			exp:   false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Test existExpr.
			exist := ExistExpr{tc.query}
			a.Equal(tc.exp, exist.testFilter(tc.current, tc.root))
			buf := new(strings.Builder)
			exist.writeTo(buf)
			a.Equal(tc.query.String(), buf.String())

			// Test notExistExpr.
			ne := NotExistsExpr{tc.query}
			a.Equal(!tc.exp, ne.testFilter(tc.current, tc.root))
			buf.Reset()
			ne.writeTo(buf)
			a.Equal("!"+tc.query.String(), buf.String())
		})
	}
}
