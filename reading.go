package xobis

import "fmt"

// Reading is an immutable numeric meter reading with an OBIS identifier and an
// optional unit. Construct it with [ParseReading] or [NewReading]. It is comparable;
// equality includes identifier presence, decimal spelling and unit spelling.
// Its zero value is invalid. Use [ReadingField] for mutable text decoding.
type Reading struct {
	identifier Pattern
	value      Decimal
	unit       Unit
}

// NewReading validates and constructs a single numeric reading. The identifier
// retains omitted groups without assigning defaults. [UnitNone] omits the unit.
// Unit meaning and compatibility with the identifier are not checked.
// On error, it returns the zero [Reading] and an error wrapping [ErrSyntax] or [ErrRange].
func NewReading(identifier Pattern, value Decimal, unit Unit) (Reading, error) {
	r := Reading{identifier: identifier, value: value, unit: unit}
	if err := r.Validate(); err != nil {
		return Reading{}, err
	}
	return r, nil
}

// Identifier returns the immutable identifier, including its presence flags.
func (r Reading) Identifier() Pattern { return r.identifier }

// Value returns the exact decimal value, preserving its original spelling.
func (r Reading) Value() Decimal { return r.value }

// Unit returns the supplied symbol, or [UnitNone] when absent.
func (r Reading) Unit() Unit { return r.unit }

// Validate checks the identifier, decimal value and unit syntax. It does not
// check measurement meaning, unit compatibility or physical plausibility.
func (r Reading) Validate() error {
	if err := r.identifier.Validate(); err != nil {
		return fmt.Errorf("xobis: reading identifier: %w", err)
	}
	if err := r.value.Validate(); err != nil {
		return fmt.Errorf("xobis: reading value: %w", err)
	}
	if offset := unitErrorOffset(string(r.unit)); offset >= 0 {
		return fmt.Errorf("xobis: invalid unit at byte %d: %w", offset, ErrSyntax)
	}
	return nil
}

// String formats identifier(value[*unit]) without a line ending. The identifier
// uses canonical decimal notation; decimal and unit spelling are preserved.
// The zero [Reading] formats as an empty string.
func (r Reading) String() string {
	if r.value.text == "" {
		return ""
	}
	return string(appendReading(nil, r))
}

// MarshalText implements [encoding.TextMarshaler]. Invalid readings return nil
// bytes and a validation error; valid readings use the same format as [Reading.String].
func (r Reading) MarshalText() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	return appendReading(nil, r), nil
}

func appendReading(dst []byte, r Reading) []byte {
	dst = appendCode(dst, r.identifier.groups, r.identifier.present)
	dst = append(dst, '(')
	dst = append(dst, r.value.text...)
	if r.unit != UnitNone {
		dst = append(dst, '*')
		dst = append(dst, r.unit...)
	}
	return append(dst, ')')
}
