/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"

	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
)

// signHandler sign or verify p2p message
type msgSigner interface {
	// signMsg calulate signature and fill related fields in msg(peerid, pubkey, signature or etc)
	signMsg(msg *types.P2PMessage) error
	// verifyMsg check signature is valid
	vefifyMsg(msg *types.P2PMessage, pubKey crypto.PubKey) error
}

type defaultMsgSigner struct {
	selfPeerID peer.ID
	privateKey crypto.PrivKey
	pubKey     crypto.PubKey

	pidBytes    []byte
	pubKeyBytes []byte
}

func newDefaultMsgSigner(privKey crypto.PrivKey, pubKey crypto.PubKey, peerID peer.ID) msgSigner {
	pidBytes := []byte(peerID)
	pubKeyBytes, _ := pubKey.Bytes()
	return &defaultMsgSigner{selfPeerID: peerID, privateKey: privKey, pubKey: pubKey, pidBytes: pidBytes, pubKeyBytes: pubKeyBytes}
}

// sign an outgoing p2p message payload
func (pm *defaultMsgSigner) signMsg(message *types.P2PMessage) error {
	// TODO this code modify caller's parameter.
	message.Header.PeerID = pm.pidBytes
	message.Header.NodePubKey = pm.pubKeyBytes
	data, err := proto.Marshal(&types.P2PMessage{Header: canonicalizeHeader(message.Header), Data: message.Data})
	if err != nil {
		return err
	}
	signature, err := pm.signBytes(data)
	if err != nil {
		return err
	}
	message.Header.Sign = signature
	return nil
}

func canonicalizeHeader(src *types.MsgHeader) *types.MsgHeader {
	// copy fields excluding generated fields and signature itself
	return &types.MsgHeader{
		ClientVersion: src.ClientVersion,
		Gossip:        src.Gossip,
		Id:            src.Id,
		Length:        src.Length,
		NodePubKey:    src.NodePubKey,
		PeerID:        src.PeerID,
		Sign:          nil,
		Subprotocol:   src.Subprotocol,
		Timestamp:     src.Timestamp,
	}
}

// sign binary data using the local node's private key
func (pm *defaultMsgSigner) signBytes(data []byte) ([]byte, error) {
	key := pm.privateKey
	res, err := key.Sign(data)
	return res, err
}

func (pm *defaultMsgSigner) vefifyMsg(msg *types.P2PMessage, pubKey crypto.PubKey) error {
	signature := msg.Header.Sign
	checkOrigin := false
	if checkOrigin {
		// TODO it can be needed, and if that modify code to get peerid from caller and enable this code
		if err := checkPidWithPubkey(peer.ID("dummy"), pubKey); err != nil {
			return err
		}
	}

	data, _ := proto.Marshal(&types.P2PMessage{Header: canonicalizeHeader(msg.Header), Data: msg.Data})
	return verifyBytes(data, signature, pubKey)
}

func checkPidWithPubkey(peerID peer.ID, pubKey crypto.PubKey) error {
	// extract node peer.ID from the provided public key
	idFromKey, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return err
	}

	// verify that message author node peer.ID matches the provided node public key
	if idFromKey != peerID {
		return fmt.Errorf("PeerID mismatch")
	}
	return nil
}

// VerifyData Verifies incoming p2p message data integrity
// data: data to verify
// signature: author signature provided in the message payload
// pubkey: author public key from the message payload
func verifyBytes(data []byte, signature []byte, pubkey crypto.PubKey) error {
	res, err := pubkey.Verify(data, signature)
	if err != nil {
		return err
	}
	if !res {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

var dummyBytes = []byte{}

type dummySigner struct{}

func (d *dummySigner) signMsg(msg *types.P2PMessage) error {
	msg.Header.Sign = dummyBytes
	return nil
}

func (d *dummySigner) vefifyMsg(msg *types.P2PMessage, pubKey crypto.PubKey) error {
	return nil
}

var _ msgSigner = (*dummySigner)(nil)
