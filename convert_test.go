package xobis_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

func TestCodeConversions(t *testing.T) {
	// Arrange.
	tests := []struct {
		name   string
		groups xobis.Groups
		bytes  [6]byte
		hex    string
		value  uint64
	}{
		{"energy", xobis.Groups{A: 1, C: 1, D: 8, F: 255},
			[6]byte{1, 0, 1, 8, 0, 255}, "0100010800ff", 0x0100010800ff},
		{"zero", xobis.Groups{}, [6]byte{}, "000000000000", 0},
		{"leading zeroes", xobis.Groups{F: 1}, [6]byte{0, 0, 0, 0, 0, 1}, "000000000001", 1},
		{"distinct groups", xobis.Groups{A: 1, B: 2, C: 3, D: 4, E: 5, F: 6},
			[6]byte{1, 2, 3, 4, 5, 6}, "010203040506", 0x010203040506},
		{"maximum values", xobis.Groups{A: 15, B: 255, C: 255, D: 255, E: 255, F: 255},
			[6]byte{15, 255, 255, 255, 255, 255}, "0fffffffffff", 0x0fffffffffff},
		{"hexadecimal letters", xobis.Groups{A: 10, B: 188, C: 222, D: 241, E: 35, F: 69},
			[6]byte{10, 188, 222, 241, 35, 69}, "0abcdef12345", 0x0abcdef12345},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			code, err := xobis.NewCode(tt.groups)
			require.NoError(t, err)

			// Act.
			raw := xobis.ToBytes(code)
			text := xobis.ToHex(code)
			value := xobis.ToUint64(code)
			bytesRoundtrip, bytesRoundtripErr := xobis.FromBytes(raw[:])
			hexRoundtrip, hexRoundtripErr := xobis.ParseHex(text)
			uintRoundtrip, uintRoundtripErr := xobis.FromUint64(value)

			// Assert. Known outputs also catch paired encoder/decoder mistakes.
			assert.Equal(t, tt.bytes, raw)
			assert.Equal(t, tt.hex, text)
			assert.Equal(t, tt.value, value)
			require.NoError(t, bytesRoundtripErr)
			require.NoError(t, hexRoundtripErr)
			require.NoError(t, uintRoundtripErr)
			assert.Equal(t, code, bytesRoundtrip)
			assert.Equal(t, code, hexRoundtrip)
			assert.Equal(t, code, uintRoundtrip)
		})
	}
}

func TestToBytesReturnsIndependentArray(t *testing.T) {
	// Arrange.
	want := xobis.Groups{A: 1, B: 2, C: 3, D: 4, E: 5, F: 6}
	code, err := xobis.NewCode(want)
	require.NoError(t, err)

	// Act.
	first := xobis.ToBytes(code)
	second := xobis.ToBytes(code)
	clear(first[:])

	// Assert.
	assert.Equal(t, [6]byte{1, 2, 3, 4, 5, 6}, second)
	assert.Equal(t, want, code.Groups())
}

func ExampleToBytes() {
	// Arrange.
	code, err := xobis.NewCode(xobis.Groups{A: 1, C: 1, D: 8, F: 255})
	if err != nil {
		panic(err)
	}

	// Act.
	raw := xobis.ToBytes(code)
	fmt.Println(raw)

	// Output:
	// [1 0 1 8 0 255]
}

func ExampleToHex() {
	// Arrange.
	code, err := xobis.NewCode(xobis.Groups{A: 1, C: 1, D: 8, F: 255})
	if err != nil {
		panic(err)
	}

	// Act.
	text := xobis.ToHex(code)
	fmt.Println(text)

	// Output:
	// 0100010800ff
}

func ExampleToUint64() {
	// Arrange.
	code, err := xobis.NewCode(xobis.Groups{A: 1, C: 1, D: 8, F: 255})
	if err != nil {
		panic(err)
	}

	// Act.
	value := xobis.ToUint64(code)
	fmt.Printf("0x%012x\n", value)

	// Output:
	// 0x0100010800ff
}
