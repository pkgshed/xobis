package xobis_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

func TestParseDecimal(t *testing.T) {
	// Arrange.
	for _, input := range []string{
		"0",
		"123",
		"+00123.4500",
		"-0",
		"-0.000",
		"+0.00",
		"0.5",
		"9007199254740993.0000000000000000001",
		strings.Repeat("9", 4096) + "." + strings.Repeat("0", 4096),
	} {
		t.Run(input[:min(len(input), 40)], func(t *testing.T) {
			// Act.
			value, err := xobis.ParseDecimal(input)
			text := value.String()
			validationErr := value.Validate()
			roundtrip, roundtripErr := xobis.ParseDecimal(text)

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, input, text)
			assert.NoError(t, validationErr)
			require.NoError(t, roundtripErr)
			assert.Equal(t, value, roundtrip)
		})
	}
}

func TestParseDecimalErrors(t *testing.T) {
	// Arrange.
	for _, tt := range []struct {
		input  string
		offset int
	}{
		{"", 0},
		{"+", 1},
		{"-", 1},
		{".1", 0},
		{"1.", 2},
		{"-.1", 1},
		{"1e2", 1},
		{"1E+2", 1},
		{"1,2", 1},
		{"1_000", 1},
		{"1.2.3", 3},
		{"NaN", 0},
		{"Inf", 0},
		{"-Inf", 1},
		{" 1", 0},
		{"1 ", 1},
		{"1\n", 1},
		{"1\r\n", 1},
		{"1\x00", 1},
		{"++1", 1},
		{"١", 0},
		{"1\xff", 1},
	} {
		t.Run(fmt.Sprintf("%q", tt.input), func(t *testing.T) {
			// Arrange.
			var pe *xobis.ParseError

			// Act.
			value, err := xobis.ParseDecimal(tt.input)

			// Assert.
			assertParseError(t, tt.input, err, xobis.ErrSyntax)
			require.ErrorAs(t, err, &pe)
			assert.Equal(t, tt.offset, pe.Offset)
			assert.Zero(t, value)
		})
	}
}

func TestZeroDecimal(t *testing.T) {
	// Arrange.
	var value xobis.Decimal

	// Act.
	err := value.Validate()
	text := value.String()

	// Assert.
	assert.ErrorIs(t, err, xobis.ErrSyntax)
	assert.Empty(t, text)
}

func ExampleParseDecimal() {
	// Arrange.
	input := "+00123.4500"

	// Act.
	value, err := xobis.ParseDecimal(input)
	if err != nil {
		panic(err)
	}
	fmt.Println(value)

	// Output:
	// +00123.4500
}
