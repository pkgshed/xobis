---
title: Hashing
weight: 50
---

# Hashing

For collections that accept [`maphash.Hasher`](https://pkg.go.dev/hash/maphash@go1.27.0#Hasher) (Go 1.27+):

```go
import (
    "hash/maphash"

    "github.com/pkgshed/xobis"
)

var codeHasher maphash.Hasher[xobis.Code] = xobis.CodeHasher{}
var patternHasher maphash.Hasher[xobis.Pattern] = xobis.PatternHasher{}
```

Equality uses Go's `==`, including a pattern's [presence flags](values-and-patterns.md#parsing-and-matching). Hashing accepts the zero `Pattern` without validation.

`Hash` appends to the supplied hash. Use the collection's seed; hash values are not stable identifiers to store or transmit.

Use one hasher implementation consistently within a collection. Switching to `maphash.ComparableHasher[T]` may produce different hashes.
