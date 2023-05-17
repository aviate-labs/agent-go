.PHONY: test test-ledger gen gen-ic fmt

test:
	go test -v -cover ./...
	
test-ledger:
	cd ic; dfx start --background --clean
	cd ic/testdata; dfx deploy --no-wallet
	cd ic; DFX=true go test -v icpledger_test.go; dfx stop

gen:
	cd candid && go generate

gen-ic:
	go run ic/testdata/gen.go

fmt:
	go mod tidy
	gofmt -s -w .
	goarrange run -r .
