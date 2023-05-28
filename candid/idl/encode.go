package idl

import (
	"fmt"
	"github.com/aviate-labs/leb128"
	"math/big"
)

func Encode(argumentTypes []Type, arguments []any) ([]byte, error) {
	if len(arguments) < len(argumentTypes) {
		return nil, fmt.Errorf("invalid number of arguments")
	}

	// T
	tdt := &TypeDefinitionTable{
		Indexes: make(map[string]int),
	}
	for _, t := range argumentTypes {
		if err := t.AddTypeDefinition(tdt); err != nil {
			return nil, err
		}
	}

	tdtl, err := leb128.EncodeUnsigned(big.NewInt(int64(len(tdt.Types))))
	if err != nil {
		return nil, err
	}
	var tdte []byte
	for _, t := range tdt.Types {
		tdte = append(tdte, t...)
	}

	tsl, err := leb128.EncodeSigned(big.NewInt(int64(len(argumentTypes))))
	if err != nil {
		return nil, err
	}
	var (
		ts []byte
		vs []byte
	)
	for i, t := range argumentTypes {
		{ // I
			t, err := t.EncodeType(tdt)
			if err != nil {
				return nil, err
			}
			ts = append(ts, t...)
		}
		{ // M
			v, err := t.EncodeValue(arguments[i])
			if err != nil {
				return nil, err
			}
			vs = append(vs, v...)
		}
	}

	return concat(
		// magic number
		[]byte{'D', 'I', 'D', 'L'},
		// type definition table: T*(<datatype>*)
		tdtl, tdte,
		// types of the argument list: I*(<datatype>*)
		tsl, ts,
		// values of argument list: M(<datatype>*)
		vs,
	), nil
}

func Marshal(args []any) ([]byte, error) {
	var types []Type
	for _, a := range args {
		t, err := TypeOf(a)
		if err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return Encode(types, args)
}
