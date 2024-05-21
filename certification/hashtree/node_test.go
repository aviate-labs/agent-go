package hashtree_test

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/aviate-labs/agent-go/certification/hashtree"
)

var pruned = hashtree.Fork{
	LeftTree: hashtree.Fork{
		LeftTree: hashtree.Labeled{
			Label: []byte("a"),
			Tree: hashtree.Fork{
				LeftTree: hashtree.Pruned(h2b("1b4feff9bef8131788b0c9dc6dbad6e81e524249c879e9f10f71ce3749f5a638")),
				RightTree: hashtree.Labeled{
					Label: []byte("y"),
					Tree:  hashtree.Leaf("world"),
				},
			},
		},
		RightTree: hashtree.Labeled{
			Label: []byte("b"),
			Tree:  hashtree.Pruned(h2b("7b32ac0c6ba8ce35ac82c255fc7906f7fc130dab2a090f80fe12f9c2cae83ba6")),
		},
	},
	RightTree: hashtree.Fork{
		LeftTree: hashtree.Pruned(h2b("ec8324b8a1f1ac16bd2e806edba78006479c9877fed4eb464a25485465af601d")),
		RightTree: hashtree.Labeled{
			Label: []byte("d"),
			Tree:  hashtree.Leaf("morning"),
		},
	},
}

var tree = hashtree.Fork{
	LeftTree: hashtree.Fork{
		LeftTree: hashtree.Labeled{
			Label: []byte("a"),
			Tree: hashtree.Fork{
				LeftTree: hashtree.Fork{
					LeftTree: hashtree.Labeled{
						Label: []byte("x"),
						Tree:  hashtree.Leaf("hello"),
					},
					RightTree: hashtree.Empty{},
				},
				RightTree: hashtree.Labeled{
					Label: []byte("y"),
					Tree:  hashtree.Leaf("world"),
				},
			},
		},
		RightTree: hashtree.Labeled{
			Label: []byte("b"),
			Tree:  hashtree.Leaf("good"),
		},
	},
	RightTree: hashtree.Fork{
		LeftTree: hashtree.Labeled{
			Label: []byte("c"),
			Tree:  hashtree.Empty{},
		},
		RightTree: hashtree.Labeled{
			Label: []byte("d"),
			Tree:  hashtree.Leaf("morning"),
		},
	},
}

func ExampleAllPaths() {
	paths, _ := hashtree.AllPaths(tree)
	for _, path := range paths {
		var p []string
		for _, l := range path.Path {
			p = append(p, string(l))
		}
		fmt.Printf(
			"%s: %s\n",
			strings.Join(p, "/"),
			string(path.Value),
		)
	}
	// Output:
	// a/x: hello
	// a/y: world
	// b: good
	// d: morning
}

func ExampleDeserialize() {
	data, _ := hex.DecodeString("8301830183024161830183018302417882034568656c6c6f810083024179820345776f726c6483024162820344676f6f648301830241638100830241648203476d6f726e696e67")
	fmt.Println(hashtree.Deserialize(data))
	// Output:
	// {{a:{{x:hello|∅}|y:world}|b:good}|{c:∅|d:morning}} <nil>
}

func ExamplePruned() {
	fmt.Printf("%x", pruned.Reconstruct())
	// Output:
	// eb5c5b2195e62d996b84c9bcc8259d19a83786a2f59e0878cec84c811f669aa0
}

func ExampleSerialize() {
	b, _ := hashtree.Serialize(tree)
	fmt.Printf("%x", b)
	// Output:
	// 8301830183024161830183018302417882034568656c6c6f810083024179820345776f726c6483024162820344676f6f648301830241638100830241648203476d6f726e696e67
}

func Example_b() {
	fmt.Printf("%x", hashtree.Leaf("good").Reconstruct())
	// Output:
	// 7b32ac0c6ba8ce35ac82c255fc7906f7fc130dab2a090f80fe12f9c2cae83ba6
}

func Example_c() {
	fmt.Printf("%x", hashtree.Labeled{
		Label: []byte("c"),
		Tree:  hashtree.Empty{},
	}.Reconstruct())
	// Output:
	// ec8324b8a1f1ac16bd2e806edba78006479c9877fed4eb464a25485465af601d
}

func Example_root() {
	fmt.Printf("%x", tree.Reconstruct())
	// Output:
	// eb5c5b2195e62d996b84c9bcc8259d19a83786a2f59e0878cec84c811f669aa0
}

// Source: https://sdk.dfinity.org/docs/interface-spec/index.html#_example
// ─┬─┬╴"a" ─┬─┬╴"x" ─╴"hello"
//
//	│ │      │ └╴Empty
//	│ │      └╴  "y" ─╴"world"
//	│ └╴"b" ──╴"good"
//	└─┬╴"c" ──╴Empty
//	  └╴"d" ──╴"morning"
func Example_x() {
	fmt.Printf("%x", hashtree.Fork{
		LeftTree: hashtree.Labeled{
			Label: []byte("x"),
			Tree:  hashtree.Leaf("hello"),
		},
		RightTree: hashtree.Empty{},
	}.Reconstruct())
	// Output:
	// 1b4feff9bef8131788b0c9dc6dbad6e81e524249c879e9f10f71ce3749f5a638
}

func TestUFT8Leaf(t *testing.T) {
	// []byte{0x90, 0xe4, 0xcf, 0xfc, 0xda, 0x94, 0x83, 0xec, 0x16} is the unsigned leb128 encoding of 1646079569558762000.
	// Which is Mon Feb 28 2022 20:19:29 GMT+0000 in nanoseconds unix time.
	l := hashtree.Leaf([]byte{0x90, 0xe4, 0xcf, 0xfc, 0xda, 0x94, 0x83, 0xec, 0x16})
	if l.String() != "0x90e4cffcda9483ec16" {
		t.Error(l)
	}

	s := hashtree.Leaf("some string")
	if s.String() != "some string" {
		t.Error(l)
	}
}

func h2b(s string) [32]byte {
	var bs [32]byte
	b, _ := hex.DecodeString(s)
	copy(bs[:], b)
	return bs
}
