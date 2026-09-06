package xobis

import (
	"io"
	"strings"
)

// parser shares one grammar between strings and streaming cursors. The embedded
// string adapter avoids interface allocation for the string entry points.
// Errors are sticky; helpers stop inspecting and consuming input after failure.
type parser struct {
	cursor    Cursor
	text      stringCursor
	i         int64
	err       error
	errOffset int64
	peeked    bool
	next      byte
	eof       bool
	capturing bool
	buf       strings.Builder
}

func stringParser(s string) parser { return parser{text: stringCursor{s: s}} }

func (p *parser) fail(offset int64, err error) {
	if p.err == nil {
		p.err, p.errOffset = err, offset
	}
}

func (p *parser) cursorError() error { return &CursorError{Offset: p.errOffset, Err: p.err} }

func (p *parser) stringError(s string) error {
	return &ParseError{Input: s, Offset: int(p.errOffset), Err: p.err}
}

// peek distinguishes end of input from transport failure and caches lookahead.
func (p *parser) peek() (byte, bool) {
	if p.cursor == nil {
		return p.peekString()
	}
	return p.peekCursor()
}

func (p *parser) peekString() (byte, bool) {
	if p.err != nil || p.text.i == len(p.text.s) {
		return 0, false
	}
	return p.text.s[p.text.i], true
}

func (p *parser) peekCursor() (byte, bool) {
	if p.err != nil || p.eof {
		return 0, false
	}
	if !p.peeked {
		var err error
		p.next, err = p.cursor.PeekByte()
		if err == io.EOF {
			p.eof = true
			return 0, false
		}
		if err != nil {
			p.fail(p.i, err)
			return 0, false
		}
		p.peeked = true
	}
	return p.next, true
}

// take consumes one byte. Call only after a successful peek.
func (p *parser) take() {
	if p.cursor == nil {
		p.text.i++
		p.i++
		return
	}
	p.takeCursor()
}

func (p *parser) takeCursor() {
	if p.err != nil {
		return
	}
	b, err := p.cursor.ReadByte()
	if err != nil {
		if err == io.EOF {
			err = ErrSyntax
		}
		p.fail(p.i, err)
		return
	}
	p.peeked = false
	p.i++
	if p.capturing {
		p.buf.WriteByte(b)
	}
}

func (p *parser) at(b byte) bool {
	next, ok := p.peek()
	return ok && next == b
}

func (p *parser) consume(b byte) bool {
	if !p.at(b) {
		return false
	}
	p.take()
	return p.err == nil
}

func (p *parser) expect(b byte) {
	if !p.consume(b) {
		p.fail(p.i, ErrSyntax)
	}
}

func (p *parser) end() {
	if _, ok := p.peek(); ok {
		p.fail(p.i, ErrSyntax)
	}
}

func (p *parser) number() uint8 {
	if p.err != nil {
		return 0
	}
	start := p.i
	var n uint16
	for {
		b, ok := p.peek()
		if !ok || b < '0' || b > '9' {
			break
		}
		n = n*10 + uint16(b-'0')
		if n > 255 {
			p.fail(start, ErrRange)
			return 0
		}
		p.take()
	}
	if p.i == start {
		p.fail(start, ErrSyntax)
	}
	return uint8(n)
}

// digits consumes one or more ASCII digits without interpreting their value.
func (p *parser) digits() {
	if p.err != nil {
		return
	}
	start := p.i
	for {
		b, ok := p.peek()
		if !ok || b < '0' || b > '9' {
			break
		}
		p.take()
	}
	if p.i == start {
		p.fail(start, ErrSyntax)
	}
}

// captured returns immutable token text, borrowing only from immutable strings.
func (p *parser) captured(start int64) string {
	p.capturing = false
	if p.cursor == nil {
		return p.text.s[start:p.i]
	}
	text := p.buf.String()
	p.buf.Reset()
	return text
}
