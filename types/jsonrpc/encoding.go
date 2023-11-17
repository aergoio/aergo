package jsonrpc

import (
	"encoding/json"
	"fmt"

	"github.com/aergoio/aergo/v2/internal/enc/proto"
)

type EncodingType int

const (
	Raw EncodingType = 0 + iota
	Base58
)

// Deprecated - TODO remove
func JSON(pb proto.Message) string {
	jsonout, err := json.MarshalIndent(pb, "", " ")
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return ""
	}
	return string(jsonout)
}

func MarshalJSON(i interface{}) string {
	jsonout, err := json.MarshalIndent(i, "", " ")
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return ""
	}
	return string(jsonout)
}
