package spec

import (
	"strings"
)

// NormalSelector represents a single selector in a normalized path.
// Implemented by [Name] and [Index].
type NormalSelector interface {
	// writeNormalizedTo writes n to buf formatted as a [normalized path] element.
	//
	// [normalized path]: https://www.rfc-editor.org/rfc/rfc9535#section-2.7
	writeNormalizedTo(buf *strings.Builder)
}

// NormalizedPath represents a normalized path identifying a single value in a
// JSON query argument, as [defined by RFC 9535].
//
// [defined by RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535#name-normalized-paths
type NormalizedPath []NormalSelector

// String returns the string representation of np.
func (np NormalizedPath) String() string {
	buf := new(strings.Builder)
	buf.WriteRune('$')
	for _, e := range np {
		e.writeNormalizedTo(buf)
	}
	return buf.String()
}

// MarshalText marshals np into text. It implements [encoding.TextMarshaler].
func (np NormalizedPath) MarshalText() ([]byte, error) {
	return []byte(np.String()), nil
}

// LocatedNode pairs a value with its location within the JSON query argument
// from which it was selected.
type LocatedNode struct {
	// Node is the value selected from a JSON query argument.
	Node any `json:"node"`

	// Path is the normalized path that uniquely identifies the location of
	// Node in a JSON query argument.
	Path NormalizedPath `json:"path"`
}

// newLocatedNode creates and returns a new [Node]. It makes a copy of path.
func newLocatedNode(path NormalizedPath, node any) *LocatedNode {
	return &LocatedNode{
		Path: NormalizedPath(append(make([]NormalSelector, 0, len(path)), path...)),
		Node: node,
	}
}
