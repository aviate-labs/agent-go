package idl

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/aviate-labs/leb128"
)

func Decode(bs []byte) ([]Type, []any, error) {
	if len(bs) == 0 {
		return nil, nil, &FormatError{
			Description: "empty",
		}
	}

	r := bytes.NewReader(bs)

	{ // 'DIDL'
		magic := make([]byte, 4)
		n, err := r.Read(magic)
		if err != nil {
			return nil, nil, err
		}
		if n < 4 {
			return nil, nil, &FormatError{
				Description: "no magic bytes",
			}
		}
		if !bytes.Equal(magic, []byte{'D', 'I', 'D', 'L'}) {
			return nil, nil, &FormatError{
				Description: "wrong magic bytes",
			}
		}
	}

	var tds []Type
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
			switch tid.Int64() {
			case optType:
				tid, err := leb128.DecodeSigned(r)
				if err != nil {
					return nil, nil, err
				}
				f := func(tdt []Type) (Type, error) {
					v, err := getType(tid.Int64(), tdt)
					if err != nil {
						return nil, err
					}
					return &OptionalType{v}, nil
				}
				if v, err := f(tds); err == nil {
					tds = append(tds, v)
				} else {
					tc = append(tc, delayType{
						index: len(tds),
						f:     f,
					})
					tds = append(tds, nil)
				}
			case vecType:
				tid, err := leb128.DecodeSigned(r)
				if err != nil {
					return nil, nil, err
				}
				f := func(tdt []Type) (Type, error) {
					v, err := getType(tid.Int64(), tdt)
					if err != nil {
						return nil, err
					}
					return &VectorType{v}, nil
				}
				if v, err := f(tds); err == nil {
					tds = append(tds, v)
				} else {
					tc = append(tc, delayType{
						index: len(tds),
						f:     f,
					})
					tds = append(tds, nil)
				}
			case recType:
				l, err := leb128.DecodeUnsigned(r)
				if err != nil {
					return nil, nil, err
				}
				var fields []FieldType
				for i := 0; i < int(l.Int64()); i++ {
					h, err := leb128.DecodeUnsigned(r)
					if err != nil {
						return nil, nil, err
					}
					tid, err := leb128.DecodeSigned(r)
					if err != nil {
						return nil, nil, err
					}
					if v, err := getType(tid.Int64(), tds); err != nil {
						fields = append(fields, FieldType{
							Name:  h.String(),
							index: tid.Int64(),
						})
					} else {
						fields = append(fields, FieldType{
							Name: h.String(),
							Type: v,
						})
					}
				}
				tds = append(tds, &RecordType{Fields: fields})
			case varType:
				l, err := leb128.DecodeUnsigned(r)
				if err != nil {
					return nil, nil, err
				}
				var fields []FieldType
				for i := 0; i < int(l.Int64()); i++ {
					h, err := leb128.DecodeUnsigned(r)
					if err != nil {
						return nil, nil, err
					}
					tid, err := leb128.DecodeSigned(r)
					if err != nil {
						return nil, nil, err
					}
					if v, err := getType(tid.Int64(), tds); err != nil {
						fields = append(fields, FieldType{
							Name:  h.String(),
							index: tid.Int64(),
						})
					} else {
						fields = append(fields, FieldType{
							Name: h.String(),
							Type: v,
						})
					}
				}
				tds = append(tds, &VariantType{Fields: fields})
			case funcType:
				la, err := leb128.DecodeUnsigned(r)
				if err != nil {
					return nil, nil, err
				}
				var args []Type
				for i := 0; i < int(la.Int64()); i++ {
					tid, err = leb128.DecodeSigned(r)
					if err != nil {
						return nil, nil, err
					}
					v, err := getType(tid.Int64(), tds)
					if err != nil {
						return nil, nil, err
					}
					args = append(args, v)
				}
				lr, err := leb128.DecodeUnsigned(r)
				if err != nil {
					return nil, nil, err
				}
				var rets []Type
				for i := 0; i < int(lr.Int64()); i++ {
					tid, err = leb128.DecodeSigned(r)
					if err != nil {
						return nil, nil, err
					}
					v, err := getType(tid.Int64(), tds)
					if err != nil {
						return nil, nil, err
					}
					rets = append(rets, v)
				}
				l, err := leb128.DecodeUnsigned(r)
				if err != nil {
					return nil, nil, err
				}
				ann := make([]byte, l.Int64())
				if _, err := r.Read(ann); err != nil {
					return nil, nil, err
				}
				var anns []string
				if len(ann) != 0 {
					anns = append(anns, string(ann))
				}
				tds = append(tds, &FunctionType{
					ArgTypes:    args,
					RetTypes:    rets,
					Annotations: anns,
				})
			case serviceType:
				l, err := leb128.DecodeUnsigned(r)
				if err != nil {
					return nil, nil, err
				}
				var methods []Method
				for i := 0; i < int(l.Int64()); i++ {
					lm, err := leb128.DecodeUnsigned(r)
					if err != nil {
						return nil, nil, err
					}
					name := make([]byte, lm.Int64())
					n, err := r.Read(name)
					if err != nil {
						return nil, nil, err
					}
					if n != int(lm.Int64()) {
						return nil, nil, fmt.Errorf("invalid method name: %d", bs)
					}

					tid, err = leb128.DecodeSigned(r)
					if err != nil {
						return nil, nil, err
					}
					v, err := getType(tid.Int64(), tds)
					if err != nil {
						return nil, nil, err
					}
					f, ok := v.(*FunctionType)
					if !ok {
						fmt.Println(reflect.TypeOf(v))
					}
					methods = append(methods, Method{
						Name: string(name),
						Func: f,
					})
				}
				tds = append(tds, &Service{
					methods: methods,
				})
			}
		}

		for len(tc) != 0 {
			resolved := false
			for i, d := range tc {
				if v, err := d.f(tds); v != nil && err == nil {
					tds[d.index] = v
					tc = append(tc[:i], tc[i+1:]...)
					resolved = true
					break
				}
			}
			if !resolved {
				return nil, nil, fmt.Errorf("failed to resolve all types")
			}
		}

		for i, tb := range tds {
			switch t := tb.(type) {
			case *VariantType:
				resolved := true
				for _, f := range t.Fields {
					if f.Type == nil {
						resolved = false
					}
				}
				if resolved {
					continue
				}

				f := func(tds []Type) (Type, error) {
					for i, f := range t.Fields {
						v, err := getType(f.index, tds)
						if err != nil {
							return nil, err
						}
						t.Fields[i].Type = v
					}
					return t, nil
				}
				if v, err := f(tds); v == nil || err != nil {
					tc = append(tc, delayType{
						index: i,
						f:     f,
					})
				}
			case *RecordType:
				resolved := true
				for _, f := range t.Fields {
					if f.Type == nil {
						resolved = false
					}
				}
				if resolved {
					continue
				}

				f := func(tds []Type) (Type, error) {
					for i, f := range t.Fields {
						v, err := getType(f.index, tds)
						if err != nil {
							return nil, err
						}
						t.Fields[i].Type = v
					}
					return t, nil
				}
				if v, err := f(tds); v == nil || err != nil {
					tc = append(tc, delayType{
						index: i,
						f:     f,
					})
				}
			}
		}
	}

	tsl, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, nil, err
	}

	var ts []Type
	{ // I
		for i := 0; i < int(tsl.Int64()); i++ {
			tid, err := leb128.DecodeSigned(r)
			if err != nil {
				return nil, nil, err
			}
			t, err := getType(tid.Int64(), tds)
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

type delayType struct {
	// index is the index of the type in the type list.
	index int
	// f is a function that takes the type list and returns the resolved type.
	f func(tdt []Type) (Type, error)
}

// typeCache is a cache of types that are not yet fully decoded.
// It is used to resolve recursive types.
// - int is the index of the type in the type list.
// - []delayType is a list of type that depend on the type to be resolved.
type typeCache []delayType
