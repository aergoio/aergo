/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"fmt"

	"github.com/aergoio/aergo/v2/cmd/aergocli/util/encoding/json"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

type PolarisClient struct {
	types.PolarisRPCServiceClient
	conn *grpc.ClientConn
}

func GetClient(serverAddr string, opts []grpc.DialOption) interface{} {
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil || conn == nil {
		fmt.Println(err)
		panic("connection failed")
	}

	connClient := &PolarisClient{
		PolarisRPCServiceClient: types.NewPolarisRPCServiceClient(conn),
		conn:                    conn,
	}

	return connClient
}

func (c *PolarisClient) Close() {
	c.conn.Close()
	c.conn = nil
}

// JSON converts protobuf message(struct) to json notation
func JSON(pb proto.Message) string {
	jsonout, err := json.MarshalIndent(pb, "", " ")
	if err != nil {
		fmt.Printf("Failed: %s\n", err.Error())
		return ""
	}
	return string(jsonout)
}
