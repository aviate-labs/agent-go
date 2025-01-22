package internal

//go:generate go run github.com/0x51-dev/upeg/cmd/abnf --in=blob/grammar.abnf --out=blob/grammar.go --package=blob
//go:generate go run github.com/0x51-dev/upeg/cmd/abnf --in=candid/grammar.abnf --out=candid/grammar.go --package=candid --ignore=Def,DataType,NumType,ConsType,Fields,RefType,Name,Char,Num,HexNum,Utf,UtfEnc,utfcont,ascii,escape,letter,digit,hex,Comment,Nl,Ws,Sp,ESC
//go:generate go run github.com/0x51-dev/upeg/cmd/abnf --in=candidtest/grammar.abnf --out=candidtest/grammar.go --package=candidtest --ignore=Comment,MultiComment,Ws,EndLine,TestGoodTmpl,TestBadTmpl,ValuesBr,Values,Input,TextInputTmpl,BlobInputTmpl,String,Char,HexNum,Utf,UtfEnc,utfcont,ascii,escape,letter,digit,hex,ESC
//go:generate go run github.com/0x51-dev/upeg/cmd/abnf --in=candidvalue/grammar.abnf --out=candidvalue/grammar.go --package=candidvalue --ignore=Value,Bool,RecordFields,VariantField,VecFields,Sp,Spp,Ws,Char,HexNum,Utf,UtfEnc,utfcont,ascii,escape,letter,digit,hex,ESC
