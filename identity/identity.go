package identity

import "github.com/aviate-labs/agent-go/principal"

type Identity interface {
	Sender() principal.Principal
	Sign(msg []byte) []byte
	PublicKey() []byte
}
