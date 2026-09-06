package xobis

import (
	"encoding/hex"
	"fmt"
)

// ParseHex parses six hexadecimal bytes in A through F order, such as
// "0100010800ff". It accepts exactly 12 hexadecimal digits, optionally prefixed
// by "0x" or "0X". Digits are case-insensitive; whitespace, signs, separators,
// and underscores are rejected. All six groups are present, including zeroes.
// Numeric ranges are the same as for [Parse].
// On error, [ParseHex] returns the zero [Code] and a *[ParseError].
func ParseHex(s string) (Code, error) {
	start := 0
	if len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		start = 2
	}
	if len(s)-start != 12 {
		return Code{}, &ParseError{s, start + min(len(s)-start, 12),
			fmt.Errorf("expected exactly 12 hexadecimal digits: %w", ErrSyntax)}
	}
	var value uint64
	for i := start; i < len(s); i++ {
		var digit byte
		switch ch := s[i]; {
		case ch >= '0' && ch <= '9':
			digit = ch - '0'
		case ch >= 'a' && ch <= 'f':
			digit = ch - 'a' + 10
		case ch >= 'A' && ch <= 'F':
			digit = ch - 'A' + 10
		default:
			return Code{}, &ParseError{s, i, ErrSyntax}
		}
		value = value<<4 | uint64(digit)
	}
	code, err := FromUint64(value)
	if err != nil {
		return Code{}, &ParseError{s, start, err}
	}
	return code, nil
}

// ToHex returns exactly 12 lowercase hexadecimal digits in
// A through F order, including leading zeroes and without a prefix.
// For example, 1-0:1.8.0*255 becomes "0100010800ff".
func ToHex(c Code) string {
	value := ToBytes(c)
	return hex.EncodeToString(value[:])
}

// FromUint64 constructs a [Code] from six bytes packed in A through F order into
// the low 48 bits of value. A occupies bits 40 through 47 and F the lowest byte;
// for example, 0x0100010800ff represents 1-0:1.8.0*255. Leading zero bytes are
// retained as explicit groups. The high 16 bits must be zero, and A must be in
// 0..15. B through F may each be in 0..255.
// On error, [FromUint64] returns the zero [Code] and an error wrapping [ErrRange].
func FromUint64(value uint64) (Code, error) {
	if value>>48 != 0 {
		return Code{}, fmt.Errorf("xobis: packed code %#x exceeds 48 bits: %w", value, ErrRange)
	}
	return NewCode(Groups{
		A: Medium(value >> 40),
		B: Channel(value >> 32),
		C: Quantity(value >> 24),
		D: Processing(value >> 16),
		E: Classification(value >> 8),
		F: Storage(value),
	})
}

// ToUint64 packs c's six bytes in A through F order into the
// low 48 bits of a uint64. A occupies bits 40 through 47 and F the lowest byte;
// the high 16 bits are zero. For example, 1-0:1.8.0*255 becomes 0x0100010800ff.
func ToUint64(c Code) uint64 {
	g := c.groups
	return uint64(g.A)<<40 | uint64(g.B)<<32 | uint64(g.C)<<24 |
		uint64(g.D)<<16 | uint64(g.E)<<8 | uint64(g.F)
}
