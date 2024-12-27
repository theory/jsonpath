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
		exp  string
	}{
		{
			name: "object_value",
			elem: Name("a"),
			exp:  `['a']`,
		},
		{
			name: "array_index",
			elem: Index(1),
			exp:  `[1]`,
		},
		{
			name: "escape_apostrophes",
			elem: Name("'hi'"),
			exp:  `['\'hi\'']`,
		},
		{
			name: "escapes",
			elem: Name("'\b\f\n\r\t\\'"),
			exp:  `['\'\b\f\n\r\t\\\'']`,
		},
		{
			name: "escape_vertical_unicode",
			elem: Name("\u000B"),
			exp:  `['\u000b']`,
		},
		{
			name: "escape_unicode_null",
			elem: Name("\u0000"),
			exp:  `['\u0000']`,
		},
		{
			name: "escape_unicode_runes",
			elem: Name("\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u000e\u000F"),
			exp:  `['\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u000e\u000f']`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := new(strings.Builder)
			tc.elem.writeNormalizedTo(buf)
			a.Equal(tc.exp, buf.String())
		})
	}
}

func TestNormalizedPath(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		path NormalizedPath
		exp  string
	}{
		{
			name: "object_value",
			path: NormalizedPath{Name("a")},
			exp:  "$['a']",
		},
		{
			name: "array_index",
			path: NormalizedPath{Index(1)},
			exp:  "$[1]",
		},
		{
			name: "neg_for_len_5",
			path: NormalizedPath{Index(2)},
			exp:  "$[2]",
		},
		{
			name: "nested_structure",
			path: NormalizedPath{Name("a"), Name("b"), Index(1)},
			exp:  "$['a']['b'][1]",
		},
		{
			name: "unicode_escape",
			path: NormalizedPath{Name("\u000B")},
			exp:  `$['\u000b']`,
		},
		{
			name: "unicode_character",
			path: NormalizedPath{Name("\u0061")},
			exp:  "$['a']",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.exp, tc.path.String())
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
