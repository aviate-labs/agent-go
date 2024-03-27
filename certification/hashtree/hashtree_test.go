package hashtree

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
)

func TestHashTree_Lookup(t *testing.T) {
	t.Run("Empty Nodes", func(t *testing.T) {
		tree := NewHashTree(Fork{
			LeftTree: Labeled{
				Label: Label("label 1"),
				Tree:  Empty{},
			},
			RightTree: Fork{
				LeftTree: Pruned{},
				RightTree: Fork{
					LeftTree: Labeled{
						Label: Label("label 3"),
						Tree:  Leaf{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
					},
					RightTree: Labeled{
						Label: Label("label 5"),
						Tree:  Empty{},
					},
				},
			},
		})

		var lookupError LookupError
		for _, i := range []int{0, 1} {
			if _, err := tree.Lookup(Label(fmt.Sprintf("label %d", i))); !errors.As(err, &lookupError) || lookupError.Type != LookupResultAbsent {
				t.Fatalf("unexpected lookup result")
			} else if e := lookupError.Error(); e != fmt.Sprintf(`lookup error (path: "label %d") at "label %d": not found, not present in the tree`, i, i) {
				t.Fatalf("unexpected error message: %s", e)
			}
		}
		if _, err := tree.Lookup(Label("label 2")); !errors.As(err, &lookupError) || lookupError.Type != LookupResultUnknown {
			t.Fatalf("unexpected lookup result")
		} else if e := lookupError.Error(); e != `lookup error (path: "label 2") at "label 2": not found, could be pruned` {
			t.Fatalf("unexpected error message: %s", e)
		}
		if v, err := tree.Lookup(Label("label 3")); err != nil {
			t.Fatalf("unexpected lookup result")
		} else {
			if !bytes.Equal(v, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}) {
				t.Fatalf("unexpected node value")
			}
		}
		for _, i := range []int{4, 5, 6} {
			if _, err := tree.Lookup(Label(fmt.Sprintf("label %d", i))); !errors.As(err, &lookupError) || lookupError.Type != LookupResultAbsent {
				t.Fatalf("unexpected lookup result")
			} else if e := lookupError.Error(); e != fmt.Sprintf(`lookup error (path: "label %d") at "label %d": not found, not present in the tree`, i, i) {
				t.Fatalf("unexpected error message: %s", e)
			}
		}
	})
	t.Run("Nil Nodes", func(t *testing.T) {
		// let tree: HashTree<Vec<u8>> = fork(
		//        label("label 1", empty()),
		//        fork(
		//            fork(
		//                label("label 3", leaf(vec![1, 2, 3, 4, 5, 6])),
		//                label("label 5", empty()),
		//            ),
		//            pruned([1; 32]),
		//        ),
		//    );
		tree := NewHashTree(Fork{
			LeftTree: Labeled{
				Label: Label("label 1"),
			},
			RightTree: Fork{
				LeftTree: Fork{
					LeftTree: Labeled{
						Label: Label("label 3"),
						Tree:  Leaf{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
					},
					RightTree: Labeled{
						Label: Label("label 5"),
					},
				},
				RightTree: Pruned{},
			},
		})
		var lookupError LookupError
		for _, i := range []int{0, 1, 2} {
			if _, err := tree.Lookup(Label(fmt.Sprintf("label %d", i))); !errors.As(err, &lookupError) || lookupError.Type != LookupResultAbsent {
				t.Fatalf("unexpected lookup result")
			} else if e := lookupError.Error(); e != fmt.Sprintf(`lookup error (path: "label %d") at "label %d": not found, not present in the tree`, i, i) {
				t.Fatalf("unexpected error message: %s", e)
			}
		}
		if v, err := tree.Lookup(Label("label 3")); err != nil {
			t.Fatalf("unexpected lookup result")
		} else {
			if !bytes.Equal(v, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}) {
				t.Fatalf("unexpected node value")
			}
		}
		for _, i := range []int{4, 5} {
			if _, err := tree.Lookup(Label(fmt.Sprintf("label %d", i))); !errors.As(err, &lookupError) || lookupError.Type != LookupResultAbsent {
				t.Fatalf("unexpected lookup result")
			} else if e := lookupError.Error(); e != fmt.Sprintf(`lookup error (path: "label %d") at "label %d": not found, not present in the tree`, i, i) {
				t.Fatalf("unexpected error message: %s", e)
			}
		}
		if _, err := tree.Lookup(Label("label 6")); !errors.As(err, &lookupError) || lookupError.Type != LookupResultUnknown {
			t.Fatalf("unexpected lookup result")
		} else if e := lookupError.Error(); e != `lookup error (path: "label 6") at "label 6": not found, could be pruned` {
			t.Fatalf("unexpected error message: %s", e)
		}
	})
}

func TestHashTree_simple(t *testing.T) {
	tree := NewHashTree(Fork{
		LeftTree: Labeled{
			Label: Label("label 1"),
			Tree:  Empty{},
		},
		RightTree: Fork{
			LeftTree: Pruned{
				0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
				0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
				0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
				0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
			},
			RightTree: Leaf{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		},
	})
	digest := tree.Digest()
	if hex.EncodeToString(digest[:]) != "69cf325d0f20505b261821a7e77ff72fb9a8753a7964f0b587553bfb44e72532" {
		t.Fatalf("unexpected digest: %x", digest)
	}
}
