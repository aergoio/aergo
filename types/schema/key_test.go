package schema

import (
	"math"
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestReceiptsKey(t *testing.T) {
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
		key := ReceiptsKey(test.blockHash, test.blockNo)
		assert.Equal(t, test.expectKey, key, "TestReceiptsKey(%v, %v)", test.blockHash, test.blockNo)
	}
}

// raft
func TestRaftEntryKey(t *testing.T) {
	for _, test := range []struct {
		blockNo   types.BlockNo
		expectKey []byte
	}{
		{0, append([]byte(RaftEntry), 0, 0, 0, 0, 0, 0, 0, 0)},
		{1, append([]byte(RaftEntry), 1, 0, 0, 0, 0, 0, 0, 0)},
		{255, append([]byte(RaftEntry), 255, 0, 0, 0, 0, 0, 0, 0)},
		{math.MaxUint64, append([]byte(RaftEntry), 255, 255, 255, 255, 255, 255, 255, 255)},
	} {
		key := RaftEntryKey(test.blockNo)
		assert.Equal(t, test.expectKey, key, "TestRaftEntryKey(%v)", test.blockNo)
	}
}

func TestRaftEntryInvertKey(t *testing.T) {
	for _, test := range []struct {
		blockHash []byte
		expectKey []byte
	}{
		{[]byte{1, 2, 3, 4}, append([]byte(RaftEntryInvert), 1, 2, 3, 4)},
		{decodeB58("AiGVpwGUUs1kjK2oZkAEkzBzptZs25LoSakEtu5cCqFV"), append([]byte(RaftEntryInvert), decodeB58("AiGVpwGUUs1kjK2oZkAEkzBzptZs25LoSakEtu5cCqFV")...)},
		{decodeB58("5bSKqpcWnMgrr1GhU1Ed5yHajRC4WwZEZYxFtw3fVBmq"), append([]byte(RaftEntryInvert), decodeB58("5bSKqpcWnMgrr1GhU1Ed5yHajRC4WwZEZYxFtw3fVBmq")...)},
	} {
		key := RaftEntryInvertKey(test.blockHash)
		assert.Equal(t, test.expectKey, key, "TestRaftEntryInvertKey(%v)", test.blockHash)
	}
}

func TestRaftConfChangeProgressKey(t *testing.T) {
	for _, test := range []struct {
		id        uint64
		expectKey []byte
	}{
		{0, append([]byte(RaftConfChangeProgress), 0, 0, 0, 0, 0, 0, 0, 0)},
		{1, append([]byte(RaftConfChangeProgress), 1, 0, 0, 0, 0, 0, 0, 0)},
		{255, append([]byte(RaftConfChangeProgress), 255, 0, 0, 0, 0, 0, 0, 0)},
		{math.MaxUint64, append([]byte(RaftConfChangeProgress), 255, 255, 255, 255, 255, 255, 255, 255)},
	} {
		key := RaftConfChangeProgressKey(test.id)
		assert.Equal(t, test.expectKey, key, "TestRaftConfChangeProgressKey(%v)", test.id)
	}
}

// governance
func TestEnterpriseConfKey(t *testing.T) {
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
		key := EnterpriseConfKey(test.conf)
		assert.Equal(t, test.expectKey, key, "TestEnterpriseConfKey(%v)", test.conf)
	}
}

func TestNameKey(t *testing.T) {
	for _, test := range []struct {
		name      []byte
		expectKey []byte
	}{
		{nil, []byte(Name)},
		{[]byte("aergo.name"), append([]byte(Name), []byte("aergo.name")...)},
		{[]byte("AERGO.NAME"), append([]byte(Name), []byte("aergo.name")...)},
	} {
		key := NameKey(test.name)
		assert.Equal(t, test.expectKey, key, "TestNameKey(%v)", test.name)
	}
}

func TestSystemParamKey(t *testing.T) {
	for _, test := range []struct {
		param     string
		expectKey []byte
	}{
		{"", []byte(SystemParam)},
		{"bpCount", append([]byte(SystemParam), []byte("BPCOUNT")...)},
		{"stakingMin", append([]byte(SystemParam), []byte("STAKINGMIN")...)},
		{"gasPrice", append([]byte(SystemParam), []byte("GASPRICE")...)},
		{"namePrice", append([]byte(SystemParam), []byte("NAMEPRICE")...)},
	} {
		key := SystemParamKey(test.param)
		assert.Equal(t, test.expectKey, key, "TestSystemParamKey(%v)", test.param)
	}
}

func TestSystemStakingKey(t *testing.T) {
	for _, test := range []struct {
		account   []byte
		expectKey []byte
	}{
		{nil, []byte(SystemStaking)},
		{decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), append([]byte(SystemStaking), decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2")...)},
	} {
		key := SystemStakingKey(test.account)
		assert.Equal(t, test.expectKey, key, "TestSystemStakingKey(%v)", test.account)
	}
}

func TestSystemVoteKey(t *testing.T) {
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
		key := SystemVoteKey(test.key, test.voter)
		assert.Equal(t, test.expectKey, key, "TestSystemVoteKey(%v, %v)", test.key, test.voter)
	}
}

func TestSystemVoteSortKey(t *testing.T) {
	for _, test := range []struct {
		key       []byte
		expectKey []byte
	}{
		{[]byte("OpvoteBP"), append([]byte(SystemVoteSort), []byte("OpvoteBP")...)},
		{[]byte("OpvoteDAO"), append([]byte(SystemVoteSort), []byte("OpvoteDAO")...)},
		{[]byte("Opstake"), append([]byte(SystemVoteSort), []byte("Opstake")...)},
		{[]byte("Opunstake"), append([]byte(SystemVoteSort), []byte("Opunstake")...)},
	} {
		key := SystemVoteSortKey(test.key)
		assert.Equal(t, test.expectKey, key, "TestSystemVoteSortKey(%v)", test.key)
	}
}

func TestSystemVoteTotalKey(t *testing.T) {
	for _, test := range []struct {
		key       []byte
		expectKey []byte
	}{
		{[]byte("OpvoteBP"), append([]byte(SystemVoteTotal), []byte("OpvoteBP")...)},
		{[]byte("OpvoteDAO"), append([]byte(SystemVoteTotal), []byte("OpvoteDAO")...)},
		{[]byte("Opstake"), append([]byte(SystemVoteTotal), []byte("Opstake")...)},
		{[]byte("Opunstake"), append([]byte(SystemVoteTotal), []byte("Opunstake")...)},
	} {
		key := SystemVoteTotalKey(test.key)
		assert.Equal(t, test.expectKey, key, "TestSystemVoteTotalKey(%v)", test.key)
	}
}

func TestSystemVprKey(t *testing.T) {
	for _, test := range []struct {
		i         uint8
		expectKey []byte
	}{
		{0, append([]byte(SystemVpr), '0')},
		{1, append([]byte(SystemVpr), '1')},
		{255, append([]byte(SystemVpr), '2', '5', '5')},
	} {
		key := SystemVprKey(test.i)
		assert.Equal(t, test.expectKey, key, "TestSystemVprKey(%v)", test.i)
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
