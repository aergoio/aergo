package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGovernanceTypeTransaction(t *testing.T) {
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
	transaction := NewTransaction(tx)
	err = transaction.Validate()
	assert.EqualError(t, ErrTxHasInvalidHash, err.Error(), "empty hash")

	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate()
	assert.EqualError(t, ErrTxInvalidRecipient, err.Error(), "recipient null")

	transaction.GetTx().GetBody().Recipient = tx.Body.Payload
	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate()
	assert.EqualError(t, ErrTxInvalidRecipient, err.Error(), "wrong recipient case")

	transaction.GetTx().GetBody().Recipient = recipient
	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate()
	assert.EqualError(t, ErrTxInvalidRecipient, err.Error(), "recipient should be aergo.*")

	transaction.GetTx().GetBody().Recipient = []byte(AergoSystem)
	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate()
	assert.EqualError(t, ErrTooSmallAmount, err.Error(), "recipient should be aergo.*")

	transaction.GetTx().GetBody().Amount = StakingMinimum.Bytes()
	transaction.GetTx().Hash = transaction.CalculateTxHash()
	err = transaction.Validate()
	assert.NoError(t, err, "should success")
}
