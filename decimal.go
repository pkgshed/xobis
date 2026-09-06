package xobis

import "fmt"

// Decimal is an immutable, exact decimal representation. Construct it with
// [ParseDecimal]. Equality compares spelling, so 123 and 123.00 are distinct.
// Its zero value is invalid. [Decimal] does not provide arithmetic or rounding.
type Decimal struct {
	text string
}

// ParseDecimal accepts [+-]?[0-9]+(\.[0-9]+)? with ASCII digits. It preserves
// the sign, leading zeroes, fractional scale and negative zero. There is no
// fixed-width numeric limit. Whitespace, exponents and decimal commas are rejected.
// On error, it returns the zero [Decimal] and a *[ParseError] wrapping [ErrSyntax].
func ParseDecimal(s string) (Decimal, error) {
	p := stringParser(s)
	value := p.decimal()
	p.end()
	if p.err != nil {
		return Decimal{}, p.stringError(s)
	}
	return value, nil
}

// String returns the original decimal spelling, or "" for the zero [Decimal].
func (d Decimal) String() string { return d.text }

// Validate rejects the zero [Decimal] with an error wrapping [ErrSyntax].
// All other values produced by the public API are valid by construction.
func (d Decimal) Validate() error {
	if d.text == "" {
		return fmt.Errorf("xobis: a decimal value is required: %w", ErrSyntax)
	}
	return nil
}

// decimal consumes one signed fixed-point value, preserving its exact spelling.
func (p *parser) decimal() Decimal {
	if p.err != nil {
		return Decimal{}
	}
	start := p.i
	p.capturing = p.cursor != nil
	if !p.consume('+') {
		p.consume('-')
	}
	p.digits()
	if p.consume('.') {
		p.digits()
	}
	text := p.captured(start)
	if p.err != nil {
		return Decimal{}
	}
	return Decimal{text: text}
}
