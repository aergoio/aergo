/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package util

import (
	"bytes"
	"fmt"

	"github.com/aergoio/aergo/types"
	"github.com/gogo/protobuf/jsonpb"
	protobuf "github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

type ConnClient struct {
	types.AergoRPCServiceClient
	conn *grpc.ClientConn
}

func GetClient(serverAddr string, opts []grpc.DialOption) interface{} {
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil || conn == nil {
		fmt.Println(err)
		panic("connection failed")
	}

	connClient := &ConnClient{
		AergoRPCServiceClient: types.NewAergoRPCServiceClient(conn),
		conn: conn,
	}

	return connClient
}

func (c *ConnClient) Close() {
	c.conn.Close()
	c.conn = nil
}

// JSON converts protobuf message(struct) to json notation
func JSON(pb protobuf.Message) string {
	var w bytes.Buffer
	var marshaler jsonpb.Marshaler
	marshaler.Indent = "\t"
	err := marshaler.Marshal(&w, pb)
	if err != nil {
		return "[marshal fail]"
	}
	return w.String()
}
