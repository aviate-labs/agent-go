package registry

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/certification"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	v1 "github.com/aviate-labs/agent-go/clients/registry/proto/v1"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/aviate-labs/leb128"
	"github.com/fxamacker/cbor/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var REGISTRY_PRINCIPAL = principal.MustDecode("rwlgt-iiaaa-aaaaa-aaaaa-cai")

type DataProvider struct {
	a *agent.Agent
}

func NewDataProvider() (*DataProvider, error) {
	a, err := agent.New(agent.DefaultConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}
	return &DataProvider{a: a}, nil
}

func (d DataProvider) GetCertifiedChangesSince(version uint64, publicKey []byte) ([]VersionedRecord, uint64, error) {
	var resp v1.CertifiedResponse
	if err := d.a.QueryProto(
		REGISTRY_PRINCIPAL,
		"get_certified_changes_since",
		&v1.RegistryGetChangesSinceRequest{
			Version: version,
		},
		&resp,
	); err != nil {
		return nil, 0, fmt.Errorf("failed to get certified changes: %w", err)
	}
	ht, err := NewHashTree(resp.HashTree)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create hash tree: %w", err)
	}
	rawCurrentVersion, err := ht.Lookup(hashtree.Label("current_version"))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to lookup current version: %w", err)
	}
	currentVersion, err := leb128.DecodeUnsigned(bytes.NewReader(rawCurrentVersion))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode current version: %w", err)
	}

	deltaNodes, err := ht.LookupSubTree(hashtree.Label("delta"))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to lookup delta nodes: %w", err)
	}
	rawDeltas, err := hashtree.AllChildren(deltaNodes)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all children: %w", err)
	}

	var deltas []VersionedRecord
	lastVersion := version
	for _, delta := range rawDeltas {
		var value []byte
		switch delta := delta.Value.(type) {
		case hashtree.Leaf:
			value = delta
		default:
			return nil, 0, fmt.Errorf("unexpected type: %T", delta)
		}
		req := new(v1.RegistryAtomicMutateRequest)
		if err := proto.Unmarshal(value, req); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal atomic mutate request: %w", err)
		}

		v := binary.BigEndian.Uint64(delta.Path[0])
		if v != lastVersion+1 {
			return nil, 0, fmt.Errorf("unexpected version: %d", v)
		}
		lastVersion = v

		for _, m := range req.Mutations {
			var value []byte
			if m.MutationType != v1.RegistryMutation_DELETE {
				value = m.Value
			}
			deltas = append(deltas, VersionedRecord{
				Key:     string(m.Key),
				Version: v,
				Value:   value,
			})
		}
	}

	var certificate certification.Certificate
	if err := cbor.Unmarshal(resp.Certificate, &certificate); err != nil {
		return nil, 0, err
	}
	digest := ht.Digest()
	if err := certification.VerifyCertifiedData(
		certificate,
		REGISTRY_PRINCIPAL,
		publicKey,
		digest[:],
	); err != nil {
		return nil, 0, fmt.Errorf("failed to verify certified data: %w", err)
	}

	return deltas, currentVersion.Uint64(), nil
}

// GetChangesSince returns the changes since the given version.
func (d DataProvider) GetChangesSince(version uint64) ([]*v1.RegistryDelta, uint64, error) {
	var resp v1.RegistryGetChangesSinceResponse
	if err := d.a.QueryProto(
		REGISTRY_PRINCIPAL,
		"get_changes_since",
		&v1.RegistryGetChangesSinceRequest{
			Version: version,
		},
		&resp,
	); err != nil {
		return nil, 0, fmt.Errorf("failed to get changes since: %w", err)
	}
	if resp.Error != nil {
		return nil, 0, fmt.Errorf("error: %s", resp.Error.String())
	}
	return resp.Deltas, resp.Version, nil
}

func (d DataProvider) GetLatestVersion() (uint64, error) {
	var resp v1.RegistryGetLatestVersionResponse
	if err := d.a.QueryProto(
		REGISTRY_PRINCIPAL,
		"get_latest_version",
		nil,
		&resp,
	); err != nil {
		return 0, fmt.Errorf("failed to get latest version: %w", err)
	}
	return resp.Version, nil
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
		REGISTRY_PRINCIPAL,
		"get_value",
		&v1.RegistryGetValueRequest{
			Key:     key,
			Version: v,
		},
		&resp,
	); err != nil {
		return nil, 0, fmt.Errorf("failed to get value: %w", err)
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
		REGISTRY_PRINCIPAL,
		"get_value",
		&v1.RegistryGetValueRequest{
			Key:     key,
			Version: v,
		},
		&resp,
	); err != nil {
		return nil, 0, fmt.Errorf("failed to get value: %w", err)
	}
	if resp.Error != nil {
		return nil, 0, fmt.Errorf("error: %s", resp.Error.String())
	}
	return resp.Value, resp.Version, nil
}

type VersionedRecord struct {
	Key     string
	Version uint64
	Value   []byte
}
