package xobis_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

// recordingCursor tracks logical reads and allows a parent to checkpoint its
// own buffered input independently of the OBIS parser.
type recordingCursor struct {
	input  string
	offset int64
	reads  int
	peeks  int
}

func (c *recordingCursor) Offset() int64 { return c.offset }
func (c *recordingCursor) PeekByte() (byte, error) {
	c.peeks++
	if c.offset == int64(len(c.input)) {
		return 0, io.EOF
	}
	return c.input[c.offset], nil
}
func (c *recordingCursor) ReadByte() (byte, error) {
	if c.offset == int64(len(c.input)) {
		return 0, io.EOF
	}
	b := c.input[c.offset]
	c.offset++
	c.reads++
	return b, nil
}

var _ xobis.Cursor = (*recordingCursor)(nil)

func TestParsePatternFrom(t *testing.T) {
	// Arrange.
	tests := []struct{ token, suffix string }{
		{"1.8", "xyz.0.0.0.0"},
		{"1.8.0", "(1.2.3.4)"},
		{"1.8", " 2.8"},
		{"1.8", "\r\n"},
		{"1.8", ";"},
		{"1.8", ")"},
		{"1.8", "\xff"},
		{"01.08.00", ",next"},
		{"1-0:1.8.0*255", "(123*kWh)"},
		{"01.000.001.08.000.0255", "xyz"},
		{"0-0:0.0.0*0", ""},
		{"15.255.255.255.255.255", ";1.8"},
		{"0:1.8*0", "[x]"},
		{"0-1.8.0", "(0)"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.token+tt.suffix), func(t *testing.T) {
			// Arrange.
			const prefix = "µ:"
			input := prefix + tt.token + tt.suffix
			cursor := recordingCursor{input: input, offset: int64(len(prefix))}
			want, err := xobis.ParsePattern(tt.token)
			require.NoError(t, err)

			// Act.
			got, err := xobis.ParsePatternFrom(&cursor)

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, want, got)
			assert.Equal(t, int64(len(prefix)+len(tt.token)), cursor.offset)
			assert.Equal(t, len(tt.token), cursor.reads)
			assert.Equal(t, tt.suffix, cursor.input[cursor.offset:])
			assert.Equal(t, input, cursor.input)
		})
	}
}

func TestParsePatternFromFailure(t *testing.T) {
	// Arrange.
	tests := []struct {
		input  string
		kind   error
		offset int
	}{
		{"", xobis.ErrSyntax, 0}, {" 1.8", xobis.ErrSyntax, 0},
		{"1", xobis.ErrSyntax, 1}, {"1.8.", xobis.ErrSyntax, 4},
		{"1.8.xyz", xobis.ErrSyntax, 4}, {"1.8*", xobis.ErrSyntax, 4},
		{"1.8*256;", xobis.ErrRange, 4}, {"16-1.8", xobis.ErrRange, 0},
		{"16.0.1.8.0.255;", xobis.ErrRange, 0},
		{"1.8.0.255", xobis.ErrSyntax, 9}, {"1.8.0.1.255", xobis.ErrSyntax, 11},
		{"1.0.1.8.0.255.1", xobis.ErrSyntax, 13},
		{"1-0:1.8.0.255", xobis.ErrSyntax, 9},
		{"1.8:0", xobis.ErrSyntax, 3}, {"1.8-1", xobis.ErrSyntax, 3},
		{"1.8*0.1", xobis.ErrSyntax, 5}, {"1.8*0*1", xobis.ErrSyntax, 5},
		{"1.0.1.8.0.255*1", xobis.ErrSyntax, 13},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.input), func(t *testing.T) {
			// Arrange.
			const prefix = "α="
			input := prefix + tt.input
			cursor := recordingCursor{input: input, offset: int64(len(prefix))}
			var pe *xobis.CursorError

			// Act.
			got, err := xobis.ParsePatternFrom(&cursor)

			// Assert.
			assert.ErrorIs(t, err, tt.kind)
			require.ErrorAs(t, err, &pe)
			assert.Equal(t, int64(len(prefix)+tt.offset), pe.Offset)
			assert.Zero(t, got)
			assert.GreaterOrEqual(t, cursor.offset, int64(len(prefix)))
			assert.LessOrEqual(t, cursor.offset, int64(len(input)))
			assert.Equal(t, int(cursor.offset)-len(prefix), cursor.reads)
		})
	}
}

func TestParseReadingFrom(t *testing.T) {
	// Arrange.
	tokens := []string{
		"1.8(0)", "1.8.0(+00123.4500*kWh)", "1-0:1.8.0*255(-0.00*°C)",
		"01.000.001.08.000.0255(1*m³/h)",
	}
	for _, token := range tokens {
		for _, suffix := range []string{"", "\n", "\r\n", "\r", ";", "(2)", "1.8(2)", ".1.2.3", "\xff"} {
			t.Run(fmt.Sprintf("%q", token+suffix), func(t *testing.T) {
				// Arrange.
				const prefix = "µ:"
				input := prefix + token + suffix
				cursor := recordingCursor{input: input, offset: int64(len(prefix))}
				want, err := xobis.ParseReading(token)
				require.NoError(t, err)

				// Act.
				got, err := xobis.ParseReadingFrom(&cursor)

				// Assert.
				require.NoError(t, err)
				assert.Equal(t, want, got)
				assert.Equal(t, int64(len(prefix)+len(token)), cursor.offset)
				assert.Equal(t, len(token), cursor.reads)
				assert.Equal(t, suffix, cursor.input[cursor.offset:])
				assert.Equal(t, input, cursor.input)
			})
		}
	}
}

func TestParseReadingFromFailure(t *testing.T) {
	// Arrange.
	tests := []struct {
		input  string
		kind   error
		offset int
	}{
		{"", xobis.ErrSyntax, 0}, {" 1.8(1)", xobis.ErrSyntax, 0},
		{"16-1.8(1)", xobis.ErrRange, 0}, {"1.256(1)", xobis.ErrRange, 2},
		{"1.8xyz(1)", xobis.ErrSyntax, 3}, {"1.8", xobis.ErrSyntax, 3},
		{"1.8(", xobis.ErrSyntax, 4}, {"1.8((1))", xobis.ErrSyntax, 4},
		{"1.8(NaN)", xobis.ErrSyntax, 4}, {"1.8(1e2)", xobis.ErrSyntax, 5},
		{"1.8(1.*kWh)", xobis.ErrSyntax, 6},
		{"1.8(1*)", xobis.ErrSyntax, 6}, {"1.8(1*kWh*x)", xobis.ErrSyntax, 9},
		{"1.8(1*°C\xff)\r\n", xobis.ErrSyntax, 9},
		{"1.8(1*°C", xobis.ErrSyntax, 9}, {"1.8(1*kWh\n)", xobis.ErrSyntax, 9},
		{"1.8(1", xobis.ErrSyntax, 5}, {"1.8(1\n", xobis.ErrSyntax, 5},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.input), func(t *testing.T) {
			// Arrange.
			const prefix = "α="
			input := prefix + tt.input
			cursor := recordingCursor{input: input, offset: int64(len(prefix))}
			var pe *xobis.CursorError

			// Act.
			got, err := xobis.ParseReadingFrom(&cursor)

			// Assert.
			assert.ErrorIs(t, err, tt.kind)
			require.ErrorAs(t, err, &pe)
			assert.Equal(t, int64(len(prefix)+tt.offset), pe.Offset)
			assert.Zero(t, got)
			assert.GreaterOrEqual(t, cursor.offset, int64(len(prefix)))
			assert.LessOrEqual(t, cursor.offset, int64(len(input)))
			assert.Equal(t, int(cursor.offset)-len(prefix), cursor.reads)
		})
	}
}

func TestCursorInvalidOffsets(t *testing.T) {
	// Arrange.
	const input = "1.8(1)"
	for _, offset := range []int64{-1, -100, -1 << 63} {
		t.Run(fmt.Sprint(offset), func(t *testing.T) {
			// Arrange.
			patternCursor := recordingCursor{input: input, offset: offset}
			readingCursor := patternCursor
			var pe *xobis.CursorError

			// Act.
			pattern, patternErr := xobis.ParsePatternFrom(&patternCursor)
			reading, readingErr := xobis.ParseReadingFrom(&readingCursor)

			// Assert.
			assert.ErrorIs(t, patternErr, xobis.ErrRange)
			assert.ErrorIs(t, readingErr, xobis.ErrRange)
			assert.NotErrorAs(t, patternErr, &pe)
			assert.NotErrorAs(t, readingErr, &pe)
			assert.Zero(t, pattern)
			assert.Zero(t, reading)
			assert.Equal(t, offset, patternCursor.offset)
			assert.Equal(t, offset, readingCursor.offset)
			assert.Zero(t, patternCursor.reads)
			assert.Zero(t, patternCursor.peeks)
			assert.Zero(t, readingCursor.reads)
			assert.Zero(t, readingCursor.peeks)
		})
	}
}

func TestCursorAtEndOfInput(t *testing.T) {
	// Arrange.
	for _, input := range []string{"", "1.8(1)"} {
		t.Run(input, func(t *testing.T) {
			// Arrange.
			patternCursor := recordingCursor{input: input, offset: int64(len(input))}
			readingCursor := patternCursor

			// Act.
			pattern, patternErr := xobis.ParsePatternFrom(&patternCursor)
			reading, readingErr := xobis.ParseReadingFrom(&readingCursor)

			// Assert.
			assert.ErrorIs(t, patternErr, xobis.ErrSyntax)
			assert.ErrorIs(t, readingErr, xobis.ErrSyntax)
			assert.Zero(t, pattern)
			assert.Zero(t, reading)
			assert.Equal(t, int64(len(input)), patternCursor.offset)
			assert.Equal(t, int64(len(input)), readingCursor.offset)
			assert.Zero(t, patternCursor.reads)
			assert.Equal(t, 1, patternCursor.peeks)
			assert.Zero(t, readingCursor.reads)
			assert.Equal(t, 1, readingCursor.peeks)
		})
	}
}

func TestCursorParentCanRestoreCheckpoint(t *testing.T) {
	// Arrange.
	cursor := recordingCursor{input: "1.8xyz"}
	checkpoint := cursor.Offset()
	want, err := xobis.ParsePattern("1.8")
	require.NoError(t, err)

	// Act.
	reading, readingErr := xobis.ParseReadingFrom(&cursor)
	consumedOnFailure := cursor.Offset() - checkpoint
	cursor.offset = checkpoint // The parent owns checkpointing and restoration.
	pattern, patternErr := xobis.ParsePatternFrom(&cursor)

	// Assert.
	assert.ErrorIs(t, readingErr, xobis.ErrSyntax)
	assert.Zero(t, reading)
	require.NoError(t, patternErr)
	assert.Equal(t, want, pattern)
	assert.Equal(t, int64(3), consumedOnFailure)
	assert.Equal(t, int64(3), cursor.offset)
	assert.Equal(t, 6, cursor.reads)
}

func TestCursorParsesAdjacentReadings(t *testing.T) {
	// Arrange.
	const first = "1.8(123*kWh)"
	const second = "2.8(-1.00)"
	cursor := recordingCursor{input: first + second + "\r\n"}

	// Act.
	x, xErr := xobis.ParseReadingFrom(&cursor)
	y, yErr := xobis.ParseReadingFrom(&cursor)

	// Assert.
	require.NoError(t, xErr)
	require.NoError(t, yErr)
	assert.Equal(t, first, x.String())
	assert.Equal(t, second, y.String())
	assert.Equal(t, len(first)+len(second), cursor.reads)
	assert.Equal(t, "\r\n", cursor.input[cursor.offset:])
}
