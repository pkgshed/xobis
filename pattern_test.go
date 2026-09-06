package xobis_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

func TestNewPattern(t *testing.T) {
	// Arrange.
	groups := xobis.Groups{A: 1, B: 2, C: 3, D: 4, E: 5, F: 6}
	required := xobis.PresentC | xobis.PresentD
	all := xobis.PresentA | xobis.PresentB | required | xobis.PresentE | xobis.PresentF
	tests := []struct {
		name     string
		groups   xobis.Groups
		optional xobis.Presence
		text     string
		wantErr  error
	}{
		{"required only", groups, 0, "3.4", nil},
		{"A", groups, xobis.PresentA, "1-3.4", nil},
		{"B", groups, xobis.PresentB, "2:3.4", nil},
		{"AB", groups, xobis.PresentA | xobis.PresentB, "1-2:3.4", nil},
		{"E", groups, xobis.PresentE, "3.4.5", nil},
		{"AE", groups, xobis.PresentA | xobis.PresentE, "1-3.4.5", nil},
		{"BE", groups, xobis.PresentB | xobis.PresentE, "2:3.4.5", nil},
		{"ABE", groups, xobis.PresentA | xobis.PresentB | xobis.PresentE, "1-2:3.4.5", nil},
		{"F", groups, xobis.PresentF, "3.4*6", nil},
		{"AF", groups, xobis.PresentA | xobis.PresentF, "1-3.4*6", nil},
		{"BF", groups, xobis.PresentB | xobis.PresentF, "2:3.4*6", nil},
		{"ABF", groups, xobis.PresentA | xobis.PresentB | xobis.PresentF, "1-2:3.4*6", nil},
		{"EF", groups, xobis.PresentE | xobis.PresentF, "3.4.5*6", nil},
		{"AEF", groups, xobis.PresentA | xobis.PresentE | xobis.PresentF, "1-3.4.5*6", nil},
		{"BEF", groups, xobis.PresentB | xobis.PresentE | xobis.PresentF, "2:3.4.5*6", nil},
		{"ABEF", groups, all &^ required, "1-2:3.4.5*6", nil},
		{"explicit zeroes", xobis.Groups{}, all &^ required, "0-0:0.0.0*0", nil},
		{"zeroes with optional groups omitted", xobis.Groups{}, 0, "0.0", nil},
		{"upper bounds", xobis.Groups{A: 15, B: 255, C: 255, D: 255, E: 255, F: 255}, all &^ required, "15-255:255.255.255*255", nil},
		{"reserved and manufacturer values", xobis.Groups{A: 2, B: 200, C: 240, D: 128, E: 254, F: 128}, all &^ required, "2-200:240.128.254*128", nil},
		{"A out of range", xobis.Groups{A: 16}, all &^ required, "", xobis.ErrRange},
		{"A maximum byte", xobis.Groups{A: 255}, all &^ required, "", xobis.ErrRange},
		{"omitted values are ignored", xobis.Groups{A: 255, B: 255, C: 1, D: 8, E: 255, F: 255}, 0, "1.8", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			present := tt.optional | required
			var want xobis.Pattern
			var candidate xobis.Code
			if tt.wantErr == nil {
				var err error
				want, err = xobis.ParsePattern(tt.text)
				require.NoError(t, err)
				candidateGroups := tt.groups
				if !present.PresentA() {
					// The candidate must have a valid medium even when the pattern omits it.
					candidateGroups.A = xobis.MediumElectricity
				}
				candidate = newTestCode(t, candidateGroups)
			}
			patterns := map[xobis.Pattern]string{want: "found"}

			// Act.
			p, err := xobis.NewPattern(tt.groups, present)
			validationErr := p.Validate()
			text, marshalErr := p.MarshalText()
			formatted := p.String()
			matched := p.Match(candidate)
			lookup := patterns[p]

			// Assert.
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Zero(t, p)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, want, p)
			assert.Equal(t, present, p.Presence())
			assert.NoError(t, validationErr)
			require.NoError(t, marshalErr)
			assert.Equal(t, tt.text, string(text))
			assert.Equal(t, tt.text, formatted)
			assert.True(t, matched)
			assert.Equal(t, "found", lookup)
		})
	}
}

func TestNewPatternPresenceDomain(t *testing.T) {
	for value := range 256 {
		t.Run(fmt.Sprintf("%02x", value), func(t *testing.T) {
			// Arrange.
			present := xobis.Presence(value)
			var code xobis.Groups
			// Six defined bits, with bits 2 (C) and 3 (D) required.
			valid := value < 64 && value&12 == 12

			// Act.
			p, err := xobis.NewPattern(code, present)
			groups, mask := p.Groups()

			// Assert.
			if valid {
				require.NoError(t, err)
				assert.Equal(t, present, p.Presence())
				assert.Zero(t, groups)
				assert.Equal(t, present, mask)
			} else {
				require.ErrorIs(t, err, xobis.ErrSyntax)
				assert.Zero(t, p)
			}
		})
	}
}

func ExampleNewPattern() {
	// Arrange.
	code := xobis.Groups{
		C: xobis.ElectricityActivePowerImport,
		D: xobis.ElectricityTimeIntegral1,
		E: xobis.ElectricityTariffTotal,
	}
	present := xobis.PresentC | xobis.PresentD | xobis.PresentE

	// Act.
	p, err := xobis.NewPattern(code, present)
	if err != nil {
		panic(err)
	}
	fmt.Println(p)
	fmt.Println(p.Presence().PresentA())
	fmt.Println(p.Presence().PresentE())

	// Output:
	// 1.8.0
	// false
	// true
}
