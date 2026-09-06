package xobis_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

func TestParseReading(t *testing.T) {
	// Arrange.
	tests := []struct {
		input, identifier, value string
		unit                     xobis.Unit
	}{
		{"1.8.0(123*kWh)", "1.8.0", "123", xobis.UnitKilowattHour},
		{"1.8(0)", "1.8", "0", xobis.UnitNone},
		{"0:1.8(+00123.4500*kWh)", "0:1.8", "+00123.4500", xobis.UnitKilowattHour},
		{"0-1.8(-0.000*W)", "0-1.8", "-0.000", xobis.UnitWatt},
		{"1-0:1.8.0*255(-123.45*kW)", "1-0:1.8.0*255", "-123.45", xobis.UnitKilowatt},
		{"01.000.001.08.000.0255(+0.00*Wh)", "1-0:1.8.0*255", "+0.00", xobis.UnitWattHour},
		{"1.8*0(230*V)", "1.8*0", "230", xobis.UnitVolt},
		{"1.8(1.25*A)", "1.8", "1.25", xobis.UnitAmpere},
		{"1.8(50*Hz)", "1.8", "50", xobis.UnitHertz},
		{"1.8(-10.25*°C)", "1.8", "-10.25", xobis.Unit("°C")},
		{"1.8(0.001*m³/h)", "1.8", "0.001", xobis.Unit("m³/h")},
		{"1.8(50*vendor-unit^2%)", "1.8", "50", xobis.Unit("vendor-unit^2%")},
		{"1.8(1*e\u0301)", "1.8", "1", xobis.Unit("e\u0301")},
		{"1.8(1*\ufffd)", "1.8", "1", xobis.Unit("\ufffd")},
		{"1.8(9007199254740993.0000000000000000001)", "1.8", "9007199254740993.0000000000000000001", xobis.UnitNone},
		{"1.8(" + strings.Repeat("9", 4096) + "*kWh)", "1.8", strings.Repeat("9", 4096), xobis.UnitKilowattHour},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.input[:min(len(tt.input), 60)]), func(t *testing.T) {
			// Arrange.
			identifier, err := xobis.ParsePattern(tt.identifier)
			require.NoError(t, err)
			value, err := xobis.ParseDecimal(tt.value)
			require.NoError(t, err)
			wantText := tt.identifier + "(" + tt.value
			if tt.unit != xobis.UnitNone {
				wantText += "*" + string(tt.unit)
			}
			wantText += ")"

			// Act.
			got, parseErr := xobis.ParseReading(tt.input)
			constructed, constructErr := xobis.NewReading(identifier, value, tt.unit)
			validationErr := got.Validate()
			text, marshalErr := got.MarshalText()
			formatted := got.String()
			roundtrip, roundtripErr := xobis.ParseReading(formatted)

			// Assert.
			require.NoError(t, parseErr)
			require.NoError(t, constructErr)
			assert.Equal(t, constructed, got)
			assert.Equal(t, identifier, got.Identifier())
			assert.Equal(t, value, got.Value())
			assert.Equal(t, tt.unit, got.Unit())
			assert.NoError(t, validationErr)
			require.NoError(t, marshalErr)
			assert.Equal(t, wantText, string(text))
			assert.Equal(t, wantText, formatted)
			require.NoError(t, roundtripErr)
			assert.Equal(t, got, roundtrip)
		})
	}
}

func TestParseReadingErrors(t *testing.T) {
	// Arrange.
	tests := []struct {
		input  string
		kind   error
		offset int
	}{
		{"", xobis.ErrSyntax, 0}, {"1.8", xobis.ErrSyntax, 3},
		{"(1)", xobis.ErrSyntax, 0}, {"1..8(1)", xobis.ErrSyntax, 2},
		{"16-1.8(1)", xobis.ErrRange, 0}, {"1.256(1)\r\n", xobis.ErrRange, 2},
		{"1.8(", xobis.ErrSyntax, 4}, {"1.8(1", xobis.ErrSyntax, 5},
		{"1.8(1\n", xobis.ErrSyntax, 5}, {"1.8(1)\r", xobis.ErrSyntax, 6},
		{"1.8(1)\n", xobis.ErrSyntax, 6}, {"1.8(1)\r\n", xobis.ErrSyntax, 6},
		{"1.8(1)\n\n", xobis.ErrSyntax, 6}, {"1.8(1)\r\n\r\n", xobis.ErrSyntax, 6},
		{"1.8(1)\n1.8(2)", xobis.ErrSyntax, 6}, {" 1.8(1)", xobis.ErrSyntax, 0},
		{"\n1.8(1)", xobis.ErrSyntax, 0}, {"\r\n1.8(1)", xobis.ErrSyntax, 0},
		{"1.8(1)\t", xobis.ErrSyntax, 6}, {"1.8(1)\u00a0", xobis.ErrSyntax, 6},
		{"1.8(1) ", xobis.ErrSyntax, 6}, {"1.8()", xobis.ErrSyntax, 4},
		{"1.8(+)", xobis.ErrSyntax, 5}, {"1.8(.1)", xobis.ErrSyntax, 4},
		{"1.8(1.)", xobis.ErrSyntax, 6}, {"1.8(1e2)", xobis.ErrSyntax, 5},
		{"1.8(1,2)", xobis.ErrSyntax, 5}, {"1.8(NaN)", xobis.ErrSyntax, 4},
		{"1.8(Inf)", xobis.ErrSyntax, 4}, {"1.8(1_000)", xobis.ErrSyntax, 5},
		{"1.8(1\n)", xobis.ErrSyntax, 5}, {"1.8(1 2)", xobis.ErrSyntax, 5},
		{"1.8(*kWh)", xobis.ErrSyntax, 4}, {"1.8(1*)", xobis.ErrSyntax, 6},
		{"1.8(1**kWh)", xobis.ErrSyntax, 6}, {"1.8(1*k W)", xobis.ErrSyntax, 7},
		{"1.8(1*k\u00a0W)", xobis.ErrSyntax, 7}, {"1.8(1*k\u200bW)", xobis.ErrSyntax, 7},
		{"1.8(1*°C\x00)", xobis.ErrSyntax, 9}, {"1.8(1*°C\xff)\r\n", xobis.ErrSyntax, 9},
		{"1.8(1*kWh*x)", xobis.ErrSyntax, 9}, {"1.8(1*k(Wh)", xobis.ErrSyntax, 7},
		{"1.8((1))", xobis.ErrSyntax, 4}, {"1.8(1)(2)", xobis.ErrSyntax, 6},
		{"1.8(1))", xobis.ErrSyntax, 6}, {"1.8(1)junk", xobis.ErrSyntax, 6},
		{"1.8(１)", xobis.ErrSyntax, 4}, {"1.8(1\xff)", xobis.ErrSyntax, 5},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.input), func(t *testing.T) {
			// Arrange.
			var pe *xobis.ParseError

			// Act.
			got, err := xobis.ParseReading(tt.input)

			// Assert.
			assertParseError(t, tt.input, err, tt.kind)
			require.ErrorAs(t, err, &pe)
			assert.Equal(t, tt.offset, pe.Offset)
			assert.Zero(t, got)
		})
	}
}

func TestNewReadingRejectsInvalidComponents(t *testing.T) {
	// Arrange.
	identifier, err := xobis.ParsePattern("1.8")
	require.NoError(t, err)
	value, err := xobis.ParseDecimal("123.00")
	require.NoError(t, err)
	tests := []struct {
		name       string
		identifier xobis.Pattern
		value      xobis.Decimal
		unit       xobis.Unit
	}{
		{"zero identifier", xobis.Pattern{}, value, xobis.UnitNone},
		{"zero decimal", identifier, xobis.Decimal{}, xobis.UnitNone},
	}
	for _, unit := range []xobis.Unit{"k W", "k\tW", "\n", "\x00", "\xff", "\u00a0", "\u200b", "k(Wh", "kWh)", "kWh*"} {
		tests = append(tests, struct {
			name       string
			identifier xobis.Pattern
			value      xobis.Decimal
			unit       xobis.Unit
		}{fmt.Sprintf("unit %q", unit), identifier, value, unit})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			got, err := xobis.NewReading(tt.identifier, tt.value, tt.unit)

			// Assert.
			assert.ErrorIs(t, err, xobis.ErrSyntax)
			assert.Zero(t, got)
		})
	}
}

func TestReadingEquality(t *testing.T) {
	// Arrange.
	tests := []struct {
		x, y  string
		equal bool
	}{
		{"1.8(123)", "01.08(123)", true},
		{"1.8(123)", "1.8.0(123)", false},
		{"1.8(123)", "1.8(123.00)", false},
		{"1.8(0)", "1.8(-0)", false},
		{"1.8(1)", "1.8(+1)", false},
		{"1.8(1*kWh)", "1.8(1*KWH)", false},
		{"1.8(1)", "1.8(1*kWh)", false},
	}
	for _, tt := range tests {
		t.Run(tt.x+"/"+tt.y, func(t *testing.T) {
			// Arrange.
			x, err := xobis.ParseReading(tt.x)
			require.NoError(t, err)
			y, err := xobis.ParseReading(tt.y)
			require.NoError(t, err)
			keys := map[xobis.Reading]bool{x: true}

			// Act.
			equal := x == y
			found := keys[y]

			// Assert.
			assert.Equal(t, tt.equal, equal)
			assert.Equal(t, tt.equal, found)
		})
	}
}

func TestZeroReading(t *testing.T) {
	// Arrange.
	var reading xobis.Reading

	// Act.
	err := reading.Validate()
	text, marshalErr := reading.MarshalText()
	formatted := reading.String()

	// Assert.
	assert.ErrorIs(t, err, xobis.ErrSyntax)
	assert.ErrorIs(t, marshalErr, xobis.ErrSyntax)
	assert.Nil(t, text)
	assert.Empty(t, formatted)
	assert.Zero(t, reading.Identifier())
	assert.Zero(t, reading.Value())
	assert.Equal(t, xobis.UnitNone, reading.Unit())
}

func ExampleParseReading() {
	// Arrange.
	input := "1.8.0(123.4500*kWh)"

	// Act.
	reading, err := xobis.ParseReading(input)
	if err != nil {
		panic(err)
	}
	fmt.Println(reading.Identifier())
	fmt.Println(reading.Value())
	fmt.Println(reading.Unit() == xobis.UnitKilowattHour)
	fmt.Println(reading)

	// Output:
	// 1.8.0
	// 123.4500
	// true
	// 1.8.0(123.4500*kWh)
}

func ExampleNewReading() {
	// Arrange.
	identifier, err := xobis.ParsePattern("1.8.0")
	if err != nil {
		panic(err)
	}
	value, err := xobis.ParseDecimal("+00123.4500")
	if err != nil {
		panic(err)
	}

	// Act.
	reading, err := xobis.NewReading(identifier, value, xobis.UnitKilowattHour)
	if err != nil {
		panic(err)
	}
	fmt.Println(reading)

	// Output:
	// 1.8.0(+00123.4500*kWh)
}
