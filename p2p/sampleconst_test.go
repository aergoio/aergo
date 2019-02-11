/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"encoding/base64"
	"encoding/hex"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/types"
	"github.com/aergoio/aergo/p2p/p2pcommon"

	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
)

// this file collect sample global constants used in unit test. I'm tired of creating less meaningfule variables in each tests.

var dummyBlockHash, _ = hex.DecodeString("4f461d85e869ade8a0544f8313987c33a9c06534e50c4ad941498299579bd7ac")
var dummyBlockHeight uint64 = 100215
var dummyTxHash, _ = enc.ToBytes("4H4zAkAyRV253K5SNBJtBxqUgHEbZcXbWFFc6cmQHY45")

const (
	sampleKey1PrivBase64 = "CAISIM1yE7XjJyKTw4fQYMROnlxmEBut5OPPGVde7PeVAf0x"
	sampelKey1PubBase64  = "CAISIQOMA3AHgprpAb7goiDGLI6b/je3JKiYSBHyb46twYV7RA=="
	sampleKey1IDbase58   = "16Uiu2HAmN5YU8V2LnTy9neuuJCLNsxLnd5xVSRZqkjvZUHS3mLoD"

	sampleKey2PrivBase64 = "CAISIPK7nwYwZwl0LVvjJjZ58gN/4z0iAIOEOi5fKDIMBKCN"
)

var sampleKey1Priv crypto.PrivKey
var sampleKey1Pub crypto.PubKey
var sampleKey1ID peer.ID

var sampleKey2Priv crypto.PrivKey
var sampleKey2Pub crypto.PubKey
var sampleKey2ID peer.ID

func init() {
	bytes, _ := base64.StdEncoding.DecodeString(sampleKey1PrivBase64)
	sampleKey1Priv, _ = crypto.UnmarshalPrivateKey(bytes)
	bytes, _ = base64.StdEncoding.DecodeString(sampelKey1PubBase64)
	sampleKey1Pub, _ = crypto.UnmarshalPublicKey(bytes)
	if !sampleKey1Priv.GetPublic().Equals(sampleKey1Pub) {
		panic("problem in pk generation ")
	}
	sampleKey1ID, _ = peer.IDFromPublicKey(sampleKey1Pub)
	if sampleKey1ID.Pretty() != sampleKey1IDbase58 {
		panic("problem in id generation")
	}

	bytes, _ = base64.StdEncoding.DecodeString(sampleKey2PrivBase64)
	sampleKey2Priv, _ = crypto.UnmarshalPrivateKey(bytes)
	sampleKey2Pub = sampleKey2Priv.GetPublic()
	sampleKey2ID, _ = peer.IDFromPublicKey(sampleKey2Pub)
}

var dummyPeerID peer.ID
var dummyPeerID2 peer.ID
var dummyPeerID3 peer.ID

func init() {
	dummyPeerID = sampleKey1ID
	dummyPeerID2, _ = peer.IDB58Decode("16Uiu2HAmFqptXPfcdaCdwipB2fhHATgKGVFVPehDAPZsDKSU7jRm")
	dummyPeerID3, _ = peer.IDB58Decode("16Uiu2HAmU8Wc925gZ5QokM4sGDKjysdPwRCQFoYobvoVnyutccCD")
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

var dummyMo *MockMsgOrder

func init() {
	dummyMo = new(MockMsgOrder)
	dummyMo.On("IsNeedSign").Return(true)
	dummyMo.On("IsRequest").Return(true)
	dummyMo.On("GetProtocolID").Return(NewTxNotice)
	dummyMo.On("GetMsgID").Return(p2pcommon.NewMsgID())
}
