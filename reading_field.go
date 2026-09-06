package xobis

// ReadingField is a mutable text decoding adapter for [Reading]. Its zero value
// contains an invalid reading and fails marshaling. It does not track whether
// an input field was absent or explicitly supplied.
type ReadingField struct {
	value Reading
}

// NewReadingField initializes an adapter from an immutable reading. The zero
// [Reading] is permitted and retains its invalid-state behavior.
func NewReadingField(value Reading) ReadingField { return ReadingField{value: value} }

// Value returns an immutable snapshot unaffected by subsequent decoding into f.
func (f ReadingField) Value() Reading { return f.value }

// MarshalText implements [encoding.TextMarshaler] by delegating to the reading.
func (f ReadingField) MarshalText() ([]byte, error) { return f.value.MarshalText() }

// UnmarshalText implements [encoding.TextUnmarshaler] using [ParseReading]. It
// replaces the value only on success, preserving the previous value on error.
func (f *ReadingField) UnmarshalText(text []byte) error {
	value, err := ParseReading(string(text))
	if err != nil {
		return err
	}
	f.value = value
	return nil
}
