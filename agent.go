package agent

import (
	"fmt"
	"net/url"
	"time"

	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/candid-go"
	"github.com/aviate-labs/candid-go/idl"
	cert "github.com/aviate-labs/certificate-go"
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

func (a Agent) GetCanisterControllers(canisterID principal.Principal) ([]principal.Principal, error) {
	resp, err := a.GetCanisterInfo(canisterID, "controllers")
	if err != nil {
		return nil, err
	}
	var m []principal.Principal
	if err := cbor.Unmarshal(resp, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (a Agent) GetCanisterInfo(canisterID principal.Principal, subPath string) ([]byte, error) {
	path := [][]byte{[]byte("canister"), canisterID, []byte(subPath)}
	c, err := a.readStateCertificate(canisterID, [][][]byte{path})
	if err != nil {
		return nil, err
	}
	var state map[string]interface{}
	if err := cbor.Unmarshal(c, &state); err != nil {
		return nil, err
	}
	node, err := cert.DeserializeNode(state["tree"].([]interface{}))
	if err != nil {
		return nil, err
	}
	return cert.Lookup(path, node), nil
}

func (a Agent) GetCanisterModuleHash(canisterID principal.Principal) ([]byte, error) {
	return a.GetCanisterInfo(canisterID, "module_hash")
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

func (a Agent) readState(canisterID principal.Principal, data []byte) (map[string][]byte, error) {
	resp, err := a.client.readState(canisterID, data)
	if err != nil {
		return nil, err
	}
	var m map[string][]byte
	return m, cbor.Unmarshal(resp, &m)
}

func (a Agent) readStateCertificate(canisterID principal.Principal, paths [][][]byte) ([]byte, error) {
	_, data, err := a.sign(Request{
		Type:          RequestTypeReadState,
		Sender:        a.Sender(),
		Paths:         paths,
		IngressExpiry: a.expiryDate(),
	})
	if err != nil {
		return nil, err
	}

	resp, err := a.readState(canisterID, data)
	if err != nil {
		return nil, err
	}
	return resp["certificate"], nil
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
