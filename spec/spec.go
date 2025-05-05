// Package spec provides the [RFC 9535 JSONPath] [AST] and execution for
// [github.com/theory/jsonpath]. It will mainly be of interest to those
// wishing to implement their own parsers, to convert from other JSON
// path-style languages, and to implement functions for
// [github.com/theory/jsonpath/registry].
//
// # Stability
//
// The following types and constructors are considered stable:
//
//   - [Index]
//   - [Name]
//   - [SliceSelector] and [Slice]
//   - [WildcardSelector] and [Wildcard]
//   - [FilterSelector]
//   - [Segment] and [Child] and [Descendant]
//   - [PathQuery] and [Query]
//   - [LocatedNode]
//   - [NormalizedPath]
//
// The rest of the structs, constructors, and methods in this package remain
// subject to change, although we anticipate no significant revisions.
//
// [RFC 9535 JSONPath]: https://www.rfc-editor.org/rfc/rfc9535.html
// [AST]: https://en.wikipedia.org/wiki/Abstract_syntax_tree
package spec
