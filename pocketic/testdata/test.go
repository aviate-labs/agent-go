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

	// Call the actor
	var greeting string
	if err := pic.QueryCall(*cID, "hello", nil, []any{&greeting}); err != nil {
		t.Fatal(err)
	}

	_ = greeting
}
