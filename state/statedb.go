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
	stateAccounts = "state.accounts"
	stateLatest   = "state.latest"
)

var (
	logger = log.NewLogger(log.StateDB)
)

type BlockInfo struct {
	blockNo   types.BlockNo
	blockHash types.BlockKey
	prevHash  types.BlockKey
}
type BlockState struct {
	BlockInfo
	accounts map[types.AccountKey]*types.State
	undolog  map[types.AccountKey]*types.State
}

func NewBlockState(blockNo types.BlockNo, blockHash, prevHash types.BlockKey) *BlockState {
	return &BlockState{
		BlockInfo: BlockInfo{
			blockNo:   blockNo,
			blockHash: blockHash,
			prevHash:  prevHash,
		},
		accounts: make(map[types.AccountKey]*types.State),
		undolog:  make(map[types.AccountKey]*types.State),
	}
}

/*
func (bs *BlockState) PutAccount(akey types.AccountKey, state *types.State) {
	bs.accounts[akey] = state
}
func (bs *BlockState) GetAccount(akey types.AccountKey) *types.State {
	return bs.accounts[akey]
}
*/

// StateDB ...
// type StateDB interface {
// 	GetAccount(akey types.AccountKey) (*types.State, error)
// 	PutAccount(akey types.AccountKey, state *types.State) error
// }

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
	err := sdb.load()
	if err != nil {
		return err
	}
	return nil
}

func (sdb *CachedStateDB) Close() error {
	sdb.Lock()
	defer sdb.Unlock()
	// save data to db
	err := sdb.save()
	if err != nil {
		return err
	}
	// close db
	if sdb.statedb != nil {
		sdb.statedb.Close()
	}
	return nil
}

/*
func (sdb *CachedStateDB) GetBestBlockState() (*BlockState, error) {
	return sdb.GetBlockState(sdb.latest)
}
func (sdb *CachedStateDB) GetBlockState(bkey types.BlockKey) (*BlockState, error) {
	if bstate, ok := sdb.bstates[bkey]; ok {
		return bstate, nil
	}
	bs := &BlockState{}
	err := sdb.loadData(bkey[:], bs)
	if err != nil {
		return nil, err
	}
	sdb.bstates[bkey] = bs
	return bs, nil
}
func (sdb *CachedStateDB) PutBlockState(bkey types.BlockKey, bs *BlockState) error {
	sdb.saveData(bkey[:], bs)
	sdb.bstates[bkey] = bs
	sdb.latest = bkey
	return nil
}
func (sdb *CachedStateDB) PutStateCache(sc *StateCache) {
	if sc == nil {
		return
	}
	for k, v := range sc.cache {
		sdb.PutState(k, v)
	}
}

func (sdb *CachedStateDB) GetBlockAccount(bstate *BlockState, akey types.AccountKey) (*types.State, error) {
	if akey == types.EmptyAccountKey {
		return nil, errors.New("failed to get account state. account key is empty")
	}
	if skey, ok := bstate.Accounts[akey]; ok {
		return sdb.GetState(skey)
	}
	return types.NewState(akey), nil
}
func (sdb *CachedStateDB) GetAccount(akey types.AccountKey) (*types.State, error) {
	bstate, err := sdb.GetBestBlockState()
	if err != nil {
		return nil, err
	}
	return sdb.GetBlockAccount(bstate, akey)
}

func (sdb *CachedStateDB) NewBlockState(prev types.BlockKey) (*BlockState, error) {
	prevbs, err := sdb.GetBlockState(prev)
	if err != nil {
		return nil, err
	}
	bs := NewBlockState()
	bs.PutAccounts(prevbs)
	return bs, nil
}
*/

func (sdb *CachedStateDB) GetBlockAccount(bstate *BlockState, akey types.AccountKey) (*types.State, error) {
	if akey == types.EmptyAccountKey {
		return nil, fmt.Errorf("Failed to get block account: invalid account key")
	}
	if account, ok := bstate.accounts[akey]; ok {
		return types.Clone(account).(*types.State), nil
	}

	if state, ok := sdb.accounts[akey]; ok {
		return types.Clone(state).(*types.State), nil
	}
	state := types.NewState(akey)
	sdb.accounts[akey] = state
	return types.Clone(state).(*types.State), nil
}

// func (sdb *CachedStateDB) PutAccount(akey types.AccountKey, state *types.State) error {
// 	sdb.accounts[akey] = state
// 	return nil
// }

func (sdb *CachedStateDB) Apply(bstate *BlockState) error {
	if sdb.latest.blockNo+1 != bstate.blockNo {
		return fmt.Errorf("Failed to apply: invalid block no")
	}
	if sdb.latest.blockHash != bstate.prevHash {
		return fmt.Errorf("Failed to apply: invalid previous block")
	}
	sdb.Lock()
	defer sdb.Unlock()

	sdb.saveData(bstate.blockHash[:], bstate)
	for i, v := range bstate.accounts {
		sdb.accounts[i] = v
	}
	sdb.latest = &bstate.BlockInfo
	sdb.save()
	return nil
}

func (sdb *CachedStateDB) Rollback(blockNo types.BlockNo) error {
	if sdb.latest.blockNo <= blockNo {
		return fmt.Errorf("Failed to rollback: invalid block no")
	}
	sdb.Lock()
	defer sdb.Unlock()

	for sdb.latest.blockNo > blockNo {
		bs := &BlockState{}
		err := sdb.loadData(sdb.latest.blockHash[:], bs)
		if err != nil {
			return err
		}
		for k, v := range bs.undolog {
			sdb.accounts[k] = v
		}
		sdb.latest = &bs.BlockInfo
	}
	sdb.save()
	return nil
}
