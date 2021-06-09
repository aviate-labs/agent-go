package agent

import (
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/fxamacker/cbor/v2"
)

type ClientConfig struct {
	Host *url.URL
}

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

func (c Client) get(path string) ([]byte, error) {
	resp, err := c.client.Get(c.url(path))
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

func (c Client) url(p string) string {
	url := c.config.Host
	url.Path = path.Join(url.Path, p)
	return url.String()
}
