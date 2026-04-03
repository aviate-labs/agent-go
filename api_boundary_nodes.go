package agent

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"sync/atomic"

	"github.com/niccolofant/agent-go/certification"
	"github.com/niccolofant/agent-go/certification/hashtree"
	"github.com/niccolofant/agent-go/principal"
)

// DiscoverRoutes enumerates API boundary nodes on-chain and returns their
// https://<domain> URLs. Pair with RoundRobinRoute or RandomRoute to build a
// RouteProvider, then call Agent.Client().SetRouteProvider to use it.
//
// Example:
//
//	a, _ := agent.New(agent.Config{})
//	hosts, _ := agent.DiscoverRoutes(a)
//	rp, _ := agent.RoundRobinRoute(hosts)
//	a.Client().SetRouteProvider(rp)
func DiscoverRoutes(a *Agent) ([]*url.URL, error) {
	nodes, err := a.GetAPIBoundaryNodes()
	if err != nil {
		return nil, fmt.Errorf("api boundary node discovery: %w", err)
	}
	if len(nodes) == 0 {
		return nil, errors.New("api boundary node discovery: no nodes published on-chain")
	}
	var hosts []*url.URL
	for _, n := range nodes {
		if n.Domain == "" {
			continue
		}
		u, err := url.Parse("https://" + n.Domain)
		if err != nil {
			return nil, fmt.Errorf("api boundary node discovery: parse %q: %w", n.Domain, err)
		}
		hosts = append(hosts, u)
	}
	if len(hosts) == 0 {
		return nil, errors.New("api boundary node discovery: no nodes have a domain")
	}
	return hosts, nil
}

// APIBoundaryNode describes a single API boundary node as published in the
// /api_boundary_nodes/<node_id> sub-tree of the IC state.
//
// IPv4Address and IPv6Address are UTF-8 strings ("192.168.10.150" /
// "3002:0bd6:..."). Either may be empty if the node does not publish that
// address family.
type APIBoundaryNode struct {
	NodeID      principal.Principal
	Domain      string
	IPv4Address string
	IPv6Address string
}

// GetAPIBoundaryNodes enumerates the API boundary nodes published on-chain.
// This is the authoritative way to discover boundary nodes; hardcoding
// icp0.io/ic0.app is a fallback for bootstrapping only.
func (a Agent) GetAPIBoundaryNodes() ([]APIBoundaryNode, error) {
	root := []hashtree.Label{hashtree.Label("api_boundary_nodes")}
	cert, err := a.readSubnetStateCertificate(
		principal.MustDecode(certification.RootSubnetID),
		[][]hashtree.Label{root},
	)
	if err != nil {
		return nil, err
	}
	tree, err := cert.Tree.LookupSubTree(root...)
	if err != nil {
		var lookupErr hashtree.LookupError
		if errors.As(err, &lookupErr) && lookupErr.Type == hashtree.LookupResultAbsent {
			return nil, nil
		}
		return nil, err
	}
	children, err := hashtree.AllChildren(tree)
	if err != nil {
		return nil, err
	}
	var nodes []APIBoundaryNode
	for _, child := range children {
		node := APIBoundaryNode{NodeID: principal.Principal{Raw: child.Path[0]}}
		if d, err := hashtree.Lookup(child.Value, hashtree.Label("domain")); err == nil {
			node.Domain = string(d)
		}
		if v4, err := hashtree.Lookup(child.Value, hashtree.Label("ipv4_address")); err == nil {
			node.IPv4Address = string(v4)
		}
		if v6, err := hashtree.Lookup(child.Value, hashtree.Label("ipv6_address")); err == nil {
			node.IPv6Address = string(v6)
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// RouteProvider supplies the boundary-node URL for each outgoing request.
// Implementations are called once per HTTP request and must be safe for
// concurrent use.
type RouteProvider interface {
	// Route returns the host URL (scheme + host, no path) for the next request.
	Route() (*url.URL, error)
}

// RandomRoute returns a RouteProvider that picks a uniformly random host on
// each call using crypto/rand.
func RandomRoute(hosts []*url.URL) (RouteProvider, error) {
	if len(hosts) == 0 {
		return nil, errors.New("random route: no hosts")
	}
	return randomRoute{hosts: hosts}, nil
}

// RoundRobinRoute returns a RouteProvider that cycles through the given hosts
// in order, wrapping at the end. Safe for concurrent use.
func RoundRobinRoute(hosts []*url.URL) (RouteProvider, error) {
	if len(hosts) == 0 {
		return nil, errors.New("round-robin route: no hosts")
	}
	return &roundRobinRoute{hosts: hosts}, nil
}

// StaticRoute returns a RouteProvider that always serves the given URL.
func StaticRoute(host *url.URL) RouteProvider {
	return staticRoute{host: host}
}

type randomRoute struct{ hosts []*url.URL }

func (r randomRoute) Route() (*url.URL, error) {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(len(r.hosts))))
	if err != nil {
		return nil, fmt.Errorf("rand: %w", err)
	}
	return r.hosts[v.Int64()], nil
}

type roundRobinRoute struct {
	hosts []*url.URL
	idx   atomic.Uint64
}

func (r *roundRobinRoute) Route() (*url.URL, error) {
	i := r.idx.Add(1) - 1
	return r.hosts[int(i%uint64(len(r.hosts)))], nil
}

type staticRoute struct{ host *url.URL }

func (s staticRoute) Route() (*url.URL, error) { return s.host, nil }
