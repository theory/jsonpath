package registry

import (
	"errors"
	"fmt"
	"regexp"
	"regexp/syntax"
	"unicode/utf8"

	"github.com/theory/jsonpath/spec"
)

// checkLengthArgs checks the argument expressions to length() and returns an
// error if there is not exactly one expression that results in a
// [PathValue]-compatible value.
func checkLengthArgs(fea []spec.FunctionExprArg) error {
	if len(fea) != 1 {
		return fmt.Errorf("expected 1 argument but found %v", len(fea))
	}

	kind := fea[0].ResultType()
	if !kind.ConvertsTo(spec.PathValue) {
		return errors.New("cannot convert argument to ValueType")
	}

	return nil
}

// lengthFunc extracts the single argument passed in jv and returns its
// length. Panics if jv[0] doesn't exist or is not convertible to [ValueType].
//
//   - if jv[0] is nil, the result is nil
//   - If jv[0] is a string, the result is the number of Unicode scalar values
//     in the string.
//   - If jv[0] is a []any, the result is the number of elements in the slice.
//   - If jv[0] is an map[string]any, the result is the number of members in
//     the map.
//   - For any other value, the result is nil.
func lengthFunc(jv []spec.JSONPathValue) spec.JSONPathValue {
	v := spec.ValueFrom(jv[0])
	if v == nil {
		return nil
	}
	switch v := v.Value().(type) {
	case string:
		// Unicode scalar values
		return spec.Value(utf8.RuneCountInString(v))
	case []any:
		return spec.Value(len(v))
	case map[string]any:
		return spec.Value(len(v))
	default:
		return nil
	}
}

// checkCountArgs checks the argument expressions to count() and returns an
// error if there is not exactly one expression that results in a
// [PathNodes]-compatible value.
func checkCountArgs(fea []spec.FunctionExprArg) error {
	if len(fea) != 1 {
		return fmt.Errorf("expected 1 argument but found %v", len(fea))
	}

	kind := fea[0].ResultType()
	if !kind.ConvertsTo(spec.PathNodes) {
		return errors.New("cannot convert argument to PathNodes")
	}

	return nil
}

// countFunc implements the [RFC 9535]-standard count function. The result is
// a ValueType containing an unsigned integer for the number of nodes
// in jv[0]. Panics if jv[0] doesn't exist or is not convertible to
// [NodesType].
func countFunc(jv []spec.JSONPathValue) spec.JSONPathValue {
	return spec.Value(len(spec.NodesFrom(jv[0])))
}

// checkValueArgs checks the argument expressions to value() and returns an
// error if there is not exactly one expression that results in a
// [PathNodes]-compatible value.
func checkValueArgs(fea []spec.FunctionExprArg) error {
	if len(fea) != 1 {
		return fmt.Errorf("expected 1 argument but found %v", len(fea))
	}

	kind := fea[0].ResultType()
	if !kind.ConvertsTo(spec.PathNodes) {
		return errors.New("cannot convert argument to PathNodes")
	}

	return nil
}

// valueFunc implements the [RFC 9535]-standard value function. Panics if
// jv[0] doesn't exist or is not convertible to [NodesType]. Otherwise:
//
//   - If jv[0] contains a single node, the result is the value of the node.
//   - If jv[0] is empty or contains multiple nodes, the result is nil.
func valueFunc(jv []spec.JSONPathValue) spec.JSONPathValue {
	nodes := spec.NodesFrom(jv[0])
	if len(nodes) == 1 {
		return spec.Value(nodes[0])
	}
	return nil
}

// checkMatchArgs checks the argument expressions to match() and returns an
// error if there are not exactly two expressions that result in
// [PathValue]-compatible values.
func checkMatchArgs(fea []spec.FunctionExprArg) error {
	const matchArgLen = 2
	if len(fea) != matchArgLen {
		return fmt.Errorf("expected 2 arguments but found %v", len(fea))
	}

	for i, arg := range fea {
		kind := arg.ResultType()
		if !kind.ConvertsTo(spec.PathValue) {
			return fmt.Errorf("cannot convert argument %v to PathNodes", i+1)
		}
	}

	return nil
}

// matchFunc implements the [RFC 9535]-standard match function. If jv[0] and
// jv[1] evaluate to strings, the second is compiled into a regular expression with
// implied \A and \z anchors and used to match the first, returning LogicalTrue for
// a match and LogicalFalse for no match. Returns LogicalFalse if either jv value
// is not a string or if jv[1] fails to compile.
func matchFunc(jv []spec.JSONPathValue) spec.JSONPathValue {
	if v, ok := spec.ValueFrom(jv[0]).Value().(string); ok {
		if r, ok := spec.ValueFrom(jv[1]).Value().(string); ok {
			if rc := compileRegex(`\A` + r + `\z`); rc != nil {
				return spec.LogicalFrom(rc.MatchString(v))
			}
		}
	}
	return spec.LogicalFalse
}

// checkSearchArgs checks the argument expressions to search() and returns an
// error if there are not exactly two expressions that result in
// [PathValue]-compatible values.
func checkSearchArgs(fea []spec.FunctionExprArg) error {
	const searchArgLen = 2
	if len(fea) != searchArgLen {
		return fmt.Errorf("expected 2 arguments but found %v", len(fea))
	}

	for i, arg := range fea {
		kind := arg.ResultType()
		if !kind.ConvertsTo(spec.PathValue) {
			return fmt.Errorf("cannot convert argument %v to PathNodes", i+1)
		}
	}

	return nil
}

// searchFunc implements the [RFC 9535]-standard search function. If both jv[0]
// and jv[1] contain strings, the latter is compiled into a regular expression and used
// to match the former, returning LogicalTrue for a match and LogicalFalse for no
// match. Returns LogicalFalse if either value is not a string, or if jv[1]
// fails to compile.
func searchFunc(jv []spec.JSONPathValue) spec.JSONPathValue {
	if val, ok := spec.ValueFrom(jv[0]).Value().(string); ok {
		if r, ok := spec.ValueFrom(jv[1]).Value().(string); ok {
			if rc := compileRegex(r); rc != nil {
				return spec.LogicalFrom(rc.MatchString(val))
			}
		}
	}
	return spec.LogicalFalse
}

// compileRegex compiles str into a regular expression or returns an error. To
// comply with RFC 9485 regular expression semantics, all instances of "." are
// replaced with "[^\n\r]". This sadly requires compiling the regex twice:
// once to produce an AST to replace "." nodes, and a second time for the
// final regex.
func compileRegex(str string) *regexp.Regexp {
	// First compile AST and replace "." with [^\n\r].
	// https://www.rfc-editor.org/rfc/rfc9485.html#name-pcre-re2-and-ruby-regexps
	r, err := syntax.Parse(str, syntax.Perl|syntax.DotNL)
	if err != nil {
		// Could use some way to log these errors rather than failing silently.
		return nil
	}

	replaceDot(r)
	re, _ := regexp.Compile(r.String())
	return re
}

//nolint:gochecknoglobals
var clrf, _ = syntax.Parse("[^\n\r]", syntax.Perl)

// replaceDot recurses re to replace all "." nodes with "[^\n\r]" nodes.
func replaceDot(re *syntax.Regexp) {
	if re.Op == syntax.OpAnyChar {
		*re = *clrf
	} else {
		for _, re := range re.Sub {
			replaceDot(re)
		}
	}
}
