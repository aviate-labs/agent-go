package candid

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"slices"
	"unicode/utf8"

	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/leb128"
)

// Decode decodes with DefaultDecodingQuota; use DecodeWithQuota to set it.
func Decode(bs []byte) ([]idl.Type, []any, error) {
	return DecodeWithQuota(bs, idl.NewBudget(idl.DefaultDecodingQuota))
}

// DecodeWithQuota is Decode with an explicit budget. Pass
// idl.NewUnlimitedBudget to lift the ceiling entirely.
func DecodeWithQuota(bs []byte, budget *idl.Budget) ([]idl.Type, []any, error) {
	ts, r, err := decodeTypes(bs)
	if err != nil {
		return nil, nil, err
	}

	var vs []any
	{ // M
		for i := range ts {
			v, err := ts[i].Decode(r, budget)
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

// Unmarshal decodes with DefaultDecodingQuota. Its signature is deliberately
// plain so it stays usable as a func([]byte, []any) error value; use
// UnmarshalWithQuota to set the quota.
func Unmarshal(data []byte, values []any) error {
	return UnmarshalWithQuota(data, values, idl.NewBudget(idl.DefaultDecodingQuota))
}

// UnmarshalWithQuota is Unmarshal with an explicit budget. Pass
// idl.NewUnlimitedBudget to lift the ceiling entirely.
func UnmarshalWithQuota(data []byte, values []any, budget *idl.Budget) error {
	ts, r, err := decodeTypes(data)
	if err != nil {
		return err
	}
	// The sender may supply fewer arguments than expected: the missing trailing
	// ones decode as null, but only where null is a valid value of the expected
	// type. Extra arguments are ignored.
	if len(ts) < len(values) {
		for _, v := range values[len(ts):] {
			if reflect.ValueOf(v).Elem().Kind() != reflect.Pointer {
				return fmt.Errorf("missing argument: %d of %d", len(ts), len(values))
			}
		}
		values = values[:len(ts)]
	}
	if len(ts) > len(values) {
		return fmt.Errorf("unequal value lengths: %d %d", len(ts), len(values))
	}

	for i, v := range values {
		switch v := v.(type) {
		case *idl.RawMessage:
			bs, err := ts[i].Read(r)
			if err != nil {
				return err
			}
			*v = bs
		default:
			vs, err := ts[i].Decode(r, budget)
			if err != nil {
				return err
			}
			if err := idl.UnmarshalGo(ts[i], vs, v); err != nil {
				return err
			}
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

func decodeTypes(bs []byte) ([]idl.Type, *bytes.Reader, error) {
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
		for range int(tdtl.Int64()) {
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
			default:
				if o >= 0 {
					return nil, nil, fmt.Errorf("invalid opcode: %d", o)
				}
				count, err := idl.DecodeLen(r)
				if err != nil {
					return nil, nil, err
				}
				skip := make([]byte, count)
				if _, err := io.ReadFull(r, skip); err != nil {
					return nil, nil, err
				}
				tds = append(tds, &idl.FutureType{OpCode: o})
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

		// Decoding one overflows the stack, which Go cannot recover from.
		// Replacing the slot is not enough: a field elsewhere still holds the
		// original, so substitute through the references too.
		empty := make(map[idl.Type]bool)
		for i, t := range tds {
			if isEmptyType(t, nil) {
				empty[t] = true
				tds[i] = new(idl.EmptyType)
			}
		}
		if len(empty) != 0 {
			for _, t := range tds {
				switch t := t.(type) {
				case *idl.RecordType:
					for i, f := range t.Fields {
						if empty[f.Type] {
							t.Fields[i].Type = new(idl.EmptyType)
						}
					}
				case *idl.VariantType:
					for i, f := range t.Fields {
						if empty[f.Type] {
							t.Fields[i].Type = new(idl.EmptyType)
						}
					}
				}
			}
		}
	}

	for _, t := range tds {
		if err := idl.CheckSize(t); err != nil {
			return nil, nil, err
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
	return ts, r, nil
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
	var prev *big.Int
	for i := 0; i < int(l.Int64()); i++ {
		h, err := leb128.DecodeUnsigned(r)
		if err != nil {
			return nil, err
		}
		if h.BitLen() > 32 {
			return nil, fmt.Errorf("field id out of range: %s", h)
		}
		if prev != nil {
			switch h.Cmp(prev) {
			case 0:
				return nil, fmt.Errorf("duplicate field id: %s", h)
			case -1:
				return nil, fmt.Errorf("field ids not in increasing order: %s after %s", h, prev)
			}
		}
		prev = h
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

	al, err := idl.DecodeLen(r)
	if err != nil {
		return nil, err
	}
	ann := make([]byte, al)
	if _, err := io.ReadFull(r, ann); err != nil {
		return nil, err
	}
	var anns []string
	for _, a := range ann {
		switch a {
		case 0x01:
			anns = append(anns, "query")
		case 0x02:
			anns = append(anns, "oneway")
		case 0x03:
			anns = append(anns, "composite_query")
		default:
			return nil, fmt.Errorf("invalid function annotation: %d", a)
		}
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
	var (
		methods []idl.Method
		opcodes []idl.OpCode
	)
	for i := 0; i < int(l.Int64()); i++ {
		lm, err := idl.DecodeLen(r)
		if err != nil {
			return nil, err
		}
		name := make([]byte, lm)
		if _, err := io.ReadFull(r, name); err != nil {
			return nil, err
		}
		if !utf8.Valid(name) {
			return nil, fmt.Errorf("invalid utf8 in method name")
		}
		// Methods are sorted by name, so this also rejects duplicates.
		if len(methods) != 0 && methods[len(methods)-1].Name >= string(name) {
			return nil, fmt.Errorf("method name %s duplicate or not sorted", name)
		}

		tid, err := leb128.DecodeSigned(r)
		if err != nil {
			return nil, err
		}
		methods = append(methods, idl.Method{Name: string(name)})
		opcodes = append(opcodes, idl.OpCode(tid.Int64()))
	}

	s := &idl.ServiceType{Methods: methods}
	// A method may refer to a type defined later in the table, so resolve the
	// signatures once the whole table is known. Commit only once every method
	// landed: a partial retry would leave nil Funcs behind.
	f := func(tds []idl.Type) (idl.Type, error) {
		resolved := make([]idl.Method, len(methods))
		copy(resolved, methods)
		for i, o := range opcodes {
			v, err := o.GetType(tds)
			if err != nil {
				return nil, err
			}
			fn, ok := v.(*idl.FunctionType)
			if !ok {
				return nil, fmt.Errorf("invalid method type: %s", reflect.TypeOf(v))
			}
			resolved[i].Func = fn
		}
		s.Methods = resolved
		return s, nil
	}
	if v, err := f(tds); err == nil {
		return v, nil
	}
	tc.cache = append(tc.cache, delayType{index: len(tds), f: f})
	return nil, nil
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
			// A cycle: what is left refers to a slot that is itself pending.
			// Placeholders let the rest resolve, then get filled in; restore them
			// on failure so no inner is left nil.
			prev := make([]idl.Type, len(tc.cache))
			var recs []*idl.RecursiveType
			for i, d := range tc.cache {
				prev[i] = tds[d.index]
				rec := idl.NewRecursiveType(fmt.Sprintf("rec_%d", d.index))
				tds[d.index] = rec
				recs = append(recs, rec)
			}
			for i, d := range tc.cache {
				v, err := d.f(tds)
				if v == nil || err != nil {
					for j, d := range tc.cache {
						tds[d.index] = prev[j]
					}
					return fmt.Errorf("failed to resolve all types")
				}
				recs[i].SetInner(v)
			}
			tc.cache = nil
			return nil
		}
	}

	return nil
}

// isEmptyType reports a record that reaches itself through records only: it
// has no finite value. An opt or vec on the path breaks the cycle.
func isEmptyType(t idl.Type, seen []idl.Type) bool {
	if slices.Contains(seen, t) {
		return true
	}
	// resolve() breaks cycles with a RecursiveType, so a record cycle reaches
	// itself through one rather than directly.
	if rec, ok := t.(*idl.RecursiveType); ok {
		inner := rec.Inner()
		if inner == nil {
			return false
		}
		return isEmptyType(inner, append(seen, t))
	}
	rec, ok := t.(*idl.RecordType)
	if !ok {
		return false
	}
	seen = append(seen, t)
	for _, f := range rec.Fields {
		if isEmptyType(f.Type, seen) {
			return true
		}
	}
	return false
}
