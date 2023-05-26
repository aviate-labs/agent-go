package mock

import (
	"bytes"
	"encoding/hex"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid/marshal"
	"github.com/aviate-labs/agent-go/certificate"
	"github.com/aviate-labs/agent-go/certificate/bls"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/fxamacker/cbor/v2"
	"io"
	"log"
	"net/http"
	"strings"
)

type Canister struct {
	Id      principal.Principal
	Handler HandlerFunc
}

type HandlerFunc func(request Request) ([]any, error)

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
func (r *Replica) AddCanister(id principal.Principal, handler HandlerFunc) {
	r.Canisters[id.String()] = Canister{
		Id:      id,
		Handler: handler,
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
		requestId := agent.NewRequestID(req)
		requestIdHex := hex.EncodeToString(requestId[:])
		log.Println("received call request", requestIdHex)
		r.Requests[requestIdHex] = req
		writer.WriteHeader(http.StatusAccepted)
	case "query":
		if req.Type != agent.RequestTypeQuery {
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte("expected query request"))
			return
		}
		requestId := agent.NewRequestID(req)
		requestIdHex := hex.EncodeToString(requestId[:])
		log.Println("received query request", requestIdHex)

		values, err := canister.Handler(fromAgentRequest(req))
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}

		rawReply, err := marshal.Marshal(values)
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
		requestId := req.Paths[0][1]
		requestIdHex := hex.EncodeToString(requestId)
		log.Println("received read_state request", requestIdHex)
		req, ok := r.Requests[requestIdHex]
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte("request not found: " + requestIdHex))
			return
		}
		values, err := canister.Handler(fromAgentRequest(req))
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}

		rawReply, err := marshal.Marshal(values)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}

		t := certificate.NewHashTree(certificate.Fork{
			LeftTree: certificate.Labeled{
				Label: []byte("request_status"),
				Tree: certificate.Labeled{
					Label: requestId,
					Tree: certificate.Fork{
						LeftTree: certificate.Labeled{
							Label: []byte("status"),
							Tree:  certificate.Leaf("replied"),
						},
						RightTree: certificate.Labeled{
							Label: []byte("reply"),
							Tree:  certificate.Leaf(rawReply),
						},
					},
				},
			},
			RightTree: certificate.Empty{},
		})
		d := t.Digest()
		m := make(map[string][]byte)
		s := r.rootKey.Sign(string(append(certificate.DomainSeparator("ic-state-root"), d[:]...)))
		cert := certificate.Cert{
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
	log.Println("getting status")
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
	Type       agent.RequestType
	Sender     principal.Principal
	MethodName string
	Arguments  []any
}

func fromAgentRequest(request agent.Request) Request {
	var arguments []any
	_ = marshal.Unmarshal(request.Arguments, arguments)
	return Request{
		Type:       request.Type,
		Sender:     request.Sender,
		MethodName: request.MethodName,
		Arguments:  arguments,
	}
}
