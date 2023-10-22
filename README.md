# Go Agent for the Internet Computer

[![Go Version](https://img.shields.io/github/go-mod/go-version/aviate-labs/agent-go.svg)](https://github.com/aviate-labs/agent-go)
[![GoDoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/aviate-labs/agent-go)

## Packages

| Package Name  | Links                                                                                                                                                                                                 | Description                                                                   |
|---------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------|
| `agent`       | [![README](https://img.shields.io/badge/-README-green)](https://github.com/aviate-labs/agent-go) [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/agent-go) | A library to talk directly to the Replica.                                    |  
| `candid`      | [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/agent-go/candid)                                                                                           | A Candid library for Golang.                                                  |
| `certificate` | [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/agent-go/certificate)                                                                                      | A Certification library for Golang.                                           |
| `gen`         | [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/agent-go/gen)                                                                                              | A library to generate Golang clients.                                         |
| `ic`          | [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/agent-go/ic)                                                                                               | Multiple auto-generated sub-modules to talk to the Internet Computer services |
| `identity`    | [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/agent-go/identity)                                                                                         | A library that creates/manages identities.                                    |
| `principal`   | [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/agent-go/principal)                                                                                        | Generic Identifiers for the Internet Computer                                 |

More dependencies in the [go.mod](./go.mod) file.

## CLI

```shell
go install github.com/aviate-labs/agent-go/cmd/goic@latest
```

Read more [here](cmd/goic/README.md)

## Testing

There are two types of tests within this repository; the normal go tests and [DFX](https://github.com/dfinity/sdk)
dependent tests. The test suite will run a local replica through DFX to run some e2e tests. If you do not have it
installed then those tests will be ignored.

```shell
go test -v ./...
```
