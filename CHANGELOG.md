# Changelog

All notable changes to this project will be documented in this file. It uses the
[Keep a Changelog] format, and this project adheres to [Semantic Versioning].

  [Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
  [Semantic Versioning]: https://semver.org/spec/v2.0.0.html
    "Semantic Versioning 2.0.0"

## [v0.1.2] — Unreleased

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
