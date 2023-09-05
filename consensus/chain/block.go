package chain

import (
	"errors"
	"fmt"
	"time"

	"github.com/aergoio/aergo/v2/chain"
	"github.com/aergoio/aergo/v2/internal/enc"
	"github.com/aergoio/aergo/v2/message"
	"github.com/aergoio/aergo/v2/p2p/p2putil"
	"github.com/aergoio/aergo/v2/pkg/component"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
)

var (
	// ErrQuit indicates that shutdown is initiated.
	ErrQuit           = errors.New("shutdown initiated")
	ErrBlockEmpty     = errors.New("no transactions in block")
	ErrSyncChain      = errors.New("failed to sync request")
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
func MaxBlockBodySize() uint32 {
	return chain.MaxBlockBodySize()
}

type FetchFn = func(component.ICompSyncRequester, uint32) []types.Transaction
type FetchDeco = func(FetchFn) FetchFn

type BlockGenerator struct {
	bState   *state.BlockState
	rejected *RejTxInfo
	noTTE    bool // disable eviction by timeout if true

	hs               component.ICompSyncRequester
	bi               *types.BlockHeaderInfo
	txOp             TxOp
	fetchTXs         func(component.ICompSyncRequester, uint32) []types.Transaction
	skipEmpty        bool
	maxBlockBodySize uint32
}

func NewBlockGenerator(hs component.ICompSyncRequester, bi *types.BlockHeaderInfo, bState *state.BlockState,
	txOp TxOp, skipEmpty bool) *BlockGenerator {
	return &BlockGenerator{
		bState: bState,

		hs:               hs,
		bi:               bi,
		txOp:             txOp,
		fetchTXs:         FetchTXs,
		skipEmpty:        skipEmpty,
		maxBlockBodySize: MaxBlockBodySize(),
	}
}

type RejTxInfo struct {
	tx        types.Transaction
	orig      error
	evictable bool
}

func newTxRej(tx types.Transaction, orig error, evictable bool) *RejTxInfo {
	return &RejTxInfo{tx: tx, orig: orig, evictable: evictable}
}

func (r *RejTxInfo) Tx() types.Transaction {
	return r.tx
}

func (r *RejTxInfo) Hash() []byte {
	return r.tx.GetHash()
}

func (r *RejTxInfo) Evictable() bool {
	return r.evictable
}

func (g *BlockGenerator) Rejected() *RejTxInfo {
	return g.rejected
}

// SetTimeoutTx set bState.timeoutTx to tx.
func (g *BlockGenerator) SetTimeoutTx(tx types.Transaction) {
	logger.Warn().Str("hash", enc.ToString(tx.GetHash())).Msg("timeout tx marked for eviction")
	g.bState.SetTimeoutTx(tx)
}

// GenerateBlock generate & return a new block.
func (g *BlockGenerator) GenerateBlock() (*types.Block, error) {
	bState := g.bState

	transactions, err := g.GatherTXs()
	if err != nil {
		return nil, err
	}
	n := len(transactions)
	if n == 0 && g.skipEmpty {
		logger.Debug().Msg("BF: empty block is skipped")
		return nil, ErrBlockEmpty
	}

	txs := make([]*types.Tx, n)
	for i, x := range transactions {
		txs[i] = x.GetTx()
	}

	block := types.NewBlock(g.bi, bState.GetRoot(), bState.Receipts(), txs, chain.CoinbaseAccount, bState.Consensus())
	if n != 0 && logger.IsDebugEnabled() {
		logger.Debug().
			Str("txroothash", types.EncodeB64(block.GetHeader().GetTxsRootHash())).
			Int("hashed", len(txs)).
			Int("no_receipts", len(bState.Receipts().Get())).
			Msg("BF: tx root hash")
	}

	return block, nil
}

func (g *BlockGenerator) WithDeco(fn FetchDeco) *BlockGenerator {
	if fn != nil {
		g.fetchTXs = fn(g.fetchTXs)
	}
	return g
}

func (g *BlockGenerator) SetNoTTE(noTTE bool) *BlockGenerator {
	g.noTTE = noTTE
	return g
}

func (g *BlockGenerator) setRejected(tx types.Transaction, cause error, evictable bool) {
	g.rejected = newTxRej(tx, cause, evictable)
}

func (g *BlockGenerator) tteEnabled() bool {
	return !g.noTTE
}

// ConnectBlock send an AddBlock request to the chain service. This method is called only when this node
// produced a block.
func ConnectBlock(hs component.ICompSyncRequester, block *types.Block, blockState *state.BlockState, timeout time.Duration) error {
	// blockState does not include a valid BlockHash since it is constructed
	// from an incomplete block. So set it here.
	r, err := hs.RequestFuture(message.ChainSvc, &message.AddBlock{PeerID: "", Block: block, Bstate: blockState},
		timeout, "consensus/chain/info.ConnectBlock").Result()
	if err != nil {
		logger.Error().Err(err).Uint64("no", block.Header.BlockNo).
			Str("hash", block.ID()).
			Str("prev", block.PrevID()).
			Msg("failed to connect block")
		return &ErrBlockConnect{id: block.ID(), prevID: block.PrevID(), ec: err}
	}

	reply, ok := r.(*message.AddBlockRsp)
	if !ok {
		logger.Warn().Uint64("no", block.Header.BlockNo).
			Str("hash", block.ID()).
			Str("prev", block.PrevID()).
			Msg("ignore a weird add block response from chain service")
		return nil
	}
	if reply != nil && reply.Err != nil {
		return &ErrBlockConnect{id: block.ID(), prevID: block.PrevID(), ec: reply.Err}
	}

	return nil
}

func SyncChain(hs *component.ComponentHub, targetHash []byte, targetNo types.BlockNo, peerID types.PeerID) error {
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

			return err
		}
	}

	logger.Info().Str("peer", p2putil.ShortForm(peerID)).Msg("succeeded to sync for consensus")
	// TODO check best block is equal to target Hash/no
	return nil
}
