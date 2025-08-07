package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theory/jsonpath/spec"
)

func TestRegistry(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test  string
		rType spec.FuncType
		expr  []spec.FuncExprArg
		args  []spec.PathValue
		exp   any
	}{
		// RFC 9535-defined functions.
		{
			test:  "length",
			rType: spec.FuncValue,
			expr:  []spec.FuncExprArg{spec.Literal("foo")},
			args:  []spec.PathValue{spec.Value("foo")},
			exp:   spec.Value(3),
		},
		{
			test:  "count",
			rType: spec.FuncValue,
			expr:  []spec.FuncExprArg{&spec.SingularQueryExpr{}},
			args:  []spec.PathValue{spec.Nodes(1, 2)},
			exp:   spec.Value(2),
		},
		{
			test:  "value",
			rType: spec.FuncValue,
			expr:  []spec.FuncExprArg{&spec.SingularQueryExpr{}},
			args:  []spec.PathValue{spec.Nodes(42)},
			exp:   spec.Value(42),
		},
		{
			test:  "match",
			rType: spec.FuncLogical,
			expr:  []spec.FuncExprArg{spec.Literal("foo"), spec.Literal(".*")},
			args:  []spec.PathValue{spec.Value("foo"), spec.Value(".*")},
			exp:   spec.LogicalTrue,
		},
		{
			test:  "search",
			rType: spec.FuncLogical,
			expr:  []spec.FuncExprArg{spec.Literal("foo"), spec.Literal(".")},
			args:  []spec.PathValue{spec.Value("foo"), spec.Value(".")},
			exp:   spec.LogicalTrue,
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			r := require.New(t)

			reg := New()
			a.Len(reg.funcs, 5)

			ft := reg.Get(tc.test)
			a.NotNil(ft)
			a.Equal(tc.rType, ft.ReturnType())
			r.NoError(ft.Validate(tc.expr))
			a.Equal(tc.exp, ft.Evaluate(tc.args))
		})
	}
}

func TestRegisterErr(t *testing.T) {
	t.Parallel()
	reg := New()

	for _, tc := range []struct {
		test   string
		fnName string
		valid  spec.Validator
		eval   spec.Evaluator
		err    string
	}{
		{
			test:   "existing_func",
			fnName: "length",
			valid:  func([]spec.FuncExprArg) error { return nil },
			eval:   func([]spec.PathValue) spec.PathValue { return spec.Value(42) },
			err:    "register: Register called twice for function length",
		},
		{
			test: "nil_validator",
			err:  "register: validator is nil",
		},
		{
			test:  "nil_evaluator",
			valid: func([]spec.FuncExprArg) error { return nil },
			err:   "register: evaluator is nil",
		},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)

			err := reg.Register(tc.fnName, spec.FuncValue, tc.valid, tc.eval)
			r.ErrorIs(err, ErrRegister, tc.test)
			r.EqualError(err, tc.err, tc.test)
		})
	}
}
