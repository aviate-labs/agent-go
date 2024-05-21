package registry

import (
	"fmt"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	v1 "github.com/aviate-labs/agent-go/registry/proto/v1"
)

func NewHashTree(tree *v1.MixedHashTree) (*hashtree.HashTree, error) {
	root, err := newHashTree(tree)
	if err != nil {
		return nil, err
	}
	return &hashtree.HashTree{
		Root: root,
	}, nil

}

func newHashTree(tree *v1.MixedHashTree) (hashtree.Node, error) {
	switch t := tree.TreeEnum.(type) {
	case *v1.MixedHashTree_Empty:
		return hashtree.Empty{}, nil
	case *v1.MixedHashTree_LeafData:
		return hashtree.Leaf(t.LeafData), nil
	case *v1.MixedHashTree_Labeled_:
		n, err := newHashTree(t.Labeled.Subtree)
		if err != nil {
			return nil, err
		}
		return hashtree.Labeled{
			Label: t.Labeled.Label,
			Tree:  n,
		}, nil
	case *v1.MixedHashTree_Fork_:
		left, err := newHashTree(t.Fork.LeftTree)
		if err != nil {
			return nil, err
		}
		right, err := newHashTree(t.Fork.RightTree)
		if err != nil {
			return nil, err
		}
		return hashtree.Fork{
			LeftTree:  left,
			RightTree: right,
		}, nil
	case *v1.MixedHashTree_PrunedDigest:
		return hashtree.Pruned(t.PrunedDigest), nil
	default:
		return nil, fmt.Errorf("unsupported hash tree type: %T", tree.TreeEnum)
	}
}
