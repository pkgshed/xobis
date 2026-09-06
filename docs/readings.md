# Meter readings

`ParseReading` accepts one numeric reading with an optional unit:

```go
reading, err := xobis.ParseReading("1.8.0(123.4500*kWh)")
if err != nil {
    return err
}
fmt.Println(reading.Identifier())                   // 1.8.0
fmt.Println(reading.Value())                        // 123.4500
fmt.Println(reading.Unit() == xobis.UnitKilowattHour) // true
fmt.Println(reading)                                // 1.8.0(123.4500*kWh)
```

## Accepted format

The grammar is `identifier(value[*unit])`. The identifier follows [pattern syntax](values-and-patterns.md#parsing-and-matching). For example, `1-0:1.8.0*255(123*kWh)` has both a storage group and a unit separator.

Values use ASCII `[+-]?[0-9]+(\.[0-9]+)?`, such as `+00123.4500` or `-0.000`. `Decimal` preserves the exact spelling with no fixed numeric magnitude limit. Exponents and special floating-point values are unsupported.

Omitting the unit gives `UnitNone`; an explicit `*` requires a nonempty symbol. Units are case-sensitive UTF-8 graphic characters excluding whitespace, parentheses, and `*`. Custom symbols such as `°C` and `m³/h` are accepted.

Input must end immediately after `)`, with no surrounding whitespace or line endings. Formatting emits no line ending.

For readings embedded in a larger document, use [`ParseReadingFrom`](cursors.md).

## Construction

Construct a reading from an existing pattern and a parsed decimal:

```go
identifier, err := xobis.ParsePattern("1.8.0")
if err != nil {
    return err
}
value, err := xobis.ParseDecimal("+00123.4500")
if err != nil {
    return err
}
reading, err := xobis.NewReading(identifier, value, xobis.UnitKilowattHour)
if err != nil {
    return err
}
fmt.Println(reading) // 1.8.0(+00123.4500*kWh)
```

`NewReading` validates the identifier, value, and unit. Predefined units include:

| Constant           | Symbol        |
|--------------------|---------------|
| `UnitNone`         | `""` (absent) |
| `UnitWatt`         | `W`           |
| `UnitKilowatt`     | `kW`          |
| `UnitWattHour`     | `Wh`          |
| `UnitKilowattHour` | `kWh`         |
| `UnitVolt`         | `V`           |
| `UnitAmpere`       | `A`           |
| `UnitHertz`        | `Hz`          |

Use `xobis.Unit("m³/h")` for a custom symbol.

## Equality and zero values

`Decimal` and `Reading` are immutable, comparable values suitable as map keys. Equality compares representation: `123` differs from `123.00`, and `0` differs from `-0`. Readings also compare identifier presence and unit spelling.

Formatting normalizes the identifier and preserves decimal and unit spelling. The zero `Decimal` and zero `Reading` are invalid and format as empty strings; validation returns an error wrapping `ErrSyntax`.

## JSON and mutable decoding

Use `ReadingField` to decode a JSON string:

```go
var payload struct {
    Reading xobis.ReadingField `json:"reading"`
}
if err := json.Unmarshal([]byte(`{"reading":"1.8.0(123.4500*kWh)"}`), &payload); err != nil {
    return err
}
reading := payload.Reading.Value()
fmt.Println(reading.Value()) // 123.4500
```

See [encoding and decoding](encoding.md#mutable-decoding-fields) for adapter behavior and [validation](validation.md) for errors and scope.
