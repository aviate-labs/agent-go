package identity

type AnonymousIdentity struct{}

func (id AnonymousIdentity) Sign(msg []byte) ([]byte, error) {
	return nil, nil
}

func (id AnonymousIdentity) PublicKey() []byte {
	return nil
}

type Identity interface {
	Sign(msg []byte) ([]byte, error)
	PublicKey() []byte
}
