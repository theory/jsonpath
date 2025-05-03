package registry_test

import (
	"errors"
	"fmt"
	"log"

	"github.com/theory/jsonpath/registry"
	"github.com/theory/jsonpath/spec"
)

// Create and registry a custom JSONPath expression, first(), that returns the
// first node in a list of nodes passed to it. See
// [github.com/theory/jsonpath.Parser] for a more complete example.
func Example() {
	reg := registry.New()
	err := reg.Register(
		"first",           // function name
		spec.FuncValue,    // returns a single value
		validateFirstArgs, // parse-time validation defined below
		firstFunc,         // function defined below
	)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	fmt.Printf("%v\n", reg.Get("first").ResultType())
	// Output: Value
}

// validateFirstArgs validates that a single argument is passed to the first()
// function, and that it can be converted to [spec.FuncNodes], so that first()
// can return the first node. It's called by the parser.
func validateFirstArgs(fea []spec.FuncExprArg) error {
	if len(fea) != 1 {
		return fmt.Errorf("expected 1 argument but found %v", len(fea))
	}

	if !fea[0].ResultType().ConvertsToNodes() {
		return errors.New("cannot convert argument to Nodes")
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
