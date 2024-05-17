package registry

import (
	"bytes"
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
	v1 "github.com/aviate-labs/agent-go/registry/proto/v1"
	"google.golang.org/protobuf/proto"
	"sort"
	"strings"
)

type Client struct {
	dp     *DataProvider
	deltas []*v1.RegistryDelta
}

func New() (*Client, error) {
	dp, err := NewDataProvider()
	if err != nil {
		return nil, err
	}
	deltas, _, err := dp.GetChangesSince(0)
	if err != nil {
		return nil, err
	}
	sort.Slice(deltas, func(i, j int) bool {
		return 0 < bytes.Compare(deltas[i].Key, deltas[j].Key)
	})
	return &Client{
		dp:     dp,
		deltas: deltas,
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

func (c *Client) GetNodeList() ([]*v1.NodeRecord, error) {
	var nodes []*v1.NodeRecord
	for _, delta := range c.deltas {
		key := string(delta.Key)
		if strings.HasPrefix(key, "node_record_") {
			for _, value := range delta.Values {
				record := new(v1.NodeRecord)
				if err := proto.Unmarshal(value.Value, record); err != nil {
					return nil, err
				}
				nodes = append(nodes, record)
			}
		}
	}
	return nodes, nil
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
