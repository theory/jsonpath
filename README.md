Go RFC 9535 JSONPath Playground
===============================

The source for the [Go JSONPath Playground], a stateless single-page web site
for experimenting with the [Go RFC 9535 JSONPath] package. Compiled via
[TinyGo] into a ca. 670 K (254 K compressed) [Wasm] file and loaded directly
into the page. All functionality implemented in JavaScript and Go, heavily
borrowed from the [Goldmark Playground] and [serde_json_path Sandbox].

Usage
-----

On load, the form will be filled with sample JSON and a randomly-selected
example query. Hit the "Run Query" button to see the values the path query
selects from the JSON appear in the "Query Output" field.

To try your own, paste the JSON to query into the "JSON" field and input the
JSONpath query into the "Path" field, then hit the "Run Query" button.

That's it.

Read on for details and additional features.

### Docs

The two buttons in the top-right corner provide documentation and links.

*   Hit the button with the circled question mark in the top right corner to
    reveal a table summarizing the JSONPath syntax.

*   Hit the button with the circled i for information about the JSONPath
    playground.

### Options

Select options for execution and the display of results:

*   **Located**: Show the normalized path location of each value.

### Permalink

Hit the "Permalink" button instead of "Run Query" to reload the page with a
URL that contains the contents the JSONPath, JSON, and options and executes
the results. Copy the URL to use it for sharing.

Note that the Playground is stateless; no data is stored except in the
Permalink URL itself (and whatever data collection GitHub injects; see its
[privacy statement] for details).

### Path

Input the JSONPath query to execute into this field. On load, the app will
pre-load an example query. See [RFC 9535] for details on the jsonpath
language.

### JSON Input

Input the JSON against which to execute the JSONPath query. May be any kind
of JSON value, including objects, arrays, and scalar values. On load, the
field will contain the JSON object used in examples from [RFC 9535].

## Copyright and License

Copyright (c) 2024 David E. Wheeler. Distributed under the [MIT License].

Based on [Goldmark Playground] the [serde_json_path Sandbox], with icons from
[Boxicons], all distributed under the [MIT License].

  [Go JSONPath Playground]: https://theory.github.io/jsonpath/playground
  [Go RFC 9535 JSONPath]: https://pkg.go.dev/github.com/theory/jsonpath
    "pkg.go.dev: github.com/theory/jsonpath"
  [Wasm]: https://webassembly.org "WebAssembly"
  [TinyGo]: https://tinygo.org
  [Goldmark Playground]: https://yuin.github.io/goldmark/playground
  [serde_json_path Sandbox]: https://serdejsonpath.live
  [privacy statement]: https://docs.github.com/en/site-policy/privacy-policies/github-general-privacy-statement
  [RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535.html
  [MIT License]: https://opensource.org/license/mit
  [Boxicons]: https://boxicons.com
