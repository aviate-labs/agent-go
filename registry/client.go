package registry

import (
	"fmt"
	"github.com/aviate-labs/agent-go/certification"
	"github.com/aviate-labs/agent-go/principal"
	v1 "github.com/aviate-labs/agent-go/registry/proto/v1"
	"google.golang.org/protobuf/proto"
	"strings"
)

type Client struct {
	dp *DataProvider
}

func New() (*Client, error) {
	dp, err := NewDataProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create data provider: %w", err)
	}
	return &Client{
		dp: dp,
	}, nil
}

func (c *Client) GetNNSSubnetID() (*principal.Principal, error) {
	v, _, err := c.dp.GetValueUpdate([]byte("nns_subnet_id"), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get NNS subnet ID: %w", err)
	}
	var nnsSubnetID v1.SubnetId
	if err := proto.Unmarshal(v, &nnsSubnetID); err != nil {
		return nil, fmt.Errorf("failed to unmarshal NNS subnet ID: %w", err)
	}
	return &principal.Principal{Raw: nnsSubnetID.PrincipalId.Raw}, nil
}

func (c *Client) GetLatestVersion() (uint64, error) {
	return c.dp.GetLatestVersion()
}

func (c *Client) GetNodeListSince(version uint64) (NodeMap, error) {
	nnsSubnetID, err := c.GetNNSSubnetID()
	if err != nil {
		return nil, err
	}
	nnsPublicKey, err := c.GetSubnetPublicKey(*nnsSubnetID)
	if err != nil {
		return nil, err
	}

	latestVersion, err := c.dp.GetLatestVersion()
	if err != nil {
		return nil, err
	}

	currentVersion := version
	nodeMap := make(map[string]*v1.NodeRecord)
	nodeOperatorMap := make(map[string]*v1.NodeOperatorRecord)
	for {
		records, _, err := c.dp.GetCertifiedChangesSince(currentVersion, nnsPublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get certified changes: %w", err)
		}
		currentVersion = records[len(records)-1].Version
		for _, record := range records {
			if strings.HasPrefix(record.Key, "node_record_") {
				if record.Value == nil {
					delete(nodeMap, strings.TrimPrefix(record.Key, "node_record_"))
				} else {
					var nodeRecord v1.NodeRecord
					if err := proto.Unmarshal(record.Value, &nodeRecord); err != nil {
						return nil, fmt.Errorf("failed to unmarshal node record: %w", err)
					}
					nodeMap[strings.TrimPrefix(record.Key, "node_record_")] = &nodeRecord
				}
			} else if strings.HasPrefix(record.Key, "node_operator_record_") {
				if record.Value == nil {
					delete(nodeOperatorMap, strings.TrimPrefix(record.Key, "node_operator_record_"))
				} else {
					var nodeOperatorRecord v1.NodeOperatorRecord
					if err := proto.Unmarshal(record.Value, &nodeOperatorRecord); err != nil {
						return nil, fmt.Errorf("failed to unmarshal node operator record: %w", err)
					}
					nodeOperatorMap[strings.TrimPrefix(record.Key, "node_operator_record_")] = &nodeOperatorRecord
				}
			}
		}
		if currentVersion >= latestVersion {
			break
		}
	}

	nodeDetailsMap := make(map[string]NodeDetails)
	for key, nodeRecord := range nodeMap {
		nodeOperatorID := principal.Principal{Raw: nodeRecord.NodeOperatorId}
		var nodeProviderID principal.Principal
		var dcID string
		if no, ok := nodeOperatorMap[nodeOperatorID.String()]; ok {
			nodeProviderID = principal.Principal{Raw: no.NodeProviderPrincipalId}
			dcID = no.DcId
		}
		var ipv6 string
		if nodeRecord.Http != nil {
			ipv6 = nodeRecord.Http.IpAddr
		}
		var ipv4 *IPv4Interface
		if nodeRecord.PublicIpv4Config != nil {
			ipv4 = &IPv4Interface{
				Address:      nodeRecord.PublicIpv4Config.IpAddr,
				Gateways:     nodeRecord.PublicIpv4Config.GatewayIpAddr,
				PrefixLength: nodeRecord.PublicIpv4Config.PrefixLength,
			}
		}
		nodeDetailsMap[key] = NodeDetails{
			IPv6:            ipv6,
			IPv4:            ipv4,
			NodeProviderID:  nodeProviderID,
			NodeOperatorID:  nodeOperatorID,
			DCID:            dcID,
			HostOSVersionID: nodeRecord.HostosVersionId,
			Domain:          nodeRecord.Domain,
		}
	}
	return nodeDetailsMap, nil
}

func (c *Client) GetSubnetDetails(subnetID principal.Principal) (*v1.SubnetRecord, error) {
	v, _, err := c.dp.GetValueUpdate([]byte(fmt.Sprintf("subnet_record_%s", subnetID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subnet details: %w", err)
	}
	var record v1.SubnetRecord
	if err := proto.Unmarshal(v, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subnet details: %w", err)
	}
	return &record, nil
}

func (c *Client) GetSubnetIDs() ([]principal.Principal, error) {
	v, _, err := c.dp.GetValueUpdate([]byte("subnet_list"), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subnet IDs: %w", err)
	}
	var list v1.SubnetListRecord
	if err := proto.Unmarshal(v, &list); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subnet IDs: %w", err)
	}
	var subnets []principal.Principal
	for _, subnet := range list.Subnets {
		subnets = append(subnets, principal.Principal{Raw: subnet})
	}
	return subnets, nil
}

func (c *Client) GetSubnetPublicKey(subnetID principal.Principal) ([]byte, error) {
	v, _, err := c.dp.GetValueUpdate([]byte(fmt.Sprintf("crypto_threshold_signing_public_key_%s", subnetID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subnet public key: %w", err)
	}
	var publicKey v1.PublicKey
	if err := proto.Unmarshal(v, &publicKey); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subnet public key: %w", err)
	}
	if publicKey.Algorithm != v1.AlgorithmId_ALGORITHM_ID_THRES_BLS12_381 {
		return nil, fmt.Errorf("unsupported public key algorithm")
	}
	return certification.PublicKeyToDER(publicKey.KeyValue)
}

type DataCenterMap map[string][]NodeDetails

type IPv4Interface struct {
	Address      string
	Gateways     []string
	PrefixLength uint32
}

type NodeDetails struct {
	IPv6            string
	IPv4            *IPv4Interface
	NodeOperatorID  principal.Principal
	NodeProviderID  principal.Principal
	DCID            string
	HostOSVersionID *string
	Domain          *string
}

type NodeMap map[string]NodeDetails

func (n NodeMap) ByDataCenter() DataCenterMap {
	dcMap := make(map[string][]NodeDetails)
	for _, node := range n {
		if v, ok := dcMap[node.DCID]; ok {
			dcMap[node.DCID] = append(v, node)
		} else {
			dcMap[node.DCID] = []NodeDetails{node}
		}
	}
	return dcMap
}

func (n NodeMap) ByNodeOperator() NodeOperatorMap {
	noMap := make(map[string][]NodeDetails)
	for _, node := range n {
		if v, ok := noMap[node.NodeOperatorID.String()]; ok {
			noMap[node.NodeOperatorID.String()] = append(v, node)
		} else {
			noMap[node.NodeOperatorID.String()] = []NodeDetails{node}
		}
	}
	return noMap
}

func (n NodeMap) ByNodeProvider() NodeProviderMap {
	npMap := make(map[string][]NodeDetails)
	for _, node := range n {
		if v, ok := npMap[node.NodeProviderID.String()]; ok {
			npMap[node.NodeProviderID.String()] = append(v, node)
		} else {
			npMap[node.NodeProviderID.String()] = []NodeDetails{node}
		}
	}
	return npMap
}

type NodeOperatorMap map[string][]NodeDetails

type NodeProviderMap map[string][]NodeDetails
