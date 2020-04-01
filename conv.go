package jsonvalue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

func parseUint(b []byte) (uint64, error) {
	return strconv.ParseUint(string(b), 10, 64)
}

func parseInt(b []byte) (int64, error) {
	return strconv.ParseInt(string(b), 10, 64)
}

func parseFloat(b []byte) (float64, error) {
	return strconv.ParseFloat(string(b), 64)
}

func parseString(b []byte) (string, error) {
	len := len(b)
	if len < 2 {
		return "", fmt.Errorf("invalid string")
	}

	var ret string
	err := json.Unmarshal(b, &ret)
	return ret, err
}

func parseStringNoQuote(b []byte) (string, error) {
	buf := bytes.Buffer{}
	buf.WriteRune('"')
	buf.Write(b)
	buf.WriteRune('"')
	return parseString(buf.Bytes())
}

func formatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// reference:
// - [UTF-16](https://zh.wikipedia.org/zh-cn/UTF-16)
// - [JavaScript has a Unicode problem](https://mathiasbynens.be/notes/javascript-unicode)
// - [Meaning of escaped unicode characters in JSON](https://stackoverflow.com/questions/21995410/meaning-of-escaped-unicode-characters-in-json)
func escapeUnicodeToBuff(buf *bytes.Buffer, r rune) {
	if r <= '\u0127' {
		buf.WriteRune(r)
		return
	}
	if r <= '\uffff' {
		buf.WriteString(fmt.Sprintf("\\u%04X", r))
		return
	}
	// if r > 0x10FFFF {
	// 	// invalid unicode
	// 	buf.WriteRune(r)
	// 	return
	// }

	r = r - 0x10000
	lo := r & 0x003FF
	hi := (r & 0xFFC00) >> 10
	buf.WriteString(fmt.Sprintf("\\u%04X", hi+0xD800))
	buf.WriteString(fmt.Sprintf("\\u%04X", lo+0xDC00))
	return
}

func escapeStringToBuff(s string, buf *bytes.Buffer) {
	for _, chr := range s {
		switch chr {
		case '"':
			buf.WriteString("\\\"")
		case '/':
			buf.WriteString("\\/")
		case '\\':
			buf.WriteString("\\\\")
		case '\b':
			buf.WriteString("\\b")
		case '\f':
			buf.WriteString("\\f")
		case '\t':
			buf.WriteString("\\t")
		case '\n':
			buf.WriteString("\\n")
		case '\r':
			buf.WriteString("\\r")
		case '<':
			buf.WriteString("\\u003C")
		case '>':
			buf.WriteString("\\u003E")
		case '&':
			buf.WriteString("\\u0026")
		case '%':
			buf.WriteString("\\u0025") // not standard JSON encoding
		default:
			escapeUnicodeToBuff(buf, chr)
		}
	}
	return
}

func intfToInt(v interface{}) (u int, err error) {
	switch v.(type) {
	case int:
		u = int(v.(int))
	case uint:
		u = int(v.(uint))
	case int64:
		u = int(v.(int64))
	case uint64:
		u = int(v.(uint64))
	case int32:
		u = int(v.(int32))
	case uint32:
		u = int(v.(uint32))
	case int16:
		u = int(v.(int16))
	case uint16:
		u = int(v.(uint16))
	case int8:
		u = int(v.(int8))
	case uint8:
		u = int(v.(uint8))
	default:
		err = fmt.Errorf("%s is not a number", reflect.TypeOf(v).String())
	}

	return
}

// func intfToInt64(v interface{}) (i int64, err error) {
// 	switch v.(type) {
// 	case int:
// 		i = int64(v.(int))
// 	case uint:
// 		i = int64(v.(uint))
// 	case int64:
// 		i = int64(v.(int64))
// 	case uint64:
// 		i = int64(v.(uint64))
// 	case int32:
// 		i = int64(v.(int32))
// 	case uint32:
// 		i = int64(v.(uint32))
// 	case int16:
// 		i = int64(v.(int16))
// 	case uint16:
// 		i = int64(v.(uint16))
// 	case int8:
// 		i = int64(v.(int8))
// 	case uint8:
// 		i = int64(v.(uint8))
// 	default:
// 		err = fmt.Errorf("%s is not a number", reflect.TypeOf(v).String())
// 	}

// 	return
// }

func intfToString(v interface{}) (s string, err error) {
	switch v.(type) {
	case string:
		s = v.(string)
	default:
		err = fmt.Errorf("%s is not a string", reflect.TypeOf(v).String())
	}

	return
}

// func intfToJsonvalue(v interface{}) (j *V, err error) {
// 	switch v.(type) {
// 	case *V:
// 		j = v.(*V)
// 	default:
// 		err = fmt.Errorf("%s is not a *jsonvalue.V type", reflect.TypeOf(v).String())
// 	}

// 	return
// }
