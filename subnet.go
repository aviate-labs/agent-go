package agent

import (
	"fmt"
	"github.com/aviate-labs/agent-go/certification"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/fxamacker/cbor/v2"
	"math/big"
)

func (a Agent) GetSubnetMetrics(subnetID principal.Principal) (*SubnetMetrics, error) {
	path := []hashtree.Label{hashtree.Label("subnet"), subnetID.Raw, hashtree.Label("metrics")}
	cert, err := a.readSubnetStateCertificate(subnetID, [][]hashtree.Label{path})
	if err != nil {
		return nil, err
	}
	rawMetrics, err := cert.Tree.Lookup(path...)
	if err != nil {
		return nil, err
	}

	var metrics SubnetMetrics
	if err := cbor.Unmarshal(rawMetrics, &metrics); err != nil {
		return nil, err
	}
	return &metrics, nil
}

func (a Agent) GetSubnets() ([]principal.Principal, error) {
	path := []hashtree.Label{hashtree.Label("subnet")}
	cert, err := a.readSubnetStateCertificate(principal.MustDecode(certification.RootSubnetID), [][]hashtree.Label{path})
	if err != nil {
		return nil, err
	}
	subnets, err := cert.Tree.LookupSubTree(path...)
	if err != nil {
		return nil, err
	}
	children, err := hashtree.AllChildren(subnets)
	if err != nil {
		return nil, err
	}
	var subnetsList []principal.Principal
	for _, p := range children {
		subnetsList = append(subnetsList, principal.Principal{Raw: p.Path[0]})
	}
	return subnetsList, nil
}

func (a Agent) GetSubnetsInfo() ([]SubnetInfo, error) {
	rootSubnetID := principal.MustDecode(certification.RootSubnetID)
	path := []hashtree.Label{hashtree.Label("subnet")}
	cert, err := a.readSubnetStateCertificate(rootSubnetID, [][]hashtree.Label{path})
	if err != nil {
		return nil, err
	}
	subnets, err := cert.Tree.LookupSubTree(path...)
	if err != nil {
		return nil, err
	}
	children, err := hashtree.AllChildren(subnets)
	if err != nil {
		return nil, err
	}
	var subnetsInfo []SubnetInfo
	for _, p := range children {
		subnetID := principal.Principal{Raw: p.Path[0]}
		publicKey, err := hashtree.Lookup(p.Value, hashtree.Label("public_key"))
		if err != nil {
			return nil, err
		}
		rawCanisterRanges, err := hashtree.Lookup(p.Value, hashtree.Label("canister_ranges"))
		if err != nil {
			return nil, err
		}
		var canisterRanges certification.CanisterRanges
		if err := cbor.Unmarshal(rawCanisterRanges, &canisterRanges); err != nil {
			return nil, err
		}

		nodes, err := cert.Tree.LookupSubTree(hashtree.Label("subnet"), subnetID.Raw, hashtree.Label("node"))
		if err != nil {
			path = []hashtree.Label{hashtree.Label("subnet"), subnetID.Raw, hashtree.Label("node")}
			nodesCert, err := a.readSubnetStateCertificate(subnetID, [][]hashtree.Label{path})
			if err != nil {
				return nil, err
			}
			nodes, err = nodesCert.Tree.LookupSubTree(path...)
			if err != nil {
				return nil, err
			}
		}
		nodeChildren, err := hashtree.AllChildren(nodes)
		if err != nil {
			return nil, err
		}
		var nodesInfo []NodeInfo
		for _, node := range nodeChildren {
			nodeID := principal.Principal{Raw: node.Path[0]}
			nodePublicKey, err := hashtree.Lookup(node.Value, hashtree.Label("public_key"))
			if err != nil {
				return nil, err
			}
			nodesInfo = append(
				nodesInfo,
				NodeInfo{
					NodeID:    nodeID,
					PublicKey: nodePublicKey,
				},
			)
		}
		subnetsInfo = append(subnetsInfo, SubnetInfo{
			SubnetID:       subnetID,
			PublicKey:      publicKey,
			CanisterRanges: canisterRanges,
			Nodes:          nodesInfo,
		})
	}
	return subnetsInfo, nil
}

type NodeInfo struct {
	NodeID    principal.Principal
	PublicKey []byte
}

type SubnetInfo struct {
	SubnetID       principal.Principal
	PublicKey      []byte
	CanisterRanges certification.CanisterRanges
	Nodes          []NodeInfo
}

type SubnetMetrics struct {
	NumCanisters            uint64
	CanisterStateBytes      uint64
	ConsumedCyclesTotal     big.Int
	UpdateTransactionsTotal uint64
}

func (m *SubnetMetrics) UnmarshalCBOR(bytes []byte) error {
	var metricsMap map[int]any
	if err := cbor.Unmarshal(bytes, &metricsMap); err != nil {
		return err
	}

	var ok bool

	m.NumCanisters, ok = metricsMap[0].(uint64)
	if !ok {
		return fmt.Errorf("unexpected type for NumCanisters: %T", metricsMap[0])
	}

	m.CanisterStateBytes, ok = metricsMap[1].(uint64)
	if !ok {
		return fmt.Errorf("unexpected type for CanisterStateBytes: %T", metricsMap[1])
	}

	rawConsumedCyclesTotal, ok := metricsMap[2].(map[any]any)
	if !ok {
		return fmt.Errorf("unexpected type for ConsumedCyclesTotal: %T", metricsMap[2])
	}
	rawLow, ok := rawConsumedCyclesTotal[uint64(0)].(uint64)
	if !ok {
		return fmt.Errorf("unexpected type for low: %T", rawConsumedCyclesTotal[0])
	}
	rawHigh, ok := rawConsumedCyclesTotal[uint64(1)].(uint64)
	if !ok {
		return fmt.Errorf("unexpected type for high: %T", rawConsumedCyclesTotal[1])
	}
	m.ConsumedCyclesTotal.SetUint64(rawHigh)
	m.ConsumedCyclesTotal.Lsh(&m.ConsumedCyclesTotal, 64)
	m.ConsumedCyclesTotal.Add(&m.ConsumedCyclesTotal, new(big.Int).SetUint64(rawLow))

	m.UpdateTransactionsTotal, ok = metricsMap[3].(uint64)
	if !ok {
		return fmt.Errorf("unexpected type for UpdateTransactionsTotal: %T", metricsMap[3])
	}
	return nil
}
