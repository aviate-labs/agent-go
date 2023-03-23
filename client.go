package agent

import (
	"bytes"
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
}

// NewClient creates a new client based on the given configuration.
func NewClient(cfg ClientConfig) Client {
	return Client{
		client: http.Client{},
		config: cfg,
	}
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

func (c Client) call(canisterID principal.Principal, data []byte) ([]byte, error) {
	return c.post("call", canisterID, data, 202)
}

func (c Client) get(path string) ([]byte, error) {
	resp, err := c.client.Get(c.url(path))
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

func (c Client) post(path string, canisterID principal.Principal, data []byte, statusCorePass int) ([]byte, error) {
	u := c.url(fmt.Sprintf("/api/v2/canister/%s/%s", canisterID.Encode(), path))
	resp, err := c.client.Post(u, "application/cbor", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case statusCorePass:
		return io.ReadAll(resp.Body)
	default:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("(%d) %s: %s", resp.StatusCode, resp.Status, body)
	}
}

func (c Client) query(canisterID principal.Principal, data []byte) ([]byte, error) {
	return c.post("query", canisterID, data, 200)
}

func (c Client) readState(canisterID principal.Principal, data []byte) ([]byte, error) {
	return c.post("read_state", canisterID, data, 200)
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
