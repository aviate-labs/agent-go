package identity

import "github.com/aviate-labs/principal-go"

type Identity interface {
	Sender() principal.Principal
	Sign(msg []byte) []byte
	PublicKey() []byte
}
