package ledger

//go:generate go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
//go:generate protoc -I=testdata --go_out=. testdata/ledger.proto
