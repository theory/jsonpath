package spec

import (
	"cmp"
	"strings"
)

// NormalSelector represents a single selector in a normalized path.
// Implemented by [Name] and [Index].
type NormalSelector interface {
	// writeNormalizedTo writes the selector to buf formatted as a [normalized
	// path] element.
	//
	// [normalized path]: https://www.rfc-editor.org/rfc/rfc9535#section-2.7
	writeNormalizedTo(buf *strings.Builder)

	// writePointerTo writes the selector to buf formatted as a [JSON Pointer]
	// reference token.
	//
	// [JSON Pointer]: https://www.rfc-editor.org/rfc/rfc6901
	writePointerTo(buf *strings.Builder)
}

// NormalizedPath represents a normalized path identifying a single value in a
// JSON query argument, as [defined by RFC 9535]. Defines a simplified string
// format that exclusively uses single quotation marks to quote names. Useful
// for converting to JSON Pointer (via [NormalizedPath.Pointer]) or other JSON
// pointer-type formats.
//
// Interfaces implemented:
//   - [fmt.Stringer]
//   - [encoding.TextMarshaler]
//
// [defined by RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535#name-normalized-paths
type NormalizedPath []NormalSelector

// Normalized creates and returns a [NormalizedPath] that contains sel.
func Normalized(sel ...NormalSelector) NormalizedPath {
	return NormalizedPath(sel)
}

// String returns the string representation of np.
func (np NormalizedPath) String() string {
	buf := new(strings.Builder)
	buf.WriteRune('$')
	for _, e := range np {
		e.writeNormalizedTo(buf)
	}
	return buf.String()
}

// Pointer returns an [RFC 6901 JSON Pointer] string representation of np.
//
// [RFC 6901 JSON Pointer]: https://www.rfc-editor.org/rfc/rfc6901
func (np NormalizedPath) Pointer() string {
	buf := new(strings.Builder)
	for _, e := range np {
		buf.WriteRune('/')
		e.writePointerTo(buf)
	}
	return buf.String()
}

// Compare compares np to np2 and returns -1 if np is less than np2, 1 if it's
// greater than np2, and 0 if they're equal. Indexes are always considered
// less than names.
func (np NormalizedPath) Compare(np2 NormalizedPath) int {
	for i := range np {
		if i >= len(np2) {
			return 1
		}
		switch v1 := np[i].(type) {
		case Name:
			switch v2 := np2[i].(type) {
			case Name:
				if x := cmp.Compare(v1, v2); x != 0 {
					return x
				}
			case Index:
				return 1
			}
		case Index:
			switch v2 := np2[i].(type) {
			case Index:
				if x := cmp.Compare(v1, v2); x != 0 {
					return x
				}
			case Name:
				return -1
			}
		}
	}

	if len(np2) > len(np) {
		return -1
	}
	return 0
}

// MarshalText marshals np into text. Implements [encoding.TextMarshaler].
func (np NormalizedPath) MarshalText() ([]byte, error) {
	return []byte(np.String()), nil
}

// LocatedNode pairs a value with the [NormalizedPath] for its location within
// the JSON query argument from which it was selected. Returned by
// implementations of [Selector]'s SelectLocated method.
type LocatedNode struct {
	// Node is the value selected from a JSON query argument.
	Node any `json:"node"`

	// Path is the normalized path that uniquely identifies the location of
	// Node in a JSON query argument.
	Path NormalizedPath `json:"path"`
}

// newLocatedNode creates and returns a new [LocatedNode]. It copies path.
func newLocatedNode(path NormalizedPath, node any) *LocatedNode {
	return &LocatedNode{
		Path: NormalizedPath(append(make([]NormalSelector, 0, len(path)), path...)),
		Node: node,
	}
}
