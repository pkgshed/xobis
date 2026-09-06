package xobis

import (
	"errors"
	"fmt"
)

var (
	// ErrSyntax indicates malformed input, invalid presence flags or missing required groups.
	ErrSyntax = errors.New("invalid OBIS syntax")
	// ErrRange indicates a value outside its group's range, a packed code
	// exceeding 48 bits, or a negative initial cursor offset.
	ErrRange = errors.New("OBIS value out of range")
)

const allPresent = PresentA | PresentB | PresentC | PresentD | PresentE | PresentF

// ParseError describes a parsing failure. [ParseError.Offset] is a zero-based byte
// offset in [ParseError.Input]; an offset equal to the input's length indicates
// an error at the end of the input.
type ParseError struct {
	Input  string
	Offset int
	Err    error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("xobis: parse %q at byte %d: %v", e.Input, e.Offset, e.Err)
}

// Unwrap supports [errors.Is] with [ErrSyntax] and [ErrRange].
func (e *ParseError) Unwrap() error { return e.Err }

// Parse parses a complete numeric OBIS identifier in A-B:C.D.E*F or
// A.B.C.D.E.F notation. All six groups are required. Values must be decimal,
// with A in 0..15 and B through F in 0..255. Leading zeroes are accepted;
// whitespace, signs, hexadecimal values and trailing text are rejected.
// Use [ParseHex] for hexadecimal notation.
// On error, [Parse] returns the zero [Code] and a *[ParseError].
func Parse(s string) (Code, error) {
	p, err := ParsePattern(s)
	if err != nil {
		return Code{}, err
	}
	if p.present != allPresent {
		return Code{}, &ParseError{s, len(s), fmt.Errorf("all six groups are required: %w", ErrSyntax)}
	}
	return Code{groups: p.groups}, nil
}

// ParsePattern parses a numeric OBIS identifier with optional groups A, B, E
// and F: [A-][B:]C.D[.E][*F]. It also accepts complete A.B.C.D.E.F notation.
// The same numeric and lexical rules as [Parse] apply. Omitted groups are zero
// in the [Groups] copy returned by [Pattern.Groups] and absent from [Pattern.Presence].
// No defaults are assigned.
// On error, [ParsePattern] returns the zero [Pattern] and a *[ParseError].
func ParsePattern(s string) (Pattern, error) {
	r := stringParser(s)
	pattern := r.pattern()
	r.end()
	if r.err != nil {
		return Pattern{}, r.stringError(s)
	}
	return pattern, nil
}

// pattern consumes an identifier, leaving its first non-OBIS byte untouched.
// Dotted notation is resolved from local groups, never from the source suffix.
func (p *parser) pattern() Pattern {
	if p.err != nil {
		return Pattern{}
	}
	start := p.i
	var groups Groups
	present := PresentC | PresentD
	v := p.number()
	if p.consume('-') {
		groups.A, present = Medium(v), present|PresentA
		v = p.number()
	}
	if p.consume(':') {
		groups.B, present = Channel(v), present|PresentB
		v = p.number()
	}
	if present&(PresentA|PresentB) != 0 {
		groups.C = Quantity(v)
		p.expect('.')
		groups.D = Processing(p.number())
		if p.consume('.') {
			groups.E, present = Classification(p.number()), present|PresentE
		}
	} else {
		var buf [6]uint8
		dotted := buf[:1]
		dotted[0] = v

		for p.at('.') {
			if len(dotted) == cap(dotted) {
				p.fail(p.i, ErrSyntax)
				break
			}
			p.consume('.')
			dotted = append(dotted, p.number())
		}

		switch len(dotted) {
		case 2, 3:
			groups.C, groups.D = Quantity(dotted[0]), Processing(dotted[1])
			if len(dotted) == 3 {
				groups.E, present = Classification(dotted[2]), present|PresentE
			}
		case 6:
			groups = Groups{
				A: Medium(dotted[0]),
				B: Channel(dotted[1]),
				C: Quantity(dotted[2]),
				D: Processing(dotted[3]),
				E: Classification(dotted[4]),
				F: Storage(dotted[5]),
			}
			present = allPresent
		default:
			p.fail(p.i, ErrSyntax)
		}
	}
	if !present.PresentF() && p.consume('*') {
		groups.F, present = Storage(p.number()), present|PresentF
	}
	// OBIS punctuation cannot act as an outer delimiter. Reject malformed
	// continuations instead of silently returning a shorter valid identifier.
	if p.at('.') || p.at('-') || p.at(':') || p.at('*') {
		p.fail(p.i, ErrSyntax)
	}
	if p.err != nil {
		return Pattern{}
	}
	pattern, err := NewPattern(groups, present)
	if err != nil {
		p.fail(start, err)
		return Pattern{}
	}
	return pattern
}
