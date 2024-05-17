package registry

import (
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/ic"
	"github.com/aviate-labs/agent-go/registry/proto/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type DataProvider struct {
	a *agent.Agent
}

func NewDataProvider() (*DataProvider, error) {
	a, err := agent.New(agent.DefaultConfig)
	if err != nil {
		return nil, err
	}
	return &DataProvider{a: a}, nil
}

// GetChangesSince returns the changes since the given version.
func (d DataProvider) GetChangesSince(version uint64) ([]*v1.RegistryDelta, uint64, error) {
	var resp v1.RegistryGetChangesSinceResponse
	if err := d.a.QueryProto(
		ic.REGISTRY_PRINCIPAL,
		"get_changes_since",
		&v1.RegistryGetChangesSinceRequest{
			Version: version,
		},
		&resp,
	); err != nil {
		return nil, 0, err
	}
	if resp.Error != nil {
		return nil, 0, fmt.Errorf("error: %s", resp.Error.String())
	}
	return resp.Deltas, resp.Version, nil
}

// GetValue returns the value of the given key and its version.
// If version is nil, the latest version is returned.
func (d DataProvider) GetValue(key []byte, version *uint64) ([]byte, uint64, error) {
	var v *wrapperspb.UInt64Value
	if version != nil {
		v = wrapperspb.UInt64(*version)
	}
	var resp v1.RegistryGetValueResponse
	if err := d.a.QueryProto(
		ic.REGISTRY_PRINCIPAL,
		"get_value",
		&v1.RegistryGetValueRequest{
			Key:     key,
			Version: v,
		},
		&resp,
	); err != nil {
		return nil, 0, err
	}
	if resp.Error != nil {
		return nil, 0, fmt.Errorf("error: %s", resp.Error.String())
	}
	return resp.Value, resp.Version, nil
}

// GetValueUpdate returns the value of the given key and its version.
func (d DataProvider) GetValueUpdate(key []byte, version *uint64) ([]byte, uint64, error) {
	var v *wrapperspb.UInt64Value
	if version != nil {
		v = wrapperspb.UInt64(*version)
	}
	var resp v1.RegistryGetValueResponse
	if err := d.a.CallProto(
		ic.REGISTRY_PRINCIPAL,
		"get_value",
		&v1.RegistryGetValueRequest{
			Key:     key,
			Version: v,
		},
		&resp,
	); err != nil {
		return nil, 0, err
	}
	if resp.Error != nil {
		return nil, 0, fmt.Errorf("error: %s", resp.Error.String())
	}
	return resp.Value, resp.Version, nil
}
