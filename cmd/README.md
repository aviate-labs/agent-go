# Agent CLI

```
go install github.com/aviate-labs/agent-go/cmd
cmd --help
> ERROR: command "--help" not found
```

```shell
go run main.go fetch ryjl3-tyaaa-aaaaa-aaaba-cai --output=ledger.did
go run main.go generate did ledger.did ledger --output=ledger.go --packageName=main
go fmt ledger.go
```

**OR**

```shell
go run main.go generate remote ryjl3-tyaaa-aaaaa-aaaba-cai ledger --output=ledger.go --packageName=main
go fmt ledger.go
```
