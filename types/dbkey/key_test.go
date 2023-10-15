package dbkey

import (
	"math"
	"testing"

	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestReceipts(t *testing.T) {
	for _, test := range []struct {
		blockHash []byte
		blockNo   types.BlockNo
		expectKey []byte
	}{
		{nil, 0, append([]byte(receiptsPrefix), 0, 0, 0, 0, 0, 0, 0, 0)},
		{nil, 1, append([]byte(receiptsPrefix), 1, 0, 0, 0, 0, 0, 0, 0)},
		{nil, 255, append([]byte(receiptsPrefix), 255, 0, 0, 0, 0, 0, 0, 0)},
		{nil, math.MaxUint64, append([]byte(receiptsPrefix), 255, 255, 255, 255, 255, 255, 255, 255)},
		{[]byte{1, 2, 3, 4}, 0, append([]byte(receiptsPrefix), 1, 2, 3, 4, 0, 0, 0, 0, 0, 0, 0, 0)},
		{decodeB58("AiGVpwGUUs1kjK2oZkAEkzBzptZs25LoSakEtu5cCqFV"), 0, append([]byte(receiptsPrefix), append(decodeB58("AiGVpwGUUs1kjK2oZkAEkzBzptZs25LoSakEtu5cCqFV"), 0, 0, 0, 0, 0, 0, 0, 0)...)},
		{decodeB58("5bSKqpcWnMgrr1GhU1Ed5yHajRC4WwZEZYxFtw3fVBmq"), 0, append([]byte(receiptsPrefix), append(decodeB58("5bSKqpcWnMgrr1GhU1Ed5yHajRC4WwZEZYxFtw3fVBmq"), 0, 0, 0, 0, 0, 0, 0, 0)...)},
	} {
		key := Receipts(test.blockHash, test.blockNo)
		assert.Equal(t, test.expectKey, key, "TestReceipts(%v, %v)", test.blockHash, test.blockNo)
	}
}

// raft
func TestRaftEntry(t *testing.T) {
	for _, test := range []struct {
		blockNo   types.BlockNo
		expectKey []byte
	}{
		{0, append([]byte(raftEntry), 0, 0, 0, 0, 0, 0, 0, 0)},
		{1, append([]byte(raftEntry), 1, 0, 0, 0, 0, 0, 0, 0)},
		{255, append([]byte(raftEntry), 255, 0, 0, 0, 0, 0, 0, 0)},
		{math.MaxUint64, append([]byte(raftEntry), 255, 255, 255, 255, 255, 255, 255, 255)},
	} {
		key := RaftEntry(test.blockNo)
		assert.Equal(t, test.expectKey, key, "TestraftEntry(%v)", test.blockNo)
	}
}

func TestRaftEntryInvert(t *testing.T) {
	for _, test := range []struct {
		blockHash []byte
		expectKey []byte
	}{
		{[]byte{1, 2, 3, 4}, append([]byte(raftEntryInvert), 1, 2, 3, 4)},
		{decodeB58("AiGVpwGUUs1kjK2oZkAEkzBzptZs25LoSakEtu5cCqFV"), append([]byte(raftEntryInvert), decodeB58("AiGVpwGUUs1kjK2oZkAEkzBzptZs25LoSakEtu5cCqFV")...)},
		{decodeB58("5bSKqpcWnMgrr1GhU1Ed5yHajRC4WwZEZYxFtw3fVBmq"), append([]byte(raftEntryInvert), decodeB58("5bSKqpcWnMgrr1GhU1Ed5yHajRC4WwZEZYxFtw3fVBmq")...)},
	} {
		key := RaftEntryInvert(test.blockHash)
		assert.Equal(t, test.expectKey, key, "TestraftEntryInvert(%v)", test.blockHash)
	}
}

func TestRaftConfChangeProgress(t *testing.T) {
	for _, test := range []struct {
		id        uint64
		expectKey []byte
	}{
		{0, append([]byte(raftConfChangeProgress), 0, 0, 0, 0, 0, 0, 0, 0)},
		{1, append([]byte(raftConfChangeProgress), 1, 0, 0, 0, 0, 0, 0, 0)},
		{255, append([]byte(raftConfChangeProgress), 255, 0, 0, 0, 0, 0, 0, 0)},
		{math.MaxUint64, append([]byte(raftConfChangeProgress), 255, 255, 255, 255, 255, 255, 255, 255)},
	} {
		key := RaftConfChangeProgress(test.id)
		assert.Equal(t, test.expectKey, key, "TestraftConfChangeProgress(%v)", test.id)
	}
}

// governance
func TestEnterpriseConf(t *testing.T) {
	for _, test := range []struct {
		conf      []byte
		expectKey []byte
	}{
		{[]byte("rpcpermissions"), append([]byte(enterpriseConf), []byte("RPCPERMISSIONS")...)},
		{[]byte("RPCPERMISSIONS"), append([]byte(enterpriseConf), []byte("RPCPERMISSIONS")...)},
		{[]byte("p2pwhite"), append([]byte(enterpriseConf), []byte("P2PWHITE")...)},
		{[]byte("P2PWHITE"), append([]byte(enterpriseConf), []byte("P2PWHITE")...)},
		{[]byte("p2pblack"), append([]byte(enterpriseConf), []byte("P2PBLACK")...)},
		{[]byte("P2PBLACK"), append([]byte(enterpriseConf), []byte("P2PBLACK")...)},
		{[]byte("accountwhite"), append([]byte(enterpriseConf), []byte("ACCOUNTWHITE")...)},
		{[]byte("ACCOUNTWHITE"), append([]byte(enterpriseConf), []byte("ACCOUNTWHITE")...)},
	} {
		key := EnterpriseConf(test.conf)
		assert.Equal(t, test.expectKey, key, "TestEnterpriseConf(%v)", test.conf)
	}
}

func TestName(t *testing.T) {
	for _, test := range []struct {
		name      []byte
		expectKey []byte
	}{
		{nil, []byte(name)},
		{[]byte("aergo.name"), append([]byte(name), []byte("aergo.name")...)},
		{[]byte("AERGO.NAME"), append([]byte(name), []byte("aergo.name")...)},
	} {
		key := Name(test.name)
		assert.Equal(t, test.expectKey, key, "TestName(%v)", test.name)
	}
}

func TestSystemParam(t *testing.T) {
	for _, test := range []struct {
		param     string
		expectKey []byte
	}{
		{"", []byte(systemParam)},
		{"bpCount", append([]byte(systemParam), []byte("BPCOUNT")...)},
		{"stakingMin", append([]byte(systemParam), []byte("STAKINGMIN")...)},
		{"gasPrice", append([]byte(systemParam), []byte("GASPRICE")...)},
		{"namePrice", append([]byte(systemParam), []byte("NAMEPRICE")...)},
	} {
		key := SystemParam(test.param)
		assert.Equal(t, test.expectKey, key, "TestSystemParam(%v)", test.param)
	}
}

func TestSystemStaking(t *testing.T) {
	for _, test := range []struct {
		account   []byte
		expectKey []byte
	}{
		{nil, []byte(systemStaking)},
		{decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), append([]byte(systemStaking), decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2")...)},
	} {
		key := SystemStaking(test.account)
		assert.Equal(t, test.expectKey, key, "TestSystemStaking(%v)", test.account)
	}
}

func TestSystemVote(t *testing.T) {
	for _, test := range []struct {
		key       []byte
		voter     []byte
		expectKey []byte
	}{
		{decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("OpvoteBP"), append([]byte(systemVote), append(decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("OpvoteBP")...)...)},
		{decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("OpvoteDAO"), append([]byte(systemVote), append(decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("OpvoteDAO")...)...)},
		{decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("Opstake"), append([]byte(systemVote), append(decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("Opstake")...)...)},
		{decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("Opunstake"), append([]byte(systemVote), append(decodeAddr("AmNpn7K9wg6wsn6oMkTirQSUNdqtDm94iCrrpP5ZpwCAAxxPrsU2"), []byte("Opunstake")...)...)},
	} {
		key := SystemVote(test.key, test.voter)
		assert.Equal(t, test.expectKey, key, "TestSystemVote(%v, %v)", test.key, test.voter)
	}
}

func TestSystemVoteSort(t *testing.T) {
	for _, test := range []struct {
		key       []byte
		expectKey []byte
	}{
		{[]byte("OpvoteBP"), append([]byte(systemVoteSort), []byte("OpvoteBP")...)},
		{[]byte("OpvoteDAO"), append([]byte(systemVoteSort), []byte("OpvoteDAO")...)},
		{[]byte("Opstake"), append([]byte(systemVoteSort), []byte("Opstake")...)},
		{[]byte("Opunstake"), append([]byte(systemVoteSort), []byte("Opunstake")...)},
	} {
		key := SystemVoteSort(test.key)
		assert.Equal(t, test.expectKey, key, "TestSystemVoteSort(%v)", test.key)
	}
}

func TestSystemVoteTotal(t *testing.T) {
	for _, test := range []struct {
		key       []byte
		expectKey []byte
	}{
		{[]byte("OpvoteBP"), append([]byte(systemVoteTotal), []byte("OpvoteBP")...)},
		{[]byte("OpvoteDAO"), append([]byte(systemVoteTotal), []byte("OpvoteDAO")...)},
		{[]byte("Opstake"), append([]byte(systemVoteTotal), []byte("Opstake")...)},
		{[]byte("Opunstake"), append([]byte(systemVoteTotal), []byte("Opunstake")...)},
	} {
		key := SystemVoteTotal(test.key)
		assert.Equal(t, test.expectKey, key, "TestSystemVoteTotal(%v)", test.key)
	}
}

func TestSystemVpr(t *testing.T) {
	for _, test := range []struct {
		i         uint8
		expectKey []byte
	}{
		{0, append([]byte(systemVpr), '0')},
		{1, append([]byte(systemVpr), '1')},
		{255, append([]byte(systemVpr), '2', '5', '5')},
	} {
		key := SystemVpr(test.i)
		assert.Equal(t, test.expectKey, key, "TestSystemVpr(%v)", test.i)
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
