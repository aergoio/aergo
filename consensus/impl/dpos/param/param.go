package param

import "time"

const (
	// BlockProducers is the number of block producers
	BlockProducers = 23
	BpMajority     = BlockProducers*2/3 + 1

	DefaultBlockIntervalSec = int64(1) // block production interval in sec
)

var (
	BlockIntervalSec int64
	BlockIntervalMs  int64
	BpMinTimeLimitMs int64
	BpMaxTimeLimitMs int64
	BlockInterval    time.Duration
	LoopInterval     time.Duration
)

func init() {
	Init(DefaultBlockIntervalSec)
}

// Init initilizes the DPoS paramters.
func Init(blockIntervalSec int64) {
	BlockIntervalSec = blockIntervalSec
	BlockIntervalMs = BlockIntervalSec * 1000
	BpMinTimeLimitMs = BlockIntervalMs / 4
	BpMaxTimeLimitMs = BlockIntervalMs / 2
	BlockInterval = time.Duration(BlockIntervalSec) * time.Second
	LoopInterval = BlockInterval / 10
}
