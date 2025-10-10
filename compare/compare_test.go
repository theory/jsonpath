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
		"array_slice_with_step_and_leading_zeros": "RFC 9535 § 2.3.3.1, 2.3.4.1: leading zeros disallowed in integers",
		"dot_notation_with_number_on_object":      "RFC 9535 § 2.5.1.1: leading digits disallowed in shorthand names",
		"dot_notation_with_dash":                  "RFC 9535 § 2.5.1.1: dash disallowed in shorthand hames",
	}

	for _, q := range queries(t) {
		t.Run(q.ID, func(t *testing.T) {
			t.Parallel()

			// Skip tests with no consensus.
			// https://github.com/cburgmer/json-path-comparison/pull/153#issuecomment-3374075044
			if q.Consensus == nil {
				t.Skip("No consensus")
			}

			// Skip tests where the consensus NOT_SUPPORTED.
			if q.Consensus == "NOT_SUPPORTED" {
				t.Skip(q.Consensus)
			}

			// Skip tests that violate RFC 9535.
			if r, ok := skip[q.ID]; ok {
				t.Skip(r)
			}

			path, err := jsonpath.Parse(q.Selector)
			require.NoError(t, err)
			result := []any(path.Select(q.Document))

			if q.Ordered {
				assert.Equal(t, q.Consensus, result)
			} else {
				assert.ElementsMatch(t, q.Consensus, result)
			}
		})
	}
}
