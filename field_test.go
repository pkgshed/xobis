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
	_ encoding.TextMarshaler   = xobis.CodeField{}
	_ encoding.TextMarshaler   = xobis.PatternField{}
	_ encoding.TextUnmarshaler = (*xobis.CodeField)(nil)
	_ encoding.TextUnmarshaler = (*xobis.PatternField)(nil)
)

func TestCodeFieldDecoding(t *testing.T) {
	// Arrange.
	tests := []struct {
		input string
		kind  error
	}{
		{"0-0:0.0.0*0", nil},
		{"01.000.002.08.000.0255", nil},
		{"", xobis.ErrSyntax},
		{"1.8.0", xobis.ErrSyntax},
		{"0100010800ff", xobis.ErrRange},
		{"1-0:1.8.0*256", xobis.ErrRange},
		{"16-0:1.8.0*255", xobis.ErrRange},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Arrange.
			original := newTestCode(t, xobis.Groups{A: 1, C: 1, D: 8, F: 255})
			field := xobis.NewCodeField(original)
			copy := field
			snapshot := field.Value()
			keys := map[xobis.Code]string{snapshot: "original"}
			input := []byte(tt.input)
			want := original
			if tt.kind == nil {
				var err error
				want, err = xobis.Parse(tt.input)
				require.NoError(t, err)
			}

			// Act.
			err := field.UnmarshalText(input)
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

func TestPatternFieldDecoding(t *testing.T) {
	// Arrange.
	tests := []struct {
		input string
		kind  error
	}{
		{"2.8", nil},
		{"1.8*0", nil},
		{"0-0:0.0.0*0", nil},
		{"", xobis.ErrSyntax},
		{"1.8.0*", xobis.ErrSyntax},
		{"1.8.256", xobis.ErrRange},
		{"16-1.8", xobis.ErrRange},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Arrange.
			original, err := xobis.ParsePattern("0:1.8.0")
			require.NoError(t, err)
			field := xobis.NewPatternField(original)
			copy := field
			snapshot := field.Value()
			keys := map[xobis.Pattern]string{snapshot: "original"}
			input := []byte(tt.input)
			want := original
			if tt.kind == nil {
				want, err = xobis.ParsePattern(tt.input)
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

func TestFieldZeroValues(t *testing.T) {
	// Arrange.
	var code xobis.CodeField
	var pattern xobis.PatternField

	// Act.
	codeValue := code.Value()
	patternValue := pattern.Value()
	codeText, codeErr := code.MarshalText()
	patternText, patternErr := pattern.MarshalText()
	wrappedCode := xobis.NewCodeField(xobis.Code{})
	wrappedPattern := xobis.NewPatternField(xobis.Pattern{})

	// Assert.
	assert.Zero(t, codeValue)
	assert.Zero(t, patternValue)
	require.NoError(t, codeErr)
	assert.Equal(t, "0-0:0.0.0*0", string(codeText))
	assert.ErrorIs(t, patternErr, xobis.ErrSyntax)
	assert.Nil(t, patternText)
	assert.Equal(t, code, wrappedCode)
	assert.Equal(t, pattern, wrappedPattern)
}

func TestFieldJSONRoundTrip(t *testing.T) {
	// Arrange.
	type payload struct {
		Code    xobis.CodeField    `json:"code"`
		Pattern xobis.PatternField `json:"pattern"`
	}
	wantCode := newTestCode(t, xobis.Groups{A: 1, C: 1, D: 8, F: 255})
	wantPattern, err := xobis.ParsePattern("1.8.0")
	require.NoError(t, err)
	var decoded payload
	input := []byte(`{"code":"01.0.1.8.0.255","pattern":"01.08.000"}`)

	// Act.
	unmarshalErr := json.Unmarshal(input, &decoded)
	text, marshalErr := json.Marshal(decoded)
	code := decoded.Code.Value()
	pattern := decoded.Pattern.Value()
	codeText, coreCodeErr := json.Marshal(code)
	patternText, corePatternErr := json.Marshal(pattern)

	// Assert.
	require.NoError(t, unmarshalErr)
	require.NoError(t, marshalErr)
	assert.Equal(t, `{"code":"1-0:1.8.0*255","pattern":"1.8.0"}`, string(text))
	assert.Equal(t, wantCode, code)
	assert.Equal(t, wantPattern, pattern)
	require.NoError(t, coreCodeErr)
	require.NoError(t, corePatternErr)
	assert.Equal(t, `"1-0:1.8.0*255"`, string(codeText))
	assert.Equal(t, `"1.8.0"`, string(patternText))
}

func TestFieldJSONMapKeys(t *testing.T) {
	t.Run("code", func(t *testing.T) {
		// Arrange.
		code := newTestCode(t, xobis.Groups{A: 1, C: 1, D: 8, F: 255})
		input := map[xobis.Code]string{code: "energy"}
		var decoded map[xobis.CodeField]string

		// Act.
		text, marshalErr := json.Marshal(input)
		unmarshalErr := json.Unmarshal(text, &decoded)
		label := decoded[xobis.NewCodeField(code)]

		// Assert.
		require.NoError(t, marshalErr)
		require.NoError(t, unmarshalErr)
		assert.Equal(t, "energy", label)
	})
	t.Run("pattern", func(t *testing.T) {
		// Arrange.
		omitted, err := xobis.ParsePattern("1.8")
		require.NoError(t, err)
		explicitZero, err := xobis.ParsePattern("1.8.0")
		require.NoError(t, err)
		input := map[xobis.Pattern]string{omitted: "all tariffs", explicitZero: "total"}
		var decoded map[xobis.PatternField]string

		// Act.
		text, marshalErr := json.Marshal(input)
		unmarshalErr := json.Unmarshal(text, &decoded)

		// Assert.
		require.NoError(t, marshalErr)
		require.NoError(t, unmarshalErr)
		assert.Len(t, decoded, 2)
		assert.Equal(t, "all tariffs", decoded[xobis.NewPatternField(omitted)])
		assert.Equal(t, "total", decoded[xobis.NewPatternField(explicitZero)])
	})
}

func ExampleCodeField() {
	// Arrange.
	var config struct {
		OBIS xobis.CodeField `json:"obis"`
	}
	input := []byte(`{"obis":"1-0:1.8.0*255"}`)

	// Act.
	if err := json.Unmarshal(input, &config); err != nil {
		panic(err)
	}
	code := config.OBIS.Value()
	fmt.Println(code)

	// Output:
	// 1-0:1.8.0*255
}

func ExamplePatternField() {
	// Arrange.
	var field xobis.PatternField

	// Act.
	if err := field.UnmarshalText([]byte("1.8.0")); err != nil {
		panic(err)
	}
	pattern := field.Value()
	_, present := pattern.Groups()
	fmt.Println(pattern)
	fmt.Println(present.PresentA())
	fmt.Println(present.PresentE())

	// Output:
	// 1.8.0
	// false
	// true
}
