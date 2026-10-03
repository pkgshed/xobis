---
title: Encoding and decoding
weight: 40
---

# Encoding and decoding

## Text marshalling

The types `Code`, `Pattern`, and `Reading` implement the `encoding.TextMarshaler` interface for JSON string encoding and use as map keys. `String` and text marshalling use decimal notation. Marshalling rejects the invalid zero values of `Pattern` and `Reading`. The `encoding.TextUnmarshaler` interface is deliberately unsupported to preserve the immutability of these types.

## Mutable decoding fields

Use `CodeField`, `PatternField`, or `ReadingField` for text and JSON decoding. Their `Value()` methods return the underlying values:

```go
var config struct {
    OBIS   xobis.CodeField    `json:"obis"`
    Filter xobis.PatternField `json:"filter"`
}
if err := json.Unmarshal([]byte(`{"obis":"1-0:1.8.0*255","filter":"1.8.0"}`), &config); err != nil {
    return err
}
code := config.OBIS.Value()
pattern := config.Filter.Value()
fmt.Println(pattern.Match(code)) // true
```

The adapters decode using `Parse`, `ParsePattern`, and `ParseReading`, respectively, and also support text marshaling. Failed decoding preserves the previous value; successful decoding leaves previously extracted values unchanged.

Use `NewCodeField(code)`, `NewPatternField(pattern)`, or `NewReadingField(reading)` to wrap an existing value. Zero fields contain the corresponding core zero values. The adapters do not track whether an input field was missing or explicitly supplied.

For JSON map decoding, use `CodeField`, `PatternField`, or `ReadingField` as the key type and extract each key's `Value()`.

See [meter readings](readings.md#json-and-mutable-decoding) for a `ReadingField` example.

## Hexadecimal, bytes, and packed codes

Use `ParseHex` for six hexadecimal bytes in A through F order, or `FromUint64` for a packed integer:

```go
code, err := xobis.ParseHex("0100010800ff") // Also accepts "0x0100010800ff".
if err != nil {
    return err
}
same, err := xobis.FromUint64(0x0100010800ff)
if err != nil {
    return err
}
fmt.Println(code)         // 1-0:1.8.0*255
fmt.Println(code == same) // true
```

Hexadecimal text requires exactly 12 hexadecimal digits after an optional `0x` or `0X` prefix. Both letter cases are accepted. `FromUint64` uses the packed representation below and rejects nonzero bits above bit 47.

Use `FromBytes` for exactly six raw bytes in the same A through F order:

```go
code, err := xobis.FromBytes([]byte{0x01, 0x00, 0x01, 0x08, 0x00, 0xff})
if err != nil {
    return err
}
fmt.Println(code) // 1-0:1.8.0*255
```

`FromBytes` copies the input without modifying it. All three constructors check the [group ranges](validation.md).

Convert a `Code` back with:

| Function         | Result    | Representation                                            |
|------------------|-----------|-----------------------------------------------------------|
| `ToBytes(code)`  | `[6]byte` | Six raw bytes in A through F order                        |
| `ToHex(code)`    | `string`  | Exactly 12 lowercase hexadecimal digits, without a prefix |
| `ToUint64(code)` | `uint64`  | A in bits 40–47, F in the lowest byte, high 16 bits zero  |

These conversions return no errors. Use `raw[:]` to pass the array returned by `ToBytes` to `FromBytes`:

```go
raw := xobis.ToBytes(code)
restored, err := xobis.FromBytes(raw[:])
if err != nil {
    return err
}
fmt.Println(restored == code) // true
```

Hexadecimal, byte, and integer formats represent complete codes. For a pattern, use `CompleteCode()` when all groups are present, or text marshaling to preserve omissions. To construct a pattern from a code, pass `code.Groups()` and the desired presence flags to [`NewPattern`](values-and-patterns.md#parsing-and-matching).
