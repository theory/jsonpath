// Package jsonpath implements RFC 9535 JSONPath query expressions.
package jsonpath

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Selector represents a single Selector in an RFC 9535 JSONPath query.
type Selector interface {
	fmt.Stringer
	// writeTo writes a string representation of a selector to buf.
	writeTo(buf *strings.Builder)
	// Select selects values from input and returns them.
	Select(input any) []any
}

// Name is a key name selector, e.g., .name or ["name"].
type Name string

// String returns a quoted string representation of n.
func (n Name) String() string {
	return strconv.Quote(string(n))
}

// writeTo writes a quoted string representation of i to buf.
func (n Name) writeTo(buf *strings.Builder) {
	buf.WriteString(n.String())
}

// Select selects n from input and returns it as a single value in a slice.
// Returns an empty slice if input is not a map[string]any or if it does not
// contain n.
func (n Name) Select(input any) []any {
	if obj, ok := input.(map[string]any); ok {
		if val, ok := obj[string(n)]; ok {
			return []any{val}
		}
	}
	return make([]any, 0)
}

// wc is the underlying nil value used by [Wildcard].
type wc struct{}

// Wildcard is a wildcard selector, e.g., * or [*].
//
//nolint:gochecknoglobals
var Wildcard = wc{}

// writeTo  writes "*" to buf.
func (wc) writeTo(buf *strings.Builder) { buf.WriteByte('*') }

// String returns "*".
func (wc) String() string { return "*" }

// Select selects the values from input and returns them in a slice. Returns
// an empty slice if input is not []any map[string]any.
func (wc) Select(input any) []any {
	switch val := input.(type) {
	case []any:
		return val
	case map[string]any:
		vals := make([]any, 0, len(val))
		for _, v := range val {
			vals = append(vals, v)
		}
		return vals
	}
	return make([]any, 0)
}

// Index is an array index selector, e.g., [3].
type Index int

// writeTo writes a string representation of i to buf.
func (i Index) writeTo(buf *strings.Builder) {
	buf.WriteString(i.String())
}

// String returns a string representation of i.
func (i Index) String() string { return strconv.FormatInt(int64(i), 10) }

// Select selects i from input and returns it as a single value in a slice.
// Returns an empty slice if input is not a slice or if i it outside the
// bounds of input.
func (i Index) Select(input any) []any {
	if val, ok := input.([]any); ok {
		if int(i) < len(val) {
			return []any{val[i]}
		}
	}
	return make([]any, 0)
}

// SliceSelector is a slice selector, e.g., [0:100:5].
type SliceSelector struct {
	// Start of the slice; defaults to 0.
	start int
	// End of the slice; defaults to math.MaxInt.
	end int
	// Steps between start and end; defaults to 0.
	step int
}

// Slice creates a new SliceSelector. Pass up to three integers or nils for
// the start, end, and step arguments. Subsequent arguments are ignored.
func Slice(args ...any) SliceSelector {
	const (
		startArg = 0
		endArg   = 1
		stepArg  = 2
	)
	// Set defaults.
	s := SliceSelector{0, math.MaxInt, 1}
	switch len(args) - 1 {
	case stepArg:
		switch step := args[stepArg].(type) {
		case int:
			s.step = step
		case nil:
			// Nothing to do
		default:
			panic("Third value passed to NewSlice is not an integer")
		}
		fallthrough
	case endArg:
		switch end := args[endArg].(type) {
		case int:
			s.end = end
		case nil:
			// Negative step: end with minimum int.
			if s.step < 0 {
				s.end = math.MinInt
			}
		default:
			panic("Second value passed to NewSlice is not an integer")
		}
		fallthrough
	case startArg:
		switch start := args[startArg].(type) {
		case int:
			s.start = start
		case nil:
			// Negative step: start with maximum int.
			if s.step < 0 {
				s.start = math.MaxInt
			}
		default:
			panic("First value passed to NewSlice is not an integer")
		}
	}
	return s
}

// writeTo writes a string representation of s to buf.
func (s SliceSelector) writeTo(buf *strings.Builder) {
	if s.start != 0 {
		buf.WriteString(strconv.FormatInt(int64(s.start), 10))
	}
	buf.WriteByte(':')
	if s.end != math.MaxInt {
		buf.WriteString(strconv.FormatInt(int64(s.end), 10))
	}
	if s.step != 1 {
		buf.WriteByte(':')
		buf.WriteString(strconv.FormatInt(int64(s.step), 10))
	}
}

// String returns a quoted string representation of s.
func (s SliceSelector) String() string {
	buf := new(strings.Builder)
	s.writeTo(buf)
	return buf.String()
}

// Select selects and returns the values from input for the indexes specified
// by s. Returns an empty slice if input is not a slice. Indexes outside the
// bounds of input will not be included in the return value.
func (s SliceSelector) Select(input any) []any {
	if val, ok := input.([]any); ok {
		lower, upper := s.bounds(len(val))
		res := make([]any, 0, len(val))
		switch {
		case s.step > 0:
			for i := lower; i < upper; i += s.step {
				res = append(res, val[i])
			}
		case s.step < 0:
			for i := upper; lower < i; i += s.step {
				res = append(res, val[i])
			}
		}
		return res
	}
	return make([]any, 0)
}

// Start returns the start position.
func (s SliceSelector) Start() int {
	return s.start
}

// End returns the end position.
func (s SliceSelector) End() int {
	return s.end
}

// Step returns the step value.
func (s SliceSelector) Step() int {
	return s.step
}

// bounds returns the lower and upper bounds for selecting from a slice of
// length.
func (s SliceSelector) bounds(length int) (int, int) {
	start := normalize(s.start, length)
	end := normalize(s.end, length)
	switch {
	case s.step > 0:
		return max(min(start, length), 0), max(min(end, length), 0)
	case s.step < 0:
		return max(min(end, length-1), -1), max(min(start, length-1), -1)
	default:
		return 0, 0
	}
}

// normalize normalizes index i relative to a slice of length.
func normalize(i, length int) int {
	if i >= 0 {
		return i
	}

	return length + i
}
