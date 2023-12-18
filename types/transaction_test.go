package types

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/stretchr/testify/assert"
)

var (
	TestNormal          = 0
	TestDuplicatePeerID = 1
	TestInvalidPeerID   = 2
	TestInvalidString   = 3
)

func TestGovernanceTypeTransaction(t *testing.T) {
	InitGovernance("dpos", false)

	const testSender = "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"
	const testReceiver = "AmNhXiU3s2BN26v5B5hT2bbEjvSjqyrBY7DGnD9UqVcwkTrDYyJN"
	account, err := DecodeAddress(testSender)
	assert.NoError(t, err, "should success to decode test address")
	recipient, err := DecodeAddress(testReceiver)
	assert.NoError(t, err, "should success to decode test address")
	tx := &Tx{
		Body: &TxBody{
			Account: account,
			Payload: []byte(`{"Name":"v1unstake"}`),
			Type:    TxType_GOVERNANCE,
		},
	}

	chainid := []byte("chainid")
	fakechainid := []byte("fake")
	transaction := NewTransaction(tx)

	transaction.GetBody().ChainIdHash = fakechainid
	err = transaction.Validate(chainid, false)
	assert.EqualError(t, ErrTxInvalidChainIdHash, err.Error(), "invalid chainid")

	transaction.GetBody().ChainIdHash = chainid
	err = transaction.Validate(chainid, false)
	assert.EqualError(t, ErrTxHasInvalidHash, err.Error(), "empty hash")

	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate(chainid, false)
	assert.EqualError(t, ErrTxInvalidRecipient, err.Error(), "recipient null")

	transaction.GetTx().GetBody().Recipient = tx.Body.Payload
	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate(chainid, false)
	assert.EqualError(t, ErrTxInvalidRecipient, err.Error(), "wrong recipient case")

	transaction.GetTx().GetBody().Recipient = recipient
	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate(chainid, false)
	assert.EqualError(t, ErrTxInvalidRecipient, err.Error(), "recipient should be aergo.*")

	transaction.GetTx().GetBody().Recipient = []byte(AergoSystem)
	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate(chainid, false)
	assert.NoError(t, err, "should success")

	transaction.GetTx().GetBody().Amount = StakingMinimum.Bytes()
	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate(chainid, false)
	assert.NoError(t, err, "should success")

	transaction.GetTx().GetBody().Payload = buildVoteBPPayloadEx(2, TestInvalidString)
	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate(chainid, false)
	assert.EqualError(t, err, ErrTxInvalidPayload.Error(), "invalid string")

	transaction.GetTx().GetBody().Payload = buildVoteBPPayloadEx(2, TestInvalidPeerID)
	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate(chainid, false)
	assert.EqualError(t, err, ErrTxInvalidPayload.Error(), "invalid peer id")

	transaction.GetTx().GetBody().Payload = buildVoteBPPayloadEx(2, TestDuplicatePeerID)
	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate(chainid, false)
	assert.EqualError(t, err, ErrTxInvalidPayload.Error(), "dup peer id")

	transaction.GetTx().GetBody().Payload = buildVoteBPPayloadEx(2, TestNormal)
	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate(chainid, false)
	t.Log(string(transaction.GetTx().GetBody().Payload))
	assert.NoError(t, err, "should success")

	transaction.GetTx().GetBody().Recipient = []byte(`aergo.name`)
	transaction.GetTx().GetBody().Payload = []byte(`{"Name":"v1createName", "Args":["1"]}`)
	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate(chainid, false)
	assert.Error(t, err, "invalid name length in create")

	transaction.GetTx().GetBody().Payload = []byte(`{"Name":"v1updateName", "Args":["1234567890","AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"]}`)
	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate(chainid, false)
	assert.Error(t, err, "invalid name length in update")

	transaction.GetTx().GetBody().Payload = []byte(`{"Name":"v1voteDAO", "Args":["BPCOUNT","3", "3"]}`)
	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate(chainid, false)
	assert.Error(t, err, "duplicate arguments")

}

func buildVoteBPPayloadEx(count int, err int) []byte {
	var ci CallInfo
	ci.Name = OpvoteBP.Cmd()
	_, pub, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	peerid, _ := IDFromPublicKey(pub)
	for i := 0; i < count; i++ {
		if err == TestDuplicatePeerID {
			ci.Args = append(ci.Args, IDB58Encode(peerid))
		} else if err == TestInvalidString {
			ci.Args = append(ci.Args, (i + 1))
		} else if err == TestInvalidPeerID {
			ci.Args = append(ci.Args, strconv.Itoa(i+1))
		} else {
			_, pub, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
			peerid, _ := IDFromPublicKey(pub)
			ci.Args = append(ci.Args, IDB58Encode(peerid))
		}
	}
	payload, _ := json.Marshal(ci)
	return payload
}
