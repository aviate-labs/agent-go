package ledger

//go:generate go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.3
//go:generate protoc -I=testdata --go_out=. testdata/ledger.proto
