package p2p

import (
	"encoding/base64"
	"encoding/hex"

	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
)

// this file collect sample global constants used in unit test. I'm tired of creating less meaningfule variables in each tests.

var dummyBlockHash, _ = hex.DecodeString("4f461d85e869ade8a0544f8313987c33a9c06534e50c4ad941498299579bd7ac")
var dummyBlockHeight uint64 = 100215

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
