// Package main performs a basic JSONPath query in order to test WASM compilation.
package main

import (
	"encoding/json"
	"fmt"

	"github.com/theory/jsonpath"
)

func main() {
	// Parse a jsonpath query.
	p, _ := jsonpath.Parse(`$.foo`)

	// Select values from unmarshaled JSON input.
	result := p.Select([]byte(`{"foo": "bar"}`))

	// Show the result.
	//nolint:errchkjson
	items, _ := json.Marshal(result)

	//nolint:forbidigo
	fmt.Printf("%s\n", items)
}
