package pocketic

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
)

// GetBlob retrieves a binary blob from the PocketIC server.
func (pic PocketIC) GetBlob(blobID []byte) ([]byte, error) {
	var bytes []byte
	if err := pic.do(
		http.MethodGet,
		fmt.Sprintf("%s/blobstore/%s", pic.server.URL(), hex.EncodeToString(blobID)),
		nil,
		&bytes,
	); err != nil {
		return nil, err
	}
	return bytes, nil
}

// UploadBlob uploads and stores a binary blob to the PocketIC server.
func (pic PocketIC) UploadBlob(bytes []byte, gzipCompression bool) ([]byte, error) {
	method := http.MethodPost
	url := fmt.Sprintf("%s/blobstore", pic.server.URL())
	pic.logger.Printf("[POCKETIC] %s %s %+v", method, url, bytes)
	req, err := newRequest(method, url, bytes)
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/octet-stream")
	if gzipCompression {
		req.Header.Set("content-encoding", "gzip")
	}
	resp, err := pic.client.Do(req)
	if err != nil {
		return nil, err
	}
	hexBlobID, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	blobID, err := hex.DecodeString(string(hexBlobID))
	if err != nil {
		return nil, err
	}
	return blobID, nil
}
