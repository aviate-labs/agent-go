package certificate_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	cert "github.com/aviate-labs/agent-go/certificate"
)

var pruned = cert.Fork{
	LeftTree: cert.Fork{
		LeftTree: cert.Labeled{
			Label: []byte("a"),
			Tree: cert.Fork{
				LeftTree: cert.Pruned(h2b("1B4FEFF9BEF8131788B0C9DC6DBAD6E81E524249C879E9F10F71CE3749F5A638")),
				RightTree: cert.Labeled{
					Label: []byte("y"),
					Tree:  cert.Leaf("world"),
				},
			},
		},
		RightTree: cert.Labeled{
			Label: []byte("b"),
			Tree:  cert.Pruned(h2b("7B32AC0C6BA8CE35AC82C255FC7906F7FC130DAB2A090F80FE12F9C2CAE83BA6")),
		},
	},
	RightTree: cert.Fork{
		LeftTree: cert.Pruned(h2b("EC8324B8A1F1AC16BD2E806EDBA78006479C9877FED4EB464A25485465AF601D")),
		RightTree: cert.Labeled{
			Label: []byte("d"),
			Tree:  cert.Leaf("morning"),
		},
	},
}

var tree = cert.Fork{
	LeftTree: cert.Fork{
		LeftTree: cert.Labeled{
			Label: []byte("a"),
			Tree: cert.Fork{
				LeftTree: cert.Fork{
					LeftTree: cert.Labeled{
						Label: []byte("x"),
						Tree:  cert.Leaf("hello"),
					},
					RightTree: cert.Empty{},
				},
				RightTree: cert.Labeled{
					Label: []byte("y"),
					Tree:  cert.Leaf("world"),
				},
			},
		},
		RightTree: cert.Labeled{
			Label: []byte("b"),
			Tree:  cert.Leaf("good"),
		},
	},
	RightTree: cert.Fork{
		LeftTree: cert.Labeled{
			Label: []byte("c"),
			Tree:  cert.Empty{},
		},
		RightTree: cert.Labeled{
			Label: []byte("d"),
			Tree:  cert.Leaf("morning"),
		},
	},
}

func ExampleB() {
	fmt.Printf("%X", cert.Leaf("good").Reconstruct())
	// Output:
	// 7B32AC0C6BA8CE35AC82C255FC7906F7FC130DAB2A090F80FE12F9C2CAE83BA6
}

func ExampleC() {
	fmt.Printf("%X", cert.Labeled{
		Label: []byte("c"),
		Tree:  cert.Empty{},
	}.Reconstruct())
	// Output:
	// EC8324B8A1F1AC16BD2E806EDBA78006479C9877FED4EB464A25485465AF601D
}

func ExampleDeserialize() {
	data, _ := hex.DecodeString("8301830183024161830183018302417882034568656c6c6f810083024179820345776f726c6483024162820344676f6f648301830241638100830241648203476d6f726e696e67")
	fmt.Println(cert.Deserialize(data))
	// Output:
	// {{a:{{x:hello|∅}|y:world}|b:good}|{c:∅|d:morning}} <nil>
}

func ExamplePruned() {
	fmt.Printf("%X", pruned.Reconstruct())
	// Output:
	// EB5C5B2195E62D996B84C9BCC8259D19A83786A2F59E0878CEC84C811F669AA0
}

func ExampleRoot() {
	fmt.Printf("%X", tree.Reconstruct())
	// Output:
	// EB5C5B2195E62D996B84C9BCC8259D19A83786A2F59E0878CEC84C811F669AA0
}

func ExampleSerialize() {
	b, _ := cert.Serialize(tree)
	fmt.Printf("%x", b)
	// Output:
	// 8301830183024161830183018302417882034568656c6c6f810083024179820345776f726c6483024162820344676f6f648301830241638100830241648203476d6f726e696e67
}

// Source: https://sdk.dfinity.org/docs/interface-spec/index.html#_example
// ─┬─┬╴"a" ─┬─┬╴"x" ─╴"hello"
//
//	│ │      │ └╴Empty
//	│ │      └╴  "y" ─╴"world"
//	│ └╴"b" ──╴"good"
//	└─┬╴"c" ──╴Empty
//	  └╴"d" ──╴"morning"
func ExampleX() {
	fmt.Printf("%X", cert.Fork{
		LeftTree: cert.Labeled{
			Label: []byte("x"),
			Tree:  cert.Leaf("hello"),
		},
		RightTree: cert.Empty{},
	}.Reconstruct())
	// Output:
	// 1B4FEFF9BEF8131788B0C9DC6DBAD6E81E524249C879E9F10F71CE3749F5A638
}

func h2b(s string) [32]byte {
	var bs [32]byte
	b, _ := hex.DecodeString(s)
	copy(bs[:], b)
	return bs
}

func TestUFT8Leaf(t *testing.T) {
	// []byte{0x90, 0xe4, 0xcf, 0xfc, 0xda, 0x94, 0x83, 0xec, 0x16} is the unsigned leb128 encoding of 1646079569558762000.
	// Which is Mon Feb 28 2022 20:19:29 GMT+0000 in nanoseconds unix time.
	l := cert.Leaf([]byte{0x90, 0xe4, 0xcf, 0xfc, 0xda, 0x94, 0x83, 0xec, 0x16})
	if l.String() != "0x90e4cffcda9483ec16" {
		t.Error(l)
	}

	s := cert.Leaf([]byte("some string"))
	if s.String() != "some string" {
		t.Error(l)
	}
}
