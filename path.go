// Package jsonpath implements RFC 9535 JSONPath query expressions.
package jsonpath

import (
	"iter"
	"slices"

	"github.com/theory/jsonpath/parser"
	"github.com/theory/jsonpath/registry"
	"github.com/theory/jsonpath/spec"
)

// ErrPathParse errors are returned for path parse errors.
var ErrPathParse = parser.ErrPathParse

// Path represents a [RFC 9535] JSONPath query.
//
// [RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535.html
type Path struct {
	q *spec.PathQuery
}

// New creates and returns a new Path consisting of q.
func New(q *spec.PathQuery) *Path {
	return &Path{q: q}
}

// Parse parses path, a JSONPath query string, into a Path. Returns an
// ErrPathParse on parse failure.
func Parse(path string) (*Path, error) {
	return NewParser().Parse(path)
}

// MustParse parses path into a Path. Panics with an ErrPathParse on parse
// failure.
func MustParse(path string) *Path {
	return NewParser().MustParse(path)
}

// String returns a string representation of p.
func (p *Path) String() string {
	return p.q.String()
}

// Query returns p's root Query.
func (p *Path) Query() *spec.PathQuery {
	return p.q
}

// Select returns the values that JSONPath query p selects from input.
func (p *Path) Select(input any) NodeList {
	return p.q.Select(nil, input)
}

// SelectLocated returns the values that JSONPath query p selects from input
// as [spec.LocatedNode] values that pair the values with the [normalized
// paths] that identify them. Unless you have a specific need for the unique
// normalized path for each value, you probably want to use [Path.Select].
//
// [normalized paths]: https://www.rfc-editor.org/rfc/rfc9535#section-2.7
func (p *Path) SelectLocated(input any) LocatedNodeList {
	return p.q.SelectLocated(nil, input, spec.Normalized())
}

// Parser parses JSONPath strings into [Path] values.
type Parser struct {
	reg *registry.Registry
}

// Option defines a parser option.
type Option func(*Parser)

// WithRegistry configures a Parser with a function Registry, which may
// contain function extensions. See [Parser] for an example.
func WithRegistry(reg *registry.Registry) Option {
	return func(p *Parser) { p.reg = reg }
}

// NewParser creates a new Parser configured by opt.
func NewParser(opt ...Option) *Parser {
	p := &Parser{}
	for _, o := range opt {
		o(p)
	}

	if p.reg == nil {
		p.reg = registry.New()
	}

	return p
}

// Parse parses path, a JSON Path query string, into a Path. Returns an
// ErrPathParse on parse failure.
//
//nolint:wrapcheck
func (c *Parser) Parse(path string) (*Path, error) {
	q, err := parser.Parse(c.reg, path)
	if err != nil {
		return nil, err
	}
	return New(q), nil
}

// MustParse parses path, a JSON Path query string, into a Path. Panics with
// an ErrPathParse on parse failure.
func (c *Parser) MustParse(path string) *Path {
	q, err := parser.Parse(c.reg, path)
	if err != nil {
		panic(err)
	}
	return New(q)
}

// NodeList is a list of nodes selected by a JSONPath query. Each node
// represents a single JSON value selected from the JSON query argument.
// Returned by [Path.Select].
type NodeList []any

// All returns an iterator over all the nodes in list.
//
// Range over list itself to get indexes as well as values.
func (list NodeList) All() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, v := range list {
			if !yield(v) {
				return
			}
		}
	}
}

// LocatedNodeList is a list of nodes selected by a JSONPath query, along with
// their locations. Returned by [Path.SelectLocated].
type LocatedNodeList []*spec.LocatedNode

// All returns an iterator over all the nodes in list.
//
// Range over list itself to get indexes and node values.
func (list LocatedNodeList) All() iter.Seq[*spec.LocatedNode] {
	return func(yield func(*spec.LocatedNode) bool) {
		for _, v := range list {
			if !yield(v) {
				return
			}
		}
	}
}

// Nodes returns an iterator over all the nodes in list. This is effectively
// the same data a returned by [Path.Select].
func (list LocatedNodeList) Nodes() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, v := range list {
			if !yield(v.Node) {
				return
			}
		}
	}
}

// Paths returns an iterator over all the normalized paths in list.
func (list LocatedNodeList) Paths() iter.Seq[spec.NormalizedPath] {
	return func(yield func(spec.NormalizedPath) bool) {
		for _, v := range list {
			if !yield(v.Path) {
				return
			}
		}
	}
}

// Deduplicate deduplicates the nodes in list based on their normalized paths,
// modifying the contents of list. It returns the modified list, which may
// have a smaller length, and zeroes the elements between the new length and
// the original length.
func (list LocatedNodeList) Deduplicate() LocatedNodeList {
	if len(list) <= 1 {
		return list
	}

	seen := map[string]struct{}{}
	uniq := list[:0]
	for _, n := range list {
		p := n.Path.String()
		if _, x := seen[p]; !x {
			seen[p] = struct{}{}
			uniq = append(uniq, n)
		}
	}
	clear(list[len(uniq):]) // zero/nil out the obsolete elements, for GC
	return slices.Clip(uniq)
}

// Sort sorts list by the normalized path of each node.
func (list LocatedNodeList) Sort() {
	slices.SortFunc(list, func(a, b *spec.LocatedNode) int {
		return a.Path.Compare(b.Path)
	})
}

// Clone returns a shallow copy of list.
func (list LocatedNodeList) Clone() LocatedNodeList {
	return append(make(LocatedNodeList, 0, len(list)), list...)
}
