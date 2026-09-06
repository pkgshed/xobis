package xobis

import (
	"fmt"
	"io"
)

// Cursor provides byte-oriented access to a parser's logical input.
// [Cursor.Offset] is the nonnegative, zero-based position of the next byte.
// [Cursor.ReadByte] advances it by one on success; [Cursor.PeekByte] and failed
// reads leave it unchanged.
// A successful peek exposes the byte returned by the next successful read.
// Operations may block. [io.EOF] means the input has ended, not that another
// chunk is temporarily unavailable.
//
// Implementations must be non-nil and must not be used concurrently or advanced
// independently during parsing. Parsers may consume input before failing;
// checkpointing, rollback, resource limits and framing belong to the caller.
type Cursor interface {
	PeekByte() (byte, error)
	ReadByte() (byte, error)
	Offset() int64
}

// CursorError describes a parsing or I/O failure at an absolute byte offset in
// a [Cursor]. [CursorError.Offset] can precede the cursor's current position.
// No input is retained.
type CursorError struct {
	Offset int64
	Err    error
}

func (e *CursorError) Error() string {
	return fmt.Sprintf("xobis: parse at byte %d: %v", e.Offset, e.Err)
}

// Unwrap exposes [ErrSyntax], [ErrRange], or the underlying I/O error.
func (e *CursorError) Unwrap() error { return e.Err }

// ParsePatternFrom consumes one pattern at the cursor's current offset. It
// leaves the first unrelated byte untouched: "1.8xyz" consumes only "1.8".
// OBIS punctuation continues the grammar, so "1.8." fails instead of returning
// a shorter pattern. Completion requires a following delimiter or [io.EOF] and may
// block until one is available. The caller decides which following bytes are legal.
//
// Failure returns the zero [Pattern] and a *[CursorError] without rolling back
// consumed bytes. [io.EOF] where input is required becomes [ErrSyntax]; other I/O
// errors retain their identity through wrapping. A negative initial offset
// returns an error wrapping [ErrRange], rather than a [CursorError], without reading.
// Use [ParsePattern] to require complete string input.
func ParsePatternFrom(cursor Cursor) (Pattern, error) {
	p, err := parserFromCursor(cursor)
	if err != nil {
		return Pattern{}, err
	}
	pattern := p.pattern()
	if p.err != nil {
		return Pattern{}, p.cursorError()
	}
	return pattern, nil
}

// ParseReadingFrom consumes one reading at the cursor's current offset,
// stopping immediately after ')' without inspecting the following byte.
// Whitespace, line endings and subsequent input belong to the caller.
// Use [ParseReading] to require a complete string ending immediately after ')'.
// Failure and error-position rules are the same as [ParsePatternFrom].
func ParseReadingFrom(cursor Cursor) (Reading, error) {
	p, err := parserFromCursor(cursor)
	if err != nil {
		return Reading{}, err
	}
	reading := p.reading()
	if p.err != nil {
		return Reading{}, p.cursorError()
	}
	return reading, nil
}

func parserFromCursor(cursor Cursor) (parser, error) {
	offset := cursor.Offset()
	if offset < 0 {
		return parser{}, fmt.Errorf("xobis: negative cursor offset %d: %w", offset, ErrRange)
	}
	return parser{cursor: cursor, i: offset}, nil
}

// stringCursor adapts complete strings without allocating a reader buffer.
type stringCursor struct {
	s string
	i int
}

func (c *stringCursor) PeekByte() (byte, error) {
	if c.i == len(c.s) {
		return 0, io.EOF
	}
	return c.s[c.i], nil
}

func (c *stringCursor) ReadByte() (byte, error) {
	b, err := c.PeekByte()
	if err == nil {
		c.i++
	}
	return b, err
}

func (c *stringCursor) Offset() int64 { return int64(c.i) }
