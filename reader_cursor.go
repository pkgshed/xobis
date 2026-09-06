package xobis

import (
	"bufio"
	"io"
)

// ReaderCursor adapts an [io.Reader] to [Cursor] using a fixed-size read buffer
// and one byte of lookahead. Construct it with [NewReaderCursor]; its zero value
// is not usable. It must not be copied or used concurrently.
//
// The buffer may read ahead from the underlying reader. Continue consuming
// input through this cursor, not through the underlying reader. [ReaderCursor]
// does not close the reader, retain consumed input, or support rollback.
type ReaderCursor struct {
	reader  *bufio.Reader
	offset  int64
	peeked  bool
	next    byte
	nextErr error
}

// NewReaderCursor returns a cursor at logical offset zero. The reader must be
// non-nil. Deadlines, cancellation and input limits are the reader's responsibility.
func NewReaderCursor(r io.Reader) *ReaderCursor {
	return &ReaderCursor{reader: bufio.NewReader(r)}
}

// PeekByte returns the next byte without advancing the logical offset.
// Repeated peeks return the same byte or error until [ReaderCursor.ReadByte] is called.
func (c *ReaderCursor) PeekByte() (byte, error) {
	if !c.peeked {
		c.next, c.nextErr = c.reader.ReadByte()
		c.peeked = true
	}
	return c.next, c.nextErr
}

// ReadByte consumes the next byte, advancing [ReaderCursor.Offset] only on success.
// After a failed read, a subsequent operation may retry the underlying reader.
func (c *ReaderCursor) ReadByte() (byte, error) {
	b, err := c.PeekByte()
	c.peeked = false
	if err == nil {
		c.offset++
	}
	return b, err
}

// Offset returns the number of bytes successfully consumed through this cursor.
func (c *ReaderCursor) Offset() int64 { return c.offset }

var _ Cursor = (*ReaderCursor)(nil)
