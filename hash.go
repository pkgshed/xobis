package xobis

import "hash/maphash"

// CodeHasher implements [maphash.Hasher] for [Code] using its == equality.
// It is stateless and its zero value is ready to use.
type CodeHasher struct{}

// Hash packs c's groups A through F into the low 48 bits of a uint64, with A in
// the lowest byte, and adds it to h using [maphash.WriteComparable]. It preserves
// h's existing contents and seed. All [Code] values can be hashed, including zero.
func (CodeHasher) Hash(h *maphash.Hash, c Code) {
	h.Seed() // Initialize a zero [maphash.Hash] before writing comparable values.
	maphash.WriteComparable(h, packGroups(c.groups))
}

// Equal reports whether x == y.
func (CodeHasher) Equal(x, y Code) bool { return x == y }

// PatternHasher implements [maphash.Hasher] for [Pattern] using its == equality.
// It is stateless and its zero value is ready to use. Equality includes presence
// flags and all groups. Omitted groups are always stored as zero.
type PatternHasher struct{}

// Hash packs p's groups A through F into the low 48 bits of a uint64, with A in
// the lowest byte, and the presence mask into bits 48 through 55. It adds that
// value to h using [maphash.WriteComparable], preserving h's contents and seed.
// It does not validate or normalize p, and it accepts the zero [Pattern].
func (PatternHasher) Hash(h *maphash.Hash, p Pattern) {
	h.Seed() // Initialize a zero [maphash.Hash] before writing comparable values.
	maphash.WriteComparable(h, packGroups(p.groups)|uint64(p.present)<<48)
}

// Equal reports whether x == y. It does not perform pattern matching.
func (PatternHasher) Equal(x, y Pattern) bool { return x == y }

func packGroups(c Groups) uint64 {
	return uint64(c.A) | uint64(c.B)<<8 | uint64(c.C)<<16 |
		uint64(c.D)<<24 | uint64(c.E)<<32 | uint64(c.F)<<40
}
