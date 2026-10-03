package xobis_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

func TestCodePattern(t *testing.T) {
	// Arrange.
	tests := []struct {
		name   string
		groups xobis.Groups
		text   string
	}{
		{"zero", xobis.Groups{}, "0-0:0.0.0*0"},
		{"energy", xobis.Groups{A: 1, C: 1, D: 8, F: 255}, "1-0:1.8.0*255"},
		{"explicit zero storage", xobis.Groups{A: 1, C: 1, D: 8}, "1-0:1.8.0*0"},
		{"distinct groups", xobis.Groups{A: 1, B: 2, C: 3, D: 4, E: 5, F: 6}, "1-2:3.4.5*6"},
		{"upper bounds", xobis.Groups{A: 15, B: 255, C: 255, D: 255, E: 255, F: 255}, "15-255:255.255.255*255"},
		{"reserved values", xobis.Groups{A: 2, B: 200, C: 240, D: 128, E: 254, F: 128}, "2-200:240.128.254*128"},
	}
	all := xobis.PresentA | xobis.PresentB | xobis.PresentC | xobis.PresentD | xobis.PresentE | xobis.PresentF
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			code := newTestCode(t, tt.groups)
			want, err := xobis.ParsePattern(tt.text)
			require.NoError(t, err)

			// Act.
			pattern := code.Pattern()
			groups, presence := pattern.Groups()
			validationErr := pattern.Validate()
			roundtrip, roundtripErr := pattern.CompleteCode()
			matched := pattern.Match(code)

			// Assert.
			assert.NoError(t, validationErr)
			assert.Equal(t, want, pattern)
			assert.Equal(t, tt.groups, groups)
			assert.Equal(t, all, presence)
			assert.True(t, matched)
			require.NoError(t, roundtripErr)
			assert.Equal(t, code, roundtrip)
		})
	}
}

func TestCodePatternInteroperability(t *testing.T) {
	// Arrange.
	code, err := xobis.ParseHex("0100010800ff")
	require.NoError(t, err)
	abbreviated, err := xobis.ParsePattern("1-0:1.8.0")
	require.NoError(t, err)
	zeroStorage, err := xobis.ParsePattern("1-0:1.8.0*0")
	require.NoError(t, err)

	// Act.
	complete := code.Pattern()
	covered := abbreviated.Covers(complete)
	reverseCovered := complete.Covers(abbreviated)
	overlaps := abbreviated.Overlaps(complete)
	zeroOverlaps := zeroStorage.Overlaps(complete)

	// Assert. Converting the hexadecimal code preserves its explicit F=255.
	assert.True(t, covered)
	assert.False(t, reverseCovered)
	assert.True(t, overlaps)
	assert.False(t, zeroOverlaps)
}

func ExampleCode_Pattern() {
	// Arrange.
	code, err := xobis.ParseHex("0100010800ff")
	if err != nil {
		panic(err)
	}
	abbreviated, err := xobis.ParsePattern("1-0:1.8.0")
	if err != nil {
		panic(err)
	}

	// Act.
	pattern := code.Pattern()
	fmt.Println(pattern)
	fmt.Println(abbreviated.Covers(pattern))
	fmt.Println(pattern.Covers(abbreviated))

	// Output:
	// 1-0:1.8.0*255
	// true
	// false
}
