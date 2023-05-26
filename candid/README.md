# Candid

## Exceptions

### Nat and Int Types

Since `int` and `uint` types are usually 32 bits wide on 32-bit systems and 64 bits wide on 64-bit systems
we do not support them in Candid, except passing it to a `idl.Nat` and `idl.Int` respectively.

### Vector Types

Note that for arrays, you must pass the correct amount of elements.

### Variant Types

Variants always expect exactly one label and one value, labels are case-sensitive.

## Mapping between Candid and Go Types

| IDL Type    | Accepted Go Value(s)                                     | Accepted Go Type(s)                               |
|-------------|----------------------------------------------------------|---------------------------------------------------|
| `null`      | `idl.Null`, `nil`                                        | `idl.Null`                                        |
| `bool`      | `bool`                                                   | `bool`                                            |
| `nat`       | `idl.Nat`, `uint`, `uint64`, `uint32`, `uint16`, `uint8` | `idl.Nat`                                         |
| `int`       | `idl.Int`, `int`, `int64`, `int32`, `int16`, `int8`      | `idl.Int`                                         |
| `nat64`     | `uint64`, `uint32`, `uint16`, `uint8`                    | `idl.Nat`, `uint64`                               |
| `nat32`     | `uint32`, `uint16`, `uint8`                              | `idl.Nat`,  `uint64`, `uint32`                    |
| `nat16`     | `uint16`, `uint8`                                        | `idl.Nat`,  `uint64`, `uint32`, `uint16`          |
| `nat8`      | `uint8`                                                  | `idl.Nat`,  `uint64`, `uint32`, `uint16`, `uint8` |
| `int64`     | `int64`, `int32`, `int16`, `int8`                        | `idl.Int`, `int64`                                |
| `int32`     | `int32`, `int16`, `int8`                                 | `idl.Int`, `int64`, `int32`                       |
| `int16`     | `int16`, `int8`                                          | `idl.Int`, `int64`, `int32`, `int16`              |
| `int8`      | `int8`                                                   | `idl.Int`, `int64`, `int32`, `int16`, `int8`      |
| `float64`   | `float64`, `float32`                                     | `float64`                                         |
| `float32`   | `float32`                                                | `float64`, `float32`                              |
| `text`      | `string`                                                 | `string`                                          |
| `reserved`  | N/A                                                      | `idl.Reserved`                                    |
| `empty`     | N/A                                                      | `idl.Empty`                                       |
| `principal` | `principal.Principal`, `[]byte`                          | `principal.Principal`                             |
| `opt {x}`   | `nil`, `{x}`                                             | `*{x}`                                            |
| `vec {x}`   | `nil`, `[]{x}`, `[i]{x}`, `[]any`, `[i]{any}`,           | `[]{x}`, `[i]{x}`                                 |
| `record`    | `struct`, `map[string]any`                               | `struct`, `map[string]any`                        |
| `variant`   | `struct`, `map[string]any`                               | `struct`, `map[string]any`                        |
