package leb128_test

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"testing"

	"github.com/aviate-labs/agent-go/leb128"
)

func TestUnsigned(t *testing.T) {
	for _, test := range []struct {
		Hex   string
		Value *big.Int
	}{
		{"00", big.NewInt(0)},
		{"07", big.NewInt(7)},
		{"7F", big.NewInt(127)},
		{"E58E26", big.NewInt(624485)},
		{"80897A", big.NewInt(2000000)},
		{"808098F4E9B5CA6A", big.NewInt(60000000000000000)},
		{"EF9BAF8589CF959A92DEB7DE8A929EABB424", newInt(t, "24197857200151252728969465429440056815")},
	} {
		t.Run(test.Hex, func(t *testing.T) {
			e := new(big.Int).Set(test.Value)
			bs, err := leb128.EncodeUnsigned(e)
			if err != nil {
				t.Fatal(err)
			}
			if h := fmt.Sprintf("%X", bs); h != test.Hex {
				t.Errorf("\n%50s\n%50s", h, test.Hex)
			}

			d := new(big.Int).Set(test.Value)
			r := bytes.NewReader(bs)
			bi, err := leb128.DecodeUnsigned(r)
			if err != nil {
				t.Fatal(err)
			}
			if bi.Cmp(d) != 0 {
				t.Errorf("%s, \n%s\n%s", test.Hex, d, bi)
			}
			if r.Len() != 0 {
				t.Error()
			}
		})
	}
}

func TestUnsignedMultiple(t *testing.T) {
	v := big.NewInt(127)
	b, err := leb128.EncodeUnsigned(v)
	if err != nil {
		t.Fatal(err)
	}
	var bs []byte
	for range 10 {
		bs = append(bs, b...)
	}
	r := bytes.NewReader(bs)
	for range 10 {
		bi, err := leb128.DecodeUnsigned(r)
		if err != nil {
			t.Error(err)
		}
		if bi.Cmp(v) != 0 {
			t.Error(bi)
		}
	}
	if r.Len() != 0 {
		raw, _ := io.ReadAll(r)
		t.Fatalf("%x", raw)
	}
}

// TestUnsignedNonCanonical: non-canonical encodings must decode to the
// same value (WASM binary format spec).
func TestUnsignedNonCanonical(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  []byte
		want int64
	}{
		{"0x03", []byte{0x03}, 3},
		{"0x83 0x00", []byte{0x83, 0x00}, 3},
		{"0x83 0x80 0x00", []byte{0x83, 0x80, 0x00}, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := leb128.DecodeUnsigned(bytes.NewReader(tc.raw))
			if err != nil {
				t.Fatal(err)
			}
			if got.Cmp(big.NewInt(tc.want)) != 0 {
				t.Errorf("got %s, want %d", got, tc.want)
			}
		})
	}
}

func TestUnsignedStreamingTrailingBytes(t *testing.T) {
	enc, err := leb128.EncodeUnsigned(big.NewInt(127))
	if err != nil {
		t.Fatal(err)
	}
	buf := append(append([]byte(nil), enc...), 0xAA, 0xBB, 0xCC)
	r := bytes.NewReader(buf)
	if _, err := leb128.DecodeUnsigned(r); err != nil {
		t.Fatal(err)
	}
	if r.Len() != 3 {
		t.Fatalf("trailing bytes consumed: r.Len()=%d, want 3", r.Len())
	}
	rest := make([]byte, 3)
	if _, err := r.Read(rest); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rest, []byte{0xAA, 0xBB, 0xCC}) {
		t.Fatalf("trailing bytes: got %x, want AABBCC", rest)
	}
}

func TestUnsignedUnterminated(t *testing.T) {
	if _, err := leb128.DecodeUnsigned(bytes.NewReader([]byte{0x80, 0x80, 0x80})); err == nil {
		t.Fatal("expected error on unterminated ULEB128")
	}
}
