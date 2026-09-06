package xobis

// Presence is a bit set identifying supplied OBIS value groups.
//
//go:generate go tool -modfile=tools/tools.mod github.com/0x5a17ed/stringer/v2 -output presence_flags.go -flags Presence=getterSetter
type Presence uint8

const (
	PresentA Presence = 1 << iota
	PresentB
	PresentC
	PresentD
	PresentE
	PresentF
)
