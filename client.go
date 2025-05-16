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

// ic0 is the old (default) host for the Internet Computer.
// var ic0, _ = url.Parse("https://ic0.app/")

// icp0 is the default host for the Internet Computer.
var icp0, _ = url.Parse("https://icp0.io/")

// Client is a client for the IC agent.
type Client struct {
	client *http.Client
	host   *url.URL
	logger Logger
}

// NewClient creates a new client based on the given configuration.
func NewClient(options ...ClientOption) Client {
	c := Client{
		client: http.DefaultClient,
		host:   icp0,
		logger: new(NoopLogger),
	}
	for _, o := range options {
		o(&c)
	}
	return c
}

func (c Client) Call(ctx context.Context, canisterID principal.Principal, data []byte) ([]byte, error) {
	u := c.url(fmt.Sprintf("/api/v3/canister/%s/call", canisterID.Encode()))
	c.logger.Printf("[CLIENT] CALL %s", u)
	req, err := c.newRequest(ctx, "POST", u, bytes.NewBuffer(data))
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
		var status struct {
			Status string `cbor:"status"`
		}
		if err := cbor.Unmarshal(body, &status); err != nil {
			return nil, err
		}
		switch status.Status {
		case "replied":
			var certificate struct {
				Certificate []byte `cbor:"certificate"`
			}
			return certificate.Certificate, cbor.Unmarshal(body, &certificate)
		case "non_replicated_rejection":
			var pErr preprocessingError
			if err := cbor.Unmarshal(body, &pErr); err != nil {
				return nil, err
			}
			return nil, pErr
		default:
			return nil, fmt.Errorf("unknown status: %s", status)
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

func (c Client) url(p string) string {
	u := *c.host
	u.Path = path.Join(u.Path, p)
	return u.String()
}

type ClientOption func(c *Client)

func WithHostURL(host *url.URL) ClientOption {
	return func(c *Client) {
		c.host = host
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
