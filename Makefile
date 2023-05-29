.PHONY: test test-cover gen gen-ic fmt

test:
	go test -v -cover ./...

test-cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

gen:
	cd candid && go generate

gen-ic:
	go run ic/testdata/gen.go

fmt:
	go mod tidy
	gofmt -s -w .
	goarrange run -r .
	golangci-lint run ./...
