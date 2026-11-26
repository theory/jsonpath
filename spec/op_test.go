package spec

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestCompOp(t *testing.T) {
	t.Parallel()

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
		assert.Equal(t, tc.str, tc.op.String())
	}
}

type TestInterface interface {
	X()
}

type TestInterfaceImpl struct {
	Num int
}

func (*TestInterfaceImpl) X() {}

//nolint:ireturn
func NilTestInterface() TestInterface { return nil }

//nolint:ireturn
func NumTestInterface(n int) TestInterface { return &TestInterfaceImpl{n} }

func TestEqualTo(t *testing.T) {
	t.Parallel()

	aSlice := []int{0, 1, 4}
	cycle := []any{1, 2, nil}
	cycle[2] = cycle
	pointer := &TestInterfaceImpl{42}
	aMap := map[string]int{"hi": 2}
	cycleMap := map[string]any{}
	cycleMap["x"] = cycleMap
	ptr := NumTestInterface(42)
	var (
		nilSlice []any
		nilMap   map[string]any
		nilFunc  func()
		nilChan  chan int
		nilFace  any
		face     any = 1
		face1    any = 1
		face2    any = 2
	)

	for _, tc := range []struct {
		test  string
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
		{"json_number_invalid_ref", json.Number("0.0"), json.Number("not a number"), false},
		{"int_float_true", int64(10), float64(10), true},
		{"int_float_false", int64(10), float64(11), false},
		{"empty_strings", "", "", true},
		{"strings", "xyz", "xyz", true},
		{"strings_false", "xyz", "abc", false},
		{"unicode_strings", "foü", "foü", true},
		{"emoji_strings", "hi 😀", "hi 😀", true},
		{"trues", true, true, true},
		{"true_false", true, false, false},
		{"slices_equal", []any{1, 2, 3}, []any{1, 2, 3}, true},
		{"slices_ne_len", []any{1, 2, 3}, []any{1, 2, 3, 4}, false},
		{"slices_ne", []any{1, 2, 3}, []any{1, 2, 4}, false},
		{"slice_numbers", []any{1, 2, 3}, []any{1.0, 2, 3.0}, true},
		{"slice_diff_types", []int{1, 2, 3}, []float32{1.0, 2, 3.0}, true},
		{"arrays_equal", [3]any{1, 2, 3}, [3]any{1, 2, 3}, true},
		{"arrays_ne_len", [3]any{1, 2, 3}, [4]any{1, 2, 3, 4}, false},
		{"arrays_ne", [3]any{1, 2, 3}, [3]any{1, 2, 4}, false},
		{"array_numbers", [3]any{1, 2, 3}, [3]any{1.0, 2, 3.0}, true},
		{"array_diff_types", [3]int{1, 2, 3}, [3]float32{1.0, 2, 3.0}, true},
		{"nils", nil, nil, true},
		{"nil_slices", nilSlice, nilSlice, true},
		{"nested_nil_slices", []any{nilSlice, nilSlice}, []any{nilSlice, nilSlice}, true},
		{"nil_maps", nilMap, nilMap, true},
		{"nested_nil_maps", []any{nilMap, nilMap}, []any{nilMap, nilMap}, true},
		{"nil_not_nil", nil, 2, false},
		{"objects_eq", map[string]any{"x": 1, "y": 2}, map[string]any{"x": 1, "y": 2}, true},
		{"object_keys_ne", map[string]any{"x": 1, "y": 2}, map[string]any{"x": 1, "z": 2}, false},
		{"object_vals_ne", map[string]any{"x": 1, "y": 2}, map[string]any{"x": 1, "y": 3}, false},
		{"object_lens_ne", map[string]any{"x": 1, "y": 2}, map[string]any{"x": 1}, false},
		{"object_numbers", map[string]any{"x": 1, "y": 2}, map[string]any{"x": 1.0, "y": 2.0}, true},
		{
			"nested_numbers",
			map[string]any{"x": []int{42, 99}, "y": map[string]int{"a": 6, "b": 7}},
			map[string]any{"x": []float32{42, 99}, "y": map[string]float32{"a": 6.0, "b": 7.0}},
			true,
		},
		{"repeat_slices", []any{aSlice, aSlice}, []any{aSlice, aSlice}, true},
		{"cycle_slices", cycle, cycle, true},
		{"nested_cycles_eq", []any{cycle, cycle}, []any{cycle, cycle}, true},
		{"multi_nested_cycle_slices", map[string]any{"x": cycle}, map[string]any{"x": cycle}, true},
		{"repeat_maps", []any{aMap, aMap}, []any{aMap, aMap}, true},
		{"cycle_maps", cycleMap, cycleMap, true},
		{"nested_cycle_maps", []any{1, cycleMap}, []any{1, cycleMap}, true},
		{"multi_nested_cycle_maps", []any{cycleMap, cycleMap}, []any{cycleMap, cycleMap}, true},
		{"nils", nil, nil, true},
		{"nil_ptrs", NilTestInterface(), NilTestInterface(), true},
		{"one_nil_ptrs", NilTestInterface(), NumTestInterface(1), false},
		{"ptrs", NumTestInterface(1), NumTestInterface(1), true},
		{"ptrs_ne", NumTestInterface(1), NumTestInterface(2), false},
		{"nested_nil_ptrs", []any{NilTestInterface()}, []any{NilTestInterface()}, true},
		{"nested_one_nil_ptrs", []any{NilTestInterface()}, []any{NumTestInterface(1)}, false},
		{"nested_ptrs_eq", []any{NumTestInterface(1)}, []any{NumTestInterface(1)}, true},
		{"same_ptrs", ptr, ptr, true},
		{"same_nested_ptrs", []any{ptr, ptr}, []any{ptr, ptr}, true},
		{"empty_structs", TestInterfaceImpl{}, TestInterfaceImpl{}, true},
		{"same_structs", TestInterfaceImpl{1}, TestInterfaceImpl{1}, true},
		{"diff_structs", TestInterfaceImpl{1}, TestInterfaceImpl{2}, false},
		{"empty_p_structs", &TestInterfaceImpl{}, &TestInterfaceImpl{}, true},
		{"same_p_structs", &TestInterfaceImpl{1}, &TestInterfaceImpl{1}, true},
		{"diff_p_structs", &TestInterfaceImpl{1}, &TestInterfaceImpl{2}, false},
		{"struct_private", reflect.ValueOf(cycle).Index(0), reflect.ValueOf(cycle).Index(0), false},
		{"diff_struct_private", reflect.ValueOf(cycle).Index(0), reflect.ValueOf(cycle).Index(1), false},
		{"same_pointer", pointer, pointer, true},
		{"same_nested_pointer", []*TestInterfaceImpl{pointer}, []*TestInterfaceImpl{pointer}, true},
		{"repeated_pointer", []any{pointer, pointer}, []any{pointer, pointer}, true},
		{"same_map", aMap, aMap, true},
		{"string_num", "string", 0, false},
		{"num_string", 0, "string", false},
		{"array_slice_eq", []any{1, true}, [2]any{1, true}, true},
		{"array_slice_ne", []any{1, true}, [2]any{1, false}, false},
		{"array_slice_len_ne", []any{1, true}, [3]any{1, true}, false},
		{"slice_array_eq", [2]any{1, true}, []any{1, true}, true},
		{"slice_array_ne", [2]any{1, true}, []any{1, false}, false},
		{"slice_array_len_ne", [3]any{1, true}, []any{1, true}, false},
		{"complex", complex(22, 3), complex(22, 3), true},
		{"complex_ne", complex(22, 3), complex(22, 2), false},
		{"complex64", complex64(22), complex64(22), true},
		{"complex64_ne", complex64(22), complex64(12), false},
		{"complex128", complex128(22), complex128(22), true},
		{"complex128_ne", complex128(22), complex128(12), false},
		{"nil_funcs", nilFunc, nilFunc, true},
		{"nested_nil_funcs", []any{nilFunc, nilFunc}, []any{nilFunc, nilFunc}, true},
		{"nil_channels", nilChan, nilChan, true},
		{"nested_nil_channels", []any{nilChan, nilChan}, []any{nilChan, nilChan}, true},
		{"funcs", TestEqualTo, TestEqualTo, false},
		{"chans", make(chan int), make(chan int), false},
		//nolint:gosec // disable G103
		{"unsafe_pointers_eq", unsafe.Pointer(t), unsafe.Pointer(t), true},
		//nolint:gosec // disable G103
		{"unsafe_pointers_ne", unsafe.Pointer(t), unsafe.Pointer(pointer), false},
		{"nested_repeat_slices", []any{aSlice, aSlice}, []any{aSlice, aSlice}, true},
		{"nil_interfaces", nilFace, nilFace, true},
		{"nested_nil_interfaces", []any{nilFace, nilFace}, []any{nilFace, nilFace}, true},
		{"same_interfaces", face, face, true},
		{"same_nested_interfaces", []any{face, face}, []any{face, face}, true},
		{"dif_nested_interfaces_ne", []any{face, face2}, []any{face, face}, false},
		{"same_nested_interfaces_eq", []any{face, face1}, []any{face, face}, true},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.Equal(tc.exp, valueEqualTo(tc.left, tc.right))
			a.Equal(tc.exp, equalTo(Value(tc.left), Value(tc.right)))
		})
	}

	t.Run("not_comparable", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		a.False(valueEqualTo(reflect.ValueOf(42), reflect.ValueOf("x")))
		a.False(equalTo(nil, Value(42)))
		a.False(equalTo(Value(42), nil))
		a.False(equalTo(LogicalFalse, LogicalFalse))
	})
}

func TestLessThan(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test  string
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
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.Equal(tc.exp, valueLessThan(tc.left, tc.right))
			a.Equal(tc.exp, lessThan(Value(tc.left), Value(tc.right)))
		})
	}

	t.Run("not_comparable", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		a.False(lessThan(LogicalFalse, Value(".")))
		a.False(lessThan(Value("x"), LogicalFalse))
		a.False(valueLessThan(42, "x"))
		a.False(valueLessThan([]any{0}, []any{1}))
	})
}

func TestSameType(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test  string
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
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.exp, sameType(tc.left, tc.right))
		})
	}
}

func TestComparisonExpr(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test    string
		left    CompVal
		right   CompVal
		root    any
		current any
		expect  []bool
		str     string
	}{
		{
			test:   "literal_numbers_eq",
			left:   Literal(42),
			right:  Literal(42),
			expect: []bool{true, false, false, false, true, true},
			str:    "42 %v 42",
		},
		{
			test:   "literal_numbers_lt",
			left:   Literal(42),
			right:  Literal(43),
			expect: []bool{false, true, true, false, true, false},
			str:    "42 %v 43",
		},
		{
			test:   "literal_numbers_gt",
			left:   Literal(43),
			right:  Literal(42),
			expect: []bool{false, true, false, true, false, true},
			str:    "43 %v 42",
		},
		{
			test:   "literal_strings_eq",
			left:   Literal("x"),
			right:  Literal("x"),
			expect: []bool{true, false, false, false, true, true},
			str:    `"x" %v "x"`,
		},
		{
			test:   "literal_strings_lt",
			left:   Literal("x"),
			right:  Literal("y"),
			expect: []bool{false, true, true, false, true, false},
			str:    `"x" %v "y"`,
		},
		{
			test:   "literal_strings_gt",
			left:   Literal("y"),
			right:  Literal("x"),
			expect: []bool{false, true, false, true, false, true},
			str:    `"y" %v "x"`,
		},
		{
			test:   "query_numbers_eq",
			left:   SingularQuery(true, Name("x")),
			right:  SingularQuery(true, Name("y")),
			root:   map[string]any{"x": 42, "y": 42},
			expect: []bool{true, false, false, false, true, true},
			str:    `$["x"] %v $["y"]`,
		},
		{
			test:    "query_numbers_lt",
			left:    SingularQuery(false, Name("x")),
			right:   SingularQuery(false, Name("y")),
			current: map[string]any{"x": 42, "y": 43},
			expect:  []bool{false, true, true, false, true, false},
			str:     `@["x"] %v @["y"]`,
		},
		{
			test:   "query_string_gt",
			left:   SingularQuery(true, Name("y")),
			right:  SingularQuery(true, Name("x")),
			root:   map[string]any{"x": "x", "y": "y"},
			expect: []bool{false, true, false, true, false, true},
			str:    `$["y"] %v $["x"]`,
		},
		{
			test: "func_numbers_eq",
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
			test: "func_numbers_lt",
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
			test: "func_strings_gt",
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
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			for i, op := range []struct {
				text string
				op   CompOp
			}{
				{"eq", EqualTo},
				{"ne", NotEqualTo},
				{"lt", LessThan},
				{"gt", GreaterThan},
				{"le", LessThanEqualTo},
				{"ge", GreaterThanEqualTo},
			} {
				t.Run(op.text, func(t *testing.T) {
					t.Parallel()
					a := assert.New(t)

					cmp := Comparison(tc.left, op.op, tc.right)
					a.Equal(tc.expect[i], cmp.testFilter(tc.current, tc.root))
					a.Equal(fmt.Sprintf(tc.str, op.op), bufString(cmp))
				})
			}
		})

		t.Run("unknown_op", func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			cmp := Comparison(tc.left, CompOp(16), tc.right)
			a.Equal(fmt.Sprintf(tc.str, cmp.op), bufString(cmp))
			a.PanicsWithValue("Unknown operator CompOp(16)", func() {
				cmp.testFilter(tc.current, tc.root)
			})
		})
	}
}

func TestFloat(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		test string
		val  any
		exp  float64
		ok   bool
	}{
		{"int", int(42), 42.0, true},
		{"int8", int8(42), 42.0, true},
		{"int16", int16(42), 42.0, true},
		{"int32", int32(42), 42.0, true},
		{"int364", int64(42), 42.0, true},
		{"neg_int", -42, -42.0, true},
		{"neg_int8", int8(-42), -42.0, true},
		{"neg_int16", int16(-42), -42.0, true},
		{"neg_int32", int32(-42), -42.0, true},
		{"neg_int364", int64(-42), -42.0, true},
		{"uint", uint(42), 42.0, true},
		{"uint8", uint8(42), 42.0, true},
		{"uint16", uint16(42), 42.0, true},
		{"uint32", uint32(42), 42.0, true},
		{"uint64", uint64(42), 42.0, true},
		{"float32", float32(98.6), float64(float32(98.6)), true},
		{"float64", float64(98.6), 98.6, true},
		{"json_number_int", json.Number("42"), 42, true},
		{"json_number_float", json.Number("98.6"), 98.6, true},
		// Invalid values.
		{"invalid_json_number", json.Number(""), 0, false},
		{"string", "42", 0, false},
		{"map", map[string]any{"x": 42}, 0, false},
		{"slice", []int{42}, 0, false},
		{"bool", true, 0, false},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			got, ok := toFloat(tc.val)
			//nolint:testifylint // Disable float-compare: equivalence assured
			a.Equal(tc.exp, got)
			a.Equal(tc.ok, ok)
		})
	}
}

func TestPointerFor(t *testing.T) {
	t.Parallel()
	v := reflect.ValueOf([]any{1, 1})
	aSlice := []any{1, true}
	aMap := map[string]any{"x": true, "y": 1}
	ptr := &TestInterfaceImpl{}

	for _, tc := range []struct {
		test   string
		first  reflect.Value
		second reflect.Value
		kind   reflect.Kind
		equal  bool
	}{
		{"same_internal_interface_pointer", v.Index(0), v.Index(0), reflect.Interface, true},
		{"diff_internal_interface_pointer", v.Index(0), v.Index(1), reflect.Interface, false},
		{"same_slice", reflect.ValueOf(aSlice), reflect.ValueOf(aSlice), reflect.Slice, true},
		{"diff_slice", reflect.ValueOf(aSlice), reflect.ValueOf(slices.Clone(aSlice)), reflect.Slice, false},
		{"same_map", reflect.ValueOf(aMap), reflect.ValueOf(aMap), reflect.Map, true},
		{"diff_map", reflect.ValueOf(aMap), reflect.ValueOf(maps.Clone(aMap)), reflect.Map, false},
		{"same_pointer", reflect.ValueOf(ptr), reflect.ValueOf(ptr), reflect.Pointer, true},
		{"diff_pointer", reflect.ValueOf(ptr), reflect.ValueOf(&TestInterfaceImpl{}), reflect.Pointer, false},
	} {
		t.Run(tc.test, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			a.Equal(tc.kind, tc.first.Kind())
			a.Equal(tc.kind, tc.second.Kind())
			if tc.equal {
				a.Equal(pointerFor(tc.first), pointerFor(tc.second))
			} else {
				a.NotEqual(pointerFor(tc.first), pointerFor(tc.second))
			}
		})
	}
}
