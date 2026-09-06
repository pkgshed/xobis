# Validation and scope

Identifier validation checks syntax, required groups, and numeric ranges: A is 0..15 and B through F are 0..255. See [values and patterns](values-and-patterns.md) for decimal notation, [encoding](encoding.md) for binary and hexadecimal forms, and [meter readings](readings.md) for values and units.

Errors wrap `ErrSyntax` for malformed input or missing required groups, and `ErrRange` for values outside the allowed ranges. Use `errors.Is(err, xobis.ErrSyntax)` or `errors.Is(err, xobis.ErrRange)` to distinguish them. Parsers and constructors return zero values on failure.

String parsers return `*ParseError`, which records the original input and a zero-based byte offset. Cursor parsers return `*CursorError` with an absolute offset and a wrapped syntax, range or I/O error. See [parser composition](cursors.md#errors-and-recovery) for streaming errors and recovery.

Validation accepts reserved and manufacturer-specific identifiers within range. It does not establish that an identifier is assigned, supported by a device, or compatible with a reading's unit. The library does not implement full OBIS display notation, meter protocols, or measurement arithmetic.

The implemented subset draws on [IEC 62056-6-1:2023, clauses 4–5 (preview)](https://cdn.standards.iteh.ai/samples/104764/8356233bf80942e9ba6d5ac9cbdf3af0/IEC-62056-6-1-2023.pdf) for structure and ranges, and [DLMS UA Blue Book edition 7, clauses 5.4, 5.5.1, and 5.7 (public excerpt)](https://www.cs.ru.nl/~marko/onderwijs/bss/SmartMeter/Excerpt_BB7.pdf) for electricity constants and shortened notation.
