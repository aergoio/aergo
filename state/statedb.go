/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package state

import (
	"crypto/sha512"
	"errors"
	"os"
	"path"
	"sync"

	"github.com/aergoio/aergo/pkg/db"
	"github.com/aergoio/aergo/pkg/log"
	"github.com/aergoio/aergo/pkg/trie"
	"github.com/aergoio/aergo/types"
)

const (
	stateName = "state"
)

var (
	logger = log.NewLogger(log.StateDB)
)

type StateCache struct {
	cache map[types.StateKey]*types.State
}

func NewStateCache() *StateCache {
	return &StateCache{
		cache: make(map[types.StateKey]*types.State),
	}
}

func (sc *StateCache) Get(skey types.StateKey) *types.State {
	return sc.cache[skey]
}

func (sc *StateCache) Put(state *types.State) types.StateKey {
	skey := types.ToStateKeyPb(state)
	sc.cache[skey] = state
	return skey
}

type BlockState struct {
	Accounts map[types.AccountKey]types.StateKey
}

func NewBlockState() *BlockState {
	return &BlockState{
		Accounts: make(map[types.AccountKey]types.StateKey),
	}
}
func (bs *BlockState) PutAccounts(bstate *BlockState) {
	for i, v := range bstate.Accounts {
		bs.Accounts[i] = v
	}
}

func (bs *BlockState) PutAccount(akey types.AccountKey, skey types.StateKey) {
	bs.Accounts[akey] = skey
}

func (bs *BlockState) CalculateRootHash() []byte {
	hasher := func(data ...[]byte) []byte {
		hasher := sha512.New512_256()
		for i := 0; i < len(data); i++ {
			hasher.Write(data[i])
		}
		return hasher.Sum(nil)
	}
	smt := trie.NewSMT(32, hasher, nil)
	keys := trie.DataArray{}
	vals := trie.DataArray{}
	for k, v := range bs.Accounts {
		keys = append(keys, k[:])
		vals = append(vals, v[:])
	}
	smt.Update(keys, vals)
	return smt.Root
}

// StateDB ...
// type StateDB interface {
// 	GetAccount(akey types.AccountKey) (*types.State, error)
// 	PutAccount(akey types.AccountKey, state *types.State) error
// }

type CachedStateDB struct {
	sync.RWMutex
	bstates map[types.BlockKey]*BlockState
	cache   map[types.StateKey]*types.State
	latest  types.BlockKey
	statedb db.DB
}

func NewCachedStateDB() *CachedStateDB {
	return &CachedStateDB{
		bstates: make(map[types.BlockKey]*BlockState),
		cache:   make(map[types.StateKey]*types.State),
	}
}

func InitDB(basePath, dbName string) db.DB {
	dbPath := path.Join(basePath, dbName)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dbPath, 0711)
	}
	st := db.NewDB(db.BadgerImpl, dbPath)
	return st
}

func (sdb *CachedStateDB) Init(dataDir string) error {
	if sdb.statedb == nil {
		sdb.statedb = InitDB(dataDir, stateName)
	}
	return nil
}

func (sdb *CachedStateDB) Close() {
	if sdb.statedb != nil {
		sdb.statedb.Close()
	}
}

func (sdb *CachedStateDB) GetBestBlockState() (*BlockState, error) {
	return sdb.GetBlockState(sdb.latest)
}
func (sdb *CachedStateDB) GetBlockState(bkey types.BlockKey) (*BlockState, error) {
	if bstate, ok := sdb.bstates[bkey]; ok {
		return bstate, nil
	}
	bs := &BlockState{}
	err := sdb.getData(bkey[:], bs)
	if err != nil {
		return nil, err
	}
	sdb.bstates[bkey] = bs
	return bs, nil
}
func (sdb *CachedStateDB) PutBlockState(bkey types.BlockKey, bs *BlockState) error {
	sdb.putData(bkey[:], bs)
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
