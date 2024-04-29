package pocketic_test

import (
	"fmt"
	"github.com/aviate-labs/agent-go/pocketic"
	"testing"
)

func TestPocketIC_CreateCanister(t *testing.T) {
	pic, err := pocketic.New(pocketic.WithLogger(new(testLogger)))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := pic.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	canisterID, err := pic.CreateCanister()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pic.AddCycles(*canisterID, 2_000_000_000_000); err != nil {
		t.Fatal(err)
	}
}

type testLogger struct{}

func (t testLogger) Printf(format string, v ...any) {
	fmt.Printf(format+"\n", v...)
}
