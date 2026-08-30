RFC 9535 JSONPath in Go
=======================

[![⚖️ MIT]][mit] [![📚 Docs]][docs] [![🗃️ Report Card]][card] [![🛠️ Build Status]][ci] [![📊 Coverage]][cov]

The jsonpath package provides [RFC 9535 JSONPath] functionality in Go.
It operates on any type of slice or string-keyed map.

## Learn More

*   📦 See the [package documentation][docs] for usage and examples.
*   📚 See the [RFC 9535 JSONPath] standard for details on the JSONPath query
    syntax and examples of its usage.
*   🛝 Try it out in the [Playground].

## JSONPath Expressions

A brief overview of [RFC 9535 JSONPath] syntax:

| Syntax Element     | Description                                                             |
| ------------------ | ----------------------------------------------------------------------- |
| `$`                | root node identifier                                                    |
| `@`                | current node identifier (valid only within filter selectors)            |
| `[<selectors>]`    | child segment: selects zero or more children of a node                  |
| `.name`            | shorthand for `['name']`                                                |
| `.*`               | shorthand for `[*]`                                                     |
| `..[<selectors>]`  | descendant segment: selects zero or more descendants of a node          |
| `..name`           | shorthand for `..['name']`                                              |
| `..*`              | shorthand for `..[*]`                                                   |
| `'name'`           | name selector: selects a named child of an object                       |
| `*`                | wildcard selector: selects all children of a node                       |
| `3`                | index selector: selects an indexed child of an array (from 0)           |
| `0:100:5`          | array slice selector: `start:end:step` for arrays                       |
| `?<logical-expr>`  | filter selector: selects particular children using a logical expression |
| `length(@.foo)`    | function extension: invokes  a function in a filter expression          |

## Dependencies

This package has only test dependencies.

## Strict No LLM/No AI Policy

This software is 100% made by humans for humans and welcomes human
contributions. That means:

*   No LLM-generated content, whether it be code or prose.
*   No paraphrasing LLM-generated content.
*   No LLMs for editing, including fixing spelling or grammatical errors.
*   No LLMs for translation. English is encouraged, but not required. You are
    welcome to post in your native language and rely on others to have their
    own translation tools of choice to interpret your words.
*   No LLMs for brainstorming and then sharing the results of that
    brainstorming, even if you create the prose. If you use a chatbot to give
    you advice on a comment on the issue tracker, that comment is unwelcome.
*   No LLMs for finding bugs.
*   No talking about use of chatbot/LLM services.

Borrowed from the [Zig Code of Conduct].

## Copyright

Copyright © 2024-2026 David E. Wheeler

  [⚖️ MIT]: https://img.shields.io/badge/License-MIT-blue.svg "⚖️ MIT License"
  [mit]: https://opensource.org/license/MIT "⚖️ MIT License"
  [📚 Docs]: https://godoc.org/github.com/theory/jsonpath?status.svg "📚 Documentation"
  [docs]: https://pkg.go.dev/github.com/theory/jsonpath "📄 Documentation"
  [🗃️ Report Card]: https://goreportcard.com/badge/github.com/theory/jsonpath
    "🗃️ Report Card"
  [card]: https://goreportcard.com/report/github.com/theory/jsonpath
    "🗃️ Report Card"
  [🛠️ Build Status]: https://github.com/theory/jsonpath/actions/workflows/ci.yml/badge.svg
    "🛠️ Build Status"
  [ci]: https://github.com/theory/jsonpath/actions/workflows/ci.yml
    "🛠️ Build Status"
  [📊 Coverage]: https://codecov.io/gh/theory/jsonpath/graph/badge.svg?token=UB1UJ95NIK
    "📊 Code Coverage"
  [cov]: https://codecov.io/gh/theory/jsonpath "📊 Code Coverage"
  [RFC 9535 JSONPath]: https://www.rfc-editor.org/rfc/rfc9535.html
    "RFC 9535 JSONPath: Query Expressions for JSON"
  [Playground]: https://theory.github.io/jsonpath/ "Go JSONPath Playground"
  [Zig Code of Conduct]: https://ziglang.org/code-of-conduct/
