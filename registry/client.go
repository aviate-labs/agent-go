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
		return nil, err
	}
	return &Client{
		dp: dp,
	}, nil
}

func (c *Client) GetNNSSubnetID() (*principal.Principal, error) {
	v, _, err := c.dp.GetValueUpdate([]byte("nns_subnet_id"), nil)
	if err != nil {
		return nil, err
	}
	var nnsSubnetID v1.SubnetId
	if err := proto.Unmarshal(v, &nnsSubnetID); err != nil {
		return nil, err
	}
	return &principal.Principal{Raw: nnsSubnetID.PrincipalId.Raw}, nil
}

func (c *Client) GetNodeListSince(version uint64) (map[string]NodeDetails, error) {
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
			return nil, err
		}
		currentVersion = records[len(records)-1].Version
		for _, record := range records {
			if strings.HasPrefix(record.Key, "node_record_") {
				if record.Value == nil {
					delete(nodeMap, strings.TrimPrefix(record.Key, "node_record_"))
				} else {
					var nodeRecord v1.NodeRecord
					if err := proto.Unmarshal(record.Value, &nodeRecord); err != nil {
						return nil, err
					}
					nodeMap[strings.TrimPrefix(record.Key, "node_record_")] = &nodeRecord
				}
			} else if strings.HasPrefix(record.Key, "node_operator_record_") {
				if record.Value == nil {
					delete(nodeOperatorMap, strings.TrimPrefix(record.Key, "node_operator_record_"))
				} else {
					var nodeOperatorRecord v1.NodeOperatorRecord
					if err := proto.Unmarshal(record.Value, &nodeOperatorRecord); err != nil {
						return nil, err
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
			nodeProviderID = principal.Principal{Raw: no.NodeOperatorPrincipalId}
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
		return nil, err
	}
	var record v1.SubnetRecord
	if err := proto.Unmarshal(v, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (c *Client) GetSubnetIDs() ([]principal.Principal, error) {
	v, _, err := c.dp.GetValueUpdate([]byte("subnet_list"), nil)
	if err != nil {
		return nil, err
	}
	var list v1.SubnetListRecord
	if err := proto.Unmarshal(v, &list); err != nil {
		return nil, err
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
		return nil, err
	}
	var publicKey v1.PublicKey
	if err := proto.Unmarshal(v, &publicKey); err != nil {
		return nil, err
	}
	if publicKey.Algorithm != v1.AlgorithmId_ALGORITHM_ID_THRES_BLS12_381 {
		return nil, fmt.Errorf("unsupported public key algorithm")
	}
	return certification.PublicKeyToDER(publicKey.KeyValue)
}

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
