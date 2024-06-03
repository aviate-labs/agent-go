package agent

import (
	"crypto/ed25519"
	"fmt"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/certification"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/aviate-labs/leb128"
	"github.com/fxamacker/cbor/v2"
	"google.golang.org/protobuf/proto"
	"math/big"
)

// CreateQuery creates a new Query to the given canister and method.
func (a *Agent) CreateQuery(canisterID principal.Principal, methodName string, args ...any) (*Query, error) {
	rawArgs, err := idl.Marshal(args)
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		// Default to the empty Candid argument list.
		rawArgs = []byte{'D', 'I', 'D', 'L', 0, 0}
	}
	nonce, err := newNonce()
	if err != nil {
		return nil, err
	}
	requestID, data, err := a.sign(Request{
		Type:          RequestTypeQuery,
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
	return &Query{
		a:                   a,
		methodName:          methodName,
		effectiveCanisterID: effectiveCanisterID(canisterID, args),
		requestID:           *requestID,
		data:                data,
	}, nil
}

// Query calls a method on a canister and unmarshals the result into the given values.
func (a Agent) Query(canisterID principal.Principal, methodName string, args, values []any) error {
	query, err := a.CreateQuery(canisterID, methodName, args...)
	if err != nil {
		return err
	}
	return query.Query(values...)
}

// QueryProto calls a method on a canister and unmarshals the result into the given proto message.
func (a Agent) QueryProto(canisterID principal.Principal, methodName string, in, out proto.Message) error {
	payload, err := proto.Marshal(in)
	if err != nil {
		return err
	}
	if len(payload) == 0 {
		payload = []byte{}
	}
	_, data, err := a.sign(Request{
		Type:          RequestTypeQuery,
		Sender:        a.Sender(),
		IngressExpiry: a.expiryDate(),
		CanisterID:    canisterID,
		MethodName:    methodName,
		Arguments:     payload,
	})
	if err != nil {
		return err
	}
	resp, err := a.client.Query(canisterID, data)
	if err != nil {
		return err
	}
	var response Response
	if err := cbor.Unmarshal(resp, &response); err != nil {
		return err
	}
	if response.Status != "replied" {
		return fmt.Errorf("status: %s", response.Status)
	}
	var reply map[string][]byte
	if err := cbor.Unmarshal(response.Reply, &reply); err != nil {
		return err
	}
	return proto.Unmarshal(reply["arg"], out)
}

// Query is an intermediate representation of a Query to a canister.
type Query struct {
	a                   *Agent
	methodName          string
	effectiveCanisterID principal.Principal
	requestID           RequestID
	data                []byte
}

// Query calls a method on a canister and unmarshals the result into the given values.
func (q Query) Query(values ...any) error {
	q.a.logger.Printf("[AGENT] QUERY %s %s", q.effectiveCanisterID, q.methodName)
	rawResp, err := q.a.client.Query(q.effectiveCanisterID, q.data)
	if err != nil {
		return err
	}
	var resp Response
	if err := cbor.Unmarshal(rawResp, &resp); err != nil {
		return err
	}

	// Verify query signatures.
	if q.a.verifySignatures {
		if len(resp.Signatures) == 0 {
			return fmt.Errorf("no signatures")
		}

		for _, signature := range resp.Signatures {
			if len(q.effectiveCanisterID.Raw) == 0 {
				// TODO: temporary fix, signed queries did not take query calls from the management canister into account...
				return nil
			}
			c, err := q.a.readStateCertificate(q.effectiveCanisterID, [][]hashtree.Label{{hashtree.Label("subnet")}})
			if err != nil {
				return err
			}
			if err := certification.VerifyCertificate(*c, q.effectiveCanisterID, q.a.rootKey); err != nil {
				return err
			}
			subnetID := principal.MustDecode(certification.RootSubnetID)
			if c.Delegation != nil {
				subnetID = c.Delegation.SubnetId
			}
			pk, err := c.Tree.Lookup(hashtree.Label("subnet"), subnetID.Raw, hashtree.Label("node"), signature.Identity.Raw, hashtree.Label("public_key"))
			if err != nil {
				return err
			}
			publicKey, err := certification.PublicED25519KeyFromDER(pk)
			if err != nil {
				return err
			}
			switch resp.Status {
			case "replied":
				sig, err := hashOfMap(map[string]any{
					"status":     resp.Status,
					"reply":      resp.Reply,
					"timestamp":  signature.Timestamp,
					"request_id": q.requestID[:],
				})
				if err != nil {
					return err
				}
				if !ed25519.Verify(
					*publicKey,
					append([]byte("\x0Bic-response"), sig[:]...),
					signature.Signature,
				) {
					return fmt.Errorf("invalid signature")
				}
			case "rejected":
				code, err := leb128.EncodeUnsigned(big.NewInt(int64(resp.RejectCode)))
				if err != nil {
					return err
				}
				sig, err := hashOfMap(map[string]any{
					"status":         resp.Status,
					"reject_code":    code,
					"reject_message": resp.RejectMsg,
					"error_code":     resp.ErrorCode,
					"timestamp":      signature.Timestamp,
					"request_id":     q.requestID[:],
				})
				if err != nil {
					return err
				}
				if !ed25519.Verify(
					*publicKey,
					append([]byte("\x0Bic-response"), sig[:]...),
					signature.Signature,
				) {
					return fmt.Errorf("invalid signature")
				}
			default:
				panic("unreachable")
			}
		}
	}
	switch resp.Status {
	case "replied":
		var reply map[string][]byte
		if err := cbor.Unmarshal(resp.Reply, &reply); err != nil {
			return err
		}
		return idl.Unmarshal(reply["arg"], values)
	case "rejected":
		return fmt.Errorf("(%d) %s: %s", resp.RejectCode, resp.ErrorCode, resp.RejectMsg)
	default:
		panic("unreachable")
	}
}
