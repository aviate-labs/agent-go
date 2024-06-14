package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/aviate-labs/agent-go/principal"
	"github.com/fxamacker/cbor/v2"
)

// Client is a client for the IC agent.
type Client struct {
	client http.Client
	config ClientConfig
	logger Logger
}

// NewClient creates a new client based on the given configuration.
func NewClient(cfg ClientConfig) Client {
	return Client{
		client: http.Client{},
		config: cfg,
		logger: new(NoopLogger),
	}
}

// NewClientWithLogger creates a new client based on the given configuration and logger.
func NewClientWithLogger(cfg ClientConfig, logger Logger) Client {
	if logger == nil {
		logger = new(NoopLogger)
	}
	return Client{
		client: http.Client{},
		config: cfg,
		logger: logger,
	}
}

func (c Client) Call(ctx context.Context, canisterID principal.Principal, data []byte) ([]byte, error) {
	u := c.url(fmt.Sprintf("/api/v2/canister/%s/call", canisterID.Encode()))
	c.logger.Printf("[CLIENT] CALL %s", u)
	req, err := c.newRequest(ctx, "POST", u, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusAccepted:
		return io.ReadAll(resp.Body)
	case http.StatusOK:
		body, _ := io.ReadAll(resp.Body)
		var err preprocessingError
		if err := cbor.Unmarshal(body, &err); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("(%d) %s: %s", err.RejectCode, err.Message, err.ErrorCode)
	default:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("(%d) %s: %s", resp.StatusCode, resp.Status, body)
	}
}

func (c Client) Query(ctx context.Context, canisterID principal.Principal, data []byte) ([]byte, error) {
	return c.post(ctx, "query", canisterID, data)
}

func (c Client) ReadState(ctx context.Context, canisterID principal.Principal, data []byte) ([]byte, error) {
	return c.post(ctx, "read_state", canisterID, data)
}

func (c Client) ReadSubnetState(ctx context.Context, subnetID principal.Principal, data []byte) ([]byte, error) {
	return c.postSubnet(ctx, "read_state", subnetID, data)
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
	c.logger.Printf("[CLIENT] GET %s", c.url(path))
	resp, err := c.client.Get(c.url(path))
	if err != nil {
		return nil, err
	}
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

func (c Client) post(ctx context.Context, path string, canisterID principal.Principal, data []byte) ([]byte, error) {
	u := c.url(fmt.Sprintf("/api/v2/canister/%s/%s", canisterID.Encode(), path))
	c.logger.Printf("[CLIENT] POST %s", u)
	req, err := c.newRequest(ctx, "POST", u, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		return io.ReadAll(resp.Body)
	default:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("(%d) %s: %s", resp.StatusCode, resp.Status, body)
	}
}

func (c Client) postSubnet(ctx context.Context, path string, subnetID principal.Principal, data []byte) ([]byte, error) {
	u := c.url(fmt.Sprintf("/api/v2/subnet/%s/%s", subnetID.Encode(), path))
	c.logger.Printf("[CLIENT] POST %s", u)
	req, err := c.newRequest(ctx, "POST", u, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		return io.ReadAll(resp.Body)
	default:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("(%d) %s: %s", resp.StatusCode, resp.Status, body)
	}
}

func (c Client) url(p string) string {
	u := *c.config.Host
	u.Path = path.Join(u.Path, p)
	return u.String()
}

// ClientConfig is the configuration for a client.
type ClientConfig struct {
	Host *url.URL
}

type preprocessingError struct {
	// The reject code.
	RejectCode uint64 `cbor:"reject_code"`
	// A textual diagnostic message.
	Message string `cbor:"reject_message"`
	// An optional implementation-specific textual error code.
	ErrorCode string `cbor:"error_code"`
}
