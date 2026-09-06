// Package xobis parses, formats and validates numeric OBIS identifiers and readings.
//
// [Code] and [Pattern] are immutable, comparable values. [NewCode] and [NewPattern]
// validate and copy mutable [Groups] inputs. [Code.Groups] and [Pattern.Groups]
// return copies; [Pattern.Groups] also returns presence flags to distinguish
// omission from explicit zero.
//
// [Parse] accepts complete identifiers in A-B:C.D.E*F or A.B.C.D.E.F notation.
// [ParsePattern] also accepts shortened identifiers without inventing defaults.
// [ParseHex], [FromUint64] and [FromBytes] accept hexadecimal text, packed integers and
// raw bytes in A through F order. [ToHex], [ToUint64] and [ToBytes] convert valid [Code]
// values back to those representations without errors.
//
// [ParsePatternFrom] and [ParseReadingFrom] integrate with streaming parsers through
// [Cursor]. [NewReaderCursor] adapts an [io.Reader]. Parsing may consume input before
// failing; the enclosing parser owns checkpointing, delimiters and line endings.
// [CursorError] reports absolute byte offsets and wraps syntax, range or I/O errors.
// String parsers require complete input and return [ParseError] with the full source.
//
// [ParseReading] accepts a single numeric meter reading such as 1.8.0(123*kWh),
// with no surrounding whitespace or line endings. [Reading] preserves omitted
// identifier groups, exact [Decimal] spelling and an optional, extensible [Unit].
// [NewReading] constructs immutable readings from a [Pattern], a parsed [Decimal]
// and a [Unit].
//
// [Code], [Pattern] and [Reading] support text marshaling. [CodeField], [PatternField]
// and [ReadingField] provide mutable text unmarshaling adapters for configuration
// and transport structs. [CodeField.Value], [PatternField.Value] and [ReadingField.Value]
// return immutable snapshots, and failed decoding preserves prior state.
// Validation checks syntax and numeric ranges, not whether an identifier has an
// assigned measurement meaning or a reading uses an appropriate unit.
package xobis
