/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package state

import (
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/aergoio/aergo/pkg/db"
	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/types"
)

const (
	stateName     = "state"
	stateAccounts = stateName + ".accounts"
	stateLatest   = stateName + ".latest"
)

var (
	logger = log.NewLogger(log.StateDB)
)

type BlockInfo struct {
	blockNo   types.BlockNo
	blockHash types.BlockKey
	prevHash  types.BlockKey
}
type StateEntry struct {
	state *types.State
	undo  *types.State
}
type BlockState struct {
	BlockInfo
	accounts map[types.AccountKey]*StateEntry
}

func NewBlockState(blockNo types.BlockNo, blockHash, prevHash types.BlockKey) *BlockState {
	return &BlockState{
		BlockInfo: BlockInfo{
			blockNo:   blockNo,
			blockHash: blockHash,
			prevHash:  prevHash,
		},
		accounts: make(map[types.AccountKey]*StateEntry),
	}
}

func (bs *BlockState) PutAccount(akey types.AccountKey, state, change *types.State) {
	if prev, ok := bs.accounts[akey]; ok {
		prev.state = change
	} else {
		bs.accounts[akey] = &StateEntry{
			state: change,
			undo:  state,
		}
	}
}

type CachedStateDB struct {
	sync.RWMutex
	accounts map[types.AccountKey]*types.State
	latest   *BlockInfo
	statedb  db.DB
}

func NewCachedStateDB() *CachedStateDB {
	return &CachedStateDB{
		accounts: make(map[types.AccountKey]*types.State),
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

func (sdb *CachedStateDB) Init(dataDir string) error {
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

func (sdb *CachedStateDB) Close() error {
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

func (sdb *CachedStateDB) GetAccountState(akey types.AccountKey) (*types.State, error) {
	if akey == types.EmptyAccountKey {
		return nil, fmt.Errorf("Failed to get block account: invalid account key")
	}
	if state, ok := sdb.accounts[akey]; ok {
		return state, nil
	}
	state := types.NewState(akey)
	sdb.accounts[akey] = state
	return state, nil
}
func (sdb *CachedStateDB) GetAccount(bs *BlockState, akey types.AccountKey) (*types.State, error) {
	if akey == types.EmptyAccountKey {
		return nil, fmt.Errorf("Failed to get block account: invalid account key")
	}

	if prev, ok := bs.accounts[akey]; ok {
		return prev.state, nil
	}
	return sdb.GetAccountState(akey)
}
func (sdb *CachedStateDB) GetAccountClone(bs *BlockState, akey types.AccountKey) (*types.State, error) {
	state, err := sdb.GetAccount(bs, akey)
	if err != nil {
		return nil, err
	}
	return types.Clone(state).(*types.State), nil
}

func (sdb *CachedStateDB) Apply(bstate *BlockState) error {
	if sdb.latest.blockNo+1 != bstate.blockNo {
		return fmt.Errorf("Failed to apply: invalid block no")
	}
	if sdb.latest.blockHash != bstate.prevHash {
		return fmt.Errorf("Failed to apply: invalid previous block")
	}
	sdb.Lock()
	defer sdb.Unlock()

	sdb.saveBlockState(bstate)
	for k, v := range bstate.accounts {
		sdb.accounts[k] = v.state
	}
	sdb.latest = &bstate.BlockInfo
	sdb.saveStateDB()
	return nil
}

func (sdb *CachedStateDB) Rollback(blockNo types.BlockNo) error {
	if sdb.latest.blockNo <= blockNo {
		return fmt.Errorf("Failed to rollback: invalid block no")
	}
	sdb.Lock()
	defer sdb.Unlock()

	for sdb.latest.blockNo > blockNo {
		bs, err := sdb.loadBlockState(sdb.latest.blockHash)
		if err != nil {
			return err
		}
		for k, v := range bs.accounts {
			sdb.accounts[k] = v.undo
		}
		sdb.latest = &bs.BlockInfo
	}
	sdb.saveStateDB()
	return nil
}
