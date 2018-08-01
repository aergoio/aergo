/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"errors"
	"net"
	"reflect"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
	peer "github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
	uuid "github.com/satori/go.uuid"
)

// ActorService is collection of helper methods to use actor
// FIXME move to more general package. it used in p2p and rpc
type ActorService interface {
	// SendRequest send actor request, which does not need to get return value, and forget it.
	SendRequest(actor string, msg interface{})
	// CallReqeust send actor request and wait the handling of that message to finished,
	// and get return value.
	CallRequest(actor string, msg interface{}) (interface{}, error)
	// FutureRequest send actor reqeust and get the Future object to get the state and return value of message
	FutureRequest(actor string, msg interface{}) *actor.Future
}

func extractBlockFromRequest(rawResponse interface{}, err error) (*types.Block, error) {
	if err != nil {
		return nil, err
	}
	var blockRsp *message.GetBlockRsp
	switch v := rawResponse.(type) {
	case message.GetBlockRsp:
		blockRsp = &v
	case message.GetBestBlockRsp:
		blockRsp = (*message.GetBlockRsp)(&v)
	case message.GetBlockByNoRsp:
		blockRsp = (*message.GetBlockRsp)(&v)
	default:
		panic("unexpected data type " + reflect.TypeOf(rawResponse).Name() + "is passed. check if there is a bug. ")
	}
	return extractBlock(blockRsp)
}

func extractBlock(from *message.GetBlockRsp) (*types.Block, error) {
	if nil != from.Err {
		return nil, from.Err
	}
	return from.Block, nil

}

func extractTXsFromRequest(rawResponse interface{}, err error) ([]*types.Tx, bool) {
	if err != nil {
		return nil, false
	}
	var rsponse *message.MemPoolGetRsp
	switch v := rawResponse.(type) {
	case *message.MemPoolGetRsp:
		rsponse = v
	case message.MemPoolGetRsp:
		rsponse = &v
	default:
		panic("unexpected data type " + reflect.TypeOf(rawResponse).Name() + "is passed. check if there is a bug. ")
	}
	return extractTXs(rsponse)
}

func extractTXs(from *message.MemPoolGetRsp) ([]*types.Tx, bool) {
	if nil != from.Err {
		return nil, false
	}
	return from.Txs, true
}

func getIP(a *types.PeerAddress) net.IP {
	return net.IP(a.Address)
}

func setIP(a *types.PeerAddress, ipAddress net.IP) {
	a.Address = []byte(ipAddress)
}

// RandomUUID generate random UUID and return in form of string
func RandomUUID() string {
	return uuid.Must(uuid.NewV4()).String()
}

func externalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip, nil
		}
	}
	return nil, errors.New("no external ip address found")
}

// warnLogUnknownPeer log warning that tell unknown peer sent message
func warnLogUnknownPeer(logger log.ILogger, protocol protocol.ID, peerID peer.ID) {
	logger.Warnf("Message %v from Unknown peer %s, ignoring it.", protocol, peerID.Pretty())

}

func debugLogReceiveMsg(logger log.ILogger, protocol protocol.ID, msgID string, peerID peer.ID,
	additional interface{}) {
	if additional != nil {
		//		logger.Debugf("Received %v:%s request from %s. %v", protocol, msgID, peerID.Pretty(),
		//			additional)
	} else {
		logger.Debugf("Received %v:%s request from %s", protocol, msgID, peerID.Pretty())
	}
}
