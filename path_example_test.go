package jsonpath_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/theory/jsonpath"
	"github.com/theory/jsonpath/registry"
	"github.com/theory/jsonpath/spec"
)

// Select all the authors of the books in a bookstore object.
func Example() {
	// Parse a jsonpath query.
	p, err := jsonpath.Parse(`$.store.book[*].author`)
	if err != nil {
		log.Fatal(err)
	}

	// Select values from unmarshaled JSON input.
	store := bookstore()
	result := p.Select(store)

	// Show the result.
	items, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", items)

	// Output: ["Nigel Rees","Evelyn Waugh","Herman Melville","J. R. R. Tolkien"]
}

func ExamplePath_Select() {
	// Load some JSON.
	menu := map[string]any{
		"apps": map[string]any{
			"guacamole": 19.99,
			"salsa":     5.99,
		},
	}

	// Parse a JSONPath and select from the input.
	p := jsonpath.MustParse("$.apps.*")
	result := p.Select(menu)

	// Show the result.
	for _, val := range result {
		fmt.Printf("%v\n", val)
	}
	// Unordered output:
	// 19.99
	// 5.99
}

func ExamplePath_SelectLocated() {
	// Load some JSON.
	menu := map[string]any{
		"apps": map[string]any{
			"guacamole": 19.99,
			"salsa":     5.99,
		},
	}

	// Parse a JSONPath and select from the input.
	p := jsonpath.MustParse("$.apps.*")
	result := p.SelectLocated(menu)

	// Show the result.
	for _, node := range result {
		fmt.Printf("%v: %v\n", node.Path, node.Node)
	}

	// Unordered output:
	// $['apps']['guacamole']: 19.99
	// $['apps']['salsa']: 5.99
}

// Use the Parser to parse a collection of paths.
func ExampleParser() {
	// Create a new parser using the default function registry.
	parser := jsonpath.NewParser()

	// Parse a list of paths.
	paths := []*jsonpath.Path{}
	for _, path := range []string{
		"$.store.book[*].author",
		"$..author",
		"$.store..color",
		"$..book[2].author",
		"$..book[2].publisher",
		"$..book[?@.isbn].title",
		"$..book[?@.price<10].title",
	} {
		p, err := parser.Parse(path)
		if err != nil {
			log.Fatal(err)
		}
		paths = append(paths, p)
	}

	// Later, use the paths to select from JSON inputs.
	store := bookstore()
	for _, p := range paths {
		items := p.Select(store)
		array, err := json.Marshal(items)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s\n", array)
	}
	// Output:
	// ["Nigel Rees","Evelyn Waugh","Herman Melville","J. R. R. Tolkien"]
	// ["Nigel Rees","Evelyn Waugh","Herman Melville","J. R. R. Tolkien"]
	// ["red"]
	// ["Herman Melville"]
	// []
	// ["Moby Dick","The Lord of the Rings"]
	// ["Sayings of the Century","Moby Dick"]
}

// The second use case for the Parser is to provide a [registry.Registry]
// containing function extensions, as [defined by the standard]. This example
// creates a function named "first" that returns the first item in a list of
// nodes.
//
// [defined by the standard]: https://www.rfc-editor.org/rfc/rfc9535.html#name-function-extensions
func ExampleParser_functionExtension() {
	// Register the first function.
	reg := registry.New()
	err := reg.Register(
		"first",           // name
		spec.FuncValue,    // returns a single value
		validateFirstArgs, // parse-time validation defined below
		firstFunc,         // function defined below
	)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	// Create a parser with the registry that contains the extension.
	parser := jsonpath.NewParser(jsonpath.WithRegistry(reg))

	// Use the function to select lists that start with 6.
	path, err := parser.Parse("$[? first(@.*) == 6]")
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	// Do any of these arrays start with 6?
	input := []any{
		[]any{1, 2, 3, 4, 5},
		[]any{6, 7, 8, 9},
		[]any{4, 8, 12},
	}
	result := path.Select(input)
	fmt.Printf("%v\n", result)
	// Output: [[6 7 8 9]]
}

// validateFirstArgs validates that a single argument is passed to the first()
// function, and that it can be converted to [spec.PathNodes], so that first()
// can return the first node. It's called by the parser.
func validateFirstArgs(fea []spec.FunctionExprArg) error {
	if len(fea) != 1 {
		return fmt.Errorf("expected 1 argument but found %v", len(fea))
	}

	if !fea[0].ResultType().ConvertsTo(spec.PathNodes) {
		return errors.New("cannot convert argument to PathNodes")
	}

	return nil
}

// firstFunc defines the custom first() JSONPath function. It converts its
// single argument to a [spec.NodesType] value and returns a [*spec.ValueType]
// that contains the first node. If there are no nodes it returns nil.
func firstFunc(jv []spec.JSONPathValue) spec.JSONPathValue {
	nodes := spec.NodesFrom(jv[0])
	if len(nodes) == 0 {
		return nil
	}
	return spec.Value(nodes[0])
}

// bookstore returns an unmarshaled JSON object.
func bookstore() any {
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

	// Parse the JSON.
	var value any
	if err := json.Unmarshal(src, &value); err != nil {
		log.Fatal(err)
	}
	return value
}
