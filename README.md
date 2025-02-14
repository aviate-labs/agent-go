# Go Agent for the Internet Computer

[![Go Version](https://img.shields.io/github/go-mod/go-version/aviate-labs/agent-go.svg)](https://github.com/aviate-labs/agent-go)
[![GoDoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/aviate-labs/agent-go)

```shell
go get github.com/aviate-labs/agent-go
```

## Getting Started

The agent is a library that allows you to talk to the Internet Computer.

```go
package main

import (
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/ic"
	"log"
)

type (
	Account struct {
		Account string `ic:"account"`
	}

	Balance struct {
		E8S uint64 `ic:"e8s"`
	}
)

func main() {
	a, _ := agent.New(agent.DefaultConfig)

	var balance Balance
	if err := a.Query(
		principal.MustDecode("ryjl3-tyaaa-aaaaa-aaaba-cai"), "account_balance_dfx",
		[]any{Account{"9523dc824aa062dcd9c91b98f4594ff9c6af661ac96747daef2090b7fe87037d"}},
		[]any{&balance},
	); err != nil {
		log.Fatal(err)
	}

	_ = balance // Balance{E8S: 0}
}

```

### Using an Identity

Supported identities are `Ed25519` and `Secp256k1`. By default, the agent uses the anonymous identity.

```go
id, _ := identity.NewEd25519Identity(publicKey, privateKey)
config := agent.Config{
    Identity: id,
}
```

### Using the Local Replica

If you are running a local replica, you can use the `FetchRootKey` option to fetch the root key from the replica.

```go
u, _ := url.Parse("http://localhost:8000")
config := agent.Config{
    ClientConfig: &agent.ClientConfig{Host: u},
    FetchRootKey: true,
    DisableSignedQueryVerification: true,
}
```

### Using WalletCanister as proxy 
Please refer ic/wallet/README.md for detail.



## Packages

You can find the documentation for each package in the links below. Examples can be found throughout the documentation.

| Package Name      | Links                                                                                                                                                                                                   | Description                                                                     |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| `agent`           | [![README](https://img.shields.io/badge/-README-green)](https://github.com/aviate-labs/agent-go) [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/agent-go)   | A library to talk directly to the Replica.                                      |
| `candid`          | [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/agent-go/candid)                                                                                             | A Candid library for Golang.                                                    |
| `certification`   | [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/agent-go/certificate)                                                                                        | A Certification library for Golang.                                             |
| `gen`             | [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/agent-go/gen)                                                                                                | A library to generate Golang clients.                                           |
| `identity`        | [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/agent-go/identity)                                                                                           | A library that creates/manages identities.                                      |
| `principal`       | [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/agent-go/principal)                                                                                          | Generic Identifiers for the Internet Computer                                   |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| `ic-go`           | [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/ic-go)                                                                                                       | Multiple auto-generated sub-modules to talk to the Internet Computer services   |
| `pocketic-go`     | [![DOC](https://img.shields.io/badge/-DOC-blue)](https://pkg.go.dev/github.com/aviate-labs/pocketic-go)                                                                                                 | A client library to talk to the PocketIC Server.                                |

More dependencies in the [go.mod](./go.mod) file.

## CLI

```shell
go install github.com/aviate-labs/agent-go/cmd/goic@latest
```

Read more [here](cmd/goic/README.md)

## Testing

This repository contains two types of tests: standard Go tests and [PocketIC](https://github.com/dfinity/pocketic)
-dependent tests. The test suite runs a local PocketIC server using the installed pocket-ic-server to execute some
end-to-end (e2e) tests. If pocket-ic-server is not installed, those specific tests will be skipped.

```shell
go test -v ./...
```

## Reference Implementations

- [Rust Agent](https://github.com/dfinity/agent-rs/)
- [JavaScript Agent](https://github.com/dfinity/agent-js/)
