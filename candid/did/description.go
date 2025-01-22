package did

import (
	"strings"

	"github.com/0x51-dev/upeg/parser"
	"github.com/aviate-labs/agent-go/candid/internal/candid"
)

// Description represents the interface description of a program. An interface description consists of a sequence of
// imports and type definitions, possibly followed by a service declaration.
type Description struct {
	// Definitions is the sequence of import and type definitions.
	Definitions []Definition
	// Services is a list of service declarations.
	Services []Service
}

func ConvertDescription(n *parser.Node) Description {
	var desc Description
	for _, n := range n.Children() {
		switch n.Name {
		case candid.Type.Name:
			desc.Definitions = append(
				desc.Definitions,
				convertType(n),
			)
		case candid.Import.Name:
			desc.Definitions = append(
				desc.Definitions,
				Import{
					Text: "",
				},
			)
		case candid.Actor.Name:
			desc.Services = append(
				desc.Services,
				convertService(n),
			)
		case candid.CommentText.Name:
			// Ignore comments.
		default:
			panic(n)
		}
	}
	return desc
}

func (p Description) String() string {
	var s []string
	for _, d := range p.Definitions {
		s = append(s, d.String())
	}
	for _, a := range p.Services {
		s = append(s, a.String())
	}
	return strings.Join(s, ";\n")
}
