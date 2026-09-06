//go:build go1.27

package xobis_test

import (
	"hash/maphash"

	"github.com/pkgshed/xobis"
)

var (
	_ maphash.Hasher[xobis.Code]    = maphash.ComparableHasher[xobis.Code]{}
	_ maphash.Hasher[xobis.Pattern] = maphash.ComparableHasher[xobis.Pattern]{}
	_ maphash.Hasher[xobis.Code]    = xobis.CodeHasher{}
	_ maphash.Hasher[xobis.Pattern] = xobis.PatternHasher{}
)
