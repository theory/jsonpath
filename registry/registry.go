// Package registry provides a RFC 9535 JSONPath function registry.
package registry

//go:generate stringer -linecomment -output registry_string.go -type FuncType

import (
	"errors"
	"fmt"
	"sync"

	"github.com/theory/jsonpath/spec"
)

// Registry maintains a registry of JSONPath functions, including both
// [RFC 9535]-required functions and function extensions.
//
// [RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535.html
type Registry struct {
	mu    sync.RWMutex
	funcs map[string]*Function
}

// New returns a new [Registry] loaded with the [RFC 9535]-mandated functions:
//
//   - [length]
//   - [count]
//   - [value]
//   - [match]
//   - [search]
//
// [RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535.html
// [length]: https://www.rfc-editor.org/rfc/rfc9535.html#name-length-function-extension
// [count]: https://www.rfc-editor.org/rfc/rfc9535.html#name-count-function-extension
// [value]: https://www.rfc-editor.org/rfc/rfc9535.html#name-value-function-extension
// [match]: https://www.rfc-editor.org/rfc/rfc9535.html#name-match-function-extension
// [search]: https://www.rfc-editor.org/rfc/rfc9535.html#name-search-function-extension
func New() *Registry {
	return &Registry{
		mu: sync.RWMutex{},
		funcs: map[string]*Function{
			"length": {
				name:       "length",
				resultType: spec.FuncValue,
				validator:  checkLengthArgs,
				evaluator:  lengthFunc,
			},
			"count": {
				name:       "count",
				resultType: spec.FuncValue,
				validator:  checkCountArgs,
				evaluator:  countFunc,
			},
			"value": {
				name:       "value",
				resultType: spec.FuncValue,
				validator:  checkValueArgs,
				evaluator:  valueFunc,
			},
			"match": {
				name:       "match",
				resultType: spec.FuncLogical,
				validator:  checkMatchArgs,
				evaluator:  matchFunc,
			},
			"search": {
				name:       "search",
				resultType: spec.FuncLogical,
				validator:  checkSearchArgs,
				evaluator:  searchFunc,
			},
		},
	}
}

// Validator functions validate that the args expressions to a function can be
// processed by the function.
type Validator func(args []spec.FunctionExprArg) error

// Evaluator functions execute a function against the values returned by args.
type Evaluator func(args []spec.JSONPathValue) spec.JSONPathValue

// ErrRegister errors are returned by [Register].
var ErrRegister = errors.New("register")

// Register registers a function extension by its name. Returns an
// [ErrRegister] error if validator or nil or evaluator is nil or if r
// already contains name.
func (r *Registry) Register(
	name string,
	resultType spec.FuncType,
	validator Validator,
	evaluator Evaluator,
) error {
	if validator == nil {
		return fmt.Errorf("%w: validator is nil", ErrRegister)
	}
	if evaluator == nil {
		return fmt.Errorf("%w: evaluator is nil", ErrRegister)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.funcs[name]; dup {
		return fmt.Errorf(
			"%w: Register called twice for function %v",
			ErrRegister, name,
		)
	}

	r.funcs[name] = &Function{name, resultType, validator, evaluator}
	return nil
}

// Get returns a reference to the registered function named name. Returns nil
// if no function with that name has been registered.
func (r *Registry) Get(name string) *Function {
	r.mu.RLock()
	defer r.mu.RUnlock()
	function := r.funcs[name]
	return function
}

// Function defines a JSONPath function. Use [Register] to register a new
// function.
type Function struct {
	// name is the name of the function. Must be unique among all functions in
	// a registry.
	name string

	// resultType defines the type of the function return value.
	resultType spec.FuncType

	// validator executes at parse time to validate that all the args to
	// the function are compatible with the function.
	validator func(args []spec.FunctionExprArg) error

	// evaluator executes the function against args and returns the result of
	// type ResultType.
	evaluator func(args []spec.JSONPathValue) spec.JSONPathValue
}

// NewFunction creates a new JSONPath function extension. The parameters are:
//
//   - name: the name of the function as used in JSONPath queries.
//   - resultType: The data type of the function return value.
//   - validator: A validation function that will be called by at parse time
//     to validate that all the function args are compatible with the function.
//   - evaluator: The implementation of the function itself that executes the
//     against args and returns the result defined by resultType.
func NewFunction(
	name string,
	resultType spec.FuncType,
	validator func(args []spec.FunctionExprArg) error,
	evaluator func(args []spec.JSONPathValue,
	) spec.JSONPathValue,
) *Function {
	return &Function{name, resultType, validator, evaluator}
}

// Name returns the name of the function.
func (f *Function) Name() string { return f.name }

// ResultType returns the data type of the function return value.
func (f *Function) ResultType() spec.FuncType { return f.resultType }

// Evaluate executes the function against args and returns the result of type
// [ResultType].
func (f *Function) Evaluate(args []spec.JSONPathValue) spec.JSONPathValue {
	return f.evaluator(args)
}

// Validate executes at parse time to validate that all the args to the
// function are compatible with the function.
func (f *Function) Validate(args []spec.FunctionExprArg) error {
	return f.validator((args))
}

// type Function interface {
// 	Evaluate(args []JSONPathValue) JSONPathValue
// 	ResultType() FuncType
// 	Name() string
// }
