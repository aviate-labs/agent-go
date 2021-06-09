package crc32

import (
	"hash"
	"hash/crc32"
)

type Sequence struct {
	b []byte
	h hash.Hash
}

func New(raw []byte) Sequence {
	h := crc32.NewIEEE()
	h.Write(raw)
	return Sequence{
		b: raw,
		h: h,
	}
}

func (h Sequence) Bytes() []byte {
	return h.b
}

func (h Sequence) Value() []byte {
	return h.h.Sum(nil)
}
