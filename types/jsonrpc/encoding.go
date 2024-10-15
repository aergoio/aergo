package jsonrpc

import (
	"encoding/json"
	"fmt"
	"strings"
)

type EncodingType int

const (
	Raw EncodingType = 0 + iota
	Base58
	Obj
)

func ParseEncodingType(s string) EncodingType {
	flag := strings.ToLower(s)
	switch flag {
	case "raw":
		return Raw
	case "obj":
		return Obj
	case "base58":
		fallthrough
	default:
		return Base58
	}
}

func MarshalJSON(i interface{}) string {
	jsonout, err := json.MarshalIndent(i, "", " ")
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return ""
	}
	return string(jsonout)
}
