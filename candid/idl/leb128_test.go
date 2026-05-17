package idl

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/aviate-labs/agent-go/leb128"
)

func TestReadLEB128(t *testing.T) {
	for _, test := range []struct {
		hex string
	}{
		{"0"},
		{"FF"},
		{"FFFFFFFFFFFFFFFF"},
	} {
		t.Run(test.hex, func(t *testing.T) {
			bi := new(big.Int)
			if _, ok := bi.SetString(test.hex, 16); !ok {
				t.Fatal()
			}
			n, err := leb128.EncodeUnsigned(bi)
			if err != nil {
				t.Fatal(err)
			}
			r := bytes.NewReader(n)
			if _, err := readLEB128(r); err != nil {
				t.Fatal(err)
			}
			if r.Len() != 0 {
				t.Error("should have read everything")
			}
		})
	}
}

func TestReadSLEB128(t *testing.T) {
	for _, test := range []struct {
		hex string
		neg bool
	}{
		{hex: "0"},
		{hex: "FF"},
		{hex: "FFFFFFFFFFFFFFFF"},
		{hex: "FF", neg: true},
		{hex: "FFFFFFFFFFFFFFFF", neg: true},
	} {
		t.Run(fmt.Sprintf("%s %t", test.hex, test.neg), func(t *testing.T) {
			bi := new(big.Int)
			if _, ok := bi.SetString(test.hex, 16); !ok {
				t.Fatal()
			}
			if test.neg {
				bi = new(big.Int).Sub(big.NewInt(0), bi)
			}
			n, err := leb128.EncodeSigned(bi)
			if err != nil {
				t.Fatal(err)
			}
			r := bytes.NewReader(n)
			if _, err := readLEB128(r); err != nil {
				t.Fatal(err)
			}
			if r.Len() != 0 {
				t.Error("should have read everything")
			}
		})
	}
}
