package idl

type IDL struct {
	Null      *NullType
	Bool      *BoolType
	Nat       *NatType
	Int       *IntType
	Nat8      *NatType
	Nat16     *NatType
	Nat32     *NatType
	Nat64     *NatType
	Int8      *IntType
	Int16     *IntType
	Int32     *IntType
	Int64     *IntType
	Float32   *FloatType
	Float64   *FloatType
	Text      *TextType
	Reserved  *ReservedType
	Empty     *EmptyType
	Opt       func(typ Type) *OptionalType
	Tuple     func(ts ...Type) *TupleType
	Vec       func(t Type) *VectorType
	Record    func(fields map[string]Type) *RecordType
	Variant   func(fields map[string]Type) *VariantType
	Func      func(args, ret []FunctionParameter, annotations []string) *FunctionType
	Service   func(functions map[string]*FunctionType) *Service
	Principal *PrincipalType
}

type IDLFactory = func(types IDL) *Service

func NewInterface(factory IDLFactory) *Service {
	return factory(IDL{
		Bool:     new(BoolType),
		Null:     new(NullType),
		Nat:      new(NatType),
		Int:      new(IntType),
		Nat8:     Nat8Type(),
		Nat16:    Nat16Type(),
		Nat32:    Nat32Type(),
		Nat64:    Nat64Type(),
		Int8:     Int8Type(),
		Int16:    Int16Type(),
		Int32:    Int32Type(),
		Int64:    Int64Type(),
		Text:     new(TextType),
		Float32:  Float32Type(),
		Float64:  Float64Type(),
		Reserved: new(ReservedType),
		Empty:    new(EmptyType),
		Opt: func(typ Type) *OptionalType {
			return &OptionalType{Type: typ}
		},
		Tuple: func(ts ...Type) *TupleType {
			tuple := TupleType(ts)
			return &tuple
		},
		Vec: func(t Type) *VectorType {
			return NewVectorType(t)
		},
		Record: func(fields map[string]Type) *RecordType {
			return NewRecordType(fields)
		},
		Variant: func(fields map[string]Type) *VariantType {
			return NewVariantType(fields)
		},
		Func: func(argumentTypes, returnTypes []FunctionParameter, annotations []string) *FunctionType {
			return NewFunctionType(argumentTypes, returnTypes, annotations)
		},
		Service: func(methods map[string]*FunctionType) *Service {
			return NewServiceType(methods)
		},
		Principal: new(PrincipalType),
	})
}
