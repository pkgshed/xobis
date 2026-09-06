package xobis_test

import (
	"encoding"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

// Only valid fixtures use this helper; constructor rejection is tested directly.
func newTestCode(t testing.TB, groups xobis.Groups) xobis.Code {
	t.Helper()
	code, err := xobis.NewCode(groups)
	require.NoError(t, err)
	return code
}

func TestCoreImmutabilityContract(t *testing.T) {
	// Arrange.
	for _, value := range []any{xobis.Code{}, xobis.Pattern{}, xobis.Decimal{}, xobis.Reading{}} {
		t.Run(reflect.TypeOf(value).Name(), func(t *testing.T) {
			// Arrange.
			typ := reflect.TypeOf(value)
			pointer := reflect.PointerTo(typ)
			decoder := reflect.TypeFor[encoding.TextUnmarshaler]()

			// Act.
			decodes := pointer.Implements(decoder)
			isComparable := typ.Comparable()
			var exported, embedded []string
			for field := range typ.NumField() {
				f := typ.Field(field)
				if f.IsExported() {
					exported = append(exported, f.Name)
				}
				if f.Anonymous {
					embedded = append(embedded, f.Name)
				}
			}

			// Assert.
			assert.False(t, decodes, "decoding belongs on the mutable Field adapter")
			assert.True(t, isComparable)
			assert.Empty(t, exported)
			assert.Empty(t, embedded)
			assert.Equal(t,
				typ.NumMethod(),
				pointer.NumMethod(),
				"core types must not acquire pointer-only mutators")
		})
	}
}

func TestCodeGroupsAreIndependent(t *testing.T) {
	// Arrange.
	input := xobis.Groups{A: 1, B: 2, C: 3, D: 4, E: 5, F: 6}
	want := input
	code, err := xobis.NewCode(input)
	require.NoError(t, err)
	keys := map[xobis.Code]string{code: "original"}

	// Act.
	input.A = 255
	clone := code.Groups()
	clone.C = 42
	changed, changeErr := xobis.NewCode(clone)
	groups := code.Groups()
	lookup := keys[code]

	// Assert.
	assert.Equal(t, want, groups)
	assert.Equal(t, "original", lookup)
	require.NoError(t, changeErr)
	assert.NotEqual(t, code, changed)
	assert.Equal(t, xobis.Quantity(42), changed.Groups().C)
}

func TestPatternGroupsAreIndependent(t *testing.T) {
	// Arrange.
	input := xobis.Groups{A: 255, B: 64, C: 1, D: 8, E: 255, F: 255}
	mask := xobis.PresentC | xobis.PresentD
	pattern, err := xobis.NewPattern(input, mask)
	require.NoError(t, err)
	keys := map[xobis.Pattern]string{pattern: "original"}
	candidate := newTestCode(t, xobis.Groups{A: 1, C: 1, D: 8, F: 255})

	// Act.
	input.C = 99
	mask = mask.SetPresentA()
	clone, presence := pattern.Groups()
	clone.A = 7
	presence = presence.SetPresentA()
	changed, changeErr := xobis.NewPattern(clone, presence)
	groups, gotPresence := pattern.Groups()
	matched := pattern.Match(candidate)
	lookup := keys[pattern]

	// Assert.
	assert.Equal(t, xobis.Groups{C: 1, D: 8}, groups)
	assert.Equal(t, xobis.PresentA|xobis.PresentC|xobis.PresentD, mask)
	assert.Equal(t, xobis.PresentC|xobis.PresentD, gotPresence)
	assert.True(t, matched)
	assert.Equal(t, "original", lookup)
	require.NoError(t, changeErr)
	assert.NotEqual(t, pattern, changed)
}

func TestPatternCompleteCode(t *testing.T) {
	// Arrange.
	for _, input := range []string{"", "1.8", "1.8.0", "1-0:1.8.0*255", "0-0:0.0.0*0"} {
		t.Run(input, func(t *testing.T) {
			// Arrange.
			var pattern xobis.Pattern
			if input != "" {
				var err error
				pattern, err = xobis.ParsePattern(input)
				require.NoError(t, err)
			}
			want, wantErr := xobis.Parse(input)

			// Act.
			code, err := pattern.CompleteCode()

			// Assert.
			if wantErr != nil {
				assert.ErrorIs(t, err, xobis.ErrSyntax)
				assert.Zero(t, code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, want, code)
			}
		})
	}
}

func TestZeroCoreValues(t *testing.T) {
	// Arrange.
	var code xobis.Code
	var pattern xobis.Pattern

	// Act.
	codeErr := code.Validate()
	codeText, codeMarshalErr := code.MarshalText()
	patternErr := pattern.Validate()
	patternText, patternMarshalErr := pattern.MarshalText()
	matched := pattern.Match(code)
	formatted := pattern.String()

	// Assert.
	assert.NoError(t, codeErr)
	require.NoError(t, codeMarshalErr)
	assert.Equal(t, "0-0:0.0.0*0", string(codeText))
	assert.ErrorIs(t, patternErr, xobis.ErrSyntax)
	assert.ErrorIs(t, patternMarshalErr, xobis.ErrSyntax)
	assert.Nil(t, patternText)
	assert.False(t, matched)
	assert.Empty(t, formatted)
}
