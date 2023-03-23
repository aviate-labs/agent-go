.PHONY: test

test:
	go test -v ./...
	
test-ledger:
	cd ledger; dfx start --background --clean
	cd ledger/testdata; dfx deploy --no-wallet
	cd ledger; DFX=true go test -v ./...; dfx stop

