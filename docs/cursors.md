---
title: Parser composition
weight: 30
---

# Parser composition

Use `ParsePatternFrom` or `ParseReadingFrom` to parse one value from a stream or an enclosing parser. Use `ParsePattern` or `ParseReading` for a complete string.

## Streaming cursor

The enclosing parser implements this interface:

```go
type Cursor interface {
    PeekByte() (byte, error)
    ReadByte() (byte, error)
    Offset() int64
}
```

`Offset()` is the nonnegative, zero-based position of the next byte in the logical input. A successful `ReadByte` advances it by one. Peeks and failed reads do not advance it. A successful peek exposes the byte returned by the next successful read.

Operations may block awaiting input. `io.EOF` means input has ended, not that another chunk is temporarily unavailable. Pass a non-nil cursor and do not use it concurrently or advance it independently during a parse.

## Reading from an io.Reader

`NewReaderCursor(reader)` supplies a cursor with a buffered reader and one byte of lookahead. Its logical offset starts at zero. Construct it with a non-nil reader; the zero `ReaderCursor` is not usable. Do not copy the cursor.

```go
cursor := xobis.NewReaderCursor(reader)
reading, err := xobis.ParseReadingFrom(cursor)
if err != nil {
    return err
}
fmt.Println(reading)
```

The buffer can read ahead from the underlying reader. Continue consuming through the same cursor, including any separators and subsequent values. Repeated `PeekByte` calls return the same byte or error until `ReadByte` is called. After a failed `ReadByte`, another operation may retry the reader.

The adapter does not close the reader or retain consumed input. Cancellation, deadlines, and input limits belong to the supplied reader or enclosing parser. Readings own their decimal and unit text, so later buffer reuse cannot change returned values. Those strings can grow with the value length; identifiers are parsed without retaining their text.

`ParseReadingFrom` stops immediately after `)` without inspecting the following byte. It does not wait for a separator or EOF. Neither cursor parser skips whitespace or line endings; the enclosing parser handles them. Adjacent complete readings can be parsed with consecutive calls.

## Pattern boundaries

`ParsePatternFrom` leaves the first byte unrelated to the identifier untouched. For example, `1.8xyz` returns `1.8` and leaves `xyz`; the caller decides whether the suffix is legal. OBIS punctuation (`.`, `-`, `:`, `*`) continues the grammar, so `1.8.` fails.

A pattern needs a following delimiter or EOF to establish completion. On a live stream, parsing `1.8` may block waiting to learn whether another digit or group follows. An I/O failure at that point is an error, not a successful boundary.

## Errors and recovery

Cursor parsing returns a zero value and `*CursorError` on failure. Its `Offset` is an absolute `int64` byte position; its `Err` wraps `ErrSyntax`, `ErrRange`, or the original I/O error. Use `errors.Is` and `errors.As` to inspect the cause. EOF where required input is missing, including empty input, becomes `ErrSyntax`. An invalid or incomplete UTF-8 sequence points to its starting byte; an I/O failure points to the position where reading failed.

A negative initial cursor offset returns a configuration error wrapping `ErrRange`, without a `CursorError` or any input access. String APIs retain `*ParseError`, including the complete source and existing byte-offset behavior.

A failed parse may consume input. The error offset can precede the cursor's current position, and it must not be used as a rollback position. To try another grammar, the parent must establish a checkpoint **before** parsing, restore it on failure, and release it on success. For streaming inputs that requires retaining and replaying bytes read since the checkpoint, with a caller-defined buffer limit. Logical rollback cannot undo reads from the underlying stream. This library does not provide a checkpoint or transaction adapter.
