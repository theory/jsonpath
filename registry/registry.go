// Package registry provides a RFC 9535 JSONPath function extension registry.
package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/theory/jsonpath/spec"
)

// Registry maintains a registry of JSONPath function extensions, including
// both [RFC 9535]-required functions and custom functions.
//
// [RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535.html
type Registry struct {
	mu    sync.RWMutex
	funcs map[string]*spec.FuncExtension
}

// New returns a new [Registry] loaded with the [RFC 9535]-mandated function
// extensions:
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
		funcs: map[string]*spec.FuncExtension{
			"length": spec.Extension("length", spec.FuncValue, checkLengthArgs, lengthFunc),
			"count":  spec.Extension("count", spec.FuncValue, checkCountArgs, countFunc),
			"value":  spec.Extension("value", spec.FuncValue, checkValueArgs, valueFunc),
			"match":  spec.Extension("match", spec.FuncLogical, checkMatchArgs, matchFunc),
			"search": spec.Extension("search", spec.FuncLogical, checkSearchArgs, searchFunc),
		},
	}
}

// ErrRegister errors are returned by [Register].
var ErrRegister = errors.New("register")

// Register registers a function extension. The parameters are:
//
//   - name: the name of the function extension as used in JSONPath queries.
//   - returnType: The data type of the function return value.
//   - validator: A validation function that will be called at parse time
//     to validate that all the function args are compatible with the function.
//   - evaluator: The implementation of the function itself that executes
//     against args and returns the result of the type defined by resultType.
//
// Returns [ErrRegister] if validator or evaluator is nil or if r already
// contains name.
func (r *Registry) Register(
	name string,
	resultType spec.FuncType,
	validator spec.Validator,
	evaluator spec.Evaluator,
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

	r.funcs[name] = spec.Extension(name, resultType, validator, evaluator)
	return nil
}

// Get returns a reference to the registered function extension named name.
// Returns nil if no function with that name has been registered.
func (r *Registry) Get(name string) *spec.FuncExtension {
	r.mu.RLock()
	defer r.mu.RUnlock()
	function := r.funcs[name]
	return function
}
