package xobis

import (
	"unicode"
	"unicode/utf8"
)

// Unit is a case-sensitive unit symbol, or [UnitNone] when no unit was supplied.
// The constants are a convenience subset; unfamiliar symbols are supported.
// [NewReading] validates units as UTF-8 graphic characters excluding whitespace,
// parentheses and '*'. It does not verify their meaning or convert units.
type Unit string

const (
	UnitNone         Unit = ""
	UnitWatt         Unit = "W"
	UnitKilowatt     Unit = "kW"
	UnitWattHour     Unit = "Wh"
	UnitKilowattHour Unit = "kWh"
	UnitVolt         Unit = "V"
	UnitAmpere       Unit = "A"
	UnitHertz        Unit = "Hz"
)

// unitErrorOffset returns the first invalid byte or -1. The empty symbol is
// valid here because [UnitNone] represents absence; parsing rejects an empty '*'.
func unitErrorOffset(s string) int {
	for i := 0; i < len(s); {
		size := unitRuneSize(s[i:])
		if size == 0 {
			return i
		}
		i += size
	}
	return -1
}

// unitRuneSize returns the byte length of the first valid unit rune, or zero.
func unitRuneSize(s string) int {
	r, size := utf8.DecodeRuneInString(s)
	if (r == utf8.RuneError && size == 1) || !unicode.IsGraphic(r) ||
		unicode.IsSpace(r) || r == '(' || r == ')' || r == '*' {
		return 0
	}
	return size
}
