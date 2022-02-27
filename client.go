package agent

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/aviate-labs/principal-go"
	"github.com/fxamacker/cbor/v2"
)

type Client struct {
	client http.Client
	config ClientConfig
}

func NewClient(cfg ClientConfig) Client {
	return Client{
		client: http.Client{},
		config: cfg,
	}
}

func (c Client) Status() (Status, error) {
	raw, err := c.get("/api/v2/status")
	if err != nil {
		return Status{}, err
	}
	var status Status
	return status, cbor.Unmarshal(raw, &status)
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
	url := c.url(fmt.Sprintf("/api/v2/canister/%s/%s", canisterID.Encode(), path))
	resp, err := c.client.Post(url, "application/cbor", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case statusCorePass:
		defer resp.Body.Close()
		return io.ReadAll(resp.Body)
	default:
		return nil, fmt.Errorf("(%d) %s", resp.StatusCode, resp.Status)
	}
}

func (c Client) query(canisterID principal.Principal, data []byte) ([]byte, error) {
	return c.post("query", canisterID, data, 200)
}

func (c Client) readState(canisterID principal.Principal, data []byte) ([]byte, error) {
	return c.post("read_state", canisterID, data, 200)
}

func (c Client) url(p string) string {
	url := *c.config.Host
	url.Path = path.Join(url.Path, p)
	return url.String()
}

type ClientConfig struct {
	Host *url.URL
}
