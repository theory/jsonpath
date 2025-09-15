//go:build compare

// Package compare  tests theory/jsonpath against the [json-path-comparison]
// project's regression suite. It requires the file regression_suite.yaml to
// be in this directory. The test only runs with the "compare" tag. Use make
// for the easiest way to download regression_suite.yaml and run the tests:
//
//	make test-compare
//
// [json-path-comparison]: https://github.com/cburgmer/json-path-comparison
package compare

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theory/jsonpath"
	"gopkg.in/yaml.v3"
)

type Query struct {
	ID        string `yaml:"id"`
	Selector  string `yaml:"selector"`
	Document  any    `yaml:"document"`
	Consensus any    `yaml:"consensus"`
	Ordered   bool   `yaml:"ordered"`
}

func file(t *testing.T) string {
	t.Helper()
	_, fn, _, ok := runtime.Caller(0)
	assert.True(t, ok)
	return filepath.Clean(filepath.Join(
		filepath.Dir(fn),
		"regression_suite.yaml",
	))
}

func queries(t *testing.T) []Query {
	t.Helper()
	data, err := os.ReadFile(file(t))
	require.NoError(t, err)
	var q struct {
		Queries []Query `yaml:"queries"`
	}
	require.NoError(t, yaml.Unmarshal(data, &q))
	return q.Queries
}

func TestConsensus(t *testing.T) {
	t.Parallel()

	skip := map[string]string{
		"array_slice_with_step_and_leading_zeros": "leading zeros disallowed integers; see RFC 9535 sections 2.3.3.1, 2.3.4.1",
		"dot_notation_with_number_on_object":      "leading digits disallowed in shorthand names; see RFC 9535 section 2.5.1.1",
		"dot_notation_with_dash":                  "dash disallowed in shorthand hames; see RFC 9535 section 2.5.1.1",
	}

	for _, q := range queries(t) {
		t.Run(q.ID, func(t *testing.T) {
			t.Parallel()

			if q.Consensus == "NOT_SUPPORTED" {
				t.Skip(q.Consensus)
			}

			if r, ok := skip[q.ID]; ok {
				t.Skip(r)
			}

			path, err := jsonpath.Parse(q.Selector)
			// XXX Why is consensus empty?
			if q.Consensus != nil {
				require.NoError(t, err)
			}
			if err != nil {
				// XXX Why is there an error? TODOs?
				assert.Nil(t, path)
				return
			}
			result := []any(path.Select(q.Document))

			switch {
			case q.Ordered:
				assert.Equal(t, q.Consensus, result)
			case q.Consensus == nil:
				// XXX What to do here?
				// assert.Empty(t, result)
			default:
				assert.ElementsMatch(t, q.Consensus, result)
			}
		})
	}
}
