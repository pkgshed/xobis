package xobis_test

import (
	"encoding"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

var (
	_ encoding.TextMarshaler = xobis.Code{}
	_ encoding.TextMarshaler = xobis.Pattern{}
)

func TestNumericDomain(t *testing.T) {
	for group := range 6 {
		for value := range 256 {
			// Arrange.
			g := [6]int{1, 0, 1, 8, 0, 255}
			g[group] = value
			input := fmt.Sprintf("%d-%d:%d.%d.%d*%d", g[0], g[1], g[2], g[3], g[4], g[5])
			want := xobis.Groups{
				A: xobis.Medium(g[0]), B: xobis.Channel(g[1]), C: xobis.Quantity(g[2]),
				D: xobis.Processing(g[3]), E: xobis.Classification(g[4]), F: xobis.Storage(g[5]),
			}

			// Act.
			c, parseErr := xobis.Parse(input)
			validationErr := want.Validate()
			constructed, constructErr := xobis.NewCode(want)
			text, marshalErr := c.MarshalText()
			formatted := c.String()

			// Assert.
			if group == 0 && value > 15 {
				assert.ErrorIs(t, parseErr, xobis.ErrRange, "input: %q", input)
				assert.ErrorIs(t, validationErr, xobis.ErrRange, "input: %q", input)
				assert.ErrorIs(t, constructErr, xobis.ErrRange, "input: %q", input)
				assert.Zero(t, c)
				assert.Zero(t, constructed)
				continue
			}
			require.NoError(t, parseErr, "input: %q", input)
			assert.Equal(t, want, c.Groups(), "input: %q", input)
			require.NoError(t, constructErr)
			assert.Equal(t, c, constructed)
			assert.NoError(t, validationErr, "input: %q", input)
			require.NoError(t, marshalErr, "input: %q", input)
			assert.Equal(t, input, string(text))
			assert.Equal(t, input, formatted)
		}
	}
}

func TestPatternMatch(t *testing.T) {
	// Arrange.
	tests := []struct {
		pattern string
		code    string
		want    bool
	}{
		{"1.8", "1-0:1.8.0*255", true},
		{"1.8", "7-99:1.8.3*4", true},
		{"1.8", "1-0:2.8.0*255", false},
		{"1.8", "1-0:1.7.0*255", false},
		{"0-1.8", "0-2:1.8.0*255", true},
		{"0-1.8", "1-2:1.8.0*255", false},
		{"0:1.8", "1-0:1.8.0*255", true},
		{"0:1.8", "1-1:1.8.0*255", false},
		{"1.8.0", "1-0:1.8.0*255", true},
		{"1.8.0", "1-0:1.8.1*255", false},
		{"1.8*0", "1-0:1.8.2*0", true},
		{"1.8*0", "1-0:1.8.2*255", false},
		{"1.8*255", "1-0:1.8.2*255", true},
		{"1.8*255", "1-0:1.8.2*0", false},
		{"1-0:1.8.0*255", "1-0:1.8.0*255", true},
		{"1-0:1.8.0*255", "1-0:1.8.0*0", false},
	}
	for _, tt := range tests {
		t.Run(tt.pattern+"/"+tt.code, func(t *testing.T) {
			// Arrange.
			p, err := xobis.ParsePattern(tt.pattern)
			require.NoError(t, err)
			c, err := xobis.Parse(tt.code)
			require.NoError(t, err)

			// Act.
			got := p.Match(c)

			// Assert.
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMediumString(t *testing.T) {
	// Arrange.
	tests := []struct {
		medium xobis.Medium
		want   string
	}{
		{xobis.MediumElectricity, "MediumElectricity"},
		{xobis.Medium(2), "Medium(2)"},
		{xobis.Medium(255), "Medium(255)"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			// Act.
			got := tt.medium.String()

			// Assert.
			assert.Equal(t, tt.want, got)
		})
	}
}

func ExampleCode() {
	// Arrange.
	groups := xobis.Groups{
		A: xobis.MediumElectricity,
		B: xobis.ChannelUnspecified,
		C: xobis.ElectricityActivePowerImport,
		D: xobis.ElectricityTimeIntegral1,
		E: xobis.ElectricityTariffTotal,
		F: xobis.StorageNotUsed,
	}

	// Act.
	c, err := xobis.NewCode(groups)
	if err != nil {
		panic(err)
	}
	fmt.Println(c)
	fmt.Println(c.Groups().A)

	// Output:
	// 1-0:1.8.0*255
	// MediumElectricity
}

func ExampleParsePattern() {
	// Arrange.
	p, err := xobis.ParsePattern("1.8.0")
	if err != nil {
		panic(err)
	}
	c, err := xobis.Parse("1-0:1.8.0*255")
	if err != nil {
		panic(err)
	}

	// Act.
	fmt.Println(p.Presence().PresentA())
	fmt.Println(p.Presence().PresentE())
	fmt.Println(p.Match(c))

	// Output:
	// false
	// true
	// true
}
