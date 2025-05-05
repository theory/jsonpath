package spec

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompOp(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		op  CompOp
		str string
	}{
		{EqualTo, "=="},
		{NotEqualTo, "!="},
		{LessThan, "<"},
		{LessThanEqualTo, "<="},
		{GreaterThan, ">"},
		{GreaterThanEqualTo, ">="},
	} {
		a.Equal(tc.str, tc.op.String())
	}
}

func TestEqualTo(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name  string
		left  any
		right any
		exp   bool
	}{
		{"int_zeros", 0, 0, true},
		{"int_ones", 1, 1, true},
		{"int_zero_one", 0, 1, false},
		{"int8_zeros", int8(0), int8(0), true},
		{"int8_ones", int8(1), int8(1), true},
		{"int8_zero_one", int8(0), int8(1), false},
		{"int16_zeros", int16(0), int16(0), true},
		{"int16_ones", int16(1), int16(1), true},
		{"int16_zero_one", int16(0), int16(1), false},
		{"int32_zeros", int32(0), int32(0), true},
		{"int32_ones", int32(1), int32(1), true},
		{"int32_zero_one", int32(0), int32(1), false},
		{"int64_zeros", int64(0), int64(0), true},
		{"int64_ones", int64(1), int64(1), true},
		{"int64_zero_one", int64(0), int64(1), false},
		{"uint_zeros", uint(0), uint(0), true},
		{"uint_ones", uint(1), uint(1), true},
		{"uint_zero_one", uint(0), uint(1), false},
		{"uint8_zeros", uint8(0), uint8(0), true},
		{"uint8_ones", uint8(1), uint8(1), true},
		{"uint8_zero_one", uint8(0), uint8(1), false},
		{"uint16_zeros", uint16(0), uint16(0), true},
		{"uint16_ones", uint16(1), uint16(1), true},
		{"uint16_zero_one", uint16(0), uint16(1), false},
		{"uint32_zeros", uint32(0), uint32(0), true},
		{"uint32_ones", uint32(1), uint32(1), true},
		{"uint32_zero_one", uint32(0), uint32(1), false},
		{"uint64_zeros", uint64(0), uint64(0), true},
		{"uint64_ones", uint64(1), uint64(1), true},
		{"uint64_zero_one", uint64(0), uint64(1), false},
		{"float32_zeros", float32(0), float32(0), true},
		{"float32_ones", float32(1), float32(1), true},
		{"float32_zero_one", float32(0), float32(1), false},
		{"float64_zeros", float64(0), float64(0), true},
		{"float64_ones", float64(1), float64(1), true},
		{"float64_zero_one", float64(0), float64(1), false},
		{"json_number_eq", json.Number("0"), json.Number("0.0"), true},
		{"json_number_ne", json.Number("1024"), json.Number("0.0"), false},
		{"json_number_invalid", json.Number("not a number"), json.Number("0.0"), false},
		{"int_float_true", int64(10), float64(10), true},
		{"int_float_false", int64(10), float64(11), false},
		{"empty_strings", "", "", true},
		{"strings", "xyz", "xyz", true},
		{"strings_false", "xyz", "abc", false},
		{"unicode_strings", "foü", "foü", true},
		{"emoji_strings", "hi 😀", "hi 😀", true},
		{"trues", true, true, true},
		{"true_false", true, false, false},
		{"arrays_equal", []any{1, 2, 3}, []any{1, 2, 3}, true},
		{"arrays_ne", []any{1, 2, 3}, []any{1, 2, 3, 4}, false},
		{"nils", nil, nil, true},
		{"nil_not_nil", nil, 2, false},
		{"objects_eq", map[string]any{"x": 1, "y": 2}, map[string]any{"x": 1, "y": 2}, true},
		{"object_keys_ne", map[string]any{"x": 1, "y": 2}, map[string]any{"x": 1, "z": 2}, false},
		{"object_vals_ne", map[string]any{"x": 1, "y": 2}, map[string]any{"x": 1, "y": 3}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.exp, valueEqualTo(tc.left, tc.right))
			a.Equal(tc.exp, equalTo(Value(tc.left), Value(tc.right)))
		})
	}

	t.Run("not_comparable", func(t *testing.T) {
		t.Parallel()
		a.False(valueEqualTo(42, "x"))
		a.False(equalTo(nil, Value(42)))
		a.False(equalTo(Value(42), nil))
		a.False(equalTo(LogicalFalse, LogicalFalse))
	})
}

func TestLessThan(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name  string
		left  any
		right any
		exp   bool
	}{
		{"int_zeros", 0, 0, false},
		{"int_zero_one", 0, 1, true},
		{"int_one_zero", 1, 0, false},
		{"int8_zeros", 0, 0, false},
		{"int8_zero_one", 0, 1, true},
		{"int8_one_zero", 1, 0, false},
		{"int16_zeros", 0, 0, false},
		{"int16_zero_one", 0, 1, true},
		{"int16_one_zero", 1, 0, false},
		{"int32_zeros", 0, 0, false},
		{"int32_zero_one", 0, 1, true},
		{"int32_one_zero", 1, 0, false},
		{"int64_zeros", 0, 0, false},
		{"int64_zero_one", 0, 1, true},
		{"int64_one_zero", 1, 0, false},
		{"uint_zeros", 0, 0, false},
		{"uint_zero_one", 0, 1, true},
		{"uint_one_zero", 1, 0, false},
		{"uint8_zeros", 0, 0, false},
		{"uint8_zero_one", 0, 1, true},
		{"uint8_one_zero", 1, 0, false},
		{"uint16_zeros", 0, 0, false},
		{"uint16_zero_one", 0, 1, true},
		{"uint16_one_zero", 1, 0, false},
		{"uint32_zeros", 0, 0, false},
		{"uint32_zero_one", 0, 1, true},
		{"uint32_one_zero", 1, 0, false},
		{"uint64_zeros", 0, 0, false},
		{"uint64_zero_one", 0, 1, true},
		{"uint64_one_zero", 1, 0, false},
		{"float32_zeros", 0, 0, false},
		{"float32_zero_one", 0, 1, true},
		{"float32_one_zero", 1, 0, false},
		{"float64_zeros", 0, 0, false},
		{"float64_zero_one", 0, 1, true},
		{"float64_one_zero", 1, 0, false},
		{"int_float_true", 12, 98.6, true},
		{"int_float_false", 99, 98.6, false},
		{"float_int_false", 98.6, 98, false},
		{"float_int_true", 98.6, 99, true},
		{"empty_string_sting", "", "x", true},
		{"empty_strings", "", "", false},
		{"string_a_b", "a", "b", true},
		{"string_c_b", "c", "b", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.exp, valueLessThan(tc.left, tc.right))
			a.Equal(tc.exp, lessThan(Value(tc.left), Value(tc.right)))
		})
	}

	t.Run("not_comparable", func(t *testing.T) {
		t.Parallel()
		a.False(lessThan(LogicalFalse, Value(".")))
		a.False(lessThan(Value("x"), LogicalFalse))
		a.False(valueLessThan(42, "x"))
		a.False(valueLessThan([]any{0}, []any{1}))
	})
}

func TestSameType(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name  string
		left  PathValue
		right PathValue
		exp   bool
	}{
		{"int_nodes", Nodes(1), Nodes(0), true},
		{"float_nodes", Nodes(98.6), Nodes(22.4), true},
		{"bool_nodes", Nodes(true), Nodes(false), true},
		{"string_nodes", Nodes("hi"), Nodes("go"), true},
		{"object_nodes", Nodes(map[string]any{}), Nodes(map[string]any{}), true},
		{"array_nodes", Nodes([]any{}), Nodes([]any{}), true},
		{"nil_nodes", Nodes(nil), Nodes(nil), true},
		{"int_float_nodes", Nodes(1), Nodes(98.6), true},
		{"int64_uint32_nodes", Nodes(int64(1)), Nodes(uint32(8)), true},
		{"int_bool_nodes", Nodes(1), Nodes(false), false},
		{"string_obj_nodes", Nodes("hi"), Nodes(map[string]any{}), false},
		{"int64_array_nodes", Nodes(int64(9)), Nodes([]any{}), false},
		{"int_vals", Value(1), Value(0), true},
		{"float_vals", Value(98.6), Value(22.4), true},
		{"bool_vals", Value(true), Value(false), true},
		{"string_vals", Value("hi"), Value("go"), true},
		{"object_vals", Value(map[string]any{}), Value(map[string]any{}), true},
		{"array_vals", Value([]any{}), Value([]any{}), true},
		{"nil_vals", Value(nil), Value(nil), true},
		{"int_float_vals", Value(1), Value(98.6), true},
		{"int64_uint32_vals", Value(int64(1)), Value(uint32(8)), true},
		{"int_bool_vals", Value(1), Value(false), false},
		{"string_obj_vals", Value("hi"), Value(map[string]any{}), false},
		{"int64_array_vals", Value(int64(9)), Value([]any{}), false},
		{"nodes_multi", Nodes(1, 1), Nodes(1, 1), false},
		{"nodes_multi_sing", Nodes(1, 1), Nodes(1), false},

		{"nodes_val_int", Nodes(0), Value(1), true},
		{"nodes_val_float", Nodes(1.1), Value(2.2), true},
		{"nodes_val_numbers", Nodes(1), Value(2.2), true},
		{"nodes_val_bool", Nodes(true), Value(false), true},
		{"nodes_val_string", Nodes("hi"), Value("go"), true},
		{"nodes_val_object", Nodes(map[string]any{}), Value(map[string]any{}), true},
		{"nodes_val_array", Nodes([]any{"x"}), Value([]any{1}), true},
		{"nodes_val_nil", Nodes(nil), Value(nil), true},
		{"nodes_val_int_bool", Nodes(21), Value(false), false},
		{"nodes_val_string_nil", Nodes("hi"), Value(nil), false},
		{"nodes_val_obj_array", Nodes(map[string]any{}), Value([]any{}), false},
		{"nodes_bool_logical", Nodes(true), LogicalFalse, true},
		{"nodes_string_logical", Nodes("x"), LogicalFalse, false},
		{"nodes_int_logical", Nodes(42), LogicalFalse, false},
		{"multi_nodes_val", Nodes(0, 0), Value(1), false},
		{"multi_nodes_logical", Nodes(true, true), LogicalTrue, false},
		{"nodes_json_number", Nodes(json.Number("1")), Value(1), true},
		{"nodes_json_numbers", Nodes(json.Number("1")), Value(json.Number("10")), true},

		{"val_nodes_int", Value(0), Nodes(1), true},
		{"val_nodes_float", Value(1.1), Nodes(2.2), true},
		{"val_nodes_numbers", Value(1), Nodes(2.2), true},
		{"val_nodes_bool", Value(true), Nodes(false), true},
		{"val_nodes_string", Value("hi"), Nodes("go"), true},
		{"val_nodes_object", Value(map[string]any{}), Nodes(map[string]any{}), true},
		{"val_nodes_array", Value([]any{"x"}), Nodes([]any{1}), true},
		{"val_nodes_nil", Value(nil), Nodes(nil), true},
		{"val_nodes_int_bool", Value(21), Nodes(false), false},
		{"val_nodes_string_nil", Value("hi"), Nodes(nil), false},
		{"val_nodes_obj_array", Value(map[string]any{}), Nodes([]any{}), false},
		{"val_bool_logical", Value(true), LogicalFalse, true},
		{"val_string_logical", Value("x"), LogicalFalse, false},
		{"val_int_logical", Value(42), LogicalFalse, false},

		{"logical_types", LogicalFalse, LogicalTrue, true},
		{"logical_val_bool", LogicalFalse, Value(false), true},
		{"logical_nodes_bool", LogicalFalse, Nodes(false), true},
		{"logical_val_string", LogicalFalse, Value("true"), false},
		{"logical_nodes_string", LogicalFalse, Nodes("true"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a.Equal(tc.exp, sameType(tc.left, tc.right))
		})
	}
}

func TestComparisonExpr(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	for _, tc := range []struct {
		name    string
		left    CompVal
		right   CompVal
		root    any
		current any
		expect  []bool
		str     string
	}{
		{
			name:   "literal_numbers_eq",
			left:   Literal(42),
			right:  Literal(42),
			expect: []bool{true, false, false, false, true, true},
			str:    "42 %v 42",
		},
		{
			name:   "literal_numbers_lt",
			left:   Literal(42),
			right:  Literal(43),
			expect: []bool{false, true, true, false, true, false},
			str:    "42 %v 43",
		},
		{
			name:   "literal_numbers_gt",
			left:   Literal(43),
			right:  Literal(42),
			expect: []bool{false, true, false, true, false, true},
			str:    "43 %v 42",
		},
		{
			name:   "literal_strings_eq",
			left:   Literal("x"),
			right:  Literal("x"),
			expect: []bool{true, false, false, false, true, true},
			str:    `"x" %v "x"`,
		},
		{
			name:   "literal_strings_lt",
			left:   Literal("x"),
			right:  Literal("y"),
			expect: []bool{false, true, true, false, true, false},
			str:    `"x" %v "y"`,
		},
		{
			name:   "literal_strings_gt",
			left:   Literal("y"),
			right:  Literal("x"),
			expect: []bool{false, true, false, true, false, true},
			str:    `"y" %v "x"`,
		},
		{
			name:   "query_numbers_eq",
			left:   SingularQuery(true, Name("x")),
			right:  SingularQuery(true, Name("y")),
			root:   map[string]any{"x": 42, "y": 42},
			expect: []bool{true, false, false, false, true, true},
			str:    `$["x"] %v $["y"]`,
		},
		{
			name:    "query_numbers_lt",
			left:    SingularQuery(false, Name("x")),
			right:   SingularQuery(false, Name("y")),
			current: map[string]any{"x": 42, "y": 43},
			expect:  []bool{false, true, true, false, true, false},
			str:     `@["x"] %v @["y"]`,
		},
		{
			name:   "query_string_gt",
			left:   SingularQuery(true, Name("y")),
			right:  SingularQuery(true, Name("x")),
			root:   map[string]any{"x": "x", "y": "y"},
			expect: []bool{false, true, false, true, false, true},
			str:    `$["y"] %v $["x"]`,
		},
		{
			name: "func_numbers_eq",
			left: &FuncExpr{
				args: []FuncExprArg{SingularQuery(true, Name("x"))},
				fn:   newValueFunc(1),
			},
			right: &FuncExpr{
				args: []FuncExprArg{SingularQuery(true, Name("y"))},
				fn:   newValueFunc(1),
			},
			root:   map[string]any{"x": "xx", "y": "yy"},
			expect: []bool{true, false, false, false, true, true},
			str:    `__val($["x"]) %v __val($["y"])`,
		},
		{
			name: "func_numbers_lt",
			left: &FuncExpr{
				args: []FuncExprArg{SingularQuery(true, Name("x"))},
				fn:   newValueFunc(1),
			},
			right: &FuncExpr{
				args: []FuncExprArg{SingularQuery(true, Name("y"))},
				fn:   newValueFunc(2),
			},
			root:   map[string]any{"x": "xx", "y": "yyy"},
			expect: []bool{false, true, true, false, true, false},
			str:    `__val($["x"]) %v __val($["y"])`,
		},
		{
			name: "func_strings_gt",
			left: &FuncExpr{
				args: []FuncExprArg{Query(false, Child(Name("y")))},
				fn:   newValueFunc(42),
			},
			right: &FuncExpr{
				args: []FuncExprArg{Query(false, Child(Name("x")))},
				fn:   newValueFunc(41),
			},
			current: map[string]any{"x": "x", "y": "y"},
			expect:  []bool{false, true, false, true, false, true},
			str:     `__val(@["y"]) %v __val(@["x"])`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for i, op := range []struct {
				name string
				op   CompOp
			}{
				{"eq", EqualTo},
				{"ne", NotEqualTo},
				{"lt", LessThan},
				{"gt", GreaterThan},
				{"le", LessThanEqualTo},
				{"ge", GreaterThanEqualTo},
			} {
				t.Run(op.name, func(t *testing.T) {
					t.Parallel()
					cmp := Comparison(tc.left, op.op, tc.right)
					a.Equal(tc.expect[i], cmp.testFilter(tc.current, tc.root))
					a.Equal(fmt.Sprintf(tc.str, op.op), bufString(cmp))
				})
			}
		})

		t.Run("unknown_op", func(t *testing.T) {
			t.Parallel()
			cmp := Comparison(tc.left, CompOp(16), tc.right)
			a.Equal(fmt.Sprintf(tc.str, cmp.op), bufString(cmp))
			a.PanicsWithValue("Unknown operator CompOp(16)", func() {
				cmp.testFilter(tc.current, tc.root)
			})
		})
	}
}
