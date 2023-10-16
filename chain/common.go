/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"errors"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/types"
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

	Genesis *types.Genesis
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

// IsPublic reports whether the blockchain is public or not.
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
	if genesis.TotalBalance() != nil {
		types.MaxAER = genesis.TotalBalance()
		logger.Info().Str("TotalBalance", types.MaxAER.String()).Msg("set total from genesis")
	}

	Genesis = genesis
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
	if size > types.BlockSizeHardLimit() {
		logger.Panic().Uint32("block size", size).Msg("too large block size, hard limit = 8MiB")
	}
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
