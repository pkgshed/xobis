package xobis

// Code is an immutable, complete OBIS identifier with value groups A through F.
// Its zero value is the valid code 0-0:0.0.0*0. Construct other values with
// [NewCode], [Parse], [ParseHex], [FromUint64] or [FromBytes]. [Code] is comparable and
// suitable as a map key. Use [CodeField] for mutable text decoding.
type Code struct {
	groups Groups
}

// NewCode validates and copies groups into an immutable [Code].
// On error, it returns the zero [Code] and an error wrapping [ErrRange].
func NewCode(groups Groups) (Code, error) {
	if err := groups.Validate(); err != nil {
		return Code{}, err
	}
	return Code{groups: groups}, nil
}

// Groups returns an independent copy of all six groups.
// Edit the copy and call [NewCode] to construct a different code.
func (c Code) Groups() Groups { return c.groups }

// Validate checks numeric ranges. Every [Code] produced by the public API,
// including the zero value, is valid. It does not check assigned meanings.
func (c Code) Validate() error { return c.groups.Validate() }

// String returns the decimal A-B:C.D.E*F representation without leading zeroes.
func (c Code) String() string {
	return string(appendCode(make([]byte, 0, 23), c.groups, allPresent))
}

// MarshalText implements [encoding.TextMarshaler] using decimal notation.
// [Code] values are valid by construction, so the error is always nil.
func (c Code) MarshalText() ([]byte, error) {
	return appendCode(make([]byte, 0, 23), c.groups, allPresent), nil
}
