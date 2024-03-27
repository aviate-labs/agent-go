package hashtree_test

import (
	"bytes"
	"errors"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"testing"
)

func TestHashTree_Lookup_absent(t *testing.T) {
	tree := hashtree.NewHashTree(hashtree.Labeled{
		Label: hashtree.Label("a"),
		Tree: hashtree.Labeled{
			Label: hashtree.Label("b"),
			Tree: hashtree.Labeled{
				Label: hashtree.Label("c"),
				Tree: hashtree.Fork{
					LeftTree: hashtree.Labeled{
						Label: hashtree.Label("d0"),
					},
					RightTree: hashtree.Labeled{
						Label: hashtree.Label("d1"),
						Tree:  hashtree.Leaf("d"),
					},
				},
			},
		},
	})
	var lookupError hashtree.LookupError
	if _, err := tree.Lookup(hashtree.Label("a"), hashtree.Label("b"), hashtree.Label("c0"), hashtree.Label("d0")); !errors.As(err, &lookupError) || lookupError.Type != hashtree.LookupResultAbsent {
		t.Fatalf("unexpected lookup result")
	}
	if _, err := tree.Lookup(hashtree.Label("a"), hashtree.Label("b"), hashtree.Label("c"), hashtree.Label("d0")); !errors.As(err, &lookupError) || lookupError.Type != hashtree.LookupResultAbsent {
		t.Fatalf("unexpected lookup result")
	}

	v, err := tree.Lookup(hashtree.Label("a"), hashtree.Label("b"), hashtree.Label("c"), hashtree.Label("d1"))
	if err != nil {
		t.Fatalf("unexpected lookup result")
	}
	if !bytes.Equal(v, hashtree.Label("d")) {
		t.Fatalf("unexpected node value")
	}
}
