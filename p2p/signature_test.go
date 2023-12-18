/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/libp2p/go-libp2p/core/crypto"
	"testing"

	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func Test_defaultMsgSigner_signMsg(t *testing.T) {
	t.Run("TSameKey", func(t *testing.T) {
		// msg and msg2 is same at first
		sampleMsg1 := &types.P2PMessage{Header: &types.MsgHeader{Subprotocol: p2pcommon.PingResponse.Uint32(), Length: 5}, Data: []byte{0, 1, 2, 3, 4}}
		sampleMsg2 := &types.P2PMessage{Header: &types.MsgHeader{Subprotocol: p2pcommon.PingResponse.Uint32(), Length: 5}, Data: []byte{0, 1, 2, 3, 4}}
		pm := newDefaultMsgSigner(sampleKey1Priv, sampleKey1Pub, sampleKey1ID)
		if err := pm.SignMsg(sampleMsg1); (err != nil) != false {
			t.Errorf("defaultMsgSigner.signMsg() error = %v, wantErr %v", err, false)
		}
		assert.NotNil(t, sampleMsg1.Header.Sign)
		assert.True(t, len(sampleMsg1.Header.Sign) > 0)

		// different memory but same value is same signature
		if err := pm.SignMsg(sampleMsg2); (err != nil) != false {
			t.Errorf("defaultMsgSigner.signMsg() error = %v, wantErr %v", err, false)
		}
		assert.Equal(t, sampleMsg1.Header.Sign, sampleMsg2.Header.Sign)
	})

	t.Run("TDiffKey", func(t *testing.T) {
		// msg and msg2 is same at first
		sampleMsg1 := &types.P2PMessage{Header: &types.MsgHeader{Subprotocol: p2pcommon.PingResponse.Uint32(), Length: 5}, Data: []byte{0, 1, 2, 3, 4}}
		sampleMsg2 := &types.P2PMessage{Header: &types.MsgHeader{Subprotocol: p2pcommon.PingResponse.Uint32(), Length: 5}, Data: []byte{0, 1, 2, 3, 4}}
		pm := newDefaultMsgSigner(sampleKey1Priv, sampleKey1Pub, sampleKey1ID)
		if err := pm.SignMsg(sampleMsg1); (err != nil) != false {
			t.Errorf("defaultMsgSigner.signMsg() error = %v, wantErr %v", err, false)
		}
		assert.NotNil(t, sampleMsg1.Header.Sign)
		assert.True(t, len(sampleMsg1.Header.Sign) > 0)

		// same value but different pk is different signature
		pm2 := newDefaultMsgSigner(sampleKey2Priv, sampleKey2Pub, sampleKey2ID)
		if err := pm2.SignMsg(sampleMsg2); (err != nil) != false {
			t.Errorf("defaultMsgSigner.signMsg() error = %v, wantErr %v", err, false)
		}
		assert.NotEqual(t, sampleMsg1.Header.Sign, sampleMsg2.Header.Sign)
	})

}

func Test_defaultMsgSigner_verifyMsg(t *testing.T) {
	pubkey1bytes, _ := crypto.MarshalPublicKey(sampleKey1Pub)
	pubkey2bytes, _ := crypto.MarshalPublicKey(sampleKey2Pub)
	t.Run("TSucc", func(t *testing.T) {
		// msg and msg2 is same at first
		sampleMsg1 := &types.P2PMessage{Header: &types.MsgHeader{Subprotocol: p2pcommon.PingResponse.Uint32(), Length: 5, NodePubKey: pubkey1bytes}, Data: []byte{0, 1, 2, 3, 4}}

		pm := newDefaultMsgSigner(sampleKey1Priv, sampleKey1Pub, sampleKey1ID)
		assert.Nil(t, pm.SignMsg(sampleMsg1))
		expectedSig := append(make([]byte, 0), sampleMsg1.GetHeader().GetSign()...)
		assert.Equal(t, expectedSig, sampleMsg1.GetHeader().GetSign())

		pm2 := newDefaultMsgSigner(sampleKey2Priv, sampleKey2Pub, sampleKey2ID)
		// different memory but same value is same signature
		if err := pm.VerifyMsg(sampleMsg1, sampleKey1ID); (err != nil) != false {
			t.Errorf("defaultMsgSigner.verifyMsg() error = %v, wantErr %v", err, false)
		}
		if err := pm2.VerifyMsg(sampleMsg1, sampleKey1ID); (err != nil) != false {
			t.Errorf("defaultMsgSigner.verifyMsg() error = %v, wantErr %v", err, false)
		}
	})

	t.Run("TDiffKey", func(t *testing.T) {
		// msg and msg2 is same at first
		sampleMsg1 := &types.P2PMessage{Header: &types.MsgHeader{Subprotocol: p2pcommon.PingResponse.Uint32(), Length: 5}, Data: []byte{0, 1, 2, 3, 4}}
		pm := newDefaultMsgSigner(sampleKey1Priv, sampleKey1Pub, sampleKey1ID)
		assert.Nil(t, pm.SignMsg(sampleMsg1))
		expectedSig := append(make([]byte, 0), sampleMsg1.GetHeader().GetSign()...)
		assert.Equal(t, expectedSig, sampleMsg1.GetHeader().GetSign())

		sampleMsg1.Header.NodePubKey = pubkey2bytes
		pm2 := newDefaultMsgSigner(sampleKey2Priv, sampleKey2Pub, sampleKey2ID)
		// different memory but same value is same signature
		if err := pm.VerifyMsg(sampleMsg1, sampleKey2ID); (err != nil) != true {
			t.Errorf("defaultMsgSigner.verifyMsg() error = %v, wantErr %v", err, false)
		}
		if err := pm2.VerifyMsg(sampleMsg1, sampleKey2ID); (err != nil) != true {
			t.Errorf("defaultMsgSigner.verifyMsg() error = %v, wantErr %v", err, false)
		}
	})

	t.Run("TDiffFields", func(t *testing.T) {
		// msg and msg2 is same at first
		sampleMsg1 := &types.P2PMessage{Header: &types.MsgHeader{Subprotocol: p2pcommon.PingResponse.Uint32(), Length: 5}, Data: []byte{0, 1, 2, 3, 4}}

		pm := newDefaultMsgSigner(sampleKey1Priv, sampleKey1Pub, sampleKey1ID)
		assert.Nil(t, pm.SignMsg(sampleMsg1))
		expectedSig := append(make([]byte, 0), sampleMsg1.GetHeader().GetSign()...)
		assert.Equal(t, expectedSig, sampleMsg1.GetHeader().GetSign())
		sampleMsg1.Data = append(sampleMsg1.Data, 5, 6)
		pm2 := newDefaultMsgSigner(sampleKey2Priv, sampleKey2Pub, sampleKey2ID)
		// different memory but same value is same signature
		if err := pm2.VerifyMsg(sampleMsg1, sampleKey1ID); (err != nil) != true {
			t.Errorf("defaultMsgSigner.verifyMsg() error = %v, wantErr %v", err, false)
		}
	})

}
