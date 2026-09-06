package xobis_test

import (
	"encoding"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

var (
	_ encoding.TextMarshaler   = xobis.Reading{}
	_ encoding.TextMarshaler   = xobis.ReadingField{}
	_ encoding.TextUnmarshaler = (*xobis.ReadingField)(nil)
)

func TestReadingFieldDecoding(t *testing.T) {
	// Arrange.
	tests := []struct {
		input string
		kind  error
	}{
		{"2.8(-001.2500*°C)", nil},
		{"1.8(0)", nil},
		{"1.8(0)\n", xobis.ErrSyntax},
		{"1.8(0)\r\n", xobis.ErrSyntax},
		{"", xobis.ErrSyntax},
		{"1.8(1*)", xobis.ErrSyntax},
		{"1.8(1*k\xffWh)", xobis.ErrSyntax},
		{"1.8(1)(2)", xobis.ErrSyntax},
		{"16-1.8(1)", xobis.ErrRange},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.input), func(t *testing.T) {
			// Arrange.
			original, err := xobis.ParseReading("0:1.8.0(+123.00*kWh)")
			require.NoError(t, err)
			field := xobis.NewReadingField(original)
			copy := field
			snapshot := field.Value()
			keys := map[xobis.Reading]string{snapshot: "original"}
			input := []byte(tt.input)
			want := original
			if tt.kind == nil {
				want, err = xobis.ParseReading(tt.input)
				require.NoError(t, err)
			}

			// Act.
			err = field.UnmarshalText(input)
			clear(input)
			value := field.Value()
			text, marshalErr := field.MarshalText()
			lookup := keys[snapshot]

			// Assert.
			if tt.kind == nil {
				require.NoError(t, err)
			} else {
				assertParseError(t, tt.input, err, tt.kind)
			}
			assert.Equal(t, want, value)
			assert.Equal(t, original, snapshot)
			assert.Equal(t, original, copy.Value())
			assert.Equal(t, "original", lookup)
			require.NoError(t, marshalErr)
			assert.Equal(t, want.String(), string(text))
		})
	}
}

func TestZeroReadingField(t *testing.T) {
	// Arrange.
	var field xobis.ReadingField

	// Act.
	value := field.Value()
	text, err := field.MarshalText()
	wrapped := xobis.NewReadingField(xobis.Reading{})

	// Assert.
	assert.Zero(t, value)
	assert.ErrorIs(t, err, xobis.ErrSyntax)
	assert.Nil(t, text)
	assert.Equal(t, field, wrapped)
}

func TestReadingFieldJSON(t *testing.T) {
	// Arrange.
	type payload struct {
		Reading xobis.ReadingField `json:"reading"`
	}
	var decoded payload
	input := []byte(`{"reading":"01.08.000(+00123.4500*kWh)"}`)
	want := `{"reading":"1.8.0(+00123.4500*kWh)"}`

	// Act.
	unmarshalErr := json.Unmarshal(input, &decoded)
	text, marshalErr := json.Marshal(decoded)
	coreText, coreErr := json.Marshal(decoded.Reading.Value())

	// Assert.
	require.NoError(t, unmarshalErr)
	require.NoError(t, marshalErr)
	assert.Equal(t, want, string(text))
	require.NoError(t, coreErr)
	assert.Equal(t, `"1.8.0(+00123.4500*kWh)"`, string(coreText))
}

func TestReadingFieldJSONFailureIsAtomic(t *testing.T) {
	// Arrange.
	reading, err := xobis.ParseReading("1.8(123*kWh)")
	require.NoError(t, err)
	for _, input := range []string{`"1.8(NaN*kWh)"`, `"1.8(1*kWh)\n"`, `"1.8(1*kWh)\r\n"`} {
		t.Run(input, func(t *testing.T) {
			// Arrange.
			field := xobis.NewReadingField(reading)
			text := []byte(input)

			// Act.
			err := json.Unmarshal(text, &field)

			// Assert.
			assert.ErrorIs(t, err, xobis.ErrSyntax)
			assert.Equal(t, reading, field.Value())
		})
	}
}

func TestReadingJSONMapKeys(t *testing.T) {
	// Arrange.
	first, err := xobis.ParseReading("1.8(123*kWh)")
	require.NoError(t, err)
	second, err := xobis.ParseReading("1.8.0(123.00*kWh)")
	require.NoError(t, err)
	input := map[xobis.Reading]string{first: "first", second: "second"}
	var decoded map[xobis.ReadingField]string

	// Act.
	text, marshalErr := json.Marshal(input)
	unmarshalErr := json.Unmarshal(text, &decoded)

	// Assert.
	require.NoError(t, marshalErr)
	require.NoError(t, unmarshalErr)
	assert.Len(t, decoded, 2)
	assert.Equal(t, "first", decoded[xobis.NewReadingField(first)])
	assert.Equal(t, "second", decoded[xobis.NewReadingField(second)])
}

func ExampleReadingField() {
	// Arrange.
	var payload struct {
		Reading xobis.ReadingField `json:"reading"`
	}
	input := []byte(`{"reading":"1.8.0(123.4500*kWh)"}`)

	// Act.
	if err := json.Unmarshal(input, &payload); err != nil {
		panic(err)
	}
	reading := payload.Reading.Value()
	fmt.Println(reading.Value())
	fmt.Println(reading.Unit())

	// Output:
	// 123.4500
	// kWh
}
