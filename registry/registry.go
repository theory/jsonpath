// Package registry provides a RFC 9535 JSONPath function registry.
package registry

//go:generate stringer -linecomment -output registry_string.go -type FuncType

import (
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
				Name:       "length",
				ResultType: spec.FuncValue,
				Validate:   checkLengthArgs,
				Evaluate:   lengthFunc,
			},
			"count": {
				Name:       "count",
				ResultType: spec.FuncValue,
				Validate:   checkCountArgs,
				Evaluate:   countFunc,
			},
			"value": {
				Name:       "value",
				ResultType: spec.FuncValue,
				Validate:   checkValueArgs,
				Evaluate:   valueFunc,
			},
			"match": {
				Name:       "match",
				ResultType: spec.FuncLogical,
				Validate:   checkMatchArgs,
				Evaluate:   matchFunc,
			},
			"search": {
				Name:       "search",
				ResultType: spec.FuncLogical,
				Validate:   checkSearchArgs,
				Evaluate:   searchFunc,
			},
		},
	}
}

// Register registers a function extension by its name. Panics if fn is nil or
// Register is called twice with the same fn.name.
func (r *Registry) Register(fn *Function) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if fn == nil {
		panic("jsonpath: Register function is nil")
	}
	if _, dup := r.funcs[fn.Name]; dup {
		panic("jsonpath: Register called twice for function " + fn.Name)
	}
	r.funcs[fn.Name] = fn
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
	// Name is the name of the function. Must be unique among all functions.
	Name string

	// ResultType defines the type of the function return value.
	ResultType spec.FuncType

	// Validate executes at parse time to validate that all the args to
	// the function are compatible with the function.
	Validate func(args []spec.FunctionExprArg) error

	// Evaluate executes the function against args and returns the result of
	// type ResultType.
	Evaluate func(args []spec.JSONPathValue) spec.JSONPathValue
}
