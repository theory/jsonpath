// Package jsonpath implements RFC 9535 JSONPath query expressions.
package jsonpath

import (
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
func (p *Path) Select(input any) []any {
	return p.q.Select(nil, input)
}

// Parser parses JSONPath strings into [*Path]s.
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
