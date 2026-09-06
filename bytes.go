package xobis

import "fmt"

// FromBytes constructs a [Code] from exactly six raw bytes in A through F order.
// A must be in 0..15; B through F may each be in 0..255. Zero bytes remain
// explicit groups. The input is neither modified nor retained.
// On error, [FromBytes] returns the zero [Code] and an error wrapping [ErrSyntax]
// for an incorrect length or [ErrRange] for an out-of-range group A.
func FromBytes(value []byte) (Code, error) {
	if len(value) != 6 {
		return Code{}, fmt.Errorf("xobis: expected exactly 6 bytes, got %d: %w", len(value), ErrSyntax)
	}
	return NewCode(Groups{
		A: Medium(value[0]),
		B: Channel(value[1]),
		C: Quantity(value[2]),
		D: Processing(value[3]),
		E: Classification(value[4]),
		F: Storage(value[5]),
	})
}

// ToBytes returns c's six raw bytes in A through F order.
// Zero bytes remain explicit groups. The returned array is an independent copy.
func ToBytes(c Code) [6]byte {
	g := c.groups
	return [6]byte{byte(g.A), byte(g.B), byte(g.C), byte(g.D), byte(g.E), byte(g.F)}
}
