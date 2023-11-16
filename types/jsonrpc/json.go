package jsonrpc

import (
	"encoding/json"
	"fmt"

	"github.com/aergoio/aergo/v2/internal/enc/proto"
)

// JSON converts protobuf message(struct) to json notation
func JSON(pb proto.Message) string {
	jsonout, err := json.MarshalIndent(pb, "", " ")
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return ""
	}
	return string(jsonout)
}

func B58JSON(i interface{}) string {
	jsonout, err := json.MarshalIndent(i, "", " ")
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return ""
	}
	return string(jsonout)
}
