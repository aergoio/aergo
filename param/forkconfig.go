package param

import (
	"bytes"
	"github.com/aergoio/aergo/internal/enc"
)

var (
	mainnetGenesisHash, _ = enc.ToBytes("000000000000000")
	testnetGenesisHash, _ = enc.ToBytes("FnzS3cTd6ikjkvTDWoq3JHY7taAQqZLMfbaiR8w4QtA3")

	chainForkCfg *ForkConfig
)
var (
	mainnetForkConfig = &ForkConfig{AIP1: 0}
	testnetForkConfig = &ForkConfig{AIP1: 1000}
	testForkConfig    = &ForkConfig{AIP1: 0}
)

type ForkConfig struct {
	AIP1 uint64
	AIP2 uint64
}

func SetForkConfig(genesisHash []byte) error {
	if bytes.Equal(genesisHash, mainnetGenesisHash) {
		chainForkCfg = mainnetForkConfig
	} else if bytes.Equal(genesisHash, testnetGenesisHash) {
		chainForkCfg = testnetForkConfig
	}
	chainForkCfg = testForkConfig

	return nil
}

func GetForkConfig() *ForkConfig {
	return chainForkCfg
}

func (fc *ForkConfig) ISAIP1(BlkNo uint64) bool {
	if fc.AIP1 <= BlkNo {
		return true
	}
	return false
}

func (fc *ForkConfig) ISAIP2(BlkNo uint64) bool {
	if fc.AIP1 <= BlkNo {
		return true
	}
	return false
}
