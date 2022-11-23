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
				v, err := getType(tid.Int64(), tds)
				if err != nil {
					return nil, nil, err
				}
				tds = append(tds, &OptionalType{v})
			case vecType:
				tid, err := leb128.DecodeSigned(r)
				if err != nil {
					return nil, nil, err
				}
				v, err := getType(tid.Int64(), tds)
				if err != nil {
					return nil, nil, err
				}
				tds = append(tds, &VectorType{v})
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
					v, err := getType(tid.Int64(), tds)
					if err != nil {
						return nil, nil, err
					}
					fields = append(fields, FieldType{
						Name: h.String(),
						Type: v,
					})
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
					v, err := getType(tid.Int64(), tds)
					if err != nil {
						return nil, nil, err
					}
					fields = append(fields, FieldType{
						Name: h.String(),
						Type: v,
					})
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
