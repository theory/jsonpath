package spec

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

// Select selects and returns values from current or root for each of seg's
// selectors. Defined by the [Selector] interface.
func (s *Segment) Select(current, root any) []any {
	ret := []any{}
	for _, sel := range s.selectors {
		ret = append(ret, sel.Select(current, root)...)
	}
	if s.descendant {
		ret = append(ret, s.descend(current, root)...)
	}
	return ret
}

// SelectLocated selects and returns values as [LocatedNode] structs from
// current or root for each of seg's selectors. Defined by the [Selector]
// interface.
func (s *Segment) SelectLocated(current, root any, parent NormalizedPath) []*LocatedNode {
	ret := []*LocatedNode{}
	for _, sel := range s.selectors {
		ret = append(ret, sel.SelectLocated(current, root, parent)...)
	}
	if s.descendant {
		ret = append(ret, s.descendLocated(current, root, parent)...)
	}
	return ret
}

// descend recursively executes seg.Select for each value in current and/or
// root and returns the results.
func (s *Segment) descend(current, root any) []any {
	ret := []any{}
	switch val := current.(type) {
	case []any:
		for _, v := range val {
			ret = append(ret, s.Select(v, root)...)
		}
	case map[string]any:
		for _, v := range val {
			ret = append(ret, s.Select(v, root)...)
		}
	}
	return ret
}

// descend recursively executes seg.Select for each value in current and/or
// root and returns the results.
func (s *Segment) descendLocated(current, root any, parent NormalizedPath) []*LocatedNode {
	ret := []*LocatedNode{}
	switch val := current.(type) {
	case []any:
		for i, v := range val {
			ret = append(ret, s.SelectLocated(v, root, append(parent, Index(i)))...)
		}
	case map[string]any:
		for k, v := range val {
			ret = append(ret, s.SelectLocated(v, root, append(parent, Name(k)))...)
		}
	}
	return ret
}

// isSingular returns true if the segment selects at most one node. Defined by
// the [Selector] interface.
func (s *Segment) isSingular() bool {
	if s.descendant || len(s.selectors) != 1 {
		return false
	}
	return s.selectors[0].isSingular()
}

// IsDescendant returns true if the segment is a descendant selector that
// recursively select the children of a JSON value.
func (s *Segment) IsDescendant() bool { return s.descendant }
