package jsonnetenc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/mkmik/multierror"

	"github.com/juju/errors"
)

const (
	jsonnetEscapeOpen  = "%%%jsonnet "
	jsonnetEscapeClose = " tennojs%%%" // sdrawkcab jsonnet
)

var (
	exprRegexp = regexp.MustCompile(fmt.Sprintf(`"(?U)%s(.*)%s"`, jsonnetEscapeOpen, jsonnetEscapeClose))
	// we cannot control the rendering of the keys, yet that is needed to generate the short for of: `a: super.a +`
	superHack = regexp.MustCompile(`"(?U)(.*\+)":`)
)

// Import renders as jsonnet import.
type Import string

// MarshalJSON implements the json.Marshaler interface
func (s Import) MarshalJSON() ([]byte, error) {
	return wrap(fmt.Sprintf("(import %q)", s)), nil
}

// Var renders as jsonnet variable reference.
type Var string

// MarshalJSON implements the json.Marshaler interface
func (s Var) MarshalJSON() ([]byte, error) {
	return wrap(string(s)), nil
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
			return nil, errors.Trace(err)
		}
		l = append(l, bytes.TrimSpace(b))
	}
	return wrap(string(bytes.Join(l, []byte("+")))), nil
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
		return nil, errors.Trace(err)
	}

	r, err := marshal(i.RHS)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return wrap(string(fmt.Sprintf("%s[%s]", string(bytes.TrimSpace(l)), string(bytes.TrimSpace(r))))), nil
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
		return nil, errors.Trace(err)
	}

	return wrap(string(fmt.Sprintf("%s.%s", string(bytes.TrimSpace(l)), m.Field))), nil
}

func wrap(s string) []byte {
	return []byte(fmt.Sprintf("%q", fmt.Sprintf("%s%s%s", jsonnetEscapeOpen, s, jsonnetEscapeClose)))
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
		return nil, errors.Trace(err)
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
		return nil, errors.Trace(err)
	}
	b = superHack.ReplaceAll(b, []byte("$1:"))
	return jsonnetFmt(b)
}

func marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "    ")
	if err := enc.Encode(v); err != nil {
		return nil, errors.Trace(err)
	}
	b, err := unwrap(buf.Bytes())
	return b, errors.Trace(err)
}

func jsonnetFmt(b []byte) ([]byte, error) {
	// NOP because not implemented yet in go-jsonnet
	return b, nil
}
