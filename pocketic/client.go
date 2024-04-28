package pocketic

import "github.com/aviate-labs/agent-go/principal"

// PocketIC is a client for the local PocketIC server.
type PocketIC struct {
	server     *server
	instanceID int
	topology   map[string]Topology
	sender     principal.Principal
}

type Config struct {
	subnetConfig ExtendedSubnetConfigSet
}

type Option func(*Config)

func WithSubnetConfig(config ExtendedSubnetConfigSet) Option {
	return func(p *Config) {
		p.subnetConfig = config
	}
}

// New creates a new PocketIC client.
func New(opts ...Option) (*PocketIC, error) {
	config := Config{
		subnetConfig: DefaultSubnetConfig,
	}
	for _, fn := range opts {
		fn(&config)
	}

	s, err := newServer()
	if err != nil {
		return nil, err
	}
	resp, err := s.NewInstance(config.subnetConfig)
	if err != nil {
		return nil, err
	}
	return &PocketIC{
		server:     s,
		instanceID: resp.InstanceID,
		topology:   resp.Topology,
		sender:     principal.AnonymousID,
	}, nil
}
