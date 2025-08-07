package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSegmentString(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
		seg  *Segment
		str  string
		sing bool
	}{
		{
			test: "no_selectors",
			seg:  Child(),
			str:  "[]",
		},
		{
			test: "descendant_no_selectors",
			seg:  Descendant(),
			str:  "..[]",
		},
		{
			test: "name",
			seg:  Child(Name("hi")),
			str:  `["hi"]`,
			sing: true,
		},
		{
			test: "index",
			seg:  Child(Index(2)),
			str:  `[2]`,
			sing: true,
		},
		{
			test: "wildcard",
			seg:  Child(Wildcard()),
			str:  `[*]`,
		},
		{
			test: "slice",
			seg:  Child(Slice(2)),
			str:  `[2:]`,
		},
		{
			test: "multiples",
			seg:  Child(Slice(2), Name("hi"), Index(3)),
			str:  `[2:,"hi",3]`,
		},
		{
			test: "descendant_multiples",
			seg:  Descendant(Slice(2), Name("hi"), Index(3)),
			str:  `..[2:,"hi",3]`,
		},
		{
			test: "wildcard_override",
			seg:  Child(Slice(2), Name("hi"), Index(3), Wildcard()),
			str:  `[2:,"hi",3,*]`,
		},
		{
			test: "descendant_wildcard_override",
			seg:  Descendant(Slice(2), Name("hi"), Index(3), Wildcard()),
			str:  `..[2:,"hi",3,*]`,
		},
		{
			test: "dupes",
			seg:  Child(Slice(2), Name("hi"), Slice(2), Slice(2), Name("hi"), Index(3), Name("go"), Index(3)),
			str:  `[2:,"hi",2:,2:,"hi",3,"go",3]`,
		},
		{
			test: "descendant_dupes",
			seg:  Descendant(Slice(2), Name("hi"), Slice(2), Slice(2), Name("hi"), Index(3), Name("go"), Index(3)),
			str:  `..[2:,"hi",2:,2:,"hi",3,"go",3]`,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.Equal(tc.str, tc.seg.String())
			a.Equal(tc.sing, tc.seg.isSingular())
			a.Equal(tc.seg.descendant, tc.seg.IsDescendant())
		})
	}
}

func TestSegmentSelect(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
		seg  *Segment
		src  any
		exp  []any
		loc  []*LocatedNode
		rand bool
		sing bool
	}{
		{
			test: "no_selectors",
			seg:  Child(),
			src:  []any{1, 3},
			exp:  []any{},
			loc:  []*LocatedNode{},
		},
		{
			test: "name",
			seg:  Child(Name("hi")),
			src:  map[string]any{"hi": 42, "go": 98.6, "x": true},
			exp:  []any{42},
			loc: []*LocatedNode{
				{Path: Normalized(Name("hi")), Node: 42},
			},
			sing: true,
		},
		{
			test: "two_names",
			seg:  Child(Name("hi"), Name("go")),
			src:  map[string]any{"hi": 42, "go": 98.6, "x": true},
			exp:  []any{42, 98.6},
			loc: []*LocatedNode{
				{Path: Normalized(Name("hi")), Node: 42},
				{Path: Normalized(Name("go")), Node: 98.6},
			},
			rand: true,
		},
		{
			test: "dupe_name",
			seg:  Child(Name("hi"), Name("hi")),
			src:  map[string]any{"hi": 42, "go": 98.6, "x": true},
			exp:  []any{42, 42},
			loc: []*LocatedNode{
				{Path: Normalized(Name("hi")), Node: 42},
				{Path: Normalized(Name("hi")), Node: 42},
			},
			rand: true,
		},
		{
			test: "three_names",
			seg:  Child(Name("hi"), Name("go"), Name("x")),
			src:  map[string]any{"hi": 42, "go": 98.6, "x": true},
			exp:  []any{42, 98.6, true},
			loc: []*LocatedNode{
				{Path: Normalized(Name("hi")), Node: 42},
				{Path: Normalized(Name("go")), Node: 98.6},
				{Path: Normalized(Name("x")), Node: true},
			},
			rand: true,
		},
		{
			test: "name_and_others",
			seg:  Child(Name("hi"), Index(1), Slice()),
			src:  map[string]any{"hi": 42, "go": 98.6, "x": true},
			exp:  []any{42},
			loc: []*LocatedNode{
				{Path: Normalized(Name("hi")), Node: 42},
			},
		},
		{
			test: "index",
			seg:  Child(Index(1)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 42},
			},
			sing: true,
		},
		{
			test: "two_indexes",
			seg:  Child(Index(1), Index(4)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "x"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 42},
				{Path: Normalized(Index(4)), Node: "x"},
			},
		},
		{
			test: "dupe_index",
			seg:  Child(Index(1), Index(1)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, 42},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 42},
				{Path: Normalized(Index(1)), Node: 42},
			},
		},
		{
			test: "three_indexes",
			seg:  Child(Index(1), Index(4), Index(0)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "x", "hi"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 42},
				{Path: Normalized(Index(4)), Node: "x"},
				{Path: Normalized(Index(0)), Node: "hi"},
			},
		},
		{
			test: "index_and_name",
			seg:  Child(Name("hi"), Index(2)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{"go"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(2)), Node: "go"},
			},
		},
		{
			test: "slice",
			seg:  Child(Slice(1, 3)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "go"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 42},
				{Path: Normalized(Index(2)), Node: "go"},
			},
		},
		{
			test: "two_slices",
			seg:  Child(Slice(1, 3), Slice(0, 1)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "go", "hi"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 42},
				{Path: Normalized(Index(2)), Node: "go"},
				{Path: Normalized(Index(0)), Node: "hi"},
			},
		},
		{
			test: "overlapping_slices",
			seg:  Child(Slice(1, 3), Slice(0, 3)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "go", "hi", 42, "go"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 42},
				{Path: Normalized(Index(2)), Node: "go"},
				{Path: Normalized(Index(0)), Node: "hi"},
				{Path: Normalized(Index(1)), Node: 42},
				{Path: Normalized(Index(2)), Node: "go"},
			},
		},
		{
			test: "slice_plus_index",
			seg:  Child(Slice(1, 3), Index(0)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "go", "hi"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 42},
				{Path: Normalized(Index(2)), Node: "go"},
				{Path: Normalized(Index(0)), Node: "hi"},
			},
		},
		{
			test: "slice_plus_overlapping_indexes",
			seg:  Child(Slice(1, 3), Index(0), Index(1)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "go", "hi", 42},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 42},
				{Path: Normalized(Index(2)), Node: "go"},
				{Path: Normalized(Index(0)), Node: "hi"},
				{Path: Normalized(Index(1)), Node: 42},
			},
		},
		{
			test: "slice_and_others",
			seg:  Child(Name("hi"), Index(1), Slice()),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "hi", 42, "go", 98.6, "x", true},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 42},
				{Path: Normalized(Index(0)), Node: "hi"},
				{Path: Normalized(Index(1)), Node: 42},
				{Path: Normalized(Index(2)), Node: "go"},
				{Path: Normalized(Index(3)), Node: 98.6},
				{Path: Normalized(Index(4)), Node: "x"},
				{Path: Normalized(Index(5)), Node: true},
			},
		},
		{
			test: "wildcard_array",
			seg:  Child(Wildcard()),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{"hi", 42, "go", 98.6, "x", true},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: "hi"},
				{Path: Normalized(Index(1)), Node: 42},
				{Path: Normalized(Index(2)), Node: "go"},
				{Path: Normalized(Index(3)), Node: 98.6},
				{Path: Normalized(Index(4)), Node: "x"},
				{Path: Normalized(Index(5)), Node: true},
			},
		},
		{
			test: "wildcard_object",
			seg:  Child(Wildcard()),
			src:  map[string]any{"hi": 42, "go": 98.6, "x": true},
			exp:  []any{42, 98.6, true},
			loc: []*LocatedNode{
				{Path: Normalized(Name("hi")), Node: 42},
				{Path: Normalized(Name("go")), Node: 98.6},
				{Path: Normalized(Name("x")), Node: true},
			},
			rand: true,
		},
		{
			test: "wildcard_others_array",
			seg:  Child(Wildcard(), Slice(1, 3), Index(0), Name("go")),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{"hi", 42, "go", 98.6, "x", true, 42, "go", "hi"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: "hi"},
				{Path: Normalized(Index(1)), Node: 42},
				{Path: Normalized(Index(2)), Node: "go"},
				{Path: Normalized(Index(3)), Node: 98.6},
				{Path: Normalized(Index(4)), Node: "x"},
				{Path: Normalized(Index(5)), Node: true},
				{Path: Normalized(Index(1)), Node: 42},
				{Path: Normalized(Index(2)), Node: "go"},
				{Path: Normalized(Index(0)), Node: "hi"},
			},
		},
		{
			test: "wildcard_others_object",
			seg:  Child(Wildcard(), Slice(1, 3), Index(0), Name("go")),
			src:  map[string]any{"hi": 42, "go": 98.6, "x": true},
			exp:  []any{42, 98.6, true, 98.6},
			loc: []*LocatedNode{
				{Path: Normalized(Name("hi")), Node: 42},
				{Path: Normalized(Name("go")), Node: 98.6},
				{Path: Normalized(Name("x")), Node: true},
				{Path: Normalized(Name("go")), Node: 98.6},
			},
			rand: true,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.Equal(tc.seg.selectors, tc.seg.Selectors())
			a.Equal(tc.sing, tc.seg.isSingular())
			a.Equal(tc.seg.descendant, tc.seg.IsDescendant())
			if tc.rand {
				a.ElementsMatch(tc.exp, tc.seg.Select(tc.src, nil))
				a.ElementsMatch(tc.loc, tc.seg.SelectLocated(tc.src, nil, Normalized()))
			} else {
				a.Equal(tc.exp, tc.seg.Select(tc.src, nil))
				a.Equal(tc.loc, tc.seg.SelectLocated(tc.src, nil, Normalized()))
			}
		})
	}
}

func TestDescendantSegmentSelect(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
		seg  *Segment
		src  any
		exp  []any
		loc  []*LocatedNode
		rand bool
	}{
		{
			test: "no_selectors",
			seg:  Descendant(),
			src:  []any{1, 3, []any{3, 5, []any{42, 98.6, true}}},
			exp:  []any{},
			loc:  []*LocatedNode{},
		},
		{
			test: "root_name",
			seg:  Descendant(Name("hi")),
			src:  map[string]any{"hi": 42, "go": map[string]any{"x": 98.6}},
			exp:  []any{42},
			loc: []*LocatedNode{
				{Path: Normalized(Name("hi")), Node: 42},
			},
			rand: true,
		},
		{
			test: "name",
			seg:  Descendant(Name("hi")),
			src: map[string]any{
				"hi": 42, "go": map[string]any{
					"hi": 98.6, "x": map[string]any{"hi": true},
				},
			},
			exp: []any{42, 98.6, true},
			loc: []*LocatedNode{
				{Path: Normalized(Name("hi")), Node: 42},
				{Path: Normalized(Name("go"), Name("hi")), Node: 98.6},
				{Path: Normalized(Name("go"), Name("x"), Name("hi")), Node: true},
			},
			rand: true,
		},
		{
			test: "name_in_name",
			seg:  Descendant(Name("hi")),
			src: map[string]any{
				"hi": 42, "go": map[string]any{
					"hi": map[string]any{"hi": true},
				},
			},
			exp: []any{42, map[string]any{"hi": true}, true},
			loc: []*LocatedNode{
				{Path: Normalized(Name("hi")), Node: 42},
				{Path: Normalized(Name("go"), Name("hi")), Node: map[string]any{"hi": true}},
				{Path: Normalized(Name("go"), Name("hi"), Name("hi")), Node: true},
			},
			rand: true,
		},
		{
			test: "name_under_array",
			seg:  Descendant(Name("hi")),
			src: []any{
				map[string]any{
					"hi": 98.6, "x": map[string]any{"hi": true},
				},
			},
			exp: []any{98.6, true},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Name("hi")), Node: 98.6},
				{Path: Normalized(Index(0), Name("x"), Name("hi")), Node: true},
			},
			rand: true,
		},
		{
			test: "two_names",
			seg:  Descendant(Name("hi"), Name("x")),
			src: map[string]any{
				"hi": 42, "go": map[string]any{
					"hi": 98.6, "go": map[string]any{"hi": true, "x": 12},
				},
				"x": map[string]any{"x": 99},
			},
			exp: []any{42, 98.6, true, 12, map[string]any{"x": 99}, 99},
			loc: []*LocatedNode{
				{Path: Normalized(Name("hi")), Node: 42},
				{Path: Normalized(Name("go"), Name("hi")), Node: 98.6},
				{Path: Normalized(Name("go"), Name("go"), Name("hi")), Node: true},
				{Path: Normalized(Name("go"), Name("go"), Name("x")), Node: 12},
				{Path: Normalized(Name("x")), Node: map[string]any{"x": 99}},
				{Path: Normalized(Name("x"), Name("x")), Node: 99},
			},
			rand: true,
		},
		{
			test: "root_index",
			seg:  Descendant(Index(1)),
			src:  []any{1, 3, []any{3}},
			exp:  []any{3},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 3},
			},
		},
		{
			test: "index",
			seg:  Descendant(Index(1)),
			src:  []any{1, 3, []any{3, 5, []any{42, 98.6, true}}},
			exp:  []any{3, 5, 98.6},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 3},
				{Path: Normalized(Index(2), Index(1)), Node: 5},
				{Path: Normalized(Index(2), Index(2), Index(1)), Node: 98.6},
			},
		},
		{
			test: "two_indexes",
			seg:  Descendant(Index(1), Index(0)),
			src:  []any{1, 3, []any{3, 5, []any{42, 98.6, true}}},
			exp:  []any{3, 1, 5, 3, 98.6, 42},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 3},
				{Path: Normalized(Index(0)), Node: 1},
				{Path: Normalized(Index(2), Index(1)), Node: 5},
				{Path: Normalized(Index(2), Index(0)), Node: 3},
				{Path: Normalized(Index(2), Index(2), Index(1)), Node: 98.6},
				{Path: Normalized(Index(2), Index(2), Index(0)), Node: 42},
			},
		},
		{
			test: "index_under_object",
			seg:  Descendant(Index(1)),
			src:  map[string]any{"x": 1, "y": 3, "z": []any{3, 5, []any{42, 98.6, true}}},
			exp:  []any{5, 98.6},
			loc: []*LocatedNode{
				{Path: Normalized(Name("z"), Index(1)), Node: 5},
				{Path: Normalized(Name("z"), Index(2), Index(1)), Node: 98.6},
			},
		},
		{
			test: "root_slice",
			seg:  Descendant(Slice(1, 3)),
			src:  []any{1, 3, 4, []any{3}},
			exp:  []any{3, 4},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 3},
				{Path: Normalized(Index(2)), Node: 4},
			},
		},
		{
			test: "slice",
			seg:  Descendant(Slice(1, 2)),
			src:  []any{1, 3, 4, []any{3, 5, "x", []any{42, 98.6, "y", "z", true}}},
			exp:  []any{3, 5, 98.6},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 3},
				{Path: Normalized(Index(3), Index(1)), Node: 5},
				{Path: Normalized(Index(3), Index(3), Index(1)), Node: 98.6},
			},
		},
		{
			test: "two_more_slices",
			seg:  Descendant(Slice(1, 2), Slice(3, 4)),
			src:  []any{1, 3, 4, []any{3, 5, "x", []any{42, 98.6, "y", "z", true}}},
			exp: []any{
				3,
				[]any{3, 5, "x", []any{42, 98.6, "y", "z", true}},
				5,
				[]any{42, 98.6, "y", "z", true},
				98.6,
				"z",
			},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 3},
				{Path: Normalized(Index(3)), Node: []any{3, 5, "x", []any{42, 98.6, "y", "z", true}}},
				{Path: Normalized(Index(3), Index(1)), Node: 5},
				{Path: Normalized(Index(3), Index(3)), Node: []any{42, 98.6, "y", "z", true}},
				{Path: Normalized(Index(3), Index(3), Index(1)), Node: 98.6},
				{Path: Normalized(Index(3), Index(3), Index(3)), Node: "z"},
			},
		},
		{
			test: "slice_and_index",
			seg:  Descendant(Slice(1, 2), Index(3)),
			src:  []any{1, 3, 4, []any{3, 5, "x", []any{42, 98.6, "y", "z", true}}},
			exp: []any{
				3,
				[]any{3, 5, "x", []any{42, 98.6, "y", "z", true}},
				5,
				[]any{42, 98.6, "y", "z", true},
				98.6,
				"z",
			},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: 3},
				{Path: Normalized(Index(3)), Node: []any{3, 5, "x", []any{42, 98.6, "y", "z", true}}},
				{Path: Normalized(Index(3), Index(1)), Node: 5},
				{Path: Normalized(Index(3), Index(3)), Node: []any{42, 98.6, "y", "z", true}},
				{Path: Normalized(Index(3), Index(3), Index(1)), Node: 98.6},
				{Path: Normalized(Index(3), Index(3), Index(3)), Node: "z"},
			},
		},
		{
			test: "slice_under_object",
			seg:  Descendant(Slice(1, 2)),
			src:  map[string]any{"x": []any{3, 5, "x", []any{42, 98.6, "y", "z", true}}},
			exp:  []any{5, 98.6},
			loc: []*LocatedNode{
				{Path: Normalized(Name("x"), Index(1)), Node: 5},
				{Path: Normalized(Name("x"), Index(3), Index(1)), Node: 98.6},
			},
		},
		{
			test: "root_wildcard_array",
			seg:  Descendant(Wildcard()),
			src:  []any{1, 3, 4},
			exp:  []any{1, 3, 4},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: 1},
				{Path: Normalized(Index(1)), Node: 3},
				{Path: Normalized(Index(2)), Node: 4},
			},
		},
		{
			test: "root_wildcard_object",
			seg:  Descendant(Wildcard()),
			src:  map[string]any{"x": 42, "y": true},
			exp:  []any{42, true},
			loc: []*LocatedNode{
				{Path: Normalized(Name("x")), Node: 42},
				{Path: Normalized(Name("y")), Node: true},
			},
			rand: true,
		},
		{
			test: "wildcard_nested_array",
			seg:  Descendant(Wildcard()),
			src:  []any{1, 3, []any{4, 5}},
			exp:  []any{1, 3, []any{4, 5}, 4, 5},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: 1},
				{Path: Normalized(Index(1)), Node: 3},
				{Path: Normalized(Index(2)), Node: []any{4, 5}},
				{Path: Normalized(Index(2), Index(0)), Node: 4},
				{Path: Normalized(Index(2), Index(1)), Node: 5},
			},
		},
		{
			test: "wildcard_nested_object",
			seg:  Descendant(Wildcard()),
			src:  map[string]any{"x": 42, "y": map[string]any{"z": "hi"}},
			exp:  []any{42, map[string]any{"z": "hi"}, "hi"},
			loc: []*LocatedNode{
				{Path: Normalized(Name("x")), Node: 42},
				{Path: Normalized(Name("y")), Node: map[string]any{"z": "hi"}},
				{Path: Normalized(Name("y"), Name("z")), Node: "hi"},
			},
			rand: true,
		},
		{
			test: "wildcard_mixed",
			seg:  Descendant(Wildcard()),
			src:  []any{1, 3, map[string]any{"z": "hi"}},
			exp:  []any{1, 3, map[string]any{"z": "hi"}, "hi"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: 1},
				{Path: Normalized(Index(1)), Node: 3},
				{Path: Normalized(Index(2)), Node: map[string]any{"z": "hi"}},
				{Path: Normalized(Index(2), Name("z")), Node: "hi"},
			},
			rand: true,
		},
		{
			test: "wildcard_mixed_index",
			seg:  Descendant(Wildcard(), Index(0)),
			src:  []any{1, 3, map[string]any{"z": "hi"}},
			exp:  []any{1, 3, map[string]any{"z": "hi"}, 1, "hi"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: 1},
				{Path: Normalized(Index(1)), Node: 3},
				{Path: Normalized(Index(2)), Node: map[string]any{"z": "hi"}},
				{Path: Normalized(Index(0)), Node: 1},
				{Path: Normalized(Index(2), Name("z")), Node: "hi"},
			},
		},
		{
			test: "wildcard_mixed_name",
			seg:  Descendant(Wildcard(), Name("z")),
			src:  []any{1, 3, map[string]any{"z": "hi", "y": "x"}},
			exp:  []any{1, 3, map[string]any{"z": "hi", "y": "x"}, "hi", "x", "hi"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: 1},
				{Path: Normalized(Index(1)), Node: 3},
				{Path: Normalized(Index(2)), Node: map[string]any{"z": "hi", "y": "x"}},
				{Path: Normalized(Index(2), Name("z")), Node: "hi"},
				{Path: Normalized(Index(2), Name("y")), Node: "x"},
				{Path: Normalized(Index(2), Name("z")), Node: "hi"},
			},
			rand: true,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.False(tc.seg.isSingular())
			a.True(tc.seg.IsDescendant())
			if tc.rand {
				a.ElementsMatch(tc.exp, tc.seg.Select(tc.src, nil))
				a.ElementsMatch(tc.loc, tc.seg.SelectLocated(tc.src, nil, Normalized()))
			} else {
				a.Equal(tc.exp, tc.seg.Select(tc.src, nil))
				a.Equal(tc.loc, tc.seg.SelectLocated(tc.src, nil, Normalized()))
			}
		})
	}
}
