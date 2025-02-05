package registry_test

import (
	"os"
	"testing"

	"github.com/aviate-labs/agent-go/clients/registry"
	"github.com/aviate-labs/agent-go/principal"
)

func TestClient_GetNNSSubnetID(t *testing.T) {
	checkEnabled(t)

	c, err := registry.New()
	if err != nil {
		t.Fatal(err)
	}

	id, err := c.GetNNSSubnetID()
	if err != nil {
		t.Fatal(err)
	}
	if !id.Equal(principal.MustDecode("tdb26-jop6k-aogll-7ltgs-eruif-6kk7m-qpktf-gdiqx-mxtrf-vb5e6-eqe")) {
		t.Error(id)
	}
}

func TestClient_GetNodeListSince(t *testing.T) {
	checkEnabled(t)

	c, err := registry.New()
	if err != nil {
		t.Fatal(err)
	}

	latestVersion, err := c.GetLatestVersion()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := c.GetNodeListSince(latestVersion - 100); err != nil {
		t.Fatal(err)
	}
}

func checkEnabled(t *testing.T) {
	// The reason for this is that the tests are very slow.
	if os.Getenv("REGISTRY_TEST_ENABLE") != "true" {
		t.Skip("Skipping registry tests. Set REGISTRY_TEST_ENABLE=true to enable.")
	}
}
