package certificate

import "github.com/aviate-labs/agent-go/principal"

// Certificate is a certificate gets returned by the IC.
type Certificate struct {
	// Tree is the certificate tree.
	Tree HashTree
	// Signature is the signature of the certificate tree.
	Signature []byte
	// Delegation is the delegation of the certificate.
	Delegation *Delegation
}

// Delegation is a delegation of a certificate.
type Delegation struct {
	// SubnetId is the subnet ID of the delegation.
	SubnetId principal.Principal
	// The nested certificate typically does not itself again contain a
	// delegation, although there is no reason why agents should enforce that
	// property.
	Certificate Certificate
}
