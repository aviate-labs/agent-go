package did

import (
	"fmt"

	"github.com/0x51-dev/upeg/parser"
	"github.com/aviate-labs/agent-go/candid/idl"
)

func ConvertDataNode(n *parser.Node) Data {
	return convertData(n)
}

// TypeEnv resolves named type references while converting to idl types.
type TypeEnv struct {
	defs   map[string]Data
	rec    map[string]*idl.RecursiveType // in progress
	done   map[string]idl.Type
	cyclic map[string]bool
}

func NewTypeEnv(defs map[string]Data) *TypeEnv {
	return &TypeEnv{
		defs:   defs,
		rec:    make(map[string]*idl.RecursiveType),
		done:   make(map[string]idl.Type),
		cyclic: make(map[string]bool),
	}
}

// IDLType converts a Data to its idl.Type. A reference reached while its own
// definition is still being converted becomes a RecursiveType, which is what
// makes `type Opt = opt Opt` terminate.
func (e *TypeEnv) IDLType(d Data) (idl.Type, error) {
	switch d := d.(type) {
	case Primitive:
		return primitiveIDLType(string(d))
	case Blob:
		return idl.NewVectorType(idl.Nat8Type()), nil
	case Principal:
		return new(idl.PrincipalType), nil
	case Optional:
		inner, err := e.IDLType(d.Data)
		if err != nil {
			return nil, err
		}
		return idl.NewOptionalType(inner), nil
	case Vector:
		inner, err := e.IDLType(d.Data)
		if err != nil {
			return nil, err
		}
		return idl.NewVectorType(inner), nil
	case Record:
		fields, err := e.fields(d)
		if err != nil {
			return nil, err
		}
		return idl.NewRecordType(fields), nil
	case Variant:
		fields, err := e.fields(d)
		if err != nil {
			return nil, err
		}
		return idl.NewVariantType(fields), nil
	case Func:
		return e.funcType(d)
	case Service:
		var ms []idl.Method
		for _, m := range d.Methods {
			if m.Func == nil {
				continue
			}
			f, err := e.funcType(*m.Func)
			if err != nil {
				return nil, err
			}
			ms = append(ms, idl.Method{Name: m.Name, Func: f.(*idl.FunctionType)})
		}
		return &idl.ServiceType{Methods: ms}, nil
	case DataId:
		return e.resolve(string(d))
	default:
		return nil, fmt.Errorf("unsupported data type: %T", d)
	}
}

func (e *TypeEnv) resolve(name string) (idl.Type, error) {
	if r, ok := e.rec[name]; ok {
		e.cyclic[name] = true
		return r, nil
	}
	if t, ok := e.done[name]; ok {
		return t, nil
	}
	def, ok := e.defs[name]
	if !ok {
		return nil, fmt.Errorf("undefined type: %s", name)
	}
	r := idl.NewRecursiveType(name)
	e.rec[name] = r
	inner, err := e.IDLType(def)
	delete(e.rec, name)
	if err != nil {
		return nil, err
	}
	r.SetInner(inner)
	// Only a definition that referred back to itself needs the indirection.
	t := idl.Type(r)
	if !e.cyclic[name] {
		t = inner
	}
	e.done[name] = t
	return t, nil
}

// fields keys members by label; an unnamed one takes its positional index.
func (e *TypeEnv) fields(fs []Field) (map[string]idl.Type, error) {
	m := make(map[string]idl.Type, len(fs))
	for i, f := range fs {
		var label string
		switch {
		case f.Name != nil:
			label = *f.Name
		case f.Nat != nil:
			label = f.Nat.String()
		case f.NatData != nil:
			label = f.NatData.String()
		case f.NameData != nil && e.defs[*f.NameData] == nil:
			// A lone name nothing defines is a payload-less tag: `variant {nil}`.
			label = *f.NameData
		default:
			label = fmt.Sprint(i)
		}

		// A named type reference parses into NameData, alongside the label.
		if f.Data == nil && f.NameData != nil && e.defs[*f.NameData] != nil {
			t, err := e.resolve(*f.NameData)
			if err != nil {
				return nil, err
			}
			m[label] = t
			continue
		}
		if f.Data == nil {
			m[label] = new(idl.NullType)
			continue
		}
		t, err := e.IDLType(*f.Data)
		if err != nil {
			return nil, err
		}
		m[label] = t
	}
	return m, nil
}

func primitiveIDLType(p string) (idl.Type, error) {
	switch p {
	case "bool":
		return new(idl.BoolType), nil
	case "text":
		return new(idl.TextType), nil
	case "null":
		return new(idl.NullType), nil
	case "reserved":
		return new(idl.ReservedType), nil
	case "empty":
		return new(idl.EmptyType), nil
	case "nat":
		return new(idl.NatType), nil
	case "nat8":
		return idl.Nat8Type(), nil
	case "nat16":
		return idl.Nat16Type(), nil
	case "nat32":
		return idl.Nat32Type(), nil
	case "nat64":
		return idl.Nat64Type(), nil
	case "int":
		return new(idl.IntType), nil
	case "int8":
		return idl.Int8Type(), nil
	case "int16":
		return idl.Int16Type(), nil
	case "int32":
		return idl.Int32Type(), nil
	case "int64":
		return idl.Int64Type(), nil
	case "float32":
		return idl.Float32Type(), nil
	case "float64":
		return idl.Float64Type(), nil
	default:
		return nil, fmt.Errorf("unsupported primitive: %s", p)
	}
}

func (e *TypeEnv) funcType(f Func) (idl.Type, error) {
	args, err := e.params(f.ArgTypes)
	if err != nil {
		return nil, err
	}
	rets, err := e.params(f.ResTypes)
	if err != nil {
		return nil, err
	}
	var anns []string
	if f.Annotation != nil {
		anns = append(anns, string(*f.Annotation))
	}
	return idl.NewFunctionType(args, rets, anns), nil
}

func (e *TypeEnv) params(t Tuple) ([]idl.FunctionParameter, error) {
	ps := make([]idl.FunctionParameter, 0, len(t))
	for _, a := range t {
		v, err := e.IDLType(a.Data)
		if err != nil {
			return nil, err
		}
		ps = append(ps, idl.FunctionParameter{Type: v})
	}
	return ps, nil
}
