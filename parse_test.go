package jsonpath

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRoot(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)

	p, err := Parse("$")
	r.NoError(err)
	a.Equal("$", p.String())
	a.Empty(p.q.segments)
}

func TestParseSimple(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)

	for _, tc := range []struct {
		name string
		path string
		exp  *Query
		err  string
	}{
		{
			name: "root",
			path: "$",
			exp:  NewQuery([]*Segment{}),
		},
		{
			name: "name",
			path: "$.x",
			exp:  NewQuery([]*Segment{Child(Name("x"))}),
		},
		{
			name: "trim_leading_space",
			path: "   $.x",
			err:  `jsonpath: unexpected blank space at position 1`,
		},
		{
			name: "trim_trailing_space",
			path: "$.x    ",
			err:  `jsonpath: unexpected blank space at position 4`,
		},
		{
			name: "no_interim_space",
			path: "$.x   .y",
			exp:  NewQuery([]*Segment{Child(Name("x")), Child(Name("y"))}),
		},
		{
			name: "unexpected_integer",
			path: "$.62",
			err:  `jsonpath: unexpected integer at position 3`,
		},
		{
			name: "unexpected_token",
			path: "$.==12",
			err:  `jsonpath: unexpected '=' at position 3`,
		},
		{
			name: "name_name",
			path: "$.x.y",
			exp:  NewQuery([]*Segment{Child(Name("x")), Child(Name("y"))}),
		},
		{
			name: "wildcard",
			path: "$.*",
			exp:  NewQuery([]*Segment{Child(Wildcard)}),
		},
		{
			name: "wildcard_wildcard",
			path: "$.*.*",
			exp:  NewQuery([]*Segment{Child(Wildcard), Child(Wildcard)}),
		},
		{
			name: "name_wildcard",
			path: "$.x.*",
			exp:  NewQuery([]*Segment{Child(Name("x")), Child(Wildcard)}),
		},
		{
			name: "desc_name",
			path: "$..x",
			exp:  NewQuery([]*Segment{Descendant(Name("x"))}),
		},
		{
			name: "desc_name_2x",
			path: "$..x..y",
			exp:  NewQuery([]*Segment{Descendant(Name("x")), Descendant(Name("y"))}),
		},
		{
			name: "desc_wildcard",
			path: "$..*",
			exp:  NewQuery([]*Segment{Descendant(Wildcard)}),
		},
		{
			name: "desc_wildcard_2x",
			path: "$..*..*",
			exp:  NewQuery([]*Segment{Descendant(Wildcard), Descendant(Wildcard)}),
		},
		{
			name: "desc_wildcard_name",
			path: "$..*.xyz",
			exp:  NewQuery([]*Segment{Descendant(Wildcard), Child(Name("xyz"))}),
		},
		{
			name: "wildcard_desc_name",
			path: "$.*..xyz",
			exp:  NewQuery([]*Segment{Child(Wildcard), Descendant(Name("xyz"))}),
		},
		{
			name: "empty_string",
			path: "",
			err:  "jsonpath: unexpected end of input",
		},
		{
			name: "bad_start",
			path: ".x",
			err:  `jsonpath: unexpected '.' at position 1`,
		},
		{
			name: "not_a_segment",
			path: "$foo",
			err:  "jsonpath: unexpected identifier at position 1",
		},
		{
			name: "not_a_dot_segment",
			path: "$.{x}",
			err:  `jsonpath: unexpected '{' at position 3`,
		},
		{
			name: "not_a_descendant",
			path: "$..{x}",
			err:  `jsonpath: unexpected '{' at position 4`,
		},
		{
			name: "name_with_dollar",
			path: "$.x$.$y",
			exp:  NewQuery([]*Segment{Child(Name("x$")), Child(Name("$y"))}),
		},
		{
			name: "name_with_escape",
			path: `$.x\r.y\n`,
			exp:  NewQuery([]*Segment{Child(Name("x\r")), Child(Name("y\n"))}),
		},
		{
			name: "name_with_unicode_escape",
			path: `$.fo\u00f8.tune\uD834\uDD1E`,
			exp:  NewQuery([]*Segment{Child(Name("foø")), Child(Name("tune𝄞"))}),
		},
		{
			name: "name_with_leading_unicode_escape",
			path: `$.\u00f8ps`,
			exp:  NewQuery([]*Segment{Child(Name("øps"))}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p, err := Parse(tc.path)
			if tc.err == "" {
				r.NoError(err)
				a.Equal(New(tc.exp), p)
			} else {
				a.Nil(p)
				r.EqualError(err, tc.err)
				r.ErrorIs(err, ErrPathParse)
			}
		})
	}
}

func TestParseSelectors(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)

	for _, tc := range []struct {
		name string
		path string
		exp  *Query
		err  string
	}{
		{
			name: "index",
			path: "$[0]",
			exp:  NewQuery([]*Segment{Child(Index(0))}),
		},
		{
			name: "two_indexes",
			path: "$[0, 1]",
			exp:  NewQuery([]*Segment{Child(Index(0), Index(1))}),
		},
		{
			name: "name",
			path: `$["foo"]`,
			exp:  NewQuery([]*Segment{Child(Name("foo"))}),
		},
		{
			name: "sq_name",
			path: `$['foo']`,
			exp:  NewQuery([]*Segment{Child(Name("foo"))}),
		},
		{
			name: "two_names",
			path: `$["foo", "🐦‍🔥"]`,
			exp:  NewQuery([]*Segment{Child(Name("foo"), Name("🐦‍🔥"))}),
		},
		{
			name: "json_escapes",
			path: `$["abx_xyx\ryup", "\b\f\n\r\t\/\\"]`,
			exp:  NewQuery([]*Segment{Child(Name("abx_xyx\ryup"), Name("\b\f\n\r\t/\\"))}),
		},
		{
			name: "unicode_escapes",
			path: `$["fo\u00f8", "tune \uD834\uDD1E"]`,
			exp:  NewQuery([]*Segment{Child(Name("foø"), Name("tune 𝄞"))}),
		},
		{
			name: "slice_start",
			path: `$[1:]`,
			exp:  NewQuery([]*Segment{Child(Slice(1))}),
		},
		{
			name: "slice_start_2",
			path: `$[2:]`,
			exp:  NewQuery([]*Segment{Child(Slice(2))}),
		},
		{
			name: "slice_start_end",
			path: `$[2:6]`,
			exp:  NewQuery([]*Segment{Child(Slice(2, 6))}),
		},
		{
			name: "slice_end",
			path: `$[:6]`,
			exp:  NewQuery([]*Segment{Child(Slice(nil, 6))}),
		},
		{
			name: "slice_start_end_step",
			path: `$[2:6:2]`,
			exp:  NewQuery([]*Segment{Child(Slice(2, 6, 2))}),
		},
		{
			name: "slice_start_step",
			path: `$[2::2]`,
			exp:  NewQuery([]*Segment{Child(Slice(2, nil, 2))}),
		},
		{
			name: "slice_step",
			path: `$[::2]`,
			exp:  NewQuery([]*Segment{Child(Slice(nil, nil, 2))}),
		},
		{
			name: "slice_defaults",
			path: `$[:]`,
			exp:  NewQuery([]*Segment{Child(Slice())}),
		},
		{
			name: "slice_spacing",
			path: `$[   1:  2  : 2   ]`,
			exp:  NewQuery([]*Segment{Child(Slice(1, 2, 2))}),
		},
		{
			name: "slice_slice",
			path: `$[:,:]`,
			exp:  NewQuery([]*Segment{Child(Slice(), Slice())}),
		},
		{
			name: "slice_slice_slice",
			path: `$[2:,:4,7:9]`,
			exp:  NewQuery([]*Segment{Child(Slice(2), Slice(nil, 4), Slice(7, 9))}),
		},
		{
			name: "slice_name",
			path: `$[:,"hi"]`,
			exp:  NewQuery([]*Segment{Child(Slice(), Name("hi"))}),
		},
		{
			name: "name_slice",
			path: `$["hi",2:]`,
			exp:  NewQuery([]*Segment{Child(Name("hi"), Slice(2))}),
		},
		{
			name: "slice_index",
			path: `$[:,42]`,
			exp:  NewQuery([]*Segment{Child(Slice(), Index(42))}),
		},
		{
			name: "index_slice",
			path: `$[42,:3]`,
			exp:  NewQuery([]*Segment{Child(Index(42), Slice(nil, 3))}),
		},
		{
			name: "slice_wildcard",
			path: `$[:,   *]`,
			exp:  NewQuery([]*Segment{Child(Slice(), Wildcard)}),
		},
		{
			name: "wildcard_slice",
			path: `$[  *,  :   ]`,
			exp:  NewQuery([]*Segment{Child(Wildcard, Slice())}),
		},
		{
			name: "slice_neg_start",
			path: `$[-3:]`,
			exp:  NewQuery([]*Segment{Child(Slice(-3))}),
		},
		{
			name: "slice_neg_end",
			path: `$[:-3:]`,
			exp:  NewQuery([]*Segment{Child(Slice(nil, -3))}),
		},
		{
			name: "slice_neg_step",
			path: `$[::-2]`,
			exp:  NewQuery([]*Segment{Child(Slice(nil, nil, -2))}),
		},
		{
			name: "index_name_slice, wildcard",
			path: `$[3, "🦀", :3,*]`,
			exp:  NewQuery([]*Segment{Child(Index(3), Name("🦀"), Slice(nil, 3), Wildcard)}),
		},
		{
			name: "filter_unsupported",
			path: `$[?@.x == 'y']`,
			err:  `jsonpath: filter selectors not yet supported at position 3`,
		},
		{
			name: "slice_bad_start",
			path: `$[:d]`,
			err:  `jsonpath: unexpected identifier at position 4`,
		},
		{
			name: "slice_four_parts",
			path: `$[0:0:0:0]`,
			err:  `jsonpath: unexpected integer at position 9`,
		},
		{
			name: "invalid_selector",
			path: `$[{}]`,
			err:  `jsonpath: unexpected '{' at position 3`,
		},
		{
			name: "invalid_second_selector",
			path: `$[1, hi]`,
			err:  `jsonpath: unexpected identifier at position 6`,
		},
		{
			name: "missing_segment_comma",
			path: `$[1 "hi"]`,
			err:  `jsonpath: unexpected string at position 5`,
		},
		{
			name: "space_index",
			path: "$[   0]",
			exp:  NewQuery([]*Segment{Child(Index(0))}),
		},
		{
			name: "index_space_comma_index",
			path: "$[0    , 12]",
			exp:  NewQuery([]*Segment{Child(Index(0), Index(12))}),
		},
		{
			name: "index_comma_space_name",
			path: `$[0, "xyz"]`,
			exp:  NewQuery([]*Segment{Child(Index(0), Name("xyz"))}),
		},
		{
			name: "tab_index",
			path: "$[\t0]",
			exp:  NewQuery([]*Segment{Child(Index(0))}),
		},
		{
			name: "newline_index",
			path: "$[\n0]",
			exp:  NewQuery([]*Segment{Child(Index(0))}),
		},
		{
			name: "return_index",
			path: "$[\r0]",
			exp:  NewQuery([]*Segment{Child(Index(0))}),
		},
		{
			name: "name_space",
			path: `$["hi"   ]`,
			exp:  NewQuery([]*Segment{Child(Name("hi"))}),
		},
		{
			name: "wildcard_tab",
			path: "$[*\t]",
			exp:  NewQuery([]*Segment{Child(Wildcard)}),
		},
		{
			name: "slice_newline",
			path: "$[2:\t]",
			exp:  NewQuery([]*Segment{Child(Slice(2))}),
		},
		{
			name: "index_return",
			path: "$[0\r]",
			exp:  NewQuery([]*Segment{Child(Index(0))}),
		},
		{
			name: "descendant_index",
			path: "$..[0]",
			exp:  NewQuery([]*Segment{Descendant(Index(0))}),
		},
		{
			name: "descendant_name",
			path: `$..["hi"]`,
			exp:  NewQuery([]*Segment{Descendant(Name("hi"))}),
		},
		{
			name: "descendant_multi",
			path: `$..[  "hi", 2, *, 4:5  ]`,
			exp:  NewQuery([]*Segment{Descendant(Name("hi"), Index(2), Wildcard, Slice(4, 5))}),
		},
		{
			name: "invalid_descendant",
			path: "$..[oops]",
			err:  `jsonpath: unexpected identifier at position 5`,
		},
		{
			name: "invalid_unicode_escape",
			path: `$["fo\uu0f8"]`,
			err:  `jsonpath: invalid escape after backslash at position 8`,
		},
		{
			name: "invalid_integer",
			path: `$[170141183460469231731687303715884105727]`, // too large
			err:  `jsonpath: cannot parse "170141183460469231731687303715884105727", value out of range at position 3`,
		},
		{
			name: "invalid_slice_float",
			path: `$[:170141183460469231731687303715884105727]`, // too large
			err:  `jsonpath: cannot parse "170141183460469231731687303715884105727", value out of range at position 4`,
		},
		{
			name: `name_sq_name_desc_wild`,
			path: `$.names['first_name']..*`,
			exp: NewQuery([]*Segment{
				Child(Name("names")),
				Child(Name("first_name")),
				Descendant(Wildcard),
			}),
		},
		{
			name: "no_tail",
			path: `$.a['b']tail`,
			err:  `jsonpath: unexpected identifier at position 9`,
		},
		{
			name: "dq_name",
			path: `$["name"]`,
			exp:  NewQuery([]*Segment{Child(Name("name"))}),
		},
		{
			name: "sq_name",
			path: `$['name']`,
			exp:  NewQuery([]*Segment{Child(Name("name"))}),
		},
		{
			name: "two_name_segment",
			path: `$["name","test"]`,
			exp:  NewQuery([]*Segment{Child(Name("name"), Name("test"))}),
		},
		{
			name: "name_index_slice_segment",
			path: `$['name',10,0:3]`,
			exp:  NewQuery([]*Segment{Child(Name("name"), Index(10), Slice(0, 3))}),
		},
		{
			name: "default_slice_wildcard_segment",
			path: `$[::,*]`,
			exp:  NewQuery([]*Segment{Child(Slice(), Wildcard)}),
		},
		{
			name: "leading_zero_index",
			path: `$[010]`,
			err:  `jsonpath: invalid number literal at position 3`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p, err := Parse(tc.path)
			if tc.err == "" {
				r.NoError(err)
				a.Equal(New(tc.exp), p)
			} else {
				a.Nil(p)
				r.EqualError(err, tc.err)
				r.ErrorIs(err, ErrPathParse)
			}
		})
	}
}

func TestMakeNumErr(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	t.Run("parse_int", func(t *testing.T) {
		t.Parallel()
		_, numErr := strconv.ParseInt("170141183460469231731687303715884105727", 10, 64)
		r.Error(numErr)
		tok := token{invalid, "", 6}
		err := makeNumErr(tok, numErr)
		r.EqualError(
			err,
			`jsonpath: cannot parse "170141183460469231731687303715884105727", value out of range at position 7`,
		)
		r.ErrorIs(err, ErrPathParse)
	})

	t.Run("parse_float", func(t *testing.T) {
		t.Parallel()
		_, numErr := strconv.ParseFloat("99e+1234", 64)
		r.Error(numErr)
		tok := token{invalid, "", 12}
		err := makeNumErr(tok, numErr)
		r.EqualError(
			err,
			`jsonpath: cannot parse "99e+1234", value out of range at position 13`,
		)
		r.ErrorIs(err, ErrPathParse)
	})

	t.Run("other error", func(t *testing.T) {
		t.Parallel()
		myErr := errors.New("oops")
		tok := token{invalid, "", 19}
		err := makeNumErr(tok, myErr)
		r.EqualError(err, `jsonpath: oops at position 20`)
		r.ErrorIs(err, ErrPathParse)
	})
}
