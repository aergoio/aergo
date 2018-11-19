/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package chain

import (
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/types"
)

type ChainAnchor []([]byte)

const (
	//MaxAnchors = 32
	MaxAnchors = 1000000 //FIXME modify after implementing finder.fullscan
	Skip       = 16
)

// returns anchor blocks of chain
// use config
func (cs *ChainService) getAnchorsNew() (ChainAnchor, error) {
	//from top : 8 * 32 = 256
	anchors := make(ChainAnchor, 0)
	cnt := MaxAnchors
	logger.Debug().Msg("get anchors")

	blkNo := cs.getBestBlockNo()
LOOP:
	for i := 0; i < cnt; i++ {
		blockHash, err := cs.getHashByNo(blkNo)
		if err != nil {
			logger.Info().Msg("assertion - hash get failed")
			// assertion!
			return nil, err
		}

		anchors = append(anchors, blockHash)

		logger.Debug().Uint64("no", blkNo).Msg("anchor added")

		switch {
		case blkNo == 0:
			break LOOP
		case blkNo < Skip:
			blkNo = 0
		default:
			blkNo -= Skip
		}
	}

	return anchors, nil
}

// returns anchor blocks of chain
// use config
func (cs *ChainService) getAnchorsFromHash(blockHash []byte) ChainAnchor {
	/* TODO: use config */
	anchors := make(ChainAnchor, 0, 1000)
	anchors = append(anchors, blockHash)

	// collect 10 latest hashes
	latestNo := cs.getBestBlockNo()
	for i := 0; i < 10; i++ {
		blockHash, err := cs.getHashByNo(latestNo)
		if err != nil {
			logger.Info().Msg("assertion - hash get failed")
			// assertion!
			return nil
		}

		logger.Debug().Uint64("no", latestNo).Str("hash", enc.ToString(blockHash)).Msg("anchor")

		anchors = append(anchors, blockHash)
		if latestNo == 0 {
			return anchors
		}
		latestNo--
	}

	count := MaxAnchorCount
	// collect exponential
	var dec types.BlockNo = 1
	for i := 0; i < count; i++ {
		blockHash, err := cs.getHashByNo(latestNo)
		if err != nil {
			// assertion!
			return nil
		}

		logger.Debug().Uint64("no", latestNo).Str("hash", enc.ToString(blockHash)).Msg("anchor")

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
