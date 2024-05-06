package pocketic

import (
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
	"net/http"
	"time"
)

type InstanceConfig struct {
	InstanceID int                 `json:"instance_id"`
	Topology   map[string]Topology `json:"topology"`
}

// CreateInstance creates a new PocketIC instance.
func (pic PocketIC) CreateInstance(config SubnetConfigSet) (*InstanceConfig, error) {
	var a createResponse[InstanceConfig]
	if err := pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/instances", pic.server.URL()),
		config,
		&a,
	); err != nil {
		return nil, err
	}
	if a.Error != nil {
		return nil, a.Error
	}
	return a.Created, nil
}

// DeleteInstance deletes a PocketIC instance.
func (pic PocketIC) DeleteInstance(instanceID int) error {
	return pic.do(
		http.MethodDelete,
		fmt.Sprintf("%s/instances/%d", pic.server.URL(), instanceID),
		nil,
		nil,
	)
}

// GetCycles returns the cycles of a canister.
func (pic PocketIC) GetCycles(canisterID principal.Principal) (int, error) {
	var cycles rawCycles
	if err := pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/read/get_cycles", pic.InstanceURL()),
		&RawCanisterID{CanisterID: canisterID.Raw},
		&cycles,
	); err != nil {
		return 0, err
	}
	return cycles.Cycles, nil
}

// GetInstances lists all PocketIC instance availabilities.
func (pic PocketIC) GetInstances() ([]string, error) {
	var instances []string
	if err := pic.do(
		http.MethodGet,
		fmt.Sprintf("%s/instances", pic.server.URL()),
		nil,
		&instances,
	); err != nil {
		return nil, err
	}
	return instances, nil
}

// GetStableMemory returns the stable memory of a canister.
func (pic PocketIC) GetStableMemory(canisterID principal.Principal) ([]byte, error) {
	var data rawStableMemory
	if err := pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/read/get_stable_memory", pic.InstanceURL()),
		&RawCanisterID{CanisterID: canisterID.Raw},
		&data,
	); err != nil {
		return nil, err
	}
	return data.Blob, nil
}

// GetSubnet returns the subnet of a canister.
func (pic PocketIC) GetSubnet(canisterID principal.Principal) (*principal.Principal, error) {
	var subnetID SubnetID
	if err := pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/read/get_subnet", pic.InstanceURL()),
		&RawCanisterID{CanisterID: canisterID.Raw},
		&subnetID,
	); err != nil {
		return nil, err
	}
	return &principal.Principal{Raw: subnetID.SubnetID}, nil
}

// GetTime returns the current time of the PocketIC instance.
func (pic PocketIC) GetTime() (*time.Time, error) {
	var t rawTime
	if err := pic.do(
		http.MethodGet,
		fmt.Sprintf("%s/read/get_time", pic.InstanceURL()),
		nil,
		&t,
	); err != nil {
		return nil, err
	}
	now := time.Unix(0, t.NanosSinceEpoch)
	return &now, nil
}

// RootKey returns the root key of the NNS subnet.
func (pic PocketIC) RootKey() ([]byte, error) {
	var subnetID *principal.Principal
	for k, v := range pic.topology {
		if v.SubnetKind == NNSSubnet {
			id, err := principal.Decode(k)
			if err != nil {
				return nil, err
			}
			subnetID = &id
			break
		}
	}
	if subnetID == nil {
		return nil, fmt.Errorf("no NNS subnet found")
	}
	var key []byte
	if err := pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/read/pub_key", pic.InstanceURL()),
		&SubnetID{SubnetID: subnetID.Raw},
		&key,
	); err != nil {
		return nil, err
	}
	return key, nil
}

// SetStableMemory sets the stable memory of a canister.
func (pic PocketIC) SetStableMemory(canisterID principal.Principal, data []byte, gzipCompression bool) error {
	blobID, err := pic.UploadBlob(data, gzipCompression)
	if err != nil {
		return err
	}
	return pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/update/set_stable_memory", pic.InstanceURL()),
		rawSetStableMemory{
			CanisterID: canisterID.Raw,
			BlobID:     blobID,
		},
		nil,
	)
}

func (pic PocketIC) Tick() error {
	return pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/update/tick", pic.InstanceURL()),
		nil,
		nil,
	)
}
