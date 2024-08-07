package jsonpath

import (
	"strings"
)

// Segment represents a single segment in an RFC 9535 JSONPath query,
// consisting of a list of Selectors and child Segments.
type Segment struct {
	selectors  []Selector
	descendant bool
}

// Child creates and returns a Segment that uses one or more Selectors
// to select the children of a JSON value.
func Child(sel ...Selector) *Segment {
	return &Segment{selectors: sel}
}

// Descendant creates and returns a Segment that uses one or more Selectors to
// select the children of a JSON value, together with the children of its
// children, and so forth recursively.
func Descendant(sel ...Selector) *Segment {
	return &Segment{selectors: sel, descendant: true}
}

// Selectors returns s's Selectors.
func (s *Segment) Selectors() []Selector {
	return s.selectors
}

// String returns a string representation of seg, including all of its child
// segments in as a tree diagram.
func (s *Segment) String() string {
	buf := new(strings.Builder)
	if s.descendant {
		buf.WriteString("..")
	}
	buf.WriteByte('[')
	for i, sel := range s.selectors {
		if i > 0 {
			buf.WriteByte(',')
		}
		sel.writeTo(buf)
	}
	buf.WriteByte(']')
	return buf.String()
}

// Select selects and returns values from input for each of seg's selectors.
func (s *Segment) Select(input any) []any {
	ret := []any{}
	for _, sel := range s.selectors {
		ret = append(ret, sel.Select(input)...)
	}
	if s.descendant {
		ret = append(ret, s.descend(input)...)
	}
	return ret
}

// descend recursively executes seg.Select for each value in input and returns
// the results.
func (s *Segment) descend(input any) []any {
	ret := []any{}
	switch val := input.(type) {
	case []any:
		for _, v := range val {
			ret = append(ret, s.Select(v)...)
		}
	case map[string]any:
		for _, v := range val {
			ret = append(ret, s.Select(v)...)
		}
	}
	return ret
}
