// package main provides the Wasm app.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/theory/jsonpath"
)

const (
	optIndent int = 1 << iota
)

func main() {
	stream := make(chan struct{})

	js.Global().Set("query", js.FuncOf(query))
	js.Global().Set("optIndent", js.ValueOf(optIndent))

	<-stream
}

func query(_ js.Value, args []js.Value) any {
	query := args[0].String()
	target := args[1].String()
	opts := args[2].Int()

	return execute(query, target, opts)
}

func execute(query, target string, opts int) string {
	// Parse the JSON.
	var value any
	if err := json.Unmarshal([]byte(target), &value); err != nil {
		return fmt.Sprintf("Error parsing JSON: %v", err)
	}

	// Parse the JSONPath query.
	path, err := jsonpath.Parse(query)
	if err != nil {
		return fmt.Sprintf("Error parsing %v", err)
	}

	// Execute the query against the JSON.
	res := path.Select(value)

	// Serialize the result
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if opts&optIndent == optIndent {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(res); err != nil {
		return fmt.Sprintf("Error parsing results: %v", err)
	}

	return buf.String()
}
