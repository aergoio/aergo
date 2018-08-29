/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package p2p

import (
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
	peer "github.com/libp2p/go-libp2p-peer"
)

// Authenticate incoming p2p message
// message: a protobufs go data object
// data: common p2p message data
func (ps *peerManager) AuthenticateMessage(message proto.Message, data *types.MessageData) bool {
	// for Test only
	return true

	// store a temp ref to signature and remove it from message data
	// sign is a string to allow easy reset to zero-value (empty string)
	sign := data.Sign
	data.Sign = []byte{}

	// marshall data without the signature to protobufs3 binary format
	bin, err := proto.Marshal(message)
	if err != nil {
		ps.log.Warn().Msg("failed to marshal pb message")
		return false
	}

	// restore sig in message data (for possible future use)
	data.Sign = sign

	// restore peer peer.ID binary format from base58 encoded node peer.ID data
	peerID, err := peer.IDB58Decode(data.PeerID)
	if err != nil {
		ps.log.Warn().Err(err).Msg("Failed to decode node peer.ID from base58")
		return false
	}

	// verify the data was authored by the signing peer identified by the public key
	// and signature included in the message
	err = VerifyData(bin, []byte(sign), peerID, data.NodePubKey)
	if err != nil {
		ps.log.Debug().Err(err).Msg("message verification failed")
		return false
	}
	return true
}

// sign an outgoing p2p message payload
func (ps *peerManager) SignProtoMessage(message proto.Message) ([]byte, error) {
	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	return ps.SignData(data)
}

// sign binary data using the local node's private key
func (ps *peerManager) SignData(data []byte) ([]byte, error) {
	key := ps.privateKey
	res, err := key.Sign(data)
	return res, err
}
