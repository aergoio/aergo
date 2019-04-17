package chain

import (
	"errors"
	"fmt"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/p2p/p2putil"
	"github.com/libp2p/go-libp2p-peer"
	"time"

	"github.com/aergoio/aergo/chain"
	"github.com/aergoio/aergo/message"
	"github.com/aergoio/aergo/pkg/component"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
)

var (
	// ErrQuit indicates that shutdown is initiated.
	ErrQuit           = errors.New("shutdown initiated")
	errBlockSizeLimit = errors.New("the transactions included exceeded the block size limit")
	ErrBlockEmpty     = errors.New("no transactions in block")
	ErrSyncChain      = errors.New("failed to sync request")
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
func MaxBlockBodySize() uint32 {
	return chain.MaxBlockBodySize()
}

// GenerateBlock generate & return a new block
func GenerateBlock(hs component.ICompSyncRequester, prevBlock *types.Block, bState *state.BlockState, txOp TxOp, ts int64, skipEmpty bool) (*types.Block, error) {
	transactions, err := GatherTXs(hs, bState, txOp, MaxBlockBodySize())
	if err != nil {
		return nil, err
	}

	txs := make([]*types.Tx, 0)
	for _, x := range transactions {
		txs = append(txs, x.GetTx())
	}

	if len(txs) == 0 && skipEmpty {
		logger.Debug().Msg("BF: empty block is skipped")
		return nil, ErrBlockEmpty
	}

	block := types.NewBlock(prevBlock, bState.GetRoot(), bState.Receipts(), txs, chain.CoinbaseAccount, ts)
	if len(txs) != 0 && logger.IsDebugEnabled() {
		logger.Debug().
			Str("txroothash", types.EncodeB64(block.GetHeader().GetTxsRootHash())).
			Int("hashed", len(txs)).
			Msg("BF: tx root hash")
	}

	return block, nil
}

// ConnectBlock send an AddBlock request to the chain service.
func ConnectBlock(hs component.ICompSyncRequester, block *types.Block, blockState *state.BlockState) error {
	// blockState does not include a valid BlockHash since it is constructed
	// from an incomplete block. So set it here.
	_, err := hs.RequestFuture(message.ChainSvc, &message.AddBlock{PeerID: "", Block: block, Bstate: blockState},
		time.Second, "consensus/chain/info.ConnectBlock").Result()
	if err != nil {
		logger.Error().Err(err).Uint64("no", block.Header.BlockNo).
			Str("hash", block.ID()).
			Str("prev", block.PrevID()).
			Msg("failed to connect block")

		return &ErrBlockConnect{id: block.ID(), prevID: block.PrevID(), ec: err}
	}

	return nil
}

func SyncChain(hs *component.ComponentHub, targetHash []byte, targetNo types.BlockNo, peerID peer.ID) error {
	logger.Info().Str("peer", p2putil.ShortForm(peerID)).Uint64("no", targetNo).
		Str("hash", enc.ToString(targetHash)).Msg("request to sync for consensus")

	notiC := make(chan error)
	hs.Tell(message.SyncerSvc, &message.SyncStart{PeerID: peerID, TargetNo: targetNo, NotifyC: notiC})

	// wait end of sync every 1sec
	select {
	case err := <-notiC:
		if err != nil {
			logger.Error().Err(err).Uint64("no", targetNo).
				Str("hash", enc.ToString(targetHash)).
				Msg("failed to sync")

			return ErrSyncChain
		}
	}

	logger.Info().Str("peer", p2putil.ShortForm(peerID)).Msg("succeeded to sync for consensus")
	// TODO check best block is equal to target Hash/no
	return nil
}
