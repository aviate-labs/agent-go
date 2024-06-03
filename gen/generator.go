package gen

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/aviate-labs/agent-go/candid"
	"github.com/aviate-labs/agent-go/candid/did"
	"io"
	"io/fs"
	"strings"
	"text/template"
)

const (
	templatesDir = "templates"
)

var (
	//go:embed templates/*
	files     embed.FS
	templates map[string]*template.Template
)

func funcName(prefix, name string) string {
	if strings.HasPrefix(name, "\"") {
		name = name[1 : len(name)-1]
	}
	var str string
	for _, p := range strings.Split(name, "_") {
		str += strings.ToUpper(string(p[0])) + p[1:]
	}
	if prefix != "" {
		return fmt.Sprintf("%s.%s", prefix, str)
	}
	return str
}

func init() {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	tmplFiles, err := fs.ReadDir(files, templatesDir)
	if err != nil {
		panic(err)
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}

		pt, err := template.ParseFS(files, templatesDir+"/"+tmpl.Name())
		if err != nil {
			panic(err)
		}

		templates[strings.TrimSuffix(tmpl.Name(), ".gotmpl")] = pt
	}
}

func rawName(name string) string {
	if strings.HasPrefix(name, "\"") {
		return name[1 : len(name)-1]
	}
	return name
}

// Generator is a generator for a given service description.
type Generator struct {
	AgentName          string
	ModulePath         string
	CanisterName       string
	PackageName        string
	ServiceDescription did.Description
	usedIDL            bool

	indirect bool
}

// NewGenerator creates a new generator for the given service description.
func NewGenerator(agentName, canisterName, packageName string, rawDID []byte) (*Generator, error) {
	desc, err := candid.ParseDID(rawDID)
	if err != nil {
		return nil, err
	}
	return &Generator{
		AgentName:          agentName,
		CanisterName:       canisterName,
		PackageName:        packageName,
		ServiceDescription: desc,
	}, nil
}

func (g *Generator) Generate() ([]byte, error) {
	var definitions []agentArgsDefinition
	for _, definition := range g.ServiceDescription.Definitions {
		switch definition := definition.(type) {
		case did.Type:
			typ := g.dataToString("", definition.Data)
			definitions = append(definitions, agentArgsDefinition{
				Name: funcName("", definition.Id),
				Type: typ,
				Eq:   !strings.HasPrefix(typ, "struct"),
			})
		}
	}

	var methods []agentArgsMethod
	for _, service := range g.ServiceDescription.Services {
		for _, method := range service.Methods {
			name := rawName(method.Name)
			f := method.Func

			var argumentTypes []agentArgsMethodArgument
			for i, t := range f.ArgTypes {
				var n string
				if (t.Name != nil) && (*t.Name != "") {
					n = *t.Name
				} else {
					n = fmt.Sprintf("arg%d", i)
				}
				argumentTypes = append(argumentTypes, agentArgsMethodArgument{
					Name: n,
					Type: g.dataToString("", t.Data),
				})
			}

			var returnTypes []string
			for _, t := range f.ResTypes {
				returnTypes = append(returnTypes, g.dataToString("", t.Data))
			}

			typ := "Call"
			if f.Annotation != nil && *f.Annotation == did.AnnQuery {
				typ = "Query"
			}

			methods = append(methods, agentArgsMethod{
				RawName:       name,
				Name:          funcName("", name),
				Type:          typ,
				ArgumentTypes: argumentTypes,
				ReturnTypes:   returnTypes,
			})
		}
	}
	tmplName := "agent"
	if g.indirect {
		tmplName = "agent_indirect"
	}
	t, ok := templates[tmplName]
	if !ok {
		return nil, fmt.Errorf("template not found")
	}
	var tmpl bytes.Buffer
	if err := t.Execute(&tmpl, agentArgs{
		AgentName:    g.AgentName,
		CanisterName: g.CanisterName,
		PackageName:  g.PackageName,
		UsedIDL:      g.usedIDL,
		Definitions:  definitions,
		Methods:      methods,
	}); err != nil {
		return nil, err
	}
	return io.ReadAll(&tmpl)
}

// Indirect sets the generator to generate indirect calls.
func (g *Generator) Indirect() *Generator {
	g.indirect = true
	return g
}

func (g *Generator) dataToString(prefix string, data did.Data) string {
	switch t := data.(type) {
	case did.Blob:
		return "[]byte"
	case did.DataId:
		return funcName(prefix, string(t))
	case did.Func:
		return "struct { /* NOT SUPPORTED */ }"
	case did.Optional:
		return fmt.Sprintf("*%s", g.dataToString(prefix, t.Data))
	case did.Primitive:
		switch t {
		case "nat8", "nat16", "nat32", "nat64":
			return strings.ReplaceAll(data.String(), "nat", "uint")
		case "bool", "float32", "float64", "int8", "int16", "int32", "int64":
			return data.String()
		case "int":
			g.usedIDL = true
			return "idl.Int"
		case "nat":
			g.usedIDL = true
			return "idl.Nat"
		case "text":
			return "string"
		case "null":
			g.usedIDL = true
			return "idl.Null"
		default:
			panic(fmt.Sprintf("unknown primitive: %s", t))
		}
	case did.Principal:
		return "principal.Principal"
	case did.Record:
		var sizeName int
		var sizeType int
		var records []struct {
			originalName string
			name         string
			typ          string
		}
		var tuple = true
		for i, field := range t {
			originalName := fmt.Sprintf("Field%d", i)
			name := originalName
			if n := field.Name; n != nil {
				originalName = *n
				name = funcName("", *n)
				tuple = false
			}
			if l := len(name); l > sizeName {
				sizeName = l
			}
			var typ string
			if field.Data != nil {
				typ = g.dataToString(prefix, *field.Data)
			} else {
				typ = funcName(prefix, *field.NameData)
			}
			for _, typ := range strings.Split(typ, "\n") {
				if l := len(typ); l > sizeType {
					sizeType = l
				}
			}
			records = append(records, struct {
				originalName string
				name         string
				typ          string
			}{
				originalName: originalName,
				name:         name,
				typ:          typ,
			})
		}
		var record string
		for _, r := range records {
			tag := r.originalName
			if tuple {
				tag = strings.TrimPrefix(tag, "Field")
			}
			if strings.HasPrefix(r.typ, "*") {
				tag += ",omitempty" // optional value.
			}
			record += fmt.Sprintf("\t%-*s %-*s `ic:\"%s\" json:\"%s\"`\n", sizeName, r.name, sizeType, r.typ, tag, tag)
		}
		return fmt.Sprintf("struct {\n%s}", record)
	case did.Variant:
		var sizeName int
		var sizeType int
		var records []struct {
			originalName string
			name         string
			typ          string
		}
		for _, field := range t {
			if field.Name == nil {
				name := *field.NameData
				if l := len(name); l > sizeName {
					sizeName = l
				}
				if 8 > sizeType {
					sizeType = 8
				}
				g.usedIDL = true
				records = append(records, struct {
					originalName string
					name         string
					typ          string
				}{originalName: name, name: funcName("", name), typ: "idl.Null"})
			} else {
				name := funcName("", *field.Name)
				if l := len(name); l > sizeName {
					sizeName = l
				}
				var typ string
				if field.Data != nil {
					typ = g.dataToString(prefix, *field.Data)
				} else {
					typ = funcName(prefix, *field.NameData)
				}
				for _, typ := range strings.Split(typ, "\n") {
					if l := len(typ); l > sizeType {
						sizeType = l
					}
				}
				records = append(records, struct {
					originalName string
					name         string
					typ          string
				}{
					originalName: *field.Name,
					name:         name,
					typ:          typ,
				})
			}
		}
		var record string
		for _, r := range records {
			record += fmt.Sprintf("\t%-*s *%-*s `ic:\"%s,variant\"`\n", sizeName, r.name, sizeType, r.typ, r.originalName)
		}
		return fmt.Sprintf("struct {\n%s}", record)
	case did.Vector:
		return fmt.Sprintf("[]%s", g.dataToString(prefix, t.Data))
	default:
		panic(fmt.Sprintf("unknown type: %T", t))
	}
}

type agentArgs struct {
	AgentName    string
	CanisterName string
	PackageName  string
	UsedIDL      bool
	Definitions  []agentArgsDefinition
	Methods      []agentArgsMethod
}

type agentArgsDefinition struct {
	Name string
	Type string
	Eq   bool
}

type agentArgsMethod struct {
	RawName             string
	Name                string
	Type                string
	ArgumentTypes       []agentArgsMethodArgument
	FilledArgumentTypes []agentArgsMethodArgument
	ReturnTypes         []string
}

type agentArgsMethodArgument struct {
	Name string
	Type string
}
