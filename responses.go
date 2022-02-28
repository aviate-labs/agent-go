package agent

type Response struct {
	Status     string            `cbor:"status"`
	Reply      map[string][]byte `cbor:"reply"`
	RejectCode uint64            `cbor:"reject_code"`
	RejectMsg  string            `cbor:"reject_message"`
}
