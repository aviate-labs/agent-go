package leb128

import "math/big"

var (
	x00 = big.NewInt(0x00)
	x7F = big.NewInt(0x7F)
	x80 = big.NewInt(0x80)
)
