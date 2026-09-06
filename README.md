# xobis

A Go library for parsing, formatting, matching, and validating numeric OBIS identifiers. Immutable `Code` and `Pattern` values have strongly typed groups and work as map keys.

Supports decimal and hexadecimal notation, byte and integer conversions, [numeric meter readings](docs/readings.md), [parser composition](docs/cursors.md), JSON decoding adapters, and custom hashers. Requires Go 1.24; the runtime package uses only the standard library.

## Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/pkgshed/xobis"
)

func main() {
	code, err := xobis.Parse("1-0:1.8.0*255")
	if err != nil {
		log.Fatal(err)
	}
	pattern, err := xobis.ParsePattern("1.8.0") // Omitted groups match any value.
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(code)                                       // 1-0:1.8.0*255
	fmt.Println(pattern.Match(code))                        // true
	fmt.Println(code.Groups().A == xobis.MediumElectricity) // true
	fmt.Println(xobis.ToBytes(code))                        // [1 0 1 8 0 255]
	fmt.Println(xobis.ToHex(code))                          // 0100010800ff
	fmt.Printf("0x%012x\n", xobis.ToUint64(code))           // 0x0100010800ff
}
```

Validation checks syntax and numeric ranges. It does not verify that an identifier is assigned by a standard or supported by a device.

[Read the documentation](docs/index.md).
