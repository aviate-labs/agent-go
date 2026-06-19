package idl

import (
	"bytes"
	"fmt"

	"github.com/aviate-labs/agent-go/leb128"
)

// FutureType is a placeholder for type-table entries whose opcode is not
// recognised by this decoder. Per Candid spec ("Deserialisation of future
// types"), the decoder skips them and continues so older clients can read
// payloads produced by newer schemas.
type FutureType struct {
	primType
	OpCode OpCode
}

func (FutureType) Decode(r *bytes.Reader) (any, error) {
	m, err := decodeLen(r)
	if err != nil {
		return nil, err
	}
	if _, err := leb128.DecodeUnsigned(r); err != nil {
		return nil, err
	}
	skip := make([]byte, m)
	if _, err := r.Read(skip); err != nil {
		return nil, err
	}
	return nil, nil
}

func (f FutureType) EncodeType(_ *TypeDefinitionTable) ([]byte, error) {
	return nil, fmt.Errorf("cannot encode future type")
}

func (FutureType) EncodeValue(_ any) ([]byte, error) {
	return nil, fmt.Errorf("cannot encode future value")
}

func (FutureType) Read(r *bytes.Reader) ([]byte, error) {
	m, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, err
	}
	n, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, err
	}
	ml, err := checkLen(m, r)
	if err != nil {
		return nil, err
	}
	body := make([]byte, ml)
	if _, err := r.Read(body); err != nil {
		return nil, err
	}
	mEnc, _ := leb128.EncodeUnsigned(m)
	nEnc, _ := leb128.EncodeUnsigned(n)
	return append(append(mEnc, nEnc...), body...), nil
}

func (f FutureType) String() string {
	return fmt.Sprintf("future(%d)", f.OpCode)
}

func (FutureType) UnmarshalGo(raw any, _v any) error {
	if raw == nil {
		return nil
	}
	return NewUnmarshalGoError(raw, _v)
}
