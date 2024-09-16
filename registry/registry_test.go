package registry

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theory/jsonpath/spec"
)

func TestRegistry(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)

	for _, tc := range []struct {
		name  string
		rType spec.FuncType
		expr  []spec.FunctionExprArg
		args  []spec.JSONPathValue
		exp   any
	}{
		// RFC 9535-defined functions.
		{
			name:  "length",
			rType: spec.FuncValue,
			expr:  []spec.FunctionExprArg{spec.Literal("foo")},
			args:  []spec.JSONPathValue{spec.Value("foo")},
			exp:   spec.Value(3),
		},
		{
			name:  "count",
			rType: spec.FuncValue,
			expr:  []spec.FunctionExprArg{&spec.SingularQueryExpr{}},
			args:  []spec.JSONPathValue{spec.NodesType([]any{1, 2})},
			exp:   spec.Value(2),
		},
		{
			name:  "value",
			rType: spec.FuncValue,
			expr:  []spec.FunctionExprArg{&spec.SingularQueryExpr{}},
			args:  []spec.JSONPathValue{spec.NodesType([]any{42})},
			exp:   spec.Value(42),
		},
		{
			name:  "match",
			rType: spec.FuncLogical,
			expr:  []spec.FunctionExprArg{spec.Literal("foo"), spec.Literal(".*")},
			args:  []spec.JSONPathValue{spec.Value("foo"), spec.Value(".*")},
			exp:   spec.LogicalTrue,
		},
		{
			name:  "search",
			rType: spec.FuncLogical,
			expr:  []spec.FunctionExprArg{spec.Literal("foo"), spec.Literal(".")},
			args:  []spec.JSONPathValue{spec.Value("foo"), spec.Value(".")},
			exp:   spec.LogicalTrue,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			reg := New()
			a.Len(reg.funcs, 5)

			ft := reg.Get(tc.name)
			a.NotNil(ft)
			a.Equal(tc.rType, ft.resultType)
			r.NoError(ft.validator(tc.expr))
			a.Equal(tc.exp, ft.evaluator(tc.args))
		})
	}
}

func TestRegisterErr(t *testing.T) {
	t.Parallel()
	r := require.New(t)
	reg := New()

	for _, tc := range []struct {
		name   string
		fnName string
		valid  Validator
		eval   Evaluator
		err    string
	}{
		{
			name: "nil_validator",
			err:  "register: validator is nil",
		},
		{
			name:  "nil_evaluator",
			valid: func([]spec.FunctionExprArg) error { return nil },
			err:   "register: evaluator is nil",
		},
		{
			name:   "existing_func",
			fnName: "length",
			valid:  func([]spec.FunctionExprArg) error { return nil },
			eval:   func([]spec.JSONPathValue) spec.JSONPathValue { return spec.Value(42) },
			err:    "register: Register called twice for function length",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := reg.Register(tc.fnName, spec.FuncValue, tc.valid, tc.eval)
			r.ErrorIs(err, ErrRegister, tc.name)
			r.EqualError(err, tc.err, tc.name)
		})
	}
}

func TestFunction(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name string
		fn   *Function
		args []spec.JSONPathValue
		err  error
		exp  spec.JSONPathValue
	}{
		{
			name: "valid_err_value",
			fn: NewFunction(
				"xyz", spec.FuncValue,
				func([]spec.FunctionExprArg) error { return errors.New("oops") },
				func([]spec.JSONPathValue) spec.JSONPathValue { return spec.Value(42) },
			),
			args: []spec.JSONPathValue{},
			exp:  spec.Value(42),
			err:  errors.New("oops"),
		},
		{
			name: "no_valid_err_nodes",
			fn: NewFunction(
				"abc", spec.FuncNodeList,
				func([]spec.FunctionExprArg) error { return nil },
				func([]spec.JSONPathValue) spec.JSONPathValue { return spec.NodesType{"hi"} },
			),
			args: []spec.JSONPathValue{},
			exp:  spec.NodesType{"hi"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.fn.name, tc.fn.Name())
			a.Equal(tc.err, tc.fn.Validate(nil))
			a.Equal(tc.exp, tc.fn.Evaluate(tc.args))
		})
	}
}
