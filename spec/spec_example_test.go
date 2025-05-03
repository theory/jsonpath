package spec_test

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/theory/jsonpath"
	"github.com/theory/jsonpath/registry"
	"github.com/theory/jsonpath/spec"
)

// Select all the authors of the books in a bookstore object.
func Example() {
	// Construct a jsonpath query.
	p := jsonpath.New(spec.Query(
		true,
		spec.Child(spec.Name("store")),
		spec.Child(spec.Name("book")),
		spec.Child(spec.Wildcard),
		spec.Child(spec.Name("author")),
	))

	// Select values from unmarshaled JSON input.
	store := bookstore()
	nodes := p.Select(store)

	// Show the selected values.
	items, err := json.Marshal(nodes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", items)

	// Output: ["Nigel Rees","Evelyn Waugh","Herman Melville","J. R. R. Tolkien"]
}

// Construct a couple of different path queries.
func ExamplePathQuery() {
	// Create a query and its segments.
	q := spec.Query(
		true,
		spec.Child(spec.Name("store")),
		spec.Child(spec.Name("book")),
		spec.Child(spec.Wildcard),
		spec.Child(spec.Name("author")),
	)
	fmt.Printf("%v\n", q)

	// Create a query with multi-selector segments.
	q = spec.Query(
		true,
		spec.Child(spec.Name("profile")),
		spec.Descendant(spec.Name("email"), spec.Name("phone")),
	)
	fmt.Printf("%v\n", q)

	// Output:
	// $["store"]["book"][*]["author"]
	// $["profile"]..["email","phone"]
}

// Create a child segment that selects the name "hi", th index 2, or slice
// 1-3.
func ExampleSegment_child() {
	child := spec.Child(
		spec.Name("hi"),
		spec.Index(2),
		spec.Slice(1, 3, 1),
	)
	fmt.Printf("%v\n", child)
	// Output: ["hi",2,1:3]
}

// Create a descendant segment that selects the name "email" or array index
// zero from a node or any of its descendant nodes.
func ExampleSegment_descendant() {
	child := spec.Descendant(
		spec.Name("email"),
		spec.Index(0),
	)
	fmt.Printf("%v\n", child)
	// Output: ..["email",0]
}

// Create a few slice selectors.
func ExampleSliceSelector() {
	// Select all values in a slice.
	for _, sliceSelector := range []any{
		spec.Slice(),         // full slice
		spec.Slice(1, 4),     // items 1-3
		spec.Slice(nil, 8),   // items 0-7
		spec.Slice(4, -1, 3), // items 4-last step by 3
		spec.Slice(5, 1, -2), // items 5-2, step by -2

	} {
		fmt.Printf("%v\n", sliceSelector)
	}
	// Output:
	// :
	// 1:4
	// :8
	// 4:-1:3
	// 5:1:-2
}

// Create a few name selectors.
func ExampleName() {
	for _, nameSelector := range []any{
		spec.Name("hello"),                 // ascii
		spec.Name(`Charlie "Bird" Parker`), // quoted
		spec.Name("އައްސަލާމް ޢަލައިކުމް"), // Unicode
		spec.Name("📡🪛🪤"),                   // emoji
	} {
		fmt.Printf("%v\n", nameSelector)
	}
	// Output:
	// "hello"
	// "Charlie \"Bird\" Parker"
	// "އައްސަލާމް ޢަލައިކުމް"
	// "📡🪛🪤"
}

// Create a few index selectors.
func ExampleIndex() {
	for _, indexSelector := range []any{
		spec.Index(0),  // first item
		spec.Index(3),  // fourth item
		spec.Index(-1), // last item
	} {
		fmt.Printf("%v\n", indexSelector)
	}
	// Output:
	// 0
	// 3
	// -1
}

func ExampleWildcard() {
	fmt.Printf("%v\n", spec.Wildcard)
	// Output: *
}

// Create a filter selector that selects nodes with a descendant that contains
// an object field named "x".
func ExampleFilterSelector() {
	f := spec.Filter(spec.And(
		spec.Existence(
			spec.Query(false, spec.Descendant(spec.Name("x"))),
		),
	))
	fmt.Printf("%v\n", f)
	// Output: ?@..["x"]
}

// Create a comparison expression that compares a literal value to a path
// expression.
func ExampleComparisonExpr() {
	cmp := spec.Comparison(
		spec.Literal(42),
		spec.EqualTo,
		spec.SingularQuery(false, spec.Name("age")),
	)
	fmt.Printf("%v\n", cmp)
	// Output: 42 == @["age"]
}

// Create an existence expression as a filter expression.
func ExampleExistExpr() {
	filter := spec.Filter(spec.And(
		spec.Existence(
			spec.Query(false, spec.Child(spec.Name("x"))),
		),
	))
	fmt.Printf("%v\n", filter)
	// Output: ?@["x"]
}

// Create a nonexistence expression as a filter expression.
func ExampleNonExistExpr() {
	filter := spec.Filter(spec.And(
		spec.Nonexistence(
			spec.Query(false, spec.Child(spec.Name("x"))),
		),
	))
	fmt.Printf("%v\n", filter)
	// Output: ?!@["x"]
}

// Construct a path expression as a function argument.
func ExampleNodesQueryExpr() {
	reg := registry.New()
	fnExpr := spec.Function(
		reg.Get("length"),
		spec.NodesQuery(
			spec.Query(false, spec.Child(spec.Index(0))),
		),
	)
	fmt.Printf("%v\n", fnExpr)
	// Output: length(@[0])
}

// Each [spec.FuncType] converted to one of the spec-standard function types,
// [spec.ValueType], [spec.NodesType], or [spec.LogicalType].
func ExampleFuncType() {
	fmt.Println("FuncType       Converts To FuncTypes")
	fmt.Println("-------------- ---------------------")
	for _, at := range []spec.FuncType{
		spec.FuncLiteral,
		spec.FuncSingularQuery,
		spec.FuncValue,
		spec.FuncNodes,
		spec.FuncLogical,
	} {
		to := []string{}
		if at.ConvertsToValue() {
			to = append(to, "Value")
		}
		if at.ConvertsToLogical() {
			to = append(to, "Logical")
		}
		if at.ConvertsToNodes() {
			to = append(to, "Nodes")
		}

		fmt.Printf("%-15v %v\n", at, strings.Join(to, ", "))
	}
	// Output:
	// FuncType       Converts To FuncTypes
	// -------------- ---------------------
	// Literal         Value
	// SingularQuery   Value, Logical, Nodes
	// Value           Value
	// Nodes           Logical, Nodes
	// Logical         Logical
}

// Print the [spec.FuncType] for each [spec.JSONPathValue]
// implementation.
func ExampleJSONPathValue() {
	fmt.Printf("Implementation    FuncType\n")
	fmt.Printf("----------------- --------\n")
	for _, jv := range []spec.JSONPathValue{
		spec.Value(nil),
		spec.Nodes(1, 2),
		spec.Logical(true),
	} {
		fmt.Printf("%-17T %v\n", jv, jv.FuncType())
	}
	// Output:
	// Implementation    FuncType
	// ----------------- --------
	// *spec.ValueType   Value
	// spec.NodesType    Nodes
	// spec.LogicalType  Logical
}

// Use the standard match() function in a function expression.
func ExampleFuncExpr() {
	reg := registry.New()
	fe := spec.Function(
		reg.Get("match"),
		spec.NodesQuery(
			spec.Query(false, spec.Child(spec.Name("rating"))),
		),
		spec.Literal("good$"),
	)
	fmt.Printf("%v\n", fe)
	// Output: match(@["rating"], "good$")
}

// Use the standard count() function in a function expression.
func ExampleNotFuncExpr() {
	reg := registry.New()
	nf := spec.NotFunction(spec.Function(
		reg.Get("length"),
		spec.NodesQuery(
			spec.Query(false, spec.Child(spec.Index(0))),
		),
	))
	fmt.Printf("%v\n", nf)
	// Output: !length(@[0])
}

// Pass a LiteralArg to a function.
func ExampleLiteralArg() {
	reg := registry.New()
	fe := spec.Function(
		reg.Get("length"),
		spec.Literal("some string"),
	)
	fmt.Printf("%v\n", fe)
	// Output: length("some string")
}

// Assemble a LogicalAnd consisting of multiple expressions.
func ExampleLogicalAnd() {
	and := spec.And(
		spec.Value(42),
		spec.Existence(
			spec.Query(false, spec.Child(spec.Name("answer"))),
		),
	)
	fmt.Printf("%v\n", and)
	// Output: 42 && @["answer"]
}

// Assemble a LogicalOr consisting of multiple expressions.
func ExampleLogicalOr() {
	and := spec.Or(
		spec.And(spec.Existence(
			spec.Query(false, spec.Child(spec.Name("answer"))),
		)),
		spec.And(spec.Value(42)),
	)
	fmt.Printf("%v\n", and)
	// Output: @["answer"] || 42
}

func ExampleLogicalType() {
	fmt.Printf("%v\n", spec.Logical(true))
	fmt.Printf("%v\n", spec.Logical(false))
	// Output:
	// true
	// false
}

// Use a [spec.NodesType] to create a list of nodes, a.k.a. a JSON array.
func ExampleNodesType() {
	fmt.Printf("%v\n", spec.Nodes("hi", 42, true))
	// Output: [hi 42 true]
}

// Use a [spec.ParenExpr] to group the result of a [LogicalOr] expression.
func ExampleParenExpr() {
	paren := spec.Paren(
		spec.And(
			spec.Existence(
				spec.Query(false, spec.Child(spec.Name("answer"))),
			),
		),
		spec.And(
			spec.Existence(
				spec.Query(false, spec.Child(spec.Name("question"))),
			),
		),
	)
	fmt.Printf("%v\n", paren)
	// Output: (@["answer"] || @["question"])
}

// Use a [spec.NotParenExpr] to negate the result of a [LogicalOr] expression.
func ExampleNotParenExpr() {
	not := spec.NotParen(
		spec.And(
			spec.Existence(
				spec.Query(false, spec.Child(spec.Name("answer"))),
			),
		),
		spec.And(
			spec.Existence(
				spec.Query(false, spec.Child(spec.Name("question"))),
			),
		),
	)
	fmt.Printf("%v\n", not)
	// Output: !(@["answer"] || @["question"])
}

// Compare a normalized JSONPath and its JSON Pointer equivalent.
func ExampleNormalizedPath() {
	norm := spec.Normalized(
		spec.Name("x"),
		spec.Index(1),
		spec.Name("y"),
	)
	fmt.Printf("%v\n", norm)
	fmt.Printf("%v\n", norm.Pointer())
	// Output:
	// $['x'][1]['y']
	// /x/1/y
}

func ExampleSingularQueryExpr() {
	singular := spec.SingularQuery(
		true,
		spec.Name("profile"),
		spec.Name("contacts"),
		spec.Index(0),
		spec.Name("email"),
	)
	fmt.Printf("%v\n", singular)
	// Output: $["profile"]["contacts"][0]["email"]
}

// Create A [spec.ValueType] for each supported JSON type.
func ExampleValueType() {
	for _, val := range []any{
		"hello",
		42,
		98.6,
		json.Number("1024"),
		true,
		nil,
		map[string]any{"x": true},
		[]any{1, 2, false},
	} {
		fmt.Printf("%v\n", spec.Value(val))
	}
	// Output:
	// hello
	// 42
	// 98.6
	// 1024
	// true
	// <nil>
	// map[x:true]
	// [1 2 false]
}

func ExampleValueFrom() {
	for _, val := range []spec.JSONPathValue{
		spec.Value("hello"), // converts to value
		spec.Nodes(1, 2, 3), // does not convert to value
		spec.Logical(false), // does not convert to value
	} {
		if val.FuncType().ConvertsToValue() {
			fmt.Printf("val: %v\n", spec.ValueFrom(val))
		}
	}
	// Output: val: hello
}

func ExampleNodesFrom() {
	for _, val := range []spec.JSONPathValue{
		spec.Value("hello"), // does not convert to nodes
		spec.Nodes(1, 2, 3), // converts to nodes
		spec.Logical(false), // does not convert to nodes
	} {
		if val.FuncType().ConvertsToNodes() {
			fmt.Printf("nodes: %v\n", spec.NodesFrom(val))
		}
	}
	// Output: nodes: [1 2 3]
}

func ExampleLogicalFrom() {
	for _, val := range []spec.JSONPathValue{
		spec.Value("hello"), // does not convert to logical
		spec.Nodes(1, 2, 3), // converts to logical (existence)
		spec.Logical(false), // converts to logical
	} {
		if val.FuncType().ConvertsToLogical() {
			fmt.Printf("logical: %v\n", spec.LogicalFrom(val))
		}
	}
	// Output:
	// logical: true
	// logical: false
}

func ExampleLocatedNode() {
	// Query all "author" object properties.
	p := jsonpath.New(spec.Query(
		true,
		spec.Descendant(spec.Name("author")),
	))

	// Select LocatedNodes from unmarshaled JSON input.
	store := bookstore()
	nodes := p.SelectLocated(store)

	// Show the LocatedNodes.
	items, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", items)

	// Output:
	// [
	//   {
	//     "node": "Nigel Rees",
	//     "path": "$['store']['book'][0]['author']"
	//   },
	//   {
	//     "node": "Evelyn Waugh",
	//     "path": "$['store']['book'][1]['author']"
	//   },
	//   {
	//     "node": "Herman Melville",
	//     "path": "$['store']['book'][2]['author']"
	//   },
	//   {
	//     "node": "J. R. R. Tolkien",
	//     "path": "$['store']['book'][3]['author']"
	//   }
	// ]
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
