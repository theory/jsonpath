package spec

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalSelector(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		elem NormalSelector
		str  string
		ptr  string
	}{
		{
			name: "object_value",
			elem: Name("a"),
			str:  `['a']`,
			ptr:  `a`,
		},
		{
			name: "array_index",
			elem: Index(1),
			str:  `[1]`,
			ptr:  `1`,
		},
		{
			name: "escape_apostrophes",
			elem: Name("'hi'"),
			str:  `['\'hi\'']`,
			ptr:  "'hi'",
		},
		{
			name: "escapes",
			elem: Name("'\b\f\n\r\t\\'"),
			str:  `['\'\b\f\n\r\t\\\'']`,
			ptr:  "'\b\f\n\r\t\\'",
		},
		{
			name: "escape_vertical_unicode",
			elem: Name("\u000B"),
			str:  `['\u000b']`,
			ptr:  "\u000B",
		},
		{
			name: "escape_unicode_null",
			elem: Name("\u0000"),
			str:  `['\u0000']`,
			ptr:  "\u0000",
		},
		{
			name: "escape_unicode_runes",
			elem: Name("\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u000e\u000F"),
			str:  `['\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u000e\u000f']`,
			ptr:  "\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u000e\u000F",
		},
		{
			name: "escape_pointer",
			elem: Name("this / ~that"),
			str:  `['this / ~that']`,
			ptr:  "this ~1 ~0that",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := new(strings.Builder)
			tc.elem.writeNormalizedTo(buf)
			a.Equal(tc.str, buf.String())
			buf.Reset()
			tc.elem.writePointerTo(buf)
			a.Equal(tc.ptr, buf.String())
		})
	}
}

func TestNormalizedPath(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		path NormalizedPath
		str  string
		ptr  string
	}{
		{
			name: "empty_path",
			path: NormalizedPath{},
			str:  "$",
			ptr:  "",
		},
		{
			name: "object_value",
			path: NormalizedPath{Name("a")},
			str:  "$['a']",
			ptr:  "/a",
		},
		{
			name: "array_index",
			path: NormalizedPath{Index(1)},
			str:  "$[1]",
			ptr:  "/1",
		},
		{
			name: "neg_for_len_5",
			path: NormalizedPath{Index(2)},
			str:  "$[2]",
			ptr:  "/2",
		},
		{
			name: "nested_structure",
			path: NormalizedPath{Name("a"), Name("b"), Index(1)},
			str:  "$['a']['b'][1]",
			ptr:  "/a/b/1",
		},
		{
			name: "unicode_escape",
			path: NormalizedPath{Name("\u000B")},
			str:  `$['\u000b']`,
			ptr:  "/\u000b",
		},
		{
			name: "unicode_character",
			path: NormalizedPath{Name("\u0061")},
			str:  "$['a']",
			ptr:  "/a",
		},
		{
			name: "nested_structure_pointer_stuff",
			path: NormalizedPath{Name("a~x"), Name("b/2"), Index(1)},
			str:  "$['a~x']['b/2'][1]",
			ptr:  "/a~0x/b~12/1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.str, tc.path.String())
			a.Equal(tc.ptr, tc.path.Pointer())
		})
	}
}

func TestNormalizedPathCompare(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		p1   NormalizedPath
		p2   NormalizedPath
		exp  int
	}{
		{
			name: "empty_paths",
			exp:  0,
		},
		{
			name: "same_name",
			p1:   NormalizedPath{Name("a")},
			p2:   NormalizedPath{Name("a")},
			exp:  0,
		},
		{
			name: "diff_names",
			p1:   NormalizedPath{Name("a")},
			p2:   NormalizedPath{Name("b")},
			exp:  -1,
		},
		{
			name: "diff_names_rev",
			p1:   NormalizedPath{Name("b")},
			p2:   NormalizedPath{Name("a")},
			exp:  1,
		},
		{
			name: "same_name_diff_lengths",
			p1:   NormalizedPath{Name("a"), Name("b")},
			p2:   NormalizedPath{Name("a")},
			exp:  1,
		},
		{
			name: "same_name_diff_lengths_rev",
			p1:   NormalizedPath{Name("a")},
			p2:   NormalizedPath{Name("a"), Name("b")},
			exp:  -1,
		},
		{
			name: "same_multi_names",
			p1:   NormalizedPath{Name("a"), Name("b")},
			p2:   NormalizedPath{Name("a"), Name("b")},
			exp:  0,
		},
		{
			name: "diff_nested_names",
			p1:   NormalizedPath{Name("a"), Name("a")},
			p2:   NormalizedPath{Name("a"), Name("b")},
			exp:  -1,
		},
		{
			name: "diff_nested_names_rev",
			p1:   NormalizedPath{Name("a"), Name("b")},
			p2:   NormalizedPath{Name("a"), Name("a")},
			exp:  1,
		},
		{
			name: "name_vs_index",
			p1:   NormalizedPath{Name("a")},
			p2:   NormalizedPath{Index(0)},
			exp:  1,
		},
		{
			name: "name_vs_index_rev",
			p1:   NormalizedPath{Index(0)},
			p2:   NormalizedPath{Name("a")},
			exp:  -1,
		},
		{
			name: "diff_nested_types",
			p1:   NormalizedPath{Name("a"), Index(1024)},
			p2:   NormalizedPath{Name("a"), Name("b")},
			exp:  -1,
		},
		{
			name: "diff_nested_types_rev",
			p1:   NormalizedPath{Name("a"), Name("b")},
			p2:   NormalizedPath{Name("a"), Index(1024)},
			exp:  1,
		},
		{
			name: "same_index",
			p1:   NormalizedPath{Index(42)},
			p2:   NormalizedPath{Index(42)},
			exp:  0,
		},
		{
			name: "diff_indexes",
			p1:   NormalizedPath{Index(42)},
			p2:   NormalizedPath{Index(99)},
			exp:  -1,
		},
		{
			name: "diff_indexes_rev",
			p1:   NormalizedPath{Index(99)},
			p2:   NormalizedPath{Index(42)},
			exp:  1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.exp, tc.p1.Compare(tc.p2))
		})
	}
}

func TestLocatedNode(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)

	for _, tc := range []struct {
		name string
		node LocatedNode
		exp  string
	}{
		{
			name: "simple",
			node: LocatedNode{Path: NormalizedPath{Name("a")}, Node: "foo"},
			exp:  `{"path": "$['a']", "node": "foo"}`,
		},
		{
			name: "double_quoted_path",
			node: LocatedNode{Path: NormalizedPath{Name(`"a"`)}, Node: 42},
			exp:  `{"path": "$['\"a\"']", "node": 42}`,
		},
		{
			name: "single_quoted_path",
			node: LocatedNode{Path: NormalizedPath{Name(`'a'`)}, Node: true},
			exp:  `{"path": "$['\\'a\\'']", "node": true}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			json, err := json.Marshal(tc.node)
			r.NoError(err)
			a.JSONEq(tc.exp, string(json))
		})
	}
}
