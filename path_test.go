package jsonpath

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSpecExamples(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)
	val := specExampleJSON(t)
	store, _ := val["store"].(map[string]any)

	for _, tc := range []struct {
		name string
		path string
		exp  []any
		size int
		rand bool
	}{
		//nolint:dupword
		{
			name: "example_1",
			path: `$.store.book[*].author`,
			exp:  []any{"Nigel Rees", "Evelyn Waugh", "Herman Melville", "J. R. R. Tolkien"},
		},
		//nolint:dupword
		{
			name: "example_2",
			path: `$..author`,
			exp:  []any{"Nigel Rees", "Evelyn Waugh", "Herman Melville", "J. R. R. Tolkien"},
		},
		{
			name: "example_3",
			path: `$.store.*`,
			exp:  []any{store["book"], store["bicycle"]},
			rand: true,
		},
		{
			name: "example_4",
			path: `$.store..price`,
			exp:  []any{399., 8.95, 12.99, 8.99, 22.99},
			rand: true,
		},
		{
			name: "example_5",
			path: `$..book[2]`,
			exp: []any{map[string]any{
				"category": "fiction",
				"author":   "Herman Melville",
				"title":    "Moby Dick",
				"isbn":     "0-553-21311-3",
				"price":    8.99,
			}},
		},
		{
			name: "example_6",
			path: `$..book[-1]`,
			//nolint:dupword
			exp: []any{map[string]any{
				"category": "fiction",
				"author":   "J. R. R. Tolkien",
				"title":    "The Lord of the Rings",
				"isbn":     "0-395-19395-8",
				"price":    22.99,
			}},
		},
		{
			name: "example_7",
			path: `$..book[0,1]`,
			exp: []any{
				map[string]any{
					"category": "reference",
					"author":   "Nigel Rees",
					"title":    "Sayings of the Century",
					"price":    8.95,
				},
				map[string]any{
					"category": "fiction",
					"author":   "Evelyn Waugh",
					"title":    "Sword of Honour",
					"price":    12.99,
				},
			},
		},
		{
			name: "example_8",
			path: `$..book[?(@.isbn)]`,
			exp: []any{
				map[string]any{
					"category": "fiction",
					"author":   "Herman Melville",
					"title":    "Moby Dick",
					"isbn":     "0-553-21311-3",
					"price":    8.99,
				},
				//nolint:dupword
				map[string]any{
					"category": "fiction",
					"author":   "J. R. R. Tolkien",
					"title":    "The Lord of the Rings",
					"isbn":     "0-395-19395-8",
					"price":    22.99,
				},
			},
		},
		{
			name: "example_9",
			path: `$..book[?(@.price<10)]`,
			exp: []any{
				map[string]any{
					"category": "reference",
					"author":   "Nigel Rees",
					"title":    "Sayings of the Century",
					"price":    8.95,
				},
				map[string]any{
					"category": "fiction",
					"author":   "Herman Melville",
					"title":    "Moby Dick",
					"isbn":     "0-553-21311-3",
					"price":    8.99,
				},
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
			p, err := Parse(tc.path)
			r.NoError(err)
			a.Equal(p.q, p.Query())
			a.Equal(p.q.String(), p.String())
			res := p.Select(val)

			if tc.exp != nil {
				if tc.rand {
					a.ElementsMatch(tc.exp, res)
				} else {
					a.Equal(tc.exp, res)
				}
			} else {
				a.Len(res, tc.size)
			}
		})
	}
}

func specExampleJSON(t *testing.T) map[string]any {
	t.Helper()
	//nolint:dupword
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

			// Replace invalid JSON start character in test case 4.
			if i == 4 {
				tc.Selector = strings.ReplaceAll(tc.Selector, "☺", "々")
				if doc, ok := tc.Document.(map[string]any); ok {
					doc["々"] = doc["☺"]
					delete(doc, "☺")
				} else {
					t.Fatalf("expected map[string]any but got %T", tc.Document)
				}
			}

			description := fmt.Sprintf("%v: `%v`", tc.Name, tc.Selector)
			p, err := Parse(tc.Selector)
			if tc.InvalidSelector {
				r.Error(err, description)
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
