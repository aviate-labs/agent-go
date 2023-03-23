package agent

// Envelope is a wrapper for a Request that includes the sender's public key and signature.
type Envelope struct {
	Content      Request `cbor:"content,omitempty"`
	SenderPubKey []byte  `cbor:"sender_pubkey,omitempty"`
	SenderSig    []byte  `cbor:"sender_sig,omitempty"`
}
