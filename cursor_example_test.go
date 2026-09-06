package xobis_test

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkgshed/xobis"
)

// textCursor is an example adapter for a parser that owns a buffered string.
type textCursor struct {
	input    string
	position int
}

func (c *textCursor) Offset() int64 { return int64(c.position) }
func (c *textCursor) PeekByte() (byte, error) {
	if c.position == len(c.input) {
		return 0, io.EOF
	}
	return c.input[c.position], nil
}
func (c *textCursor) ReadByte() (byte, error) {
	b, err := c.PeekByte()
	if err == nil {
		c.position++
	}
	return b, err
}

func ExampleParseReadingFrom() {
	// Arrange.
	cursor := xobis.NewReaderCursor(strings.NewReader("1.8.0(+00123.4500*kWh)\r\n"))

	// Act.
	reading, err := xobis.ParseReadingFrom(cursor)
	if err != nil {
		panic(err)
	}
	fmt.Println(reading)
	next, err := cursor.PeekByte()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%q\n", next)

	// Output:
	// 1.8.0(+00123.4500*kWh)
	// '\r'
}

func ExampleParsePatternFrom() {
	// Arrange.
	cursor := textCursor{input: "1.8xyz"}

	// Act.
	pattern, err := xobis.ParsePatternFrom(&cursor)
	if err != nil {
		panic(err)
	}
	fmt.Println(pattern)
	fmt.Println(cursor.input[cursor.position:])

	// Output:
	// 1.8
	// xyz
}
