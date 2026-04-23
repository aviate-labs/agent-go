package registry_test

import (
	"testing"

	"github.com/niccolofant/agent-go"
	"github.com/niccolofant/agent-go/clients/registry"
)

func TestDataProvider_GetLatestVersion(t *testing.T) {
	checkEnabled(t)

	dp := registry.NewDataProvider(&agent.Agent{})
	if _, err := dp.GetLatestVersion(); err != nil {
		t.Error(err)
	}
}
