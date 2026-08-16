.PHONY: test test-cover gen fmt

# The candid conformance tests read CANDID_TEST_DIR, which the dev shell
# exports along with the pinned go and golangci-lint. Re-enter it unless we are
# already inside one: CI runs `nix develop --command make ...`, which sets
# IN_NIX_SHELL.
ifdef IN_NIX_SHELL
NIX :=
else
NIX := nix develop --command
endif

test:
	$(NIX) go test -v -cover ./...

test-cover:
	$(NIX) go test -v -coverprofile=coverage.out ./...
	$(NIX) go tool cover -html=coverage.out

gen:
	cd candid/internal && $(NIX) go generate
	cd certification/http/certexp && $(NIX) go generate

fmt:
	$(NIX) go mod tidy
	$(NIX) gofmt -s -w .
	$(NIX) go fix ./...
	$(NIX) golangci-lint run ./...
