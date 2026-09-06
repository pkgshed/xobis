package xobis

import "unicode/utf8"

// ParseReading parses a complete identifier(value[*unit]).
// The identifier uses [ParsePattern]'s grammar; the value uses
// [ParseDecimal]'s grammar. A supplied unit must be nonempty and satisfy [Unit]'s
// syntax. Omitted identifier groups and exact value and unit spelling are retained.
// Surrounding whitespace, line endings, additional groups and trailing text are rejected.
// On error, it returns the zero [Reading] and a *[ParseError] whose [ParseError.Input]
// and byte [ParseError.Offset] refer to the original string, including any line ending.
func ParseReading(s string) (Reading, error) {
	p := stringParser(s)
	reading := p.reading()
	p.end()
	if p.err != nil {
		return Reading{}, p.stringError(s)
	}
	return reading, nil
}

// reading composes grammars without taking ownership of surrounding framing.
func (p *parser) reading() Reading {
	if p.err != nil {
		return Reading{}
	}
	start := p.i
	identifier := p.pattern()
	p.expect('(')
	value := p.decimal()
	unit := UnitNone
	if p.consume('*') {
		unit = p.unit()
	}
	p.expect(')')
	if p.err != nil {
		return Reading{}
	}
	reading, err := NewReading(identifier, value, unit)
	if err != nil {
		p.fail(start, err)
		return Reading{}
	}
	return reading
}

// unit consumes a nonempty symbol, leaving the reading's closing ')' untouched.
func (p *parser) unit() Unit {
	if p.err != nil {
		return UnitNone
	}
	start := p.i
	p.capturing = p.cursor != nil
	for {
		b, ok := p.peek()
		if !ok || b == ')' {
			break
		}
		p.unitRune()
	}
	text := p.captured(start)
	if p.i == start {
		p.fail(start, ErrSyntax)
	}
	if p.err != nil {
		return UnitNone
	}
	return Unit(text)
}

// unitRune validates one UTF-8 rune across arbitrary input chunk boundaries.
// Invalid and truncated encodings point to their first byte; transport errors
// retain the position where reading failed.
func (p *parser) unitRune() {
	start := p.i
	var buf [utf8.UTFMax]byte
	for n := 1; n <= len(buf); n++ {
		b, ok := p.peek()
		if !ok {
			p.fail(start, ErrSyntax)
			return
		}
		buf[n-1] = b
		if utf8.FullRune(buf[:n]) {
			if unitRuneSize(string(buf[:n])) == 0 {
				p.fail(start, ErrSyntax)
				return
			}
			p.take()
			return
		}
		p.take()
	}
}
