package jsonsrc

import (
	"encoding/json"
	"fmt"

	"github.com/kud360/goxsd5/xsd"
)

// jvalue is an order-preserving JSON value tree. encoding/json's Decoder.Token
// stream is consumed into it so object member order (which sequence content
// models depend on) survives; a plain map[string]any would lose it.
type jvalue interface{ isJValue() }

// jobject is a JSON object with members in source order.
type jobject struct {
	members []jmember
	pos     xsd.Pos
}

func (*jobject) isJValue() {}

type jmember struct {
	key   string
	value jvalue
	pos   xsd.Pos // position of the member's value
}

// jarray is a JSON array.
type jarray struct {
	items []jvalue
	pos   xsd.Pos
}

func (*jarray) isJValue() {}

// jscalar is a JSON string, number, or boolean, carried as its lexical text.
// number holds the json.Number token text so 5.0 is not renormalized to 5.
type jscalar struct {
	lexical string
	pos     xsd.Pos
}

func (*jscalar) isJValue() {}

// jnull is JSON null.
type jnull struct{ pos xsd.Pos }

func (*jnull) isJValue() {}

// decode consumes the whole token stream from dec into one jvalue. It uses
// InputOffset for rough source positions.
func decode(dec *json.Decoder, uri string) (jvalue, error) {
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	v, err := decodeValue(dec, uri, tok)
	if err != nil {
		return nil, err
	}
	// Reject trailing content: a JSON instance is a single document value.
	if dec.More() {
		return nil, fmt.Errorf("jsonsrc: trailing content after top-level JSON value")
	}
	return v, nil
}

func posAt(uri string, off int64) xsd.Pos {
	return xsd.Pos{URI: uri, Line: 1, Column: int(off)}
}

// decodeValue decodes the value whose first token is tok.
func decodeValue(dec *json.Decoder, uri string, tok json.Token) (jvalue, error) {
	pos := posAt(uri, dec.InputOffset())
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			return decodeObject(dec, uri, pos)
		case '[':
			return decodeArray(dec, uri, pos)
		}
		return nil, fmt.Errorf("jsonsrc: unexpected %q", t)
	case string:
		return &jscalar{lexical: t, pos: pos}, nil
	case json.Number:
		return &jscalar{lexical: t.String(), pos: pos}, nil
	case bool:
		if t {
			return &jscalar{lexical: "true", pos: pos}, nil
		}
		return &jscalar{lexical: "false", pos: pos}, nil
	case nil:
		return &jnull{pos: pos}, nil
	}
	return nil, fmt.Errorf("jsonsrc: unexpected token %v", tok)
}

// decodeObject decodes members until the matching '}', preserving order.
func decodeObject(dec *json.Decoder, uri string, pos xsd.Pos) (jvalue, error) {
	obj := &jobject{pos: pos}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("jsonsrc: object key is not a string")
		}
		valTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		val, err := decodeValue(dec, uri, valTok)
		if err != nil {
			return nil, err
		}
		obj.members = append(obj.members, jmember{key: key, value: val, pos: posAt(uri, dec.InputOffset())})
	}
	// Consume the closing '}'.
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	return obj, nil
}

// decodeArray decodes items until the matching ']'.
func decodeArray(dec *json.Decoder, uri string, pos xsd.Pos) (jvalue, error) {
	arr := &jarray{pos: pos}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		val, err := decodeValue(dec, uri, tok)
		if err != nil {
			return nil, err
		}
		arr.items = append(arr.items, val)
	}
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	return arr, nil
}
