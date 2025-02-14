package registry_test

import (
	"testing"

	"github.com/aviate-labs/agent-go/clients/registry"
)

func TestDataProvider_GetLatestVersion(t *testing.T) {
	checkEnabled(t)

	dp, err := registry.NewDataProvider()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dp.GetLatestVersion(); err != nil {
		t.Error(err)
	}
}
