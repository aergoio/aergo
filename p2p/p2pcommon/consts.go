/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	protocol "github.com/libp2p/go-libp2p-protocol"
	"time"
)

// constants of p2p protocol since v0.3
const (
	// this magic number is useful only in handshaking
	MAGICMain uint32 = 0x47416841
	MAGICTest uint32 = 0x2e415429

	P2PVersion030 uint32 = 0x00000300

	SigLength = 16

	MaxPayloadLength = 1 << 23 // 8MB

	MaxBlockHeaderResponseCount = 10000
	MaxBlockResponseCount       = 2000
)

// context of multiaddr, as higher type of p2p message
const (
	AergoP2PSub protocol.ID = "/aergop2p/0.3"
)

// constants about private key
const (
	DefaultPkKeyPrefix = "aergo-peer"
	DefaultPkKeyExt    = ".key"
	DefaultPubKeyExt   = ".pub"
	DefaultPeerIDExt   = ".id"
)

// constants for inter-communication of aergosvr
const (
	// other actor
	DefaultActorMsgTTL = time.Second * 4
)

