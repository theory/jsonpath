package spec

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectorInterface(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		tok  any
	}{
		{"name", Name("hi")},
		{"index", Index(42)},
		{"slice", Slice()},
		{"wildcard", Wildcard()},
		{"filter", Filter(nil)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Implements(t, (*Selector)(nil), tc.tok)
		})
	}
}

func TestSelectorString(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		tok  Selector
		str  string
		sing bool
	}{
		{
			name: "name",
			tok:  Name("hi"),
			str:  `"hi"`,
			sing: true,
		},
		{
			name: "name_space",
			tok:  Name("hi there"),
			str:  `"hi there"`,
			sing: true,
		},
		{
			name: "name_quote",
			tok:  Name(`hi "there"`),
			str:  `"hi \"there\""`,
			sing: true,
		},
		{
			name: "name_unicode",
			tok:  Name(`hi 😀`),
			str:  `"hi 😀"`,
			sing: true,
		},
		{
			name: "name_digits",
			tok:  Name(`42`),
			str:  `"42"`,
			sing: true,
		},
		{
			name: "index",
			tok:  Index(42),
			str:  "42",
			sing: true,
		},
		{
			name: "index_big",
			tok:  Index(math.MaxUint32),
			str:  "4294967295",
			sing: true,
		},
		{
			name: "index_zero",
			tok:  Index(0),
			str:  "0",
			sing: true,
		},
		{
			name: "slice_0_4",
			tok:  Slice(0, 4),
			str:  ":4",
		},
		{
			name: "slice_4_5",
			tok:  Slice(4, 5),
			str:  "4:5",
		},
		{
			name: "slice_end_42",
			tok:  Slice(nil, 42),
			str:  ":42",
		},
		{
			name: "slice_start_4",
			tok:  Slice(4),
			str:  "4:",
		},
		{
			name: "slice_start_end_step",
			tok:  Slice(4, 7, 2),
			str:  "4:7:2",
		},
		{
			name: "slice_start_step",
			tok:  Slice(4, nil, 2),
			str:  "4::2",
		},
		{
			name: "slice_end_step",
			tok:  Slice(nil, 4, 2),
			str:  ":4:2",
		},
		{
			name: "slice_step",
			tok:  Slice(nil, nil, 3),
			str:  "::3",
		},
		{
			name: "slice_neg_step",
			tok:  Slice(nil, nil, -1),
			str:  "::-1",
		},
		{
			name: "slice_max_start",
			tok:  Slice(math.MaxInt),
			str:  fmt.Sprintf("%v:", math.MaxInt),
		},
		{
			name: "slice_max_start_neg_step",
			tok:  Slice(math.MaxInt, nil, -1),
			str:  "::-1",
		},
		{
			name: "slice_min_start",
			tok:  Slice(math.MinInt),
			str:  fmt.Sprintf("%v:", math.MinInt),
		},
		{
			name: "slice_min_start_neg_step",
			tok:  Slice(math.MinInt, nil, -1),
			str:  fmt.Sprintf("%v::-1", math.MinInt),
		},
		{
			name: "slice_max_end",
			tok:  Slice(0, math.MaxInt),
			str:  ":",
		},
		{
			name: "slice_max_end_neg_step",
			tok:  Slice(0, math.MaxInt, -1),
			str:  "::-1",
		},
		{
			name: "slice_min_end",
			tok:  Slice(0, math.MinInt),
			str:  fmt.Sprintf(":%v", math.MinInt),
		},
		{
			name: "slice_min_end_neg_step",
			tok:  Slice(0, math.MinInt, -1),
			str:  "::-1",
		},
		{
			name: "wildcard",
			tok:  Wildcard(),
			str:  "*",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.Equal(tc.sing, tc.tok.isSingular())
			buf := new(strings.Builder)
			tc.tok.writeTo(buf)
			a.Equal(tc.str, buf.String())
			a.Equal(tc.str, tc.tok.String())
		})
	}
}

func TestSliceBounds(t *testing.T) {
	t.Parallel()

	json := []any{"a", "b", "c", "d", "e", "f", "g"}

	extract := func(s SliceSelector) []any {
		lower, upper := s.Bounds(len(json))
		res := make([]any, 0, len(json))
		switch {
		case s.step > 0:
			for i := lower; i < upper; i += s.step {
				res = append(res, json[i])
			}
		case s.step < 0:
			for i := upper; lower < i; i += s.step {
				res = append(res, json[i])
			}
		}
		return res
	}

	type lenCase struct {
		length int
		lower  int
		upper  int
	}

	for _, tc := range []struct {
		name  string
		slice SliceSelector
		cases []lenCase
		exp   []any
	}{
		{
			name:  "defaults",
			slice: Slice(),
			exp:   json,
			cases: []lenCase{
				{10, 0, 10},
				{3, 0, 3},
				{99, 0, 99},
			},
		},
		{
			name:  "step_0",
			slice: Slice(nil, nil, 0),
			exp:   []any{},
			cases: []lenCase{
				{10, 0, 0},
				{3, 0, 0},
				{99, 0, 0},
			},
		},
		{
			name:  "nil_nil_nil",
			slice: Slice(nil, nil, nil),
			exp:   json,
			cases: []lenCase{
				{10, 0, 10},
				{3, 0, 3},
				{99, 0, 99},
			},
		},
		{
			name:  "3_8_2",
			slice: Slice(3, 8, 2),
			exp:   []any{"d", "f"},
			cases: []lenCase{
				{10, 3, 8},
				{3, 3, 3},
				{99, 3, 8},
			},
		},
		{
			name:  "1_3_1",
			slice: Slice(1, 3, 1),
			exp:   []any{"b", "c"},
			cases: []lenCase{
				{10, 1, 3},
				{2, 1, 2},
				{99, 1, 3},
			},
		},
		{
			name:  "5_defaults",
			slice: Slice(5),
			exp:   []any{"f", "g"},
			cases: []lenCase{
				{10, 5, 10},
				{8, 5, 8},
				{99, 5, 99},
			},
		},
		{
			name:  "1_5_2",
			slice: Slice(1, 5, 2),
			exp:   []any{"b", "d"},
			cases: []lenCase{
				{10, 1, 5},
				{4, 1, 4},
				{99, 1, 5},
			},
		},
		{
			name:  "5_1_neg2",
			slice: Slice(5, 1, -2),
			exp:   []any{"f", "d"},
			cases: []lenCase{
				{10, 1, 5},
				{4, 1, 3},
				{99, 1, 5},
			},
		},
		{
			name:  "def_def_neg1",
			slice: Slice(nil, nil, -1),
			exp:   []any{"g", "f", "e", "d", "c", "b", "a"},
			cases: []lenCase{
				{10, -1, 9},
				{4, -1, 3},
				{99, -1, 98},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.False(tc.slice.isSingular())
			for _, lc := range tc.cases {
				lower, upper := tc.slice.Bounds(lc.length)
				a.Equal(lc.lower, lower)
				a.Equal(lc.upper, upper)
			}
			a.Equal(tc.exp, extract(tc.slice))
			a.Equal(tc.slice.start, tc.slice.Start())
			a.Equal(tc.slice.end, tc.slice.End())
			a.Equal(tc.slice.step, tc.slice.Step())
		})
	}
}

func TestSlicePanic(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	a.PanicsWithValue(
		"First value passed to Slice is not an integer",
		func() { Slice("hi") },
	)
	a.PanicsWithValue(
		"Second value passed to Slice is not an integer",
		func() { Slice(nil, "hi") },
	)
	a.PanicsWithValue(
		"Third value passed to Slice is not an integer",
		func() { Slice(nil, 42, "hi") },
	)
}

func TestNameSelect(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		sel  Name
		src  any
		exp  []any
		loc  []*LocatedNode
	}{
		{
			name: "got_name",
			sel:  Name("hi"),
			src:  map[string]any{"hi": 42},
			exp:  []any{42},
			loc:  []*LocatedNode{{Path: Normalized(Name("hi")), Node: 42}},
		},
		{
			name: "got_name_array",
			sel:  Name("hi"),
			src:  map[string]any{"hi": []any{42, true}},
			exp:  []any{[]any{42, true}},
			loc:  []*LocatedNode{{Path: Normalized(Name("hi")), Node: []any{42, true}}},
		},
		{
			name: "no_name",
			sel:  Name("hi"),
			src:  map[string]any{"oy": []any{42, true}},
			exp:  []any{},
			loc:  []*LocatedNode{},
		},
		{
			name: "src_array",
			sel:  Name("hi"),
			src:  []any{42, true},
			exp:  []any{},
			loc:  []*LocatedNode{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.Equal(tc.exp, tc.sel.Select(tc.src, nil))
			a.Equal(tc.loc, tc.sel.SelectLocated(tc.src, nil, Normalized()))
		})
	}
}

func TestIndexSelect(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		sel  Index
		src  any
		exp  []any
		loc  []*LocatedNode
	}{
		{
			name: "index_zero",
			sel:  Index(0),
			src:  []any{42, true, "hi"},
			exp:  []any{42},
			loc:  []*LocatedNode{{Path: Normalized(Index(0)), Node: 42}},
		},
		{
			name: "index_two",
			sel:  Index(2),
			src:  []any{42, true, "hi"},
			exp:  []any{"hi"},
			loc:  []*LocatedNode{{Path: Normalized(Index(2)), Node: "hi"}},
		},
		{
			name: "index_neg_one",
			sel:  Index(-1),
			src:  []any{42, true, "hi"},
			exp:  []any{"hi"},
			loc:  []*LocatedNode{{Path: Normalized(Index(2)), Node: "hi"}},
		},
		{
			name: "index_neg_two",
			sel:  Index(-2),
			src:  []any{42, true, "hi"},
			exp:  []any{true},
			loc:  []*LocatedNode{{Path: Normalized(Index(1)), Node: true}},
		},
		{
			name: "out_of_range",
			sel:  Index(4),
			src:  []any{42, true, "hi"},
			exp:  []any{},
			loc:  []*LocatedNode{},
		},
		{
			name: "neg_out_of_range",
			sel:  Index(-4),
			src:  []any{42, true, "hi"},
			exp:  []any{},
			loc:  []*LocatedNode{},
		},
		{
			name: "src_object",
			sel:  Index(0),
			src:  map[string]any{"hi": 42},
			exp:  []any{},
			loc:  []*LocatedNode{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.Equal(tc.exp, tc.sel.Select(tc.src, nil))
			a.Equal(tc.loc, tc.sel.SelectLocated(tc.src, nil, Normalized()))
		})
	}
}

func TestWildcardSelect(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		src  any
		exp  []any
		loc  []*LocatedNode
	}{
		{
			name: "object",
			src:  map[string]any{"x": true, "y": []any{true}},
			exp:  []any{true, []any{true}},
			loc: []*LocatedNode{
				{Path: Normalized(Name("x")), Node: true},
				{Path: Normalized(Name("y")), Node: []any{true}},
			},
		},
		{
			name: "array",
			src:  []any{true, 42, map[string]any{"x": 6}},
			exp:  []any{true, 42, map[string]any{"x": 6}},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: true},
				{Path: Normalized(Index(1)), Node: 42},
				{Path: Normalized(Index(2)), Node: map[string]any{"x": 6}},
			},
		},
		{
			name: "something_else",
			src:  42,
			exp:  []any{},
			loc:  []*LocatedNode{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			if _, ok := tc.src.(map[string]any); ok {
				a.ElementsMatch(tc.exp, Wildcard().Select(tc.src, nil))
				a.ElementsMatch(tc.loc, Wildcard().SelectLocated(tc.src, nil, Normalized()))
			} else {
				a.Equal(tc.exp, Wildcard().Select(tc.src, nil))
				a.Equal(tc.loc, Wildcard().SelectLocated(tc.src, nil, Normalized()))
			}
		})
	}
}

func TestSliceSelect(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		sel  SliceSelector
		src  any
		exp  []any
		loc  []*LocatedNode
	}{
		{
			name: "0_2",
			sel:  Slice(0, 2),
			src:  []any{42, true, "hi"},
			exp:  []any{42, true},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: 42},
				{Path: Normalized(Index(1)), Node: true},
			},
		},
		{
			name: "0_1",
			sel:  Slice(0, 1),
			src:  []any{[]any{42, false}, true, "hi"},
			exp:  []any{[]any{42, false}},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: []any{42, false}},
			},
		},
		{
			name: "2_5",
			sel:  Slice(2, 5),
			src:  []any{[]any{42, false}, true, "hi", 98.6, 73, "hi", 22},
			exp:  []any{"hi", 98.6, 73},
			loc: []*LocatedNode{
				{Path: Normalized(Index(2)), Node: "hi"},
				{Path: Normalized(Index(3)), Node: 98.6},
				{Path: Normalized(Index(4)), Node: 73},
			},
		},
		{
			name: "2_5_over_len",
			sel:  Slice(2, 5),
			src:  []any{"x", true, "y"},
			exp:  []any{"y"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(2)), Node: "y"},
			},
		},
		{
			name: "defaults",
			sel:  Slice(),
			src:  []any{"x", nil, "y", 42},
			exp:  []any{"x", nil, "y", 42},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: "x"},
				{Path: Normalized(Index(1)), Node: nil},
				{Path: Normalized(Index(2)), Node: "y"},
				{Path: Normalized(Index(3)), Node: 42},
			},
		},
		{
			name: "default_start",
			sel:  Slice(nil, 3),
			src:  []any{"x", nil, "y", 42, 98.6, 54},
			exp:  []any{"x", nil, "y"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: "x"},
				{Path: Normalized(Index(1)), Node: nil},
				{Path: Normalized(Index(2)), Node: "y"},
			},
		},
		{
			name: "default_end",
			sel:  Slice(2),
			src:  []any{"x", true, "y", 42, 98.6, 54},
			exp:  []any{"y", 42, 98.6, 54},
			loc: []*LocatedNode{
				{Path: Normalized(Index(2)), Node: "y"},
				{Path: Normalized(Index(3)), Node: 42},
				{Path: Normalized(Index(4)), Node: 98.6},
				{Path: Normalized(Index(5)), Node: 54},
			},
		},
		{
			name: "step_2",
			sel:  Slice(nil, nil, 2),
			src:  []any{"x", true, "y", 42, 98.6, 54},
			exp:  []any{"x", "y", 98.6},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: "x"},
				{Path: Normalized(Index(2)), Node: "y"},
				{Path: Normalized(Index(4)), Node: 98.6},
			},
		},
		{
			name: "step_3",
			sel:  Slice(nil, nil, 3),
			src:  []any{"x", true, "y", 42, 98.6, 54, 98, 73},
			exp:  []any{"x", 42, 98},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: "x"},
				{Path: Normalized(Index(3)), Node: 42},
				{Path: Normalized(Index(6)), Node: 98},
			},
		},
		{
			name: "negative_step",
			sel:  Slice(nil, nil, -1),
			src:  []any{"x", true, "y", []any{1, 2}},
			exp:  []any{[]any{1, 2}, "y", true, "x"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(3)), Node: []any{1, 2}},
				{Path: Normalized(Index(2)), Node: "y"},
				{Path: Normalized(Index(1)), Node: true},
				{Path: Normalized(Index(0)), Node: "x"},
			},
		},
		{
			name: "5_0_neg2",
			sel:  Slice(5, 0, -2),
			src:  []any{"x", true, "y", 8, 13, 25, 23, 78, 13},
			exp:  []any{25, 8, true},
			loc: []*LocatedNode{
				{Path: Normalized(Index(5)), Node: 25},
				{Path: Normalized(Index(3)), Node: 8},
				{Path: Normalized(Index(1)), Node: true},
			},
		},
		{
			name: "src_object",
			sel:  Slice(0, 2),
			src:  map[string]any{"hi": 42},
			exp:  []any{},
			loc:  []*LocatedNode{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.Equal(tc.exp, tc.sel.Select(tc.src, nil))
			a.Equal(tc.loc, tc.sel.SelectLocated(tc.src, nil, Normalized()))
		})
	}
}

func TestFilterSelector(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		filter  *FilterSelector
		root    any
		current any
		exp     []any
		loc     []*LocatedNode
		str     string
		rand    bool
	}{
		{
			name:   "no_filter",
			filter: Filter(),
			exp:    []any{},
			loc:    []*LocatedNode{},
			str:    "?",
		},
		{
			name:    "array_root",
			filter:  Filter(And(Existence(Query(true, Child(Index(0)))))),
			root:    []any{42, true, "hi"},
			current: map[string]any{"x": 2},
			exp:     []any{2},
			loc: []*LocatedNode{
				{Path: Normalized(Name("x")), Node: 2},
			},
			str: `?$[0]`,
		},
		{
			name:    "array_root_false",
			filter:  Filter(And(Existence(Query(true, Child(Index(4)))))),
			root:    []any{42, true, "hi"},
			current: map[string]any{"x": 2},
			exp:     []any{},
			loc:     []*LocatedNode{},
			str:     `?$[4]`,
		},
		{
			name:    "object_root",
			filter:  Filter(And(Existence(Query(true, Child(Name("y")))))),
			root:    map[string]any{"x": 42, "y": "hi"},
			current: map[string]any{"a": 2, "b": 3},
			exp:     []any{2, 3},
			loc: []*LocatedNode{
				{Path: Normalized(Name("a")), Node: 2},
				{Path: Normalized(Name("b")), Node: 3},
			},
			str:  `?$["y"]`,
			rand: true,
		},
		{
			name:    "object_root_false",
			filter:  Filter(And(Existence(Query(true, Child(Name("z")))))),
			root:    map[string]any{"x": 42, "y": "hi"},
			current: map[string]any{"a": 2, "b": 3},
			exp:     []any{},
			loc:     []*LocatedNode{},
			str:     `?$["z"]`,
			rand:    true,
		},
		{
			name:    "array_current",
			filter:  Filter(And(Existence(Query(false, Child(Index(0)))))),
			current: []any{[]any{42}},
			exp:     []any{[]any{42}},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: []any{42}},
			},
			str: `?@[0]`,
		},
		{
			name:    "array_current_false",
			filter:  Filter(And(Existence(Query(false, Child(Index(1)))))),
			current: []any{[]any{42}},
			exp:     []any{},
			loc:     []*LocatedNode{},
			str:     `?@[1]`,
		},
		{
			name:    "object_current",
			filter:  Filter(And(Existence(Query(false, Child(Name("x")))))),
			current: []any{map[string]any{"x": 42}},
			exp:     []any{map[string]any{"x": 42}},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: map[string]any{"x": 42}},
			},
			str: `?@["x"]`,
		},
		{
			name:    "object_current_false",
			filter:  Filter(And(Existence(Query(false, Child(Name("y")))))),
			current: []any{map[string]any{"x": 42}},
			exp:     []any{},
			loc:     []*LocatedNode{},
			str:     `?@["y"]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			if tc.rand {
				a.ElementsMatch(tc.exp, tc.filter.Select(tc.current, tc.root))
				a.ElementsMatch(tc.loc, tc.filter.SelectLocated(tc.current, tc.root, Normalized()))
			} else {
				a.Equal(tc.exp, tc.filter.Select(tc.current, tc.root))
				a.Equal(tc.loc, tc.filter.SelectLocated(tc.current, tc.root, Normalized()))
			}
			a.Equal(tc.str, tc.filter.String())
			a.Equal(tc.str, bufString(tc.filter))
			a.False(tc.filter.isSingular())
		})
	}
}
