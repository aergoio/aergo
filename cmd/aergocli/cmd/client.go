/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"fmt"

	"github.com/aergoio/aergo/v2/types"
	"google.golang.org/grpc"
)

type ConnClient struct {
	types.AergoRPCServiceClient
	conn *grpc.ClientConn
}

func GetClient(serverAddr string, opts []grpc.DialOption) interface{} {
	conn := GetConn(serverAddr, opts)
	connClient := &ConnClient{
		AergoRPCServiceClient: types.NewAergoRPCServiceClient(conn),
		conn:                  conn,
	}

	return connClient
}

func GetConn(serverAddr string, opts []grpc.DialOption) *grpc.ClientConn {
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil || conn == nil {
		fmt.Println(err)
		panic("connection failed")
	}
	return conn
}

func (c *ConnClient) Close() {
	c.conn.Close()
	c.conn = nil
}
