package xobis_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

func TestFromBytes(t *testing.T) {
	// Arrange.
	tests := []struct {
		name  string
		input []byte
		want  xobis.Groups
	}{
		{"energy", []byte{1, 0, 1, 8, 0, 255}, xobis.Groups{A: 1, C: 1, D: 8, F: 255}},
		{"zero", make([]byte, 6), xobis.Groups{}},
		{"leading zeroes", []byte{0, 0, 0, 0, 0, 1}, xobis.Groups{F: 1}},
		{"distinct groups", []byte{1, 2, 3, 4, 5, 6}, xobis.Groups{A: 1, B: 2, C: 3, D: 4, E: 5, F: 6}},
		{"maximum values", []byte{15, 255, 255, 255, 255, 255}, xobis.Groups{A: 15, B: 255, C: 255, D: 255, E: 255, F: 255}},
		{"reserved values", []byte{2, 200, 240, 128, 254, 128}, xobis.Groups{A: 2, B: 200, C: 240, D: 128, E: 254, F: 128}},
		{"subslice", []byte{255, 1, 2, 3, 4, 5, 6, 255}[1:7], xobis.Groups{A: 1, B: 2, C: 3, D: 4, E: 5, F: 6}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			original := bytes.Clone(tt.input)

			// Act.
			got, err := xobis.FromBytes(tt.input)
			validationErr := got.Validate()

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Groups())
			assert.NoError(t, validationErr)
			assert.Equal(t, original, tt.input)
		})
	}
}

func TestFromBytesRejectsInvalidInput(t *testing.T) {
	// Arrange.
	tests := []struct {
		name  string
		input []byte
		kind  error
	}{
		{"nil", nil, xobis.ErrSyntax},
		{"hexadecimal text", []byte("0100010800ff"), xobis.ErrSyntax},
		{"invalid medium", []byte{16, 0, 1, 8, 0, 255}, xobis.ErrRange},
		{"maximum byte medium", []byte{255, 0, 1, 8, 0, 255}, xobis.ErrRange},
	}
	for _, length := range []int{0, 1, 2, 3, 4, 5, 7, 8, 4096} {
		tests = append(tests, struct {
			name  string
			input []byte
			kind  error
		}{fmt.Sprintf("length %d", length), make([]byte, length), xobis.ErrSyntax})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			original := bytes.Clone(tt.input)

			// Act.
			got, err := xobis.FromBytes(tt.input)

			// Assert.
			assert.ErrorIs(t, err, tt.kind)
			assert.Zero(t, got)
			assert.Equal(t, original, tt.input)
		})
	}
}

func TestFromBytesDoesNotRetainInput(t *testing.T) {
	// Arrange.
	input := []byte{1, 0, 1, 8, 0, 255}
	want := xobis.Groups{A: 1, C: 1, D: 8, F: 255}

	// Act.
	got, err := xobis.FromBytes(input)
	clear(input)

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, want, got.Groups())
}

func ExampleFromBytes() {
	// Arrange.
	input := []byte{0x01, 0x00, 0x01, 0x08, 0x00, 0xff}

	// Act.
	code, err := xobis.FromBytes(input)
	if err != nil {
		panic(err)
	}
	fmt.Println(code)

	// Output:
	// 1-0:1.8.0*255
}
