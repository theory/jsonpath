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
			q := Query(false)
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
			q := Query(false, tc.segs...)
			a.Equal("@"+tc.str, q.String())
			a.Equal("@"+tc.str, bufString(q))
			q = Query(true, tc.segs...)
			a.Equal("$"+tc.str, q.String())
			a.Equal("$"+tc.str, bufString(q))
		})
	}
}

type queryTestCase struct {
	name    string
	resType FuncType
	segs    []*Segment
	input   any
	exp     []any
	loc     []*LocatedNode
	rand    bool
}

func (tc queryTestCase) run(a *assert.Assertions) {
	// Set up Query.
	q := Query(false, tc.segs...)
	a.Equal(tc.segs, q.Segments())
	a.False(q.root)

	// Test Select and SelectLocated.
	if tc.rand {
		a.ElementsMatch(tc.exp, q.Select(tc.input, nil))
		a.ElementsMatch(tc.loc, q.SelectLocated(tc.input, nil, NormalizedPath{}))
	} else {
		a.Equal(tc.exp, q.Select(tc.input, nil))
		a.Equal(tc.loc, q.SelectLocated(tc.input, nil, NormalizedPath{}))
	}

	// Test result type and conversion.
	a.Equal(tc.resType, q.ResultType())
	a.Equal(tc.resType == FuncValue, q.ConvertsTo(FuncValue))
	a.True(q.ConvertsTo(FuncNodes))
	a.False(q.ConvertsTo(FuncLogical))
}

func TestQueryObject(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []queryTestCase{
		{
			name:    "root",
			resType: FuncValue,
			input:   map[string]any{"x": true, "y": []any{1, 2}},
			exp:     []any{map[string]any{"x": true, "y": []any{1, 2}}},
			loc: []*LocatedNode{
				{Path: NormalizedPath{}, Node: map[string]any{"x": true, "y": []any{1, 2}}},
			},
		},
		{
			name:    "one_key_scalar",
			resType: FuncValue,
			segs:    []*Segment{Child(Name("x"))},
			input:   map[string]any{"x": true, "y": []any{1, 2}},
			exp:     []any{true},
			loc: []*LocatedNode{
				{Path: Normalized(Name("x")), Node: true},
			},
		},
		{
			name:    "one_key_array",
			resType: FuncValue,
			segs:    []*Segment{Child(Name("y"))},
			input:   map[string]any{"x": true, "y": []any{1, 2}},
			exp:     []any{[]any{1, 2}},
			loc: []*LocatedNode{
				{Path: Normalized(Name("y")), Node: []any{1, 2}},
			},
		},
		{
			name:    "one_key_object",
			resType: FuncValue,
			segs:    []*Segment{Child(Name("y"))},
			input:   map[string]any{"x": true, "y": map[string]any{"a": 1}},
			exp:     []any{map[string]any{"a": 1}},
			loc: []*LocatedNode{
				{Path: Normalized(Name("y")), Node: map[string]any{"a": 1}},
			},
		},
		{
			name:    "multiple_keys",
			resType: FuncNodes,
			segs:    []*Segment{Child(Name("x"), Name("y"))},
			input:   map[string]any{"x": true, "y": []any{1, 2}, "z": "hi"},
			exp:     []any{true, []any{1, 2}},
			loc: []*LocatedNode{
				{Path: Normalized(Name("x")), Node: true},
				{Path: Normalized(Name("y")), Node: []any{1, 2}},
			},
			rand: true,
		},
		{
			name:    "three_level_path",
			resType: FuncValue,
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
			loc: []*LocatedNode{
				{Path: Normalized(Name("x"), Name("a"), Name("i")), Node: []any{1, 2}},
			},
		},
		{
			name:    "wildcard_keys",
			resType: FuncNodes,
			segs: []*Segment{
				Child(Wildcard),
				Child(Name("a"), Name("b")),
			},
			input: map[string]any{
				"x": map[string]any{"a": "go", "b": 2, "c": 5},
				"y": map[string]any{"a": 2, "b": 3, "d": 3},
			},
			exp: []any{"go", 2, 2, 3},
			loc: []*LocatedNode{
				{Path: Normalized(Name("x"), Name("a")), Node: "go"},
				{Path: Normalized(Name("x"), Name("b")), Node: 2},
				{Path: Normalized(Name("y"), Name("a")), Node: 2},
				{Path: Normalized(Name("y"), Name("b")), Node: 3},
			},
			rand: true,
		},
		{
			name:    "any_key_indexes",
			resType: FuncNodes,
			segs: []*Segment{
				Child(Wildcard),
				Child(Index(0), Index(1)),
			},
			input: map[string]any{
				"x": []any{"a", "go", "b", 2, "c", 5},
				"y": []any{"a", 2, "b", 3, "d", 3},
			},
			exp: []any{"a", "go", "a", 2},
			loc: []*LocatedNode{
				{Path: Normalized(Name("x"), Index(0)), Node: "a"},
				{Path: Normalized(Name("x"), Index(1)), Node: "go"},
				{Path: Normalized(Name("y"), Index(0)), Node: "a"},
				{Path: Normalized(Name("y"), Index(1)), Node: 2},
			},
			rand: true,
		},
		{
			name:    "any_key_nonexistent_index",
			resType: FuncNodes,
			segs:    []*Segment{Child(Wildcard), Child(Index(1))},
			input: map[string]any{
				"x": []any{"a", "go", "b", 2, "c", 5},
				"y": []any{"a"},
			},
			exp: []any{"go"},
			loc: []*LocatedNode{
				{Path: Normalized(Name("x"), Index(1)), Node: "go"},
			},
		},
		{
			name:    "nonexistent_key",
			resType: FuncValue,
			segs:    []*Segment{Child(Name("x"))},
			input:   map[string]any{"y": []any{1, 2}},
			exp:     []any{},
			loc:     []*LocatedNode{},
		},
		{
			name:    "nonexistent_branch_key",
			resType: FuncValue,
			segs:    []*Segment{Child(Name("x")), Child(Name("z"))},
			input:   map[string]any{"y": []any{1, 2}},
			exp:     []any{},
			loc:     []*LocatedNode{},
		},
		{
			name:    "wildcard_then_nonexistent_key",
			resType: FuncNodes,
			segs:    []*Segment{Child(Wildcard), Child(Name("x"))},
			input:   map[string]any{"y": map[string]any{"a": 1}},
			exp:     []any{},
			loc:     []*LocatedNode{},
		},
		{
			name:    "not_an_object",
			resType: FuncValue,
			segs:    []*Segment{Child(Name("x")), Child(Name("y"))},
			input:   map[string]any{"x": true},
			exp:     []any{},
			loc:     []*LocatedNode{},
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
			name:    "root",
			resType: FuncValue,
			input:   []any{"x", true, "y", []any{1, 2}},
			exp:     []any{[]any{"x", true, "y", []any{1, 2}}},
			loc: []*LocatedNode{
				{Path: NormalizedPath{}, Node: []any{"x", true, "y", []any{1, 2}}},
			},
		},
		{
			name:    "index_zero",
			resType: FuncValue,
			segs:    []*Segment{Child(Index(0))},
			input:   []any{"x", true, "y", []any{1, 2}},
			exp:     []any{"x"},
			loc:     []*LocatedNode{{Path: Normalized(Index(0)), Node: "x"}},
		},
		{
			name:    "index_one",
			resType: FuncValue,
			segs:    []*Segment{Child(Index(1))},
			input:   []any{"x", true, "y", []any{1, 2}},
			exp:     []any{true},
			loc:     []*LocatedNode{{Path: Normalized(Index(1)), Node: true}},
		},
		{
			name:    "index_three",
			resType: FuncValue,
			segs:    []*Segment{Child(Index(3))},
			input:   []any{"x", true, "y", []any{1, 2}},
			exp:     []any{[]any{1, 2}},
			loc:     []*LocatedNode{{Path: Normalized(Index(3)), Node: []any{1, 2}}},
		},
		{
			name:    "multiple_indexes",
			resType: FuncNodes,
			segs:    []*Segment{Child(Index(1), Index(3))},
			input:   []any{"x", true, "y", []any{1, 2}},
			exp:     []any{true, []any{1, 2}},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1)), Node: true},
				{Path: Normalized(Index(3)), Node: []any{1, 2}},
			},
		},
		{
			name:    "nested_indices",
			resType: FuncValue,
			segs:    []*Segment{Child(Index(0)), Child(Index(0))},
			input:   []any{[]any{1, 2}, "x", true, "y"},
			exp:     []any{1},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(0)), Node: 1},
			},
		},
		{
			name:    "nested_multiple_indices",
			resType: FuncNodes,
			segs:    []*Segment{Child(Index(0)), Child(Index(0), Index(1))},
			input:   []any{[]any{1, 2, 3}, "x", true, "y"},
			exp:     []any{1, 2},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(0)), Node: 1},
				{Path: Normalized(Index(0), Index(1)), Node: 2},
			},
		},
		{
			name:    "nested_index_gaps",
			resType: FuncValue,
			segs:    []*Segment{Child(Index(1)), Child(Index(1))},
			input:   []any{"x", []any{1, 2}, true, "y"},
			exp:     []any{2},
			loc: []*LocatedNode{
				{Path: Normalized(Index(1), Index(1)), Node: 2},
			},
		},
		{
			name:    "three_level_index_path",
			resType: FuncValue,
			segs: []*Segment{
				Child(Index(0)),
				Child(Index(0)),
				Child(Index(0)),
			},
			input: []any{[]any{[]any{42, 12}, 2}, "x", true, "y"},
			exp:   []any{42},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(0), Index(0)), Node: 42},
			},
		},
		{
			name:    "mixed_nesting",
			resType: FuncNodes,
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
			exp: []any{2, "hi", 1},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(1)), Node: 2},
				{Path: Normalized(Index(3), Name("y")), Node: "hi"},
				{Path: Normalized(Index(3), Name("z")), Node: 1},
			},
			rand: true,
		},
		{
			name:    "wildcard_indexes_index",
			resType: FuncNodes,
			segs:    []*Segment{Child(Wildcard), Child(Index(0), Index(2))},
			input:   []any{[]any{1, 2, 3}, []any{3, 2, 1}, []any{4, 5, 6}},
			exp:     []any{1, 3, 3, 1, 4, 6},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(0)), Node: 1},
				{Path: Normalized(Index(0), Index(2)), Node: 3},
				{Path: Normalized(Index(1), Index(0)), Node: 3},
				{Path: Normalized(Index(1), Index(2)), Node: 1},
				{Path: Normalized(Index(2), Index(0)), Node: 4},
				{Path: Normalized(Index(2), Index(2)), Node: 6},
			},
		},
		{
			name:    "nonexistent_index",
			resType: FuncValue,
			segs:    []*Segment{Child(Index(3))},
			input:   []any{"y", []any{1, 2}},
			exp:     []any{},
			loc:     []*LocatedNode{},
		},
		{
			name:    "nonexistent_child_index",
			resType: FuncNodes,
			segs:    []*Segment{Child(Wildcard), Child(Index(3))},
			input:   []any{[]any{0, 1, 2, 3}, []any{0, 1, 2}},
			exp:     []any{3},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(3)), Node: 3},
			},
		},
		{
			name:    "not_an_array_index_1",
			resType: FuncValue,
			segs:    []*Segment{Child(Index(1)), Child(Index(0))},
			input:   []any{"x", true},
			exp:     []any{},
			loc:     []*LocatedNode{},
		},
		{
			name:    "not_an_array_index_0",
			resType: FuncValue,
			segs:    []*Segment{Child(Index(0)), Child(Index(0))},
			input:   []any{"x", true},
			exp:     []any{},
			loc:     []*LocatedNode{},
		},
		{
			name:    "wildcard_not_an_array_index_1",
			resType: FuncNodes,
			segs:    []*Segment{Child(Wildcard), Child(Index(0))},
			input:   []any{"x", true},
			exp:     []any{},
			loc:     []*LocatedNode{},
		},
		{
			name:    "mix_wildcard_keys",
			resType: FuncNodes,
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
			exp: []any{"hi", "go", "bo", 42, true, 21, 53, "bo", 42},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Name("x")), Node: "hi"},
				{Path: Normalized(Index(0), Name("y")), Node: "go"},
				{Path: Normalized(Index(1), Name("x")), Node: "bo"},
				{Path: Normalized(Index(1), Name("y")), Node: 42},
				{Path: Normalized(Index(2), Name("x")), Node: true},
				{Path: Normalized(Index(2), Name("y")), Node: 21},
				{Path: Normalized(Index(3), Index(1)), Node: 53},
				{Path: Normalized(Index(1), Name("x")), Node: "bo"},
				{Path: Normalized(Index(1), Name("y")), Node: 42},
			},
			rand: true,
		},
		{
			name:    "mix_wildcard_nonexistent_key",
			resType: FuncNodes,
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Name("x")), Node: "hi"},
				{Path: Normalized(Index(1), Name("x")), Node: "bo"},
				{Path: Normalized(Index(2), Name("x")), Node: true},
				{Path: Normalized(Index(1), Name("x")), Node: "bo"},
			},
		},
		{
			name:    "mix_wildcard_index",
			resType: FuncNodes,
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(0)), Node: "x"},
				{Path: Normalized(Index(0), Index(1)), Node: "hi"},
				{Path: Normalized(Index(1), Index(0)), Node: "x"},
				{Path: Normalized(Index(1), Index(1)), Node: "bo"},
				{Path: Normalized(Index(2), Index(0)), Node: "x"},
				{Path: Normalized(Index(2), Index(1)), Node: true},
				{Path: Normalized(Index(1), Index(0)), Node: "x"},
				{Path: Normalized(Index(1), Index(1)), Node: "bo"},
			},
		},
		{
			name:    "mix_wildcard_nonexistent_index",
			resType: FuncNodes,
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(0)), Node: "x"},
				{Path: Normalized(Index(1), Index(0)), Node: "x"},
				{Path: Normalized(Index(2), Index(0)), Node: "x"},
				{Path: Normalized(Index(1), Index(0)), Node: "x"},
			},
		},
		{
			name:    "wildcard_nonexistent_key",
			resType: FuncNodes,
			segs:    []*Segment{Child(Wildcard), Child(Name("a"))},
			input: []any{
				map[string]any{"a": 1, "b": 2},
				map[string]any{"z": 3, "b": 4},
			},
			exp: []any{1},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Name("a")), Node: 1},
			},
		},
		{
			name:    "wildcard_nonexistent_middle_key",
			resType: FuncNodes,
			segs:    []*Segment{Child(Wildcard), Child(Name("a"))},
			input: []any{
				map[string]any{"a": 1, "b": 2},
				map[string]any{"z": 3, "b": 4},
				map[string]any{"a": 5},
				map[string]any{"z": 3, "b": 4},
			},
			exp: []any{1, 5},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Name("a")), Node: 1},
				{Path: Normalized(Index(2), Name("a")), Node: 5},
			},
		},
		{
			name:    "wildcard_nested_nonexistent_key",
			resType: FuncNodes,
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Name("x"), Name("a")), Node: 1},
			},
		},
		{
			name:    "wildcard_nested_nonexistent_index",
			resType: FuncNodes,
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Name("x"), Index(1)), Node: 2},
			},
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: "x"},
				{Path: Normalized(Index(1)), Node: true},
			},
		},
		{
			name:  "slice_0_1",
			segs:  []*Segment{Child(Slice(0, 1))},
			input: []any{"x", true, "y", []any{1, 2}},
			exp:   []any{"x"},
			loc:   []*LocatedNode{{Path: Normalized(Index(0)), Node: "x"}},
		},
		{
			name:  "slice_2_5",
			segs:  []*Segment{Child(Slice(2, 5))},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"y", []any{1, 2}, 42},
			loc: []*LocatedNode{
				{Path: Normalized(Index(2)), Node: "y"},
				{Path: Normalized(Index(3)), Node: []any{1, 2}},
				{Path: Normalized(Index(4)), Node: 42},
			},
		},
		{
			name:  "slice_2_5_over_len",
			segs:  []*Segment{Child(Slice(2, 5))},
			input: []any{"x", true, "y"},
			exp:   []any{"y"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(2)), Node: "y"},
			},
		},
		{
			name:  "slice_defaults",
			segs:  []*Segment{Child(Slice())},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: "x"},
				{Path: Normalized(Index(1)), Node: true},
				{Path: Normalized(Index(2)), Node: "y"},
				{Path: Normalized(Index(3)), Node: []any{1, 2}},
				{Path: Normalized(Index(4)), Node: 42},
				{Path: Normalized(Index(5)), Node: nil},
				{Path: Normalized(Index(6)), Node: 78},
			},
		},
		{
			name:  "default_start",
			segs:  []*Segment{Child(Slice(nil, 2))},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"x", true},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: "x"},
				{Path: Normalized(Index(1)), Node: true},
			},
		},
		{
			name:  "default_end",
			segs:  []*Segment{Child(Slice(2))},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"y", []any{1, 2}, 42, nil, 78},
			loc: []*LocatedNode{
				{Path: Normalized(Index(2)), Node: "y"},
				{Path: Normalized(Index(3)), Node: []any{1, 2}},
				{Path: Normalized(Index(4)), Node: 42},
				{Path: Normalized(Index(5)), Node: nil},
				{Path: Normalized(Index(6)), Node: 78},
			},
		},
		{
			name:  "step_2",
			segs:  []*Segment{Child(Slice(nil, nil, 2))},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"x", "y", 42, 78},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: "x"},
				{Path: Normalized(Index(2)), Node: "y"},
				{Path: Normalized(Index(4)), Node: 42},
				{Path: Normalized(Index(6)), Node: 78},
			},
		},
		{
			name:  "step_3",
			segs:  []*Segment{Child(Slice(nil, nil, 3))},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"x", []any{1, 2}, 78},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: "x"},
				{Path: Normalized(Index(3)), Node: []any{1, 2}},
				{Path: Normalized(Index(6)), Node: 78},
			},
		},
		{
			name:  "multiple_slices",
			segs:  []*Segment{Child(Slice(0, 1), Slice(3, 4))},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"x", []any{1, 2}},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: "x"},
				{Path: Normalized(Index(3)), Node: []any{1, 2}},
			},
		},
		{
			name:  "overlapping_slices",
			segs:  []*Segment{Child(Slice(0, 3), Slice(2, 4))},
			input: []any{"x", true, "y", []any{1, 2}, 42, nil, 78},
			exp:   []any{"x", true, "y", "y", []any{1, 2}},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0)), Node: "x"},
				{Path: Normalized(Index(1)), Node: true},
				{Path: Normalized(Index(2)), Node: "y"},
				{Path: Normalized(Index(2)), Node: "y"},
				{Path: Normalized(Index(3)), Node: []any{1, 2}},
			},
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(1)), Node: 42},
				{Path: Normalized(Index(1), Index(1)), Node: "on"},
			},
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(1)), Node: 42},
				{Path: Normalized(Index(0), Index(3)), Node: 64},
				{Path: Normalized(Index(0), Index(4)), Node: []any{}},
				{Path: Normalized(Index(1), Index(1)), Node: "on"},
				{Path: Normalized(Index(1), Index(3)), Node: 88},
				{Path: Normalized(Index(1), Index(4)), Node: []any{1}},
			},
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(0), Index(0)), Node: 42},
				{Path: Normalized(Index(1), Index(0), Index(0)), Node: 16},
			},
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(0), Index(0)), Node: 42},
				{Path: Normalized(Index(2), Index(0), Name("y")), Node: "hi"},
				{Path: Normalized(Index(2), Index(0), Name("z")), Node: 1},
			},
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(0)), Node: 1},
				{Path: Normalized(Index(0), Index(1)), Node: 2},
				{Path: Normalized(Index(0), Index(3)), Node: 4},
				{Path: Normalized(Index(1), Index(0)), Node: 3},
				{Path: Normalized(Index(1), Index(1)), Node: 2},
				{Path: Normalized(Index(1), Index(3)), Node: 0},
				{Path: Normalized(Index(2), Index(0)), Node: 4},
				{Path: Normalized(Index(2), Index(1)), Node: 5},
				{Path: Normalized(Index(2), Index(3)), Node: 7},
			},
		},
		{
			name:  "nonexistent_slice",
			segs:  []*Segment{Child(Slice(3, 5))},
			input: []any{"y", []any{1, 2}},
			exp:   []any{},
			loc:   []*LocatedNode{},
		},
		{
			name:  "nonexistent_branch_index",
			segs:  []*Segment{Child(Wildcard), Child(Slice(3, 5))},
			input: []any{[]any{0, 1, 2, 3, 4}, []any{0, 1, 2}},
			exp:   []any{3, 4},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(3)), Node: 3},
				{Path: Normalized(Index(0), Index(4)), Node: 4},
			},
		},
		{
			name:    "not_an_array_index_1",
			resType: FuncValue,
			segs:    []*Segment{Child(Index(1)), Child(Index(0))},
			input:   []any{"x", true},
			exp:     []any{},
			loc:     []*LocatedNode{},
		},
		{
			name:  "not_an_array",
			segs:  []*Segment{Child(Slice(0, 5)), Child(Index(0))},
			input: []any{"x", true},
			exp:   []any{},
			loc:   []*LocatedNode{},
		},
		{
			name:  "wildcard_not_an_array_index_1",
			segs:  []*Segment{Child(Wildcard), Child(Slice(0, 5))},
			input: []any{"x", true},
			exp:   []any{},
			loc:   []*LocatedNode{},
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
			exp: []any{"hi", "go", "bo", 42, true, 21, "bo", 42},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Name("x")), Node: "hi"},
				{Path: Normalized(Index(0), Name("y")), Node: "go"},
				{Path: Normalized(Index(1), Name("x")), Node: "bo"},
				{Path: Normalized(Index(1), Name("y")), Node: 42},
				{Path: Normalized(Index(2), Name("x")), Node: true},
				{Path: Normalized(Index(2), Name("y")), Node: 21},
				{Path: Normalized(Index(1), Name("x")), Node: "bo"},
				{Path: Normalized(Index(1), Name("y")), Node: 42},
			},
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Name("x")), Node: "hi"},
				{Path: Normalized(Index(1), Name("x")), Node: "bo"},
				{Path: Normalized(Index(2), Name("x")), Node: true},
				{Path: Normalized(Index(1), Name("x")), Node: "bo"},
			},
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(0)), Node: "x"},
				{Path: Normalized(Index(0), Index(1)), Node: "hi"},
				{Path: Normalized(Index(1), Index(0)), Node: "y"},
				{Path: Normalized(Index(1), Index(1)), Node: "bo"},
				{Path: Normalized(Index(2), Index(0)), Node: "z"},
				{Path: Normalized(Index(2), Index(1)), Node: true},
				{Path: Normalized(Index(1), Index(0)), Node: "y"},
				{Path: Normalized(Index(1), Index(1)), Node: "bo"},
			},
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(0)), Node: "x"},
				{Path: Normalized(Index(1), Index(0)), Node: "y"},
				{Path: Normalized(Index(2), Index(0)), Node: "z"},
				{Path: Normalized(Index(1), Index(0)), Node: "y"},
			},
		},
		{
			name: "slice_nonexistent_key",
			segs: []*Segment{Child(Slice(0, 5)), Child(Name("a"))},
			input: []any{
				map[string]any{"a": 1, "b": 2},
				map[string]any{"z": 3, "b": 4},
			},
			exp: []any{1},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Name("a")), Node: 1},
			},
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Name("a")), Node: 1},
				{Path: Normalized(Index(2), Name("a")), Node: 5},
			},
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Name("x"), Name("a")), Node: 1},
			},
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Name("x"), Index(1)), Node: 2},
			},
		},
		{
			name:  "slice_neg",
			segs:  []*Segment{Child(Slice(nil, nil, -1))},
			input: []any{"x", true, "y", []any{1, 2}},
			exp:   []any{[]any{1, 2}, "y", true, "x"},
			loc: []*LocatedNode{
				{Path: Normalized(Index(3)), Node: []any{1, 2}},
				{Path: Normalized(Index(2)), Node: "y"},
				{Path: Normalized(Index(1)), Node: true},
				{Path: Normalized(Index(0)), Node: "x"},
			},
		},
		{
			name:  "slice_5_0_neg2",
			segs:  []*Segment{Child(Slice(5, 0, -2))},
			input: []any{"x", true, "y", 8, 13, 25, 23, 78, 13},
			exp:   []any{25, 8, true},
			loc: []*LocatedNode{
				{Path: Normalized(Index(5)), Node: 25},
				{Path: Normalized(Index(3)), Node: 8},
				{Path: Normalized(Index(1)), Node: true},
			},
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
			loc: []*LocatedNode{
				{Path: Normalized(Index(2), Index(2)), Node: false},
				{Path: Normalized(Index(2), Index(1)), Node: 98.6},
				{Path: Normalized(Index(1), Index(1)), Node: "on"},
				{Path: Normalized(Index(0), Index(2)), Node: true},
				{Path: Normalized(Index(0), Index(1)), Node: 42},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.resType == 0 {
				tc.resType = FuncNodes
			}
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
			loc: []*LocatedNode{
				{Path: Normalized(Name("o"), Name("j")), Node: 1},
				{Path: Normalized(Name("a"), Index(2), Index(0), Name("j")), Node: 4},
			},
			rand: true,
		},
		{
			name:  "un_descendant_name",
			segs:  []*Segment{Descendant(Name("o"))},
			input: json,
			exp:   []any{map[string]any{"j": 1, "k": 2}},
			loc: []*LocatedNode{
				{Path: Normalized(Name("o")), Node: map[string]any{"j": 1, "k": 2}},
			},
		},
		{
			name:  "nested_name",
			segs:  []*Segment{Child(Name("o")), Descendant(Name("k"))},
			input: json,
			exp:   []any{2},
			loc: []*LocatedNode{
				{Path: Normalized(Name("o"), Name("k")), Node: 2},
			},
		},
		{
			name:  "nested_wildcard",
			segs:  []*Segment{Child(Name("o")), Descendant(Wildcard)},
			input: json,
			exp:   []any{1, 2},
			loc: []*LocatedNode{
				{Path: Normalized(Name("o"), Name("j")), Node: 1},
				{Path: Normalized(Name("o"), Name("k")), Node: 2},
			},
			rand: true,
		},
		{
			name:  "single_index",
			segs:  []*Segment{Descendant(Index(0))},
			input: json,
			exp:   []any{5, map[string]any{"j": 4}},
			loc: []*LocatedNode{
				{Path: Normalized(Name("a"), Index(0)), Node: 5},
				{Path: Normalized(Name("a"), Index(2), Index(0)), Node: map[string]any{"j": 4}},
			},
		},
		{
			name:  "nested_index",
			segs:  []*Segment{Child(Name("a")), Descendant(Index(0))},
			input: json,
			exp:   []any{5, map[string]any{"j": 4}},
			loc: []*LocatedNode{
				{Path: Normalized(Name("a"), Index(0)), Node: 5},
				{Path: Normalized(Name("a"), Index(2), Index(0)), Node: map[string]any{"j": 4}},
			},
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
			loc: []*LocatedNode{
				{Path: Normalized(Name("profile"), Name("name"), Name("last")), Node: "Obama"},
				{
					Path: Normalized(Name("profile"), Name("contacts"), Name("email"), Name("primary")),
					Node: "foo@example.com",
				},
				{
					Path: Normalized(Name("profile"), Name("contacts"), Name("email"), Name("secondary")),
					Node: "2nd@example.net",
				},
				{
					Path: Normalized(Name("profile"), Name("contacts"), Name("phones"), Name("primary")),
					Node: "123456789",
				},
				{
					Path: Normalized(Name("profile"), Name("contacts"), Name("phones"), Name("secondary")),
					Node: "987654321",
				},
				{
					Path: Normalized(Name("profile"), Name("contacts"), Name("addresses"), Name("primary")),
					Node: []any{
						"123 Main Street",
						"Whatever", "OR", "98754",
					},
				},
			},
			rand: true,
		},
		{
			name:  "do_not_include_parent_key",
			segs:  []*Segment{Descendant(Name("o")), Child(Name("k"))},
			input: map[string]any{"o": map[string]any{"o": "hi", "k": 2}},
			exp:   []any{2},
			loc: []*LocatedNode{
				{Path: Normalized(Name("o"), Name("k")), Node: 2},
			},
		},
		{
			name:  "do_not_include_parent_index",
			segs:  []*Segment{Descendant(Index(0)), Child(Index(1))},
			input: []any{[]any{42, 98}},
			exp:   []any{98},
			loc: []*LocatedNode{
				{Path: Normalized(Index(0), Index(1)), Node: 98},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.resType = FuncNodes
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
	q := Query(false, Child(Name("x")))
	a.False(q.root)
	a.Equal([]any{"x"}, q.Select(x, y))
	a.Equal([]any{}, q.Select(y, x))
	a.Equal(
		[]*LocatedNode{{Path: Normalized(Name("x")), Node: "x"}},
		q.SelectLocated(x, y, NormalizedPath{}),
	)
	a.Equal([]*LocatedNode{}, q.SelectLocated(y, x, NormalizedPath{}))

	// Test root.
	q.root = true
	a.Equal([]any{}, q.Select(x, y))
	a.Equal([]any{"x"}, q.Select(y, x))
	a.Equal([]*LocatedNode{}, q.SelectLocated(x, y, NormalizedPath{}))
	a.Equal(
		[]*LocatedNode{{Path: Normalized(Name("x")), Node: "x"}},
		q.SelectLocated(y, x, NormalizedPath{}),
	)
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
			query: Query(false, Child(Name("j"))),
			sing:  SingularQuery(false, Name("j")),
		},
		{
			name:  "root_singular",
			query: &PathQuery{segments: []*Segment{Child(Name("j")), Child(Index(0))}, root: true},
			sing:  SingularQuery(true, Name("j"), Index(0)),
		},
		{
			name:  "descendant",
			query: Query(false, Descendant(Name("j"))),
		},
		{
			name:  "multi_selector",
			query: Query(false, Child(Name("j"), Name("x"))),
		},
		{
			name:  "single_slice",
			query: Query(false, Child(Slice())),
		},
		{
			name:  "wildcard",
			query: Query(false, Child(Wildcard)),
		},
		{
			name:  "filter",
			query: Query(false, Child(&FilterSelector{})),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.sing == nil {
				a.False(tc.query.isSingular())
				a.Nil(tc.query.Singular())
				a.Equal(tc.query, tc.query.Expression())
			} else {
				a.True(tc.query.isSingular())
				a.Equal(tc.sing, tc.query.Singular())
				a.Equal(tc.sing, tc.query.Expression())
			}
		})
	}
}
