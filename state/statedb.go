/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package state

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/internal/enc"
	"github.com/aergoio/aergo/pkg/trie"
	"github.com/aergoio/aergo/types"
	"github.com/golang/protobuf/proto"
)

const (
	stateName   = "state"
	stateLatest = stateName + ".latest"
)

var (
	logger = log.NewLogger(stateName)
)

var (
	emptyHashID    = types.HashID{}
	emptyBlockID   = types.BlockID{}
	emptyAccountID = types.AccountID{}
)

var (
	errSaveData       = errors.New("Failed to save data: invalid key")
	errLoadData       = errors.New("Failed to load data: invalid key")
	errSaveBlockState = errors.New("Failed to save blockState: invalid BlockID")
	errLoadBlockState = errors.New("Failed to load blockState: invalid BlockID")
	errSaveStateData  = errors.New("Failed to save StateData: invalid HashID")
	errLoadStateData  = errors.New("Failed to load StateData: invalid HashID")
)

type ChainStateDB struct {
	sync.RWMutex
	trie   *trie.Trie
	latest *types.BlockInfo
	store  *db.DB
}

func NewChainStateDB() *ChainStateDB {
	return &ChainStateDB{}
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
	if sdb.store == nil {
		sdb.store = InitDB(dataDir, stateName)
	}

	// init trie
	sdb.trie = trie.NewTrie(nil, types.TrieHasher, *sdb.store)

	// load data from db
	err := sdb.loadStateDB()
	if err != nil {
		return err
	}
	if sdb.latest != nil && !sdb.latest.StateRoot.Equal(emptyHashID) {
		sdb.trie.Root = sdb.latest.StateRoot[:]
		sdb.trie.LoadCache(sdb.trie.Root)
	}
	return nil
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
	if sdb.store != nil {
		(*sdb.store).Close()
	}
	return nil
}

func (sdb *ChainStateDB) SetGenesis(genesisBlock *types.Genesis) error {
	block := genesisBlock.Block
	gbInfo := &types.BlockInfo{
		BlockNo:   0,
		BlockHash: types.ToBlockID(block.BlockHash()),
	}
	sdb.latest = gbInfo

	// create state of genesis block
	gbState := types.NewBlockState(gbInfo)
	for address, balance := range genesisBlock.Balance {
		bytes := types.ToAddress(address)
		id := types.ToAccountID(bytes)
		gbState.PutAccount(id, &types.State{}, balance)
	}

	if genesisBlock.VoteState != nil {
		aid := types.ToAccountID([]byte(types.AergoSystem))
		gbState.PutAccount(aid, &types.State{}, genesisBlock.VoteState)
	}
	// save state of genesis block
	if err := sdb.apply(gbState); err != nil {
		return err
	}
	return nil
}

func (sdb *ChainStateDB) getAccountState(aid types.AccountID) (*types.State, error) {
	if aid == emptyAccountID {
		return nil, fmt.Errorf("Failed to get block account: invalid account id")
	}
	state, err := sdb.getAccountStateData(aid)
	if err != nil {
		return nil, err
	}
	return state, nil
}
func (sdb *ChainStateDB) getAccountStateData(aid types.AccountID) (*types.State, error) {
	dkey, err := sdb.trie.Get(aid[:])
	if err != nil {
		return nil, fmt.Errorf("Failed to get account state from trie: %s", err)
	}
	if len(dkey) == 0 {
		return types.NewState(), nil
	}
	data, err := sdb.loadStateData(dkey)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (sdb *ChainStateDB) GetAccountStateClone(aid types.AccountID) (*types.State, error) {
	state, err := sdb.getAccountState(aid)
	if err != nil {
		return nil, err
	}
	res := types.State(*state)
	return &res, nil
}
func (sdb *ChainStateDB) getBlockAccount(bs *types.BlockState, aid types.AccountID) (*types.State, error) {
	if aid == emptyAccountID {
		return nil, fmt.Errorf("Failed to get block account: invalid account id")
	}

	if prev, ok := bs.GetAccount(aid); ok {
		return prev, nil
	}
	return sdb.getAccountState(aid)
}
func (sdb *ChainStateDB) GetBlockAccountClone(bs *types.BlockState, aid types.AccountID) (*types.State, error) {
	state, err := sdb.getBlockAccount(bs, aid)
	if err != nil {
		return nil, err
	}
	res := types.State(*state)
	return &res, nil
}

func (sdb *ChainStateDB) updateTrie(bstate *types.BlockState) error {
	accounts := bstate.GetAccountStates()
	if len(accounts) <= 0 {
		// do nothing
		return nil
	}
	bufs := []cacheEntry{}
	for k, v := range accounts {
		data, err := proto.Marshal(v)
		if err != nil {
			return err
		}
		et := newCacheEntry(types.HashID(k), data)
		bufs = append(bufs, et)
	}
	caches := newStateCaches()
	caches.puts(bufs...)
	keys, vals := caches.export()
	_, err := sdb.trie.Update(keys, vals)
	if err != nil {
		return err
	}
	caches.commit(sdb.store)
	sdb.trie.Commit()
	return nil
}

func (sdb *ChainStateDB) revertTrie(prevBlockStateRoot types.HashID) error {
	var targetRoot []byte
	if !prevBlockStateRoot.Equal(emptyHashID) {
		targetRoot = prevBlockStateRoot[:]
	}

	if bytes.Equal(sdb.trie.Root, targetRoot) {
		// same root, do nothing
		return nil
	}
	err := sdb.trie.Revert(targetRoot)
	if err != nil {
		// FIXME: is that enough?
		// if prevRoot is not contained in the cached tries.
		sdb.trie.Root = targetRoot
		err = sdb.trie.LoadCache(sdb.trie.Root)
		return err
	}
	return nil
}

func (sdb *ChainStateDB) Apply(bstate *types.BlockState) error {
	if sdb.latest.BlockNo+1 != bstate.BlockNo {
		return fmt.Errorf("Failed to apply: invalid block no - latest=%v, this=%v", sdb.latest.BlockNo, bstate.BlockNo)
	}
	if sdb.latest.BlockHash != bstate.PrevHash {
		return fmt.Errorf("Failed to apply: invalid previous block latest=%v, bstate=%v",
			sdb.latest.BlockHash, bstate.PrevHash)
	}
	return sdb.apply(bstate)
}

func (sdb *ChainStateDB) apply(bstate *types.BlockState) error {
	sdb.Lock()
	defer sdb.Unlock()

	// rollback and revert trie requires state root before apply
	if bstate.Undo.StateRoot == emptyHashID {
		bstate.Undo.StateRoot = types.ToHashID(sdb.trie.Root)
	}

	// apply blockState to trie
	err := sdb.updateTrie(bstate)
	if err != nil {
		return err
	}

	// check state root
	if bstate.BlockInfo.StateRoot != types.ToHashID(sdb.GetHash()) {
		// TODO: if validation failed, than revert statedb.
		bstate.BlockInfo.StateRoot = types.ToHashID(sdb.GetHash())
	}
	logger.Debug().Str("stateRoot", enc.ToString(sdb.GetHash())).Msg("apply block state")

	// save blockState
	err = sdb.saveBlockState(bstate)
	if err != nil {
		return err
	}

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

		err = sdb.revertTrie(bs.Undo.StateRoot)
		if err != nil {
			return err
		}
		// logger.Debugf("- trie.root: %v", base64.StdEncoding.EncodeToString(sdb.GetHash()))

		target = &types.BlockInfo{
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

func (sdb *ChainStateDB) IsExistState(hash []byte) bool {
	//TODO : StateRootValidation
	return false
}
