.PHONY: test test-cover gen gen-ic fmt

test:
	go test -v -cover ./...

test-registry:
	REGISTRY_TEST_ENABLE=true go test -v -cover ./clients/registry/...

test-ledger:
	LEDGER_TEST_ENABLE=true go test -v -cover ./clients/ledger/...

test-all:
	REGISTRY_TEST_ENABLE=true LEDGER_TEST_ENABLE=true go test -v -cover ./...

check-moc:
	find ic -type f -name '*.mo' -print0 | xargs -0 $(shell dfx cache show)/moc --check

test-cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

gen:
	cd candid/internal && go generate
	cd certification/http/certexp && go generate
	cd clients/ledger && go generate
	cd clients/registry && go generate

fmt:
	go mod tidy
	gofmt -s -w .
	go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -fix -test ./...
	goarrange run -r .
	golangci-lint run ./...
