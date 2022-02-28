# Go Agent for the Internet Computer

[![Go Version](https://img.shields.io/github/go-mod/go-version/aviate-labs/agent-go.svg)](https://github.com/aviate-labs/agent-go)
[![GoDoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/aviate-labs/agent-go)

## Testing

There are two types of tests within this repository; the normal go tests and [DFX](https://github.com/dfinity/sdk) dependent tests. The test suite will run a local replica through DFX to run some e2e tests. If you do not have it installed then those tests will be ignored.

```shell
go test -v ./...
```

## Packages

| Package Name | Links | Description |
|---|---|---|
| `agent` | [![README](https://img.shields.io/badge/-README-green)](https://github.com/aviate-labs/agent-go) [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/agent-go) | A library to talk directly to the Replica. |  
| `identity` | [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/agent-go/identity) | A library that creates/manages identities. |

### Dependencies

| Package Name | Links | Description |
|---|---|---|
| `candid` | [![README](https://img.shields.io/badge/-README-green)](https://github.com/aviate-labs/candid-go) [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/candid-go) | A Candid library for Golang |
| `principal` | [![README](https://img.shields.io/badge/-README-green)](https://github.com/aviate-labs/principal-go) [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/principal-go) | Generic Identifiers for the Internet Computer |

More dependencies in the [go.mod](./go.mod) file.
