package http

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/aviate-labs/agent-go/certification"

	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/certification/http/certexp"
	"github.com/aviate-labs/leb128"

	"github.com/fxamacker/cbor/v2"
)

func CalculateRequestHash(r *Request, reqCert *certexp.CertificateExpressionRequestCertification) ([32]byte, error) {
	m := make(map[string]any)
	for _, header := range r.Headers {
		k := strings.ToLower(header.Field0)
		v := header.Field1

		if slices.Contains(reqCert.CertifiedRequestHeaders, k) {
			// TODO: allow duplicates?
			m[k] = v
		}
	}
	m[":ic-cert-method"] = r.Method
	u, err := url.Parse(r.Url)
	if err != nil {
		return [32]byte{}, err
	}
	params, err := parseQueryInOrder(u.RawQuery)
	if err != nil {
		return [32]byte{}, err
	}
	if len(params) > 0 {
		var certifiedParams []queryParameter
		for _, certifiedParam := range reqCert.CertifiedQueryParameters {
			for _, param := range params {
				if param.k == certifiedParam {
					certifiedParams = append(certifiedParams, param)
				}
			}
		}
		params = certifiedParams
	}
	if len(params) > 0 {
		var query string
		for i, param := range params {
			if i != 0 {
				query += "&"
			}
			query += fmt.Sprintf("%s=%s", url.QueryEscape(param.k), url.QueryEscape(param.v))
		}
		m[":ic-cert-query"] = sha256.Sum256([]byte(query))
	}

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
func hashAny(a any) ([32]byte, error) {
	switch a := a.(type) {
	case []byte:
		return sha256.Sum256(a), nil
	case [32]byte:
		return a, nil
	case string:
		return sha256.Sum256([]byte(a)), nil
	case *big.Int:
		n, err := leb128.EncodeUnsigned(a)
		if err != nil {
			return [32]byte{}, err
		}
		return sha256.Sum256(n), nil
	case []any:
		var hashes []byte
		for _, v := range a {
			h, err := hashAny(v)
			if err != nil {
				return [32]byte{}, err
			}
			hashes = append(hashes, h[:]...)
		}
		return sha256.Sum256(hashes), nil
	case map[string]any:
		return hashMap(a)
	default:
		return [32]byte{}, fmt.Errorf("invalid value type: %T", a)
	}
}

func hashMap(m map[string]any) ([32]byte, error) {
	var hashes [][]byte
	for k, v := range m {
		kh := sha256.Sum256([]byte(k))
		vh, err := hashAny(v)
		if err != nil {
			return [32]byte{}, err
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
	if err := certification.VerifyCertificate(
		certificateHeader.Certificate,
		a.canisterId,
		a.GetRootKey(),
	); err != nil {
		return err
	}

	// The timestamp at the /time path must be recent, e.g. 5 minutes.
	rawTime, err := certificateHeader.Certificate.Tree.Lookup(hashtree.Label("time"))
	if err != nil {
		return fmt.Errorf("no time found: %w", err)
	}
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
	fmt.Println(certificateExpression)
	certExpr, err := certexp.ParseCertificateExpression(certificateExpression)
	if err != nil {
		return err
	}

	exprPathNode, err := certificateHeader.Tree.LookupSubTree(exprPath.GetPath()...)
	if err != nil {
		return fmt.Errorf("no expression path found: %w", err)
	}
	var exprHash hashtree.Labeled
	switch n := (exprPathNode).(type) {
	case hashtree.Labeled:
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
		n, err := hashtree.NewHashTree(exprHash.Tree).LookupSubTree(hashtree.Label(""), respHash[:])
		if err != nil {
			return fmt.Errorf("response hash not found: %w", err)
		}
		switch n := (n).(type) {
		case hashtree.Leaf:
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
		n, err := hashtree.NewHashTree(exprHash.Tree).LookupSubTree(reqHash[:], respHash[:])
		if err != nil {
			return fmt.Errorf("response hash not found: %w", err)
		}
		switch n := (n).(type) {
		case hashtree.Leaf:
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

	witness, err := certificateHeader.Certificate.Tree.Lookup(hashtree.Label("canister"), a.canisterId.Raw, hashtree.Label("certified_data"))
	if err != nil {
		return fmt.Errorf("no witness found: %w", err)
	}

	if len(witness) != 32 {
		return fmt.Errorf("invalid witness length")
	}

	reconstruct := certificateHeader.Tree.Root.Reconstruct()
	if !bytes.Equal(witness, reconstruct[:32]) {
		return fmt.Errorf("invalid witness")
	}

	treeHash, err := certificateHeader.Tree.Lookup(hashtree.Label("http_assets"), hashtree.Label(path))
	if err != nil || len(treeHash) == 0 {
		treeHash, _ = certificateHeader.Tree.Lookup(hashtree.Label("http_assets"))
	}

	if !bytes.Equal(hash[:], treeHash) {
		return fmt.Errorf("invalid hash")
	}
	return nil
}

type CertificateHeader struct {
	Certificate certification.Certificate
	Tree        hashtree.HashTree
	Version     int
	ExprPath    []hashtree.Label
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
			var cert certification.Certificate
			if err := cbor.Unmarshal(raw, &cert); err != nil {
				return nil, err
			}
			certificateHeader.Certificate = cert
		case "tree":
			raw, err := base64.StdEncoding.DecodeString(vs[1][1 : len(vs[1])-1])
			if err != nil {
				return nil, err
			}
			var tree hashtree.HashTree
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
			var strPath []string
			if err := cbor.Unmarshal(raw, &strPath); err != nil {
				return nil, err
			}
			var path []hashtree.Label
			for _, s := range strPath {
				path = append(path, hashtree.Label(s))
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
	Path     []hashtree.Label
}

func ParseExpressionPath(path []hashtree.Label) (*ExpressionPath, error) {
	if len(path) < 2 || !bytes.Equal(path[0], hashtree.Label("http_expr")) {
		return nil, fmt.Errorf("invalid expression path")
	}
	var wilcard bool
	switch string(path[len(path)-1]) {
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

func (e ExpressionPath) GetPath() []hashtree.Label {
	path := make([]hashtree.Label, len(e.Path)+2)
	copy(path[1:], e.Path)
	path[0] = hashtree.Label("http_expr")
	if e.Wildcard {
		path[len(path)-1] = hashtree.Label("<*>")
	} else {
		path[len(path)-1] = hashtree.Label("<$>")
	}
	return path
}

type queryParameter struct {
	k, v string
}

func parseQueryInOrder(query string) (params []queryParameter, err error) {
	for query != "" {
		var key string
		key, query, _ = strings.Cut(query, "&")
		if strings.Contains(key, ";") {
			err = fmt.Errorf("invalid semicolon separator in query")
			continue
		}
		if key == "" {
			continue
		}
		key, value, _ := strings.Cut(key, "=")
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		params = append(params, queryParameter{k: key, v: value})
	}
	return params, nil
}
