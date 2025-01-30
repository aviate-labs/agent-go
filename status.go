package agent

// Status describes various status fields of the Internet Computer.
type Status struct {
	// The public key (a DER-encoded BLS key) of the root key of this Internet Computer instance.
	RootKey []byte `cbor:"root_key"`
}
