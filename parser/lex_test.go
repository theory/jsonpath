package parser

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanString(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		tok  token
	}{
		{
			name: "empty_dq",
			in:   ``,
			tok:  token{goString, "", 0},
		},
		{
			name: "one_char_dq",
			in:   `x`,
			tok:  token{goString, "x", 0},
		},
		{
			name: "multi_char_dq",
			in:   `hello there`,
			tok:  token{goString, "hello there", 0},
		},
		{
			name: "utf8_dq",
			in:   `hello ø`,
			tok:  token{goString, "hello ø", 0},
		},
		{
			name: "emoji_dq",
			in:   `hello 👋🏻`,
			tok:  token{goString, "hello 👋🏻", 0},
		},
		{
			name: "emoji_escape_q_dq",
			in:   `hello \"👋🏻\" there`,
			tok:  token{goString, `hello "👋🏻" there`, 0},
		},
		{
			name: "escapes_dq",
			in:   `\b\f\n\r\t\/\\`,
			tok:  token{goString, "\b\f\n\r\t/\\", 0},
		},
		{
			name: "unicode_dq",
			in:   `fo\u00f8`,
			tok:  token{goString, "foø", 0},
		},
		{
			name: "non_surrogate_start_dq",
			in:   `\u00f8y vey`,
			tok:  token{goString, "øy vey", 0},
		},
		{
			name: "non_surrogate_end_dq",
			in:   `fo\u00f8`,
			tok:  token{goString, "foø", 0},
		},
		{
			name: "non_surrogate_mid_dq",
			in:   `fo\u00f8 bar`,
			tok:  token{goString, "foø bar", 0},
		},
		{
			name: "non_surrogate_start_d_dq",
			in:   `\ud3c0 yep`,
			tok:  token{goString, "폀 yep", 0},
		},
		{
			name: "non_surrogate_end_d_dq",
			in:   `got \ud3c0`,
			tok:  token{goString, "got 폀", 0},
		},
		{
			name: "non_surrogate_mid_d_dq",
			in:   `got \ud3c0 yep`,
			tok:  token{goString, "got 폀 yep", 0},
		},
		{
			name: "surrogate_pair_dq",
			in:   `\uD834\uDD1E`,
			tok:  token{goString, "\U0001D11E", 0},
		},
		{
			name: "surrogate_pair_start_dq",
			in:   `\uD834\uDD1E yep`,
			tok:  token{goString, "\U0001D11E yep", 0},
		},
		{
			name: "surrogate_pair_end_dq",
			in:   `go \uD834\uDD1E`,
			tok:  token{goString, "go \U0001D11E", 0},
		},
		{
			name: "invalid_unicode_dq",
			in:   `fo\u0f8`,
			tok:  token{invalid, "invalid escape after backslash", 5},
		},
		{
			name: "invalid_non_surrogate_start_d_dq",
			in:   `\ud30 yep`,
			tok:  token{invalid, "invalid escape after backslash", 3},
		},
		{
			name: "invalid_surrogate_high_dq",
			in:   `\uD8x4\uDD1E`,
			tok:  token{invalid, "invalid escape after backslash", 3},
		},
		{
			name: "invalid_surrogate_low_dq",
			in:   `\uD834\uDDxE`,
			tok:  token{invalid, "invalid escape after backslash", 9},
		},
		{
			name: "surrogate_low_not_d_dq",
			in:   `\uD834\uED1E`,
			tok:  token{invalid, "invalid escape after backslash", 9},
		},
		{
			name: "surrogate_low_not_a_f_dq",
			in:   `\uD834\ud11E`,
			tok:  token{invalid, "invalid escape after backslash", 9},
		},
		{
			name: "no_surrogate_low_dq",
			in:   `\uD834 oops`,
			tok:  token{invalid, "invalid escape after backslash", 7},
		},
		{
			name: "bad_escape_dq",
			in:   `left \7 right`,
			tok:  token{invalid, "invalid escape after backslash", 7},
		},
		{
			name: "unicode_not_hex_dq",
			in:   `hi \ux234 oops`,
			tok:  token{invalid, "invalid escape after backslash", 6},
		},
		{
			name: "dollar_start",
			in:   `$xyz`,
			tok:  token{goString, "$xyz", 0},
		},
		{
			name: "dollar_end",
			in:   `xyz$`,
			tok:  token{goString, "xyz$", 0},
		},
		{
			name: "dollar_mid",
			in:   `xy$z`,
			tok:  token{goString, "xy$z", 0},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			// Test double-quoted.
			lex := newLexer(`"` + tc.in + `"`)
			a.Equal('"', lex.r)
			a.Equal(tc.tok, lex.scanString())

			// Test single-quoted.
			lex = newLexer(`'` + strings.ReplaceAll(tc.in, `"`, `'`) + `'`)
			a.Equal('\'', lex.r)
			if tc.tok.tok != invalid {
				tc.tok.val = strings.ReplaceAll(tc.tok.val, `"`, `'`)
			}
			a.Equal(tc.tok, lex.scanString())

			// Test identifier.
			if tc.in != "" && !strings.ContainsAny(tc.in, "\\$") {
				lex := newLexer(strings.ReplaceAll(tc.in, ` `, `_`))
				a.Equal(rune(tc.in[0]), lex.r)
				if tc.tok.tok == invalid {
					tc.tok.pos--
				} else {
					tc.tok.val = strings.ReplaceAll(tc.tok.val, ` `, `_`)
				}
				if tc.tok.tok == goString {
					tc.tok.tok = identifier
				}
				a.Equal(tc.tok, lex.scanIdentifier())
			}
		})
	}

	// Test unclosed strings.
	t.Run("unclosed", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		a.Equal(rune(invalid), newLexer(`"food`).scanString().tok)
		a.Equal(rune(invalid), newLexer(`'food`).scanString().tok)
	})
}

func TestScanIdentifier(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		tok  token
	}{
		{
			name: "with_emoji",
			in:   "say_😀",
			tok:  token{identifier, "say_😀", 0},
		},
		{
			name: "with_surrogate_pair",
			in:   "say_\U0001D11E",
			tok:  token{identifier, "say_𝄞", 0},
		},
		{
			name: "newline",
			in:   "say\n",
			tok:  token{identifier, "say", 0},
		},
		{
			name: "linefeed",
			in:   "xxx\f",
			tok:  token{identifier, "xxx", 0},
		},
		{
			name: "return",
			in:   "abx_xyx\ryup",
			tok:  token{identifier, "abx_xyx", 0},
		},
		{
			name: "whitespace",
			in:   "go on",
			tok:  token{identifier, "go", 0},
		},
		{
			name: "true",
			in:   "true",
			tok:  token{boolTrue, "true", 0},
		},
		{
			name: "false",
			in:   "false",
			tok:  token{boolFalse, "false", 0},
		},
		{
			name: "null",
			in:   "null",
			tok:  token{jsonNull, "null", 0},
		},
		{
			name: "true_stop_at_escaped",
			in:   `tru\u0065`,
			tok:  token{identifier, "tru", 0},
		},
		{
			name: "false_stop_at_escaped",
			in:   `fals\u0065`,
			tok:  token{identifier, "fals", 0},
		},
		{
			name: "null_stop_at_escaped",
			in:   `n\u0075ll`,
			tok:  token{identifier, "n", 0},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			lex := newLexer(tc.in)
			assert.Equal(t, tc.tok, lex.scanIdentifier())
		})
	}
}

func TestScanNumber(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		tok  token
		num  any
	}{
		{
			name: "zero",
			in:   "0",
			tok:  token{integer, "0", 0},
			num:  int64(0),
		},
		{
			name: "zero_and_more",
			in:   "0 say",
			tok:  token{integer, "0", 0},
			num:  int64(0),
		},
		{
			name: "one",
			in:   "1",
			tok:  token{integer, "1", 0},
			num:  int64(1),
		},
		{
			name: "nine",
			in:   "9",
			tok:  token{integer, "9", 0},
			num:  int64(9),
		},
		{
			name: "twelve",
			in:   "12",
			tok:  token{integer, "12", 0},
			num:  int64(12),
		},
		{
			name: "hundred",
			in:   "100",
			tok:  token{integer, "100", 0},
			num:  int64(100),
		},
		{
			name: "max_int",
			in:   strconv.FormatInt(math.MaxInt64, 10),
			tok:  token{integer, strconv.FormatInt(math.MaxInt64, 10), 0},
			num:  int64(math.MaxInt64),
		},
		{
			name: "neg_one",
			in:   "-1",
			tok:  token{integer, "-1", 0},
			num:  int64(-1),
		},
		{
			name: "neg_42",
			in:   "-42",
			tok:  token{integer, "-42", 0},
			num:  int64(-42),
		},
		{
			name: "neg_zero",
			in:   "-0 oops",
			tok:  token{integer, "-0", 0},
			num:  int64(0),
		},
		{
			name: "leading_zero",
			in:   "032",
			tok:  token{invalid, "invalid number literal", 0},
		},
		{
			name: "zero_frac",
			in:   "0.1",
			tok:  token{number, "0.1", 0},
			num:  float64(0.1),
		},
		{
			name: "zero_frac_more",
			in:   "0.09323200/",
			tok:  token{number, "0.09323200", 0},
			num:  float64(0.093232),
		},
		{
			name: "more_frac",
			in:   "42.234853+",
			tok:  token{number, "42.234853", 0},
			num:  float64(42.234853),
		},
		{
			name: "neg_frac",
			in:   "-42.734/",
			tok:  token{number, "-42.734", 0},
			num:  float64(-42.734),
		},
		{
			name: "neg_zero_frac",
			in:   "-0.23",
			tok:  token{number, "-0.23", 0},
			num:  float64(-0.23),
		},
		{
			name: "double_zero_frac",
			in:   "01.23",
			tok:  token{invalid, "invalid number literal", 0},
		},
		{
			name: "neg_double_zero_frac",
			in:   "-01.23",
			tok:  token{invalid, "invalid number literal", 0},
		},
		{
			name: "missing_frac",
			in:   "42.x",
			tok:  token{invalid, "invalid number literal", 0},
		},
		{
			name: "missing_neg_frac",
			in:   "-42.x",
			tok:  token{invalid, "invalid number literal", 0},
		},
		{
			name: "zero_exp",
			in:   "0e12",
			tok:  token{number, "0e12", 0},
			num:  float64(0e12),
		},
		{
			name: "numb_exp",
			in:   "42E124",
			tok:  token{number, "42E124", 0},
			num:  float64(42e124),
		},
		{
			name: "neg_zero_exp",
			in:   "-0e123",
			tok:  token{number, "-0e123", 0},
			num:  float64(-0e123),
		},
		{
			name: "neg_exp",
			in:   "-42E123",
			tok:  token{number, "-42E123", 0},
			num:  float64(-42e123),
		},
		{
			name: "lead_zero_exp",
			in:   "00e12",
			tok:  token{invalid, "invalid number literal", 0},
		},
		{
			name: "exp_plus",
			in:   "99e+123",
			tok:  token{number, "99e+123", 0},
			num:  float64(99e+123),
		},
		{
			name: "exp_minus",
			in:   "99e-01234",
			tok:  token{number, "99e-01234", 0},
			num:  float64(99e-01234),
		},
		{
			name: "exp_decimal",
			in:   "12.32E3",
			tok:  token{number, "12.32E3", 0},
			num:  float64(12.32e3),
		},
		{
			name: "exp_plus_no_digits",
			in:   "99e++",
			tok:  token{invalid, "invalid number literal", 0},
		},
		{
			name: "exp_minus_no_digits",
			in:   "99e-x lol",
			tok:  token{invalid, "invalid number literal", 0},
		},
		{
			name: "neg_no_digits",
			in:   "-lol",
			tok:  token{invalid, "invalid number literal", 0},
		},
		{
			name: "exp_no_digits",
			in:   "42eek",
			tok:  token{invalid, "invalid number literal", 0},
		},
		// https://go.dev/ref/spec#Integer_literals
		{
			name: "integer",
			in:   "42",
			tok:  token{integer, "42", 0},
			num:  int64(42),
		},
		{
			name: "long_int",
			in:   "170141183460469231731687303715884105727",
			tok:  token{integer, "170141183460469231731687303715884105727", 0},
			num:  false, // integer too large, will not parse
		},
		// https://go.dev/ref/spec#Floating-point_literals
		{
			name: "float",
			in:   "72.40",
			tok:  token{number, "72.40", 0},
			num:  float64(72.4),
		},
		{
			name: "float_2",
			in:   "2.71828",
			tok:  token{number, "2.71828", 0},
			num:  float64(2.71828),
		},
		{
			name: "float_3",
			in:   "6.67428e-11",
			tok:  token{number, "6.67428e-11", 0},
			num:  float64(6.67428e-11),
		},
		{
			name: "float_4",
			in:   "1e6",
			tok:  token{number, "1e6", 0},
			num:  float64(1e6),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			r := require.New(t)

			lex := newLexer(tc.in)
			tok := lex.scanNumber()
			a.Equal(tc.tok, tok)

			// Test that we can parse the values.
			var (
				num any
				err error
			)

			switch tc.tok.tok {
			case integer:
				num, err = strconv.ParseInt(tok.val, 10, 64)
			case number:
				num, err = strconv.ParseFloat(tok.val, 64)
			default:
				return
			}

			if _, ok := tc.num.(bool); ok {
				// Not a valid value.
				r.Error(err)
			} else {
				r.NoError(err)
				a.Equal(tc.num, num)
			}
		})
	}
}

func TestScanBlankSpace(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		tok  token
		next rune
	}{
		{
			name: "empty",
			in:   "",
			tok:  token{blankSpace, "", 0},
			next: eof,
		},
		{
			name: "no_spaces",
			in:   "xxx",
			tok:  token{blankSpace, "", 0},
			next: 'x',
		},
		{
			name: "space",
			in:   " ",
			tok:  token{blankSpace, " ", 0},
			next: eof,
		},
		{
			name: "spaces",
			in:   "     ",
			tok:  token{blankSpace, "     ", 0},
			next: eof,
		},
		{
			name: "spacey",
			in:   "     y",
			tok:  token{blankSpace, "     ", 0},
			next: 'y',
		},
		{
			name: "newline",
			in:   "\n",
			tok:  token{blankSpace, "\n", 0},
			next: eof,
		},
		{
			name: "newlines",
			in:   "\n\n\n\n",
			tok:  token{blankSpace, "\n\n\n\n", 0},
			next: eof,
		},
		{
			name: "newline_plus",
			in:   "\n\n\n\ngo on",
			tok:  token{blankSpace, "\n\n\n\n", 0},
			next: 'g',
		},
		{
			name: "linefeed",
			in:   "\r",
			tok:  token{blankSpace, "\r", 0},
			next: eof,
		},
		{
			name: "multiple_linefeed",
			in:   "\r\r\r\r",
			tok:  token{blankSpace, "\r\r\r\r", 0},
			next: eof,
		},
		{
			name: "linefeed_plus",
			in:   "\r\r\r\rgo on",
			tok:  token{blankSpace, "\r\r\r\r", 0},
			next: 'g',
		},
		{
			name: "tab",
			in:   "\t",
			tok:  token{blankSpace, "\t", 0},
			next: eof,
		},
		{
			name: "multiple_tab",
			in:   "\t\t\t\t",
			tok:  token{blankSpace, "\t\t\t\t", 0},
			next: eof,
		},
		{
			name: "tab_plus",
			in:   "\t\t\t\tgo on",
			tok:  token{blankSpace, "\t\t\t\t", 0},
			next: 'g',
		},
		{
			name: "mix_blanks",
			in:   "\t    \r\n\t   \r\n\t lol",
			tok:  token{blankSpace, "\t    \r\n\t   \r\n\t ", 0},
			next: 'l',
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			lex := newLexer(tc.in)
			a.Equal(tc.tok, lex.scanBlankSpace())
			lex = newLexer(tc.in)
			lex.skipBlankSpace()
			a.Equal(tc.next, lex.r)
		})
	}
}

func TestScanTokens(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		in     string
		tokens []token
	}{
		{
			name:   "empty",
			in:     "",
			tokens: []token{},
		},
		{
			name:   "dollar",
			in:     "$",
			tokens: []token{{'$', "", 0}},
		},
		{
			name: "dollar_dot_string",
			in:   "$.foo",
			tokens: []token{
				{'$', "", 0},
				{'.', "", 1},
				{identifier, "foo", 2},
			},
		},
		{
			name: "bracket_space_int_bracket",
			in:   "[  42]",
			tokens: []token{
				{'[', "", 0},
				{blankSpace, "  ", 1},
				{integer, "42", 3},
				{']', "", 5},
			},
		},
		{
			name: "string_bracket_int_bracket",
			in:   "'hello'[42]",
			tokens: []token{
				{goString, "hello", 0},
				{'[', "", 7},
				{integer, "42", 8},
				{']', "", 10},
			},
		},
		{
			name: "number_space_unclosed_string",
			in:   `98.6 "foo`,
			tokens: []token{
				{number, "98.6", 0},
				{blankSpace, " ", 4},
				{invalid, "unterminated string literal", 9},
			},
		},
		{
			name: "number_space_string_invalid_escape",
			in:   `98.6 'foo\x'`,
			tokens: []token{
				{number, "98.6", 0},
				{blankSpace, " ", 4},
				{invalid, "invalid escape after backslash", 10},
			},
		},
		{
			name: "number_space_string_invalid_unicode",
			in:   `98.6 'foo\uf3xx'`,
			tokens: []token{
				{number, "98.6", 0},
				{blankSpace, " ", 4},
				{invalid, "invalid escape after backslash", 11},
			},
		},
		{
			name: "number_space_string_invalid_low_surrogate",
			in:   `98.6 "foo\uD834\uED1E"`,
			tokens: []token{
				{number, "98.6", 0},
				{blankSpace, " ", 4},
				{invalid, "invalid escape after backslash", 17},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			lex := newLexer(tc.in)
			a.Equal(token{}, lex.prev)
			tokens := make([]token, 0, len(tc.tokens))
			for t := lex.scan(); t.tok != eof; t = lex.scan() {
				tokens = append(tokens, t)
				a.Equal(t, lex.prev)
			}
			a.Equal(tc.tokens, tokens)
		})
	}
}

func TestToken(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		tok  token
		id   string
		str  string
		err  string
	}{
		{
			name: "invalid",
			id:   "invalid",
			tok:  token{invalid, "oops", 12},
			str:  `Token{invalid, "oops", 12}`,
			err:  "jsonpath: oops at 12",
		},
		{
			name: "eof",
			id:   "eof",
			tok:  token{eof, "", 12},
			str:  `Token{eof, "", 12}`,
		},
		{
			name: "identifier",
			id:   "identifier",
			tok:  token{identifier, "foo", 12},
			str:  `Token{identifier, "foo", 12}`,
		},
		{
			name: "integer",
			id:   "integer",
			tok:  token{integer, "42", 12},
			str:  `Token{integer, "42", 12}`,
		},
		{
			name: "number",
			id:   "number",
			tok:  token{number, "98.6", 12},
			str:  `Token{number, "98.6", 12}`,
		},
		{
			name: "string",
			id:   "string",
			tok:  token{goString, "👋🏻 there", 12},
			str:  `Token{string, "👋🏻 there", 12}`,
		},
		{
			name: "blankSpace",
			id:   "blank space",
			tok:  token{blankSpace, "  \t", 3},
			str:  `Token{blank space, "  \t", 3}`,
		},
		{
			name: "bracket",
			id:   "'['",
			tok:  token{'[', "[", 3},
			str:  `Token{'[', "[", 3}`,
		},
		{
			name: "dollar",
			id:   "'$'",
			tok:  token{'$', "$", 3},
			str:  `Token{'$', "$", 3}`,
		},
		{
			name: "dot",
			id:   "'.'",
			tok:  token{'.', ".", 3},
			str:  `Token{'.', ".", 3}`,
		},
		{
			name: "multibyte",
			id:   "'ü'",
			tok:  token{'ü', "ü", 3},
			str:  `Token{'ü', "ü", 3}`,
		},
		{
			name: "emoji",
			id:   "'🐶'",
			tok:  token{'🐶', "🐶", 3},
			str:  `Token{'🐶', "🐶", 3}`,
		},
		{
			name: "surrogate_pair",
			id:   "'\U0001D11E'",
			tok:  token{'𝄞', "𝄞", 3},
			str:  `Token{'𝄞', "𝄞", 3}`,
		},
		{
			name: "newline",
			id:   `'\n'`,
			tok:  token{'\n', "\n", 3},
			str:  `Token{'\n', "\n", 3}`,
		},
		{
			name: "tab",
			id:   `'\t'`,
			tok:  token{'\t', "\t", 3},
			str:  `Token{'\t', "\t", 3}`,
		},
		{
			name: "null_byte",
			id:   `'\x00'`,
			tok:  token{'\u0000', "\x00", 3},
			str:  `Token{'\x00', "\x00", 3}`,
		},
		{
			name: "cancel",
			id:   `'\x18'`,
			tok:  token{'\u0018', "\x18", 3},
			str:  `Token{'\x18', "\x18", 3}`,
		},
		{
			name: "bel",
			id:   `'\a'`,
			tok:  token{'\007', "\007", 3},
			str:  `Token{'\a', "\a", 3}`,
		},
		{
			name: "unicode_space",
			id:   `'\u2028'`,
			tok:  token{'\u2028', "\u2028", 3},
			str:  `Token{'\u2028', "\u2028", 3}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			r := require.New(t)

			a.Equal(tc.id, tc.tok.name())
			a.Equal(tc.str, tc.tok.String())
			err := tc.tok.err()
			if tc.err == "" {
				r.NoError(err)
			} else {
				r.EqualError(err, tc.err)
				r.ErrorIs(err, ErrPathParse)
			}
		})
	}
}

func TestPeek(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	input := "this is it"
	lex := newLexer(input)
	for _, r := range input[1:] {
		a.Equal(r, lex.peek())
		lex.next()
	}

	for range 3 {
		a.Equal(rune(eof), lex.peek())
	}
}
