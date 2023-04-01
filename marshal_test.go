package jsonnetenc_test

import (
	"fmt"
	"log"

	"github.com/mkmik/jsonnetenc"
)

func ExampleMarshal() {
	d := struct {
		Name      string               `json:"name"`
		Stuff     jsonnetenc.Import    `json:"stuff"`
		StuffStr  jsonnetenc.ImportStr `json:"stuffstr"`
		StuffBin  jsonnetenc.ImportBin `json:"stuffbin"`
		Var       jsonnetenc.Var       `json:"var"`
		Sum       jsonnetenc.Sum       `json:"sum"`
		Index     jsonnetenc.Index     `json:"index"`
		Dot       jsonnetenc.Member    `json:"dot"`
		FieldQuot jsonnetenc.Member    `json:"fieldQuot"`
		Hack      string               `json:"hack+"`
		SelfFoo   jsonnetenc.Self      `json:"selffoo"`
		SelfQuot  jsonnetenc.Self      `json:"selfquot"`
		SelfRes   jsonnetenc.Self      `json:"selfres"`
		SuperFoo  jsonnetenc.Super     `json:"superfoo"`
		SuperQuot jsonnetenc.Super     `json:"superquot"`
		SuperRes  jsonnetenc.Super     `json:"superres"`
	}{
		Name:     "foo",
		Stuff:    "bar",
		StuffStr: "bar.txt",
		StuffBin: "bar.bin",
		Var:      "baz",
		Sum: jsonnetenc.Sum{
			40,
			jsonnetenc.Var("x"),
			"foo",
			jsonnetenc.Import("stuff"),
			struct {
				X int `json:"x"`
			}{X: 42}},
		Index:     jsonnetenc.Index{LHS: jsonnetenc.Var("a"), RHS: jsonnetenc.Sum{"k", "e", "y"}},
		Dot:       jsonnetenc.Member{LHS: jsonnetenc.Var("a"), Field: "field"},
		FieldQuot: jsonnetenc.Member{LHS: jsonnetenc.Var("a"), Field: "1f"},
		Hack:      "foo",
		SelfFoo:   "foo",
		SelfQuot:  "foo-bar",
		SelfRes:   "self",
		SuperFoo:  "foo",
		SuperQuot: "foo-bar",
		SuperRes:  "self",
	}

	b, err := jsonnetenc.Marshal(d)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", string(b))
	// Output:
	// {
	//     "name": "foo",
	//     "stuff": (import "bar"),
	//     "stuffstr": (importstr "bar.txt"),
	//     "stuffbin": (importbin "bar.bin"),
	//     "var": baz,
	//     "sum": 40+x+"foo"+(import "stuff")+{
	//     "x": 42
	//},
	//     "index": a["k"+"e"+"y"],
	//     "dot": a.field,
	//     "fieldQuot": a["1f"],
	//     hack+: "foo",
	//     "selffoo": self.foo,
	//     "selfquot": self["foo-bar"],
	//     "selfres": self["self"],
	//     "superfoo": super.foo,
	//     "superquot": super["foo-bar"],
	//     "superres": super["self"]
	// }
}
