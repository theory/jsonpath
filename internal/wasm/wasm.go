// Package main performs a basic JSONPath query in order to test WASM compilation.
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/theory/jsonpath"
)

func main() {
	// Parse a jsonpath query.
	p, err := jsonpath.Parse(`$.foo`)
	if err != nil {
		log.Fatal(err)
	}

	// Select values from unmarshaled JSON input.
	result := p.Select([]byte(`{"foo": "bar"}`))

	// Show the result.
	items, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	//nolint:forbidigo
	fmt.Printf("%s\n", items)
}
