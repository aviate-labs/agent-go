package agent

import (
	"fmt"
	"net/url"
	"time"

	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/candid-go/idl"
	"github.com/aviate-labs/principal-go"
	"github.com/fxamacker/cbor/v2"
)

var ic0, _ = url.Parse("https://ic0.app/")

type Agent struct {
	client        Client
	identity      identity.Identity
	ingressExpiry time.Duration
}

func New() Agent {
	return Agent{
		client: NewClient(ClientConfig{
			Host: ic0,
		}),
		identity:      identity.AnonymousIdentity{},
		ingressExpiry: 10 * time.Second,
	}
}

func (a Agent) expiryDate() uint64 {
	return uint64(time.Now().Add(a.ingressExpiry).UnixNano())
}

func (a Agent) Sender() principal.Principal {
	return a.identity.Sender()
}

func (a Agent) query(canisterID principal.Principal, data []byte) (*QueryResponse, error) {
	resp, err := a.client.query(canisterID, data)
	if err != nil {
		return nil, err
	}
	queryReponse := new(QueryResponse)
	return queryReponse, cbor.Unmarshal(resp, &queryReponse)
}

func (a Agent) Query(canisterID principal.Principal, methodName string, args []byte) ([]idl.Type, []interface{}, error) {
	_, data, err := a.sign(Request{
		Type:          RequestTypeQuery,
		Sender:        a.Sender(),
		CanisterID:    canisterID,
		MethodName:    methodName,
		Arguments:     args,
		IngressExpiry: a.expiryDate(),
	})
	if err != nil {
		return nil, nil, err
	}
	resp, err := a.query(canisterID, data)
	if err != nil {
		return nil, nil, err
	}
	switch resp.Status {
	case "replied":
		return idl.Decode(resp.Reply["arg"])
	case "rejected":
		return nil, nil, fmt.Errorf("(%d) %s", resp.RejectCode, resp.RejectMsg)
	default:
		panic("unreachable")
	}
}

func (a Agent) sign(request Request) (*RequestID, []byte, error) {
	requestID := NewRequestID(request)
	data, err := cbor.Marshal(struct {
		Content         Request `cbor:"content"`
		SenderPublicKey []byte  `cbor:"sender_pubkey,omitempty"`
		Signature       []byte  `cbor:"sender_sig,omitempty"`
	}{
		Content:         request,
		SenderPublicKey: a.identity.PublicKey(),
		Signature:       requestID.Sign(a.identity),
	})
	if err != nil {
		return nil, nil, err
	}
	return &requestID, data, nil
}
