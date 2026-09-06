//go:build go1.27

package xobis_test

import (
	"hash/maphash"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

func BenchmarkCodeHasher(b *testing.B) {
	// Arrange.
	code := newTestCode(b, xobis.Groups{A: 1, C: 1, D: 8, F: 255})

	b.Run("specialized", func(b *testing.B) {
		// Arrange.
		var hasher xobis.CodeHasher
		var hash maphash.Hash
		seed := maphash.MakeSeed()
		hash.SetSeed(seed)
		b.ReportAllocs()

		// Act.
		for b.Loop() {
			hash.Reset()
			hasher.Hash(&hash, code)
			_ = hash.Sum64()
		}

		// Assert outside the measured loop.
		require.Equal(b, seed, hash.Seed())
	})
	b.Run("comparable", func(b *testing.B) {
		// Arrange.
		var hasher maphash.ComparableHasher[xobis.Code]
		var hash maphash.Hash
		seed := maphash.MakeSeed()
		hash.SetSeed(seed)
		b.ReportAllocs()

		// Act.
		for b.Loop() {
			hash.Reset()
			hasher.Hash(&hash, code)
			_ = hash.Sum64()
		}

		// Assert outside the measured loop.
		require.Equal(b, seed, hash.Seed())
	})
}

func BenchmarkPatternHasher(b *testing.B) {
	// Arrange.
	pattern, err := xobis.ParsePattern("1.8.0")
	require.NoError(b, err)

	b.Run("specialized", func(b *testing.B) {
		// Arrange.
		var hasher xobis.PatternHasher
		var hash maphash.Hash
		seed := maphash.MakeSeed()
		hash.SetSeed(seed)
		b.ReportAllocs()

		// Act.
		for b.Loop() {
			hash.Reset()
			hasher.Hash(&hash, pattern)
			_ = hash.Sum64()
		}

		// Assert outside the measured loop.
		require.Equal(b, seed, hash.Seed())
	})
	b.Run("comparable", func(b *testing.B) {
		// Arrange.
		var hasher maphash.ComparableHasher[xobis.Pattern]
		var hash maphash.Hash
		seed := maphash.MakeSeed()
		hash.SetSeed(seed)
		b.ReportAllocs()

		// Act.
		for b.Loop() {
			hash.Reset()
			hasher.Hash(&hash, pattern)
			_ = hash.Sum64()
		}

		// Assert outside the measured loop.
		require.Equal(b, seed, hash.Seed())
	})
}
