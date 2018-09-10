/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

var (
	// MaxBlockSize is the maximum size of a block.
	MaxBlockSize uint32
)

// Init initializes the blockchain-related parameters.
func Init(maxBlockSize uint32) {
	MaxBlockSize = maxBlockSize
}
