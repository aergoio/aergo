/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package types

import (
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

// PeerID is identier of peer. it is alias of libp2p.PeerID for now
type PeerID = core.PeerID

func IDB58Decode(b58 string) (PeerID, error) {
	return peer.IDB58Decode(b58)
}
func IDB58Encode(pid PeerID) string {
	return peer.IDB58Encode(pid)
}

func IDFromBytes(b []byte) (PeerID, error) {
	return peer.IDFromBytes(b)
}
func IDFromPublicKey(pubKey crypto.PubKey) (PeerID,error) {
	return peer.IDFromPublicKey(pubKey)
}
func IDFromPrivateKey(sk crypto.PrivKey) (PeerID, error) {
	return peer.IDFromPrivateKey(sk)
}