package marshal

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/leb128"
)

func DecodeOpt(r *bytes.Reader, ctx Context[idl.Type]) (any, error) {
	l, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch l {
	case 0x00:
		return nil, nil
	case 0x01:
		return ctx.typ.Decode(r)
	default:
		return nil, fmt.Errorf("invalid option value")
	}
}

func DecodeRecord(r *bytes.Reader, ctx Context[*idl.RecordType]) (map[string]any, error) {
	record := make(map[string]any)
	for _, f := range ctx.typ.Fields {
		v, err := f.Type.Decode(r)
		if err != nil {
			return nil, err
		}
		record[f.Name] = v
	}
	return record, nil
}

func DecodeVariant(r *bytes.Reader, ctx Context[*idl.VariantType]) (map[string]any, error) {
	id, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, err
	}
	if id.Cmp(big.NewInt(int64(len(ctx.typ.Fields)))) >= 0 {
		return nil, fmt.Errorf("invalid variant index: %v", id)
	}
	v, err := ctx.typ.Fields[int(id.Int64())].Type.Decode(r)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		id.String(): v,
	}, nil
}

func DecodeVector(r *bytes.Reader, ctx idl.Type) ([]any, error) {
	l, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, err
	}
	var vs []any
	for i := 0; i < int(l.Int64()); i++ {
		v_, err := ctx.Decode(r)
		if err != nil {
			return nil, err
		}
		vs = append(vs, v_)
	}
	return vs, nil
}

func EncodeCons(value any, tdt *idl.TypeDefinitionTable) ([]byte, []byte, error) {
	typ, err := idl.TypeOf(value)
	if err != nil {
		return nil, nil, err
	}
	t, err := typ.EncodeType(tdt)
	if err != nil {
		return nil, nil, err
	}
	v, err := typ.EncodeValue(value)
	if err != nil {
		return nil, nil, err
	}
	return t, v, nil
}

type DecodeFunc = func(*bytes.Reader, Context[idl.Type]) (any, error)
