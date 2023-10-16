/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package client

import (
	"errors"
	"testing"

	"github.com/aergoio/aergo/v2/config"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/p2p/p2pmock"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/polaris/common"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
)

type dummyNTC struct {
	nt       p2pcommon.NetworkTransport
	chainID  *types.ChainID
	selfMeta p2pcommon.PeerMeta
}

func (dntc *dummyNTC) SelfMeta() p2pcommon.PeerMeta {
	return dntc.selfMeta
}

func (dntc *dummyNTC) GetNetworkTransport() p2pcommon.NetworkTransport {
	return dntc.nt
}
func (dntc *dummyNTC) GenesisChainID() *types.ChainID {
	return dntc.chainID
}

var (
	pmapDummyCfg = &config.Config{P2P: &config.P2PConfig{}, Polaris: &config.PolarisConfig{GenesisFile: "../../examples/genesis.json"}}
	pmapDummyNTC = &dummyNTC{chainID: &types.ChainID{}}
)

// initSvc select Polarises to connect, or disable polaris
func TestPolarisConnectSvc_initSvc(t *testing.T) {
	polarisIDMain, _ := types.IDB58Decode("16Uiu2HAkuxyDkMTQTGFpmnex2SdfTVzYfPztTyK339rqUdsv3ZUa")
	polarisIDTest, _ := types.IDB58Decode("16Uiu2HAkvJTHFuJXxr15rFEHsJWnyn1QvGatW2E9ED9Mvy4HWjVF")
	dummyPeerID2, _ := types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
	polar2 := "/ip4/172.21.1.2/tcp/8915/p2p/16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm"
	dummyPeerID3, _ := types.IDB58Decode("16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD")
	polar3 := "/ip4/172.22.2.3/tcp/8915/p2p/16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD"

	customChainID := types.ChainID{Magic: "unittest.blocko.io"}
	type args struct {
		use       bool
		polarises []string

		chainID *types.ChainID
	}
	tests := []struct {
		name string
		args args

		wantCnt int
		peerIDs []types.PeerID
	}{
		//
		{"TAergoNoPolaris", args{false, nil, &common.ONEMainNet}, 0, []types.PeerID{}},
		{"TAergoMainDefault", args{true, nil, &common.ONEMainNet}, 1, []types.PeerID{polarisIDMain}},
		{"TAergoMainPlusCfg", args{true, []string{polar2, polar3}, &common.ONEMainNet}, 3, []types.PeerID{polarisIDMain, dummyPeerID2, dummyPeerID3}},
		{"TAergoTestDefault", args{true, nil, &common.ONETestNet}, 1, []types.PeerID{polarisIDTest}},
		{"TAergoTestPlusCfg", args{true, []string{polar2, polar3}, &common.ONETestNet}, 3, []types.PeerID{polarisIDTest, dummyPeerID2, dummyPeerID3}},
		{"TCustom", args{true, nil, &customChainID}, 0, []types.PeerID{}},
		{"TCustomPlusCfg", args{true, []string{polar2, polar3}, &customChainID}, 2, []types.PeerID{dummyPeerID2, dummyPeerID3}},
		{"TWrongPolarisAddr", args{true, []string{"/ip4/256.256.1.1/tcp/8915/p2p/16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD"}, &customChainID}, 0, []types.PeerID{}},
		{"TWrongPolarisAddr2", args{true, []string{"/egwgew5/p2p/16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD"}, &customChainID}, 0, []types.PeerID{}},
		{"TWrongPolarisAddr3", args{true, []string{"/dns/nowhere1234.io/tcp/8915/p2p/16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD"}, &customChainID}, 1, []types.PeerID{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockNT := p2pmock.NewMockNetworkTransport(ctrl)
			pmapDummyNTC.nt = mockNT
			pmapDummyNTC.chainID = tt.args.chainID

			cfg := config.NewServerContext("", "").GetDefaultP2PConfig()
			cfg.NPUsePolaris = tt.args.use
			cfg.NPAddPolarises = tt.args.polarises

			pcs := NewPolarisConnectSvc(cfg, pmapDummyNTC)

			if len(pcs.mapServers) != tt.wantCnt {
				t.Errorf("NewPolarisConnectSvc() = %v, want %v", len(pcs.mapServers), tt.wantCnt)
			}
			for _, wantPeerID := range tt.peerIDs {
				found := false
				for _, polarisMeta := range pcs.mapServers {
					if wantPeerID == polarisMeta.ID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("initSvc() want exist %v but not ", wantPeerID)
				}
			}
		})
	}
}

func TestPolarisConnectSvc_BeforeStop(t *testing.T) {

	type fields struct {
		BaseComponent *component.BaseComponent
	}
	tests := []struct {
		name   string
		fields fields

		calledStreamHandler bool
	}{
		{"TNot", fields{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockNT := p2pmock.NewMockNetworkTransport(ctrl)
			pmapDummyNTC.nt = mockNT
			pms := NewPolarisConnectSvc(pmapDummyCfg.P2P, pmapDummyNTC)

			mockNT.EXPECT().AddStreamHandler(common.PolarisPingSub, gomock.Any()).Times(1)
			mockNT.EXPECT().RemoveStreamHandler(common.PolarisPingSub).Times(1)

			pms.AfterStart()

			pms.BeforeStop()

			ctrl.Finish()
		})
	}
}

func TestPolarisConnectSvc_queryToPolaris(t *testing.T) {
	sampleCahinID := &types.ChainID{Consensus: "dpos"}
	sm := p2pcommon.PeerMeta{ID: types.RandomPeerID()}
	ss := &types.Status{}

	sErr := errors.New("send erroor")
	rErr := errors.New("send erroor")

	succR := &types.MapResponse{Status: types.ResultStatus_OK, Addresses: []*types.PeerAddress{{}}}
	oldVerR := &types.MapResponse{Status: types.ResultStatus_FAILED_PRECONDITION, Message: common.TooOldVersionMsg}
	otherR := &types.MapResponse{Status: types.ResultStatus_INVALID_ARGUMENT, Message: "arg is wrong"}

	type args struct {
		mapServerMeta p2pcommon.PeerMeta
		peerStatus    *types.Status
	}
	tests := []struct {
		name string
		args args

		sendErr error
		readErr error
		mapResp *types.MapResponse

		wantCnt int
		wantErr bool
	}{
		{"TSingle", args{sm, ss}, nil, nil, succR, 1, false},
		{"TSendFail", args{sm, ss}, sErr, nil, succR, 0, true},
		{"TRecvFail", args{sm, ss}, nil, rErr, succR, 0, true},
		{"TOldVersion", args{sm, ss}, nil, nil, oldVerR, 0, true},
		{"TFailResp", args{sm, ss}, nil, nil, otherR, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cfg := config.NewServerContext("", "").GetDefaultP2PConfig()
			cfg.NPUsePolaris = true

			mockNTC := p2pmock.NewMockNTContainer(ctrl)
			mockNTC.EXPECT().GenesisChainID().Return(sampleCahinID).AnyTimes()

			mockRW := p2pmock.NewMockMsgReadWriter(ctrl)

			payload, _ := p2putil.MarshalMessageBody(tt.mapResp)
			dummyData := common.NewPolarisMessage(p2pcommon.NewMsgID(), common.MapResponse, payload)

			mockRW.EXPECT().WriteMsg(gomock.AssignableToTypeOf(&common.PolarisMessage{})).Return(tt.sendErr)
			if tt.sendErr == nil {
				mockRW.EXPECT().ReadMsg().Return(dummyData, tt.readErr)
			}
			pcs := NewPolarisConnectSvc(cfg, mockNTC)

			got, err := pcs.queryToPolaris(tt.args.mapServerMeta, mockRW, tt.args.peerStatus)
			if (err != nil) != tt.wantErr {
				t.Errorf("connectAndQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantCnt {
				t.Errorf("connectAndQuery() len(got) = %v, want %v", len(got), tt.wantCnt)
			}
		})
	}
}
