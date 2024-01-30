/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package cmd

import (
	"fmt"

	"github.com/aergoio/aergo/v2/types"
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
