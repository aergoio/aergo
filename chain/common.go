/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"errors"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/types"
)

var (
	// MaxBlockSize is the maximum size of a block.
	MaxBlockSize    uint32
	CoinbaseAccount []byte
	CoinbaseFee     uint64
	MaxAnchorCount  int
	UseFastSyncer   bool
)

var (
	ErrInvalidCoinbaseAccount = errors.New("invalid coinbase account in config")
)

// Init initializes the blockchain-related parameters.
func Init(maxBlockSize uint32, coinbaseAccountStr string, coinbaseFee uint64, isBp bool, maxAnchorCount int, useFastSyncer bool) error {
	var err error

	MaxBlockSize = maxBlockSize
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

	CoinbaseFee = coinbaseFee
	MaxAnchorCount = maxAnchorCount
	UseFastSyncer = useFastSyncer
	return nil
}
