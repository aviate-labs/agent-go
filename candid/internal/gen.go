package internal

//go:generate go run github.com/0x51-dev/upeg/cmd/abnf --in=candid/grammar.abnf --out=candid/grammar.go --package=candid --ignore=Def,DataType,NumType,ConsType,Fields,RefType,Name,Char,Num,HexNum,Utf,UtfEnc,utfcont,ascii,escape,letter,digit,hex,Comment,Nl,Ws,Sp,ESC
//go:generate go run github.com/0x51-dev/upeg/cmd/abnf --in=ctest/grammar.abnf --out=ctest/grammar.go --package=ctest --ignore=Comment,MultiComment,Ws,EndLine,TestGoodTmpl,TestBadTmpl,ValuesBr,Values,Input,TextInputTmpl,BlobInputTmpl,String,Char,UChar,EscapedDQuote,digit,hex
//go:generate go run github.com/0x51-dev/upeg/cmd/abnf --in=cvalue/grammar.abnf --out=cvalue/grammar.go --package=cvalue --ignore=Value,Bool,RecordFields,VariantField,VecFields,Sp,Spp,Ws,Char,HexNum,Utf,UtfEnc,utfcont,ascii,escape,letter,digit,hex,ESC
