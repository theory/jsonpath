# Changelog

All notable changes to this project will be documented in this file. It uses the
[Keep a Changelog] format, and this project adheres to [Semantic Versioning].

  [Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
  [Semantic Versioning]: https://semver.org/spec/v2.0.0.html
    "Semantic Versioning 2.0.0"

## [v0.4.0] — 2025-01-15

### ⚡ Improvements

*   Added the `Pointer` method to `NormalizedPath`. It returns an [RFC 9535
    JSON Pointer] string representation of the normalized path.

  [v0.4.0]: https://github.com/theory/jsonpath/compare/v0.3.0...v0.4.0
  [RFC 9535 JSON Pointer]: https://www.rfc-editor.org/rfc/rfc9535#name-normalized-paths

## [v0.3.0] — 2024-12-28

### ⚡ Improvements

*   Added `SelectLocated`. It works just like `Select`, but returns
    `LocatedNode`s that pair the selected nodes with [RFC 9535-defined]
    `NormalizedPath`s that uniquely identify their locations within the JSON
    query argument.
*   Added `LocatedNodeList`, the return value from `SelectLocated`. It
    contains methods for working with the selected nodes, including iterators
    for its nodes & `NormalizedPath`s, deduplication, sorting, and cloning.
*   Added `Compare` to `NormalizedPath`, which enables the sorting of
    `LocatedNodeList`s.

### 📔 Notes

*   Requires Go 1.23 to take advantage of its iterator support.
*   Changed the return value of `Select` from `[]any` to `NodeList`, which is
    an alias for `[]any`. Done to pair with `LocatedNodeList`, the return
    value of `SelectLocated`. Features an `All` method, which returns an
    iterator over all the nodes in the list. It may gain additional methods in
    the future.

### 📚 Documentation

*   Added `Select`, `SelectLocated`, `NodeList`, and `LocatedNodeList`
    examples to the Go docs.

  [v0.3.0]: https://github.com/theory/jsonpath/compare/v0.2.1...v0.3.0
  [RFC 9535-defined]: https://www.rfc-editor.org/rfc/rfc9535#section-2.7

## [v0.2.1] — 2024-12-12

### 🪲 Bug Fixes

*   Fixed the formatting of slice strings to omit min and max integers when
    not specified and using a negative step.

  [v0.2.1]: https://github.com/theory/jsonpath/compare/v0.2.0...v0.2.1

## [v0.2.0] — 2024-11-13

### ⚡ Improvements

*   Added `spec.Filter.Eval` to allow public evaluation of a single JSON node.
    Used internally by `spec.FilterSelector.Select`.
*   Added `spec.Segment.IsDescendant` to tell wether a segments selects just
    from the current child node or also recursively selects from all of its
    descendants.

### 🪲 Bug Fixes

*   Added missing "?" to the stringification of `spec.FilterSelector`.

### 📔 Notes

*   Made `spec.SliceSelector.Bounds` public.
*   Made the underlying struct defining `spec.Wildcard` public, named it
    `spec.WildcardSelector`.

  [v0.2.0]: https://github.com/theory/jsonpath/compare/v0.1.2...v0.2.0

## [v0.1.2] — 2024-10-28

### 🪲 Bug Fixes

*   Eliminated a lexer variable that prevented [TinyGo] compilation.

### 🏗️ Build Setup

*   Added simple tests to ensure the package compiles properly as Go and
    TinyGo WASM.
*   Added the WASM compile test to the [Test and Lint] GitHub action.

  [v0.1.2]: https://github.com/theory/jsonpath/compare/v0.1.1...v0.1.2
  [TinyGo]: https://tinygo.org "TinyGo — A Go Compiler For Small Places"
  [Test and Lint]: https://github.com/theory/jsonpath/actions/workflows/ci.yml

### 📚 Documentation

*   Fixed version header links here in CHANGELOG.md.

## [v0.1.1] — 2024-09-19

### 📚 Documentation

*   Neatened the formatting of the README table for improved display on
    pkg.go.dev.

  [v0.1.1]: https://github.com/theory/jsonpath/compare/v0.1.0...v0.1.1

## [v0.1.0] — 2024-09-19

The theme of this release is *Standards Matter.*

### ⚡ Improvements

*   First release, everything is new!
*   Full [RFC 9535] JSONPath implementation
*   All [JSONPath Compliance Test Suite] tests pass
*   Includes parser, AST, and executor

### 🏗️ Build Setup

*   Built with Go
*   Use `go get` to add to a project

### 📚 Documentation

*   Docs on [pkg.go.dev]
*   Syntax summary in `README`

### 📔 Notes

*   The `jsonpath` package is stable and unlikely to change
*   The `spec` package is not yet stable
*   The `registry` package is stable, although `spec` objects it references
    may change
*   More detailed documentation to come

  [v0.1.0]: https://github.com/theory/jsonpath/compare/a7279e6...v0.1.0
  [RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535.html
    "RFC 9535 JSONPath: Query Expressions for JSON"
  [JSONPath Compliance Test Suite]: https://github.com/jsonpath-standard/jsonpath-compliance-test-suite
    "A Compliance Test Suite for the RFC 9535 JSONPath Standard"
  [pkg.go.dev]: https://pkg.go.dev/github.com/theory/jsonpath
