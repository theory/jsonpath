package spec

//go:generate stringer -linecomment -output op_string.go -type CompOp

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

// CompOp defines the JSONPath [filter comparison operators].
//
// [filter comparison operators]: https://www.rfc-editor.org/rfc/rfc9535.html#name-comparisons
type CompOp uint8

const (
	// EqualTo is the == operator.
	EqualTo CompOp = iota + 1 // ==
	// NotEqualTo is the != operator.
	NotEqualTo // !=
	// LessThan is the < operator.
	LessThan // <
	// GreaterThan is the > operator.
	GreaterThan // >
	// LessThanEqualTo is the <= operator.
	LessThanEqualTo // <=
	// GreaterThanEqualTo is the >= operator.
	GreaterThanEqualTo // >=
)

// CompVal defines the interface for comparable values in filter
// expressions. Implemented by:
//
//   - [LiteralArg]
//   - [SingularQueryExpr]
//   - [FuncExpr]
type CompVal interface {
	stringWriter
	// asValue returns the value to be compared.
	asValue(current, root any) PathValue
}

// CompExpr is a filter expression that compares two values, which themselves
// may themselves be the output of expressions. Interfaces implemented:
//
//   - [BasicExpr]
//   - [fmt.Stringer]
type CompExpr struct {
	left  CompVal
	op    CompOp
	right CompVal
}

// Comparison creates and returns a new [CompExpr] that uses op to compare
// left and right.
func Comparison(left CompVal, op CompOp, right CompVal) *CompExpr {
	return &CompExpr{left, op, right}
}

// writeTo writes a string representation of ce to buf. Defined by
// [stringWriter].
func (ce *CompExpr) writeTo(buf *strings.Builder) {
	ce.left.writeTo(buf)
	fmt.Fprintf(buf, " %v ", ce.op)
	ce.right.writeTo(buf)
}

// String returns the string representation of ce.
func (ce *CompExpr) String() string {
	var buf strings.Builder
	ce.writeTo(&buf)
	return buf.String()
}

// testFilter uses ce.Op to compare the values returned by ce.Left and
// ce.Right relative to current and root. Defined by [BasicExpr].
func (ce *CompExpr) testFilter(current, root any) bool {
	left := ce.left.asValue(current, root)
	right := ce.right.asValue(current, root)
	switch ce.op {
	case EqualTo:
		return equalTo(left, right)
	case NotEqualTo:
		return !equalTo(left, right)
	case LessThan:
		return sameType(left, right) && lessThan(left, right)
	case GreaterThan:
		return sameType(left, right) && !lessThan(left, right) && !equalTo(left, right)
	case LessThanEqualTo:
		return sameType(left, right) && (lessThan(left, right) || equalTo(left, right))
	case GreaterThanEqualTo:
		return sameType(left, right) && !lessThan(left, right)
	default:
		panic(fmt.Sprintf("Unknown operator %v", ce.op))
	}
}

// equalTo returns true if left and right are nils, or if both are [ValueType]
// values and [valueEqualTo] returns true for their underlying values.
// Otherwise it returns false.
func equalTo(left, right PathValue) bool {
	switch left := left.(type) {
	case *ValueType:
		if right, ok := right.(*ValueType); ok {
			return valueEqualTo(left.any, right.any)
		}
	case nil:
		return right == nil
	}
	return false
}

// toFloat converts val to a float64 if it is a numeric value, setting ok to
// true. Otherwise it returns false for ok.
func toFloat(val any) (float64, bool) {
	switch val := val.(type) {
	case int:
		return float64(val), true
	case int8:
		return float64(val), true
	case int16:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint8:
		return float64(val), true
	case uint16:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

// During deepValueEqual, must keep track of checks that are in progress. The
// comparison algorithm assumes that all checks in progress are true when it
// reencounters them. Visited comparisons are stored in a map indexed by
// visit.
type visit struct {
	a1  unsafe.Pointer
	a2  unsafe.Pointer
	typ reflect.Type
}

func valueEqualTo(x, y any) bool {
	if x == nil || y == nil {
		return x == y
	}
	return deepEqual(reflect.ValueOf(x), reflect.ValueOf(y), make(map[visit]bool))
}

func deepEqualSlices(v1, v2 reflect.Value, visited map[visit]bool) bool {
	if v1.Len() != v2.Len() {
		return false
	}
	if v1.Kind() == reflect.Slice && v2.Kind() == reflect.Slice && v1.UnsafePointer() == v2.UnsafePointer() {
		return true
	}
	for i := range v1.Len() {
		if !deepEqual(v1.Index(i), v2.Index(i), visited) {
			return false
		}
	}
	return true
}

// Unfortunately, reflect.Value will not share the interface pointer.
func pointerFor(v reflect.Value) unsafe.Pointer {
	switch v.Kind() {
	case reflect.Interface:
		// Source - https://stackoverflow.com/a/57698257
		// Posted by Laevus Dexter, modified by community. See post 'Timeline' for change history
		// Retrieved 2026-11-25, License - CC BY-SA 4.0
		type face struct {
			typ, val unsafe.Pointer
		}
		//nolint:gosec // disable G103
		return (*face)(unsafe.Pointer(&v)).val
	default:
		return v.UnsafePointer()
	}
}

func seen(v1, v2 reflect.Value, visited map[visit]bool) bool {
	// Nil pointers cannot be cyclic. Avoid putting them in the visited map.
	if v1.IsNil() || v2.IsNil() {
		return false
	}
	addr1 := pointerFor(v1)
	addr2 := pointerFor(v2)
	if uintptr(addr1) > uintptr(addr2) {
		// Canonicalize order to reduce number of entries in visited.
		// Assumes non-moving garbage collector.
		addr1, addr2 = addr2, addr1
	}

	// Short circuit if references are already seen.
	typ := v1.Type()
	v := visit{addr1, addr2, typ}
	if visited[v] {
		return true
	}

	// Remember for later.
	visited[v] = true

	return false
}

// valueEqualTo tests for deep equality using reflected types. The map
// argument tracks comparisons that have already been seen, which allows short
// circuiting on recursive types. Based on [reflect.DeepEqual], but compares
// numeric, slice, and map values of any type.
func deepEqual(v1, v2 reflect.Value, visited map[visit]bool) bool {
	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid()
	}

	// Handle cases where different kinds are acceptable.
	if v1.Kind() != v2.Kind() {
		// Numbers must compare as floats.
		if left, ok := toFloat(v1.Interface()); ok {
			if right, ok := toFloat(v2.Interface()); ok {
				return left == right
			}
			return false
		}

		// Can compare any kind of array or slice.
		if (v1.Kind() == reflect.Array || v1.Kind() == reflect.Slice) &&
			(v2.Kind() == reflect.Array || v2.Kind() == reflect.Slice) {
			return deepEqualSlices(v1, v2, visited)
		}

		return false
	}

	// Everything else must be the same kind.
	switch v1.Kind() {
	case reflect.Array:
		return deepEqualSlices(v1, v2, visited)
	case reflect.Slice:
		if seen(v1, v2, visited) {
			return true
		}
		return deepEqualSlices(v1, v2, visited)
	case reflect.Interface:
		if seen(v1, v2, visited) {
			return true
		}
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		return deepEqual(v1.Elem(), v2.Elem(), visited)
	case reflect.Pointer:
		if seen(v1, v2, visited) {
			return true
		}
		if v1.UnsafePointer() == v2.UnsafePointer() {
			return true
		}
		return deepEqual(v1.Elem(), v2.Elem(), visited)
	case reflect.Struct:
		for i := range v1.NumField() {
			if !deepEqual(v1.Field(i), v2.Field(i), visited) {
				return false
			}
		}
		return true
	case reflect.Map:
		if seen(v1, v2, visited) {
			return true
		}
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.UnsafePointer() == v2.UnsafePointer() {
			return true
		}
		iter := v1.MapRange()
		for iter.Next() {
			val1 := iter.Value()
			val2 := v2.MapIndex(iter.Key())
			if !deepEqual(val1, val2, visited) {
				return false
			}
		}
		return true
	case reflect.String:
		// Handle supported numeric interfaces (json.Number).
		if f64Left, ok := toFloat(v1.Interface()); ok {
			if f64Right, ok := toFloat(v2.Interface()); ok {
				return f64Left == f64Right
			}
			return false
		}
		return v1.String() == v2.String()
	case reflect.Bool:
		return v1.Bool() == v2.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v1.Int() == v2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v1.Uint() == v2.Uint()
	case reflect.Float32, reflect.Float64:
		return v1.Float() == v2.Float()
	case reflect.Complex64, reflect.Complex128:
		return v1.Complex() == v2.Complex()
	case reflect.Chan, reflect.Func:
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		// Not comparable.
		return false
	default:
		if v1.CanInterface() {
			// Normal equality suffices.
			return v1.Interface() == v2.Interface()
		}
		return false
	}
}

// lessThan returns true if left and right are both ValueTypes and
// [valueLessThan] returns true for their underlying values. Otherwise it
// returns false.
func lessThan(left, right PathValue) bool {
	if left, ok := left.(*ValueType); ok {
		if right, ok := right.(*ValueType); ok {
			return valueLessThan(left.any, right.any)
		}
	}
	return false
}

// sameType returns true if left and right resolve to the same JSON data type.
func sameType(left, right PathValue) bool {
	switch left := left.(type) {
	case NodesType:
		if len(left) == 1 {
			switch right := right.(type) {
			case NodesType:
				return valCompType(left[0], right[0])
			case *ValueType:
				return valCompType(left[0], right.any)
			case LogicalType:
				_, ok := left[0].(bool)
				return ok
			}
		}
	case *ValueType:
		switch right := right.(type) {
		case *ValueType:
			return valCompType(left.any, right.any)
		case NodesType:
			return valCompType(left.any, right[0])
		case LogicalType:
			_, ok := left.any.(bool)
			return ok
		}
	case LogicalType:
		switch right := right.(type) {
		case LogicalType:
			return true
		case NodesType:
			if len(right) == 1 {
				_, ok := right[0].(bool)
				return ok
			}
		case *ValueType:
			_, ok := right.any.(bool)
			return ok
		}
	}
	return false
}

// valCompType returns true if left and right are comparable types, which
// means either both are a numeric type or are otherwise the same type.
func valCompType(left, right any) bool {
	switch left.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, json.Number:
		switch right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, json.Number:
			return true
		}
	}
	return reflect.TypeOf(left) == reflect.TypeOf(right)
}

// valueLessThan returns true if left and right are both numeric values or
// string values and left is less than right.
func valueLessThan(left, right any) bool {
	if left, ok := toFloat(left); ok {
		if right, ok := toFloat(right); ok {
			return left < right
		}
		return false
	}

	if left, ok := left.(string); ok {
		if right, ok := right.(string); ok {
			return left < right
		}
	}

	return false
}
