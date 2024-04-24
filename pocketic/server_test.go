package pocketic

import (
	"bytes"
	"testing"
)

func TestServer_SetBlobStoreEntry(t *testing.T) {
	s, err := newServer()
	if err != nil {
		t.Fatal(err)
	}
	blob := []byte("key")
	id, err := s.SetBlobStoreEntry(blob, false)
	if err != nil {
		t.Fatal(err)
	}

	v, err := s.GetBlobStoreEntry(id)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, blob) {
		t.Errorf("got %v, expected %v", v, blob)
	}
}

func TestServer_Status(t *testing.T) {
	s, err := newServer()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Status(); err != nil {
		t.Fatal(err)
	}
}
