/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package client

import (
	"bufio"
	"fmt"
	"github.com/aergoio/aergo/p2p/p2pkey"
	"github.com/aergoio/aergo/p2p/subproto"
	"github.com/aergoio/aergo/p2p/v030"
	"github.com/aergoio/aergo/polaris/common"
	"github.com/libp2p/go-libp2p-core/network"
	"sync"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
)

// PeerMapService is
type PolarisConnectSvc struct {
	*component.BaseComponent

	PrivateChain bool

	mapServers []p2pcommon.PeerMeta
	exposeself bool

	ntc p2pcommon.NTContainer
	nt  p2pcommon.NetworkTransport

	rwmutex *sync.RWMutex
}

func NewPolarisConnectSvc(cfg *config.P2PConfig, ntc p2pcommon.NTContainer) *PolarisConnectSvc {
	pcs := &PolarisConnectSvc{
		rwmutex:    &sync.RWMutex{},
		exposeself: cfg.NPExposeSelf,
	}
	pcs.BaseComponent = component.NewBaseComponent(message.MapSvc, pcs, log.NewLogger("pcs"))

	// init
	pcs.ntc = ntc
	pcs.initSvc(cfg)
	// TODO need more pretty way to get chainID

	return pcs
}

func (pcs *PolarisConnectSvc) initSvc(cfg *config.P2PConfig) {
	pcs.PrivateChain = !pcs.ntc.ChainID().PublicNet
	if cfg.NPUsePolaris {
		// private network does not use public polaris
		if !pcs.PrivateChain {
			servers := make([]string, 0)
			// add hardcoded built-in servers if net is ONE net.
			if *pcs.ntc.ChainID() == common.ONEMainNet {
				pcs.Logger.Info().Msg("chain is ONE Mainnet so use default polaris for mainnet")
				servers = common.MainnetMapServer
			} else if *pcs.ntc.ChainID() == common.ONETestNet {
				pcs.Logger.Info().Msg("chain is ONE Testnet so use default polaris for testnet")
				servers = common.TestnetMapServer
			} else {
				pcs.Logger.Info().Msg("chain is custom public network so only custom polaris in configuration file will be used")
			}

			for _, addrStr := range servers {
				meta, err := p2putil.ParseMultiAddrString(addrStr)
				if err != nil {
					pcs.Logger.Info().Str("addr_str", addrStr).Msg("invalid polaris server address in base setting ")
					continue
				}
				pcs.mapServers = append(pcs.mapServers, meta)
			}
		} else {
			pcs.Logger.Info().Msg("chain is private so only using polaris in config file")
		}
		// append custom polarises set in configuration file
		for _, addrStr := range cfg.NPAddPolarises {
			meta, err := p2putil.ParseMultiAddrString(addrStr)
			if err != nil {
				pcs.Logger.Info().Str("addr_str", addrStr).Msg("invalid polaris server address in config file ")
				continue
			}
			pcs.mapServers = append(pcs.mapServers, meta)
		}

		if len(pcs.mapServers) == 0 {
			pcs.Logger.Warn().Msg("using polais is enabled but no active polaris server found. node discovery by polaris is disabled")
		} else {
			pcs.Logger.Info().Array("polarises", p2putil.NewLogPeerMetasMarshaler(pcs.mapServers, 10)).Msg("using polaris")
		}
	} else {
		pcs.Logger.Info().Msg("node discovery by polaris is disabled configuration.")
	}
}

func (pcs *PolarisConnectSvc) BeforeStart() {}

func (pcs *PolarisConnectSvc) AfterStart() {
	pcs.nt = pcs.ntc.GetNetworkTransport()
	pcs.nt.AddStreamHandler(common.PolarisPingSub, pcs.onPing)

}

func (pcs *PolarisConnectSvc) BeforeStop() {
	pcs.nt.RemoveStreamHandler(common.PolarisPingSub)
}

func (pcs *PolarisConnectSvc) Statistics() *map[string]interface{} {
	return nil
	//dummy := make(map[string]interface{})
	//return &dummy
}

func (pcs *PolarisConnectSvc) Receive(context actor.Context) {
	rawMsg := context.Message()
	switch msg := rawMsg.(type) {
	case *message.MapQueryMsg:
		pcs.Hub().Tell(message.P2PSvc, pcs.queryPeers(msg))
	default:
		//		pcs.Logger.Debug().Interface("msg", msg) // TODO: temporal code for resolve compile error
	}
}

func (pcs *PolarisConnectSvc) queryPeers(msg *message.MapQueryMsg) *message.MapQueryRsp {
	succ := 0
	resultPeers := make([]*types.PeerAddress, 0, msg.Count)
	for _, meta := range pcs.mapServers {
		addrs, err := pcs.connectAndQuery(meta, msg.BestBlock.Hash, msg.BestBlock.Header.BlockNo)
		if err != nil {
			pcs.Logger.Warn().Err(err).Str("map_id", p2putil.ShortForm(meta.ID)).Msg("faild to get peer addresses")
			continue
		}
		// FIXME delete duplicated peers
		resultPeers = append(resultPeers, addrs...)
		succ++
	}
	err := error(nil)
	if succ == 0 {
		err = fmt.Errorf("all servers of polaris are down")
	}
	pcs.Logger.Debug().Int("peer_cnt", len(resultPeers)).Msg("Got map response and send back")
	resp := &message.MapQueryRsp{Peers: resultPeers, Err: err}
	return resp
}

func (pcs *PolarisConnectSvc) connectAndQuery(mapServerMeta p2pcommon.PeerMeta, bestHash []byte, bestHeight uint64) ([]*types.PeerAddress, error) {
	s, err := pcs.nt.GetOrCreateStreamWithTTL(mapServerMeta, common.PolarisConnectionTTL, common.PolarisMapSub)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	peerID := s.Conn().RemotePeer()
	if peerID != mapServerMeta.ID {
		return nil, fmt.Errorf("internal error peerid mismatch, exp %s, actual %s", mapServerMeta.ID.Pretty(), peerID.Pretty())
	}
	pcs.Logger.Debug().Str(p2putil.LogPeerID, peerID.String()).Msg("Sending map query")

	rw := v030.NewV030ReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	peerAddress := pcs.nt.SelfMeta().ToPeerAddress()
	chainBytes, _ := pcs.ntc.ChainID().Bytes()
	peerStatus := &types.Status{Sender: &peerAddress, BestBlockHash: bestHash, BestHeight: bestHeight, ChainID: chainBytes,
		Version:p2pkey.NodeVersion()}
	// receive input
	err = pcs.sendRequest(peerStatus, mapServerMeta, pcs.exposeself, 100, rw)
	if err != nil {
		return nil, err
	}
	_, resp, err := pcs.readResponse(mapServerMeta, rw)
	if err != nil {
		return nil, err
	}
	if resp.Status == types.ResultStatus_OK {
		return resp.Addresses, nil
	}
	return nil, fmt.Errorf("remote error %s", resp.Status.String())
}

func (pcs *PolarisConnectSvc) sendRequest(status *types.Status, mapServerMeta p2pcommon.PeerMeta, register bool, size int, wt p2pcommon.MsgWriter) error {
	msgID := p2pcommon.NewMsgID()
	queryReq := &types.MapQuery{Status: status, Size: int32(size), AddMe: register, Excludes: [][]byte{[]byte(mapServerMeta.ID)}}
	bytes, err := p2putil.MarshalMessageBody(queryReq)
	if err != nil {
		return err
	}
	reqMsg := common.NewPolarisMessage(msgID, common.MapQuery, bytes)

	return wt.WriteMsg(reqMsg)
}

// tryAddPeer will do check connecting peer and add. it will return peer meta information received from
// remote peer setup some
func (pcs *PolarisConnectSvc) readResponse(mapServerMeta p2pcommon.PeerMeta, rd p2pcommon.MsgReader) (p2pcommon.Message, *types.MapResponse, error) {
	data, err := rd.ReadMsg()
	if err != nil {
		return nil, nil, err
	}
	queryResp := &types.MapResponse{}
	err = p2putil.UnmarshalMessageBody(data.Payload(), queryResp)
	if err != nil {
		return data, nil, err
	}
	pcs.Logger.Debug().Str(p2putil.LogPeerID, mapServerMeta.ID.String()).Int("peer_cnt", len(queryResp.Addresses)).Msg("Received map query response")

	return data, queryResp, nil
}

func (pcs *PolarisConnectSvc) onPing(s network.Stream) {
	peerID := s.Conn().RemotePeer()
	pcs.Logger.Debug().Str(p2putil.LogPeerID, peerID.String()).Msg("Received ping from polaris (maybe)")

	rw := v030.NewV030ReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	defer s.Close()

	req, err := rw.ReadMsg()
	if err != nil {
		return
	}
	pingReq := &types.Ping{}
	err = p2putil.UnmarshalMessageBody(req.Payload(), pingReq)
	if err != nil {
		return
	}
	// TODO: check if sender is known polaris or peer and it not, ban or write to blacklist .
	pingResp := &types.Ping{}
	bytes, err := p2putil.MarshalMessageBody(pingResp)
	if err != nil {
		return
	}
	msgID := p2pcommon.NewMsgID()
	respMsg := common.NewPolarisRespMessage(msgID, req.ID(), subproto.PingResponse, bytes)
	err = rw.WriteMsg(respMsg)
	if err != nil {
		return
	}
}
