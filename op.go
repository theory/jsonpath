package jsonpath

//go:generate stringer -linecomment -output op_string.go -type CompOp

import (
	"fmt"
	"reflect"
	"strings"
)

// CompOp defines the JSONPath filter comparison operators.
type CompOp uint8

//revive:disable:exported
const (
	EqualTo            CompOp = iota + 1 // ==
	NotEqualTo                           // !=
	LessThan                             // <
	GreaterThan                          // >
	LessThanEqualTo                      // <=
	GreaterThanEqualTo                   // >=
)

// ComparisonExpr represents the comparison of two values, which themselves
// may be the output of expressions.
type ComparisonExpr struct {
	// An expression that produces the JSON value for the left side of the
	// comparison.
	Left CompVal
	// The comparison operator.
	Op CompOp
	// An expression that produces the JSON value for the right side of the
	// comparison.
	Right CompVal
}

// writeTo writes a string representation of ce to buf.
func (ce *ComparisonExpr) writeTo(buf *strings.Builder) {
	ce.Left.writeTo(buf)
	fmt.Fprintf(buf, " %v ", ce.Op)
	ce.Right.writeTo(buf)
}

// testFilter uses ce.Op to compare the values returned by ce.Left and
// ce.Right relative to current and root.
func (ce *ComparisonExpr) testFilter(current, root any) bool {
	left := ce.Left.asValue(current, root)
	right := ce.Right.asValue(current, root)
	switch ce.Op {
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
		panic(fmt.Sprintf("Unknown operator %v", ce.Op))
	}
}

// equalTo returns true if left and right are nils, or if both are
// [ValueType]s and [valueEqualTo] returns true for their underlying values.
// Otherwise it returns false.
func equalTo(left, right JSONPathValue) bool {
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
		return float64(val), true
	default:
		return 0, false
	}
}

// valueEqualTo returns true if left and right are equal.
func valueEqualTo(left, right any) bool {
	if left, ok := toFloat(left); ok {
		if right, ok := toFloat(right); ok {
			return left == right
		}
		return false
	}

	return reflect.DeepEqual(left, right)
}

// lessThan returns true if left and right are both ValueTypes and
// [valueLessThan] returns true for their underlying values. Otherwise it
// returns false.
func lessThan(left, right JSONPathValue) bool {
	if left, ok := left.(*ValueType); ok {
		if right, ok := right.(*ValueType); ok {
			return valueLessThan(left.any, right.any)
		}
	}
	return false
}

// sameType returns true if left and right resolve to the same JSON data type.
func sameType(left, right JSONPathValue) bool {
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
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		switch right.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
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
