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

const pubNetMaxBlockBodySize = 4000000

var (
	CoinbaseAccount []byte
	MaxAnchorCount  int
	VerifierCount   int
	coinbaseFee     *big.Int

	// maxBlockBodySize is the upper limit of block size.
	maxBlockBodySize uint32
	maxBlockSize     uint32
	pubNet           bool
)

var (
	// ErrInvalidCoinbaseAccount is returned by Init when the coinbase account
	// address is invalid.
	ErrInvalidCoinbaseAccount = errors.New("invalid coinbase account in config")
)

// Init initializes the blockchain-related parameters.
func Init(maxBlkBodySize uint32, coinbaseAccountStr string, isBp bool, maxAnchorCount int, verifierCount int) error {
	var err error

	setMaxBlockBodySize(maxBlkBodySize)
	setMaxBlockSize(MaxBlockBodySize() + types.DefaultMaxHdrSize)

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

func initChainEnv(genesis *types.Genesis) {
	pubNet = genesis.ID.PublicNet
	if pubNet {
		setMaxBlockBodySize(pubNetMaxBlockBodySize)
	}
	contract.PubNet = pubNet
	fee, _ := genesis.ID.GetCoinbaseFee() // no failure
	setCoinbaseFee(fee)
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

func setMaxBlockSize(size uint32) {
	maxBlockSize = size
}

func setCoinbaseFee(fee *big.Int) {
	coinbaseFee = fee
}

func CoinbaseFee() *big.Int {
	return coinbaseFee
}
