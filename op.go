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
	Left comparableVal
	// The comparison operator.
	Op CompOp
	// An expression that produces the JSON value for the right side of the
	// comparison.
	Right comparableVal
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
		return lessThan(left, right)
	case GreaterThan:
		return !lessThan(left, right) && !equalTo(left, right)
	case LessThanEqualTo:
		return lessThan(left, right) || equalTo(left, right)
	case GreaterThanEqualTo:
		return !lessThan(left, right)
	default:
		panic(fmt.Sprintf("Unknown operator %v", ce.Op))
	}
}

// equalTo returns true if left and right are both [ValueType]s and
// [valueEqualTo] returns true for their underlying values. Otherwise it
// returns false.
func equalTo(left, right JSONPathValue) bool {
	if left, ok := left.(*ValueType); ok {
		if right, ok := right.(*ValueType); ok {
			return valueEqualTo(left.any, right.any)
		}
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
