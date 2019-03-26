/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"errors"

	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/contract/system"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/types"
)

const pubNetMaxBlockBodySize = 4000000

var (
	CoinbaseAccount []byte
	MaxAnchorCount  int
	VerifierCount   int

	// maxBlockBodySize is the upper limit of block size.
	maxBlockBodySize uint32
	maxBlockSize     uint32
	pubNet           bool
	consensusName    string
)

var (
	// ErrInvalidCoinbaseAccount is returned by Init when the coinbase account
	// address is invalid.
	ErrInvalidCoinbaseAccount = errors.New("invalid coinbase account in config")
	ErrInvalidConsensus       = errors.New("invalid consensus name from genesis")
)

// Init initializes the blockchain-related parameters.
func Init(maxBlkBodySize uint32, coinbaseAccountStr string, isBp bool, maxAnchorCount int, verifierCount int) error {
	var err error

	setBlockSizeLimit(maxBlkBodySize)

	if isBp {
		if len(coinbaseAccountStr) != 0 {
			CoinbaseAccount, err = types.DecodeAddress(coinbaseAccountStr)
			if err != nil {
				return ErrInvalidCoinbaseAccount
			}
			logger.Info().Str("account", enc.ToString(CoinbaseAccount)).Str("str", coinbaseAccountStr).
				Msg("set coinbase account")

		} else {
			logger.Info().Msg("Coinbase Account is nil, so BP reward will be discarded")
		}
	}

	MaxAnchorCount = maxAnchorCount
	VerifierCount = verifierCount

	return nil
}

// IsPublic reports whether the block chain is public or not.
func IsPublic() bool {
	return pubNet
}

func initChainParams(genesis *types.Genesis) {
	pubNet = genesis.ID.PublicNet
	if pubNet {
		setBlockSizeLimit(pubNetMaxBlockBodySize)
	}
	if err := setConsensusName(genesis.ConsensusType()); err != nil {
		logger.Panic().Err(err).Msg("invalid consensus type in genesis block")
	}
	system.InitDefaultBpCount(len(genesis.BPs))
}

// MaxBlockBodySize returns the max block body size.
func MaxBlockBodySize() uint32 {
	return maxBlockBodySize
}

// MaxBlockSize returns the max block size.
func MaxBlockSize() uint32 {
	return maxBlockSize
}

func setMaxBlockBodySize(size uint32) {
	maxBlockBodySize = size
}

func setBlockSizeLimit(maxBlockBodySize uint32) {
	setMaxBlockBodySize(maxBlockBodySize)
	maxBlockSize = MaxBlockBodySize() + types.DefaultMaxHdrSize
}

func setConsensusName(val string) error {
	for _, name := range consensus.ConsensusName {
		if val == name {
			consensusName = val
		}
	}

	if consensusName == "" {
		return ErrInvalidConsensus
	}

	return nil
}

func ConsensusName() string {
	return consensusName
}
