package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSegmentString(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		seg  *Segment
		str  string
		sing bool
	}{
		{
			name: "no_selectors",
			seg:  Child(),
			str:  "[]",
		},
		{
			name: "descendant_no_selectors",
			seg:  Descendant(),
			str:  "..[]",
		},
		{
			name: "name",
			seg:  Child(Name("hi")),
			str:  `["hi"]`,
			sing: true,
		},
		{
			name: "index",
			seg:  Child(Index(2)),
			str:  `[2]`,
			sing: true,
		},
		{
			name: "wildcard",
			seg:  Child(Wildcard),
			str:  `[*]`,
		},
		{
			name: "slice",
			seg:  Child(Slice(2)),
			str:  `[2:]`,
		},
		{
			name: "multiples",
			seg:  Child(Slice(2), Name("hi"), Index(3)),
			str:  `[2:,"hi",3]`,
		},
		{
			name: "descendant_multiples",
			seg:  Descendant(Slice(2), Name("hi"), Index(3)),
			str:  `..[2:,"hi",3]`,
		},
		{
			name: "wildcard_override",
			seg:  Child(Slice(2), Name("hi"), Index(3), Wildcard),
			str:  `[2:,"hi",3,*]`,
		},
		{
			name: "descendant_wildcard_override",
			seg:  Descendant(Slice(2), Name("hi"), Index(3), Wildcard),
			str:  `..[2:,"hi",3,*]`,
		},
		{
			name: "dupes",
			seg:  Child(Slice(2), Name("hi"), Slice(2), Slice(2), Name("hi"), Index(3), Name("go"), Index(3)),
			str:  `[2:,"hi",2:,2:,"hi",3,"go",3]`,
		},
		{
			name: "descendant_dupes",
			seg:  Descendant(Slice(2), Name("hi"), Slice(2), Slice(2), Name("hi"), Index(3), Name("go"), Index(3)),
			str:  `..[2:,"hi",2:,2:,"hi",3,"go",3]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.str, tc.seg.String())
			a.Equal(tc.sing, tc.seg.isSingular())
		})
	}
}

func TestSegmentQuery(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		seg  *Segment
		src  any
		exp  []any
		rand bool
		sing bool
	}{
		{
			name: "no_selectors",
			seg:  Child(),
			src:  []any{1, 3},
			exp:  []any{},
		},
		{
			name: "name",
			seg:  Child(Name("hi")),
			src:  map[string]any{"hi": 42, "go": 98.6, "x": true},
			exp:  []any{42},
			sing: true,
		},
		{
			name: "two_names",
			seg:  Child(Name("hi"), Name("go")),
			src:  map[string]any{"hi": 42, "go": 98.6, "x": true},
			exp:  []any{42, 98.6},
			rand: true,
		},
		{
			name: "dupe_name",
			seg:  Child(Name("hi"), Name("hi")),
			src:  map[string]any{"hi": 42, "go": 98.6, "x": true},
			exp:  []any{42, 42},
			rand: true,
		},
		{
			name: "three_names",
			seg:  Child(Name("hi"), Name("go"), Name("x")),
			src:  map[string]any{"hi": 42, "go": 98.6, "x": true},
			rand: true,
			exp:  []any{42, 98.6, true},
		},
		{
			name: "name_and_others",
			seg:  Child(Name("hi"), Index(1), Slice()),
			src:  map[string]any{"hi": 42, "go": 98.6, "x": true},
			exp:  []any{42},
		},
		{
			name: "index",
			seg:  Child(Index(1)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42},
			sing: true,
		},
		{
			name: "two_indexes",
			seg:  Child(Index(1), Index(4)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "x"},
		},
		{
			name: "dupe_index",
			seg:  Child(Index(1), Index(1)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, 42},
		},
		{
			name: "three_indexes",
			seg:  Child(Index(1), Index(4), Index(0)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "x", "hi"},
		},
		{
			name: "index_and_name",
			seg:  Child(Name("hi"), Index(2)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{"go"},
		},
		{
			name: "slice",
			seg:  Child(Slice(1, 3)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "go"},
		},
		{
			name: "two_slices",
			seg:  Child(Slice(1, 3), Slice(0, 1)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "go", "hi"},
		},
		{
			name: "overlapping_slices",
			seg:  Child(Slice(1, 3), Slice(0, 3)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "go", "hi", 42, "go"},
		},
		{
			name: "slice_plus_index",
			seg:  Child(Slice(1, 3), Index(0)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "go", "hi"},
		},
		{
			name: "slice_plus_overlapping_indexes",
			seg:  Child(Slice(1, 3), Index(0), Index(1)),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "go", "hi", 42},
		},
		{
			name: "slice_and_others",
			seg:  Child(Name("hi"), Index(1), Slice()),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{42, "hi", 42, "go", 98.6, "x", true},
		},
		{
			name: "wildcard_array",
			seg:  Child(Wildcard),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{"hi", 42, "go", 98.6, "x", true},
		},
		{
			name: "wildcard_object",
			seg:  Child(Wildcard),
			src:  map[string]any{"hi": 42, "go": 98.6, "x": true},
			exp:  []any{42, 98.6, true},
			rand: true,
		},
		{
			name: "wildcard_others_array",
			seg:  Child(Wildcard, Slice(1, 3), Index(0), Name("go")),
			src:  []any{"hi", 42, "go", 98.6, "x", true},
			exp:  []any{"hi", 42, "go", 98.6, "x", true, 42, "go", "hi"},
		},
		{
			name: "wildcard_others_object",
			seg:  Child(Wildcard, Slice(1, 3), Index(0), Name("go")),
			src:  map[string]any{"hi": 42, "go": 98.6, "x": true},
			exp:  []any{42, 98.6, true, 98.6},
			rand: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.seg.selectors, tc.seg.Selectors())
			a.Equal(tc.sing, tc.seg.isSingular())
			if tc.rand {
				a.ElementsMatch(tc.exp, tc.seg.Select(tc.src, nil))
			} else {
				a.Equal(tc.exp, tc.seg.Select(tc.src, nil))
			}
		})
	}
}

func TestDescendantSegmentQuery(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		seg  *Segment
		src  any
		exp  []any
		rand bool
	}{
		{
			name: "no_selectors",
			seg:  Descendant(),
			src:  []any{1, 3, []any{3, 5, []any{42, 98.6, true}}},
			exp:  []any{},
		},
		{
			name: "root_name",
			seg:  Descendant(Name("hi")),
			src:  map[string]any{"hi": 42, "go": map[string]any{"x": 98.6}},
			exp:  []any{42},
			rand: true,
		},
		{
			name: "name",
			seg:  Descendant(Name("hi")),
			src: map[string]any{
				"hi": 42, "go": map[string]any{
					"hi": 98.6, "x": map[string]any{"hi": true},
				},
			},
			exp:  []any{42, 98.6, true},
			rand: true,
		},
		{
			name: "name_in_name",
			seg:  Descendant(Name("hi")),
			src: map[string]any{
				"hi": 42, "go": map[string]any{
					"hi": map[string]any{"hi": true},
				},
			},
			exp:  []any{42, map[string]any{"hi": true}, true},
			rand: true,
		},
		{
			name: "name_under_array",
			seg:  Descendant(Name("hi")),
			src: []any{
				map[string]any{
					"hi": 98.6, "x": map[string]any{"hi": true},
				},
			},
			exp:  []any{98.6, true},
			rand: true,
		},
		{
			name: "two_names",
			seg:  Descendant(Name("hi"), Name("x")),
			src: map[string]any{
				"hi": 42, "go": map[string]any{
					"hi": 98.6, "go": map[string]any{"hi": true, "x": 12},
				},
				"x": map[string]any{"x": 99},
			},
			exp:  []any{42, 98.6, true, 12, map[string]any{"x": 99}, 99},
			rand: true,
		},
		{
			name: "root_index",
			seg:  Descendant(Index(1)),
			src:  []any{1, 3, []any{3}},
			exp:  []any{3},
		},
		{
			name: "index",
			seg:  Descendant(Index(1)),
			src:  []any{1, 3, []any{3, 5, []any{42, 98.6, true}}},
			exp:  []any{3, 5, 98.6},
		},
		{
			name: "two_indexes",
			seg:  Descendant(Index(1), Index(0)),
			src:  []any{1, 3, []any{3, 5, []any{42, 98.6, true}}},
			exp:  []any{3, 1, 5, 3, 98.6, 42},
		},
		{
			name: "index_under_object",
			seg:  Descendant(Index(1)),
			src:  map[string]any{"x": 1, "y": 3, "z": []any{3, 5, []any{42, 98.6, true}}},
			exp:  []any{5, 98.6},
		},
		{
			name: "root_slice",
			seg:  Descendant(Slice(1, 3)),
			src:  []any{1, 3, 4, []any{3}},
			exp:  []any{3, 4},
		},
		{
			name: "slice",
			seg:  Descendant(Slice(1, 2)),
			src:  []any{1, 3, 4, []any{3, 5, "x", []any{42, 98.6, "y", "z", true}}},
			exp:  []any{3, 5, 98.6},
		},
		{
			name: "two_slices",
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
		},
		{
			name: "slice_and_index",
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
		},
		{
			name: "slice_under_object",
			seg:  Descendant(Slice(1, 2)),
			src:  map[string]any{"x": []any{3, 5, "x", []any{42, 98.6, "y", "z", true}}},
			exp:  []any{5, 98.6},
		},
		{
			name: "root_wildcard_array",
			seg:  Descendant(Wildcard),
			src:  []any{1, 3, 4},
			exp:  []any{1, 3, 4},
		},
		{
			name: "root_wildcard_object",
			seg:  Descendant(Wildcard),
			src:  map[string]any{"x": 42, "y": true},
			exp:  []any{42, true},
			rand: true,
		},
		{
			name: "wildcard_nested_array",
			seg:  Descendant(Wildcard),
			src:  []any{1, 3, []any{4, 5}},
			exp:  []any{1, 3, []any{4, 5}, 4, 5},
		},
		{
			name: "wildcard_nested_object",
			seg:  Descendant(Wildcard),
			src:  map[string]any{"x": 42, "y": map[string]any{"z": "hi"}},
			exp:  []any{42, map[string]any{"z": "hi"}, "hi"},
			rand: true,
		},
		{
			name: "wildcard_mixed",
			seg:  Descendant(Wildcard),
			src:  []any{1, 3, map[string]any{"z": "hi"}},
			exp:  []any{1, 3, map[string]any{"z": "hi"}, "hi"},
			rand: true,
		},
		{
			name: "wildcard_mixed_index",
			seg:  Descendant(Wildcard, Index(0)),
			src:  []any{1, 3, map[string]any{"z": "hi"}},
			exp:  []any{1, 3, map[string]any{"z": "hi"}, 1, "hi"},
		},
		{
			name: "wildcard_mixed_name",
			seg:  Descendant(Wildcard, Name("z")),
			src:  []any{1, 3, map[string]any{"z": "hi", "y": "x"}},
			exp:  []any{1, 3, map[string]any{"z": "hi", "y": "x"}, "hi", "x", "hi"},
			rand: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.False(tc.seg.isSingular())
			if tc.rand {
				a.ElementsMatch(tc.exp, tc.seg.Select(tc.src, nil))
			} else {
				a.Equal(tc.exp, tc.seg.Select(tc.src, nil))
			}
		})
	}
}
