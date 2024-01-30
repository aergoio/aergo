package jsonrpc

import (
	"encoding/json"
	"fmt"
)

type EncodingType int

const (
	Raw EncodingType = 0 + iota
	Base58
)

func MarshalJSON(i interface{}) string {
	jsonout, err := json.MarshalIndent(i, "", " ")
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return ""
	}
	return string(jsonout)
}
