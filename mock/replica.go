package mock

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/certification"
	"github.com/aviate-labs/agent-go/certification/bls"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/principal"

	"github.com/fxamacker/cbor/v2"
)

type Canister struct {
	Id      principal.Principal
	Methods map[string]Method
}

type HandlerFunc func(request Request) ([]any, error)

type Method struct {
	Name      string      // Name is the name of the method.
	Arguments []any       // Arguments is a list of pointers to the arguments, that will be filled by the agent.
	Handler   HandlerFunc // Handler is the function that will be called when the method is called.
}

type Replica struct {
	rootKey   *bls.SecretKey
	Canisters map[string]Canister
	Requests  map[string]agent.Request
}

func NewReplica() *Replica {
	return &Replica{
		rootKey:   bls.NewSecretKeyByCSPRNG(),
		Canisters: make(map[string]Canister),
		Requests:  make(map[string]agent.Request),
	}
}

// AddCanister adds a canister to the replica.
func (r *Replica) AddCanister(
	id principal.Principal,
	methods []Method,
) {
	ms := make(map[string]Method)
	for _, m := range methods {
		ms[m.Name] = m
	}
	r.Canisters[id.String()] = Canister{
		Id:      id,
		Methods: ms,
	}
}

func (r *Replica) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if !strings.HasPrefix(request.URL.Path, "/api/v2/") {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	path := strings.Split(request.URL.Path, "/")[3:]
	switch path[0] {
	case "canister":
		if request.Method != http.MethodPost {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(request.Body)
		r.handleCanister(writer, path[1], path[2], body)
	case "status":
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		r.handleStatus(writer)
	default:
		writer.WriteHeader(http.StatusNotFound)
	}
}

func (r *Replica) handleCanister(writer http.ResponseWriter, canisterId, typ string, body []byte) {
	canister, ok := r.Canisters[canisterId]
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte("canister not found: " + canisterId))
		return
	}
	var envelope agent.Envelope
	if err := cbor.Unmarshal(body, &envelope); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(err.Error()))
		return
	}
	// TODO: validate sender + signatures, ...
	req := envelope.Content

	switch typ {
	case "call":
		if req.Type != agent.RequestTypeCall {
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte("expected call request"))
			return
		}
		requestID := agent.NewRequestID(req)
		requestIDHex := hex.EncodeToString(requestID[:])
		r.Requests[requestIDHex] = req
		writer.WriteHeader(http.StatusAccepted)
	case "query":
		if req.Type != agent.RequestTypeQuery {
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte("expected query request"))
			return
		}

		method, ok := canister.Methods[req.MethodName]
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte("method not defined in replica: " + req.MethodName))
			return
		}

		values, err := method.Handler(fromAgentRequest(req, method.Arguments))
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}

		rawReply, err := idl.Marshal(values)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}

		reply := make(map[string][]byte)
		reply["arg"] = rawReply
		resp := agent.Response{
			Status: "replied",
			Reply:  reply,
		}

		writer.WriteHeader(http.StatusOK)
		raw, err := cbor.Marshal(resp)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}
		_, _ = writer.Write(raw)
	case "read_state":
		if !bytes.Equal(req.Paths[0][0], []byte("request_status")) {
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte("expected request_status"))
			return
		}
		requestID := req.Paths[0][1]
		requestIDHex := hex.EncodeToString(requestID)
		req, ok := r.Requests[requestIDHex]
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte("request not found: " + requestIDHex))
			return
		}

		method, ok := canister.Methods[req.MethodName]
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte("method not defined in replica: " + req.MethodName))
			return
		}

		values, err := method.Handler(fromAgentRequest(req, method.Arguments))
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}

		rawReply, err := idl.Marshal(values)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}

		t := hashtree.NewHashTree(hashtree.Fork{
			LeftTree: hashtree.Labeled{
				Label: []byte("request_status"),
				Tree: hashtree.Labeled{
					Label: requestID,
					Tree: hashtree.Fork{
						LeftTree: hashtree.Labeled{
							Label: []byte("reply"),
							Tree:  hashtree.Leaf(rawReply),
						},
						RightTree: hashtree.Labeled{
							Label: []byte("status"),
							Tree:  hashtree.Leaf("replied"),
						},
					},
				},
			},
			RightTree: hashtree.Empty{},
		})
		d := t.Digest()
		m := make(map[string][]byte)
		s := r.rootKey.Sign(string(append(hashtree.DomainSeparator("ic-state-root"), d[:]...)))
		cert := certification.Cert{
			Tree:      t,
			Signature: s.Serialize(),
		}
		rawCert, _ := cbor.Marshal(cert)
		m["certificate"] = rawCert

		rawTree, _ := cbor.Marshal(t)
		m["tree"] = rawTree

		writer.WriteHeader(http.StatusOK)
		raw, err := cbor.Marshal(m)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}
		_, _ = writer.Write(raw)
	default:
		writer.WriteHeader(http.StatusNotFound)
	}
}

func (r *Replica) handleStatus(writer http.ResponseWriter) {
	publicKey := r.rootKey.GetPublicKey().Serialize()
	status := agent.Status{
		Version: "golang-mock",
		RootKey: publicKey,
	}
	raw, _ := cbor.Marshal(status)
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(raw)
}

type Request struct {
	Type      agent.RequestType
	Sender    principal.Principal
	Arguments []any
}

func fromAgentRequest(request agent.Request, arguments []any) Request {
	if err := idl.Unmarshal(request.Arguments, arguments); err != nil {
		panic(err)
	}
	return Request{
		Type:      request.Type,
		Sender:    request.Sender,
		Arguments: arguments,
	}
}
