/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"encoding/hex"
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
)

var (
	myChainID, theirChainID * types.ChainID
	myChainBytes, theirChainBytes []byte
)

func init() {
	myChainID = types.NewChainID()
	myChainID.Magic = "itSmain1"
	myChainBytes, _ = myChainID.Bytes()

	theirChainID = types.NewChainID()
	theirChainID.Read(myChainBytes)
	theirChainID.Magic = "itsdiff2"
	theirChainBytes, _ = theirChainID.Bytes()

}

func TestDeepEqual(t *testing.T) {
	b1, _ := myChainID.Bytes()
	b2 := make([]byte, len(b1), len(b1)<<1 )
	copy( b2, b1)

	s1 := &types.Status{ChainID:b1}
	s2 := &types.Status{ChainID:b2}

	if !reflect.DeepEqual(s1, s2) {
		t.Errorf("byte slice cant do DeepEqual! %v, %v", b1, b2)
	}

}

func TestV030StatusHS_doForOutbound(t *testing.T) {
	logger = log.NewLogger("test")
	mockActor := new(MockActorService)
	mockCA := new(MockChainAccessor)
	mockPM := new(MockPeerManager)

	dummyMeta := p2pcommon.PeerMeta{ID: dummyPeerID, IPAddress:"dummy.aergo.io"}
	dummyAddr := dummyMeta.ToPeerAddress()
	mockPM.On("SelfMeta").Return(dummyMeta)
	dummyBlock := &types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	mockActor.On("GetChainAccessor").Return(mockCA)
	mockCA.On("GetBestBlock").Return(dummyBlock, nil)

	dummyStatusMsg := &types.Status{ChainID:myChainBytes, Sender:&dummyAddr}
	nilSenderStatusMsg := &types.Status{ChainID:myChainBytes, Sender:nil}
	diffStatusMsg := &types.Status{ChainID:theirChainBytes, Sender:&dummyAddr}
	tests := []struct {
		name       string
		readReturn *types.Status
		readError  error
		writeError error
		want       *types.Status
		wantErr    bool
	}{
		{"TSuccess", dummyStatusMsg, nil, nil, dummyStatusMsg, false},
		{"TUnexpMsg", nil, nil, nil, nil, true},
		{"TRFail", dummyStatusMsg, fmt.Errorf("failed"), nil, nil, true},
		{"TRNoSender", nilSenderStatusMsg, nil, nil, nil, true},
		{"TWFail", dummyStatusMsg, nil, fmt.Errorf("failed"), nil, true},
		{"TDiffChain", diffStatusMsg, nil, nil, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyReader := new(MockReader)
			dummyWriter := new(MockWriter)
			mockRW := new(MockMsgReadWriter)

			containerMsg := &V030Message{}
			if tt.readReturn != nil {
				containerMsg.subProtocol = StatusRequest
				statusBytes, _ := MarshalMessage(tt.readReturn)
				containerMsg.payload = statusBytes
			} else {
				containerMsg.subProtocol = AddressesRequest
			}
			mockRW.On("ReadMsg").Return(containerMsg, tt.readError)
			mockRW.On("WriteMsg", mock.Anything).Return(tt.writeError)

			h := newV030StateHS(mockPM, mockActor, logger, myChainID, samplePeerID, dummyReader, dummyWriter)
			h.msgRW = mockRW
			got, err := h.doForOutbound()
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got!=nil && tt.want!= nil {
				if !reflect.DeepEqual(got.ChainID, tt.want.ChainID) {
					fmt.Printf("got:(%d) %s \n", len(got.ChainID), hex.EncodeToString(got.ChainID))
					fmt.Printf("got:(%d) %s \n", len(tt.want.ChainID), hex.EncodeToString(tt.want.ChainID))
					t.Errorf("PeerHandshaker.handshakeOutboundPeer() = %v, want %v", got.ChainID, tt.want.ChainID)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerHandshaker.handshakeOutboundPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestV030StatusHS_handshakeInboundPeer(t *testing.T) {
	// t.SkipNow()
	logger = log.NewLogger("test")
	mockActor := new(MockActorService)
	mockCA := new(MockChainAccessor)
	mockPM := new(MockPeerManager)

	dummyMeta := p2pcommon.PeerMeta{ID: dummyPeerID, IPAddress:"dummy.aergo.io"}
	dummyAddr := dummyMeta.ToPeerAddress()
	mockPM.On("SelfMeta").Return(dummyMeta)
	dummyBlock := &types.Block{Hash: dummyBlockHash, Header: &types.BlockHeader{BlockNo: dummyBlockHeight}}
	//dummyBlkRsp := message.GetBestBlockRsp{Block: dummyBlock}
	mockActor.On("GetChainAccessor").Return(mockCA)
	mockCA.On("GetBestBlock").Return(dummyBlock, nil)

	dummyStatusMsg := &types.Status{ChainID:myChainBytes, Sender:&dummyAddr}
	nilSenderStatusMsg := &types.Status{ChainID:myChainBytes, Sender:nil}
	diffStatusMsg := &types.Status{ChainID:theirChainBytes, Sender:&dummyAddr}
	tests := []struct {
		name       string
		readReturn *types.Status
		readError  error
		writeError error
		want       *types.Status
		wantErr    bool
	}{
		{"TSuccess", dummyStatusMsg, nil, nil, dummyStatusMsg, false},
		{"TUnexpMsg", nil, nil, nil, nil, true},
		{"TRFail", dummyStatusMsg, fmt.Errorf("failed"), nil, nil, true},
		{"TRNoSender", nilSenderStatusMsg, nil, nil, nil, true},
		{"TWFail", dummyStatusMsg, nil, fmt.Errorf("failed"), nil, true},
		{"TDiffChain", diffStatusMsg, nil, nil, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dummyReader := new(MockReader)
			dummyWriter := new(MockWriter)
			mockRW := new(MockMsgReadWriter)
			containerMsg := &V030Message{}
			if tt.readReturn != nil {
				containerMsg.subProtocol = StatusRequest
				statusBytes, _ := MarshalMessage(tt.readReturn)
				containerMsg.payload = statusBytes
			} else {
				containerMsg.subProtocol = AddressesRequest
			}

			mockRW.On("ReadMsg").Return(containerMsg, tt.readError)
			mockRW.On("WriteMsg", mock.Anything).Return(tt.writeError)

			h := newV030StateHS(mockPM, mockActor, logger, myChainID, samplePeerID, dummyReader, dummyWriter)
			h.msgRW = mockRW
			got, err := h.doForInbound()
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerHandshaker.handshakeInboundPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got!=nil && tt.want!= nil {
				if !reflect.DeepEqual(got.ChainID, tt.want.ChainID) {
					fmt.Printf("got:(%d) %s \n", len(got.ChainID), hex.EncodeToString(got.ChainID))
					fmt.Printf("got:(%d) %s \n", len(tt.want.ChainID), hex.EncodeToString(tt.want.ChainID))
					t.Errorf("PeerHandshaker.handshakeOutboundPeer() = %v, want %v", got.ChainID, tt.want.ChainID)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PeerHandshaker.handshakeInboundPeer() = %v, want %v", got, tt.want)
			}
		})
	}
}
