package xobis_test

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

// The test oracle uses regular expressions and strconv independently of the
// production scanner. In particular, it also catches valid inputs rejected by
// the scanner, which a round-trip property alone would miss.
var (
	referenceReduced = regexp.MustCompile(`\A(?:([0-9]+)-)?(?:([0-9]+):)?([0-9]+)\.([0-9]+)(?:\.([0-9]+))?(?:\*([0-9]+))?\z`)
	referenceDotted  = regexp.MustCompile(`\A([0-9]+)\.([0-9]+)\.([0-9]+)\.([0-9]+)\.([0-9]+)\.([0-9]+)\z`)
)

func referenceParse(s string) (xobis.Groups, xobis.Presence, bool) {
	m := referenceDotted.FindStringSubmatch(s)
	if m == nil {
		m = referenceReduced.FindStringSubmatch(s)
	}
	if m == nil {
		return xobis.Groups{}, 0, false
	}
	var groups [6]uint8
	var present xobis.Presence
	for i, text := range m[1:] {
		if text == "" {
			continue
		}
		n, err := strconv.ParseUint(text, 10, 8)
		if err != nil || (i == 0 && n > 15) {
			return xobis.Groups{}, 0, false
		}
		groups[i] = uint8(n)
		present |= 1 << i
	}
	return xobis.Groups{
		A: xobis.Medium(groups[0]), B: xobis.Channel(groups[1]),
		C: xobis.Quantity(groups[2]), D: xobis.Processing(groups[3]),
		E: xobis.Classification(groups[4]), F: xobis.Storage(groups[5]),
	}, present, true
}

func FuzzParse(f *testing.F) {
	// Arrange the seed corpus.
	for _, input := range []string{
		"", "1.8", "1.8.0", "0:1.8", "1-1.8", "1-0:1.8",
		"1.8*255", "1.8.0*0", "0:1.8.0*255", "1-1.8.0*255",
		"1-0:1.8.0*255", "1.0.1.8.0.255", "15-255:255.255.255*255",
		"01-000:001.08.000*0255", "0-0:0.0.0*0", "16-0:1.8.0*255",
		"1-0:1.8.0*256", "1.8.0.255", "1.0.1.8.0.255\n", "１.8.0",
		"C.1.0", "1-0:1.8.0*255\x00", "\xff",
	} {
		f.Add(input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		// Arrange.
		want, mask, valid := referenceParse(input)
		complete := valid && mask == xobis.PresentA|xobis.PresentB|xobis.PresentC|xobis.PresentD|xobis.PresentE|xobis.PresentF
		var patternRoundtrip xobis.Pattern
		var codeRoundtrip xobis.Code
		var validationErr, patternRoundtripErr, codeRoundtripErr error
		var matched bool
		var candidate xobis.Code
		if valid {
			var err error
			candidate, err = xobis.NewCode(want)
			require.NoError(t, err)
		}

		// Act.
		p, patternErr := xobis.ParsePattern(input)
		groups, present := p.Groups()
		c, codeErr := xobis.Parse(input)
		if valid {
			validationErr = p.Validate()
			patternRoundtrip, patternRoundtripErr = xobis.ParsePattern(p.String())
			matched = p.Match(candidate)
		}
		if complete {
			codeRoundtrip, codeRoundtripErr = xobis.Parse(c.String())
		}

		// Assert.
		if valid {
			require.NoError(t, patternErr, "input: %q", input)
			assert.Equal(t, want, groups)
			assert.Equal(t, mask, present)
			assert.Equal(t, mask, p.Presence())
			assert.NoError(t, validationErr)
			require.NoError(t, patternRoundtripErr)
			assert.Equal(t, p, patternRoundtrip)
			assert.True(t, matched)
		} else {
			require.Error(t, patternErr, "input: %q", input)
			assert.Zero(t, p)
		}
		if complete {
			require.NoError(t, codeErr, "input: %q", input)
			assert.Equal(t, want, c.Groups())
			require.NoError(t, codeRoundtripErr)
			assert.Equal(t, c, codeRoundtrip)
		} else {
			require.Error(t, codeErr, "input: %q", input)
			assert.Zero(t, c)
		}
	})
}

func BenchmarkParse(b *testing.B) {
	// Arrange.
	for _, input := range []string{"1-0:1.8.0*255", "1.0.1.8.0.255"} {
		b.Run(input, func(b *testing.B) {
			// Arrange.
			var err error
			b.ReportAllocs()

			// Act.
			for b.Loop() {
				_, err = xobis.Parse(input)
				if err != nil {
					break
				}
			}

			// Assert outside the measured loop.
			require.NoError(b, err)
		})
	}
}

func BenchmarkParsePattern(b *testing.B) {
	// Arrange.
	input := "1.8.0"
	var err error
	b.ReportAllocs()

	// Act.
	for b.Loop() {
		_, err = xobis.ParsePattern(input)
		if err != nil {
			break
		}
	}

	// Assert outside the measured loop.
	require.NoError(b, err)
}
