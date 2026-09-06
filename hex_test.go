package xobis_test

import (
	"encoding/binary"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

func TestParseHex(t *testing.T) {
	// Arrange.
	tests := []struct {
		input string
		want  xobis.Groups
	}{
		{"0100010800ff", xobis.Groups{A: 1, C: 1, D: 8, F: 255}},
		{"0x0100010800ff", xobis.Groups{A: 1, C: 1, D: 8, F: 255}},
		{"0X0100010800FF", xobis.Groups{A: 1, C: 1, D: 8, F: 255}},
		{"000000000000", xobis.Groups{}},
		{"0x000000000000", xobis.Groups{}},
		{"000000000001", xobis.Groups{F: 1}},
		{"010203040506", xobis.Groups{A: 1, B: 2, C: 3, D: 4, E: 5, F: 6}},
		{"0fFfFfFfFfFf", xobis.Groups{A: 15, B: 255, C: 255, D: 255, E: 255, F: 255}},
		{"02c8f080fe80", xobis.Groups{A: 2, B: 200, C: 240, D: 128, E: 254, F: 128}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Act.
			got, err := xobis.ParseHex(tt.input)
			validationErr := got.Validate()
			roundtrip, roundtripErr := xobis.Parse(got.String())

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Groups())
			assert.NoError(t, validationErr)
			require.NoError(t, roundtripErr)
			assert.Equal(t, tt.want, roundtrip.Groups())
		})
	}
}

func TestParseHexErrors(t *testing.T) {
	// Arrange.
	tests := []struct {
		input  string
		kind   error
		offset int
	}{
		{"", xobis.ErrSyntax, 0},
		{"0x", xobis.ErrSyntax, 2},
		{"0X", xobis.ErrSyntax, 2},
		{"0100010800f", xobis.ErrSyntax, 11},
		{"0x100010800ff", xobis.ErrSyntax, 13},
		{"00100010800ff", xobis.ErrSyntax, 12},
		{"0x00100010800ff", xobis.ErrSyntax, 14},
		{"0100010800fg", xobis.ErrSyntax, 11},
		{"0x0100010800gf", xobis.ErrSyntax, 12},
		{"g100010800ff", xobis.ErrSyntax, 0},
		{"0x0g00010800ff", xobis.ErrSyntax, 3},
		{"0b0100010800ff", xobis.ErrSyntax, 12},
		{"+100010800ff", xobis.ErrSyntax, 0},
		{"-100010800ff", xobis.ErrSyntax, 0},
		{"0100_10800ff", xobis.ErrSyntax, 4},
		{"0100:10800ff", xobis.ErrSyntax, 4},
		{"01 0010800ff", xobis.ErrSyntax, 2},
		{"0100010800ff ", xobis.ErrSyntax, 12},
		{"0100010800f\n", xobis.ErrSyntax, 11},
		{"0100010800f\x00", xobis.ErrSyntax, 11},
		{"0100010800f\xff", xobis.ErrSyntax, 11},
		{"０0010800ff", xobis.ErrSyntax, 0},
		{"1000010800ff", xobis.ErrRange, 0},
		{"0x1000010800ff", xobis.ErrRange, 2},
		{"FF00010800ff", xobis.ErrRange, 0},
		{strings.Repeat("0", 4096), xobis.ErrSyntax, 12},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.input[:min(len(tt.input), 40)]), func(t *testing.T) {
			// Arrange.
			var pe *xobis.ParseError

			// Act.
			got, err := xobis.ParseHex(tt.input)

			// Assert.
			assert.Zero(t, got)
			assertParseError(t, tt.input, err, tt.kind)
			require.ErrorAs(t, err, &pe)
			assert.Equal(t, tt.offset, pe.Offset)
		})
	}
}

func TestFromUint64(t *testing.T) {
	// Arrange.
	tests := []struct {
		value uint64
		want  xobis.Groups
	}{
		{0x0100010800ff, xobis.Groups{A: 1, C: 1, D: 8, F: 255}},
		{0, xobis.Groups{}},
		{1, xobis.Groups{F: 1}},
		{0x010203040506, xobis.Groups{A: 1, B: 2, C: 3, D: 4, E: 5, F: 6}},
		{0x0fffffffffff, xobis.Groups{A: 15, B: 255, C: 255, D: 255, E: 255, F: 255}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%012x", tt.value), func(t *testing.T) {
			// Act.
			got, err := xobis.FromUint64(tt.value)
			validationErr := got.Validate()

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Groups())
			assert.NoError(t, validationErr)
		})
	}
}

func TestFromUint64RejectsOutOfRange(t *testing.T) {
	// Arrange.
	values := []uint64{0x100000000000, 0xff0000000000, 0xffffffffffff, ^uint64(0)}
	for bit := 48; bit < 64; bit++ {
		// Keep a valid code below each overflow bit to catch silent truncation.
		values = append(values, uint64(1)<<bit|0x0100010800ff)
	}
	for _, value := range values {
		t.Run(fmt.Sprintf("%x", value), func(t *testing.T) {
			// Act.
			got, err := xobis.FromUint64(value)

			// Assert.
			assert.ErrorIs(t, err, xobis.ErrRange)
			assert.Zero(t, got)
		})
	}
}

func TestEncodedNumericDomain(t *testing.T) {
	for group := range 6 {
		for value := range 256 {
			// Arrange.
			bytes := [8]byte{0, 0, 1, 0, 1, 8, 0, 255}
			bytes[group+2] = byte(value)
			input := fmt.Sprintf("%x", bytes[2:])
			packed := binary.BigEndian.Uint64(bytes[:])
			want := xobis.Groups{
				A: xobis.Medium(bytes[2]), B: xobis.Channel(bytes[3]),
				C: xobis.Quantity(bytes[4]), D: xobis.Processing(bytes[5]),
				E: xobis.Classification(bytes[6]), F: xobis.Storage(bytes[7]),
			}

			// Act.
			parsed, parseErr := xobis.ParseHex(input)
			constructed, constructErr := xobis.FromUint64(packed)
			fromBytes, bytesErr := xobis.FromBytes(bytes[2:])
			created, createErr := xobis.NewCode(want)
			raw := xobis.ToBytes(created)
			text := xobis.ToHex(created)
			number := xobis.ToUint64(created)

			// Assert.
			if group == 0 && value > 15 {
				assert.ErrorIs(t, parseErr, xobis.ErrRange, "input: %q", input)
				assert.ErrorIs(t, constructErr, xobis.ErrRange, "input: %q", input)
				assert.ErrorIs(t, bytesErr, xobis.ErrRange, "input: %q", input)
				assert.ErrorIs(t, createErr, xobis.ErrRange, "input: %q", input)
				assert.Zero(t, created)
				assert.Zero(t, parsed)
				assert.Zero(t, constructed)
				assert.Zero(t, fromBytes)
				continue
			}
			require.NoError(t, parseErr, "input: %q", input)
			require.NoError(t, constructErr, "input: %q", input)
			require.NoError(t, bytesErr, "input: %q", input)
			require.NoError(t, createErr, "input: %q", input)
			assert.Equal(t, want, created.Groups())
			assert.Equal(t, want, parsed.Groups(), "input: %q", input)
			assert.Equal(t, want, constructed.Groups(), "input: %q", input)
			assert.Equal(t, want, fromBytes.Groups(), "input: %q", input)
			assert.Equal(t, bytes[2:], raw[:], "input: %q", input)
			assert.Equal(t, input, text)
			assert.Equal(t, packed, number, "input: %q", input)
		}
	}
}

func ExampleParseHex() {
	// Arrange.
	input := "0100010800ff"

	// Act.
	code, err := xobis.ParseHex(input)
	if err != nil {
		panic(err)
	}
	fmt.Println(code)

	// Output:
	// 1-0:1.8.0*255
}

func ExampleFromUint64() {
	// Arrange.
	const packed = 0x0100010800ff

	// Act.
	code, err := xobis.FromUint64(packed)
	if err != nil {
		panic(err)
	}
	fmt.Println(code)

	// Output:
	// 1-0:1.8.0*255
}
