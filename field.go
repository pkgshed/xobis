package xobis

// CodeField is a mutable text decoding adapter for [Code], for example in a
// configuration struct. Its zero value contains the valid zero [Code].
// It does not track whether an input field was absent or explicitly supplied.
type CodeField struct {
	value Code
}

// NewCodeField initializes a mutable adapter with an immutable code.
func NewCodeField(value Code) CodeField { return CodeField{value: value} }

// Value returns a copy unaffected by subsequent decoding into f.
func (f CodeField) Value() Code { return f.value }

// MarshalText implements [encoding.TextMarshaler] using decimal notation.
func (f CodeField) MarshalText() ([]byte, error) { return f.value.MarshalText() }

// UnmarshalText implements [encoding.TextUnmarshaler] using [Parse]'s decimal
// notation. It replaces the value only after successful parsing.
func (f *CodeField) UnmarshalText(text []byte) error {
	value, err := Parse(string(text))
	if err != nil {
		return err
	}
	f.value = value
	return nil
}

// PatternField is a mutable text decoding adapter for [Pattern]. Its zero value
// contains the invalid zero [Pattern], whose [Pattern.MarshalText] returns [ErrSyntax].
// It does not track whether an input field was absent or explicitly supplied.
type PatternField struct {
	value Pattern
}

// NewPatternField initializes a mutable adapter with an immutable pattern.
// The zero [Pattern] is permitted and retains its invalid-state behavior.
func NewPatternField(value Pattern) PatternField { return PatternField{value: value} }

// Value returns a copy unaffected by subsequent decoding into f.
func (f PatternField) Value() Pattern { return f.value }

// MarshalText implements [encoding.TextMarshaler], preserving presence flags.
func (f PatternField) MarshalText() ([]byte, error) { return f.value.MarshalText() }

// UnmarshalText implements [encoding.TextUnmarshaler] using [ParsePattern]'s decimal
// notation. It preserves omitted groups and replaces the value only on success.
func (f *PatternField) UnmarshalText(text []byte) error {
	value, err := ParsePattern(string(text))
	if err != nil {
		return err
	}
	f.value = value
	return nil
}
