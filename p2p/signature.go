/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"fmt"

	"github.com/aergoio/aergo/v2/internal/enc/proto"
	"github.com/aergoio/aergo/v2/p2p/p2pcommon"
	"github.com/aergoio/aergo/v2/types"
	"github.com/libp2p/go-libp2p/core/crypto"
)

type defaultMsgSigner struct {
	selfPeerID types.PeerID
	privateKey crypto.PrivKey
	pubKey     crypto.PubKey

	pidBytes    []byte
	pubKeyBytes []byte
}

func newDefaultMsgSigner(privKey crypto.PrivKey, pubKey crypto.PubKey, peerID types.PeerID) p2pcommon.MsgSigner {
	pidBytes := []byte(peerID)
	pubKeyBytes, _ := crypto.MarshalPublicKey(pubKey)
	return &defaultMsgSigner{selfPeerID: peerID, privateKey: privKey, pubKey: pubKey, pidBytes: pidBytes, pubKeyBytes: pubKeyBytes}
}

// SignMsg sign an outgoing p2p message payload and assign the signature to field of message
func (pm *defaultMsgSigner) SignMsg(message *types.P2PMessage) error {
	message.Header.PeerID = pm.pidBytes
	message.Header.NodePubKey = pm.pubKeyBytes
	data, err := proto.Encode(&types.P2PMessage{Header: canonicalizeHeader(message.Header), Data: message.Data})
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

func (pm *defaultMsgSigner) VerifyMsg(msg *types.P2PMessage, senderID types.PeerID) error {
	// check signature
	pubKey, err := crypto.UnmarshalPublicKey(msg.Header.NodePubKey)
	if err != nil {
		return err
	}
	signature := msg.Header.Sign
	checkOrigin := false
	if checkOrigin {
		// TODO it can be needed, and if that, modify code to get peer id from caller and enable this code
		if err := checkPidWithPubkey(senderID, pubKey); err != nil {
			return err
		}
	}

	data, _ := proto.Encode(&types.P2PMessage{Header: canonicalizeHeader(msg.Header), Data: msg.Data})
	return verifyBytes(data, signature, pubKey)
}

func checkPidWithPubkey(peerID types.PeerID, pubKey crypto.PubKey) error {
	// extract node peerID from the provided public key
	idFromKey, err := types.IDFromPublicKey(pubKey)
	if err != nil {
		return err
	}

	// verify that message author node PeerID matches the provided node public key
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

func (d *dummySigner) SignMsg(msg *types.P2PMessage) error {
	msg.Header.Sign = dummyBytes
	return nil
}

func (d *dummySigner) VerifyMsg(msg *types.P2PMessage, senderID types.PeerID) error {
	return nil
}

var _ p2pcommon.MsgSigner = (*dummySigner)(nil)
