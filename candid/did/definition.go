package did

import (
	"fmt"
	"strings"

	"github.com/0x51-dev/upeg/parser"
)

// Definition represents an imports or type definition.
type Definition interface {
	def()
	fmt.Stringer
}

// Import represents an import declaration from another file. Service is true for
// `import service "..."`, which also merges the imported file's main service.
type Import struct {
	Text    string
	Service bool
}

func convertImport(n *parser.Node) Import {
	var imp Import
	for _, c := range n.Children() {
		switch c.Name {
		case "ImportService":
			imp.Service = true
		case "Text":
			imp.Text = strings.Trim(c.Value(), `"`)
		}
	}
	return imp
}

func (i Import) String() string {
	if i.Service {
		return fmt.Sprintf("import service %q", i.Text)
	}
	return fmt.Sprintf("import %q", i.Text)
}

func (i Import) def() {}

// Type represents a named type definition.
type Type struct {
	Id   string
	Data Data
}

func convertType(n *parser.Node) Type {
	cs := n.Children()
	var (
		id   = cs[0]
		data = cs[len(cs)-1]
	)
	return Type{
		Id:   id.Value(),
		Data: convertData(data),
	}
}

func (t Type) String() string {
	return fmt.Sprintf("type %s = %s", t.Id, t.Data.String())
}

func (t Type) def() {}
