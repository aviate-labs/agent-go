package marshal

import (
	"bytes"
	"math/big"
)

func concat(bs ...[]byte) []byte {
	var c []byte
	for _, b := range bs {
		c = append(c, b...)
	}
	return c
}

func pad00(n int, bs []byte) []byte {
	for len(bs) != n {
		bs = append(bs, 0)
	}
	return bs
}

func padff(n int, bs []byte) []byte {
	for len(bs) != n {
		bs = append(bs, 0xff)
	}
	return bs
}

func readInt(r *bytes.Reader, n int) (*big.Int, error) {
	bi, err := readUInt(r, n)
	if err != nil {
		return nil, err
	}
	m := big.NewInt(2)
	m = m.Exp(m, big.NewInt(int64((n-1)*8+7)), nil)
	if bi.Cmp(m) >= 0 {
		v := new(big.Int).Set(m)
		v = v.Mul(v, big.NewInt(-2))
		bi = bi.Add(bi, v)
	}
	return bi, nil
}

func readUInt(r *bytes.Reader, n int) (*big.Int, error) {
	var (
		bi  = new(big.Int)
		xFF = big.NewInt(256)
		m   = big.NewInt(1)
	)
	for i := 0; i < n; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		v := new(big.Int).SetBytes([]byte{b})
		bi = bi.Add(bi, v.Mul(v, m))
		m = m.Mul(m, xFF)
	}
	return bi, nil
}

func reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func twosComplement(bi *big.Int) *big.Int {
	inv := bi.Bytes()
	for i, b := range inv {
		inv[i] = ^b
	}
	bi.SetBytes(inv)
	return bi.Add(bi, big.NewInt(1))
}

func writeInt(bi *big.Int, n int) []byte {
	switch bi.Sign() {
	case 0:
		return zeros(n)
	case -1:
		bi := new(big.Int).Set(bi)
		return padff(n, reverse(twosComplement(bi).Bytes()))
	default:
		return pad00(n, reverse(bi.Bytes()))
	}
}

func zeros(n int) []byte {
	var z []byte
	for i := 0; i < n; i++ {
		z = append(z, 0)
	}
	return z
}
