/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package message

const HelloService = "HelloService"

type HelloReq struct{ Who string }

type HelloRsp struct{ Greeting string }
