/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package state

import (
	"crypto/sha512"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/pkg/trie"
	"github.com/aergoio/aergo/types"
)

const (
	stateName     = "state"
	stateAccounts = stateName + ".accounts"
	stateLatest   = stateName + ".latest"
)

var (
	logger = log.NewLogger("state_db")
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

type StateEntry struct {
	State *types.State
	Undo  *types.State
}
type BlockState struct {
	BlockInfo
	accounts map[types.AccountID]*StateEntry
}

func NewStateEntry(state, undo *types.State) *StateEntry {
	if undo != nil && undo.IsEmpty() {
		undo = nil
	}
	return &StateEntry{
		State: state,
		Undo:  undo,
	}
}

func NewBlockState(blockNo types.BlockNo, blockHash, prevHash types.BlockID) *BlockState {
	return &BlockState{
		BlockInfo: BlockInfo{
			BlockNo:   blockNo,
			BlockHash: blockHash,
			PrevHash:  prevHash,
		},
		accounts: make(map[types.AccountID]*StateEntry),
	}
}

func (bs *BlockState) PutAccount(aid types.AccountID, state, change *types.State) {
	if prev, ok := bs.accounts[aid]; ok {
		prev.State = change
	} else {
		bs.accounts[aid] = NewStateEntry(change, state)
	}
}

type ChainStateDB struct {
	sync.RWMutex
	accounts map[types.AccountID]*types.State
	trie     *trie.SMT
	latest   *BlockInfo
	statedb  db.DB
}

func NewStateDB() *ChainStateDB {
	hasher := func(data ...[]byte) []byte {
		hasher := sha512.New512_256()
		for i := 0; i < len(data); i++ {
			hasher.Write(data[i])
		}
		return hasher.Sum(nil)
	}
	smt := trie.NewSMT(32, hasher, nil)

	return &ChainStateDB{
		accounts: make(map[types.AccountID]*types.State),
		trie:     smt,
	}
}

func InitDB(basePath, dbName string) db.DB {
	dbPath := path.Join(basePath, dbName)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	dbInst := db.NewDB(db.BadgerImpl, dbPath)
	return dbInst
}

func (sdb *ChainStateDB) Init(dataDir string) error {
	sdb.Lock()
	defer sdb.Unlock()

	// init db
	if sdb.statedb == nil {
		sdb.statedb = InitDB(dataDir, stateName)
	}

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
		sdb.statedb.Close()
	}
	return nil
}

func (sdb *ChainStateDB) SetGenesis(genesisBlock *types.Block) error {
	sdb.latest = &BlockInfo{
		BlockNo:   0,
		BlockHash: types.ToBlockID(genesisBlock.Hash),
	}
	// TODO: process initial coin tx
	err := sdb.saveStateDB()
	return err
}

func (sdb *ChainStateDB) GetAccountState(aid types.AccountID) (*types.State, error) {
	if aid == emptyAccountID {
		return nil, fmt.Errorf("Failed to get block account: invalid account id")
	}
	if state, ok := sdb.accounts[aid]; ok {
		return state, nil
	}
	state := types.NewState(aid)
	sdb.accounts[aid] = state
	return state, nil
}
func (sdb *ChainStateDB) GetAccount(bs *BlockState, aid types.AccountID) (*types.State, error) {
	if aid == emptyAccountID {
		return nil, fmt.Errorf("Failed to get block account: invalid account id")
	}

	if prev, ok := bs.accounts[aid]; ok {
		return prev.State, nil
	}
	return sdb.GetAccountState(aid)
}
func (sdb *ChainStateDB) GetAccountClone(bs *BlockState, aid types.AccountID) (*types.State, error) {
	state, err := sdb.GetAccount(bs, aid)
	if err != nil {
		return nil, err
	}
	res := types.Clone(*state).(types.State)
	return &res, nil
}

func (sdb *ChainStateDB) Apply(bstate *BlockState) error {
	if sdb.latest.BlockNo+1 != bstate.BlockNo {
		return fmt.Errorf("Failed to apply: invalid block no - latest=%v, this=%v", sdb.latest.BlockNo, bstate.BlockNo)
	}
	if sdb.latest.BlockHash != bstate.PrevHash {
		return fmt.Errorf("Failed to apply: invalid previous block latest=%v, bstate=%v",
			sdb.latest.BlockHash, bstate.PrevHash)
	}
	sdb.Lock()
	defer sdb.Unlock()

	sdb.saveBlockState(bstate)
	keys := trie.DataArray{bstate.BlockInfo.BlockHash[:]}
	vals := trie.DataArray{bstate.BlockInfo.PrevHash[:]}
	for k, v := range bstate.accounts {
		sdb.accounts[k] = v.State
		keys = append(keys, k[:])
		vals = append(vals, v.State.GetHash())
	}
	if len(keys) > 0 && len(vals) > 0 {
		sdb.trie.Update(keys, vals)
	}
	// logger.Debugf("- trie.root: %v", base64.StdEncoding.EncodeToString(sdb.GetHash()))
	sdb.latest = &bstate.BlockInfo
	err := sdb.saveStateDB()
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
		keys := trie.DataArray{bs.BlockInfo.BlockHash[:]}
		vals := trie.DataArray{bs.BlockInfo.PrevHash[:]}
		for k, v := range bs.accounts {
			sdb.accounts[k] = v.Undo
			keys = append(keys, k[:])
			vals = append(vals, v.State.GetHash())
		}
		if len(keys) > 0 && len(vals) > 0 {
			sdb.trie.Update(keys, vals)
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
