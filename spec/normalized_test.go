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

	for _, tc := range []struct {
		test string
		elem NormalSelector
		str  string
		ptr  string
	}{
		{
			test: "object_value",
			elem: Name("a"),
			str:  `['a']`,
			ptr:  `a`,
		},
		{
			test: "array_index",
			elem: Index(1),
			str:  `[1]`,
			ptr:  `1`,
		},
		{
			test: "escape_apostrophes",
			elem: Name("'hi'"),
			str:  `['\'hi\'']`,
			ptr:  "'hi'",
		},
		{
			test: "escapes",
			elem: Name("'\b\f\n\r\t\\'"),
			str:  `['\'\b\f\n\r\t\\\'']`,
			ptr:  "'\b\f\n\r\t\\'",
		},
		{
			test: "escape_vertical_unicode",
			elem: Name("\u000B"),
			str:  `['\u000b']`,
			ptr:  "\u000B",
		},
		{
			test: "escape_unicode_null",
			elem: Name("\u0000"),
			str:  `['\u0000']`,
			ptr:  "\u0000",
		},
		{
			test: "escape_unicode_runes",
			elem: Name("\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u000e\u000F"),
			str:  `['\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u000e\u000f']`,
			ptr:  "\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u000e\u000F",
		},
		{
			test: "escape_pointer",
			elem: Name("this / ~that"),
			str:  `['this / ~that']`,
			ptr:  "this ~1 ~0that",
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

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

	for _, tc := range []struct {
		test string
		path NormalizedPath
		str  string
		ptr  string
	}{
		{
			test: "empty_path",
			path: Normalized(),
			str:  "$",
			ptr:  "",
		},
		{
			test: "object_value",
			path: Normalized(Name("a")),
			str:  "$['a']",
			ptr:  "/a",
		},
		{
			test: "array_index",
			path: Normalized(Index(1)),
			str:  "$[1]",
			ptr:  "/1",
		},
		{
			test: "neg_for_len_5",
			path: Normalized(Index(2)),
			str:  "$[2]",
			ptr:  "/2",
		},
		{
			test: "nested_structure",
			path: Normalized(Name("a"), Name("b"), Index(1)),
			str:  "$['a']['b'][1]",
			ptr:  "/a/b/1",
		},
		{
			test: "unicode_escape",
			path: Normalized(Name("\u000B")),
			str:  `$['\u000b']`,
			ptr:  "/\u000b",
		},
		{
			test: "unicode_character",
			path: Normalized(Name("\u0061")),
			str:  "$['a']",
			ptr:  "/a",
		},
		{
			test: "nested_structure_pointer_stuff",
			path: Normalized(Name("a~x"), Name("b/2"), Index(1)),
			str:  "$['a~x']['b/2'][1]",
			ptr:  "/a~0x/b~12/1",
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.Equal(tc.str, tc.path.String())
			a.Equal(tc.ptr, tc.path.Pointer())
		})
	}
}

func TestNormalizedPathCompare(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
		p1   NormalizedPath
		p2   NormalizedPath
		exp  int
	}{
		{
			test: "empty_paths",
			exp:  0,
		},
		{
			test: "same_name",
			p1:   Normalized(Name("a")),
			p2:   Normalized(Name("a")),
			exp:  0,
		},
		{
			test: "diff_names",
			p1:   Normalized(Name("a")),
			p2:   Normalized(Name("b")),
			exp:  -1,
		},
		{
			test: "diff_names_rev",
			p1:   Normalized(Name("b")),
			p2:   Normalized(Name("a")),
			exp:  1,
		},
		{
			test: "same_name_diff_lengths",
			p1:   Normalized(Name("a"), Name("b")),
			p2:   Normalized(Name("a")),
			exp:  1,
		},
		{
			test: "same_name_diff_lengths_rev",
			p1:   Normalized(Name("a")),
			p2:   Normalized(Name("a"), Name("b")),
			exp:  -1,
		},
		{
			test: "same_multi_names",
			p1:   Normalized(Name("a"), Name("b")),
			p2:   Normalized(Name("a"), Name("b")),
			exp:  0,
		},
		{
			test: "diff_nested_names",
			p1:   Normalized(Name("a"), Name("a")),
			p2:   Normalized(Name("a"), Name("b")),
			exp:  -1,
		},
		{
			test: "diff_nested_names_rev",
			p1:   Normalized(Name("a"), Name("b")),
			p2:   Normalized(Name("a"), Name("a")),
			exp:  1,
		},
		{
			test: "name_vs_index",
			p1:   Normalized(Name("a")),
			p2:   Normalized(Index(0)),
			exp:  1,
		},
		{
			test: "name_vs_index_rev",
			p1:   Normalized(Index(0)),
			p2:   Normalized(Name("a")),
			exp:  -1,
		},
		{
			test: "diff_nested_types",
			p1:   Normalized(Name("a"), Index(1024)),
			p2:   Normalized(Name("a"), Name("b")),
			exp:  -1,
		},
		{
			test: "diff_nested_types_rev",
			p1:   Normalized(Name("a"), Name("b")),
			p2:   Normalized(Name("a"), Index(1024)),
			exp:  1,
		},
		{
			test: "same_index",
			p1:   Normalized(Index(42)),
			p2:   Normalized(Index(42)),
			exp:  0,
		},
		{
			test: "diff_indexes",
			p1:   Normalized(Index(42)),
			p2:   Normalized(Index(99)),
			exp:  -1,
		},
		{
			test: "diff_indexes_rev",
			p1:   Normalized(Index(99)),
			p2:   Normalized(Index(42)),
			exp:  1,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.exp, tc.p1.Compare(tc.p2))
		})
	}
}

func TestLocatedNode(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
		node LocatedNode
		exp  string
	}{
		{
			test: "simple",
			node: LocatedNode{Path: Normalized(Name("a")), Node: "foo"},
			exp:  `{"path": "$['a']", "node": "foo"}`,
		},
		{
			test: "double_quoted_path",
			node: LocatedNode{Path: Normalized(Name(`"a"`)), Node: 42},
			exp:  `{"path": "$['\"a\"']", "node": 42}`,
		},
		{
			test: "single_quoted_path",
			node: LocatedNode{Path: Normalized(Name(`'a'`)), Node: true},
			exp:  `{"path": "$['\\'a\\'']", "node": true}`,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()

			json, err := json.Marshal(tc.node)
			require.NoError(t, err)
			assert.JSONEq(t, tc.exp, string(json))
		})
	}
}
