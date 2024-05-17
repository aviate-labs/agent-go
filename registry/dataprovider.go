package registry

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/ic"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/aviate-labs/agent-go/registry/proto/v1"
	"github.com/fxamacker/cbor/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type DataProvider struct {
	sync.RWMutex
	Records []v1.ProtoRegistryRecord
}

func (d *DataProvider) Add(
	key string,
	version uint64,
	value []byte,
) error {
	if version < 1 {
		return fmt.Errorf("version must be greater than 0")
	}

	d.Lock()
	defer d.Unlock()

	var ok bool
	idx := sort.Search(len(d.Records), func(i int) bool {
		if d.Records[i].Key == key {
			if d.Records[i].Version == version {
				ok = true // Record already exists.
			}
			return d.Records[i].Version >= version
		}
		return d.Records[i].Key >= key
	})
	if ok {
		// Key and version already exist.
		return fmt.Errorf("record already exists: %s@%d", key, version)
	}
	d.Records = append(d.Records, v1.ProtoRegistryRecord{})
	copy(d.Records[idx+1:], d.Records[idx:]) // Shift right.
	d.Records[idx] = v1.ProtoRegistryRecord{
		Key:     key,
		Version: version,
		Value:   wrapperspb.Bytes(value),
	}
	return nil
}

func (d *DataProvider) GetChangesSince(version uint64) ([]*v1.RegistryDelta, uint64, error) {
	a, err := agent.New(agent.DefaultConfig)
	if err != nil {
		return nil, 0, err
	}
	payload, err := proto.Marshal(&v1.RegistryGetChangesSinceRequest{
		Version: version,
	})
	if err != nil {
		return nil, 0, err
	}
	request := agent.Request{
		Type:          agent.RequestTypeQuery,
		Sender:        principal.AnonymousID,
		IngressExpiry: uint64(time.Now().Add(5 * time.Minute).UnixNano()),
		CanisterID:    ic.REGISTRY_PRINCIPAL,
		MethodName:    "get_changes_since",
		Arguments:     payload,
	}
	requestID := agent.NewRequestID(request)
	id := new(identity.AnonymousIdentity)
	data, err := cbor.Marshal(agent.Envelope{
		Content:      request,
		SenderPubKey: id.PublicKey(),
		SenderSig:    requestID.Sign(id),
	})
	if err != nil {
		return nil, 0, err
	}
	resp, err := a.Client().Query(ic.REGISTRY_PRINCIPAL, data)
	if err != nil {
		return nil, 0, err
	}
	var response agent.Response
	if err := cbor.Unmarshal(resp, &response); err != nil {
		return nil, 0, err
	}
	if response.Status != "replied" {
		return nil, 0, fmt.Errorf("status: %s", response.Status)
	}

	changesResponse := new(v1.RegistryGetChangesSinceResponse)
	if err := proto.Unmarshal(response.Reply["arg"], changesResponse); err != nil {
		return nil, 0, err
	}
	if changesResponse.Error != nil {
		return nil, 0, fmt.Errorf("error: %s", changesResponse.Error.String())
	}
	return changesResponse.Deltas, changesResponse.Version, nil
}

func (d *DataProvider) IsEmpty() bool {
	d.RLock()
	defer d.RUnlock()

	return len(d.Records) == 0
}
