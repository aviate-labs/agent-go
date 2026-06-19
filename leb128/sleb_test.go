package leb128_test

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/big"
	"testing"

	"github.com/aviate-labs/agent-go/leb128"
)

func TestAppendSignedInt64(t *testing.T) {
	for _, v := range []int64{
		0, 1, -1, 63, 64, -64, -65, 127, -128, 624485, -624485,
		math.MaxInt32, math.MinInt32, math.MaxInt64, math.MinInt64,
	} {
		want, err := leb128.EncodeSigned(big.NewInt(v))
		if err != nil {
			t.Fatal(err)
		}
		var buf [10]byte
		got := leb128.AppendSignedInt64(buf[:0], v)
		if fmt.Sprintf("%X", got) != fmt.Sprintf("%X", want) {
			t.Errorf("v=%d: got %X, want %X", v, got, want)
		}
	}
}

func TestSigned(t *testing.T) {
	for _, test := range []struct {
		Hex   string
		Value *big.Int
	}{
		{"2A", big.NewInt(42)},
		{"7F", big.NewInt(-1)},
		{"C0BB78", big.NewInt(-123456)},
		{"8089FA00", big.NewInt(2000000)},
		{"808098F4E9B5CAEA00", big.NewInt(60000000000000000)},
		{"EF9BAF8589CF959A92DEB7DE8A929EABB424", newInt(t, "24197857200151252728969465429440056815")},
		{"91E4D0FAF6B0EAE5EDA1C8A1F5EDE1D4CB5B", newInt(t, "-24197857200151252728969465429440056815")},
	} {
		t.Run(test.Hex, func(t *testing.T) {
			e := new(big.Int).Set(test.Value)
			bs, err := leb128.EncodeSigned(e)
			if err != nil {
				t.Fatal(err)
			}
			if h := fmt.Sprintf("%X", bs); h != test.Hex {
				t.Errorf("\n%50s\n%50s", h, test.Hex)
			}

			d := new(big.Int).Set(test.Value)
			r := bytes.NewReader(bs)
			bi, err := leb128.DecodeSigned(r)
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

// TestSignedLargeCanonical pins exact bytes for magnitudes > int64,
// guarding the v.Int64() % 0x80 path inside EncodeSigned.
func TestSignedLargeCanonical(t *testing.T) {
	for _, tc := range []struct {
		name string
		v    *big.Int
		hex  string
	}{
		{"2^70", new(big.Int).Lsh(big.NewInt(1), 70), "8080808080808080808001"},
		{"-(2^70)", new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 70)), "808080808080808080807F"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			enc, err := leb128.EncodeSigned(tc.v)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%X", enc); got != tc.hex {
				t.Errorf("encoding: got %s, want %s", got, tc.hex)
			}
			dec, err := leb128.DecodeSigned(bytes.NewReader(enc))
			if err != nil {
				t.Fatal(err)
			}
			if dec.Cmp(tc.v) != 0 {
				t.Errorf("round-trip: got %s, want %s", dec, tc.v)
			}
		})
	}
}

func TestSignedMultiple(t *testing.T) {
	v := big.NewInt(-1)
	b, err := leb128.EncodeSigned(v)
	if err != nil {
		t.Fatal(err)
	}
	var bs []byte
	for range 10 {
		bs = append(bs, b...)
	}
	r := bytes.NewReader(bs)
	for range 10 {
		bi, err := leb128.DecodeSigned(r)
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

// TestSignedNonCanonical: non-canonical signed encodings must decode
// to the same value (WASM binary format spec).
func TestSignedNonCanonical(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  []byte
		want int64
	}{
		{"0x7E", []byte{0x7E}, -2},
		{"0xFE 0x7F", []byte{0xFE, 0x7F}, -2},
		{"0xFE 0xFF 0x7F", []byte{0xFE, 0xFF, 0x7F}, -2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := leb128.DecodeSigned(bytes.NewReader(tc.raw))
			if err != nil {
				t.Fatal(err)
			}
			if got.Cmp(big.NewInt(tc.want)) != 0 {
				t.Errorf("got %s, want %d", got, tc.want)
			}
		})
	}
}

func TestSignedStreamingTrailingBytes(t *testing.T) {
	enc, err := leb128.EncodeSigned(big.NewInt(-1))
	if err != nil {
		t.Fatal(err)
	}
	buf := append(append([]byte(nil), enc...), 0xAA, 0xBB, 0xCC)
	r := bytes.NewReader(buf)
	if _, err := leb128.DecodeSigned(r); err != nil {
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

func TestSignedTooShort(t *testing.T) {
	raw, _ := leb128.EncodeSigned(big.NewInt(128))
	// [x80, x01]
	b, _ := leb128.DecodeSigned(bytes.NewReader(raw))
	if b.Cmp(big.NewInt(128)) != 0 {
		t.Fatal(b)
	}
	// [x80]
	if _, err := leb128.DecodeSigned(bytes.NewReader(raw[:len(raw)-2])); err == nil {
		t.Error()
	}
}
