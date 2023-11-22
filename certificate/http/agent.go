package http

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/certificate"
	"github.com/aviate-labs/agent-go/certificate/http/certexp"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/aviate-labs/leb128"
	"github.com/fxamacker/cbor/v2"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"time"
)

func CalculateRequestHash(r *Request, reqCert *certexp.CertificateExpressionRequestCertification) ([32]byte, error) {
	m := make(map[string]any)
	for _, header := range r.Headers {
		k := strings.ToLower(header.Field0)
		v := header.Field1
		// TODO: allow duplicates?
		m[k] = v
	}
	m[":ic-cert-method"] = r.Method
	// TODO: ic-cert-query
	requestHeaderHash, err := hashMap(m)
	if err != nil {
		return [32]byte{}, err
	}
	requestBodyHash := sha256.Sum256(r.Body)
	return sha256.Sum256(append(requestHeaderHash[:], requestBodyHash[:]...)), nil
}

func CalculateResponseHash(r *Response, respCert certexp.CertificateExpressionResponseCertification) ([32]byte, error) {
	m := make(map[string]any)
	for _, header := range r.Headers {
		k := strings.ToLower(header.Field0)
		if k == "ic-certificate" {
			continue
		}
		v := header.Field1
		// TODO: allow duplicates?
		m[k] = v
	}
	if len(respCert.CertifiedResponseHeaders) > 0 {
		tmp := make(map[string]any)
		for _, header := range respCert.CertifiedResponseHeaders {
			if v, ok := m[strings.ToLower(header)]; ok {
				tmp[strings.ToLower(header)] = v
			}
		}
		m = tmp
	}
	if len(respCert.ResponseHeaderExclusions) > 0 {
		for _, header := range respCert.ResponseHeaderExclusions {
			delete(m, strings.ToLower(header))
		}
	}
	m[":ic-cert-status"] = big.NewInt(int64(r.StatusCode))
	responseHeaderHash, err := hashMap(m)
	if err != nil {
		return [32]byte{}, err
	}
	responseBodyHash := sha256.Sum256(r.Body)
	return sha256.Sum256(append(responseHeaderHash[:], responseBodyHash[:]...)), nil
}

// ref: https://internetcomputer.org/docs/current/references/ic-interface-spec#hash-of-map
func hashMap(m map[string]any) ([32]byte, error) {
	var hashes [][]byte
	for k, v := range m {
		kh := sha256.Sum256([]byte(k))
		var vh [32]byte
		switch v := v.(type) {
		case []byte:
			vh = sha256.Sum256(v)
		case string:
			vh = sha256.Sum256([]byte(v))
		case *big.Int:
			n, err := leb128.EncodeUnsigned(v)
			if err != nil {
				return vh, err
			}
			vh = sha256.Sum256(n)
		case map[string]any:
			h, err := hashMap(v)
			if err != nil {
				return vh, err
			}
			vh = h
		default:
			return vh, fmt.Errorf("invalid value type")
		}
		hashes = append(hashes, append(kh[:], vh[:]...))
	}
	slices.SortFunc(hashes, func(a, b []byte) int { return bytes.Compare(a, b) })
	var concatHashes []byte
	for _, h := range hashes {
		concatHashes = append(concatHashes, h...)
	}
	return sha256.Sum256(concatHashes), nil
}

type Agent struct {
	canisterId          principal.Principal
	supportsV1          bool
	supportsV2, forceV2 bool
	*agent.Agent
}

func NewAgent(canisterId principal.Principal, cfg agent.Config) (*Agent, error) {
	a, err := agent.New(cfg)
	if err != nil {
		return nil, err
	}

	var supportsV1, supportsV2 bool
	if raw, err := a.GetCanisterMetadata(canisterId, "supported_certificate_versions"); err == nil {
		for _, v := range strings.Split(string(raw), ",") {
			switch v {
			case "1":
				supportsV1 = true
			case "2":
				supportsV2 = true
			}
		}
	}

	return &Agent{
		canisterId: canisterId,
		supportsV1: supportsV1,
		supportsV2: supportsV2,
		forceV2:    true,
		Agent:      a,
	}, nil
}

func (a *Agent) EnableLegacyMode() {
	a.forceV2 = false
}

func (a Agent) HttpRequest(request Request) (*Response, error) {
	var r0 Response
	if err := a.Query(
		a.canisterId,
		"http_request",
		[]any{request},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

func (a Agent) HttpRequestStreamingCallback(token StreamingCallbackToken) (**StreamingCallbackHttpResponse, error) {
	var r0 *StreamingCallbackHttpResponse
	if err := a.Query(
		a.canisterId,
		"http_request_streaming_callback",
		[]any{token},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

func (a Agent) VerifyResponse(path string, req *Request, resp *Response) error {
	var certificateHeader *CertificateHeader
	for _, header := range resp.Headers {
		if strings.ToLower(header.Field0) == "ic-certificate" {
			v, err := ParseCertificateHeader(header.Field1)
			if err != nil {
				return err
			}
			certificateHeader = v
		}
	}
	if certificateHeader == nil {
		return fmt.Errorf("no certificate header found")
	}

	// Validate the certificate.
	if err := (certificate.Certificate{
		Cert:       certificateHeader.Certificate,
		RootKey:    a.GetRootKey(),
		CanisterID: a.canisterId,
	}).Verify(); err != nil {
		return err
	}

	// The timestamp at the /time path must be recent, e.g. 5 minutes.
	rawTime := certificate.Lookup(certificate.LookupPath("time"), certificateHeader.Certificate.Tree.Root)
	t, err := leb128.DecodeUnsigned(bytes.NewReader(rawTime))
	if err != nil {
		return err
	}
	age := time.Unix(0, t.Int64())
	minAge := time.Now().Add(-time.Duration(5) * time.Minute)
	maxAge := time.Now().Add(time.Duration(5) * time.Minute)
	if age.Before(minAge) || age.After(maxAge) {
		return fmt.Errorf("certificate is not valid yet or expired")
	}

	switch certificateHeader.Version {
	case 0, 1:
		hash := sha256.Sum256(resp.Body)
		// TODO: take asset streaming into account!
		return a.verifyLegacy(path, hash, certificateHeader)
	case 2:
		return a.verify(req, resp, certificateHeader)
	default:
		return fmt.Errorf("invalid certificate version: %d", certificateHeader.Version)
	}
}

func (a *Agent) verify(req *Request, resp *Response, certificateHeader *CertificateHeader) error {
	exprPath, err := ParseExpressionPath(certificateHeader.ExprPath)
	if err != nil {
		return err
	}

	var certificateExpression string
	for _, header := range resp.Headers {
		if strings.ToLower(header.Field0) == "ic-certificateexpression" {
			certificateExpression = header.Field1
		}
	}
	if certificateExpression == "" {
		return fmt.Errorf("no certification expression found")
	}
	certExpr, err := certexp.ParseCertificateExpression(certificateExpression)
	if err != nil {
		return err
	}

	exprPathNode := certificate.LookupNode(certificate.LookupPath(exprPath.GetPath()...), certificateHeader.Tree.Root)
	if exprPathNode == nil {
		return fmt.Errorf("no expression path found")
	}
	var exprHash certificate.Labeled
	switch n := (*exprPathNode).(type) {
	case certificate.Labeled:
		exprHash = n
		certExprHash := sha256.Sum256([]byte(certificateExpression))
		if !bytes.Equal(exprHash.Label, certExprHash[:]) {
			return fmt.Errorf("invalid expression hash")
		}
	default:
		return fmt.Errorf("invalid expression path")
	}

	if certExpr.Certification == nil {
		return nil
	}

	respHash, err := CalculateResponseHash(resp, certExpr.Certification.ResponseCertification)
	if err != nil {
		return err
	}
	if certExpr.Certification.RequestCertification == nil {
		n := certificate.LookupNode(certificate.LookupPath("", string(respHash[:])), exprHash.Tree)
		if n == nil {
			return fmt.Errorf("response hash not found")
		}
		switch n := (*n).(type) {
		case certificate.Leaf:
			if len(n) != 0 {
				return fmt.Errorf("invalid response hash: not empty")
			}
			return nil
		default:
			return fmt.Errorf("invalid response hash")
		}
	} else {
		reqHash, err := CalculateRequestHash(req, certExpr.Certification.RequestCertification)
		if err != nil {
			return err
		}
		n := certificate.LookupNode(certificate.LookupPath(string(reqHash[:]), string(respHash[:])), exprHash.Tree)
		if n == nil {
			return fmt.Errorf("response hash not found")
		}
		switch n := (*n).(type) {
		case certificate.Leaf:
			if len(n) != 0 {
				return fmt.Errorf("invalid response hash: not empty")
			}
			return nil
		default:
			return fmt.Errorf("invalid response hash")
		}
	}
}

func (a *Agent) verifyLegacy(path string, hash [32]byte, certificateHeader *CertificateHeader) error {
	if a.forceV2 && a.supportsV2 {
		return fmt.Errorf("certificate version 2 is supported")
	}

	witness := certificate.Lookup(certificate.LookupPath("canister", string(a.canisterId.Raw), "certified_data"), certificateHeader.Certificate.Tree.Root)
	if len(witness) != 32 {
		return fmt.Errorf("no witness found")
	}

	reconstruct := certificateHeader.Tree.Root.Reconstruct()
	if !bytes.Equal(witness, reconstruct[:32]) {
		return fmt.Errorf("invalid witness")
	}

	treeHash := certificate.Lookup(certificate.LookupPath("http_assets", path), certificateHeader.Tree.Root)
	if len(treeHash) == 0 {
		treeHash = certificate.Lookup(certificate.LookupPath("http_assets"), certificateHeader.Tree.Root)
	}

	if !bytes.Equal(hash[:], treeHash) {
		return fmt.Errorf("invalid hash")
	}

	return nil
}

type CertificateHeader struct {
	Certificate certificate.Cert
	Tree        certificate.HashTree
	Version     int
	ExprPath    []string
}

func ParseCertificateHeader(header string) (*CertificateHeader, error) {
	var certificateHeader CertificateHeader
	for _, value := range strings.Split(header, ",") {
		vs := strings.SplitN(strings.TrimSpace(value), "=", 2)
		if len(vs) != 2 {
			return nil, fmt.Errorf("invalid header")
		}
		switch v := vs[0]; v {
		case "certificate":
			raw, err := base64.StdEncoding.DecodeString(vs[1][1 : len(vs[1])-1])
			if err != nil {
				return nil, err
			}
			var cert certificate.Cert
			if err := cbor.Unmarshal(raw, &cert); err != nil {
				return nil, err
			}
			certificateHeader.Certificate = cert
		case "tree":
			raw, err := base64.StdEncoding.DecodeString(vs[1][1 : len(vs[1])-1])
			if err != nil {
				return nil, err
			}
			var tree certificate.HashTree
			if err := cbor.Unmarshal(raw, &tree); err != nil {
				return nil, err
			}
			certificateHeader.Tree = tree
		case "version":
			version, err := strconv.Atoi(vs[1])
			if err != nil {
				return nil, err
			}
			certificateHeader.Version = version
		case "expr_path":
			raw, err := base64.StdEncoding.DecodeString(vs[1][1 : len(vs[1])-1])
			if err != nil {
				return nil, err
			}
			var path []string
			if err := cbor.Unmarshal(raw, &path); err != nil {
				return nil, err
			}
			certificateHeader.ExprPath = path
		default:
			return nil, fmt.Errorf("invalid header")
		}
	}
	return &certificateHeader, nil
}

type ExpressionPath struct {
	Wildcard bool
	Path     []string
}

func ParseExpressionPath(path []string) (*ExpressionPath, error) {
	if len(path) < 2 || path[0] != "http_expr" {
		return nil, fmt.Errorf("invalid expression path")
	}
	var wilcard bool
	switch path[len(path)-1] {
	case "<*>":
		wilcard = true
	case "<$>":
	default:
		return nil, fmt.Errorf("invalid expression path")
	}
	return &ExpressionPath{
		Wildcard: wilcard,
		Path:     path[1 : len(path)-1],
	}, nil
}

func (e ExpressionPath) GetPath() []string {
	path := make([]string, len(e.Path)+2)
	copy(path[1:], e.Path)
	path[0] = "http_expr"
	if e.Wildcard {
		path[len(path)-1] = "<*>"
	} else {
		path[len(path)-1] = "<$>"
	}
	return path
}

type HeaderField struct {
	Field0 string `ic:"0"`
	Field1 string `ic:"1"`
}

type Key = string

type Request struct {
	Method             string        `ic:"method"`
	Url                string        `ic:"url"`
	Headers            []HeaderField `ic:"headers"`
	Body               []byte        `ic:"body"`
	CertificateVersion *uint16       `ic:"certificate_version,omitempty"`
}

type Response struct {
	StatusCode        uint16             `ic:"status_code"`
	Headers           []HeaderField      `ic:"headers"`
	Body              []byte             `ic:"body"`
	Upgrade           *bool              `ic:"upgrade,omitempty"`
	StreamingStrategy *StreamingStrategy `ic:"streaming_strategy,omitempty"`
}

type StreamingCallbackHttpResponse struct {
	Body  []byte                  `ic:"body"`
	Token *StreamingCallbackToken `ic:"token,omitempty"`
}

type StreamingCallbackToken struct {
	Key             Key     `ic:"key"`
	ContentEncoding string  `ic:"content_encoding"`
	Index           idl.Nat `ic:"index"`
	Sha256          *[]byte `ic:"sha256,omitempty"`
}

type StreamingStrategy struct {
	Callback *struct {
		Callback struct { /* NOT SUPPORTED */
		} `ic:"callback"`
		Token StreamingCallbackToken `ic:"token"`
	} `ic:"Callback,variant"`
}
