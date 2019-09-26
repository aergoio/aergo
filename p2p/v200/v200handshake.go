/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package v200

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/internal/network"
	"github.com/aergoio/aergo/p2p/p2pkey"
	v030 "github.com/aergoio/aergo/p2p/v030"
	"io"
	"sort"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/aergoio/aergo/types"
)


type fakeChainID struct {
	genID types.ChainID
	versions []uint64
}
func newFC(genID types.ChainID, vers... uint64) fakeChainID {
	sort.Sort(BlkNoDesc(vers))
	return fakeChainID{genID:genID, versions:vers}
}
func (f fakeChainID) getChainID(no types.BlockNo) *types.ChainID{
	cp := f.genID
	for i:=len(f.versions)-1; i>=0 ; i-- {
		if f.versions[i] <= no {
			cp.Version = int32(i+2)
		}
	}
	return &cp
}
type BlkNoDesc []uint64

func (a BlkNoDesc) Len() int           { return len(a) }
func (a BlkNoDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BlkNoDesc) Less(i, j int) bool { return a[i] < a[j] }

// V200Handshaker exchange status data over protocol version 1.0.0
type V200Handshaker struct {
	pm      p2pcommon.PeerManager
	actor   p2pcommon.ActorService
	logger  *log.Logger
	peerID  types.PeerID

	msgRW p2pcommon.MsgReadWriter

	localGenesisHash []byte
	vm p2pcommon.VersionedManager
}

var _ p2pcommon.VersionedHandshaker = (*V200Handshaker)(nil)

func (h *V200Handshaker) GetMsgRW() p2pcommon.MsgReadWriter {
	return h.msgRW
}

func NewV200VersionedHS(pm p2pcommon.PeerManager, actor p2pcommon.ActorService, log *log.Logger, vm p2pcommon.VersionedManager, peerID types.PeerID, rwc io.ReadWriteCloser, genesis []byte) *V200Handshaker {
	h := &V200Handshaker{pm: pm, actor: actor, logger: log, peerID: peerID, localGenesisHash:genesis, vm:vm}
	// msg format is not changed
	h.msgRW = v030.NewV030MsgPipe(rwc)
	return h
}

// handshakeOutboundPeer start handshake with outbound peer
func (h *V200Handshaker) DoForOutbound(ctx context.Context) (*types.Status, error) {
	// TODO need to check auth at first...
	h.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(h.peerID)).Msg("Starting versioned handshake for outbound peer connection")

	// find my best block
	bestBlock, err := h.actor.GetChainAccessor().GetBestBlock()
	if err != nil {
		return nil, err
	}
	localID := h.vm.GetChainID(bestBlock.Header.BlockNo)

	status, err := createLocalStatus(h.pm, localID, bestBlock, h.localGenesisHash)
	if err != nil {
		h.logger.Warn().Err(err).Msg("Failed to create status message.")
		h.sendGoAway("internal error")
		return nil, err
	}

	err = h.sendLocalStatus(ctx, status)
	if err != nil {
		return nil, err
	}

	remotePeerStatus, err := h.receiveRemoteStatus(ctx)
	if err != nil {
		return nil, err
	}

	if err = h.checkRemoteStatus(remotePeerStatus); err != nil {
		return nil, err
	} else {
		return remotePeerStatus, nil
	}
}

func (h *V200Handshaker) sendLocalStatus(ctx context.Context, hostStatus *types.Status) error {
	var err error
	container := createMessage(p2pcommon.StatusRequest, p2pcommon.NewMsgID(), hostStatus)
	if container == nil {
		h.logger.Warn().Str(p2putil.LogPeerID, p2putil.ShortForm(h.peerID)).Msg("failed to create p2p message")
		h.sendGoAway("internal error")
		// h.logger.Warn().Str(LogPeerID, ShortForm(peerID)).Err(err).Msg("failed to create p2p message")
		return fmt.Errorf("failed to craete container message")
	}
	if err = h.msgRW.WriteMsg(container); err != nil {
		h.logger.Info().Str(p2putil.LogPeerID, p2putil.ShortForm(h.peerID)).Err(err).Msg("failed to write local status ")
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// go on
	}
	return nil
}

func (h *V200Handshaker) receiveRemoteStatus(ctx context.Context) (*types.Status, error) {
	// and wait to response status
	data, err := h.msgRW.ReadMsg()
	if err != nil {
		h.sendGoAway("malformed message")
		// h.logger.Info().Err(err).Msg("fail to decode")
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// go on
	}
	if data.Subprotocol() != p2pcommon.StatusRequest {
		if data.Subprotocol() == p2pcommon.GoAway {
			return h.handleGoAway(h.peerID, data)
		} else {
			h.logger.Info().Str(p2putil.LogPeerID, p2putil.ShortForm(h.peerID)).Str("expected", p2pcommon.StatusRequest.String()).Str("actual", data.Subprotocol().String()).Msg("unexpected message type")
			h.sendGoAway("unexpected message type")
			return nil, fmt.Errorf("unexpected message type")
		}
	}

	remotePeerStatus := &types.Status{}
	err = p2putil.UnmarshalMessageBody(data.Payload(), remotePeerStatus)
	if err != nil {
		h.sendGoAway("malformed status message")
		return nil, err
	}

	return remotePeerStatus, nil
}

func (h *V200Handshaker) checkRemoteStatus(remotePeerStatus *types.Status) error {
	// check if chainID is same or not
	remoteChainID := types.NewChainID()
	err := remoteChainID.Read(remotePeerStatus.ChainID)
	if err != nil {
		h.sendGoAway("wrong status")
		return err
	}
	localID := h.vm.GetChainID(remotePeerStatus.BestHeight)
	if !localID.Equals(remoteChainID) {
		h.sendGoAway("different chainID")
		return fmt.Errorf("different chainID : %s", remoteChainID.ToJSON())
	}

	peerAddress := remotePeerStatus.Sender
	if peerAddress == nil || network.CheckAddressType(peerAddress.Address) == network.AddressTypeError {
		h.sendGoAway("invalid peer address")
		return fmt.Errorf("invalid peer address : %s", peerAddress)
	}

	rMeta := p2pcommon.FromPeerAddress(peerAddress)
	if rMeta.ID != h.peerID {
		h.logger.Debug().Str("received_peer_id", rMeta.ID.Pretty()).Str(p2putil.LogPeerID, p2putil.ShortForm(h.peerID)).Msg("Inconsistent peerID")
		h.sendGoAway("Inconsistent peerID")
		return fmt.Errorf("inconsistent peerID")
	}

	// do additional check for genesises are identical
	genHash := h.localGenesisHash
	if !bytes.Equal(genHash, remotePeerStatus.Genesis) {
		h.sendGoAway("different genesis block")
		return fmt.Errorf("different genesis block local: %v , remote %v", enc.ToString(genHash), enc.ToString(remotePeerStatus.Genesis))
	}

	return nil
}

// DoForInbound is handle handshake from inbound peer
func (h *V200Handshaker) DoForInbound(ctx context.Context) (*types.Status, error) {
	// TODO need to check auth at first...
	h.logger.Debug().Str(p2putil.LogPeerID, p2putil.ShortForm(h.peerID)).Msg("Starting versioned handshake for inbound peer connection")

	// inbound: receive, check and send
	remotePeerStatus, err := h.receiveRemoteStatus(ctx)
	if err != nil {
		return nil, err
	}
	if err = h.checkRemoteStatus(remotePeerStatus); err != nil {
		return nil, err
	}

	bestBlock, err := h.actor.GetChainAccessor().GetBestBlock()
	if err != nil {
		return nil, err
	}
	localID := h.vm.GetChainID(bestBlock.Header.BlockNo)

	// send my status message as response
	localStatus, err := createLocalStatus(h.pm, localID, bestBlock, h.localGenesisHash)
	if err != nil {
		h.logger.Warn().Err(err).Msg("Failed to create localStatus message.")
		h.sendGoAway("internal error")
		return nil, err
	}
	err = h.sendLocalStatus(ctx, localStatus)
	if err != nil {
		return nil, err
	}
	return remotePeerStatus, nil
}

func (h *V200Handshaker) handleGoAway(peerID types.PeerID, data p2pcommon.Message) (*types.Status, error) {
	goAway := &types.GoAwayNotice{}
	if err := p2putil.UnmarshalMessageBody(data.Payload(), goAway); err != nil {
		h.logger.Warn().Str(p2putil.LogPeerID, p2putil.ShortForm(peerID)).Err(err).Msg("Remote peer sent goAway but failed to decode internal message")
		return nil, err
	}
	return nil, fmt.Errorf("remote peer refuse handshake: %s", goAway.GetMessage())
}

func (h *V200Handshaker) sendGoAway(msg string) {
	goMsg := createMessage(p2pcommon.GoAway, p2pcommon.NewMsgID(), &types.GoAwayNotice{Message: msg})
	if goMsg != nil {
		h.msgRW.WriteMsg(goMsg)
	}
}

func createLocalStatus(pm p2pcommon.PeerManager, chainID *types.ChainID, bestBlock *types.Block, genesis []byte) (*types.Status, error) {
	selfAddr := pm.SelfMeta().ToPeerAddress()
	chainIDbytes, err := chainID.Bytes()
	if err != nil {
		return nil, err
	}

	// create message data
	statusMsg := &types.Status{
		Sender:        &selfAddr,
		ChainID:       chainIDbytes,
		BestBlockHash: bestBlock.BlockHash(),
		BestHeight:    bestBlock.GetHeader().GetBlockNo(),
		NoExpose:      pm.SelfMeta().Hidden,
		Version:       p2pkey.NodeVersion(),
		Genesis:       genesis,
	}

	return statusMsg, nil
}

func createMessage(protocolID p2pcommon.SubProtocol, msgID p2pcommon.MsgID, msgBody p2pcommon.MessageBody) p2pcommon.Message {
	bytes, err := p2putil.MarshalMessageBody(msgBody)
	if err != nil {
		return nil
	}

	msg := p2pcommon.NewMessageValue(protocolID, msgID, p2pcommon.EmptyID, time.Now().UnixNano(), bytes)
	return msg
}
