/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package blockchain

import (
	"github.com/aergoio/aergo/types"
)

type hash []byte
type ChainAnchor []hash

// returns anchor blocks of chain
// use config
func (cs *ChainService) getAnchorsFromHash(blockHash hash) ChainAnchor {
	/* TODO: use config */
	anchors := make(ChainAnchor, 0, 1000)
	anchors = append(anchors, blockHash)

	// collect 10 latest hashes
	latestNo := cs.getBestBlockNo()
	for i := 0; i < 10; i++ {
		logger.Info().Uint64("Num", latestNo).Msg("Latest")
		blockHash, err := cs.getHashByNo(latestNo)
		if err != nil {
			logger.Info().Msg("assertion - hash get failed")
			// assertion!
			return nil
		}
		anchors = append(anchors, blockHash)
		if latestNo == 0 {
			return anchors
		}
		latestNo--
	}

	// collect exponential
	var dec types.BlockNo = 1
	for i := 0; i < 10; i++ {
		blockHash, err := cs.getHashByNo(latestNo)
		if err != nil {
			// assertion!
			return nil
		}
		anchors = append(anchors, blockHash)
		if latestNo <= dec {
			if latestNo == 0 {
				break
			}
			latestNo = 0
		} else {
			latestNo -= dec
			dec *= 2
		}
	}

	return anchors
}
