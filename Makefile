.PHONY: test test-cover gen gen-ic fmt

test:
	go test -v -cover ./...

test-registry:
	REGISTRY_TEST_ENABLE=true go test -v -cover ./registry/...

check-moc:
	find ic -type f -name '*.mo' -print0 | xargs -0 $(shell dfx cache show)/moc --check

test-cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

gen:
	cd candid && go generate
	cd pocketic && go generate
	cd registry && go generate

gen-ic:
	go run ic/testdata/gen.go
	go run ic/sns/testdata/gen.go

fmt:
	go mod tidy
	gofmt -s -w .
	goarrange run -r .
	golangci-lint run ./...
