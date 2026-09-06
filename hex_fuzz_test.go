package xobis_test

import (
	"encoding/hex"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

var referenceHex = regexp.MustCompile(`\A(?:0[xX])?([0-9a-fA-F]{12})\z`)

func FuzzParseHex(f *testing.F) {
	// Arrange the seed corpus.
	for _, input := range []string{
		"0100010800ff", "0x0100010800ff", "0X0100010800FF", "0fFfFfFfFfFf",
		"000000000000", "010203040506", "100000000000", "ff0000000000",
		"", "0x", "0x100010800ff", "00100010800ff", "0100010800fg",
		"0100010800ff\x00", "0100010800f\xff", "０0010800ff", "0100_10800ff",
	} {
		f.Add(input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		// Arrange. The oracle uses the standard hex decoder independently of
		// the production scanner, so it catches rejection of valid inputs too.
		match := referenceHex.FindStringSubmatch(input)
		var want xobis.Groups
		var packed uint64
		var wantErr error
		if match == nil {
			wantErr = xobis.ErrSyntax
		} else {
			groups, err := hex.DecodeString(match[1])
			require.NoError(t, err)
			if groups[0] > 15 {
				wantErr = xobis.ErrRange
			} else {
				want = xobis.Groups{
					A: xobis.Medium(groups[0]), B: xobis.Channel(groups[1]),
					C: xobis.Quantity(groups[2]), D: xobis.Processing(groups[3]),
					E: xobis.Classification(groups[4]), F: xobis.Storage(groups[5]),
				}
				packed, err = strconv.ParseUint(match[1], 16, 64)
				require.NoError(t, err)
			}
		}
		var constructed, roundtrip xobis.Code
		var constructErr, roundtripErr error

		// Act.
		got, err := xobis.ParseHex(input)
		if wantErr == nil {
			constructed, constructErr = xobis.FromUint64(packed)
			roundtrip, roundtripErr = xobis.Parse(got.String())
		}

		// Assert.
		if wantErr != nil {
			assertParseError(t, input, err, wantErr)
			assert.Zero(t, got)
			return
		}
		require.NoError(t, err)
		assert.Equal(t, want, got.Groups())
		require.NoError(t, constructErr)
		assert.Equal(t, want, constructed.Groups())
		require.NoError(t, roundtripErr)
		assert.Equal(t, want, roundtrip.Groups())
	})
}
