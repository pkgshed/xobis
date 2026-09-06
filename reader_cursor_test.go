package xobis_test

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

// chunkReader never supplies more than chunk bytes in one read.
type chunkReader struct {
	input string
	chunk int
}

func (r *chunkReader) Read(p []byte) (int, error) {
	if len(r.input) == 0 {
		return 0, io.EOF
	}
	n := copy(p[:min(len(p), r.chunk)], r.input)
	r.input = r.input[n:]
	return n, nil
}

// offsetCursor adapts a substream to its absolute position in a larger document.
type offsetCursor struct {
	xobis.Cursor
	base int64
}

func (c *offsetCursor) Offset() int64 { return c.base + c.Cursor.Offset() }

func TestReaderCursorLookahead(t *testing.T) {
	// Arrange.
	cursor := xobis.NewReaderCursor(&chunkReader{input: "ab", chunk: 1})

	// Act.
	initial := cursor.Offset()
	peek, peekErr := cursor.PeekByte()
	again, againErr := cursor.PeekByte()
	afterPeek := cursor.Offset()
	first, firstErr := cursor.ReadByte()
	afterFirst := cursor.Offset()
	second, secondErr := cursor.ReadByte()
	_, eofPeek := cursor.PeekByte()
	_, eofAgain := cursor.PeekByte()
	_, eofRead := cursor.ReadByte()
	afterEOF := cursor.Offset()

	// Assert.
	assert.Zero(t, initial)
	require.NoError(t, peekErr)
	require.NoError(t, againErr)
	assert.Equal(t, byte('a'), peek)
	assert.Equal(t, peek, again)
	assert.Zero(t, afterPeek)
	require.NoError(t, firstErr)
	assert.Equal(t, peek, first)
	assert.Equal(t, int64(1), afterFirst)
	require.NoError(t, secondErr)
	assert.Equal(t, byte('b'), second)
	assert.ErrorIs(t, eofPeek, io.EOF)
	assert.ErrorIs(t, eofAgain, io.EOF)
	assert.ErrorIs(t, eofRead, io.EOF)
	assert.Equal(t, int64(2), afterEOF)
}

func TestStreamingReadingsAcrossEverySplit(t *testing.T) {
	// Arrange.
	for _, input := range []string{
		"1.8(0)", "1-0:1.8.0*255(+00123.4500*kWh)",
		"01.000.001.08.000.0255(-0.00*m³/h)", "1.8(1*°C)",
		"1.8(1*e\u0301)", "1.8(1*\U0001f321)", "1.8(1*\ufffd)",
	} {
		for split := 0; split <= len(input); split++ {
			t.Run(fmt.Sprintf("%q/%d", input, split), func(t *testing.T) {
				// Arrange.
				reader := io.MultiReader(strings.NewReader(input[:split]), strings.NewReader(input[split:]), strings.NewReader(";"))
				cursor := xobis.NewReaderCursor(reader)
				want, err := xobis.ParseReading(input)
				require.NoError(t, err)

				// Act.
				got, err := xobis.ParseReadingFrom(cursor)
				next, nextErr := cursor.PeekByte()

				// Assert.
				require.NoError(t, err)
				assert.Equal(t, want, got)
				assert.Equal(t, int64(len(input)), cursor.Offset())
				require.NoError(t, nextErr)
				assert.Equal(t, byte(';'), next)
			})
		}
	}
}

func TestStreamingInvalidUTF8Offsets(t *testing.T) {
	// Arrange.
	for _, unit := range []string{"\xff", "\xc2", "\xe2\x82", "\xf0\x9f\x8c", "\xc0\xaf", "\xed\xa0\x80", "\xf4\x90\x80\x80", "\xe2(", "\xc2 ", "\x80", "\u200b"} {
		for _, suffix := range []string{"", ")"} {
			t.Run(fmt.Sprintf("%q", unit+suffix), func(t *testing.T) {
				// Arrange.
				const prefix = "1.8(1*"
				cursor := xobis.NewReaderCursor(&chunkReader{input: prefix + unit + suffix, chunk: 1})
				var pe *xobis.CursorError

				// Act.
				got, err := xobis.ParseReadingFrom(cursor)

				// Assert.
				assert.Zero(t, got)
				require.ErrorAs(t, err, &pe)
				assert.ErrorIs(t, err, xobis.ErrSyntax)
				assert.Equal(t, int64(len(prefix)), pe.Offset)
			})
		}
	}
}

func TestStreamingReadingOwnsLongText(t *testing.T) {
	// Arrange.
	value := "+" + strings.Repeat("0", 8192) + "1.000"
	unit := strings.Repeat("°", 8192)
	first := "1.8(" + value + "*" + unit + ")"
	second := "2.8(" + strings.Repeat("9", len(first)) + ")"
	cursor := xobis.NewReaderCursor(strings.NewReader(first + second))

	// Act.
	got, firstErr := xobis.ParseReadingFrom(cursor)
	_, secondErr := xobis.ParseReadingFrom(cursor)
	text := got.String()

	// Assert.
	require.NoError(t, firstErr)
	require.NoError(t, secondErr)
	assert.Equal(t, value, got.Value().String())
	assert.Equal(t, unit, string(got.Unit()))
	assert.Equal(t, first, text)
	assert.Equal(t, int64(len(first)+len(second)), cursor.Offset())
}

func TestStreamingLongIdentifier(t *testing.T) {
	// Arrange.
	input := strings.Repeat("0", 16384) + "1.8;"
	cursor := xobis.NewReaderCursor(&chunkReader{input: input, chunk: 7})

	// Act.
	got, err := xobis.ParsePatternFrom(cursor)
	next, nextErr := cursor.PeekByte()

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, "1.8", got.String())
	assert.Equal(t, int64(len(input)-1), cursor.Offset())
	require.NoError(t, nextErr)
	assert.Equal(t, byte(';'), next)
}

func TestReaderCursorDataWithEOF(t *testing.T) {
	// Arrange.
	cursor := xobis.NewReaderCursor(iotest.DataErrReader(strings.NewReader("1.8")))

	// Act.
	got, err := xobis.ParsePatternFrom(cursor)
	_, eof := cursor.ReadByte()

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, "1.8", got.String())
	assert.Equal(t, int64(3), cursor.Offset())
	assert.ErrorIs(t, eof, io.EOF)
}

// dataErrorReader returns its data and error together to exercise deferred errors.
type dataErrorReader struct {
	data  string
	err   error
	reads int
}

func (r *dataErrorReader) Read(p []byte) (int, error) {
	r.reads++
	n := copy(p, r.data)
	r.data = r.data[n:]
	if len(r.data) == 0 {
		return n, r.err
	}
	return n, nil
}

func TestStreamingIOFailures(t *testing.T) {
	// Arrange.
	failure := errors.New("transport failed")
	for _, input := range []string{"1.8(12*°C)", "1-0:1.8.0*255"} {
		for end := 0; end <= len(input); end++ {
			t.Run(fmt.Sprintf("%s/%d", input, end), func(t *testing.T) {
				// Arrange.
				reader := &dataErrorReader{data: input[:end], err: failure}
				cursor := offsetCursor{Cursor: xobis.NewReaderCursor(reader), base: 1 << 40}
				var pe *xobis.CursorError
				var parseErr error
				var reading xobis.Reading
				var pattern xobis.Pattern

				// Act.
				if strings.Contains(input, "(") {
					reading, parseErr = xobis.ParseReadingFrom(&cursor)
				} else {
					pattern, parseErr = xobis.ParsePatternFrom(&cursor)
				}

				// Assert.
				if end == len(input) && strings.Contains(input, "(") {
					require.NoError(t, parseErr)
					assert.Equal(t, input, reading.String())
				} else {
					require.ErrorAs(t, parseErr, &pe)
					assert.ErrorIs(t, parseErr, failure)
					assert.NotErrorIs(t, parseErr, xobis.ErrSyntax)
					assert.Equal(t, cursor.base+int64(end), pe.Offset)
					assert.Zero(t, reading)
					assert.Zero(t, pattern)
				}
				assert.Equal(t, cursor.base+int64(end), cursor.Offset())
				assert.Equal(t, 1, reader.reads)
			})
		}
	}
}

func TestReaderCursorRetainsPeekErrorUntilRead(t *testing.T) {
	// Arrange.
	failure := errors.New("temporary transport failure")
	reader := &dataErrorReader{err: failure}
	cursor := xobis.NewReaderCursor(reader)

	// Act.
	_, firstErr := cursor.PeekByte()
	_, repeatedErr := cursor.PeekByte()
	_, readErr := cursor.ReadByte()
	readsBeforeRetry := reader.reads
	offsetBeforeRetry := cursor.Offset()
	reader.data, reader.err = "1.8;", io.EOF
	got, retryErr := xobis.ParsePatternFrom(cursor)

	// Assert.
	assert.ErrorIs(t, firstErr, failure)
	assert.ErrorIs(t, repeatedErr, failure)
	assert.ErrorIs(t, readErr, failure)
	assert.Equal(t, 1, readsBeforeRetry)
	assert.Zero(t, offsetBeforeRetry)
	require.NoError(t, retryErr)
	assert.Equal(t, "1.8", got.String())
	assert.Equal(t, int64(3), cursor.Offset())
}

type readFailureCursor struct {
	xobis.Cursor
	after int64
	err   error
}

func (c *readFailureCursor) ReadByte() (byte, error) {
	if c.Offset() == c.after {
		return 0, c.err
	}
	return c.Cursor.ReadByte()
}

func TestCursorFailureReadingPeekedByte(t *testing.T) {
	// Arrange.
	const input = "1.8(12*°C)"
	failure := errors.New("read failed after successful peek")
	for after := range len(input) {
		t.Run(fmt.Sprint(after), func(t *testing.T) {
			// Arrange.
			cursor := readFailureCursor{
				Cursor: &recordingCursor{input: input},
				after:  int64(after), err: failure,
			}
			var pe *xobis.CursorError

			// Act.
			got, err := xobis.ParseReadingFrom(&cursor)

			// Assert.
			assert.Zero(t, got)
			require.ErrorAs(t, err, &pe)
			assert.ErrorIs(t, err, failure)
			assert.NotErrorIs(t, err, xobis.ErrSyntax)
			assert.Equal(t, int64(after), pe.Offset)
			assert.Equal(t, int64(after), cursor.Offset())
		})
	}
}

func TestCursorFailureDoesNotRollback(t *testing.T) {
	// Arrange.
	for _, tt := range []struct {
		input                 string
		kind                  error
		errorOffset, consumed int64
	}{
		{"1.8.", xobis.ErrSyntax, 4, 4},
		{"1.8*256;", xobis.ErrRange, 4, 6},
		{"16-1.8;", xobis.ErrRange, 0, 6},
	} {
		t.Run(tt.input, func(t *testing.T) {
			// Arrange.
			cursor := xobis.NewReaderCursor(&chunkReader{input: tt.input, chunk: 1})
			var pe *xobis.CursorError

			// Act.
			got, err := xobis.ParsePatternFrom(cursor)

			// Assert.
			assert.Zero(t, got)
			require.ErrorAs(t, err, &pe)
			assert.ErrorIs(t, err, tt.kind)
			assert.Equal(t, tt.errorOffset, pe.Offset)
			assert.Equal(t, tt.consumed, cursor.Offset())
		})
	}
}

func TestReaderCursorEmptyReads(t *testing.T) {
	// Arrange.
	cursor := xobis.NewReaderCursor(&dataErrorReader{})
	var pe *xobis.CursorError

	// Act.
	got, err := xobis.ParsePatternFrom(cursor)

	// Assert.
	assert.Zero(t, got)
	require.ErrorAs(t, err, &pe)
	assert.ErrorIs(t, err, io.ErrNoProgress)
	assert.Zero(t, pe.Offset)
	assert.Zero(t, cursor.Offset())
}
