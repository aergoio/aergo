/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package server

import (
	"bufio"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	v030 "github.com/aergoio/aergo/v2/p2p/v030"
	"github.com/aergoio/aergo/v2/polaris/common"
	"github.com/aergoio/aergo/v2/types"
)

type PeerHealth int

// PeersState
const (
	PeerHealth_GOOD PeerHealth = 0
	PeerHealth_MID  PeerHealth = 1
	PeerHealth_BAD  PeerHealth = 2
)

type peerState struct {
	*PeerMapService

	conn      p2pcommon.RemoteConn
	meta      p2pcommon.PeerMeta
	addr      types.PeerAddress
	connected time.Time

	// temporary means it does not affect current peer registry. TODO refactor more pretty way
	temporary bool

	bestHash   []byte
	bestNo     int64
	lCheckTime time.Time
	contFail   int32
}

func (hc *peerState) health() PeerHealth {
	// TODO make more robust if needed
	switch {
	case atomic.LoadInt32(&hc.contFail) == 0:
		return PeerHealth_GOOD
	default:
		return PeerHealth_BAD
	}
}

func (hc *peerState) lastCheck() time.Time {
	return hc.lCheckTime
}

func (hc *peerState) check(wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()
	success, err := hc.checkConnect(timeout)

	if !hc.temporary {
		if success == nil || err != nil {
			hc.unregisterPeer(hc.meta.ID)
		} else if hc.health() == PeerHealth_BAD {
			hc.unregisterPeer(hc.meta.ID)
		}
	}
}

func (hc *peerState) checkConnect(timeout time.Duration) (*types.Ping, error) {
	hc.Logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(hc.meta.ID)).Msg("staring up healthcheck")
	hc.lCheckTime = time.Now()
	s, err := hc.nt.GetOrCreateStreamWithTTL(hc.meta, PolarisPingTTL, common.PolarisPingSub)
	if err != nil {
		hc.contFail++
		hc.Logger.Debug().Err(err).Msg("Healthcheck failed to get network stream")
		hc.unregisterPeer(hc.meta.ID)
		return nil, err
	}
	defer s.Close()

	rw := v030.NewV030ReadWriter(bufio.NewReader(s), bufio.NewWriter(s), nil)
	pc := &pingChecker{peerState: hc, rw: rw}
	pingResp, err := p2putil.InvokeWithTimer(pc, time.NewTimer(timeout))
	if pingResp.(*types.Ping) == nil {
		return nil, fmt.Errorf("ping error")
	}
	if err != nil {
		return nil, err
	}
	return pingResp.(*types.Ping), nil
}

// this method MUST be called in same go routine as AergoPeer.RunPeer()
func (hc *peerState) sendPing(wt p2pcommon.MsgReadWriter) (p2pcommon.MsgID, error) {
	// find my best block
	ping := &types.Ping{}
	msgID := p2pcommon.NewMsgID()
	pingMsg, err := createV030Message(msgID, EmptyMsgID, p2pcommon.PingRequest, ping)
	if err != nil {
		hc.Logger.Warn().Err(err).Msg("failed to create ping message")
		return EmptyMsgID, err
	}

	err = wt.WriteMsg(pingMsg)
	if err != nil {
		hc.Logger.Warn().Err(err).Msg("failed to write ping message")
		return EmptyMsgID, err
	}

	return msgID, nil
}

// tryAddPeer will do check connecting peer and add. it will return peer meta information received from
// remote peer setup some
func (hc *peerState) receivePingResp(reqID p2pcommon.MsgID, rd p2pcommon.MsgReadWriter) (p2pcommon.Message, *types.Ping, error) {
	resp, err := rd.ReadMsg()
	if err != nil {
		return nil, nil, err
	}
	if resp.Subprotocol() != p2pcommon.PingResponse || reqID != resp.OriginalID() {
		return nil, nil, fmt.Errorf("Not expected response %s : req_id=%s", resp.Subprotocol().String(), resp.OriginalID().String())
	}
	pingResp := &types.Ping{}
	err = p2putil.UnmarshalMessageBody(resp.Payload(), pingResp)
	if err != nil {
		return resp, nil, err
	}

	return resp, pingResp, nil
}

// pingChecker has ttl and will try to
type pingChecker struct {
	*peerState
	rw     p2pcommon.MsgReadWriter
	cancel int32
}

func (pc *pingChecker) DoCall(done chan<- interface{}) {
	var pingResp *types.Ping = nil
	defer func() {
		if pingResp != nil {
			atomic.StoreInt32(&pc.contFail, 0)
		} else {
			atomic.AddInt32(&pc.contFail, 1)
		}
		done <- pingResp
	}()

	reqID, err := pc.sendPing(pc.rw)
	if err != nil {
		pc.Logger.Debug().Err(err).Msg("Healthcheck failed to send ping message")
		return
	}
	if atomic.LoadInt32(&pc.cancel) != 0 {
		return
	}
	_, pingResp, err = pc.receivePingResp(reqID, pc.rw)
	if err != nil {
		pc.Logger.Debug().Err(err).Msg("Healthcheck failed to receive ping response")
		return
	}
	if atomic.LoadInt32(&pc.cancel) != 0 {
		pingResp = nil
		return
	}

	pc.Logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(pc.meta.ID)).Interface("ping_resp", pingResp).Msg("Healthcheck finished successful")
	return

}

func (pc *pingChecker) Cancel() {
	atomic.StoreInt32(&pc.cancel, 1)
}
