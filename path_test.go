package jsonpath

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theory/jsonpath/registry"
	"github.com/theory/jsonpath/spec"
)

func book(idx int) spec.NormalizedPath {
	return spec.NormalizedPath{spec.Name("store"), spec.Name("book"), spec.Index(idx)}
}

func TestParseSpecExamples(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	val := specExampleJSON(t)
	store, _ := val["store"].(map[string]any)
	books, _ := store["book"].([]any)

	for _, tc := range []struct {
		name string
		path string
		exp  []any
		loc  []*spec.LocatedNode
		size int
		rand bool
	}{
		{
			name: "example_1",
			path: `$.store.book[*].author`,
			exp:  []any{"Nigel Rees", "Evelyn Waugh", "Herman Melville", "J. R. R. Tolkien"},
			loc: []*spec.LocatedNode{
				{Path: append(book(0), spec.Name("author")), Node: "Nigel Rees"},
				{Path: append(book(1), spec.Name("author")), Node: "Evelyn Waugh"},
				{Path: append(book(2), spec.Name("author")), Node: "Herman Melville"},
				{Path: append(book(3), spec.Name("author")), Node: "J. R. R. Tolkien"},
			},
		},
		{
			name: "example_2",
			path: `$..author`,
			exp:  []any{"Nigel Rees", "Evelyn Waugh", "Herman Melville", "J. R. R. Tolkien"},
			loc: []*spec.LocatedNode{
				{Path: append(book(0), spec.Name("author")), Node: "Nigel Rees"},
				{Path: append(book(1), spec.Name("author")), Node: "Evelyn Waugh"},
				{Path: append(book(2), spec.Name("author")), Node: "Herman Melville"},
				{Path: append(book(3), spec.Name("author")), Node: "J. R. R. Tolkien"},
			},
		},
		{
			name: "example_3",
			path: `$.store.*`,
			exp:  []any{store["book"], store["bicycle"]},
			loc: []*spec.LocatedNode{
				{Path: spec.NormalizedPath{spec.Name("store"), spec.Name("book")}, Node: store["book"]},
				{Path: spec.NormalizedPath{spec.Name("store"), spec.Name("bicycle")}, Node: store["bicycle"]},
			},
			rand: true,
		},
		{
			name: "example_4",
			path: `$.store..price`,
			exp:  []any{399., 8.95, 12.99, 8.99, 22.99},
			loc: []*spec.LocatedNode{
				{Path: spec.NormalizedPath{spec.Name("store"), spec.Name("bicycle"), spec.Name("price")}, Node: 399.},
				{Path: append(book(0), spec.Name("price")), Node: 8.95},
				{Path: append(book(1), spec.Name("price")), Node: 12.99},
				{Path: append(book(2), spec.Name("price")), Node: 8.99},
				{Path: append(book(3), spec.Name("price")), Node: 22.99},
			},
			rand: true,
		},
		{
			name: "example_5",
			path: `$..book[2]`,
			exp:  []any{books[2]},
			loc:  []*spec.LocatedNode{{Path: book(2), Node: books[2]}},
		},
		{
			name: "example_6",
			path: `$..book[-1]`,
			exp:  []any{books[3]},
			loc:  []*spec.LocatedNode{{Path: book(3), Node: books[3]}},
		},
		{
			name: "example_7",
			path: `$..book[0,1]`,
			exp:  []any{books[0], books[1]},
			loc: []*spec.LocatedNode{
				{Path: book(0), Node: books[0]},
				{Path: book(1), Node: books[1]},
			},
		},
		{
			name: "example_8",
			path: `$..book[?(@.isbn)]`,
			exp:  []any{books[2], books[3]},
			loc: []*spec.LocatedNode{
				{Path: book(2), Node: books[2]},
				{Path: book(3), Node: books[3]},
			},
		},
		{
			name: "example_9",
			path: `$..book[?(@.price<10)]`,
			exp:  []any{books[0], books[2]},
			loc: []*spec.LocatedNode{
				{Path: book(0), Node: books[0]},
				{Path: book(2), Node: books[2]},
			},
		},
		{
			name: "example_10",
			path: `$..*`,
			size: 27,
			rand: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := MustParse(tc.path)
			a.Equal(p.q, p.Query())
			a.Equal(p.q.String(), p.String())
			res := p.Select(val)
			loc := p.SelectLocated(val)

			if tc.exp != nil {
				if tc.rand {
					a.ElementsMatch(tc.exp, res)
					a.ElementsMatch(tc.loc, loc)
				} else {
					a.Equal(tc.exp, res)
					a.Equal(tc.loc, loc)
				}
			} else {
				a.Len(res, tc.size)
				a.Len(loc, tc.size)
			}
		})
	}
}

func specExampleJSON(t *testing.T) map[string]any {
	t.Helper()
	src := []byte(`{
	  "store": {
	    "book": [
	      {
	        "category": "reference",
	        "author": "Nigel Rees",
	        "title": "Sayings of the Century",
	        "price": 8.95
	      },
	      {
	        "category": "fiction",
	        "author": "Evelyn Waugh",
	        "title": "Sword of Honour",
	        "price": 12.99
	      },
	      {
	        "category": "fiction",
	        "author": "Herman Melville",
	        "title": "Moby Dick",
	        "isbn": "0-553-21311-3",
	        "price": 8.99
	      },
	      {
	        "category": "fiction",
	        "author": "J. R. R. Tolkien",
	        "title": "The Lord of the Rings",
	        "isbn": "0-395-19395-8",
	        "price": 22.99
	      }
	    ],
	    "bicycle": {
	      "color": "red",
	      "price": 399
	    }
	  }
	}`)

	var value map[string]any
	if err := json.Unmarshal(src, &value); err != nil {
		t.Fatal(err)
	}

	return value
}

func TestParseCompliance(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)
	p := NewParser()

	//nolint:tagliatelle
	type testCase struct {
		Name            string
		Selector        string
		Document        any
		Result          []any
		Results         [][]any
		InvalidSelector bool `json:"invalid_selector"`
	}

	rawJSON, err := os.ReadFile(
		filepath.Join("jsonpath-compliance-test-suite", "cts.json"),
	)
	r.NoError(err)
	var ts struct{ Tests []testCase }
	//nolint:musttag
	if err := json.Unmarshal(rawJSON, &ts); err != nil {
		t.Fatal(err)
	}

	for i, tc := range ts.Tests {
		t.Run(fmt.Sprintf("test_%03d", i), func(t *testing.T) {
			t.Parallel()
			description := fmt.Sprintf("%v: `%v`", tc.Name, tc.Selector)
			p, err := p.Parse(tc.Selector)
			if tc.InvalidSelector {
				r.Error(err, description)
				r.ErrorIs(err, ErrPathParse)
				a.Nil(p, description)
				return
			}

			r.NoError(err, description)
			a.NotNil(p, description)

			res := p.Select(tc.Document)
			switch {
			case tc.Result != nil:
				a.Equal(tc.Result, res, description)
			case tc.Results != nil:
				a.Contains(tc.Results, res, description)
			}
		})
	}
}

func TestParser(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)
	reg := registry.New()

	for _, tc := range []struct {
		name string
		path string
		reg  *registry.Registry
		exp  *Path
		err  string
	}{
		{
			name: "root",
			path: "$",
			exp:  MustParse("$"),
		},
		{
			name: "root_reg",
			path: "$",
			reg:  reg,
			exp:  MustParse("$"),
		},
		{
			name: "parse_error",
			path: "lol",
			err:  "jsonpath: unexpected identifier at position 1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Construct a parser.
			var parser *Parser
			if tc.reg == nil {
				parser = NewParser()
			} else {
				parser = NewParser(WithRegistry(tc.reg))
				a.Equal(tc.reg, parser.reg)
			}

			// Test Parse and MustParse methods.
			p, err := parser.Parse(tc.path)
			if tc.err == "" {
				r.NoError(err)
				a.Equal(tc.exp, p)
				a.Equal(tc.exp, parser.MustParse(tc.path))
			} else {
				r.EqualError(err, tc.err)
				r.ErrorIs(err, ErrPathParse)
				a.PanicsWithError(tc.err, func() { parser.MustParse(tc.path) })
			}

			if tc.reg == nil {
				// Test Parse and MustParse functions.
				if tc.err == "" {
					r.NoError(err)
					a.Equal(tc.exp, p)
					a.Equal(tc.exp, parser.MustParse(tc.path))
				} else {
					r.EqualError(err, tc.err)
					r.ErrorIs(err, ErrPathParse)
					a.PanicsWithError(tc.err, func() { parser.MustParse(tc.path) })
				}
			}
		})
	}
}
