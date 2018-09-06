/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package state

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"sort"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/pkg/trie"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

const (
	stateName     = "state"
	stateAccounts = stateName + ".accounts"
	stateLatest   = stateName + ".latest"
)

var (
	logger = log.NewLogger("state")
)

var (
	emptyBlockID   = types.BlockID{}
	emptyAccountID = types.AccountID{}
)

type BlockInfo struct {
	BlockNo   types.BlockNo
	BlockHash types.BlockID
	PrevHash  types.BlockID
}

type BlockState struct {
	BlockInfo
	accounts accountStates
	undo     undoStates
}

type undoStates struct {
	stateRoot types.HashID
	accounts  accountStates
}

type accountStates map[types.AccountID]*types.State

// NewBlockState create new blockState contains blockNo, blockHash and blockHash of previous block
func NewBlockState(blockNo types.BlockNo, blockHash, prevHash types.BlockID) *BlockState {
	return newBlockState(&BlockInfo{
		BlockNo:   blockNo,
		BlockHash: blockHash,
		PrevHash:  prevHash,
	})
}
func newBlockState(blockInfo *BlockInfo) *BlockState {
	return &BlockState{
		BlockInfo: *blockInfo,
		accounts:  make(accountStates),
		undo: undoStates{
			accounts: make(accountStates),
		},
	}
}

// PutAccount sets before and changed state to blockState
func (bs *BlockState) PutAccount(aid types.AccountID, stateBefore, stateChanged *types.State) {
	if _, ok := bs.undo.accounts[aid]; !ok {
		bs.undo.accounts[aid] = stateBefore
	}
	bs.accounts[aid] = stateChanged
}

// SetBlockHash sets bs.BlockInfo.BlockHash to blockHash
func (bs *BlockState) SetBlockHash(blockHash types.BlockID) {
	if bs == nil {
		return
	}

	bs.BlockInfo.BlockHash = blockHash
}

type ChainStateDB struct {
	sync.RWMutex
	accounts map[types.AccountID]*types.State
	trie     *trie.Trie
	latest   *BlockInfo
	statedb  *db.DB
}

func NewStateDB() *ChainStateDB {
	return &ChainStateDB{
		accounts: make(map[types.AccountID]*types.State),
	}
}

func InitDB(basePath, dbName string) *db.DB {
	dbPath := path.Join(basePath, dbName)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	dbInst := db.NewDB(db.BadgerImpl, dbPath)
	return &dbInst
}

func (sdb *ChainStateDB) Init(dataDir string) error {
	sdb.Lock()
	defer sdb.Unlock()

	// init db
	if sdb.statedb == nil {
		sdb.statedb = InitDB(dataDir, stateName)
	}

	// init trie
	sdb.trie = trie.NewTrie(32, types.TrieHasher, *sdb.statedb)

	// load data from db
	err := sdb.loadStateDB()
	return err
}

func (sdb *ChainStateDB) Close() error {
	sdb.Lock()
	defer sdb.Unlock()

	// save data to db
	err := sdb.saveStateDB()
	if err != nil {
		return err
	}

	// close db
	if sdb.statedb != nil {
		(*sdb.statedb).Close()
	}
	return nil
}

func (sdb *ChainStateDB) SetGenesis(genesisBlock *types.Block) error {
	gbInfo := &BlockInfo{
		BlockNo:   0,
		BlockHash: types.ToBlockID(genesisBlock.Hash),
	}
	sdb.latest = gbInfo

	// create state of genesis block
	gbState := newBlockState(gbInfo)

	// // publish initial coin
	// sampleAccount := []byte("sample")
	// logger.Debug().Str("account", string(sampleAccount)).Str("base58", base58.Encode(sampleAccount)).Msg("init genesis block")
	// gbState.PutAccount(types.ToAccountID(sampleAccount), nil, &types.State{Balance: 10000})

	// save state of genesis block
	err := sdb.apply(gbState)
	return err
}

func (sdb *ChainStateDB) getAccountState(aid types.AccountID) (*types.State, error) {
	if aid == emptyAccountID {
		return nil, fmt.Errorf("Failed to get block account: invalid account id")
	}
	if state, ok := sdb.accounts[aid]; ok {
		return state, nil
	}
	state := types.NewState()
	sdb.accounts[aid] = state
	return state, nil
}
func (sdb *ChainStateDB) GetAccountStateClone(aid types.AccountID) (*types.State, error) {
	state, err := sdb.getAccountState(aid)
	if err != nil {
		return nil, err
	}
	res := types.Clone(*state).(types.State)
	return &res, nil
}
func (sdb *ChainStateDB) getBlockAccount(bs *BlockState, aid types.AccountID) (*types.State, error) {
	if aid == emptyAccountID {
		return nil, fmt.Errorf("Failed to get block account: invalid account id")
	}

	if prev, ok := bs.accounts[aid]; ok {
		return prev, nil
	}
	return sdb.getAccountState(aid)
}
func (sdb *ChainStateDB) GetBlockAccountClone(bs *BlockState, aid types.AccountID) (*types.State, error) {
	state, err := sdb.getBlockAccount(bs, aid)
	if err != nil {
		return nil, err
	}
	res := types.Clone(*state).(types.State)
	return &res, nil
}

func (sdb *ChainStateDB) updateTrie(bstate *BlockState) error {
	size := len(bstate.accounts)
	if size <= 0 {
		// do nothing
		return nil
	}
	accs := make([]types.AccountID, 0, size)
	for k := range bstate.accounts {
		accs = append(accs, k)
	}
	sort.Slice(accs, func(i, j int) bool {
		return bytes.Compare(accs[i][:], accs[j][:]) == -1
	})
	keys := make(trie.DataArray, size)
	vals := make(trie.DataArray, size)
	var err error
	for i, v := range accs {
		keys[i] = accs[i][:]
		vals[i], err = proto.Marshal(bstate.accounts[v])
		if err != nil {
			return err
		}
	}
	_, err = sdb.trie.Update(keys, vals)
	if err != nil {
		return err
	}
	sdb.trie.Commit()
	return nil
}

func (sdb *ChainStateDB) revertTrie(prevBlockStateRoot []byte) error {
	return sdb.trie.Revert(prevBlockStateRoot)
}

func (sdb *ChainStateDB) Apply(bstate *BlockState) error {
	if sdb.latest.BlockNo+1 != bstate.BlockNo {
		return fmt.Errorf("Failed to apply: invalid block no - latest=%v, this=%v", sdb.latest.BlockNo, bstate.BlockNo)
	}
	if sdb.latest.BlockHash != bstate.PrevHash {
		return fmt.Errorf("Failed to apply: invalid previous block latest=%v, bstate=%v",
			sdb.latest.BlockHash, bstate.PrevHash)
	}
	return sdb.apply(bstate)
}

func (sdb *ChainStateDB) apply(bstate *BlockState) error {
	sdb.Lock()
	defer sdb.Unlock()

	// save blockState
	sdb.saveBlockState(bstate)

	// apply blockState to statedb
	for k, v := range bstate.accounts {
		sdb.accounts[k] = v
	}
	// apply blockState to trie
	err := sdb.updateTrie(bstate)
	if err != nil {
		return err
	}
	// logger.Debugf("- trie.root: %v", base64.StdEncoding.EncodeToString(sdb.GetHash()))
	sdb.latest = &bstate.BlockInfo
	err = sdb.saveStateDB()
	return err
}

func (sdb *ChainStateDB) Rollback(blockNo types.BlockNo) error {
	if sdb.latest.BlockNo <= blockNo {
		return fmt.Errorf("Failed to rollback: invalid block no")
	}
	sdb.Lock()
	defer sdb.Unlock()

	target := sdb.latest
	for target.BlockNo >= blockNo {
		bs, err := sdb.loadBlockState(target.BlockHash)
		if err != nil {
			return err
		}
		sdb.latest = &bs.BlockInfo

		if target.BlockNo == blockNo {
			break
		}

		for k, v := range bs.undo.accounts {
			sdb.accounts[k] = v
		}
		err = sdb.revertTrie(bs.undo.stateRoot[:])
		if err != nil {
			return err
		}
		// logger.Debugf("- trie.root: %v", base64.StdEncoding.EncodeToString(sdb.GetHash()))

		target = &BlockInfo{
			BlockNo:   sdb.latest.BlockNo - 1,
			BlockHash: sdb.latest.PrevHash,
		}
	}
	err := sdb.saveStateDB()
	return err
}

func (sdb *ChainStateDB) GetHash() []byte {
	return sdb.trie.Root
}
