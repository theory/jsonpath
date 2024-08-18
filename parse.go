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
		switch tok := lex.scan(); tok.tok {
		case '[':
			// Start of segment; scan selectors
			selectors, err := parseSelectors(lex)
			if err != nil {
				return nil, err
			}
			segs = append(segs, Child(selectors...))
		case '.':
			// Start of a name selector, wildcard, or descendant segment.
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
		case eof:
			// Done parsing.
			return NewQuery(segs), nil
		case blankSpace:
			if lex.r != eof {
				continue
			}
			fallthrough
		default:
			// No other token is valid.
			return nil, unexpected(tok)
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
			// No filter support yet.
			return nil, makeError(tok, "filter selectors not yet supported")
		case '*':
			selectors = append(selectors, Wildcard)
		case goString:
			selectors = append(selectors, Name(tok.val))
		case integer:
			// Index or slice?
			if segmentPeek(lex) == ':' {
				// Slice.
				slice, err := parseSlice(lex, tok)
				if err != nil {
					return nil, err
				}
				selectors = append(selectors, slice)
			} else {
				// Index.
				idx, err := parseInt(tok)
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
		switch segmentPeek(lex) {
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

func parseInt(tok token) (int64, error) {
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
			num, err := parseInt(tok)
			if err != nil {
				return SliceSelector{}, err
			}
			args[i] = int(num)
		default:
			// Nothing else allowed.
			return SliceSelector{}, unexpected(tok)
		}

		// What's next?
		next := segmentPeek(lex)
		if next == ']' || next == ',' {
			// We've reached the end.
			return Slice(args[0], args[1], args[2]), nil
		}
		tok = lex.scan()
	}

	// Never found the end of the slice.
	return SliceSelector{}, unexpected(tok)
}

// segmentPeek returns the next byte to be lexed, skipping any blank space.
func segmentPeek(lex *lexer) rune {
	// Skip blank space.
	if blanks&(1<<uint(lex.r)) != 0 {
		lex.scan()
	}
	return lex.r
}
