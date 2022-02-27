package agent

type Envelope struct {
	Content      Request `cbor:"content,omitempty"`
	SenderPubkey []byte  `cbor:"sender_pubkey,omitempty"`
	SenderSig    []byte  `cbor:"sender_sig,omitempty"`
}
