package xobis_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

func FuzzParsePatternFrom(f *testing.F) {
	// Arrange the seed corpus. Prefixes set arbitrary byte offsets, and suffixes
	// can either terminate an identifier or extend it into an invalid token.
	// Prefix lengths also vary the reader chunk size.
	for _, seed := range [][3]string{
		{"", "1.8", ""}, {"µ=", "1.8", "xyz.0.0.0.0"},
		{"header.", "1.8.0", "(1.2.3.4)"}, {"prefix", "1.8", "."},
		{"\xff", "01.000.001.08.000.0255", "\r\n"},
		{"", "0-0:1.8.0*255", ";"}, {"", "0-1.8", ":1"},
		{"", "1.8.0.255", ""}, {"", "1.8.0.1.255", ""},
		{"", "1.0.1.8.0.255", ".1"}, {"", "16-1.8", ""},
		{"prefix", "1.8*256", "\xff"}, {"prefix", "", ""},
	} {
		f.Add(seed[0], seed[1], seed[2])
	}
	f.Fuzz(func(t *testing.T, prefix, candidate, suffix string) {
		// Arrange. An independent character-set scan identifies the complete
		// OBIS token; the regex/strconv oracle determines whether it is valid.
		remainder := candidate + suffix
		consumed := len(remainder) - len(strings.TrimLeft(remainder, "0123456789-:.*"))
		groups, present, valid := referenceParse(remainder[:consumed])
		var want xobis.Pattern
		if valid {
			var err error
			want, err = xobis.NewPattern(groups, present)
			require.NoError(t, err)
		}
		input := prefix + remainder
		cursor := offsetCursor{Cursor: xobis.NewReaderCursor(&chunkReader{input: remainder, chunk: 1 + len(prefix)%16}), base: int64(len(prefix))}

		// Act.
		got, err := xobis.ParsePatternFrom(&cursor)
		whole, wholeErr := xobis.ParsePattern(remainder)

		// Assert.
		if !valid {
			assertCursorFailure(t, &cursor, int64(len(prefix)), int64(len(input)), err)
			assert.Zero(t, got)
			require.Error(t, wholeErr)
			assert.Zero(t, whole)
			return
		}
		require.NoError(t, err)
		assert.Equal(t, want, got)
		assert.Positive(t, consumed)
		assert.Equal(t, int64(len(prefix)+consumed), cursor.Offset())
		if consumed == len(remainder) {
			require.NoError(t, wholeErr)
			assert.Equal(t, want, whole)
		} else {
			assert.ErrorIs(t, wholeErr, xobis.ErrSyntax)
			assert.Zero(t, whole)
		}
	})
}

func FuzzParseReadingFrom(f *testing.F) {
	// Arrange the seed corpus.
	for _, seed := range [][3]string{
		{"", "1.8(0)", ""}, {"µ=", "1.8(+00123.4500*kWh)", "\r\n"},
		{"prefix", "1.8(1)", "\n\n"}, {"prefix", "1.8(1)", "\r"},
		{"header.", "1.8(1)", "1.8(2)"}, {"\xff", "1.8(1)", ".1.2.3.4"},
		{"", "1-0:1.8.0*255(1*°C)", "(2)"},
		{"", "01.000.001.08.000.0255(-0.00*m³/h)", ";"},
		{"prefix", "1.8(1*e\u0301)", "\xff"},
		{"", "1.8(1*°C\xff)", "\r\n"}, {"", "1.8(1*k\u00a0W)", ""},
		{"", "1.8(1*k\u200bW)", ""}, {"", "1.8(1e3)", ""},
		{"", "1.8((1))", ""}, {"", "1.8(1*)", ""},
		{"prefix", "1.8(1", ""}, {"prefix", "", ""},
	} {
		f.Add(seed[0], seed[1], seed[2])
	}
	f.Fuzz(func(t *testing.T, prefix, candidate, suffix string) {
		// Arrange. The independent reading regex must accept the text through
		// the first ')'; subsequent framing is exclusively the caller's input.
		remainder := candidate + suffix
		consumed := strings.IndexByte(remainder, ')') + 1
		match := referenceReading.FindStringSubmatch(remainder[:consumed])
		var groups xobis.Groups
		var present xobis.Presence
		valid := false
		if match != nil {
			groups, present, valid = referenceParse(match[1])
			valid = valid && utf8.ValidString(match[3]) && (match[3] == "" || referenceUnit.MatchString(match[3]))
		}
		var want xobis.Reading
		if valid {
			identifier, err := xobis.NewPattern(groups, present)
			require.NoError(t, err)
			value, err := xobis.ParseDecimal(match[2])
			require.NoError(t, err)
			want, err = xobis.NewReading(identifier, value, xobis.Unit(match[3]))
			require.NoError(t, err)
		}
		input := prefix + remainder
		cursor := offsetCursor{Cursor: xobis.NewReaderCursor(&chunkReader{input: remainder, chunk: 1 + len(prefix)%16}), base: int64(len(prefix))}
		tail := remainder[consumed:]

		// Act.
		got, err := xobis.ParseReadingFrom(&cursor)
		whole, wholeErr := xobis.ParseReading(remainder)

		// Assert.
		if !valid {
			assertCursorFailure(t, &cursor, int64(len(prefix)), int64(len(input)), err)
			assert.Zero(t, got)
			require.Error(t, wholeErr)
			assert.Zero(t, whole)
			return
		}
		require.NoError(t, err)
		assert.Equal(t, want, got)
		assert.Positive(t, consumed)
		assert.Equal(t, int64(len(prefix)+consumed), cursor.Offset())
		if tail == "" {
			require.NoError(t, wholeErr)
			assert.Equal(t, want, whole)
		} else {
			assert.ErrorIs(t, wholeErr, xobis.ErrSyntax)
			assert.Zero(t, whole)
		}
	})
}

func assertCursorFailure(t *testing.T, cursor xobis.Cursor, start, end int64, err error) {
	t.Helper()
	var pe *xobis.CursorError
	require.ErrorAs(t, err, &pe)
	assert.GreaterOrEqual(t, pe.Offset, start)
	assert.LessOrEqual(t, pe.Offset, end)
	assert.GreaterOrEqual(t, cursor.Offset(), start)
	assert.LessOrEqual(t, cursor.Offset(), end)
}
