/*
 * @file
 * @copyright defined in aergo/LICENSE.txt
 */

package p2p

import "time"

// constants for inter-communication of aergosvr
const (
	// other actor
	DefaultActorMsgTTL = time.Second * 4
)

// constants for peer communicating
const (
	// peer handshake will be failed when taken more than defaultHandshakeTTL
	defaultHandshakeTTL = time.Second * 20

	defaultPingInterval = time.Second * 60
	// txNoticeInterval is max wait time when not sufficient txs to notify is collected. i.e newTxNotice is sent to peer within this time.
	txNoticeInterval = time.Second * 1
	// writeMsgBufferSize is queue size of message to a peer. connection will be closed when queue is exceeded.
	writeMsgBufferSize = 20

	// time to determine which the remote designated peer is not working.
	DesignatedNodeTTL = time.Minute * 60
	// time to determine which the remote peer is not working.
	DefaultNodeTTL = time.Minute * 10
)

// constants for legacy sync algorithm. DEPRECATED newer sync loging in syncer package is used now.
const (
	SyncWorkTTL        = time.Second * 30
	AddBlockCheckpoint = 100
	AddBlockWaitTime   = time.Second * 10
)

// constants for node discovery
const (
	DiscoveryQueryInterval = time.Minute * 5

	MaxAddrListSizePolaris = 200
	MaxAddrListSizePeer    = 50
)

// constants for peer internal operations
const (
	cleanRequestInterval = time.Hour
)

// constants for caching
// TODO this value better related to max peer and block produce interval, not constant
const (
	DefaultGlobalBlockCacheSize = 300
	DefaultPeerBlockCacheSize   = 100

	DefaultGlobalTxCacheSize = 50000
	DefaultPeerTxCacheSize   = 10000
	// DefaultPeerTxQueueSize is maximum size of hashes in a single tx notice message
	DefaultPeerTxQueueSize = 2000
	// value to sent to cache, since block and tx cache need only hash itself (stored as key of map)
	cachePlaceHolder = true
)

// constants about private key
const (
	DefaultPkKeyPrefix    = "aergo-peer"
	DefaultPkKeyExt       = ".key"
	DefaultPubKeyExt      = ".pub"
	DefaultPeerIDExt      = ".id"
)
