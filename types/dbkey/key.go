package dbkey

import (
	"bytes"
	"fmt"

	"github.com/aergoio/aergo/v2/types"
)

//---------------------------------------------------------------------------------//
// state trie

func Trie(key []byte) []byte {
	return append([]byte(triePrefix), key...)
}

//---------------------------------------------------------------------------------//
// chain

func Receipts(blockHash []byte, blockNo types.BlockNo) []byte {
	key := make([]byte, len(receiptsPrefix)+len(blockHash)+8)
	copy(key, []byte(receiptsPrefix))
	copy(key[len(receiptsPrefix):], blockHash)
	copy(key[len(receiptsPrefix)+len(blockHash):], types.BlockNoToBytes(blockNo))
	return key
}

func InternalOps(txHash []byte) []byte {
	return append([]byte(internalOpsPrefix), txHash...)
}

//---------------------------------------------------------------------------------//
// metadata

func Genesis() []byte {
	return []byte(genesis)
}

func GenesisBalance() []byte {
	return []byte(genesisBalance)
}

func LatestBlock() []byte {
	return []byte(latestBlock)
}

func HardFork() []byte {
	return []byte(hardFork)
}

func ReOrg() []byte {
	return []byte(reOrg)
}

// dpos
func DposLibStatus() []byte {
	return []byte(dposLibStatus)
}

// raft
func RaftIdentity() []byte {
	return []byte(raftIdentity)
}

func RaftState() []byte {
	return []byte(raftState)
}

func RaftSnap() []byte {
	return []byte(raftSnap)
}

func RaftEntryLastIdx() []byte {
	return []byte(raftEntryLastIdx)
}

func RaftEntry(blockNo types.BlockNo) []byte {
	return append([]byte(raftEntry), types.BlockNoToBytes(blockNo)...)
}

func RaftEntryInvert(blockHash []byte) []byte {
	return append([]byte(raftEntryInvert), blockHash...)
}

func RaftConfChangeProgress(id uint64) []byte {
	return append([]byte(raftConfChangeProgress), types.Uint64ToBytes(id)...)
}

//---------------------------------------------------------------------------------//
// governance

// enterprise
func EnterpriseAdmins() []byte {
	return []byte(enterpriseAdmins)
}

func EnterpriseConf(conf []byte) []byte {
	// upper double check
	return append([]byte(enterpriseConf), bytes.ToUpper(conf)...)
}

// name
func Name(accountName []byte) []byte {
	// lower double check
	return append([]byte(name), bytes.ToLower(accountName)...)
}

// system
func SystemParam(id string) []byte {
	// upper double check
	return append([]byte(systemParam), bytes.ToUpper([]byte(id))...)
}

func SystemProposal() []byte {
	return []byte(systemProposal)
}

func SystemStaking(account []byte) []byte {
	return append([]byte(systemStaking), account...)
}

func SystemStakingTotal() []byte {
	return []byte(systemStakingTotal)
}

func SystemVote(key, voter []byte) []byte {
	return append(append([]byte(systemVote), key...), voter...)
}

func SystemVoteTotal(key []byte) []byte {
	return append([]byte(systemVoteTotal), key...)
}

func SystemVoteSort(key []byte) []byte {
	return append([]byte(systemVoteSort), key...)
}

func SystemVpr(i uint8) []byte {
	return append([]byte(systemVpr), []byte(fmt.Sprintf("%v", i))...)
}

// creator
func CreatorMeta() []byte {
	return []byte(creatorMeta)
}
