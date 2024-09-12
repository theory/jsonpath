package registry_test

import (
	"errors"
	"fmt"

	"github.com/theory/jsonpath/registry"
	"github.com/theory/jsonpath/spec"
)

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

// Create and registry a custom JSONPath expression, first(), that returns the
// first node in a list of nodes passed to it.
func Example() {
	reg := registry.New()
	reg.Register(&registry.Function{
		Name:       "first",
		ResultType: spec.FuncValue,
		Validate:   validateFirstArgs,
		Evaluate:   firstFunc,
	})
	fmt.Printf("%v\n", reg.Get("first").ResultType)
	// Output:FuncValue
}
