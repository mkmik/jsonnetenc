package jsonnetenc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-jsonnet/formatter"
	"github.com/mkmik/multierror"
)

const (
	jsonnetEscapeOpen  = "%%%jsonnet "
	jsonnetEscapeClose = " tennojs%%%" // sdrawkcab jsonnet
)

var (
	exprRegexp = regexp.MustCompile(fmt.Sprintf(`"(?U)%s(.*)%s"`, jsonnetEscapeOpen, jsonnetEscapeClose))
	// we cannot control the rendering of the keys, yet that is needed to generate the short for of: `a: super.a +`
	superHack = regexp.MustCompile(`"(?U)(.*\+)":`)

	idRegexp            = regexp.MustCompile(`^[_a-zA-Z][_a-zA-Z0-9]*$`)
	reservedIdentifiers = map[string]bool{
		"assert": true, "else": true, "error": true, "false": true, "for": true,
		"function": true, "if": true, "import": true, "importstr": true, "importbin": true,
		"in": true, "local": true, "null": true, "tailstrict": true, "then": true,
		"self": true, "super": true, "true": true,
	}
)

// Import renders as jsonnet import.
type Import string

// MarshalJSON implements the json.Marshaler interface
func (s Import) MarshalJSON() ([]byte, error) {
	return wrap(fmt.Sprintf("(import %q)", s)), nil
}

// ImportStr renders as jsonnet importstr.
type ImportStr string

// MarshalJSON implements the json.Marshaler interface
func (s ImportStr) MarshalJSON() ([]byte, error) {
	return wrap(fmt.Sprintf("(importstr %q)", s)), nil
}

// ImportBin renders as jsonnet importbin.
type ImportBin string

// MarshalJSON implements the json.Marshaler interface
func (s ImportBin) MarshalJSON() ([]byte, error) {
	return wrap(fmt.Sprintf("(importbin %q)", s)), nil
}

// Self with value "foo" renders as "self.foo".
type Self string

// MarshalJSON implements the json.Marshaler interface
func (s Self) MarshalJSON() ([]byte, error) {
	return wrap(fmt.Sprintf("self[%q]", s)), nil
}

// Super with value "foo" renders as "super.foo".
type Super string

// MarshalJSON implements the json.Marshaler interface
func (s Super) MarshalJSON() ([]byte, error) {
	// format.Format doesn't convert `super["foo"]` into `super.foo`
	// like it does for any other field.

	if !requiresEscaping(string(s)) {
		return wrap(fmt.Sprintf("super.%s", s)), nil
	} else {
		return wrap(fmt.Sprintf("super[%q]", s)), nil
	}
}

// Var renders as jsonnet variable reference.
type Var string

// MarshalJSON implements the json.Marshaler interface
func (s Var) MarshalJSON() ([]byte, error) {
	return wrap(s), nil
}

// Sum renders a sum expression
type Sum []interface{}

// MarshalJSON implements the json.Marshaler interface
func (s Sum) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	var l [][]byte
	for _, e := range s {
		b, err := marshal(e)
		if err != nil {
			return nil, err
		}
		l = append(l, bytes.TrimSpace(b))
	}
	return wrap(bytes.Join(l, []byte("+"))), nil
}

// Index renders a index expression, i.e. foo[bar]
type Index struct {
	LHS interface{}
	RHS interface{}
}

// MarshalJSON implements the json.Marshaler interface
func (i Index) MarshalJSON() ([]byte, error) {
	l, err := marshal(i.LHS)
	if err != nil {
		return nil, err
	}

	r, err := marshal(i.RHS)
	if err != nil {
		return nil, err
	}

	return wrap(fmt.Sprintf("%s[%s]", string(bytes.TrimSpace(l)), string(bytes.TrimSpace(r)))), nil
}

// Member renders a dot expression, i.e. foo.bar
type Member struct {
	LHS   interface{}
	Field string
}

// MarshalJSON implements the json.Marshaler interface
func (m Member) MarshalJSON() ([]byte, error) {
	l, err := marshal(m.LHS)
	if err != nil {
		return nil, err
	}
	return wrap(fmt.Sprintf("%s[%q]", string(bytes.TrimSpace(l)), m.Field)), nil
}

type wrappable interface {
	~string | []byte
}

func wrap[S wrappable](s S) []byte {
	return []byte(fmt.Sprintf("%q", fmt.Sprintf("%s%s%s", jsonnetEscapeOpen, string(s), jsonnetEscapeClose)))
}

func unwrap(b []byte) ([]byte, error) {
	var err error
	r := exprRegexp.ReplaceAllFunc(b, func(i []byte) []byte {
		var s string
		err = multierror.Append(err, json.Unmarshal(i, &s))
		s = strings.TrimPrefix(s, jsonnetEscapeOpen)
		s = strings.TrimSuffix(s, jsonnetEscapeClose)
		return []byte(s)
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Marshal returns the Jsonnet encoding of v.
//
// It behaves much like JSON encoding except that it also supports
// rendering values that are jsonnet expressions (such as imports).
func Marshal(v interface{}) ([]byte, error) {
	b, err := marshal(v)
	if err != nil {
		return nil, err
	}
	b = superHack.ReplaceAll(b, []byte("$1:"))
	return jsonnetFmt(b)
}

func marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "    ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	b, err := unwrap(buf.Bytes())
	return b, err
}

func jsonnetFmt(b []byte) ([]byte, error) {
	s, err := formatter.Format("", string(b), formatter.DefaultOptions())
	return []byte(s), err
}

func requiresEscaping(s string) bool {
	_, reserved := reservedIdentifiers[s]
	if reserved {
		return true
	}
	return !idRegexp.MatchString(s)
}
