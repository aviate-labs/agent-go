package registry

import (
	"os"
	"testing"
)

func TestClient_GetNodeListSince(t *testing.T) {
	checkEnabled(t)
	c, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetNodeListSince(0); err != nil {
		t.Fatal(err)
	}
}

func checkEnabled(t *testing.T) {
	// The reason for this is that the tests are very slow.
	if os.Getenv("REGISTRY_TEST_ENABLE") != "true" {
		t.Skip("Skipping registry tests. Set REGISTRY_TEST_ENABLE=true to enable.")
	}
}
