# PocketIC Golang: A Canister Testing Library

The client is currently implemented for an unreleased version of the PocketIC server.
The client is not yet stable and is subject to change.

You can download the server [here](https://download.dfinity.systems/ic/136a026d67139ecddbc48db3050e488a3c29bb74/binaries/x86_64-linux/pocket-ic.gz).

```go
package actor_test

import (
	"os"
	"testing"

	"github.com/aviate-labs/agent-go/pocketic"
)

func TestActor(t *testing.T) {
	pic, err := pocketic.New(pocketic.DefaultSubnetConfig)
	if err != nil {
		t.Fatal(err)
	}

	wasmModule, err := os.ReadFile("actor.wasm")
	if err != nil {
		t.Fatal(err)
	}

	cID, err := pic.CreateAndInstallCanister(wasmModule, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Call the actor, it has native support for the idl types of the agent-go library.
	var greeting string
	if err := pic.QueryCall(*cID, "hello", nil, []any{&greeting}); err != nil {
		t.Fatal(err)
	}
	_ = greeting
}

```

## References

- [PocketIC](https://github.com/dfinity/pocketic)
- [PocketIC Server](https://github.com/dfinity/ic/tree/master/rs/pocket_ic_server)
