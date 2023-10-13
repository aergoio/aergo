package schema

import (
	"bytes"
	"fmt"

	"github.com/aergoio/aergo/v2/types"
)

func ReceiptsKey(blockHash []byte, blockNo types.BlockNo) []byte {
	key := make([]byte, len(ReceiptsPrefix)+len(blockHash)+8)
	copy(key, []byte(ReceiptsPrefix))
	copy(key[len(ReceiptsPrefix):], blockHash)
	copy(key[len(ReceiptsPrefix)+len(blockHash):], types.BlockNoToBytes(blockNo))
	return key
}

// raft
func RaftEntryKey(blockNo types.BlockNo) []byte {
	return append([]byte(RaftEntry), types.BlockNoToBytes(blockNo)...)
}

func RaftEntryInvertKey(blockHash []byte) []byte {
	return append([]byte(RaftEntryInvert), blockHash...)
}

func RaftConfChangeProgressKey(id uint64) []byte {
	return append([]byte(RaftConfChangeProgress), types.Uint64ToBytes(id)...)
}

// governance
func EnterpriseConfKey(conf []byte) []byte {
	// upper double check
	return append([]byte(EnterpriseConf), bytes.ToUpper(conf)...)
}

func NameKey(name []byte) []byte {
	// lower double check
	return append([]byte(Name), bytes.ToLower(name)...)
}

func ParamKey(id string) []byte {
	// upper double check
	return append([]byte(SystemParam), bytes.ToUpper([]byte(id))...)
}

func StakingKey(who []byte) []byte {
	return append([]byte(SystemStaking), who...)
}

func VoteKey(key, voter []byte) []byte {
	return append(append([]byte(SystemVote), key...), voter...)
}

func VoteSortKey(key []byte) []byte {
	return append([]byte(SystemVoteSort), key...)
}

func VoteTotalKey(key []byte) []byte {
	return append([]byte(SystemVoteTotal), key...)
}

func VprKey(i uint8) []byte {
	return append([]byte(SystemVpr), []byte(fmt.Sprintf("%v", i))...)
}
