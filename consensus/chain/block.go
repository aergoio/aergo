package chain

import (
	"errors"
	"fmt"
	"time"

	"github.com/aergoio/aergo/blockchain"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

var (
	// ErrQuit indicates that shutdown is initiated.
	ErrQuit           = errors.New("shutdown initiated")
	errBlockSizeLimit = errors.New("the transactions included exceeded the block size limit")
)

// ErrTimeout can be used to indicatefor any kind of timeout.
type ErrTimeout struct {
	Kind    string
	Timeout int64
}

func (e ErrTimeout) Error() string {
	if e.Timeout != 0 {
		return fmt.Sprintf("%s timeout (%v)", e.Kind, e.Timeout)
	}
	return e.Kind + " timeout"
}

// ErrBlockConnect indicates a error indicating a failed block connected
// request.
type ErrBlockConnect struct {
	id     string
	prevID string
	ec     error
}

func (e ErrBlockConnect) Error() string {
	return fmt.Sprintf("failed to connect block (%s): id=%s, prev id=%s", e.ec.Error(), e.id, e.prevID)
}

// GetBestBlock returns the current best block from chainservice
func GetBestBlock(hs component.ICompSyncRequester) *types.Block {
	result, err := hs.RequestFuture(message.ChainSvc, &message.GetBestBlock{}, time.Second,
		"consensus/util/info.GetBestBlock").Result()
	if err != nil {
		logger.Error().Err(err).Msg("failed to get best block info")
		return nil
	}
	return result.(message.GetBestBlockRsp).Block
}

// MaxBlockBodySize returns the maximum block body size.
//
// TODO: This is not an exact size. Let's make it exact!
func MaxBlockBodySize() uint32 {
	return blockchain.MaxBlockSize - uint32(proto.Size(&types.BlockHeader{}))
}

// GenerateBlock generate & return a new block
func GenerateBlock(hs component.ICompSyncRequester, prevBlock *types.Block, txOp TxOp, ts int64) (*types.Block, *types.BlockState, error) {
	txs, blockState, err := GatherTXs(hs, txOp, MaxBlockBodySize())
	if err != nil {
		return nil, nil, err
	}

	block := types.NewBlock(prevBlock, txs, ts)

	if len(txs) != 0 {
		logger.Debug().
			Str("txroothash", types.EncodeB64(block.GetHeader().GetTxsRootHash())).
			Int("hashed", len(txs)).
			Msg("BF: tx root hash")
	}

	return block, blockState, nil
}

// ConnectBlock send an AddBlock request to the chain service.
func ConnectBlock(hs component.ICompSyncRequester, block *types.Block, blockState *types.BlockState) error {
	// blockState does not include a valid BlockHash since it is constructed
	// from an incomplete block. So set it here.
	blockState.SetBlockHash(block.BlockID())

	_, err := hs.RequestFuture(message.ChainSvc, &message.AddBlock{PeerID: "", Block: block, Bstate: blockState},
		time.Second, "consensus/chain/info.ConnectBlock").Result()
	if err != nil {
		logger.Error().Uint64("no", block.Header.BlockNo).
			Str("hash", block.ID()).
			Str("prev", block.PrevID()).
			Msg("failed to connect block")

		return &ErrBlockConnect{id: block.ID(), prevID: block.PrevID(), ec: err}
	}

	return nil
}
