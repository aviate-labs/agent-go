package registry

//go:generate go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
//go:generate protoc -I=testdata --go_out=. testdata/registry.proto
//go:generate protoc -I=testdata --go_out=. testdata/local.proto
//go:generate protoc -I=testdata --go_out=. testdata/transport.proto
