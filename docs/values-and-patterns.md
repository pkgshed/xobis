# Values and patterns

## Typed construction

Each group has a distinct named type:

| Group | Go type          | Meaning                                     |
|-------|------------------|---------------------------------------------|
| A     | `Medium`         | Energy type or abstract data                |
| B     | `Channel`        | Measurement or communication channel        |
| C     | `Quantity`       | Physical or abstract data item              |
| D     | `Processing`     | Processing or classification                |
| E     | `Classification` | Tariff, harmonic, or another classification |
| F     | `Storage`        | Historical data or further classification   |

```go
code, err := xobis.NewCode(xobis.Groups{
    A: xobis.MediumElectricity,
    B: xobis.ChannelUnspecified,
    C: xobis.ElectricityActivePowerImport,
    D: xobis.ElectricityTimeIntegral1,
    E: xobis.ElectricityTariffTotal,
    F: xobis.StorageNotUsed,
})
if err != nil {
    return err
}
fmt.Println(code) // 1-0:1.8.0*255
```

`Code` and `Pattern` are immutable, comparable values suitable as map keys. `NewCode` validates its input. To change a code, edit the copy returned by `Groups()` and construct a new value:

```go
groups := code.Groups()
groups.E = xobis.ElectricityTariff1
changed, err := xobis.NewCode(groups)
if err != nil {
    return err
}
fmt.Println(code)    // 1-0:1.8.0*255
fmt.Println(changed) // 1-0:1.8.1*255
```

## Parsing and matching

`Parse` requires all six groups and accepts the `A-B:C.D.E*F` and `A.B.C.D.E.F` notation:

```go
code, err := xobis.Parse("1.0.1.8.0.255")
if err != nil {
    return err
}
fmt.Println(code) // 1-0:1.8.0*255
```

`ParsePattern` additionally accepts `[A-][B:]C.D[.E][*F]`. Brackets mark independently optional groups and their separators; C and D are required. Six dotted groups mean A through F; two or three mean C.D or C.D.E. Four or five dotted groups are rejected.

Decimal groups accept leading zeroes. Both string parsers require the whole input and reject whitespace and signs.

```go
pattern, err := xobis.ParsePattern("1.8.0")
if err != nil {
    return err
}
fmt.Println(pattern.Presence().PresentA()) // false
fmt.Println(pattern.Presence().PresentE()) // true
fmt.Println(pattern.Match(code))          // true
```

Omitted groups match any value; explicit values, including zero and 255, match only themselves. `pattern.Groups()` returns the groups and their `Presence` flags. Omitted groups are stored as zero, so use the flags to distinguish omission from an explicit zero.

Use `NewPattern` to construct a pattern directly from typed values:

```go
pattern, err := xobis.NewPattern(xobis.Groups{
    C: xobis.ElectricityActivePowerImport,
    D: xobis.ElectricityTimeIntegral1,
    E: xobis.ElectricityTariffTotal,
}, xobis.PresentC | xobis.PresentD | xobis.PresentE)
if err != nil {
    return err
}
fmt.Println(pattern) // 1.8.0
```

`NewPattern` requires C and D, rejects undefined presence bits, validates selected groups, and clears omitted groups. Invalid presence flags wrap `ErrSyntax`; a selected group outside its range wraps `ErrRange`.

Equality compares group values and, for patterns, presence flags. An omitted group and an explicit zero are different keys.

`pattern.CompleteCode()` requires all six groups to be present. Otherwise, it returns the zero `Code` and an error wrapping `ErrSyntax`.

## Formatting and zero values

Formatting uses `A-B:C.D.E*F` notation, omits absent pattern groups, and removes leading zeroes.

The zero `Code` is the valid `0-0:0.0.0*0`. The zero `Pattern` is invalid, matches nothing, and formats as an empty string.

See [encoding](encoding.md) for text and binary conversions, and [validation](validation.md) for accepted ranges and error handling.
