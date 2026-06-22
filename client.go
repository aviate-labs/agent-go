package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/fxamacker/cbor/v2"
	"github.com/niccolofant/agent-go/principal"
)

// ic0 is the old (default) host for the Internet Computer.
// var ic0, _ = url.Parse("https://ic0.app/")

// icp0 is the default host for the Internet Computer.
var icp0, _ = url.Parse("https://icp0.io/")

// Client is a client for the IC agent.
type Client struct {
	client *http.Client
	routes RouteProvider
	logger Logger
	// callVersion / readStateVersion select the API version segment of the
	// call and read_state endpoints. The defaults certify canister ranges
	// under the sharded /canister_ranges/<subnet_id> layout; legacy uses the
	// deprecated /subnet/<subnet_id>/canister_ranges layout.
	callVersion      string
	readStateVersion string
}

// NewClient creates a new client based on the given configuration.
func NewClient(options ...ClientOption) Client {
	c := Client{
		client:           http.DefaultClient,
		routes:           StaticRoute(icp0),
		logger:           new(NoopLogger),
		callVersion:      "v4",
		readStateVersion: "v3",
	}
	for _, o := range options {
		o(&c)
	}
	return c
}

func (c Client) Call(ctx context.Context, canisterID principal.Principal, data []byte) ([]byte, error) {
	u, err := c.url(fmt.Sprintf("/api/%s/canister/%s/call", c.callVersion, canisterID.Encode()))
	if err != nil {
		return nil, err
	}
	c.logger.Printf("[CLIENT] CALL %s", u)
	req, err := c.newRequest(ctx, "POST", u, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	switch resp.StatusCode {
	case http.StatusAccepted:
		return nil, nil
	case http.StatusOK:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var reply struct {
			Status      string `cbor:"status"`
			Certificate []byte `cbor:"certificate"`
			RejectCode  uint64 `cbor:"reject_code"`
			Message     string `cbor:"reject_message"`
			ErrorCode   string `cbor:"error_code"`
		}
		if err := cbor.Unmarshal(body, &reply); err != nil {
			return nil, err
		}
		switch reply.Status {
		case "replied":
			return reply.Certificate, nil
		case "non_replicated_rejection":
			return nil, preprocessingError{
				RejectCode: reply.RejectCode,
				Message:    reply.Message,
				ErrorCode:  reply.ErrorCode,
			}
		default:
			return nil, fmt.Errorf("unknown status: %s", reply.Status)
		}
	default:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("(%d) %s: %s", resp.StatusCode, resp.Status, body)
	}
}

func (c Client) Query(ctx context.Context, canisterID principal.Principal, data []byte) ([]byte, error) {
	return c.post(ctx, "v2", "query", canisterID, data)
}

func (c Client) ReadState(ctx context.Context, canisterID principal.Principal, data []byte) ([]byte, error) {
	return c.post(ctx, c.readStateVersion, "read_state", canisterID, data)
}

func (c Client) ReadSubnetState(ctx context.Context, subnetID principal.Principal, data []byte) ([]byte, error) {
	return c.postSubnet(ctx, "read_state", subnetID, data)
}

// SetRouteProvider replaces the route provider used to pick a host URL for each
// outgoing request. Intended for runtime boundary-node selection (e.g. via
// DiscoverRoutes + RoundRobinRoute); not safe to call concurrently with
// in-flight requests.
func (c *Client) SetRouteProvider(rp RouteProvider) {
	c.routes = rp
}

// Status returns the status of the IC.
func (c Client) Status() (*Status, error) {
	raw, err := c.get("/api/v2/status")
	if err != nil {
		return nil, err
	}
	var status Status
	return &status, cbor.Unmarshal(raw, &status)
}

func (c Client) get(path string) ([]byte, error) {
	u, err := c.url(path)
	if err != nil {
		return nil, err
	}
	c.logger.Printf("[CLIENT] GET %s", u)
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return io.ReadAll(resp.Body)
}

func (c Client) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/cbor")
	return req, nil
}

func (c Client) post(ctx context.Context, version, path string, canisterID principal.Principal, data []byte) ([]byte, error) {
	u, err := c.url(fmt.Sprintf("/api/%s/canister/%s/%s", version, canisterID.Encode(), path))
	if err != nil {
		return nil, err
	}
	c.logger.Printf("[CLIENT] POST %s", u)
	req, err := c.newRequest(ctx, "POST", u, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	switch resp.StatusCode {
	case http.StatusOK:
		return io.ReadAll(resp.Body)
	default:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("(%d) %s: %s", resp.StatusCode, resp.Status, body)
	}
}

func (c Client) postSubnet(ctx context.Context, path string, subnetID principal.Principal, data []byte) ([]byte, error) {
	u, err := c.url(fmt.Sprintf("/api/v2/subnet/%s/%s", subnetID.Encode(), path))
	if err != nil {
		return nil, err
	}
	c.logger.Printf("[CLIENT] POST %s", u)
	req, err := c.newRequest(ctx, "POST", u, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	switch resp.StatusCode {
	case http.StatusOK:
		return io.ReadAll(resp.Body)
	default:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("(%d) %s: %s", resp.StatusCode, resp.Status, body)
	}
}

func (c Client) url(p string) (string, error) {
	host, err := c.routes.Route()
	if err != nil {
		return "", fmt.Errorf("route: %w", err)
	}
	u := *host
	u.Path = path.Join(u.Path, p)
	return u.String(), nil
}

type ClientOption func(c *Client)

func WithHostURL(host *url.URL) ClientOption {
	return func(c *Client) {
		c.routes = StaticRoute(host)
	}
}

func WithHttpClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.client = client
	}
}

func WithLogger(logger Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithLegacyAPI uses the deprecated /api/v3 call and /api/v2 read_state
// endpoints instead of the defaults (/api/v4 call, /api/v3 read_state).
func WithLegacyAPI() ClientOption {
	return func(c *Client) {
		c.callVersion = "v3"
		c.readStateVersion = "v2"
	}
}

type preprocessingError struct {
	// The reject code.
	RejectCode uint64 `cbor:"reject_code"`
	// A textual diagnostic message.
	Message string `cbor:"reject_message"`
	// An optional implementation-specific textual error code.
	ErrorCode string `cbor:"error_code"`
}

func (e preprocessingError) Error() string {
	return fmt.Sprintf("(%d) %s: %s", e.RejectCode, e.Message, e.ErrorCode)
}
