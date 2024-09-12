package registry

import (
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
			a.Equal(tc.rType, ft.ResultType)
			r.NoError(ft.Validate(tc.expr))
			a.Equal(tc.exp, ft.Evaluate(tc.args))
		})
	}
}

func TestRegisterErr(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	reg := New()

	for _, tc := range []struct {
		name string
		fn   *Function
		err  string
	}{
		{
			name: "nil_func",
			fn:   nil,
			err:  "jsonpath: Register function is nil",
		},
		{
			name: "existing_func",
			fn:   &Function{Name: "length"},
			err:  "jsonpath: Register called twice for function length",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.PanicsWithValue(tc.err, func() { reg.Register(tc.fn) })
		})
	}
}

func newFuncExpr(t *testing.T, name string, args []spec.FunctionExprArg) *spec.FunctionExpr {
	t.Helper()
	f, err := spec.NewFunctionExpr(name, args)
	if err != nil {
		t.Fatal(err.Error())
	}
	return f
}
