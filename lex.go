package jsonpath

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/smasher164/xid"
)

// token represents a single token in the input stream.
// Name: mnemonic name (numeric).
// Val: string value of the token from the original stream.
// Pos: position - offset from beginning of stream.
type token struct {
	// tok identifies the token, either a rune or a negative number
	// representing a special token.
	tok rune

	// val contains the string representation of the token. When tok is
	// goString, val will be a fully-parsed Go string. When tok is invalid,
	// val will be the error message.
	val string

	// pos represents the position of the start of the token.
	pos int
}

const (
	// Non-rune special tokens are negative.
	invalid = -(iota + 1)
	eof
	identifier
	integer
	number
	goString
	blankSpace
	boolTrue
	boolFalse
	jsonNull
)

// Special values.
const (
	// Numeric bases.
	decimal = 10
	hex     = 16

	// Token literals.
	nullByte  = 0x00
	backslash = '\\'

	// blanks selects blank space characters.
	blanks = 1<<'\t' | 1<<'\n' | 1<<'\r' | 1<<' '
)

// name returns the token name, a string for special tokens and a quoted rune
// or escaped rune.
func (tok token) name() string {
	switch tok.tok {
	case invalid:
		return "invalid"
	case eof:
		return "eof"
	case identifier:
		return "identifier"
	case integer:
		return "integer"
	case number:
		return "number"
	case goString:
		return "string"
	case blankSpace:
		return "blank space"
	default:
		return strconv.QuoteRune(tok.tok)
	}
}

// String returns a string representation of the token.
func (tok token) String() string {
	return fmt.Sprintf("Token{%v, %q, %v}", tok.name(), tok.val, tok.pos)
}

// err returns an error for invalid tokens and nil for all other tokens.
func (tok token) err() error {
	if tok.tok != invalid {
		return nil
	}
	return fmt.Errorf("%w: %v at %v", ErrPathParse, tok.val, tok.pos)
}

// errToken creates and returns an error token.
func (lex *lexer) errToken(pos int, msg string) token {
	// Set r to invalid so scan() ceases to scan.
	lex.r = invalid
	return token{invalid, msg, pos}
}

// lexer for [RFC 9535] JSONPath queries.
//
// Create a new lexer with NewLexer and then call NextToken repeatedly to get
// tokens from the stream. The lexer will return a token with the name EOF when
// done.
//
// Based on the public domain [TableGen lexer] by [Eli Bendersky].
//
// [RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535.html
// [TableGen lexer]: https://github.com/eliben/code-for-blog/blob/main/2014/tablegen-lexer-go/lexer-string/lexer.go
// [Eli Bendersky]: https://eli.thegreenplace.net
type lexer struct {
	buf string

	// Current rune.
	r rune

	// Position of the current rune in buf.
	rPos int

	// Position of the next rune in buf.
	nextPos int
}

// newLexer creates a new lexer for the given input.
func newLexer(buf string) *lexer {
	lex := lexer{buf, -1, 0, 0}

	// Prime the lexer by calling .next
	lex.next()
	return &lex
}

// scan returns the next token.
func (lex *lexer) scan() token {
	switch {
	case lex.r < 0:
		return token{eof, "", lex.rPos}
	case lex.r == '$':
		if isIdentRune(lex.peek(), 0) {
			return lex.scanIdentifier()
		}
	case isIdentRune(lex.r, 0):
		return lex.scanIdentifier()
	case isDigit(lex.r) || lex.r == '-':
		return lex.scanNumber()
	case lex.r == '"' || lex.r == '\'':
		return lex.scanString()
	case blanks&(1<<uint(lex.r)) != 0:
		return lex.scanBlankSpace()
	}

	ret := token{lex.r, "", lex.rPos}
	lex.next()
	return ret
}

// next advances the lexer's internal state to point to the next rune in the
// input.
func (lex *lexer) next() rune {
	if lex.nextPos < len(lex.buf) {
		lex.rPos = lex.nextPos
		r, w := rune(lex.buf[lex.nextPos]), 1

		if r >= utf8.RuneSelf {
			r, w = utf8.DecodeRuneInString(lex.buf[lex.nextPos:])
		}

		lex.nextPos += w
		lex.r = r
	} else {
		lex.rPos = len(lex.buf)
		lex.r = eof
	}

	return lex.r
}

// peek returns the next byte in the stream (the one after lex.r).
// Note: a single byte is peeked at - if there's a rune longer than a byte
// there, only its first byte is returned. Returns eof if there is no next
// byte.
func (lex *lexer) peek() rune {
	if lex.nextPos < len(lex.buf) {
		return rune(lex.buf[lex.nextPos])
	}
	return rune(eof)
}

// scanBlankSpace scan and returns a token of blank spaces.
func (lex *lexer) scanBlankSpace() token {
	startPos := lex.rPos
	for blanks&(1<<uint(lex.r)) != 0 {
		lex.next()
	}
	return token{blankSpace, lex.buf[startPos:lex.rPos], startPos}
}

// scanIdentifier scans an identifier, including escapes. lex.r should be the
// first rune in the identifier, and isIdentRune(lex.r, 0) should have already
// returned true.
func (lex *lexer) scanIdentifier() token {
	buf := new(strings.Builder)
	startPos := lex.rPos
	escaped := false

	// Scan the identifier as long as we have legit identifier runes.
	for isIdentRune(lex.r, 1) {
		switch lex.r {
		case backslash:
			// Handle escapes.
			if !lex.writeEscape(-1, buf) {
				return lex.errToken(lex.rPos, "invalid escape after backslash")
			}
			escaped = true
		default:
			buf.WriteRune(lex.r)
			lex.next()
		}
	}

	tok := token{identifier, buf.String(), startPos}
	if !escaped {
		// Set proper name for literals.
		switch tok.val {
		case "true":
			tok.tok = boolTrue
		case "false":
			tok.tok = boolFalse
		case "null":
			tok.tok = jsonNull
		}
	}
	return tok
}

// isIdentRune is a predicate controlling the characters accepted as the ith
// rune in an identifier. These follow JavaScript [identifier syntax], including
// support for \u0000 and \u{000000} unicode escapes.
//
// [identifier syntax]: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Lexical_grammar#identifiers
func isIdentRune(r rune, i int) bool {
	return r == '_' || r == '\\' || r == '$' || (i == 0 && xid.Start(r)) || (i > 0 && xid.Continue(r))
}

// scanNumber scans an integer or decimal number.
func (lex *lexer) scanNumber() token {
	startPos := lex.rPos

	// Start with integer.
	switch lex.r {
	case '0':
		next := lex.next()
		if next != '.' && next != 'e' {
			if isDigit(next) {
				// No leading zeros for integers.
				return lex.errToken(startPos, "invalid number literal")
			}
			// Standalone zero.
			return token{integer, "0", startPos}
		}
	case '-':
		// ["-"] DIGIT1 *DIGIT / (int / "-0") [ frac ] [ exp ]
		next := lex.next()
		switch {
		case next == '0':
			next := lex.peek()
			if next != '.' && next != 'e' {
				// -0 only for fractional and exponent numbers
				return lex.errToken(startPos, "invalid number literal")
			}
		case !isDigit(next):
			// Need digits after decimal.
			return lex.errToken(startPos, "invalid number literal")
		}
		fallthrough
	default:
		for isDigit(lex.r) {
			lex.next()
		}
	}

	// We have the integer part. Is there a decimal and/or exponent?
	switch lex.r {
	case '.':
		// Parse fraction: "." 1*DIGIT
		if !isDigit(lex.next()) {
			// No digits after decimal.
			return lex.errToken(startPos, "invalid number literal")
		}
		// Collect remaining digits.
		for isDigit(lex.r) {
			lex.next()
		}
		if lex.r == 'e' {
			// Exponent.
			return lex.scanExponent(startPos)
		}
		return token{number, lex.buf[startPos:lex.rPos], startPos}
	case 'e':
		// Exponent.
		return lex.scanExponent(startPos)
	default:
		// Just an integer.
		return token{integer, lex.buf[startPos:lex.rPos], startPos}
	}
}

// scanExponent scans the exponent part of a decimal number. lex.r should be
// set to 'e' and expect to be followed by an optional '-' or '+' and then one
// or more digits.
func (lex *lexer) scanExponent(startPos int) token {
	// Parse exponent: "e" [ "-" / "+" ] 1*DIGIT
	switch lex.next() {
	case '-', '+':
		if !isDigit(lex.next()) {
			// No digit after + or -
			return lex.errToken(startPos, "invalid number literal")
		}
	default:
		if !isDigit(lex.r) {
			// No digit after e
			return lex.errToken(startPos, "invalid number literal")
		}
	}

	// Consume remaining digits.
	for isDigit(lex.r) {
		lex.next()
	}
	return token{number, lex.buf[startPos:lex.rPos], startPos}
}

// scanString scans and parses a single- or double-quoted JavaScript string.
// Token.Val contains the parsed value. Returns an error token on error.
func (lex *lexer) scanString() token {
	startPos := lex.rPos
	q := lex.r
	lex.next()
	buf := new(strings.Builder)

NEXT:
	for lex.r > 0 {
		switch {
		case isUnescaped(lex.r, q):
			// Regular character.
			buf.WriteRune(lex.r)
			lex.next()
		case lex.r == '\\':
			if !lex.writeEscape(q, buf) {
				pos := lex.rPos
				lex.next()
				return lex.errToken(pos, "invalid escape after backslash")
			}
		default:
			// End of string or buffer.
			break NEXT
		}
	}

	if lex.r != q {
		return lex.errToken(lex.rPos, "unterminated string literal")
	}

	lex.next()
	return token{goString, buf.String(), startPos}
}

// writeEscape handles string escapes in the context of a string. Set q to '
// or " to indicate a single or double-quoted string context, respectively,
// and -1 for an identifier. lex.r should be set to the rune following the
// backslash that initiated the escape.
func (lex *lexer) writeEscape(q rune, buf *strings.Builder) bool {
	// Starting an escape sequence.
	next := lex.next()
	if r := unescape(next, q); r > 0 {
		// A single-character escape.
		buf.WriteRune(r)
		lex.next()
		return true
	}

	if next == '\u0075' { // uXXXX U+XXXX
		// \uXXXX unicode escape.
		if r := lex.parseUnicode(); r >= 0 {
			buf.WriteRune(r)
			lex.next()
			return true
		}
	}
	return false
}

// Returns true if r is a regular, non-escaped character. Pass q as " or ' to
// indicate the scanning of a double or single quotation string.
func isUnescaped(r rune, q rune) bool {
	switch {
	case r == q:
		return false
	case r >= '\u0020' && r <= '\u005b': // omit 0x5C \
		return true
	case r >= '\u005d' && r <= '\ud7ff': // skip surrogate code points
		return true
	case r >= '\ue000' && r <= '\U0010FFFF':
		return true
	default:
		return false
	}
}

// Returns the value corresponding to escape code r. Pass q as " or ' to
// indicate the scanning of a double or single quotation string.
func unescape(r rune, q rune) rune {
	switch r {
	case q: // ' or "
		return q
	case '\u0062': // b BS backspace U+0008
		return '\u0008'
	case '\u0066': // f FF form feed U+000C
		return '\u000c'
	case '\u006e': // n LF line feed U+000A
		return '\u000a'
	case '\u0072': // r CR carriage return U+000D
		return '\u000d'
	case '\u0074': // t HT horizontal tab U+0009
		return '\u0009'
	case '/': // / slash (solidus) U+002F
		return '/'
	case '\\': // \ backslash (reverse solidus) U+005C
		return '\\'
	default:
		return 0
	}
}

// Parses a \u unicode escape sequence. Returns invalid (-1) on error.
func (lex *lexer) parseUnicode() rune {
	if !isHexDigit(lex.next()) {
		return rune(invalid)
	}

	if lex.r != 'd' && lex.r != 'D' {
		// non-surrogate DIGIT / "A"/"B"/"C" / "E"/"F"
		return lex.scanUnicode()
	}

	switch lex.peek() {
	case '0', '1', '2', '3', '4', '5', '6', '7':
		// non-surrogate "D" %x30-37 2HEXDIG
		return lex.scanUnicode()
	}

	// potential high-surrogate "D" ("8"/"9"/"A"/"B") 2HEXDIG
	high := lex.scanUnicode()

	// Must be followed by \u
	if high < nullByte || lex.next() != '\\' || lex.next() != 'u' {
		return rune(invalid)
	}

	// potential low-surrogate "D" ("C"/"D"/"E"/"F") 2HEXDIG
	lex.next()
	low := lex.scanUnicode()
	if low < nullByte {
		return rune(invalid)
	}

	// Merge and return the surrogate pair, if valid.
	if dec := utf16.DecodeRune(high, low); dec != unicode.ReplacementChar {
		return dec
	}
	// Adjust position to start of second \u escape for error reporting.
	lex.rPos -= 3

	return rune(invalid)
}

// scanUnicode scans the current rune plus the next four and merges them into
// the resulting unicode rune. Returns invalid (-1) on error.
func (lex *lexer) scanUnicode() rune {
	rr := hexChar(lex.r)
	for i := range 3 {
		c := hexChar(lex.next())
		if c < nullByte {
			// Reset to before first rune for error reporting.
			lex.rPos -= i + 1
			return c
		}

		rr = rr*hex + c
	}
	return rr
}

// isHexDigit returns true if r represents a hex digit matching [0-9a-fA-F].
func isHexDigit(r rune) bool {
	return isDigit(r) || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f')
}

// hexChar returns the byte value corresponding to hex character c.
func hexChar(c rune) rune {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + decimal
	case 'A' <= c && c <= 'F':
		return c - 'A' + decimal
	default:
		return rune(invalid)
	}
}

// isDigit returns true if r is a digit ([0-9]).
func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}
