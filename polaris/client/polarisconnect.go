/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package client

import (
	"bufio"
	"errors"
	"fmt"
	"sync"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pkey"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	v030 "github.com/aergoio/aergo/v2/p2p/v030"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/polaris/common"
	"github.com/aergoio/aergo/v2/types"
	"github.com/libp2p/go-libp2p-core/network"
)

var ErrTooLowVersion = errors.New("aergosvr version is too low")

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

	return pcs
}

func (pcs *PolarisConnectSvc) initSvc(cfg *config.P2PConfig) {
	pcs.PrivateChain = !pcs.ntc.GenesisChainID().PublicNet
	if cfg.NPUsePolaris {
		// private network does not use public polaris
		if !pcs.PrivateChain {
			servers := make([]string, 0)
			// add hardcoded built-in servers if net is ONE net.
			if *pcs.ntc.GenesisChainID() == common.ONEMainNet {
				pcs.Logger.Info().Msg("chain is ONE Mainnet so use default polaris for mainnet")
				servers = common.MainnetMapServer
			} else if *pcs.ntc.GenesisChainID() == common.ONETestNet {
				pcs.Logger.Info().Msg("chain is ONE Testnet so use default polaris for testnet")
				servers = common.TestnetMapServer
			} else {
				pcs.Logger.Info().Msg("chain is custom public network so only custom polaris in configuration file will be used")
			}

			for _, addrStr := range servers {
				meta, err := p2putil.FromMultiAddrString(addrStr)
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
			meta, err := p2putil.FromMultiAddrString(addrStr)
			if err != nil {
				pcs.Logger.Info().Str("addr_str", addrStr).Msg("invalid polaris server address in config file ")
				continue
			}
			pcs.mapServers = append(pcs.mapServers, meta)
		}

		if len(pcs.mapServers) == 0 {
			pcs.Logger.Warn().Msg("using Polaris is enabled but no active polaris server found. node discovery by polaris will not works well")
		} else {
			pcs.Logger.Info().Array("polarises", p2putil.NewLogPeerMetasMarshaller(pcs.mapServers, 10)).Msg("using Polaris")
		}
	} else {
		pcs.Logger.Info().Msg("node discovery by Polaris is disabled by configuration.")
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
		//		pcs.Logger.Debug().Interface("msg", msg)
	}
}

func (pcs *PolarisConnectSvc) queryPeers(msg *message.MapQueryMsg) *message.MapQueryRsp {
	succ := 0
	resultPeers := make([]*types.PeerAddress, 0, msg.Count)
	for _, meta := range pcs.mapServers {
		addrs, err := pcs.connectAndQuery(meta, msg.BestBlock.Hash, msg.BestBlock.Header.BlockNo)
		if err != nil {
			if err == ErrTooLowVersion {
				pcs.Logger.Error().Err(err).Str("polarisID", p2putil.ShortForm(meta.ID)).Msg("Polaris responded this aergosvr is too low, check and upgrade aergosvr")
			} else {
				pcs.Logger.Warn().Err(err).Str("polarisID", p2putil.ShortForm(meta.ID)).Msg("failed to get peer addresses")
			}
			continue
		}
		// duplicated peers will be filtered out by caller (more precisely, p2p actor)
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
	pcs.Logger.Debug().Str("polarisID", peerID.String()).Msg("Sending map query")

	rw := v030.NewV030ReadWriter(bufio.NewReader(s), bufio.NewWriter(s), nil)

	peerAddress := pcs.ntc.SelfMeta().ToPeerAddress()
	chainBytes, _ := pcs.ntc.GenesisChainID().Bytes()
	peerStatus := &types.Status{Sender: &peerAddress, BestBlockHash: bestHash, BestHeight: bestHeight, ChainID: chainBytes,
		Version: p2pkey.NodeVersion()}

	return pcs.queryToPolaris(mapServerMeta, rw, peerStatus)
}

func (pcs *PolarisConnectSvc) queryToPolaris(mapServerMeta p2pcommon.PeerMeta, rw p2pcommon.MsgReadWriter, peerStatus *types.Status) ([]*types.PeerAddress, error) {
	// receive input
	err := pcs.sendRequest(peerStatus, mapServerMeta, pcs.exposeself, 100, rw)
	if err != nil {
		return nil, err
	}
	_, resp, err := pcs.readResponse(mapServerMeta, rw)
	if err != nil {
		return nil, err
	}
	switch resp.Status {
	case types.ResultStatus_OK:
		return resp.Addresses, nil
	case types.ResultStatus_FAILED_PRECONDITION:
		if resp.Message == common.TooOldVersionMsg {
			return nil, ErrTooLowVersion
		} else {
			return nil, fmt.Errorf("remote error %s", resp.Status.String())
		}
	default:
		return nil, fmt.Errorf("remote error %s", resp.Status.String())
	}
}

func (pcs *PolarisConnectSvc) sendRequest(status *types.Status, mapServerMeta p2pcommon.PeerMeta, register bool, size int, wt p2pcommon.MsgReadWriter) error {
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
func (pcs *PolarisConnectSvc) readResponse(mapServerMeta p2pcommon.PeerMeta, rd p2pcommon.MsgReadWriter) (p2pcommon.Message, *types.MapResponse, error) {
	data, err := rd.ReadMsg()
	if err != nil {
		return nil, nil, err
	}
	queryResp := &types.MapResponse{}
	err = p2putil.UnmarshalMessageBody(data.Payload(), queryResp)
	if err != nil {
		return data, nil, err
	}
	// old version of polaris will return old formatted PeerAddress, so conversion to new format is needed.
	for _, addr := range queryResp.Addresses {
		if len(addr.Addresses) == 0 {
			ma, err := types.ToMultiAddr(addr.Address, addr.Port)
			if err != nil {
				continue
			}
			addr.Addresses = []string{ma.String()}
		}
	}
	pcs.Logger.Debug().Str(p2putil.LogPeerID, mapServerMeta.ID.String()).Int("peer_cnt", len(queryResp.Addresses)).Msg("Received map query response")

	return data, queryResp, nil
}

func (pcs *PolarisConnectSvc) onPing(s network.Stream) {
	peerID := s.Conn().RemotePeer()
	pcs.Logger.Debug().Str("polarisID", peerID.String()).Msg("Received ping from polaris (maybe)")

	rw := v030.NewV030ReadWriter(bufio.NewReader(s), bufio.NewWriter(s), nil)
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
	respMsg := common.NewPolarisRespMessage(msgID, req.ID(), p2pcommon.PingResponse, bytes)
	err = rw.WriteMsg(respMsg)
	if err != nil {
		return
	}
}
