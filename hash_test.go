package xobis_test

import (
	"fmt"
	"hash/maphash"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

func TestCodeHasher(t *testing.T) {
	// Arrange.
	parsed, err := xobis.Parse("01.0.1.8.0.0255")
	require.NoError(t, err)
	code := newTestCode(t, xobis.Groups{
		A: xobis.MediumElectricity,
		C: xobis.ElectricityActivePowerImport,
		D: xobis.ElectricityTimeIntegral1,
		F: xobis.StorageNotUsed,
	})
	otherGroups := code.Groups()
	otherGroups.E = xobis.ElectricityTariff1
	otherTariff := newTestCode(t, otherGroups)
	tests := []struct {
		name      string
		x, y      xobis.Code
		wantEqual bool
	}{
		{"parsed and constructed", parsed, code, true},
		{"zero values", xobis.Code{}, xobis.Code{}, true},
		{"different tariffs", code, otherTariff, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var hasher xobis.CodeHasher
			seed := maphash.MakeSeed()
			var xHash, yHash maphash.Hash
			xHash.SetSeed(seed)
			yHash.SetSeed(seed)

			// Act.
			equal := hasher.Equal(tt.x, tt.y)
			hasher.Hash(&xHash, tt.x)
			hasher.Hash(&yHash, tt.y)
			xSum, ySum := xHash.Sum64(), yHash.Sum64()

			// Assert.
			assert.Equal(t, tt.wantEqual, equal)
			if tt.wantEqual {
				assert.Equal(t, xSum, ySum)
			}
			// Unequal values may legitimately have colliding hashes.
		})
	}
}

func TestPatternHasher(t *testing.T) {
	// Arrange.
	parsed, err := xobis.ParsePattern("1.8.0")
	require.NoError(t, err)
	constructed, err := xobis.NewPattern(
		xobis.Groups{A: 7, B: 64, C: 1, D: 8, E: 0, F: 255},
		xobis.PresentC|xobis.PresentD|xobis.PresentE,
	)
	require.NoError(t, err)
	omittedTariff, err := xobis.ParsePattern("1.8")
	require.NoError(t, err)
	modified, mask := parsed.Groups()
	modified.A = xobis.MediumGas
	reconstructed, err := xobis.NewPattern(modified, mask)
	require.NoError(t, err)
	tests := []struct {
		name      string
		x, y      xobis.Pattern
		wantEqual bool
	}{
		{"normalized construction", parsed, constructed, true},
		{"zero values", xobis.Pattern{}, xobis.Pattern{}, true},
		{"omitted versus explicit zero", omittedTariff, parsed, false},
		{"edited omitted group is normalized", reconstructed, parsed, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var hasher xobis.PatternHasher
			seed := maphash.MakeSeed()
			var xHash, yHash maphash.Hash
			xHash.SetSeed(seed)
			yHash.SetSeed(seed)

			// Act.
			equal := hasher.Equal(tt.x, tt.y)
			hasher.Hash(&xHash, tt.x)
			hasher.Hash(&yHash, tt.y)
			xSum, ySum := xHash.Sum64(), yHash.Sum64()

			// Assert.
			assert.Equal(t, tt.wantEqual, equal)
			if tt.wantEqual {
				assert.Equal(t, xSum, ySum)
			}
			// Unequal values may legitimately have colliding hashes.
		})
	}
}

func ExamplePatternHasher() {
	// Arrange.
	parsed, err := xobis.ParsePattern("1.8.0")
	if err != nil {
		panic(err)
	}
	constructed, err := xobis.NewPattern(
		xobis.Groups{C: xobis.ElectricityActivePowerImport, D: xobis.ElectricityTimeIntegral1},
		xobis.PresentC|xobis.PresentD|xobis.PresentE,
	)
	if err != nil {
		panic(err)
	}
	var hasher xobis.PatternHasher
	seed := maphash.MakeSeed()
	var parsedHash, constructedHash maphash.Hash
	parsedHash.SetSeed(seed)
	constructedHash.SetSeed(seed)

	// Act.
	hasher.Hash(&parsedHash, parsed)
	hasher.Hash(&constructedHash, constructed)
	fmt.Println(hasher.Equal(parsed, constructed))
	fmt.Println(parsedHash.Sum64() == constructedHash.Sum64())

	// Output:
	// true
	// true
}

func TestCodeHasherEncoding(t *testing.T) {
	// Arrange.
	tests := []struct {
		name    string
		code    xobis.Code
		encoded uint64
	}{
		{"zero", xobis.Code{}, 0},
		{"distinct groups", newTestCode(t, xobis.Groups{A: 1, B: 2, C: 3, D: 4, E: 5, F: 6}), 0x060504030201},
		{"maximum valid groups", newTestCode(t, xobis.Groups{A: 15, B: 255, C: 255, D: 255, E: 255, F: 255}), 0xffffffffff0f},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertHashAppends(t, xobis.CodeHasher{}, tt.code, tt.encoded)
		})
	}
}

func TestPatternHasherEncoding(t *testing.T) {
	// Arrange.
	full, err := xobis.NewPattern(xobis.Groups{A: 1, B: 2, C: 3, D: 4, E: 5, F: 6},
		xobis.PresentA|xobis.PresentB|xobis.PresentC|xobis.PresentD|xobis.PresentE|xobis.PresentF)
	require.NoError(t, err)
	omitted, err := xobis.ParsePattern("1.8")
	require.NoError(t, err)
	explicitZero, err := xobis.ParsePattern("1.8.0")
	require.NoError(t, err)
	modified, err := xobis.NewPattern(xobis.Groups{A: 255, B: 254, C: 253, D: 252, E: 251, F: 250}, xobis.PresentC|xobis.PresentD)
	require.NoError(t, err)
	tests := []struct {
		name    string
		pattern xobis.Pattern
		encoded uint64
	}{
		{"zero", xobis.Pattern{}, 0},
		{"all groups", full, 0x3f060504030201},
		{"omitted groups", omitted, 0x0c000008010000},
		{"explicit zero", explicitZero, 0x1c000008010000},
		{"omitted fields are normalized", modified, 0x0c0000fcfd0000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertHashAppends(t, xobis.PatternHasher{}, tt.pattern, tt.encoded)
		})
	}
}

// Check the encoding in a stream to detect resets, reseeding, omitted fields,
// and incorrect composition with other values. No fixed hash numbers or
// collision-freedom assumptions are needed.
func assertHashAppends[T any](t *testing.T, hasher interface{ Hash(*maphash.Hash, T) }, value T, encoded uint64) {
	t.Helper()
	for _, length := range []int{0, 1, 127, 128, 129, 257} {
		t.Run(fmt.Sprintf("prefix=%d", length), func(t *testing.T) {
			// Arrange.
			seed := maphash.MakeSeed()
			prefix := strings.Repeat("p", length)
			const suffix = "tail"
			var gotHash, wantHash maphash.Hash
			gotHash.SetSeed(seed)
			wantHash.SetSeed(seed)
			_, err := gotHash.WriteString(prefix)
			require.NoError(t, err)
			_, err = wantHash.WriteString(prefix)
			require.NoError(t, err)
			maphash.WriteComparable(&wantHash, encoded)
			maphash.WriteComparable(&wantHash, encoded)
			_, err = wantHash.WriteString(suffix)
			require.NoError(t, err)
			want := wantHash.Sum64()

			// Act.
			hasher.Hash(&gotHash, value)
			hasher.Hash(&gotHash, value)
			_, err = gotHash.WriteString(suffix)
			got := gotHash.Sum64()

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, seed, gotHash.Seed())
			assert.Equal(t, want, got)
		})
	}
}

func ExampleCodeHasher() {
	// Arrange.
	code, err := xobis.NewCode(xobis.Groups{A: xobis.MediumElectricity, C: xobis.ElectricityActivePowerImport,
		D: xobis.ElectricityTimeIntegral1, F: xobis.StorageNotUsed})
	if err != nil {
		panic(err)
	}
	var hasher xobis.CodeHasher
	var hash maphash.Hash

	// Act.
	hasher.Hash(&hash, code)
	first := hash.Sum64()
	hash.Reset() // Retains the seed.
	hasher.Hash(&hash, code)
	fmt.Println(first == hash.Sum64())
	fmt.Println(hasher.Equal(code, code))

	// Output:
	// true
	// true
}
