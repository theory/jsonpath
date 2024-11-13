package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryRoot(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		val  any
	}{
		{"string", "Hi there 😀"},
		{"bool", true},
		{"float", 98.6},
		{"float64", float64(98.6)},
		{"float32", float32(98.6)},
		{"int", 42},
		{"int64", int64(42)},
		{"int32", int32(42)},
		{"int16", int16(42)},
		{"int8", int8(42)},
		{"uint64", uint64(42)},
		{"uint32", uint32(42)},
		{"uint16", uint16(42)},
		{"uint8", uint8(42)},
		{"struct", struct{ x int }{}},
		{"nil", nil},
		{"map", map[string]any{"x": true, "y": []any{1, 2}}},
		{"slice", []any{1, 2, map[string]any{"x": true}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			q := Query(false, nil)
			a.Equal([]any{tc.val}, q.Select(tc.val, nil))
		})
	}
}

func TestQueryString(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		segs []*Segment
		str  string
	}{
		{
			name: "empty",
			segs: []*Segment{},
			str:  "",
		},
		{
			name: "one_key",
			segs: []*Segment{Child(Name("x"))},
			str:  `["x"]`,
		},
		{
			name: "two_keys",
			segs: []*Segment{Child(Name("x"), Name("y"))},
			str:  `["x","y"]`,
		},
		{
			name: "two_segs",
			segs: []*Segment{Child(Name("x"), Name("y")), Child(Index(0))},
			str:  `["x","y"][0]`,
		},
		{
			name: "segs_plus_descendant",
			segs: []*Segment{Child(Name("x"), Name("y")), Child(Wildcard), Descendant(Index(0))},
			str:  `["x","y"][*]..[0]`,
		},
		{
			name: "segs_with_slice",
			segs: []*Segment{Child(Name("x"), Slice(2)), Child(Wildcard), Descendant(Index(0))},
			str:  `["x",2:][*]..[0]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			q := Query(false, tc.segs)
			a.Equal("@"+tc.str, q.String())
			q = Query(true, tc.segs)
			a.Equal("$"+tc.str, q.String())
		})
	}
}

type queryTestCase struct {
	name  string
	segs  []*Segment
	input any
	exp   []any
	rand  bool
}

func (tc queryTestCase) run(a *assert.Assertions) {
	// Set up Query.
	q := Query(false, tc.segs)
	a.Equal(tc.segs, q.Segments())
	a.False(q.root)

	// Test both.
	if tc.rand {
		a.ElementsMatch(tc.exp, q.Select(tc.input, nil))
	} else {
		a.Equal(tc.exp, q.Select(tc.input, nil))
	}
}

func TestQueryObject(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []queryTestCase{
		{
			name:  "root",
			input: map[string]any{"x": true, "y": []any{1, 2}},
			exp:   []any{map[string]any{"x": true, "y": []any{1, 2}}},
		},
		{
			name:  "one_key_scalar",
			segs:  []*Segment{Child(Name("x"))},
			input: map[string]any{"x": true, "y": []any{1, 2}},
			exp:   []any{true},
		},
		{
			name:  "one_key_array",
			segs:  []*Segment{Child(Name("y"))},
			input: map[string]any{"x": true, "y": []any{1, 2}},
			exp:   []any{[]any{1, 2}},
		},
		{
			name:  "one_key_object",
			segs:  []*Segment{Child(Name("y"))},
			input: map[string]any{"x": true, "y": map[string]any{"a": 1}},
			exp:   []any{map[string]any{"a": 1}},
		},
		{
			name:  "multiple_keys",
			segs:  []*Segment{Child(Name("x"), Name("y"))},
			input: map[string]any{"x": true, "y": []any{1, 2}, "z": "hi"},
			exp:   []any{true, []any{1, 2}},
			rand:  true,
		},
		{
			name: "three_level_path",
			segs: []*Segment{
				Child(Name("x")),
				Child(Name("a")),
				Child(Name("i")),
			},
			input: map[string]any{
				"x": map[string]any{
					"a": map[string]any{
						"i": []any{1, 2},
						"j": 42,
					},
					"b": "no",
				},
				"y": 1,
			},
			exp: []any{[]any{1, 2}},
		},
		{
			name: "wildcard_keys",
			segs: []*Segment{
				Child(Wildcard),
				Child(Name("a"), Name("b")),
			},
			input: map[string]any{
				"x": map[string]any{"a": "go", "b": 2, "c": 5},
				"y": map[string]any{"a": 2, "b": 3, "d": 3},
			},
			exp:  []any{"go", 2, 2, 3},
			rand: true,
		},
		{
			name: "any_key_indexes",
			segs: []*Segment{
				Child(Wildcard),
				Child(Index(0), Index(1)),
			},
			input: map[string]any{
				"x": []any{"a", "go", "b", 2, "c", 5},
				"y": []any{"a", 2, "b", 3, "d", 3},
			},
			exp:  []any{"a", "go", "a", 2},
			rand: true,
		},
		{
			name: "any_key_nonexistent_index",
			segs: []*Segment{Child(Wildcard), Child(Index(1))},
			input: map[string]any{
				"x": []any{"a", "go", "b", 2, "c", 5},
				"y": []any{"a"},
			},
			exp: []any{"go"},
		},
		{
			name:  "nonexistent_key",
			segs:  []*Segment{Child(Name("x"))},
			input: map[string]any{"y": []any{1, 2}},
			exp:   []any{},
		},
		{
			name:  "nonexistent_branch_key",
			segs:  []*Segment{Child(Name("x")), Child(Name("z"))},
			input: map[string]any{"y": []any{1, 2}},
			exp:   []any{},
		},
		{
			name:  "wildcard_then_nonexistent_key",
			segs:  []*Segment{Child(Wildcard), Child(Name("x"))},
			input: map[string]any{"y": map[string]any{"a": 1}},
			exp:   []any{},
		},
		{
			name:  "not_an_object",
			segs:  []*Segment{Child(Name("x")), Child(Name("y"))},
			input: map[string]any{"x": true},
			exp:   []any{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(a)
		})
	}
}

func TestQueryArray(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []queryTestCase{
		{
			name:  "root",
			input: []any{"x", true, "y", []any{1, 2}},
			exp:   []any{[]any{"x", true, "y", []any{1, 2}}},
		},
		{
			name:  "index_zero",
			segs:  []*Segment{Child(Index(0))},
			input: []any{"x", true, "y", []any{1, 2}},
			exp:   []any{"x"},
		},
		{
			name:  "index_one",
			segs:  []*Segment{Child(Index(1))},
			input: []any{"x", true, "y", []any{1, 2}},
			exp:   []any{true},
		},
		{
			name:  "index_three",
			segs:  []*Segment{Child(Index(3))},
			input: []any{"x", true, "y", []any{1, 2}},
			exp:   []any{[]any{1, 2}},
		},
		{
			name:  "multiple_indexes",
			segs:  []*Segment{Child(Index(1), Index(3))},
			input: []any{"x", true, "y", []any{1, 2}},
			exp:   []any{true, []any{1, 2}},
		},
		{
			name:  "nested_indices",
			segs:  []*Segment{Child(Index(0)), Child(Index(0))},
			input: []any{[]any{1, 2}, "x", true, "y"},
			exp:   []any{1},
		},
		{
			name:  "nested_multiple_indices",
			segs:  []*Segment{Child(Index(0)), Child(Index(0), Index(1))},
			input: []any{[]any{1, 2, 3}, "x", true, "y"},
			exp:   []any{1, 2},
		},
		{
			name:  "nested_index_gaps",
			segs:  []*Segment{Child(Index(1)), Child(Index(1))},
			input: []any{"x", []any{1, 2}, true, "y"},
			exp:   []any{2},
		},
		{
			name: "three_level_index_path",
			segs: []*Segment{
				Child(Index(0)),
				Child(Index(0)),
				Child(Index(0)),
			},
			input: []any{[]any{[]any{42, 12}, 2}, "x", true, "y"},
			exp:   []any{42},
		},
		{
			name: "mixed_nesting",
			segs: []*Segment{
				Child(Index(0), Index(1), Index(3)),
				Child(Index(1), Name("y"), Name("z")),
			},
			input: []any{
				[]any{[]any{42, 12}, 2},
				"x",
				true,
				map[string]any{"y": "hi", "z": 1, "x": "no"},
			},
			exp:  []any{2, "hi", 1},
			rand: true,
		},
		{
			name:  "wildcard_indexes_index",
			segs:  []*Segment{Child(Wildcard), Child(Index(0), Index(2))},
			input: []any{[]any{1, 2, 3}, []any{3, 2, 1}, []any{4, 5, 6}},
			exp:   []any{1, 3, 3, 1, 4, 6},
		},
		{
			name:  "nonexistent_index",
			segs:  []*Segment{Child(Index(3))},
			input: []any{"y", []any{1, 2}},
			exp:   []any{},
		},
		{
			name:  "nonexistent_child_index",
			segs:  []*Segment{Child(Wildcard), Child(Index(3))},
			input: []any{[]any{0, 1, 2, 3}, []any{0, 1, 2}},
			exp:   []any{3},
		},
		{
			name:  "not_an_array_index_1",
			segs:  []*Segment{Child(Index(1)), Child(Index(0))},
			input: []any{"x", true},
			exp:   []any{},
		},
		{
			name:  "not_an_array_index_0",
			segs:  []*Segment{Child(Index(0)), Child(Index(0))},
			input: []any{"x", true},
			exp:   []any{},
		},
		{
			name:  "wildcard_not_an_array_index_1",
			segs:  []*Segment{Child(Wildcard), Child(Index(0))},
			input: []any{"x", true},
			exp:   []any{},
		},
		{
			name: "mix_wildcard_keys",
			segs: []*Segment{
				Child(Wildcard, Index(1)),
				Child(Name("x"), Index(1), Name("y")),
			},
			input: []any{
				map[string]any{"x": "hi", "y": "go"},
				map[string]any{"x": "bo", "y": 42},
				map[string]any{"x": true, "y": 21},
				[]any{34, 53, 23},
			},
			exp:  []any{"hi", "go", "bo", 42, true, 21, 53, "bo", 42},
			rand: true,
		},
		{
			name: "mix_wildcard_nonexistent_key",
			segs: []*Segment{
				Child(Wildcard, Index(1)),
				Child(Name("x"), Name("y")),
			},
			input: []any{
				map[string]any{"x": "hi"},
				map[string]any{"x": "bo"},
				map[string]any{"x": true},
			},
			exp: []any{"hi", "bo", true, "bo"},
		},
		{
			name: "mix_wildcard_index",
			segs: []*Segment{
				Child(Wildcard, Index(1)),
				Child(Index(0), Index(1)),
			},
			input: []any{
				[]any{"x", "hi", true},
				[]any{"x", "bo", 42},
				[]any{"x", true, 21},
			},
			exp: []any{"x", "hi", "x", "bo", "x", true, "x", "bo"},
		},
		{
			name: "mix_wildcard_nonexistent_index",
			segs: []*Segment{
				Child(Wildcard, Index(1)),
				Child(Index(0), Index(3)),
			},
			input: []any{
				[]any{"x", "hi", true},
				[]any{"x", "bo", 42},
				[]any{"x", true, 21},
			},
			exp: []any{"x", "x", "x", "x"},
		},
		{
			name: "wildcard_nonexistent_key",
			segs: []*Segment{Child(Wildcard), Child(Name("a"))},
			input: []any{
				map[string]any{"a": 1, "b": 2},
				map[string]any{"z": 3, "b": 4},
			},
			exp: []any{1},
		},
		{
			name: "wildcard_nonexistent_middle_key",
			segs: []*Segment{Child(Wildcard), Child(Name("a"))},
			input: []any{
				map[string]any{"a": 1, "b": 2},
				map[string]any{"z": 3, "b": 4},
				map[string]any{"a": 5},
				map[string]any{"z": 3, "b": 4},
			},
			exp: []any{1, 5},
		},
		{
			name: "wildcard_nested_nonexistent_key",
			segs: []*Segment{
				Child(Wildcard),
				Child(Wildcard),
				Child(Name("a")),
			},
			input: []any{
				map[string]any{
					"x": map[string]any{"a": 1},
					"y": map[string]any{"b": 1},
				},
				map[string]any{
					"y": map[string]any{"b": 1},
				},
			},
			exp: []any{1},
		},
		{
			name: "wildcard_nested_nonexistent_index",
			segs: []*Segment{
				Child(Wildcard),
				Child(Wildcard),
				Child(Index(1)),
			},
			input: []any{
				map[string]any{
					"x": []any{1, 2},
					"y": []any{3},
				},
				map[string]any{
					"z": []any{1},
				},
			},
			exp: []any{2},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(a)
		})
	}
}

func TestQuerySlice(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []queryTestCase{
		{
			name:  "slice_0_2",
			segs:  []*Segment{Child(Slice(0, 2))},
			input: []any{"x", true, "y", []any{1, 2}},
			exp:   []any{"x", true},
		},
		{
			name:  "slice_0_1",
			segs:  []*Segment{Child(Slice(0, 1))},
			input: []any{"x", true, "y", []any{1, 2}},
			exp:   []any{"x"},
		},
		{
			name:  "slice_2_5",
			segs:  []*Segment{Child(Slice(2, 5))},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"y", []any{1, 2}, 42},
		},
		{
			name:  "slice_2_5_over_len",
			segs:  []*Segment{Child(Slice(2, 5))},
			input: []any{"x", true, "y"},
			exp:   []any{"y"},
		},
		{
			name:  "slice_defaults",
			segs:  []*Segment{Child(Slice())},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
		},
		{
			name:  "default_start",
			segs:  []*Segment{Child(Slice(nil, 2))},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"x", true},
		},
		{
			name:  "default_end",
			segs:  []*Segment{Child(Slice(2))},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"y", []any{1, 2}, 42, nil, 78},
		},
		{
			name:  "step_2",
			segs:  []*Segment{Child(Slice(nil, nil, 2))},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"x", "y", 42, 78},
		},
		{
			name:  "step_3",
			segs:  []*Segment{Child(Slice(nil, nil, 3))},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"x", []any{1, 2}, 78},
		},
		{
			name:  "multiple_slices",
			segs:  []*Segment{Child(Slice(0, 1), Slice(3, 4))},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"x", []any{1, 2}},
		},
		{
			name:  "overlapping_slices",
			segs:  []*Segment{Child(Slice(0, 3), Slice(2, 4))},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"x", true, "y", "y", []any{1, 2}},
		},
		{
			name: "nested_slices",
			segs: []*Segment{Child(Slice(0, 2)), Child(Slice(1, 2))},
			input: []any{
				[]any{"hi", 42, true},
				[]any{"go", "on"},
				[]any{"yo", 98.6, false},
				"x", true, "y",
			},
			exp: []any{42, "on"},
		},
		{
			name: "nested_multiple_indices",
			segs: []*Segment{
				Child(Slice(0, 2)),
				Child(Slice(1, 2), Slice(3, 5)),
			},
			input: []any{
				[]any{"hi", 42, true, 64, []any{}, 7},
				[]any{"go", "on", false, 88, []any{1}, 8},
				[]any{"yo", 98.6, false, 2, []any{3, 4}, 9},
				"x", true, "y",
			},
			exp: []any{42, 64, []any{}, "on", 88, []any{1}},
		},
		{
			name: "three_level_slice_path",
			segs: []*Segment{
				Child(Slice(0, 2)),
				Child(Slice(0, 1)),
				Child(Slice(0, 1)),
			},
			input: []any{
				[]any{[]any{42, 12}, 2},
				[]any{[]any{16, true, "x"}, 7},
				"x", true, "y",
			},
			exp: []any{42, 16},
		},
		{
			name: "varying_nesting_levels_mixed",
			segs: []*Segment{
				Child(Slice(0, 2), Slice(2, 3)),
				Child(Slice(0, 1), Slice(3, 4)),
				Child(Slice(0, 1), Name("y"), Name("z")),
			},
			input: []any{
				[]any{[]any{42, 12}, 2},
				"x",
				[]any{map[string]any{"y": "hi", "z": 1, "x": "no"}},
				true,
				"go",
			},
			exp: []any{42, "hi", 1},
		},
		{
			name: "wildcard_slices_index",
			segs: []*Segment{
				Child(Wildcard),
				Child(Slice(0, 2), Slice(3, 4)),
			},
			input: []any{
				[]any{1, 2, 3, 4, 5},
				[]any{3, 2, 1, 0, -1},
				[]any{4, 5, 6, 7, 8},
			},
			exp: []any{1, 2, 4, 3, 2, 0, 4, 5, 7},
		},
		{
			name:  "nonexistent_slice",
			segs:  []*Segment{Child(Slice(3, 5))},
			input: []any{"y", []any{1, 2}},
			exp:   []any{},
		},
		{
			name:  "nonexistent_branch_index",
			segs:  []*Segment{Child(Wildcard), Child(Slice(3, 5))},
			input: []any{[]any{0, 1, 2, 3, 4}, []any{0, 1, 2}},
			exp:   []any{3, 4},
		},
		{
			name:  "not_an_array_index_1",
			segs:  []*Segment{Child(Index(1)), Child(Index(0))},
			input: []any{"x", true},
			exp:   []any{},
		},
		{
			name:  "not_an_array",
			segs:  []*Segment{Child(Slice(0, 5)), Child(Index(0))},
			input: []any{"x", true},
			exp:   []any{},
		},
		{
			name:  "wildcard_not_an_array_index_1",
			segs:  []*Segment{Child(Wildcard), Child(Slice(0, 5))},
			input: []any{"x", true},
			exp:   []any{},
		},
		{
			name: "mix_slice_keys",
			segs: []*Segment{
				Child(Slice(0, 5), Index(1)),
				Child(Name("x"), Name("y")),
			},
			input: []any{
				map[string]any{"x": "hi", "y": "go"},
				map[string]any{"x": "bo", "y": 42},
				map[string]any{"x": true, "y": 21},
			},
			exp:  []any{"hi", "go", "bo", 42, true, 21, "bo", 42},
			rand: true,
		},
		{
			name: "mix_slice_nonexistent_key",
			segs: []*Segment{
				Child(Slice(0, 5), Index(1)),
				Child(Name("x"), Name("y")),
			},
			input: []any{
				map[string]any{"x": "hi"},
				map[string]any{"x": "bo"},
				map[string]any{"x": true},
			},
			exp: []any{"hi", "bo", true, "bo"},
		},
		{
			name: "mix_slice_index",
			segs: []*Segment{
				Child(Slice(0, 5), Index(1)),
				Child(Index(0), Index(1)),
			},
			input: []any{
				[]any{"x", "hi", true},
				[]any{"y", "bo", 42},
				[]any{"z", true, 21},
			},
			exp: []any{"x", "hi", "y", "bo", "z", true, "y", "bo"},
		},
		{
			name: "mix_slice_nonexistent_index",
			segs: []*Segment{
				Child(Slice(0, 5), Index(1)),
				Child(Index(0), Index(3)),
			},
			input: []any{
				[]any{"x", "hi", true},
				[]any{"y", "bo", 42},
				[]any{"z", true, 21},
			},
			exp: []any{"x", "y", "z", "y"},
		},
		{
			name: "slice_nonexistent_key",
			segs: []*Segment{Child(Slice(0, 5)), Child(Name("a"))},
			input: []any{
				map[string]any{"a": 1, "b": 2},
				map[string]any{"z": 3, "b": 4},
			},
			exp: []any{1},
		},
		{
			name: "slice_nonexistent_middle_key",
			segs: []*Segment{Child(Slice(0, 5)), Child(Name("a"))},
			input: []any{
				map[string]any{"a": 1, "b": 2},
				map[string]any{"z": 3, "b": 4},
				map[string]any{"a": 5},
				map[string]any{"z": 3, "b": 4},
			},
			exp: []any{1, 5},
		},
		{
			name: "slice_nested_nonexistent_key",
			segs: []*Segment{
				Child(Slice(0, 5)),
				Child(Wildcard),
				Child(Name("a")),
			},
			input: []any{
				map[string]any{
					"x": map[string]any{"a": 1},
					"y": map[string]any{"b": 1},
				},
				map[string]any{
					"y": map[string]any{"b": 1},
				},
			},
			exp: []any{1},
		},
		{
			name: "slice_nested_nonexistent_index",
			segs: []*Segment{
				Child(Slice(0, 5)),
				Child(Wildcard),
				Child(Index(1)),
			},
			input: []any{
				map[string]any{
					"x": []any{1, 2},
					"y": []any{3},
				},
				map[string]any{
					"z": []any{1},
				},
			},
			exp: []any{2},
		},
		{
			name:  "slice_neg",
			segs:  []*Segment{Child(Slice(nil, nil, -1))},
			input: []any{"x", true, "y", []any{1, 2}},
			exp:   []any{[]any{1, 2}, "y", true, "x"},
		},
		{
			name:  "slice_5_0_neg2",
			segs:  []*Segment{Child(Slice(5, 0, -2))},
			input: []any{"x", true, "y", 8, 13, 25, 23, 78, 13},
			exp:   []any{25, 8, true},
		},
		{
			name: "nested_neg_slices",
			segs: []*Segment{
				Child(Slice(2, nil, -1)),
				Child(Slice(2, 0, -1)),
			},
			input: []any{
				[]any{"hi", 42, true},
				[]any{"go", "on"},
				[]any{"yo", 98.6, false},
				"x", true, "y",
			},
			exp: []any{false, 98.6, "on", true, 42},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(a)
		})
	}
}

func TestQueryDescendants(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	json := map[string]any{
		"o": map[string]any{"j": 1, "k": 2},
		"a": []any{5, 3, []any{map[string]any{"j": 4}, map[string]any{"k": 6}}},
	}

	for _, tc := range []queryTestCase{
		{
			name:  "descendant_name",
			segs:  []*Segment{Descendant(Name("j"))},
			input: json,
			exp:   []any{1, 4},
			rand:  true,
		},
		{
			name:  "un_descendant_name",
			segs:  []*Segment{Descendant(Name("o"))},
			input: json,
			exp:   []any{map[string]any{"j": 1, "k": 2}},
		},
		{
			name:  "nested_name",
			segs:  []*Segment{Child(Name("o")), Descendant(Name("k"))},
			input: json,
			exp:   []any{2},
		},
		{
			name:  "nested_wildcard",
			segs:  []*Segment{Child(Name("o")), Descendant(Wildcard)},
			input: json,
			exp:   []any{1, 2},
			rand:  true,
		},
		{
			name:  "single_index",
			segs:  []*Segment{Descendant(Index(0))},
			input: json,
			exp:   []any{5, map[string]any{"j": 4}},
		},
		{
			name:  "nested_index",
			segs:  []*Segment{Child(Name("a")), Descendant(Index(0))},
			input: json,
			exp:   []any{5, map[string]any{"j": 4}},
		},
		{
			name: "multiples",
			segs: []*Segment{
				Child(Name("profile")),
				Descendant(Name("last"), Name("primary"), Name("secondary")),
			},
			input: map[string]any{
				"profile": map[string]any{
					"name": map[string]any{
						"first": "Barrack",
						"last":  "Obama",
					},
					"contacts": map[string]any{
						"email": map[string]any{
							"primary":   "foo@example.com",
							"secondary": "2nd@example.net",
						},
						"phones": map[string]any{
							"primary":   "123456789",
							"secondary": "987654321",
							"fax":       "1029384758",
						},
						"addresses": map[string]any{
							"primary": []any{
								"123 Main Street",
								"Whatever", "OR", "98754",
							},
							"work": []any{
								"whatever",
								"XYZ", "NY", "10093",
							},
						},
					},
				},
			},
			exp: []any{
				"Obama",
				"foo@example.com",
				"2nd@example.net",
				"123456789",
				"987654321",
				[]any{
					"123 Main Street",
					"Whatever", "OR", "98754",
				},
			},
			rand: true,
		},
		{
			name:  "do_not_include_parent_key",
			segs:  []*Segment{Descendant(Name("o")), Child(Name("k"))},
			input: map[string]any{"o": map[string]any{"o": "hi", "k": 2}},
			exp:   []any{2},
		},
		{
			name:  "do_not_include_parent_index",
			segs:  []*Segment{Descendant(Index(0)), Child(Index(1))},
			input: []any{[]any{42, 98}},
			exp:   []any{98},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(a)
		})
	}
}

func TestQueryInputs(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	x := map[string]any{"x": "x"}
	y := map[string]any{"y": "y"}

	// Test current.
	q := Query(false, []*Segment{Child(Name("x"))})
	a.False(q.root)
	a.Equal([]any{"x"}, q.Select(x, y))
	a.Equal([]any{}, q.Select(y, x))

	// Test root.
	q.root = true
	a.Equal([]any{}, q.Select(x, y))
	a.Equal([]any{"x"}, q.Select(y, x))
}

func TestSingularExpr(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name  string
		query *PathQuery
		sing  *SingularQueryExpr
	}{
		{
			name:  "relative_singular",
			query: Query(false, []*Segment{Child(Name("j"))}),
			sing:  SingularQuery(false, []Selector{Name("j")}),
		},
		{
			name:  "root_singular",
			query: &PathQuery{segments: []*Segment{Child(Name("j")), Child(Index(0))}, root: true},
			sing:  SingularQuery(true, []Selector{Name("j"), Index(0)}),
		},
		{
			name:  "descendant",
			query: Query(false, []*Segment{Descendant(Name("j"))}),
		},
		{
			name:  "multi_selector",
			query: Query(false, []*Segment{Child(Name("j"), Name("x"))}),
		},
		{
			name:  "single_slice",
			query: Query(false, []*Segment{Child(Slice())}),
		},
		{
			name:  "wildcard",
			query: Query(false, []*Segment{Child(Wildcard)}),
		},
		{
			name:  "filter",
			query: Query(false, []*Segment{Child(&FilterSelector{})}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.sing == nil {
				a.False(tc.query.isSingular())
				a.Nil(tc.query.Singular())
				a.Equal(FilterQuery(tc.query), tc.query.Expression())
			} else {
				a.True(tc.query.isSingular())
				a.Equal(tc.sing, tc.query.Singular())
				a.Equal(tc.sing, tc.query.Expression())
			}
		})
	}
}
