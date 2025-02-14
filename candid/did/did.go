package did

import "github.com/aviate-labs/agent-go/candid/internal/candid"

// ParseDID parses the given raw .did files and returns the Program that is defined in it.
func ParseDID(raw []rune) (*Description, error) {
	p, err := candid.NewParser(raw)
	if err != nil {
		return nil, err
	}
	n, err := p.ParseEOF(candid.Prog)
	if err != nil {
		return nil, err
	}
	did := ConvertDescription(n)
	return &did, nil
}
