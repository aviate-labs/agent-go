package gen

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/aviate-labs/agent-go/candid"
	"github.com/aviate-labs/agent-go/candid/did"
	"github.com/aviate-labs/agent-go/candid/idl"
	"io"
	"io/fs"
	"math/rand"
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

func (g *Generator) GenerateActor() ([]byte, error) {
	definitions := make(map[string]did.Data)
	for _, definition := range g.ServiceDescription.Definitions {
		switch definition := definition.(type) {
		case did.Type:
			definitions[definition.Id] = definition.Data
		}
	}

	var methods []actorArgsMethod
	for _, service := range g.ServiceDescription.Services {
		for _, method := range service.Methods {
			name := rawName(method.Name)
			f := method.Func

			var argumentTypes []agentArgsMethodArgument
			for i, t := range f.ArgTypes {
				argumentTypes = append(argumentTypes, agentArgsMethodArgument{
					Name: fmt.Sprintf("_arg%d", i),
					Type: g.dataToMotokoString(t.Data),
				})
			}

			var returnTypes []string
			for _, t := range f.ResTypes {
				returnTypes = append(returnTypes, g.dataToMotokoString(t.Data))
			}

			r := rand.NewSource(idl.Hash(g.CanisterName).Int64())
			var returnValues []string
			for _, t := range f.ResTypes {
				returnValues = append(returnValues, g.dataToMotokoReturnValue(r, definitions, t.Data))
			}

			typ := "shared"
			if f.Annotation != nil && *f.Annotation == did.AnnQuery {
				typ = "query"
			}

			methods = append(methods, actorArgsMethod{
				Name:          name,
				Type:          typ,
				ArgumentTypes: argumentTypes,
				ReturnTypes:   returnTypes,
				ReturnValues:  returnValues,
			})
		}
	}
	t, ok := templates["actor_mo"]
	if !ok {
		return nil, fmt.Errorf("template not found")
	}
	var tmpl bytes.Buffer
	if err := t.Execute(&tmpl, actorArgs{
		CanisterName: g.CanisterName,
		Methods:      methods,
	}); err != nil {
		return nil, err
	}
	return io.ReadAll(&tmpl)
}

func (g *Generator) GenerateActorTypes() ([]byte, error) {
	var definitions []agentArgsDefinition
	for _, definition := range g.ServiceDescription.Definitions {
		switch definition := definition.(type) {
		case did.Type:
			definitions = append(definitions, agentArgsDefinition{
				Name: funcName("", definition.Id),
				Type: g.dataToMotokoString(definition.Data),
			})
		}
	}
	t, ok := templates["types_mo"]
	if !ok {
		return nil, fmt.Errorf("template not found")
	}
	var tmpl bytes.Buffer
	if err := t.Execute(&tmpl, actorTypesArgs{
		Definitions: definitions,
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

func (g *Generator) dataToMotokoReturnValue(s rand.Source, definitions map[string]did.Data, data did.Data) string {
	r := rand.New(s)
	switch t := data.(type) {
	case did.DataId:
		return g.dataToMotokoReturnValue(s, definitions, definitions[string(t)])
	case did.Blob:
		var b [32]byte
		r.Read(b[:])
		return fmt.Sprintf("\"x%02X\"", b)
	case did.Func:
		return "{ /* func */ }"
	case did.Principal:
		var b [32]byte
		r.Read(b[:])
		return fmt.Sprintf("principalOfBlob(\"x%02X\")", b[:])
	case did.Primitive:
		switch t {
		case "null":
			return "()"
		case "bool":
			return fmt.Sprintf("%t", r.Int()%2 == 0)
		case "nat", "int":
			return fmt.Sprintf("%d", r.Uint64())
		case "nat8":
			return fmt.Sprintf("%d", r.Uint64()%0xFF)
		case "nat16":
			return fmt.Sprintf("%d", r.Uint64()%0xFFFF)
		case "nat32":
			return fmt.Sprintf("%d", r.Uint64()%0xFFFFFFFF)
		case "nat64":
			return fmt.Sprintf("%d", r.Uint64()%0xFFFFFFFFFFFFFFFF)
		case "text":
			return fmt.Sprintf("\"%d\"", r.Uint64())
		}
	case did.Vector:
		n := r.Int() % 10
		var values []string
		for i := 0; i < n; i++ {
			values = append(values, g.dataToMotokoReturnValue(s, definitions, t.Data))
		}
		return fmt.Sprintf("[ %s ]", strings.Join(values, ", "))
	case did.Record:
		var tuple bool
		var fields []string
		for _, v := range t {
			if v.Name != nil {
				var data string
				if v.NameData != nil {
					data = g.dataToMotokoReturnValue(s, definitions, definitions[*v.NameData])
				} else {
					data = g.dataToMotokoReturnValue(s, definitions, *v.Data)
				}
				fields = append(fields, fmt.Sprintf("%s = %s", *v.Name, data))
			} else {
				tuple = true
				break
			}
		}
		if !tuple {
			return fmt.Sprintf("{ %s }", strings.Join(fields, "; "))
		}

		var values []string
		for _, field := range t {
			if field.Data != nil {
				values = append(values, g.dataToMotokoReturnValue(s, definitions, *field.Data))
			} else {
				values = append(values, g.dataToMotokoReturnValue(s, definitions, definitions[*field.NameData]))
			}
		}
		return fmt.Sprintf("( %s )", strings.Join(values, ", "))
	case did.Variant:
		r := s.Int63() % int64(len(t))
		field := t[r]
		if field.Data != nil {
			return fmt.Sprintf("#%s(%s)", *field.Name, g.dataToMotokoReturnValue(s, definitions, *field.Data))
		}
		if field.Name != nil {
			return fmt.Sprintf("#%s(%s)", *field.Name, g.dataToMotokoReturnValue(s, definitions, definitions[*field.NameData]))
		}
		if data := definitions[*field.NameData]; data != nil {
			return fmt.Sprintf("#%s(%s)", *field.NameData, g.dataToMotokoReturnValue(s, definitions, definitions[*field.NameData]))
		}
		return fmt.Sprintf("#%s", *field.NameData)
	case did.Optional:
		return fmt.Sprintf("?%s", g.dataToMotokoReturnValue(s, definitions, t.Data))
	}
	return fmt.Sprintf("%q # %q", "UNKNOWN", data)
}

func (g *Generator) dataToMotokoString(data did.Data) string {
	switch t := data.(type) {
	case did.Blob:
		return "Blob"
	case did.Func:
		return "{ /* func */ }"
	case did.Record:
		var fields []string
		for _, v := range t {
			if v.Name != nil {
				var data string
				if v.NameData != nil {
					data = fmt.Sprintf("T.%s", funcName("", *v.NameData))
				} else {
					data = g.dataToMotokoString(*v.Data)
				}
				fields = append(fields, fmt.Sprintf("%s : %s", *v.Name, data))
			} else {
				if v.NameData != nil {
					fields = append(fields, fmt.Sprintf("T.%s", funcName("", *v.NameData)))
				} else {
					fields = append(fields, g.dataToMotokoString(*v.Data))
				}
			}
		}
		var isTuple bool
		for _, f := range fields {
			if !strings.Contains(f, ":") {
				isTuple = true
				break
			}
		}
		if isTuple {
			return fmt.Sprintf("(%s)", strings.Join(fields, ", "))
		}
		return fmt.Sprintf("{ %s }", strings.Join(fields, "; "))
	case did.Variant:
		var variants []string
		for _, v := range t {
			if v.Name != nil {
				var data string
				if v.NameData != nil {
					data = fmt.Sprintf("T.%s", funcName("", *v.NameData))
				} else {
					data = g.dataToMotokoString(*v.Data)
				}
				variants = append(variants, fmt.Sprintf("#%s : %s", *v.Name, data))
			} else {
				variants = append(variants, fmt.Sprintf("#%s", *v.NameData))
			}
		}
		return fmt.Sprintf("{ %s }", strings.Join(variants, "; "))
	case did.Vector:
		return fmt.Sprintf("[%s]", g.dataToMotokoString(t.Data))
	case did.Optional:
		return fmt.Sprintf("?%s", g.dataToMotokoString(t.Data))
	case did.Primitive:
		switch t {
		case "nat", "nat8", "nat16", "nat32", "nat64":
			return strings.ReplaceAll(data.String(), "nat", "Nat")
		case "int", "int8", "int16", "int32", "int64":
			return strings.ReplaceAll(data.String(), "int", "Int")
		case "text":
			return "Text"
		case "bool":
			return "Bool"
		case "null":
			return "()"
		default:
			return t.String()
		}
	case did.Principal:
		return "Principal"
	case did.DataId:
		return fmt.Sprintf("T.%s", funcName("", string(t)))
	default:
		panic(fmt.Sprintf("unknown type: %T", t))
	}
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

type actorArgs struct {
	CanisterName string
	Methods      []actorArgsMethod
}

type actorArgsMethod struct {
	Name          string
	Type          string
	ArgumentTypes []agentArgsMethodArgument
	ReturnTypes   []string
	ReturnValues  []string
}

type actorTypesArgs struct {
	Definitions []agentArgsDefinition
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
