package xobis_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

func TestPatternRelations(t *testing.T) {
	// Arrange. An empty input denotes the invalid zero Pattern.
	tests := []struct {
		name              string
		left, right       string
		covers, coveredBy bool
		overlaps          bool
	}{
		{"identical abbreviated", "1-0:1.8.0", "1-0:1.8.0", true, true, true},
		{"identical complete", "1-0:1.8.0*255", "1-0:1.8.0*255", true, true, true},
		{"broad filter and BSI identifier", "1.8.0", "1-0:1.8.0", true, false, true},
		{"abbreviated and complete identifier", "1-0:1.8.0", "1-0:1.8.0*255", true, false, true},
		{"unknown storage", "1-0:1.8.0*255", "1-0:1.8.0", false, true, true},
		{"independent constraints", "1-1.8", "0:1.8", false, false, true},
		{"omitted medium and explicit zero", "1.8", "0-1.8", true, false, true},
		{"omitted channel and explicit zero", "1.8", "0:1.8", true, false, true},
		{"omitted classification and explicit zero", "1.8", "1.8.0", true, false, true},
		{"omitted storage and explicit zero", "1.8", "1.8*0", true, false, true},
		{"omitted storage and explicit 255", "1.8", "1.8*255", true, false, true},
		{"different medium", "0-1.8", "1-1.8", false, false, false},
		{"different channel", "0:1.8", "1:1.8", false, false, false},
		{"different quantity", "1.8", "2.8", false, false, false},
		{"different processing", "1.7", "1.8", false, false, false},
		{"different classification", "1.8.0", "1.8.1", false, false, false},
		{"different storage", "1.8*0", "1.8*255", false, false, false},
		{"upper bounds", "15-255:255.255.255", "15-255:255.255.255*255", true, false, true},
		{"valid all-zero groups", "0.0", "0-0:0.0.0*0", true, false, true},
		{"invalid receiver", "", "1.8", false, false, false},
		{"invalid argument", "1.8", "", false, false, false},
		{"both invalid", "", "", false, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var patterns [2]xobis.Pattern
			for i, input := range []string{tt.left, tt.right} {
				if input != "" {
					var err error
					patterns[i], err = xobis.ParsePattern(input)
					require.NoError(t, err)
				}
			}
			left, right := patterns[0], patterns[1]

			// Act.
			covers := left.Covers(right)
			coveredBy := right.Covers(left)
			overlaps := left.Overlaps(right)
			reverseOverlap := right.Overlaps(left)

			// Assert.
			assert.Equal(t, tt.covers, covers)
			assert.Equal(t, tt.coveredBy, coveredBy)
			assert.Equal(t, tt.overlaps, overlaps)
			assert.Equal(t, tt.overlaps, reverseOverlap)
		})
	}
}

func TestPatternRelationsAgainstMatchedCodes(t *testing.T) {
	// Arrange. Enumerate all presence combinations and two distinct values per
	// optional group. Their sets of matched codes independently define coverage
	// and overlap, including explicit zeroes and F=255.
	var codes [16]xobis.Code
	for values := range len(codes) {
		codes[values] = newTestCode(t, xobis.Groups{
			A: xobis.Medium(values & 1), B: xobis.Channel((values >> 1) & 1),
			C: 1, D: 8, E: xobis.Classification((values >> 2) & 1),
			F: xobis.Storage(255 * ((values >> 3) & 1)),
		})
	}
	optional := []xobis.Presence{xobis.PresentA, xobis.PresentB, xobis.PresentE, xobis.PresentF}
	type sample struct {
		pattern xobis.Pattern
		matches uint16
	}
	var samples []sample
	for mask := range 16 {
		present := xobis.PresentC | xobis.PresentD
		for i, flag := range optional {
			if mask&(1<<i) != 0 {
				present |= flag
			}
		}
		for _, code := range codes {
			pattern, err := xobis.NewPattern(code.Groups(), present)
			require.NoError(t, err)
			var matches uint16
			for i, candidate := range codes {
				if pattern.Match(candidate) {
					matches |= 1 << i
				}
			}
			samples = append(samples, sample{pattern, matches})
		}
	}

	for _, left := range samples {
		for _, right := range samples {
			// Act.
			covers := left.pattern.Covers(right.pattern)
			overlaps := left.pattern.Overlaps(right.pattern)

			// Assert.
			assert.Equal(t, right.matches&^left.matches == 0, covers,
				"%s covers %s", left.pattern, right.pattern)
			assert.Equal(t, left.matches&right.matches != 0, overlaps,
				"%s overlaps %s", left.pattern, right.pattern)
		}
	}
}

func ExamplePattern_Covers() {
	// Arrange.
	filter, err := xobis.ParsePattern("1.8.0")
	if err != nil {
		panic(err)
	}
	identifier, err := xobis.ParsePattern("1-0:1.8.0")
	if err != nil {
		panic(err)
	}

	// Act.
	fmt.Println(filter.Covers(identifier))
	fmt.Println(identifier.Covers(filter))

	// Output:
	// true
	// false
}

func ExamplePattern_Overlaps() {
	// Arrange.
	filter, err := xobis.ParsePattern("1-0:1.8.0*255")
	if err != nil {
		panic(err)
	}
	identifier, err := xobis.ParsePattern("1-0:1.8.0")
	if err != nil {
		panic(err)
	}

	// Act. Missing F allows overlap without proving that it is 255.
	fmt.Println(filter.Overlaps(identifier))
	fmt.Println(filter.Covers(identifier))

	// Output:
	// true
	// false
}
