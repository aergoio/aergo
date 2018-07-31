package param

import "time"

const (
	// BlockProducers is the number of block producers
	BlockProducers = 23

	// DefaultBlockIntervalSec  is the default block generation interval in seconds.
	DefaultBlockIntervalSec = int64(1) // block production interval in sec
)

var (
	// BlockIntervalSec is the block genration interval in seconds.
	BlockIntervalSec int64
	// BlockIntervalMs is the block genration interval in milli-seconds.
	BlockIntervalMs int64
	// BpMinTimeLimitMs is the minimum block generation time limit in milli-sconds.
	BpMinTimeLimitMs int64
	// BpMaxTimeLimitMs is the maximum block generation time limit in milli-seconds.
	BpMaxTimeLimitMs int64
	// BlockInterval is the maximum block generation time limit.
	BlockInterval time.Duration
)

func init() {
	Init(DefaultBlockIntervalSec)
}

// Init initilizes the DPoS parameters.
func Init(blockIntervalSec int64) {
	BlockIntervalSec = blockIntervalSec
	BlockIntervalMs = BlockIntervalSec * 1000
	BpMinTimeLimitMs = BlockIntervalMs / 4
	BpMaxTimeLimitMs = BlockIntervalMs / 2
	BlockInterval = time.Duration(BlockIntervalSec) * time.Second
}
