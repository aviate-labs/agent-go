package leb128_test

import (
	"math/big"
	"testing"
)

func newInt(t *testing.T, str string) *big.Int {
	bi, ok := new(big.Int).SetString(str, 10)
	if !ok {
		t.Fatal()
	}
	return bi
}
