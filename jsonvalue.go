package jsonvalue

import (
	"bytes"
	"container/list"
	"fmt"

	"github.com/buger/jsonparser"
)

// V is the main type of jsonvalue
type V struct {
	valueType      jsonparser.ValueType
	rawNumBytes    []byte
	negative       bool
	floated        bool
	stringValue    string
	int64Value     int64
	uint64Value    uint64
	boolValue      bool
	floatValue     float64
	objectChildren map[string]*V
	arrayChildren  *list.List
}

func new() *V {
	v := V{}
	v.valueType = jsonparser.NotExist
	v.objectChildren = make(map[string]*V)
	v.arrayChildren = list.New()
	return &v
}

// UnmarshalString is equavilent to Unmarshal(string(b))
func UnmarshalString(s string) (*V, error) {
	return Unmarshal([]byte(s))
}

// Unmarshal parse raw bytes and return JSON value object
func Unmarshal(b []byte) (ret *V, err error) {
	if nil == b || 0 == len(b) {
		return nil, ErrNilParameter
	}

	for i, c := range b {
		switch c {
		case ' ', '\r', '\n', '\t':
			// continue
		case '{':
			// object start
			return newFromObject(b[i:])
		case '[':
			return newFromArray(b[i:])
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
			return newFromNumber(b[i:])
		case '"':
			return newFromString(b[i:])
		case 't':
			return newFromTrue(b[i:])
		case 'f':
			return newFromFalse(b[i:])
		case 'n':
			return newFromNull(b[i:])
		default:
			return nil, ErrRawBytesUnrecignized
		}
	}

	return nil, ErrRawBytesUnrecignized
}

// ==== simple object parsing ====
func newFromNumber(b []byte) (ret *V, err error) {
	v := new()
	v.valueType = jsonparser.Number

	if bytes.Contains(b, []byte(".")) {
		v.floated = true
		v.floatValue, err = parseFloat(b)
		if err != nil {
			return
		}

		v.negative = v.floatValue < 0
		v.int64Value = int64(v.floatValue)
		v.uint64Value = uint64(v.floatValue)

	} else if '-' == b[0] {
		v.negative = true
		v.int64Value, err = parseInt(b)
		if err != nil {
			return
		}

		v.uint64Value = uint64(v.int64Value)
		v.floatValue = float64(v.int64Value)

	} else {
		v.negative = false
		v.uint64Value, err = parseUint(b)
		if err != nil {
			return
		}

		v.int64Value = int64(v.uint64Value)
		v.floatValue = float64(v.uint64Value)
	}

	v.rawNumBytes = b
	return v, nil
}

func newFromString(b []byte) (ret *V, err error) {
	v := new()
	v.valueType = jsonparser.String
	v.stringValue, err = parseString(b)
	if err != nil {
		return
	}
	return v, nil
}

func newFromTrue(b []byte) (ret *V, err error) {
	if string(b) != "true" {
		return nil, ErrNotValidBoolValue
	}

	v := new()
	v.valueType = jsonparser.Boolean
	v.boolValue = true
	return v, nil
}

func newFromFalse(b []byte) (ret *V, err error) {
	if string(b) != "false" {
		return nil, ErrNotValidBoolValue
	}

	v := new()
	v.valueType = jsonparser.Boolean
	v.boolValue = false
	return v, nil
}

func newFromBool(b []byte) (ret *V, err error) {
	v := new()
	v.valueType = jsonparser.Boolean

	switch string(b) {
	case "true":
		v.boolValue = true
	case "false":
		v.boolValue = false
	default:
		return nil, ErrNotValidBoolValue
	}

	return v, nil
}

func newFromNull(b []byte) (ret *V, err error) {
	if string(b) != "null" {
		return nil, ErrNotValidNulllValue
	}

	v := new()
	v.valueType = jsonparser.Null
	return v, nil
}

// ====
func newFromArray(b []byte) (ret *V, err error) {
	o := new()
	o.valueType = jsonparser.Array

	jsonparser.ArrayEach(b, func(v []byte, t jsonparser.ValueType, _ int, _ error) {
		if err != nil {
			return
		}

		var child *V

		switch t {
		default:
			err = fmt.Errorf("invalid value type: %v", t)
		case jsonparser.Object:
			child, err = newFromObject(v)
		case jsonparser.Array:
			child, err = newFromArray(v)
		case jsonparser.Number:
			child, err = newFromNumber(v)
		case jsonparser.Boolean:
			child, err = newFromBool(v)
		case jsonparser.Null:
			child, err = newFromNull(v)
		case jsonparser.String:
			s, err := parseStringNoQuote(v)
			if err != nil {
				return
			}
			child = new()
			child.valueType = jsonparser.String
			child.stringValue = s
		}

		if err != nil {
			return
		}
		o.arrayChildren.PushBack(child)
		return
	})

	// done
	if err != nil {
		return
	}
	return o, nil
}

// ==== object parsing ====
func newFromObject(b []byte) (ret *V, err error) {
	o := new()
	o.valueType = jsonparser.Object

	err = jsonparser.ObjectEach(b, func(k, v []byte, t jsonparser.ValueType, _ int) error {
		// key
		var child *V
		key, err := parseStringNoQuote(k)
		if err != nil {
			return err
		}

		switch t {
		default:
			return fmt.Errorf("invalid value type: %v", t)
		case jsonparser.Object:
			child, err = newFromObject(v)
		case jsonparser.Array:
			child, err = newFromArray(v)
		case jsonparser.Number:
			child, err = newFromNumber(v)
		case jsonparser.Boolean:
			child, err = newFromBool(v)
		case jsonparser.Null:
			child, err = newFromNull(v)
		case jsonparser.String:
			s, err := parseStringNoQuote(v)
			if err != nil {
				return err
			}
			child = new()
			child.valueType = jsonparser.String
			child.stringValue = s
		}

		if err != nil {
			return err
		}
		o.objectChildren[key] = child
		return nil
	})

	// done
	if err != nil {
		return
	}
	return o, nil
}

// ==== type access ====

// IsObject tells whether value is an object
func (v *V) IsObject() bool {
	return v.valueType == jsonparser.Object
}

// IsArray tells whether value is an array
func (v *V) IsArray() bool {
	return v.valueType == jsonparser.Array
}

// IsString tells whether value is a string
func (v *V) IsString() bool {
	return v.valueType == jsonparser.Null
}

// IsNumber tells whether value is a number
func (v *V) IsNumber() bool {
	return v.valueType == jsonparser.Number
}

// IsFloat tells whether value is a float point number
func (v *V) IsFloat() bool {
	return v.floated
}

// IsInteger tells whether value is a fix point interger
func (v *V) IsInteger() bool {
	return v.valueType == jsonparser.Number && false == v.floated
}

// IsBoolean tells whether value is a boolean
func (v *V) IsBoolean() bool {
	return v.valueType == jsonparser.Boolean
}

// IsNull tells whether value is null
func (v *V) IsNull() bool {
	return v.valueType == jsonparser.Null
}

// ==== value access ====

// Bool returns represented bool value
func (v *V) Bool() bool {
	return v.boolValue
}

// Int returns represented int value
func (v *V) Int() int {
	return int(v.int64Value)
}

// Uint returns represented uint value
func (v *V) Uint() uint {
	return uint(v.uint64Value)
}

// Int64 returns represented int64 value
func (v *V) Int64() int64 {
	return v.int64Value
}

// Uint64 returns represented uint64 value
func (v *V) Uint64() uint64 {
	return v.uint64Value
}

// Int32 returns represented int32 value
func (v *V) Int32() int32 {
	return int32(v.int64Value)
}

// Uint32 returns represented uint32 value
func (v *V) Uint32() uint32 {
	return uint32(v.uint64Value)
}

// Float64 returns represented float64 value
func (v *V) Float64() float64 {
	return v.floatValue
}

// Float32 returns represented float32 value
func (v *V) Float32() float32 {
	return float32(v.floatValue)
}

// String returns represented string value
func (v *V) String() string {
	switch v.valueType {
	default:
		return ""
	case jsonparser.Null:
		return "null"
	case jsonparser.Number:
		return string(v.rawNumBytes)
	case jsonparser.String:
		return v.stringValue
	case jsonparser.Boolean:
		return formatBool(v.boolValue)
	case jsonparser.Object:
		return v.packObjChildren()
	case jsonparser.Array:
		return v.packArrChildren()
	}
}

func (v *V) packObjChildren() string {
	buf := bytes.Buffer{}
	v.bufObjChildren(&buf)
	return buf.String()
}

func (v *V) bufObjChildren(buf *bytes.Buffer) {
	buf.WriteRune('{')
	i := 0
	for k, v := range v.objectChildren {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(k)
		buf.WriteString(": ")
		buf.WriteString(v.String())
		i++
	}
	buf.WriteRune('}')
}

func (v *V) packArrChildren() string {
	buf := bytes.Buffer{}
	v.bufArrChildren(&buf)
	return buf.String()
}

func (v *V) bufArrChildren(buf *bytes.Buffer) {
	buf.WriteRune('[')
	v.RangeArray(func(i int, v *V) bool {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(v.String())
		i++
		return true
	})
	buf.WriteRune(']')
}

// RangeObjects goes through each children when this is an object value
func (v *V) RangeObjects(callback func(k string, v *V) bool) {
	if false == v.IsObject() {
		return
	}
	if nil == callback {
		return
	}

	for k, v := range v.objectChildren {
		ok := callback(k, v)
		if false == ok {
			break
		}
	}
	return
}

// RangeArray goes through each children when this is an array value
func (v *V) RangeArray(callback func(i int, v *V) bool) {
	if false == v.IsArray() {
		return
	}
	if nil == callback {
		return
	}

	i := 0
	for e := v.arrayChildren.Front(); e != nil; e = e.Next() {
		v := e.Value.(*V)
		ok := callback(i, v)
		if false == ok {
			break
		}
		i++
	}
}
