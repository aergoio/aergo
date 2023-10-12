package schema

import (
	"bytes"
	"fmt"

	"github.com/aergoio/aergo/v2/types"
)

func KeyReceipts(blockHash []byte, blockNo types.BlockNo) []byte {
	key := make([]byte, len(ReceiptsPrefix)+len(blockHash)+8)
	copy(key, []byte(ReceiptsPrefix))
	copy(key[len(ReceiptsPrefix):], blockHash)
	copy(key[len(ReceiptsPrefix)+len(blockHash):], types.BlockNoToBytes(blockNo))
	return key
}

// raft
func KeyRaftEntry(blockNo types.BlockNo) []byte {
	return append([]byte(RaftEntry), types.BlockNoToBytes(blockNo)...)
}

func KeyRaftEntryInvert(blockHash []byte) []byte {
	return append([]byte(RaftEntryInvert), blockHash...)
}

func KeyRaftConfChangeProgress(id uint64) []byte {
	return append([]byte(RaftConfChangeProgress), types.Uint64ToBytes(id)...)
}

// governance
func KeyEnterpriseConf(conf []byte) []byte {
	// upper double check
	return append([]byte(EnterpriseConf), bytes.ToUpper(conf)...)
}

func KeyName(name []byte) []byte {
	// lower double check
	return append([]byte(Name), bytes.ToLower(name)...)
}

func KeyParam(id []byte) []byte {
	// upper double check
	return append([]byte(SystemParam), bytes.ToUpper(id)...)
}

func KeyStaking(who []byte) []byte {
	return append([]byte(SystemStaking), who...)
}

func KeyVote(key, voter []byte) []byte {
	return append(append([]byte(SystemVote), key...), voter...)
}

func KeyVoteSort(key []byte) []byte {
	return append([]byte(SystemVoteSort), key...)
}

func KeyVoteTotal(key []byte) []byte {
	return append([]byte(SystemVoteTotal), key...)
}

func KeyVpr(i uint8) []byte {
	return append([]byte(SystemVpr), []byte(fmt.Sprintf("%v", i))...)
}
