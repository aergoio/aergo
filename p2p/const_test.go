/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p/p2pcommon"
	"github.com/aergoio/aergo/p2p/p2pmock"
	"github.com/aergoio/aergo/types"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p-core/crypto"
)

// this file collect sample global constants used in unit test. I'm tired of creating less meaningful variables in each tests.

var dummyBlockHash, _ = hex.DecodeString("4f461d85e869ade8a0544f8313987c33a9c06534e50c4ad941498299579bd7ac")
var dummyBlockHeight uint64 = 100215
var dummyTxHash, _ = enc.ToBytes("4H4zAkAyRV253K5SNBJtBxqUgHEbZcXbWFFc6cmQHY45")

var samplePeerID types.PeerID
var sampleMeta p2pcommon.PeerMeta
var sampleErr error

var logger *log.Logger

func init() {
	logger = log.NewLogger("test")
	samplePeerID, _ = types.IDB58Decode("16Uiu2HAkvvhjxVm2WE9yFBDdPQ9qx6pX9taF6TTwDNHs8VPi1EeR")
	sampleErr = fmt.Errorf("err in unittest")
	sampleMA, _ := types.ParseMultiaddr("/ip4/192.168.1.2/tcp/7846")
	sampleMeta = p2pcommon.PeerMeta{ID: samplePeerID, Addresses:[]types.Multiaddr{sampleMA}}

}

const (
	sampleKey1PrivBase64 = "CAISIM1yE7XjJyKTw4fQYMROnlxmEBut5OPPGVde7PeVAf0x"
	sampelKey1PubBase64  = "CAISIQOMA3AHgprpAb7goiDGLI6b/je3JKiYSBHyb46twYV7RA=="
	sampleKey1IDbase58   = "16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD"

	sampleKey2PrivBase64 = "CAISIPK7nwYwZwl0LVvjJjZ58gN/4z0iAIOEOi5fKDIMBKCN"
)

var sampleMsgID p2pcommon.MsgID
var sampleKey1Priv crypto.PrivKey
var sampleKey1Pub crypto.PubKey
var sampleKey1ID types.PeerID

var sampleKey2Priv crypto.PrivKey
var sampleKey2Pub crypto.PubKey
var sampleKey2ID types.PeerID

var dummyPeerID types.PeerID
var dummyPeerID2 types.PeerID
var dummyPeerID3 types.PeerID

var dummyBestBlock *types.Block
var dummyMeta p2pcommon.PeerMeta

func init() {
	bytes, _ := base64.StdEncoding.DecodeString(sampleKey1PrivBase64)
	sampleKey1Priv, _ = crypto.UnmarshalPrivateKey(bytes)
	bytes, _ = base64.StdEncoding.DecodeString(sampelKey1PubBase64)
	sampleKey1Pub, _ = crypto.UnmarshalPublicKey(bytes)
	if !sampleKey1Priv.GetPublic().Equals(sampleKey1Pub) {
		panic("problem in pk generation ")
	}
	sampleKey1ID, _ = types.IDFromPublicKey(sampleKey1Pub)
	if sampleKey1ID.Pretty() != sampleKey1IDbase58 {
		panic("problem in id generation")
	}

	bytes, _ = base64.StdEncoding.DecodeString(sampleKey2PrivBase64)
	sampleKey2Priv, _ = crypto.UnmarshalPrivateKey(bytes)
	sampleKey2Pub = sampleKey2Priv.GetPublic()
	sampleKey2ID, _ = types.IDFromPublicKey(sampleKey2Pub)

	sampleMsgID = p2pcommon.NewMsgID()

	dummyPeerID = sampleKey1ID
	dummyPeerID2, _ = types.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
	dummyPeerID3, _ = types.IDB58Decode("16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD")

	dummyBestBlock = &types.Block{Header: &types.BlockHeader{}}
	dummyMeta = p2pcommon.PeerMeta{ID: dummyPeerID}
}

const hashSize = 32

var sampleTxsB58 = []string{
	"4H4zAkAyRV253K5SNBJtBxqUgHEbZcXbWFFc6cmQHY45",
	"6xfk39kuyDST7NwCu8tx3wqwFZ5dwKPDjxUS14tU7NZb8",
	"E8dbBGe9Hnuhk35cJoekPjL3VoL4xAxtnRuP47UoxzHd",
	"HB7Hg5GUbHuxwe8Lp5PcYUoAaQ7EZjRNG6RuvS6DnDRf",
	"BxKmDg9VbWHxrWnStEeTzJ2Ze7RF7YK4rpyjcsWSsnxs",
	"DwmGqFU4WgADpYN36FXKsYxMjeppvh9Najg4KxJ8gtX3",
}

var sampleTxs [][]byte
var sampleTxHashes []types.TxID

var sampleBlksB58 = []string{
	"v6zbuQ4aVSdbTwQhaiZGp5pcL5uL55X3kt2wfxor5W6",
	"2VEPg4MqJUoaS3EhZ6WWSAUuFSuD4oSJ645kSQsGV7H9",
	"AtzTZ2CZS45F1276RpTdLfYu2DLgRcd9HL3aLqDT1qte",
	"2n9QWNDoUvML756X7xdHWCFLZrM4CQEtnVH2RzG5FYAw",
	"6cy7U7XKYtDTMnF3jNkcJvJN5Rn85771NSKjc5Tfo2DM",
	"3bmB8D37XZr4DNPs64NiGRa2Vw3i8VEgEy6Xc2XBmRXC",
}
var sampleBlks [][]byte
var sampleBlksHashes []types.BlockID

func init() {
	sampleTxs = make([][]byte, len(sampleTxsB58))
	sampleTxHashes = make([]types.TxID, len(sampleTxsB58))
	for i, hashb58 := range sampleTxsB58 {
		hash, _ := enc.ToBytes(hashb58)
		sampleTxs[i] = hash
		copy(sampleTxHashes[i][:], hash)
	}

	sampleBlks = make([][]byte, len(sampleBlksB58))
	sampleBlksHashes = make([]types.BlockID, len(sampleBlksB58))
	for i, hashb58 := range sampleTxsB58 {
		hash, _ := enc.ToBytes(hashb58)
		sampleBlks[i] = hash
		copy(sampleBlksHashes[i][:], hash)
	}
}

var dummyMo *p2pmock.MockMsgOrder

func createDummyMo(ctrl *gomock.Controller) *p2pmock.MockMsgOrder {
	dummyMo = p2pmock.NewMockMsgOrder(ctrl)
	dummyMo.EXPECT().IsNeedSign().Return(true).AnyTimes()
	dummyMo.EXPECT().IsRequest().Return(true).AnyTimes()
	dummyMo.EXPECT().GetProtocolID().Return(p2pcommon.NewTxNotice).AnyTimes()
	dummyMo.EXPECT().GetMsgID().Return(p2pcommon.NewMsgID()).AnyTimes()
	return dummyMo
}
