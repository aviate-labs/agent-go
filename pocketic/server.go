package pocketic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var HEADER = http.Header{
	"content-type":          []string{"application/json"},
	"processing-timeout-ms": []string{"300000"},
}

func checkResponse(resp *http.Response, body any) error {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read instances: %v", err)
	}
	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted) {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	if err := json.Unmarshal(raw, body); body != nil && err != nil {
		return fmt.Errorf("failed to unmarshal instances: %v", err)
	}
	return nil
}

type CanisterIDRange struct {
	Start struct {
		CanisterID string `json:"canister_id"`
	}
	End struct {
		CanisterID string `json:"canister_id"`
	} `json:"end"`
}

type NewInstanceResponse struct {
	InstanceID int                 `json:"instance_id"`
	Topology   map[string]Topology `json:"topology"`
}

type Topology struct {
	SubnetKind     SubnetKind        `json:"subnet_kind"`
	Size           int               `json:"size"`
	CanisterRanges []CanisterIDRange `json:"canister_ranges"`
}

type server struct {
	binPath string
	port    int
	cmd     *exec.Cmd
}

func newServer() (*server, error) {
	// Try to find the pocket-ic binary.
	path, err := exec.LookPath("pocket-ic-server")
	if path, err = exec.LookPath("pocket-ic"); err != nil {
		// If the binary is not found, try to find it in the POCKET_IC_BIN environment variable.
		if pathEnv := os.Getenv("POCKET_IC_BIN"); pathEnv != "" {
			path = pathEnv
		} else {
			path = "./pocket-ic"
			if _, err := os.Stat(path); err != nil {
				return nil, fmt.Errorf("pocket-ic binary not found: %v", err)
			}
		}
	}

	versionCmd := exec.Command(path, "--version")
	rawVersion, err := versionCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get pocket-ic version: %v", err)
	}
	version := strings.TrimPrefix(strings.TrimSpace(string(rawVersion)), "pocket-ic-server ")
	if !strings.HasPrefix(version, "3.") {
		return nil, fmt.Errorf("unsupported pocket-ic version, must be v3.x: %s", version)
	}

	pid := os.Getpid()
	cmdArgs := []string{"--pid", strconv.Itoa(pid)}
	cmd := exec.Command(path, cmdArgs...)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start pocket-ic: %v", err)
	}

	readyFile := fmt.Sprintf("%spocket_ic_%d.ready", os.TempDir(), pid)
	stopAt := time.Now().Add(10 * time.Second)
	for _, err := os.Stat(readyFile); os.IsNotExist(err); _, err = os.Stat(readyFile) {
		time.Sleep(100 * time.Millisecond)
		if time.Now().After(stopAt) {
			return nil, fmt.Errorf("pocket-ic did not start in time")
		}
	}

	portFile := fmt.Sprintf("%spocket_ic_%d.port", os.TempDir(), pid)
	f, err := os.OpenFile(portFile, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open port file: %v", err)
	}
	rawPort, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read port file: %v", err)
	}
	port, err := strconv.Atoi(string(rawPort))
	if err != nil {
		return nil, fmt.Errorf("failed to convert port to int: %v", err)
	}

	return &server{
		binPath: path,
		port:    port,
		cmd:     cmd,
	}, nil
}

func (s server) Close() error {
	if err := s.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill pocket-ic: %v", err)
	}
	return nil
}

// DeleteInstance deletes an instance.
func (s server) DeleteInstance(instanceID int) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/instances/%d", s.URL(), instanceID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header = HEADER
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete instance: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (s server) GetBlobStoreEntry(id string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/blobstore/%s", s.URL(), id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header = HEADER
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob store entry: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// InstanceGet provides a generic way the HTTP GET method for an instance.
func (s server) InstanceGet(instanceID int, endpoint string, body any) error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/instances/%d/%s", s.URL(), instanceID, endpoint), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header = HEADER
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get instance: %v", err)
	}
	if err := checkResponse(resp, body); err != nil {
		return fmt.Errorf("failed to get instance: %v", err)
	}
	return nil
}

// InstancePost provides a generic way the HTTP POST method for an instance.
func (s server) InstancePost(instanceID int, endpoint string, payload, body any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/instances/%d/%s", s.URL(), instanceID, endpoint), bytes.NewBuffer(raw))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header = HEADER
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post instance: %v", err)
	}
	if err := checkResponse(resp, body); err != nil {
		return fmt.Errorf("failed to post instance: %v", err)
	}
	return nil
}

// ListInstances returns a list of all instances running on the server.
func (s server) ListInstances() ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/instances", s.URL()), nil)
	if err != nil {
		return nil, nil
	}
	req.Header = HEADER
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get instances: %v", err)
	}
	var instances []string
	if err := checkResponse(resp, &instances); err != nil {
		return nil, fmt.Errorf("failed to get instances: %v", err)
	}
	return instances, nil
}

// NewInstance creates a new instance.
func (s server) NewInstance(subnetConfig ExtendedSubnetConfigSet) (*NewInstanceResponse, error) {
	// The JSON API expects empty slices instead of nil.
	if subnetConfig.Application == nil {
		subnetConfig.Application = make([]SubnetSpec, 0)
	}
	if subnetConfig.System == nil {
		subnetConfig.System = make([]SubnetSpec, 0)
	}

	raw, err := json.Marshal(subnetConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subnet config: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/instances", s.URL()), bytes.NewBuffer(raw))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header = HEADER
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %v", err)
	}
	var respBody struct {
		Created NewInstanceResponse `json:"Created"`
	}
	if err := checkResponse(resp, &respBody); err != nil {
		return nil, fmt.Errorf("failed to create instance: %v", err)
	}
	return &respBody.Created, nil
}

// SetBlobStoreEntry sets a blob store entry.
func (s server) SetBlobStoreEntry(blob []byte, compressed bool) (string, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/blobstore", s.URL()), bytes.NewBuffer(blob))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header = HEADER
	if compressed {
		req.Header.Set("Content-Encoding", "gzip")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to set blob store entry: %v", err)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return string(raw), nil
}

func (s server) Status() error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/status", s.URL()), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header = HEADER
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get status: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (s server) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", s.port)
}
