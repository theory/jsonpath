package jsonpath

import (
	"errors"
	"fmt"
	"strconv"
)

// ErrPathParse errors are returned for path parse errors.
var ErrPathParse = errors.New("jsonpath")

func makeError(tok token, msg string) error {
	return fmt.Errorf("%w: %v at position %v", ErrPathParse, msg, tok.pos+1)
}

// unexpected creates and returns an error for an unexpected token. For
// invalid tokens, the error will be as returned by the lexer. Otherwise, the
// error will "unexpected: $name".
func unexpected(tok token) error {
	if tok.tok == invalid {
		// Lex error message in the token value.
		return makeError(tok, tok.val)
	}
	return makeError(tok, "unexpected "+tok.name())
}

// Parse parses path, a JSON Path query string, into a Path. Returns a
// PathParseError on parse failure.
func Parse(path string) (*Path, error) {
	lex := newLexer(path)
	tok := lex.scan()

	switch tok.tok {
	case '$':
		// All path queries must start with $.
		q, err := parseQuery(lex)
		if err != nil {
			return nil, err
		}
		// Should have scanned to the end of input.
		if lex.r != eof {
			return nil, unexpected(lex.scan())
		}
		return New(q), nil
	case eof:
		// The token contained nothing.
		return nil, fmt.Errorf("%w: unexpected end of input", ErrPathParse)
	default:
		return nil, unexpected(tok)
	}
}

// parseQuery parses a query expression. lex.r should be set to $ (or,
// eventually, @) before calling. Returns the parsed Query.
func parseQuery(lex *lexer) (*Query, error) {
	segs := []*Segment{}
	for {
		switch {
		case lex.r == '[':
			// Start of segment; scan selectors
			lex.scan()
			selectors, err := parseSelectors(lex)
			if err != nil {
				return nil, err
			}
			segs = append(segs, Child(selectors...))
		case lex.r == '.':
			// Start of a name selector, wildcard, or descendant segment.
			lex.scan()
			if lex.r == '.' {
				// Consume `.` and parse descendant.
				lex.scan()
				seg, err := parseDescendant(lex)
				if err != nil {
					return nil, err
				}
				segs = append(segs, seg)
				continue
			}
			// Child segment with a name or wildcard selector.
			sel, err := parseNameOrWildcard(lex)
			if err != nil {
				return nil, err
			}
			segs = append(segs, Child(sel))
		case lex.isBlankSpace(lex.r):
			switch lex.peekPastBlankSpace() {
			case '.', '[':
				lex.scanBlankSpace()
				continue
			}
			fallthrough
		default:
			// Done parsing.
			return NewQuery(segs), nil
		}
	}
}

// parseNameOrWildcard parses a name or '*' wildcard selector. Returns the
// parsed Selector.
//
//nolint:ireturn
func parseNameOrWildcard(lex *lexer) (Selector, error) {
	switch tok := lex.scan(); tok.tok {
	case identifier:
		return Name(tok.val), nil
	case '*':
		return Wildcard, nil
	default:
		return nil, unexpected(tok)
	}
}

// parseDescendant parses a ".." descendant segment, which may be a bracketed
// segment or a wildcard or name selector segment. Returns the parsed Segment.
func parseDescendant(lex *lexer) (*Segment, error) {
	switch tok := lex.scan(); tok.tok {
	case '[':
		// Start of segment; scan selectors
		selectors, err := parseSelectors(lex)
		if err != nil {
			return nil, err
		}
		return Descendant(selectors...), nil
	case identifier:
		return Descendant(Name(tok.val)), nil
	case '*':
		return Descendant(Wildcard), nil
	default:
		return nil, unexpected(tok)
	}
}

// makeNumErr converts strconv.NumErrors to jsonpath errors.
func makeNumErr(tok token, err error) error {
	var numError *strconv.NumError
	if errors.As(err, &numError) {
		return makeError(tok, fmt.Sprintf(
			"cannot parse %q, %v",
			numError.Num, numError.Err.Error(),
		))
	}
	return makeError(tok, err.Error())
}

// parseSelectors parses Selectors from a bracket segment. lex.r should be '['
// before calling. Returns the Selectors parsed.
func parseSelectors(lex *lexer) ([]Selector, error) {
	selectors := []Selector{}
	for {
		switch tok := lex.scan(); tok.tok {
		case '?':
			filter, err := parseFilter(lex)
			if err != nil {
				return nil, err
			}
			selectors = append(selectors, filter)
		case '*':
			selectors = append(selectors, Wildcard)
		case goString:
			selectors = append(selectors, Name(tok.val))
		case integer:
			// Index or slice?
			if lex.skipBlankSpace() == ':' {
				// Slice.
				slice, err := parseSlice(lex, tok)
				if err != nil {
					return nil, err
				}
				selectors = append(selectors, slice)
			} else {
				// Index.
				idx, err := parsePathInt(tok)
				if err != nil {
					return nil, err
				}
				selectors = append(selectors, Index(idx))
			}
		case ':':
			// Slice.
			slice, err := parseSlice(lex, tok)
			if err != nil {
				return nil, err
			}
			selectors = append(selectors, slice)
		case blankSpace:
			// Skip.
			continue
		default:
			return nil, unexpected(tok)
		}

		// Successfully parsed a selector. What's next?
		switch lex.skipBlankSpace() {
		case ',':
			// Consume the comma.
			lex.scan()
		case ']':
			// Consume and return.
			lex.scan()
			return selectors, nil
		default:
			// Anything else is an error.
			return nil, unexpected(lex.scan())
		}
	}
}

// parsePathInt parses an integer as used in index values and steps, which must be
// within the interval [-(253)+1, (253)-1].
func parsePathInt(tok token) (int64, error) {
	if tok.val == "-0" {
		return 0, makeError(tok, fmt.Sprintf(
			"invalid integer path value %q", tok.val,
		))
	}
	idx, err := strconv.ParseInt(tok.val, 10, 64)
	if err != nil {
		return 0, makeNumErr(tok, err)
	}
	const (
		minVal = -1<<53 + 1
		maxVal = 1<<53 - 1
	)
	if idx > maxVal || idx < minVal {
		return 0, makeError(tok, fmt.Sprintf(
			"cannot parse %q, value out of range",
			tok.val,
		))
	}
	return idx, nil
}

// parseSlice parses a slice selector, <start>:<end>:<step>. Returns the
// parsed SliceSelector.
func parseSlice(lex *lexer, tok token) (SliceSelector, error) {
	var args [3]any

	// Parse the three parts: start, end, and step.
	i := 0
	for i < 3 {
		switch tok.tok {
		case ':':
			// Skip to the next index.
			i++
		case integer:
			// Parse the integer.
			num, err := parsePathInt(tok)
			if err != nil {
				return SliceSelector{}, err
			}
			args[i] = int(num)
		default:
			// Nothing else allowed.
			return SliceSelector{}, unexpected(tok)
		}

		// What's next?
		next := lex.skipBlankSpace()
		if next == ']' || next == ',' {
			// We've reached the end.
			return Slice(args[0], args[1], args[2]), nil
		}
		tok = lex.scan()
	}

	// Never found the end of the slice.
	return SliceSelector{}, unexpected(tok)
}

// parseFilter parses a [Filter] from Lex. A [Filter] consists of a single
// [LogicalOrExpr] (logical-or-expr).
func parseFilter(lex *lexer) (*Filter, error) {
	lor, err := parseLogicalOrExpr(lex)
	if err != nil {
		return nil, err
	}
	return &Filter{lor}, nil
}

// parseLogicalOrExpr parses a [LogicalOrExpr] from lex. A [LogicalOrExpr] is
// made up of one or more [LogicalAndExpr] (logical-and-expr) separated by
// "||".
func parseLogicalOrExpr(lex *lexer) (LogicalOrExpr, error) {
	ands := []LogicalAndExpr{}
	land, err := parseLogicalAndExpr(lex)
	if err != nil {
		return nil, err
	}

	ands = append(ands, land)
	lex.scanBlankSpace()
	for {
		if lex.r != '|' {
			break
		}
		lex.scan()
		next := lex.scan()
		if next.tok != '|' {
			return nil, makeError(next, fmt.Sprintf("expected '|' but found %v", next.name()))
		}
		land, err := parseLogicalAndExpr(lex)
		if err != nil {
			return nil, err
		}
		ands = append(ands, land)
	}

	return LogicalOrExpr(ands), nil
}

// parseLogicalAndExpr parses a [LogicalAndExpr] from lex. A [LogicalAndExpr]
// is made up of one or more [BasicExpr]s (basic-expr) separated by "&&".
func parseLogicalAndExpr(lex *lexer) (LogicalAndExpr, error) {
	expr, err := parseBasicExpr(lex)
	if err != nil {
		return nil, err
	}

	ors := []BasicExpr{expr}
	lex.scanBlankSpace()
	for {
		if lex.r != '&' {
			break
		}
		lex.scan()
		next := lex.scan()
		if next.tok != '&' {
			return nil, makeError(next, fmt.Sprintf("expected '&' but found %v", next.name()))
		}
		expr, err := parseBasicExpr(lex)
		if err != nil {
			return nil, err
		}
		ors = append(ors, expr)
	}

	return LogicalAndExpr(ors), nil
}

// parseBasicExpr parses a [BasicExpr] from lex. A [BasicExpr] may be a
// parenthesized expression (paren-expr), comparison expression
// (comparison-expr), or test expression (test-expr).
//
//nolint:ireturn
func parseBasicExpr(lex *lexer) (BasicExpr, error) {
	// Consume blank space.
	lex.skipBlankSpace()

	tok := lex.scan()
	switch tok.tok {
	case '!':
		if lex.skipBlankSpace() == '(' {
			// paren-expr
			lex.scan()
			return parseNotParenExpr(lex)
		}

		next := lex.scan()
		if next.tok == identifier {
			// test-expr or comparison-expr
			f, err := parseFunction(next, lex)
			if err != nil {
				return nil, err
			}
			return NotFuncExpr{f}, nil
		}

		// test-expr or comparison-expr
		return parseNotExistsExpr(next, lex)
	case '(':
		return parseParenExpr(lex)
	case goString, integer, number, boolFalse, boolTrue, jsonNull:
		// comparison-expr
		left, err := parseLiteral(tok)
		if err != nil {
			return nil, err
		}
		return parseComparableExpr(left, lex)
	case identifier:
		if lex.r == '(' {
			return parseFunctionFilterExpr(tok, lex)
		}
	case '@', '$':
		q, err := parseFilterQuery(tok, lex)
		if err != nil {
			return nil, err
		}

		if sing := q.singular(); sing != nil {
			switch lex.skipBlankSpace() {
			// comparison-expr
			case '=', '!', '<', '>':
				return parseComparableExpr(sing, lex)
			}
		}
		return &ExistExpr{q}, nil
	}

	return nil, unexpected(tok)
}

// parseFunctionFilterExpr parses a [BasicExpr] (basic-expr) that starts with
// ident, which must be an identifier token that's expected to be the name of
// a function. The return value will be either a [FunctionExpr]
// (function-expr), if the function return value is a logical (boolean) value.
// Otherwise it will be a [ComparisonExpr] (comparison-expr), as long as the
// function call is compared to another expression. Any other configuration
// returns an error.
//
//nolint:ireturn
func parseFunctionFilterExpr(ident token, lex *lexer) (BasicExpr, error) {
	f, err := parseFunction(ident, lex)
	if err != nil {
		return nil, err
	}

	if f.fn.ResultType == FuncLogical {
		return f, nil
	}

	switch lex.skipBlankSpace() {
	case '=', '!', '<', '>':
		// comparison-expr
		return parseComparableExpr(f, lex)
	}

	return nil, makeError(lex.scan(), "missing comparison to function result")
}

// parseNotExistsExpr parses a [NotExistsExpr] (non-existence) from lex.
func parseNotExistsExpr(tok token, lex *lexer) (*NotExistsExpr, error) {
	q, err := parseFilterQuery(tok, lex)
	if err != nil {
		return nil, err
	}
	return &NotExistsExpr{q}, nil
}

// parseFilterQuery parses a *Query (rel-query / jsonpath-query) from lex.
func parseFilterQuery(tok token, lex *lexer) (*Query, error) {
	q, err := parseQuery(lex)
	if err != nil {
		return nil, err
	}
	q.root = tok.tok == '$'
	return q, nil
}

// parseLogicalOrExpr parses a [LogicalOrExpr] from lex, which should return
// the next token after '(' from scan(). Returns an error if the expression
// does not end with a closing ')'.
func parseInnerParenExpr(lex *lexer) (LogicalOrExpr, error) {
	expr, err := parseLogicalOrExpr(lex)
	if err != nil {
		return nil, err
	}

	// Make sure we ended on a parenthesis.
	next := lex.scan()
	if next.tok != ')' {
		return nil, makeError(
			next, fmt.Sprintf("expected ')' but found %v", next.name()),
		)
	}

	return expr, nil
}

// parseParenExpr parses a [ParenExpr] (paren-expr) expression from lex, which
// should return the next token after '(' from scan(). Returns an error if the
// expression does not end with a closing ')'.
func parseParenExpr(lex *lexer) (*ParenExpr, error) {
	expr, err := parseInnerParenExpr(lex)
	if err != nil {
		return nil, err
	}
	return &ParenExpr{expr}, nil
}

// parseParenExpr parses a [NotParenExpr] expression (logical-not-op
// paren-expression) from lex, which should return the next token after '('
// from scan(). Returns an error if the expression does not end with a closing
// ')'.
func parseNotParenExpr(lex *lexer) (*NotParenExpr, error) {
	expr, err := parseInnerParenExpr(lex)
	if err != nil {
		return nil, err
	}
	return &NotParenExpr{expr}, nil
}

// parseFunction parses a function named tok.val from lex. tok should be the
// token just before the next call to lex.scan, and must be an identifier
// token naming the function. Returns an error if the function is not found in
// the registry or if arguments are invalid for the function.
func parseFunction(tok token, lex *lexer) (*FunctionExpr, error) {
	function := GetFunction(tok.val)
	if function == nil {
		return nil, makeError(tok, fmt.Sprintf("unknown function %q", tok.val))
	}

	paren := lex.scan() // Drop (
	args, err := parseFunctionArgs(lex)
	if err != nil {
		return nil, err
	}

	// Validate the functions.
	if err := function.Validate(args); err != nil {
		// Return the error starting at the opening parenthesis.
		return nil, makeError(paren, err.Error())
	}

	return &FunctionExpr{fn: function, args: args}, nil
}

// parseFunctionArgs parses the comma-delimited arguments to a function from
// lex. Arguments may be one of literal, filter-query (including
// singular-query), logical-expr, or function-expr.
func parseFunctionArgs(lex *lexer) ([]FunctionExprArg, error) {
	res := []FunctionExprArg{}
	for {
		switch tok := lex.scan(); tok.tok {
		case goString, integer, number, boolFalse, boolTrue, jsonNull:
			// literal
			val, err := parseLiteral(tok)
			if err != nil {
				return nil, err
			}
			res = append(res, val)
		case '@', '$':
			// filter-query
			q, err := parseFilterQuery(tok, lex)
			if err != nil {
				return nil, err
			}

			res = append(res, q.expression())

		case identifier:
			// function-expr

			if lex.skipBlankSpace() != '(' {
				return nil, unexpected(tok)
			}
			f, err := parseFunction(tok, lex)
			if err != nil {
				return nil, err
			}
			res = append(res, f)
		case blankSpace:
			// Skip.
			continue
		case ')':
			// All done.
			return res, nil
		case '!', '(':
			ors, err := parseLogicalOrExpr(lex)
			if err != nil {
				return nil, err
			}
			res = append(res, ors)
		}

		// Successfully parsed an argument. What's next?
		switch lex.skipBlankSpace() {
		case ',':
			// Consume the comma.
			lex.scan()
		case ')':
			// Consume and return.
			lex.scan()
			return res, nil
		default:
			// Anything else is an error.
			return nil, unexpected(lex.scan())
		}
	}
}

// parseLiteral parses the literal value from tok into native Go values and
// returns them as literalArg. tok.tok must be one of goString, integer,
// number, boolFalse, boolTrue, or jsonNull.
func parseLiteral(tok token) (*literalArg, error) {
	switch tok.tok {
	case goString:
		return &literalArg{tok.val}, nil
	case integer:
		integer, err := strconv.ParseInt(tok.val, 10, 64)
		if err != nil {
			return nil, makeNumErr(tok, err)
		}
		return &literalArg{integer}, nil
	case number:
		num, err := strconv.ParseFloat(tok.val, 64)
		if err != nil {
			return nil, makeNumErr(tok, err)
		}
		return &literalArg{num}, nil
	case boolTrue:
		return &literalArg{true}, nil
	case boolFalse:
		return &literalArg{false}, nil
	case jsonNull:
		return &literalArg{nil}, nil
	default:
		return nil, unexpected(tok)
	}
}

// parseComparableExpr parses a [ComparisonExpr] (comparison-expr) from lex.
func parseComparableExpr(left CompVal, lex *lexer) (*ComparisonExpr, error) {
	// Skip blank space.
	lex.skipBlankSpace()

	op, err := parseCompOp(lex)
	if err != nil {
		return nil, err
	}

	// Skip blank space.
	lex.skipBlankSpace()

	right, err := parseComparableVal(lex.scan(), lex)
	if err != nil {
		return nil, err
	}

	return &ComparisonExpr{left, op, right}, nil
}

// parseComparableVal parses a [CompVal] (comparable) from lex.
//
//nolint:ireturn
func parseComparableVal(tok token, lex *lexer) (CompVal, error) {
	switch tok.tok {
	case goString, integer, number, boolFalse, boolTrue, jsonNull:
		// literal
		return parseLiteral(tok)
	case '@', '$':
		// singular-query
		return parseSingularQuery(tok, lex)
	case identifier:
		// function-expr
		if lex.r != '(' {
			return nil, unexpected(tok)
		}
		f, err := parseFunction(tok, lex)
		if err != nil {
			return nil, err
		}
		if f.fn.ResultType == FuncLogical {
			return nil, makeError(tok, "cannot compare result of logical function")
		}
		return f, nil
	default:
		return nil, unexpected(tok)
	}
}

// parseCompOp pares a [CompOp] (comparison-op) from lex.
func parseCompOp(lex *lexer) (CompOp, error) {
	tok := lex.scan()
	switch tok.tok {
	case '=':
		if lex.r == '=' {
			lex.scan()
			return EqualTo, nil
		}
	case '!':
		if lex.r == '=' {
			lex.scan()
			return NotEqualTo, nil
		}
	case '<':
		if lex.r == '=' {
			lex.scan()
			return LessThanEqualTo, nil
		}
		return LessThan, nil
	case '>':
		if lex.r == '=' {
			lex.scan()
			return GreaterThanEqualTo, nil
		}
		return GreaterThan, nil
	}

	return 0, makeError(tok, "invalid comparison operator")
}

// parseSingularQuery parses a [singularQuery] (singular-query) from lex. A
// singular query consists only of single-selector nodes.
func parseSingularQuery(startToken token, lex *lexer) (*singularQuery, error) {
	selectors := []Selector{}
	for {
		switch lex.r {
		case '[':
			// Index or name selector.
			lex.skipBlankSpace()
			lex.scan()
			switch tok := lex.scan(); tok.tok {
			case goString:
				selectors = append(selectors, Name(tok.val))
			case integer:
				idx, err := parsePathInt(tok)
				if err != nil {
					return nil, err
				}
				selectors = append(selectors, Index(idx))
			default:
				return nil, unexpected(tok)
			}
			// Look for closing bracket.
			lex.skipBlankSpace()
			tok := lex.scan()
			if tok.tok != ']' {
				return nil, unexpected(tok)
			}
		case '.':
			// Start of a name selector.
			lex.scan()
			tok := lex.scan()
			if tok.tok != identifier {
				return nil, unexpected(tok)
			}
			selectors = append(selectors, Name(tok.val))
		default:
			// Done parsing.
			return &singularQuery{
				selectors: selectors,
				relative:  startToken.tok == '@',
			}, nil
		}
	}
}
