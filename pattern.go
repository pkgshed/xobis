package xobis

import "fmt"

// Pattern is an immutable OBIS identifier with optional groups A, B, E and F.
// Construct one with [NewPattern] or [ParsePattern]. Omitted groups match any value;
// explicit zeroes match only zero. The zero [Pattern] is invalid and matches no codes.
// [Pattern] is comparable and suitable as a map key. [Pattern.Groups] returns a copy
// together with presence flags. Use [PatternField] for mutable text decoding.
type Pattern struct {
	groups  Groups
	present Presence
}

// NewPattern constructs a pattern from the groups selected by present.
// Groups C and D must be selected, and present must contain only defined flags.
// Selected groups must be in range. Omitted groups are ignored and their fields
// are set to zero, so equivalent inputs produce equal patterns.
//
// On error, [NewPattern] returns the zero [Pattern] and an error wrapping [ErrSyntax]
// for invalid presence flags or [ErrRange] for an out-of-range selected group.
func NewPattern(groups Groups, present Presence) (Pattern, error) {
	p := Pattern{groups: groups, present: present}
	if err := p.Validate(); err != nil {
		return Pattern{}, err
	}
	if !present.PresentA() {
		p.groups.A = 0
	}
	if !present.PresentB() {
		p.groups.B = 0
	}
	if !present.PresentE() {
		p.groups.E = 0
	}
	if !present.PresentF() {
		p.groups.F = 0
	}
	return p, nil
}

// Groups returns independent copies of the groups and their presence flags.
// Omitted groups are zero; explicit zeroes are distinguished by presence.
// Edit the copies and call [NewPattern] to construct a different pattern.
func (p Pattern) Groups() (Groups, Presence) { return p.groups, p.present }

// Presence reports which value groups were supplied.
func (p Pattern) Presence() Presence { return p.present }

// Validate checks that presence contains only defined flags, that C and D are
// present, and that present groups are in range. Values in omitted groups are ignored.
// It does not check standardized measurement meanings.
func (p Pattern) Validate() error {
	if unknown := p.present &^ allPresent; unknown != 0 {
		return fmt.Errorf("xobis: unknown presence bits %#x: %w", uint8(unknown), ErrSyntax)
	}
	if p.present&(PresentC|PresentD) != PresentC|PresentD {
		return fmt.Errorf("xobis: groups C and D are required: %w", ErrSyntax)
	}
	if p.present.PresentA() {
		return p.groups.Validate()
	}
	return nil
}

// CompleteCode returns a [Code] only when all six groups are present.
// Otherwise it returns the zero [Code] and an error wrapping [ErrSyntax].
func (p Pattern) CompleteCode() (Code, error) {
	if p.present != allPresent {
		return Code{}, fmt.Errorf("xobis: all six groups are required: %w", ErrSyntax)
	}
	return Code{groups: p.groups}, nil
}

// Match reports whether p is valid and c agrees with every present group of p.
func (p Pattern) Match(c Code) bool {
	if p.Validate() != nil {
		return false
	}
	return (!p.present.PresentA() || p.groups.A == c.groups.A) &&
		(!p.present.PresentB() || p.groups.B == c.groups.B) &&
		(!p.present.PresentC() || p.groups.C == c.groups.C) &&
		(!p.present.PresentD() || p.groups.D == c.groups.D) &&
		(!p.present.PresentE() || p.groups.E == c.groups.E) &&
		(!p.present.PresentF() || p.groups.F == c.groups.F)
}

// String formats present groups using OBIS separators without leading zeroes.
// The zero [Pattern] formats as an empty string.
func (p Pattern) String() string {
	return string(appendCode(make([]byte, 0, 23), p.groups, p.present))
}

// MarshalText implements [encoding.TextMarshaler], validating the pattern.
func (p Pattern) MarshalText() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return appendCode(make([]byte, 0, 23), p.groups, p.present), nil
}
