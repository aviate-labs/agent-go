package agent

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/aviate-labs/agent-go/candid"
	"github.com/aviate-labs/agent-go/certification"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/fxamacker/cbor/v2"
	"google.golang.org/protobuf/proto"
)

// DefaultConfig is the default configuration for an Agent.
var DefaultConfig = Config{}

func effectiveCanisterID(canisterID principal.Principal, args []any) principal.Principal {
	// If the canisterID is not aaaaa-aa (encoded as empty byte array), return it.
	if 0 < len(canisterID.Raw) || len(args) < 1 {
		return canisterID
	}

	v := reflect.ValueOf(args[0])
	switch v.Kind() {
	case reflect.Map:
		if ecid, ok := args[0].(map[string]any)["canister_id"]; ok {
			switch ecidp := ecid.(type) {
			case principal.Principal:
				return ecidp
			default:
				// If the field is not a principal, return the original canisterId.
				return canisterID
			}
		}
		return canisterID
	case reflect.Struct:
		t := v.Type()
		// Get the field with the ic tag "canister_id".
		for idx := range t.NumField() {
			if tag := t.Field(idx).Tag.Get("ic"); tag == "canister_id" {
				ecid := v.Field(idx).Interface()
				switch ecidp := ecid.(type) {
				case principal.Principal:
					return ecidp
				default:
					// If the field is not a principal, return the original canisterId.
					return canisterID
				}
			}
		}
		return canisterID
	default:
		return canisterID
	}
}

func newNonce() ([]byte, error) {
	/* Read 10 bytes of random data, which is smaller than the max allowed by the IC (32 bytes)
	 * and should still be enough from a practical point of view. */
	nonce := make([]byte, 10)
	_, err := rand.Read(nonce)
	return nonce, err
}

func uint64FromBytes(raw []byte) uint64 {
	switch len(raw) {
	case 1:
		return uint64(raw[0])
	case 2:
		return uint64(binary.BigEndian.Uint16(raw))
	case 4:
		return uint64(binary.BigEndian.Uint32(raw))
	case 8:
		return binary.BigEndian.Uint64(raw)
	default:
		panic(raw)
	}
}

type APIRequest[In, Out any] struct {
	a                   *Agent
	unmarshal           func([]byte, Out) error
	typ                 RequestType
	methodName          string
	effectiveCanisterID principal.Principal
	requestID           RequestID
	data                []byte
}

func createAPIRequest[In, Out any](
	a *Agent,
	marshal func(In) ([]byte, error),
	unmarshal func([]byte, Out) error,
	typ RequestType,
	canisterID principal.Principal,
	effectiveCanisterID principal.Principal,
	methodName string,
	in In,
) (*APIRequest[In, Out], error) {
	rawArgs, err := marshal(in)
	if err != nil {
		return nil, err
	}
	nonce, err := newNonce()
	if err != nil {
		return nil, err
	}
	requestID, data, err := a.sign(Request{
		Type:          typ,
		Sender:        a.Sender(),
		CanisterID:    canisterID,
		MethodName:    methodName,
		Arguments:     rawArgs,
		IngressExpiry: a.expiryDate(),
		Nonce:         nonce,
	})
	if err != nil {
		return nil, err
	}
	return &APIRequest[In, Out]{
		a:                   a,
		unmarshal:           unmarshal,
		typ:                 typ,
		methodName:          methodName,
		effectiveCanisterID: effectiveCanisterID,
		requestID:           *requestID,
		data:                data,
	}, nil
}

// WithEffectiveCanisterID sets the effective canister ID for the Call.
func (c *APIRequest[In, Out]) WithEffectiveCanisterID(canisterID principal.Principal) *APIRequest[In, Out] {
	c.effectiveCanisterID = canisterID
	return c
}

// Agent is a client for the Internet Computer.
type Agent struct {
	client           Client
	ctx              context.Context
	identity         identity.Identity
	ingressExpiry    time.Duration
	rootKey          []byte
	logger           Logger
	delay, timeout   time.Duration
	verifySignatures bool
}

// New returns a new Agent based on the given configuration.
func New(cfg Config) (*Agent, error) {
	if cfg.IngressExpiry == 0 {
		cfg.IngressExpiry = 5 * time.Minute
	}
	// By default, use the anonymous identity.
	var id identity.Identity = new(identity.AnonymousIdentity)
	if cfg.Identity != nil {
		id = cfg.Identity
	}
	client := NewClient(cfg.ClientConfig...)
	rootKey, _ := hex.DecodeString(certification.RootKey)
	if cfg.FetchRootKey {
		status, err := client.Status()
		if err != nil {
			return nil, err
		}
		rootKey = status.RootKey
	}
	delay := time.Second
	if cfg.PollDelay != 0 {
		delay = cfg.PollDelay
	}
	timeout := 10 * time.Second
	if cfg.PollTimeout != 0 {
		timeout = cfg.PollTimeout
	}
	return &Agent{
		client:           client,
		ctx:              context.Background(),
		identity:         id,
		ingressExpiry:    cfg.IngressExpiry,
		rootKey:          rootKey,
		logger:           client.logger,
		delay:            delay,
		timeout:          timeout,
		verifySignatures: !cfg.DisableSignedQueryVerification,
	}, nil
}

// Client returns the underlying Client of the Agent.
func (a Agent) Client() *Client {
	return &a.client
}

// CreateCandidAPIRequest creates a new api request to the given canister and method.
func (a *Agent) CreateCandidAPIRequest(typ RequestType, canisterID principal.Principal, methodName string, args ...any) (*CandidAPIRequest, error) {
	return createAPIRequest(
		a,
		candid.Marshal,
		candid.Unmarshal,
		typ,
		canisterID,
		effectiveCanisterID(canisterID, args),
		methodName,
		args,
	)
}

// CreateProtoAPIRequest creates a new api request to the given canister and method.
func (a *Agent) CreateProtoAPIRequest(typ RequestType, canisterID principal.Principal, methodName string, message proto.Message) (*ProtoAPIRequest, error) {
	return createAPIRequest(
		a,
		func(m proto.Message) ([]byte, error) {
			raw, err := proto.Marshal(m)
			if err != nil {
				return nil, err
			}
			if len(raw) == 0 {
				// Protobuf arg are not allowed to be empty.
				return []byte{}, nil
			}
			return raw, nil
		},
		proto.Unmarshal,
		typ,
		canisterID,
		canisterID,
		methodName,
		message,
	)
}

// GetCanisterControllers returns the list of principals that can control the given canister.
func (a Agent) GetCanisterControllers(canisterID principal.Principal) ([]principal.Principal, error) {
	resp, err := a.GetCanisterInfo(canisterID, "controllers")
	if err != nil {
		return nil, err
	}
	var m [][]byte
	if err := cbor.Unmarshal(resp, &m); err != nil {
		return nil, err
	}
	var p []principal.Principal
	for _, b := range m {
		p = append(p, principal.Principal{Raw: b})
	}
	return p, nil
}

// GetCanisterInfo returns the raw certificate for the given canister based on the given sub-path.
func (a Agent) GetCanisterInfo(canisterID principal.Principal, subPath string) ([]byte, error) {
	path := []hashtree.Label{hashtree.Label("canister"), canisterID.Raw, hashtree.Label(subPath)}
	node, err := a.ReadStateCertificate(canisterID, [][]hashtree.Label{path})
	if err != nil {
		return nil, err
	}
	canisterInfo, err := hashtree.Lookup(node, path...)
	if err != nil {
		return nil, err
	}
	return canisterInfo, nil
}

func (a Agent) GetCanisterMetadata(canisterID principal.Principal, subPath string) ([]byte, error) {
	path := []hashtree.Label{hashtree.Label("canister"), canisterID.Raw, hashtree.Label("metadata"), hashtree.Label(subPath)}
	c, err := a.readStateCertificate(canisterID, [][]hashtree.Label{path})
	if err != nil {
		return nil, err
	}
	metadata, err := c.Tree.Lookup(path...)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

// GetCanisterModuleHash returns the module hash for the given canister.
func (a Agent) GetCanisterModuleHash(canisterID principal.Principal) ([]byte, error) {
	h, err := a.GetCanisterInfo(canisterID, "module_hash")
	var lookupError hashtree.LookupError
	if errors.As(err, &lookupError) && lookupError.Type == hashtree.LookupResultAbsent {
		// If the canister is empty, it is expected that the module hash is not available.
		return nil, nil
	}
	return h, err
}

// GetRootKey returns the root key of the host.
func (a Agent) GetRootKey() []byte {
	return a.rootKey
}

// ReadStateCertificate reads the certificate state of the given canister at the given path.
func (a Agent) ReadStateCertificate(canisterID principal.Principal, path [][]hashtree.Label) (hashtree.Node, error) {
	c, err := a.readStateCertificate(canisterID, path)
	if err != nil {
		return nil, err
	}
	return c.Tree.Root, nil
}

// RequestStatus returns the status of the request with the given ID.
func (a Agent) RequestStatus(ecID principal.Principal, requestID RequestID) ([]byte, hashtree.Node, error) {
	a.logger.Printf("[AGENT] REQUEST STATUS %s %x", ecID, requestID)
	path := []hashtree.Label{hashtree.Label("request_status"), requestID[:]}
	certificate, err := a.readStateCertificate(ecID, [][]hashtree.Label{path})
	if err != nil {
		return nil, nil, err
	}
	if err := certification.VerifyCertificate(*certificate, ecID, a.rootKey); err != nil {
		return nil, nil, err
	}
	status, err := certificate.Tree.Lookup(append(path, hashtree.Label("status"))...)
	var lookupError hashtree.LookupError
	if errors.As(err, &lookupError) && lookupError.Type == hashtree.LookupResultAbsent {
		// The status might not be available immediately, since the request is still being processed.
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	return status, certificate.Tree.Root, nil
}

// Sender returns the principal that is sending the requests.
func (a Agent) Sender() principal.Principal {
	return a.identity.Sender()
}

func (a Agent) call(ecID principal.Principal, data []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(a.ctx, a.ingressExpiry)
	defer cancel()
	return a.client.Call(ctx, ecID, data)
}

func (a Agent) expiryDate() uint64 {
	return uint64(time.Now().Add(a.ingressExpiry).UnixNano())
}

func (a Agent) poll(ecID principal.Principal, requestID RequestID) ([]byte, error) {
	ticker := time.NewTicker(a.delay)
	timer := time.NewTimer(a.timeout)
	for {
		select {
		case <-ticker.C:
			a.logger.Printf("[AGENT] POLL %s %x", ecID, requestID)
			data, node, err := a.RequestStatus(ecID, requestID)
			if err != nil {
				return nil, err
			}
			if len(data) != 0 {
				path := []hashtree.Label{hashtree.Label("request_status"), requestID[:]}
				switch string(data) {
				case "rejected":
					tree := hashtree.NewHashTree(node)
					code, err := tree.Lookup(append(path, hashtree.Label("reject_code"))...)
					if err != nil {
						return nil, err
					}
					message, err := tree.Lookup(append(path, hashtree.Label("reject_message"))...)
					if err != nil {
						return nil, err
					}
					return nil, fmt.Errorf("(%d) %s", uint64FromBytes(code), string(message))
				case "replied":
					replied, err := hashtree.Lookup(node, append(path, hashtree.Label("reply"))...)
					if err != nil {
						return nil, fmt.Errorf("no reply found")
					}
					return replied, nil
				}
			}
		case <-timer.C:
			return nil, fmt.Errorf("out of time... waited %d seconds", a.timeout/time.Second)
		}
	}
}

func (a Agent) readState(ecID principal.Principal, data []byte) (map[string][]byte, error) {
	ctx, cancel := context.WithTimeout(a.ctx, a.ingressExpiry)
	defer cancel()
	resp, err := a.client.ReadState(ctx, ecID, data)
	if err != nil {
		return nil, err
	}
	var m map[string][]byte
	return m, cbor.Unmarshal(resp, &m)
}

func (a Agent) readStateCertificate(ecID principal.Principal, paths [][]hashtree.Label) (*certification.Certificate, error) {
	_, data, err := a.sign(Request{
		Type:          RequestTypeReadState,
		Sender:        a.Sender(),
		Paths:         paths,
		IngressExpiry: a.expiryDate(),
	})
	if err != nil {
		return nil, err
	}
	a.logger.Printf("[AGENT] READ STATE %s (ecID)", ecID)
	resp, err := a.readState(ecID, data)
	if err != nil {
		return nil, err
	}
	var certificate certification.Certificate
	if err := cbor.Unmarshal(resp["certificate"], &certificate); err != nil {
		return nil, err
	}
	if err := certificate.VerifyTime(a.ingressExpiry); err != nil {
		return nil, err
	}
	if err := certification.VerifyCertificate(certificate, ecID, a.rootKey); err != nil {
		return nil, err
	}
	return &certificate, nil
}

func (a Agent) readSubnetState(subnetID principal.Principal, data []byte) (map[string][]byte, error) {
	ctx, cancel := context.WithTimeout(a.ctx, a.ingressExpiry)
	defer cancel()
	resp, err := a.client.ReadSubnetState(ctx, subnetID, data)
	if err != nil {
		return nil, err
	}
	var m map[string][]byte
	return m, cbor.Unmarshal(resp, &m)
}

func (a Agent) readSubnetStateCertificate(subnetID principal.Principal, paths [][]hashtree.Label) (*certification.Certificate, error) {
	_, data, err := a.sign(Request{
		Type:          RequestTypeReadState,
		Sender:        a.Sender(),
		Paths:         paths,
		IngressExpiry: a.expiryDate(),
	})
	if err != nil {
		return nil, err
	}
	a.logger.Printf("[AGENT] READ SUBNET STATE %s (subnetID)", subnetID)
	resp, err := a.readSubnetState(subnetID, data)
	if err != nil {
		return nil, err
	}
	var certificate certification.Certificate
	if err := cbor.Unmarshal(resp["certificate"], &certificate); err != nil {
		return nil, err
	}
	if err := certificate.VerifyTime(a.ingressExpiry); err != nil {
		return nil, err
	}
	if err := certification.VerifySubnetCertificate(certificate, subnetID, a.rootKey); err != nil {
		return nil, err
	}
	return &certificate, nil
}

func (a Agent) sign(request Request) (*RequestID, []byte, error) {
	requestID := NewRequestID(request)
	data, err := cbor.Marshal(Envelope{
		Content:      request,
		SenderPubKey: a.identity.PublicKey(),
		SenderSig:    requestID.Sign(a.identity),
	})
	if err != nil {
		return nil, nil, err
	}
	return &requestID, data, nil
}

type CandidAPIRequest = APIRequest[[]any, []any]

// Config is the configuration for an Agent.
type Config struct {
	// Identity is the identity used by the Agent.
	Identity identity.Identity
	// IngressExpiry is the duration for which an ingress message is valid.
	// The default is set to 5 minutes.
	IngressExpiry time.Duration
	// ClientConfig is the configuration for the underlying Client.
	ClientConfig []ClientOption
	// FetchRootKey determines whether the root key should be fetched from the IC.
	FetchRootKey bool
	// Logger is the logger used by the Agent.
	Logger Logger
	// PollDelay is the delay between polling for a response.
	PollDelay time.Duration
	// PollTimeout is the timeout for polling for a response.
	PollTimeout time.Duration
	// DisableSignedQueryVerification disables the verification of signed queries.
	DisableSignedQueryVerification bool
}

type ProtoAPIRequest = APIRequest[proto.Message, proto.Message]
