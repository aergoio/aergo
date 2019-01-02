/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"errors"
	"math/big"

	"github.com/aergoio/aergo/contract"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/types"
)

const pubNetMaxBlockSize = 4000000

var (
	CoinbaseAccount []byte
	MaxAnchorCount  int
	UseFastSyncer   bool
	VerifierCount   int
	coinbaseFee     *big.Int

	// MaxBlockSize is the upper limit of block size.
	maxBlockSize uint32
	pubNet       bool
)

var (
	ErrInvalidCoinbaseAccount = errors.New("invalid coinbase account in config")
)

// Init initializes the blockchain-related parameters.
func Init(maxBlkSize uint32, coinbaseAccountStr string, isBp bool, maxAnchorCount int, useFastSyncer bool, verifierCount int) error {
	var err error

	maxBlockSize = maxBlkSize
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
	UseFastSyncer = useFastSyncer
	VerifierCount = verifierCount

	return nil
}

// IsPublic reports whether the block chain is public or not.
func IsPublic() bool {
	return pubNet
}

func initChainEnv(genesis *types.Genesis) {
	pubNet = genesis.ID.PublicNet
	if pubNet {
		setMaxBlockSize(pubNetMaxBlockSize)
	}
	contract.PubNet = pubNet
	fee, _ := genesis.ID.GetCoinbaseFee() // no failure
	setCoinbaseFee(fee)
}

// MaxBlockSize returns (kind of) the upper limit of block size.
func MaxBlockSize() uint32 {
	return maxBlockSize
}

func setMaxBlockSize(size uint32) {
	maxBlockSize = size
}

func setCoinbaseFee(fee *big.Int) {
	coinbaseFee = fee
}

func CoinbaseFee() *big.Int {
	return coinbaseFee
}
