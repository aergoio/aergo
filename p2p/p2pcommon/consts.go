/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2pcommon

import (
	"fmt"
	core "github.com/libp2p/go-libp2p-core"
	"time"
)

// constants of p2p protocol since v0.3
const (
	// this magic number is useful only in handshaking
	MAGICMain uint32 = 0x47416841
	MAGICTest uint32 = 0x2e415429

	SigLength = 16

	MaxPayloadLength = 1 << 23 // 8MB

	MaxBlockHeaderResponseCount = 10000
	MaxBlockResponseCount       = 2000
)

// P2PVersion is verion of p2p wire protocol. This version affects p2p handshake, data format transferred, etc
type P2PVersion uint32

func (v P2PVersion) Uint32() uint32 {
	return uint32(v)
}

func (v P2PVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v&0x7fff0000, v&0x0000ff00, v&0x000000ff)
}

const (
	P2PVersionUnknown P2PVersion = 0x00000000
	P2PVersion030     P2PVersion = 0x00000300
	P2PVersion031     P2PVersion = 0x00000301 // pseudo version for supporting multiversion
)

// context of multiaddr, as higher type of p2p message
const (
	LegacyP2PSubAddr core.ProtocolID = "/aergop2p/0.3"
	P2PSubAddr       core.ProtocolID = "/aergop2p"
)

// constatns for hanshake. for cacluating byte offset of wire handshake
const (
	V030HSHeaderLength = 8
	HSMagicLength      = 4
	HSVersionLength    = 4
	HSVerCntLength     = 4
)
const HSMaxVersionCnt = 16

const HSError uint32 = 0

// Codes in wire handshake
const (
	_ uint32 = iota
	ErrWrongHSReq
	ErrNoMatchedVersion  //

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

const (
	// DesignatedNodeTTL is time to determine which the remote designated peer is not working.
	DesignatedNodeTTL = time.Minute * 60

	// DefaultNodeTTL is time to determine which the remote peer is not working.
	DefaultNodeTTL = time.Minute * 10
)
