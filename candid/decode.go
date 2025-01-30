package candid

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/leb128"
)

func Decode(bs []byte) ([]idl.Type, []any, error) {
	if len(bs) == 0 {
		return nil, nil, &idl.FormatError{
			Description: "empty",
		}
	}

	r := bytes.NewReader(bs)

	if err := checkHeader(r); err != nil {
		return nil, nil, err
	}

	var tds []idl.Type
	{ // T
		tdtl, err := leb128.DecodeUnsigned(r)
		if err != nil {
			return nil, nil, err
		}

		var tc typeCache
		for i := 0; i < int(tdtl.Int64()); i++ {
			tid, err := leb128.DecodeSigned(r)
			if err != nil {
				return nil, nil, err
			}
			switch o := idl.OpCode(tid.Int64()); o {
			case idl.OptOpCode:
				typ, err := tc.decodeOptOpCode(r, tds)
				if err != nil {
					return nil, nil, err
				}
				tds = append(tds, typ)
			case idl.VecOpCode:
				typ, err := tc.decodeVecOpCode(r, tds)
				if err != nil {
					return nil, nil, err
				}
				tds = append(tds, typ)
			case idl.RecOpCode:
				typ, err := tc.decodeRecOpCode(r, tds)
				if err != nil {
					return nil, nil, err
				}
				tds = append(tds, typ)
			case idl.VarOpCode:
				typ, err := tc.decodeVarOpCode(r, tds)
				if err != nil {
					return nil, nil, err
				}
				tds = append(tds, typ)
			case idl.FuncOpCode:
				typ, err := tc.decodeFuncOpCode(r, tds)
				if err != nil {
					return nil, nil, err
				}
				tds = append(tds, typ)
			case idl.ServiceOpCode:
				typ, err := tc.decodeServiceOpCode(r, tds)
				if err != nil {
					return nil, nil, err
				}
				tds = append(tds, typ)
			}
		}

		if err := tc.resolve(tds); err != nil {
			return nil, nil, err
		}

		// Resolve records, variants and function indices.
		for _, tb := range tds {
			switch t := tb.(type) {
			case *idl.VariantType:
				resolved := true
				for _, f := range t.Fields {
					if f.Type == nil {
						resolved = false
					}
				}
				if resolved {
					continue
				}

				f := func(tds []idl.Type) (idl.Type, error) {
					for i, f := range t.Fields {
						if f.Type != nil {
							continue
						}
						o := idl.OpCode(f.Index)
						v, err := o.GetType(tds)
						if err != nil {
							return nil, err
						}
						t.Fields[i].Type = v
					}
					return t, nil
				}
				if v, err := f(tds); v == nil || err != nil {
					return nil, nil, fmt.Errorf("unable to resolve variant: %v", t)
				}
			case *idl.RecordType:
				resolved := true
				for _, f := range t.Fields {
					if f.Type == nil {
						resolved = false
					}
				}
				if resolved {
					continue
				}

				f := func(tds []idl.Type) (idl.Type, error) {
					for i, f := range t.Fields {
						if f.Type != nil {
							continue
						}
						o := idl.OpCode(f.Index)
						v, err := o.GetType(tds)
						if err != nil {
							return nil, err
						}
						t.Fields[i].Type = v
					}
					return t, nil
				}
				if v, err := f(tds); v == nil || err != nil {
					return nil, nil, fmt.Errorf("unable to resolve record: %v", t)
				}
			case *idl.FunctionType:
				resolved := true
				for _, f := range t.ArgumentParameters {
					if f.Type == nil {
						resolved = false
					}
				}
				for _, f := range t.ReturnParameters {
					if f.Type == nil {
						resolved = false
					}
				}
				if resolved {
					continue
				}
				f := func(tds []idl.Type) (idl.Type, error) {
					for i, f := range t.ArgumentParameters {
						if f.Type != nil {
							continue
						}
						o := idl.OpCode(f.Index)
						v, err := o.GetType(tds)
						if err != nil {
							return nil, err
						}
						t.ArgumentParameters[i].Type = v
					}
					for i, f := range t.ReturnParameters {
						if f.Type != nil {
							continue
						}
						o := idl.OpCode(f.Index)
						v, err := o.GetType(tds)
						if err != nil {
							return nil, err
						}
						t.ReturnParameters[i].Type = v
					}
					return t, nil
				}
				if v, err := f(tds); v == nil || err != nil {
					return nil, nil, fmt.Errorf("unable to resolve func: %v", t)
				}
			}
		}
	}

	tsl, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, nil, err
	}

	var ts []idl.Type
	{ // I
		for i := 0; i < int(tsl.Int64()); i++ {
			tid, err := leb128.DecodeSigned(r)
			if err != nil {
				return nil, nil, err
			}
			o := idl.OpCode(tid.Int64())
			t, err := o.GetType(tds)
			if err != nil {
				return nil, nil, err
			}
			ts = append(ts, t)
		}
	}

	var vs []any
	{ // M
		for i := 0; i < int(tsl.Int64()); i++ {
			v, err := ts[i].Decode(r)
			if err != nil {
				return nil, nil, err
			}
			vs = append(vs, v)
		}
	}

	if r.Len() != 0 {
		return nil, nil, fmt.Errorf("too long")
	}
	return ts, vs, nil
}

func Unmarshal(data []byte, values []any) error {
	ts, vs, err := Decode(data)
	if err != nil {
		return err
	}
	if len(ts) != len(vs) {
		return fmt.Errorf("unequal data types and value lengths: %d %d", len(ts), len(vs))
	}

	if len(vs) != len(values) {
		return fmt.Errorf("unequal value lengths: %d %d", len(vs), len(values))
	}

	for i, v := range values {
		if err := ts[i].UnmarshalGo(vs[i], v); err != nil {
			return err
		}
	}

	return nil
}

func checkHeader(r *bytes.Reader) error {
	magic := make([]byte, 4)
	n, err := r.Read(magic)
	if err != nil {
		return err
	}
	if n < 4 {
		return &idl.FormatError{
			Description: "no magic bytes",
		}
	}
	if !bytes.Equal(magic, []byte{'D', 'I', 'D', 'L'}) {
		return &idl.FormatError{
			Description: "wrong magic bytes",
		}
	}
	return nil
}

type delayType struct {
	// index is the index of the type in the type list.
	index int
	// f is a function that takes the type list and returns the resolved type.
	f func(tdt []idl.Type) (idl.Type, error)
}

// typeCache is a cache of types that are not yet fully decoded.
// It is used to resolve recursive types.
// - int is the index of the type in the type list.
// - []delayType is a list of type that depend on the type to be resolved.
type typeCache struct {
	cache []delayType
}

func (tc *typeCache) decodeFieldsSubType(r *bytes.Reader, tds []idl.Type) ([]idl.FieldType, error) {
	l, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, err
	}
	var fields []idl.FieldType
	for i := 0; i < int(l.Int64()); i++ {
		h, err := leb128.DecodeUnsigned(r)
		if err != nil {
			return nil, err
		}
		tid, err := leb128.DecodeSigned(r)
		if err != nil {
			return nil, err
		}
		o := idl.OpCode(tid.Int64())
		if v, err := o.GetType(tds); err != nil {
			fields = append(fields, idl.FieldType{
				Name:  h.String(),
				Index: o,
			})
		} else {
			fields = append(fields, idl.FieldType{
				Name: h.String(),
				Type: v,
			})
		}
	}
	return fields, nil
}

func (tc *typeCache) decodeFuncOpCode(r *bytes.Reader, tds []idl.Type) (idl.Type, error) {
	la, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, err
	}
	var args []idl.FunctionParameter
	for i := 0; i < int(la.Int64()); i++ {
		tid, err := leb128.DecodeSigned(r)
		if err != nil {
			return nil, err
		}
		o := idl.OpCode(tid.Int64())
		if v, err := o.GetType(tds); err != nil {
			args = append(args, idl.FunctionParameter{
				Index: o,
			})
		} else {
			args = append(args, idl.FunctionParameter{
				Type: v,
			})
		}
	}

	lr, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, err
	}
	var rets []idl.FunctionParameter
	for i := 0; i < int(lr.Int64()); i++ {
		tid, err := leb128.DecodeSigned(r)
		if err != nil {
			return nil, err
		}
		o := idl.OpCode(tid.Int64())
		if v, err := o.GetType(tds); err != nil {
			rets = append(rets, idl.FunctionParameter{
				Index: o,
			})
		} else {
			rets = append(rets, idl.FunctionParameter{
				Type: v,
			})
		}
	}

	l, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, err
	}
	ann := make([]byte, l.Int64())
	if _, err := r.Read(ann); err != nil {
		return nil, err
	}
	var anns []string
	if len(ann) != 0 {
		anns = append(anns, string(ann))
	}

	return &idl.FunctionType{
		ArgumentParameters: args,
		ReturnParameters:   rets,
		Annotations:        anns,
	}, nil
}

func (tc *typeCache) decodeOptOpCode(r *bytes.Reader, tds []idl.Type) (idl.Type, error) {
	tid, err := leb128.DecodeSigned(r)
	if err != nil {
		return nil, err
	}
	o := idl.OpCode(tid.Int64())
	f := func(tdt []idl.Type) (idl.Type, error) {
		v, err := o.GetType(tdt)
		if err != nil {
			return nil, err
		}
		return idl.NewOptionalType(v), nil
	}
	if v, err := f(tds); err == nil {
		return v, nil
	}

	tc.cache = append(tc.cache, delayType{
		index: len(tds),
		f:     f,
	})
	return nil, nil
}

func (tc *typeCache) decodeRecOpCode(r *bytes.Reader, tds []idl.Type) (idl.Type, error) {
	fields, err := tc.decodeFieldsSubType(r, tds)
	if err != nil {
		return nil, err
	}
	return &idl.RecordType{Fields: fields}, nil
}

func (tc *typeCache) decodeServiceOpCode(r *bytes.Reader, tds []idl.Type) (idl.Type, error) {
	l, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, err
	}
	var methods []idl.Method
	for i := 0; i < int(l.Int64()); i++ {
		lm, err := leb128.DecodeUnsigned(r)
		if err != nil {
			return nil, err
		}
		name := make([]byte, lm.Int64())
		n, err := r.Read(name)
		if err != nil {
			return nil, err
		}
		if n != int(lm.Int64()) {
			return nil, fmt.Errorf("invalid method name: %s", name)
		}

		tid, err := leb128.DecodeSigned(r)
		if err != nil {
			return nil, err
		}
		o := idl.OpCode(tid.Int64())
		v, err := o.GetType(tds)
		if err != nil {
			return nil, err
		}
		f, ok := v.(*idl.FunctionType)
		if !ok {
			return nil, fmt.Errorf("invalid method type: %s", reflect.TypeOf(v))
		}
		methods = append(methods, idl.Method{
			Name: string(name),
			Func: f,
		})
	}
	return &idl.Service{
		Methods: methods,
	}, nil
}

func (tc *typeCache) decodeVarOpCode(r *bytes.Reader, tds []idl.Type) (idl.Type, error) {
	fields, err := tc.decodeFieldsSubType(r, tds)
	if err != nil {
		return nil, err
	}
	return &idl.VariantType{Fields: fields}, nil
}

func (tc *typeCache) decodeVecOpCode(r *bytes.Reader, tds []idl.Type) (idl.Type, error) {
	tid, err := leb128.DecodeSigned(r)
	if err != nil {
		return nil, err
	}
	o := idl.OpCode(tid.Int64())
	f := func(tdt []idl.Type) (idl.Type, error) {
		v, err := o.GetType(tdt)
		if err != nil {
			return nil, err
		}
		return idl.NewVectorType(v), nil
	}
	if v, err := f(tds); err == nil {
		return v, nil
	}

	tc.cache = append(tc.cache, delayType{
		index: len(tds),
		f:     f,
	})
	return nil, nil
}

// resolve should empty out the type cache by resolving all forward references.
// It is still possible that records, vectors or functions have index
// references, these will need to be resolved separately.
func (tc *typeCache) resolve(tds []idl.Type) error {
	for len(tc.cache) != 0 {
		resolved := false
		for i, d := range tc.cache {
			if v, err := d.f(tds); v != nil && err == nil {
				tds[d.index] = v
				tc.cache = append(tc.cache[:i], tc.cache[i+1:]...)
				resolved = true
				break
			}
		}
		if !resolved {
			return fmt.Errorf("failed to resolve all types")
		}
	}

	return nil
}
