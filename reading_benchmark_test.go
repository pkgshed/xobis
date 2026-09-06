package xobis_test

import (
	"io"
	"strings"
	"testing"

	"github.com/pkgshed/xobis"
	"github.com/stretchr/testify/require"
)

func BenchmarkParseReading(b *testing.B) {
	// Arrange.
	const input = "1.8.0(+00123.4500*kWh)"
	var err error
	b.ReportAllocs()

	// Act.
	for b.Loop() {
		_, err = xobis.ParseReading(input)
		if err != nil {
			break
		}
	}

	// Assert outside the measured loop.
	require.NoError(b, err)
}

func BenchmarkParseReadingFrom(b *testing.B) {
	// Arrange. Reuse one cursor over a continuous stream of readings.
	cursor := xobis.NewReaderCursor(&repeatedReading{input: "1.8.0(+00123.4500*kWh)"})
	var err error
	b.ReportAllocs()

	// Act.
	for b.Loop() {
		_, err = xobis.ParseReadingFrom(cursor)
		if err != nil {
			break
		}
	}

	// Assert outside the measured loop.
	require.NoError(b, err)
}

type repeatedReading struct {
	input  string
	reader *strings.Reader
}

func (r *repeatedReading) Read(p []byte) (int, error) {
	if r.reader == nil {
		r.reader = strings.NewReader(r.input)
	}
	n, err := r.reader.Read(p)
	if err == io.EOF {
		r.reader.Reset(r.input)
		return r.reader.Read(p)
	}
	return n, err
}
