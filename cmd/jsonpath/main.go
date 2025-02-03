// Package main implements a simple command-line utility that allows one to extract
// data from an arbitrary JSON body that has been piped into it.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/theory/jsonpath"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:      "jsonpath",
		Usage:     "extracting data from JSON according to RFC-9535",
		UsageText: "jsonpath QUERY",
		Version:   gitrev(),
		Action:    parseAndPrint,
		Args:      true,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprint(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
}

func gitrev() string {
	version := "(git revision unavailable)"

	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, kv := range bi.Settings {
			if kv.Key == "vcs.revision" {
				version = kv.Value
			}
		}
	}

	return version
}

func parseAndPrint(ctx *cli.Context) error {
	// grab the provided jsonpath query
	q := ctx.Args().First()
	if q == "" {
		cli.ShowAppHelpAndExit(ctx, 1)
	}
	p := jsonpath.NewParser().MustParse(q)

	m, err := jsonToMap(os.Stdin)
	if err != nil {
		return fmt.Errorf("could not read JSON body from stdin: %w", err)
	}

	// apply q to map
	result := p.Select(m)

	// dump to output
	items, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("could not marshal results to JSON: %w", err)
	}
	fmt.Printf("%s\n", items) //nolint:forbidigo

	return nil
}

func jsonToMap(r io.Reader) (map[string]any, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("could not read JSON contents: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(b))
	var m map[string]any
	err = decoder.Decode(&m)
	if err != nil {
		return nil, fmt.Errorf("could not decode JSON to map: %w", err)
	}

	return m, nil
}
