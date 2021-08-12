package agent

import (
	"github.com/fxamacker/cbor/v2"
)

// Implementation identifies the implementation of the Internet Computer.
type Implementation struct {
	// Source is the canonical location of the source code.
	Source string
	// Version is the version number of the implementation.
	Version string
	// Revision is the precise git revision of the implementation.
	Revision string
}

// Status describes various status fields of the Internet Computer.
type Status struct {
	// Identifies the interface version supported.
	Version string
	// Impl describes the implementation of the Internet Computer.
	Impl *Implementation
	// The public key (a DER-encoded BLS key) of the root key of this Internet Computer instance.
	RootKey []byte
}

func (s *Status) UnmarshalCBOR(data []byte) error {
	var status struct {
		APIVersion   string `cbor:"ic_api_version"`
		ImplSource   string `cbor:"impl_source"`
		ImplVersion  string `cbor:"impl_version"`
		ImplRevision string `cbor:"impl_revision"`
		RootKey      []byte `cbor:"root_key"`
	}
	if err := cbor.Unmarshal(data, &status); err != nil {
		return err
	}
	s.Version = status.APIVersion
	s.Impl = &Implementation{
		Source:   status.ImplSource,
		Version:  status.ImplVersion,
		Revision: status.ImplRevision,
	}
	s.RootKey = status.RootKey
	return nil
}
