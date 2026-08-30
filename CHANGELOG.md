# Changelog

All notable changes to this project will be documented in this file. It uses the
[Keep a Changelog] format, and this project adheres to [Semantic Versioning].

  [Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
  [Semantic Versioning]: https://semver.org/spec/v2.0.0.html
    "Semantic Versioning 2.0.0"

## [v0.12.1] — 2026-08-30

### 📚 Documentation

*   Added "Strict No LLM/No AI Policy" section do the README, borrowed from
    the [Zig CoC]

### ⬆️ Dependency Updates

*   Upgraded to `golangci-lint` v2.13.2.
*   Added testing for Go 1.27.
*   Upgraded Testify to v1.12.1.
*   Upgraded TinyGo to v0.41.1, now supports Go v1.26.

  [v0.12.1]: https://github.com/theory/jsonpath/compare/v0.12.0...v0.12.1
  [Zig CoC]: https://ziglang.org/code-of-conduct/ "Zig Code of Conduct"

## [v0.12.0] — 2026-04-12

### ⚡ Improvements

*   Added support for selecting values from any type of slice or string-keyed
    map, not just `[]any` or `map[string]any`. Internally it still prefers
    `[]any` and `map[string]any`, to optimize for values decoded by
    encoding/json, but it now falls back on reflection to detect any other
    kind of slice or string-keyed map. Thanks to @ndsboy for the prompt (#26).
*   Updated result set creation to allocate more slots for results when the
    number of results are unknown, based on the number of selectors or items
    to select from, to improve memory efficiency. Encouraged by the recent
    [Go blog post] describing the advantages of this pattern.

### ⬆️ Dependency Updates

*   Upgraded to `golangci-lint` v2.11.4 and made suggested slice allocation
    optimizations.

### 📚 Documentation

*   Fixed some broken Go Doc links.

  [v0.11.1]: https://github.com/theory/jsonpath/compare/v0.11.0...v0.11.1
  [Go blog post]: https://go.dev/blog/allocation-optimizations
    "The Go Blog: Allocating on the Stack"

## [v0.11.0] — 2026-03-02

### ⚡ Improvements

*   Added some explicit slice size declarations to improve performance.

### ⬆️ Dependency Updates

*   Upgraded to `golangci-lint` v2.10.1
*   Added test coverage for Go v1.25 and v1.26
*   Pulled the latest `jsonpath-compliance-test-suite` and
    `regression_suite.yaml`

### 📔 Notes

*   Removed unused constant from `parser/lex.go`
*   Removed an unnecessary integer conversions reported as overflow risks by
    golangci-lint.

  [v0.10.3]: https://github.com/theory/jsonpath/compare/v0.10.2...v0.11.0

## [v0.10.2] — 2025-10-14

### 📔 Notes

*   Upgraded to `golangci-lint` v2.5.0.
*   Removed unused constant from `parser/lex.go`

### 🐞 Bug Fixes

*   Fixed bug that prevented multiple blank-space delimited ANDed comparison
    operations from being parsed. For example, `$[@.x=="hi"&&@.y!=3&&@[1]==1]`
    would parse but `$[@.x == "hi" && @.y !=3 && @[1]==1]` would not. Thanks
    to @jarangutan for the bug report and @jg-rp for the analysis ([#24]).

  [v0.10.2]: https://github.com/theory/jsonpath/compare/v0.10.1...v0.10.2
  [#24]: https://github.com/theory/jsonpath/issues/24
    "theory/jsonpath#24 Filter with 2+ consecutive '&&' operators returns parsing error"

## [v0.10.1] — 2025-09-16

### 🐞 Bug Fixes

*   Allow `true`, `false`, and `null` to be used as selectors, e.g., `$.true`.

### 📔 Notes

*   Upgraded to `golangci-lint` v2.3.1.
*   Fixed test name scoping issues with testify objects.
*   Upgraded the [compliance test suite], now with tests using `true`,
    `false`, and `null` to be used as selectors.

  [v0.10.1]: https://github.com/theory/jsonpath/compare/v0.10.0...v0.10.1
  [compliance test suite]: https://github.com/jsonpath-standard/jsonpath-compliance-test-suite

## [v0.10.0] — 2025-07-11

### ⚡ Improvements

*   Added text and binary encoding interface methods to `Path`:
    *   [encoding.TextMarshaler]
    *   [encoding.TextUnmarshaler]
    *   [encoding.BinaryMarshaler]
    *   [encoding.BinaryUnmarshaler]
    Thanks to @rkosegi for the suggestion ([#20])

### 📔 Notes

*   Upgraded to `golangci-lint` v2.2.2.

  [v0.10.0]: https://github.com/theory/jsonpath/compare/v0.9.0...v0.10.0
  [encoding.TextMarshaler]: https://pkg.go.dev/encoding#TextMarshaler
  [encoding.TextUnmarshaler]: https://pkg.go.dev/encoding#TextUnmarshaler
  [encoding.BinaryMarshaler]: https://pkg.go.dev/encoding#BinaryMarshaler
  [encoding.BinaryUnmarshaler]: https://pkg.go.dev/encoding#BinaryUnmarshaler
  [#20]: https://github.com/theory/jsonpath/issues/20
    "theory/jsonpath#20: Implement encoding.TextUnmarshaler in Path"

## [v0.9.0] — 2025-05-05

### ⚡ Improvements

*   Significantly refactored the `spec` package toward greater stability,
    usability, and increased comprehensibility. The names of things are more
    consistent, the APIs more legible and user-friendly. Quite a few types
    were renamed or merged.
*   Added support for [json.Number] values to complement the existing support
    for Go core numeric types. This should allow for transparent handling of
    values marshaled with [json.Decoder.UseNumber] enabled.
*   Moved the function extension types from the `registry` to the `spec`
    package, simplifying `registry` and the handling of function extensions,
    without changing the interface for using a registry or adding extensions
    to it.

### 📚 Documentation

*   Vastly expanded the `spec` package documentation to make it much more
    thorough and accurate, with copious links to relevant parts of [RFC 9535]
    and complete lists of interface implementations and examples for each
    significant type.
*   Removed the "Package Stability" statement from the README, as all packages
    are considered stable or else potentially unstable types in the `spec`
    package have been labeled as such.
*   Fixed links and typos in the main package documentation, and moved the
    registry example under `WithRegistry`.
*   Replaced the `spec.Wildcard` variable with a the function
    `spec.Wildcard()`.

### 📔 Notes

*   Upgraded the [compliance test suite] and integrated its located path test
    expectations. All tests passed, no changes required.

  [v0.9.0]: https://github.com/theory/jsonpath/compare/v0.4.1...v0.9.0
  [json.Number]: https://pkg.go.dev/encoding/json#Number
  [json.Decoder.UseNumber]: https://pkg.go.dev/encoding/json#Decoder.UseNumber
  [compliance test suite]: https://github.com/jsonpath-standard/jsonpath-compliance-test-suite
  [RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535.html
    "RFC 9535 JSONPath: Query Expressions for JSON"

## [v0.4.1] — 2025-05-02

### 🪲 Bug Fixes

*   Fixed the panic messages from `spec.Slice` to properly report coming from
    `Slice` and not `NewSlice`.

### 📔 Notes

*   Upgraded to and fixed issues reported by `golangci-lint` v2.1.5.
*   Added TinyGo testing on Go v1.24.

  [v0.4.1]: https://github.com/theory/jsonpath/compare/v0.4.0...v0.4.1

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
