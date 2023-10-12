package schema

import (
	"math"
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestKeyReceipts(t *testing.T) {
	for _, test := range []struct {
		blockHash []byte
		blockNo   types.BlockNo
		expectKey []byte
	}{
		{nil, 0, append([]byte(ReceiptsPrefix), 0, 0, 0, 0, 0, 0, 0, 0)},
		{nil, 1, append([]byte(ReceiptsPrefix), 1, 0, 0, 0, 0, 0, 0, 0)},
		{nil, 255, append([]byte(ReceiptsPrefix), 255, 0, 0, 0, 0, 0, 0, 0)},
		{nil, math.MaxUint64, append([]byte(ReceiptsPrefix), 255, 255, 255, 255, 255, 255, 255, 255)},
		{[]byte{1, 2, 3, 4}, 0, append([]byte(ReceiptsPrefix), 1, 2, 3, 4, 0, 0, 0, 0, 0, 0, 0, 0)},
		{decodeB58("AiGVpwGUUs1kjK2oZkAEkzBzptZs25LoSakEtu5cCqFV"), 0, append([]byte(ReceiptsPrefix), append(decodeB58("AiGVpwGUUs1kjK2oZkAEkzBzptZs25LoSakEtu5cCqFV"), 0, 0, 0, 0, 0, 0, 0, 0)...)},
		{decodeB58("5bSKqpcWnMgrr1GhU1Ed5yHajRC4WwZEZYxFtw3fVBmq"), 0, append([]byte(ReceiptsPrefix), append(decodeB58("5bSKqpcWnMgrr1GhU1Ed5yHajRC4WwZEZYxFtw3fVBmq"), 0, 0, 0, 0, 0, 0, 0, 0)...)},
	} {
		key := KeyReceipts(test.blockHash, test.blockNo)
		assert.Equal(t, test.expectKey, key, "TestKeyReceipts(%v, %v)", test.blockHash, test.blockNo)
	}
}

// raft
func TestKeyRaftEntry(t *testing.T) {
	for _, test := range []struct {
		blockNo   types.BlockNo
		expectKey []byte
	}{
		{0, append([]byte(RaftEntry), 0, 0, 0, 0, 0, 0, 0, 0)},
		{1, append([]byte(RaftEntry), 1, 0, 0, 0, 0, 0, 0, 0)},
		{255, append([]byte(RaftEntry), 255, 0, 0, 0, 0, 0, 0, 0)},
		{math.MaxUint64, append([]byte(RaftEntry), 255, 255, 255, 255, 255, 255, 255, 255)},
	} {
		key := KeyRaftEntry(test.blockNo)
		assert.Equal(t, test.expectKey, key, "TestKeyRaftEntry(%v)", test.blockNo)
	}
}

func TestKeyRaftEntryInvert(t *testing.T) {
	for _, test := range []struct {
		blockHash []byte
		expectKey []byte
	}{
		{[]byte{1, 2, 3, 4}, append([]byte(RaftEntryInvert), 1, 2, 3, 4)},
		{decodeB58("AiGVpwGUUs1kjK2oZkAEkzBzptZs25LoSakEtu5cCqFV"), append([]byte(RaftEntryInvert), decodeB58("AiGVpwGUUs1kjK2oZkAEkzBzptZs25LoSakEtu5cCqFV")...)},
		{decodeB58("5bSKqpcWnMgrr1GhU1Ed5yHajRC4WwZEZYxFtw3fVBmq"), append([]byte(RaftEntryInvert), decodeB58("5bSKqpcWnMgrr1GhU1Ed5yHajRC4WwZEZYxFtw3fVBmq")...)},
	} {
		key := KeyRaftEntryInvert(test.blockHash)
		assert.Equal(t, test.expectKey, key, "TestKeyRaftEntryInvert(%v)", test.blockHash)
	}
}

func TestKeyRaftConfChangeProgress(t *testing.T) {
	for _, test := range []struct {
		id        uint64
		expectKey []byte
	}{
		{0, append([]byte(RaftConfChangeProgress), 0, 0, 0, 0, 0, 0, 0, 0)},
		{1, append([]byte(RaftConfChangeProgress), 1, 0, 0, 0, 0, 0, 0, 0)},
		{255, append([]byte(RaftConfChangeProgress), 255, 0, 0, 0, 0, 0, 0, 0)},
		{math.MaxUint64, append([]byte(RaftConfChangeProgress), 255, 255, 255, 255, 255, 255, 255, 255)},
	} {
		key := KeyRaftConfChangeProgress(test.id)
		assert.Equal(t, test.expectKey, key, "TestKeyRaftConfChangeProgress(%v)", test.id)
	}
}

// governance
func TestKeyEnterpriseConf(t *testing.T) {
	for _, test := range []struct {
		conf      []byte
		expectKey []byte
	}{
		{[]byte("rpcpermissions"), append([]byte(EnterpriseConf), []byte("RPCPERMISSIONS")...)},
		{[]byte("RPCPERMISSIONS"), append([]byte(EnterpriseConf), []byte("RPCPERMISSIONS")...)},
		{[]byte("p2pwhite"), append([]byte(EnterpriseConf), []byte("P2PWHITE")...)},
		{[]byte("P2PWHITE"), append([]byte(EnterpriseConf), []byte("P2PWHITE")...)},
		{[]byte("p2pblack"), append([]byte(EnterpriseConf), []byte("P2PBLACK")...)},
		{[]byte("P2PBLACK"), append([]byte(EnterpriseConf), []byte("P2PBLACK")...)},
		{[]byte("accountwhite"), append([]byte(EnterpriseConf), []byte("ACCOUNTWHITE")...)},
		{[]byte("ACCOUNTWHITE"), append([]byte(EnterpriseConf), []byte("ACCOUNTWHITE")...)},
	} {
		key := KeyEnterpriseConf(test.conf)
		assert.Equal(t, test.expectKey, key, "TestKeyRaftConfChangeProgress(%v)", test.conf)
	}
}

func TestKeyName(t *testing.T) {
	for _, test := range []struct {
		name      []byte
		expectKey []byte
	}{
		{nil, []byte(Name)},
		{[]byte("aergo.name"), append([]byte(Name), []byte("aergo.name")...)},
		{[]byte("AERGO.NAME"), append([]byte(Name), []byte("aergo.name")...)},
	} {
		key := KeyName(test.name)
		assert.Equal(t, test.expectKey, key, "TestKeyName(%v)", test.name)
	}
}

func TestKeyParam(t *testing.T) {
	for _, test := range []struct {
		param     []byte
		expectKey []byte
	}{
		{nil, []byte(SystemParam)},
		{[]byte("bpCount"), append([]byte(SystemParam), []byte("BPCOUNT")...)},
		{[]byte("stakingMin"), append([]byte(SystemParam), []byte("STAKINGMIN")...)},
		{[]byte("gasPrice"), append([]byte(SystemParam), []byte("GASPRICE")...)},
		{[]byte("namePrice"), append([]byte(SystemParam), []byte("NAMEPRICE")...)},
	} {
		key := KeyParam(test.param)
		assert.Equal(t, test.expectKey, key, "TestKeyParam(%v)", test.param)
	}
}

func TestKeyStaking(t *testing.T) {
	for _, test := range []struct {
		who       []byte
		expectKey []byte
	}{
		{nil, []byte(SystemStaking)},
		{decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), append([]byte(SystemStaking), decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2")...)},
	} {
		key := KeyStaking(test.who)
		assert.Equal(t, test.expectKey, key, "TestKeyStaking(%v)", test.who)
	}
}

func TestKeyVote(t *testing.T) {
	for _, test := range []struct {
		key       []byte
		voter     []byte
		expectKey []byte
	}{
		{decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("OpvoteBP"), append([]byte(SystemVote), append(decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("OpvoteBP")...)...)},
		{decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("OpvoteDAO"), append([]byte(SystemVote), append(decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("OpvoteDAO")...)...)},
		{decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("Opstake"), append([]byte(SystemVote), append(decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("Opstake")...)...)},
		{decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("Opunstake"), append([]byte(SystemVote), append(decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("Opunstake")...)...)},
	} {
		key := KeyVote(test.key, test.voter)
		assert.Equal(t, test.expectKey, key, "TestKeyVote(%v, %v)", test.key, test.voter)
	}
}

func TestKeyVoteSort(t *testing.T) {
	for _, test := range []struct {
		key       []byte
		expectKey []byte
	}{
		{[]byte("OpvoteBP"), append([]byte(SystemVoteSort), []byte("OpvoteBP")...)},
		{[]byte("OpvoteDAO"), append([]byte(SystemVoteSort), []byte("OpvoteDAO")...)},
		{[]byte("Opstake"), append([]byte(SystemVoteSort), []byte("Opstake")...)},
		{[]byte("Opunstake"), append([]byte(SystemVoteSort), []byte("Opunstake")...)},
	} {
		key := KeyVoteSort(test.key)
		assert.Equal(t, test.expectKey, key, "TestKeyVoteSort(%v)", test.key)
	}
}

func TestKeyVoteTotal(t *testing.T) {
	for _, test := range []struct {
		key       []byte
		expectKey []byte
	}{
		{[]byte("OpvoteBP"), append([]byte(SystemVoteTotal), []byte("OpvoteBP")...)},
		{[]byte("OpvoteDAO"), append([]byte(SystemVoteTotal), []byte("OpvoteDAO")...)},
		{[]byte("Opstake"), append([]byte(SystemVoteTotal), []byte("Opstake")...)},
		{[]byte("Opunstake"), append([]byte(SystemVoteTotal), []byte("Opunstake")...)},
	} {
		key := KeyVoteTotal(test.key)
		assert.Equal(t, test.expectKey, key, "TestKeyVoteTotal(%v)", test.key)
	}
}

func TestKeyVpr(t *testing.T) {
	for _, test := range []struct {
		i         uint8
		expectKey []byte
	}{
		{0, append([]byte(SystemVpr), '0')},
		{1, append([]byte(SystemVpr), '1')},
		{255, append([]byte(SystemVpr), '2', '5', '5')},
	} {
		key := KeyVpr(test.i)
		assert.Equal(t, test.expectKey, key, "TestKeyVpr(%v)", test.i)
	}
}

//------------------------------------------------------------------//
// util

func decodeB58(s string) []byte {
	return types.DecodeB58(s)
}

func encodeB58(bt []byte) string {
	return types.EncodeB58(bt)
}

func decodeAddr(addr string) []byte {
	raw, _ := types.DecodeAddress(addr)
	return raw
}

func encodeAddr(raw []byte) string {
	return types.EncodeAddress(raw)
}
