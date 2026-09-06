package xobis_test

import (
	"errors"
	"regexp"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

// The oracle uses independent regexes for record structure, decimal syntax and
// Unicode categories, plus the existing independent OBIS identifier oracle.
var (
	referenceReading = regexp.MustCompile(`\A([^()]+)\(([+-]?[0-9]+(?:\.[0-9]+)?)(?:\*([^()*]+))?\)\z`)
	referenceUnit    = regexp.MustCompile(`\A[\pL\pM\pN\pP\pS]+\z`)
)

func FuzzParseReading(f *testing.F) {
	// Arrange the seed corpus.
	for _, input := range []string{
		"", "1.8(0)", "1.8.0(123*kWh)", "0:1.8(+00123.4500*kWh)\n",
		"0-1.8(-0.000*W)\r\n", "1-0:1.8.0*255(9007199254740993.0000000000000000001*kWh)",
		"01.000.001.08.000.0255(1*Wh)", "1.8(1*m³/h)", "1.8(1*°C)",
		"1.8(1*e\u0301)", "1.8(1*\ufffd)", "1.8(1*k\u00a0W)", "1.8(1*k\u200bW)",
		"1.8(1*k\xffW)", "1.8(NaN)", "1.8(1e3)", "1.8(1,5)", "1.8(+)",
		"1.8(*kWh)", "1.8(1*)", "1.8(1*kWh*x)", "1.8((1))", "1.8(1)(2)",
		"1.8(1)\r", "1.8(1)\n\n", "1.8(1\n", "1.256(1)\r\n", "16-1.8(1)",
	} {
		f.Add(input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		// Arrange.
		match := referenceReading.FindStringSubmatch(input)
		var groups xobis.Groups
		var present xobis.Presence
		valid := false
		if match != nil {
			groups, present, valid = referenceParse(match[1])
			valid = valid && utf8.ValidString(match[3]) && (match[3] == "" || referenceUnit.MatchString(match[3]))
		}
		var expected xobis.Reading
		if valid {
			identifier, err := xobis.NewPattern(groups, present)
			require.NoError(t, err)
			decimal, err := xobis.ParseDecimal(match[2])
			require.NoError(t, err)
			expected, err = xobis.NewReading(identifier, decimal, xobis.Unit(match[3]))
			require.NoError(t, err)
		}
		var roundtrip xobis.Reading
		var roundtripErr, marshalErr, validationErr error
		var text []byte

		// Act.
		got, err := xobis.ParseReading(input)
		if valid {
			validationErr = got.Validate()
			text, marshalErr = got.MarshalText()
			roundtrip, roundtripErr = xobis.ParseReading(string(text))
		}

		// Assert.
		if !valid {
			var pe *xobis.ParseError
			require.ErrorAs(t, err, &pe, "input: %q", input)
			assert.Equal(t, input, pe.Input)
			assert.GreaterOrEqual(t, pe.Offset, 0)
			assert.LessOrEqual(t, pe.Offset, len(input))
			assert.True(t, errors.Is(err, xobis.ErrSyntax) || errors.Is(err, xobis.ErrRange))
			assert.Zero(t, got)
			return
		}
		require.NoError(t, err, "input: %q", input)
		assert.Equal(t, expected, got)
		gotGroups, gotPresence := got.Identifier().Groups()
		assert.Equal(t, groups, gotGroups)
		assert.Equal(t, present, gotPresence)
		assert.Equal(t, match[2], got.Value().String())
		assert.Equal(t, xobis.Unit(match[3]), got.Unit())
		assert.NoError(t, validationErr)
		require.NoError(t, marshalErr)
		assert.Equal(t, got.String(), string(text))
		require.NoError(t, roundtripErr)
		assert.Equal(t, got, roundtrip)
	})
}
