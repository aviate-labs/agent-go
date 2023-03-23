package did

import (
	"strings"

	"github.com/aviate-labs/agent-go/candid/internal/candid"
	"github.com/di-wu/parser/ast"
)

// Description represents the interface description of a program. An interface description consists of a sequence of
// imports and type definitions, possibly followed by a service declaration.
type Description struct {
	// Definitions is the sequence of import and type definitions.
	Definitions []Definition
	// Services is a list of service declarations.
	Services []Service
}

func ConvertDescription(n *ast.Node) Description {
	var desc Description
	for _, n := range n.Children() {
		switch n.Type {
		case candid.TypeT:
			desc.Definitions = append(
				desc.Definitions,
				convertType(n),
			)
		case candid.ImportT:
			desc.Definitions = append(
				desc.Definitions,
				Import{
					Text: "",
				},
			)
		case candid.ActorT:
			desc.Services = append(
				desc.Services,
				convertService(n),
			)
		case candid.CommentTextT:
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
