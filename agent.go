package agent

import (
	"fmt"
	"net/url"
	"time"

	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/candid-go"
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

func New(cfg AgentConfig) Agent {
	if cfg.IngressExpiry == 0 {
		cfg.IngressExpiry = 10 * time.Second
	}
	var id identity.Identity = identity.AnonymousIdentity{}
	if cfg.Identity != nil {
		id = *cfg.Identity
	}
	ccfg := ClientConfig{
		Host: ic0,
	}
	if cfg.ClientConfig != nil {
		ccfg = *cfg.ClientConfig
	}
	return Agent{
		client:        NewClient(ccfg),
		identity:      id,
		ingressExpiry: cfg.IngressExpiry,
	}
}

func (a Agent) Query(canisterID principal.Principal, methodName string, args []byte) (string, error) {
	types, values, err := a.QueryCandid(canisterID, methodName, args)
	if err != nil {
		return "", err
	}
	return candid.DecodeValues(types, values)
}

func (a Agent) QueryCandid(canisterID principal.Principal, methodName string, args []byte) ([]idl.Type, []interface{}, error) {
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

func (a Agent) Sender() principal.Principal {
	return a.identity.Sender()
}

func (a Agent) call(canisterID principal.Principal, data []byte) (*QueryResponse, error) {
	resp, err := a.client.call(canisterID, data)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(resp))
	queryReponse := new(QueryResponse)
	return queryReponse, cbor.Unmarshal(resp, &queryReponse)
}

func (a Agent) expiryDate() uint64 {
	return uint64(time.Now().Add(a.ingressExpiry).UnixNano())
}

func (a Agent) query(canisterID principal.Principal, data []byte) (*QueryResponse, error) {
	resp, err := a.client.query(canisterID, data)
	if err != nil {
		return nil, err
	}
	queryReponse := new(QueryResponse)
	return queryReponse, cbor.Unmarshal(resp, &queryReponse)
}

func (a Agent) sign(request Request) (*RequestID, []byte, error) {
	requestID := NewRequestID(request)
	data, err := cbor.Marshal(Envelope{
		Content:      request,
		SenderPubkey: a.identity.PublicKey(),
		SenderSig:    requestID.Sign(a.identity),
	})
	if err != nil {
		return nil, nil, err
	}
	return &requestID, data, nil
}

type AgentConfig struct {
	Identity      *identity.Identity
	IngressExpiry time.Duration
	ClientConfig  *ClientConfig
}
