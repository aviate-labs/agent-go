package marshal

type Type uint8

const (
	Null      Type = 0x7f // sleb128(-1)
	Bool      Type = 0x7e // sleb128(-2)
	Nat       Type = 0x7d // sleb128(-3)
	Int       Type = 0x7c // sleb128(-4)
	Nat8      Type = 0x7b // sleb128(-5)
	Nat16     Type = 0x7a // sleb128(-6)
	Nat32     Type = 0x79 // sleb128(-7)
	Nat64     Type = 0x78 // sleb128(-8)
	Int8      Type = 0x77 // sleb128(-9)
	Int16     Type = 0x76 // sleb128(-10)
	Int32     Type = 0x75 // sleb128(-11)
	Int64     Type = 0x74 // sleb128(-12)
	Float32   Type = 0x73 // sleb128(-13)
	Float64   Type = 0x72 // sleb128(-14)
	Text      Type = 0x71 // sleb128(-15)
	Reserved  Type = 0x70 // sleb128(-16)
	Empty     Type = 0x6f // sleb128(-17)
	Option    Type = 0x6e // sleb128(-18)
	Vector    Type = 0x6d // sleb128(-19)
	Record    Type = 0x6c // sleb128(-20)
	Variant   Type = 0x6b // sleb128(-21)
	Principal Type = 0x68 // sleb128(-24)
)

func (c Type) bytes() []byte {
	return []byte{uint8(c)}
}
