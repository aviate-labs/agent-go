package idl

import "math/big"

// Hashes a string to a number.
// ( Sum_(i=0..k) utf8(id)[i] * 223^(k-i) ) mod 2^32 where k = |utf8(id)|-1
func Hash(s string) *big.Int {
	h := big.NewInt(0)
	i := big.NewInt(2)
	i = i.Exp(i, big.NewInt(32), nil)
	for _, r := range s {
		h = h.Mul(h, big.NewInt(223))
		h = h.Add(h, big.NewInt(int64(r)))
		h = h.Mod(h, i)
	}
	return h
}
