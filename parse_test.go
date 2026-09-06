package xobis_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

func TestParse(t *testing.T) {
	// Arrange.
	tests := []struct {
		input string
		want  xobis.Groups
	}{
		{"1-0:1.8.0*255", xobis.Groups{A: xobis.MediumElectricity, C: xobis.ElectricityActivePowerImport, D: xobis.ElectricityTimeIntegral1, F: xobis.StorageNotUsed}},
		{"1.0.1.8.0.255", xobis.Groups{A: 1, C: 1, D: 8, F: 255}},
		{"01-000:001.08.000*0255", xobis.Groups{A: 1, C: 1, D: 8, F: 255}},
		{"01.000.001.08.000.0255", xobis.Groups{A: 1, C: 1, D: 8, F: 255}},
		{"0-0:0.0.0*0", xobis.Groups{}},
		{"15-255:255.255.255*255", xobis.Groups{A: 15, B: 255, C: 255, D: 255, E: 255, F: 255}},
		{"7-1:23.2.0*255", xobis.Groups{A: xobis.MediumGas, B: 1, C: 23, D: 2, F: 255}},
		// Numeric validation preserves reserved and manufacturer-defined values.
		{"2-200:240.128.254*128", xobis.Groups{A: 2, B: 200, C: 240, D: 128, E: 254, F: 128}},
		{strings.Repeat("0", 4096) + "1-0:1.8.0*255", xobis.Groups{A: 1, C: 1, D: 8, F: 255}},
	}
	for _, tt := range tests {
		t.Run(tt.input[:min(len(tt.input), 40)], func(t *testing.T) {
			// Act.
			got, err := xobis.Parse(tt.input)
			validationErr := got.Validate()
			roundtrip, roundtripErr := xobis.Parse(got.String())

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Groups())
			assert.NoError(t, validationErr)
			require.NoError(t, roundtripErr)
			assert.Equal(t, got, roundtrip)
		})
	}
}

func TestParsePatternOptionalGroups(t *testing.T) {
	// Arrange.
	tests := []struct {
		input string
		mask  xobis.Presence
	}{
		{"1.8", 0},
		{"2-1.8", xobis.PresentA},
		{"3:1.8", xobis.PresentB},
		{"2-3:1.8", xobis.PresentA | xobis.PresentB},
		{"1.8.4", xobis.PresentE},
		{"2-1.8.4", xobis.PresentA | xobis.PresentE},
		{"3:1.8.4", xobis.PresentB | xobis.PresentE},
		{"2-3:1.8.4", xobis.PresentA | xobis.PresentB | xobis.PresentE},
		{"1.8*5", xobis.PresentF},
		{"2-1.8*5", xobis.PresentA | xobis.PresentF},
		{"3:1.8*5", xobis.PresentB | xobis.PresentF},
		{"2-3:1.8*5", xobis.PresentA | xobis.PresentB | xobis.PresentF},
		{"1.8.4*5", xobis.PresentE | xobis.PresentF},
		{"2-1.8.4*5", xobis.PresentA | xobis.PresentE | xobis.PresentF},
		{"3:1.8.4*5", xobis.PresentB | xobis.PresentE | xobis.PresentF},
		{"2-3:1.8.4*5", xobis.PresentA | xobis.PresentB | xobis.PresentE | xobis.PresentF},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Arrange.
			mask := tt.mask | xobis.PresentC | xobis.PresentD
			want := xobis.Groups{C: 1, D: 8}
			if tt.mask.PresentA() {
				want.A = 2
			}
			if tt.mask.PresentB() {
				want.B = 3
			}
			if tt.mask.PresentE() {
				want.E = 4
			}
			if tt.mask.PresentF() {
				want.F = 5
			}

			// Act.
			p, err := xobis.ParsePattern(tt.input)
			groups, present := p.Groups()
			validationErr := p.Validate()
			formatted := p.String()
			_, completeErr := xobis.Parse(tt.input)

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, mask, p.Presence())
			assert.Equal(t, want, groups)
			assert.Equal(t, mask, present)
			assert.Equal(t, tt.input, formatted)
			assert.NoError(t, validationErr)
			if tt.input == "2-3:1.8.4*5" {
				assert.NoError(t, completeErr)
			} else {
				assert.ErrorIs(t, completeErr, xobis.ErrSyntax)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	// Arrange.
	parsers := []struct {
		name  string
		parse func(string) (any, error)
	}{
		{"code", func(input string) (any, error) { return xobis.Parse(input) }},
		{"pattern", func(input string) (any, error) { return xobis.ParsePattern(input) }},
	}
	cases := identifierParseErrorCases()
	for _, parser := range parsers {
		t.Run(parser.name, func(t *testing.T) {
			for _, tt := range cases {
				t.Run(fmt.Sprintf("%q", tt.input[:min(len(tt.input), 40)]), func(t *testing.T) {
					// Arrange.
					var pe *xobis.ParseError

					// Act.
					got, err := parser.parse(tt.input)

					// Assert.
					assertParseError(t, tt.input, err, tt.kind)
					require.ErrorAs(t, err, &pe)
					assert.Equal(t, tt.offset, pe.Offset)
					assert.Zero(t, got)
				})
			}
		})
	}
}

// Both string parsers reject these inputs under their shared identifier grammar.
func identifierParseErrorCases() []struct {
	input  string
	kind   error
	offset int
} {
	return []struct {
		input  string
		kind   error
		offset int
	}{
		{"", xobis.ErrSyntax, 0},
		{"1", xobis.ErrSyntax, 1},
		{"1-", xobis.ErrSyntax, 2},
		{"1-0:", xobis.ErrSyntax, 4},
		{"1-0:1", xobis.ErrSyntax, 5},
		{"1.", xobis.ErrSyntax, 2},
		{"1..8", xobis.ErrSyntax, 2},
		{"1.8.", xobis.ErrSyntax, 4},
		{"1.8*", xobis.ErrSyntax, 4},
		{"-1.8", xobis.ErrSyntax, 0},
		{":1.8", xobis.ErrSyntax, 0},
		{"1-:1.8", xobis.ErrSyntax, 2},
		{"1--0:1.8", xobis.ErrSyntax, 2},
		{"1:0-1.8", xobis.ErrSyntax, 3},
		{"1:2:1.8", xobis.ErrSyntax, 3},
		{"1.8.0.255", xobis.ErrSyntax, 9},
		{"1.8.0.1.255", xobis.ErrSyntax, 11},
		{"1.0.1.8.0.255.1", xobis.ErrSyntax, 13},
		{"1-0:1.8.0.255", xobis.ErrSyntax, 9},
		{"1-0:1.8.0*255*1", xobis.ErrSyntax, 13},
		{"1-0:1.8.0*255.1", xobis.ErrSyntax, 13},
		{"1.8*255.0", xobis.ErrSyntax, 7},
		{"1-0:1.8.0&255", xobis.ErrSyntax, 9},
		{" 1-0:1.8.0*255", xobis.ErrSyntax, 0},
		{"1-0:1.8.0*255 ", xobis.ErrSyntax, 13},
		{"1-0:1.8.0*255\n", xobis.ErrSyntax, 13},
		{"1-0:1.8.0*255\x00", xobis.ErrSyntax, 13},
		{"1-0:1.8.0*255junk", xobis.ErrSyntax, 13},
		{"1-0:1.8.0*255(123*kWh)", xobis.ErrSyntax, 13},
		{"1-0:1.8.0*+1", xobis.ErrSyntax, 10},
		{"+1-0:1.8.0*255", xobis.ErrSyntax, 0},
		{"1-0:1.8.0*-1", xobis.ErrSyntax, 10},
		{"1-0:0x01.8.0*255", xobis.ErrSyntax, 5},
		{"1-0:1_0.8.0*255", xobis.ErrSyntax, 5},
		{"１-0:1.8.0*255", xobis.ErrSyntax, 0},
		{"1-0:١.8.0*255", xobis.ErrSyntax, 4},
		{"1-0:1.8.0*\xff", xobis.ErrSyntax, 10},
		{"C.1.0", xobis.ErrSyntax, 0},
		{"*.8.0", xobis.ErrSyntax, 0},
		{"1-0:1..0*255", xobis.ErrSyntax, 6},
		{"1-0:1.8.0*", xobis.ErrSyntax, 10},
		{"1-0:1.8.0*255x", xobis.ErrSyntax, 13},
		{"16-0:1.8.0*255", xobis.ErrRange, 0},
		{"255-0:1.8.0*255", xobis.ErrRange, 0},
		{"256-0:1.8.0*255", xobis.ErrRange, 0},
		{"1-256:1.8.0*255", xobis.ErrRange, 2},
		{"1-0:256.8.0*255", xobis.ErrRange, 4},
		{"1-0:1.256.0*255", xobis.ErrRange, 6},
		{"1-0:1.8.256*255", xobis.ErrRange, 8},
		{"1-0:1.8.0*256", xobis.ErrRange, 10},
		{"256.8", xobis.ErrRange, 0},
		{"1.256", xobis.ErrRange, 2},
		{"1.8.256", xobis.ErrRange, 4},
		{"1.8*256", xobis.ErrRange, 4},
		{"16.0.1.8.0.255", xobis.ErrRange, 0},
		{"1.256.1.8.0.255", xobis.ErrRange, 2},
		{"1.0.256.8.0.255", xobis.ErrRange, 4},
		{"1.0.1.256.0.255", xobis.ErrRange, 6},
		{"1.0.1.8.256.255", xobis.ErrRange, 8},
		{"1.0.1.8.0.256", xobis.ErrRange, 10},
		{"1-0:1.8.0*" + strings.Repeat("9", 4096), xobis.ErrRange, 10},
	}
}

func assertParseError(t *testing.T, input string, err, kind error) {
	t.Helper()
	var pe *xobis.ParseError
	assert.ErrorIs(t, err, kind)
	require.ErrorAs(t, err, &pe)
	assert.Equal(t, input, pe.Input)
	assert.GreaterOrEqual(t, pe.Offset, 0)
	assert.LessOrEqual(t, pe.Offset, len(input))
}
