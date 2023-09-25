package name

import (
	"testing"

	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestExcuteNameTx(t *testing.T) {
	initTest(t)
	defer deinitTest()
	txBody := &types.TxBody{}
	txBody.Account = types.ToAddress("AmMXVdJ8DnEFysN58cox9RADC74dF1CLrQimKCMdB4XXMkJeuQgL")
	txBody.Recipient = []byte(types.AergoName)
	txBody.Amount = types.NewAmount(1, types.Aergo).Bytes() //default value of name price, 1 Aergo

	name := "AB1234567890"
	txBody.Payload = buildNamePayload(name, types.NameCreate, "")

	sender, _ := sdb.GetStateDB().GetAccountStateV(txBody.Account)
	sender.AddBalance(types.MaxAER)
	receiver, _ := sdb.GetStateDB().GetAccountStateV(txBody.Recipient)
	bs := sdb.NewBlockState(sdb.GetRoot())
	scs := openContractState(t, bs)

	blockInfo := &types.BlockHeaderInfo{No: uint64(0), ForkVersion: 0}
	event, err := ExecuteNameTx(bs, scs, txBody, sender, receiver, blockInfo)
	assert.NoError(t, err, "execute name tx")
	assert.Equal(t, "create name", event[0].EventName, "event name")
	assert.Equal(t, "{\"name\":\"AB1234567890\"}", event[0].JsonArgs, "event args")
	//race
	tmpAddress := "AmNHAxiGbZJjKjdGGNj2NBoAXGwdzX9Bg59eqbek9n49JpiaZ3As"
	txBody.Account = types.ToAddress(tmpAddress)
	_, err = ExecuteNameTx(bs, scs, txBody, sender, receiver, blockInfo)
	assert.Error(t, err, "race execute name tx")

	txBody.Account = types.ToAddress("AmMXVdJ8DnEFysN58cox9RADC74dF1CLrQimKCMdB4XXMkJeuQgL")
	commitContractState(t, bs, scs)
	scs = openContractState(t, bs)

	ret := GetAddress(scs, []byte(name))
	assert.Equal(t, txBody.Account, ret, "pubkey address")
	ret = GetOwner(scs, []byte(name))
	assert.Equal(t, txBody.Account, ret, "pubkey owner")

	_, err = ExecuteNameTx(bs, scs, txBody, sender, receiver, blockInfo)
	assert.Error(t, err, "execute name tx")

	buyer := "AmMSMkVHQ6qRVA7G7rqwjvv2NBwB48tTekJ2jFMrjfZrsofePgay"
	txBody.Payload = buildNamePayload(name, types.NameUpdate, buyer)
	blockInfo.No++
	event, err = ExecuteNameTx(bs, scs, txBody, sender, receiver, blockInfo)
	assert.NoError(t, err, "execute to update name")
	assert.Equal(t, "update name", event[0].EventName, "event name")
	assert.Equal(t, "{\"name\":\"AB1234567890\",\"to\":\"AmMSMkVHQ6qRVA7G7rqwjvv2NBwB48tTekJ2jFMrjfZrsofePgay\"}", event[0].JsonArgs, "event args")

	commitContractState(t, bs, scs)
	scs = openContractState(t, bs)

	ret = GetAddress(scs, []byte(name))
	assert.Equal(t, buyer, types.EncodeAddress(ret), "pubkey address")
	ret = GetOwner(scs, []byte(name))
	assert.Equal(t, buyer, types.EncodeAddress(ret), "pubkey owner")

	//invalid case
	blockInfo.No++
	_, err = ExecuteNameTx(bs, scs, txBody, sender, receiver, blockInfo)
	assert.Error(t, err, "execute invalid updating name")

	txBody.Payload = txBody.Payload[1:]
	_, err = ExecuteNameTx(bs, scs, txBody, sender, receiver, blockInfo)
	assert.Error(t, err, "execute invalid payload")

	blockInfo.No++
	blockInfo.ForkVersion = 2
	name2 := "1234567890V2"
	txBody.Payload = buildNamePayload(name2, types.NameCreate, "")
	event, err = ExecuteNameTx(bs, scs, txBody, sender, receiver, blockInfo)
	assert.NoError(t, err, "execute name tx")
	assert.Equal(t, "create name", event[0].EventName, "event name")
	assert.Equal(t, "[\"1234567890V2\"]", event[0].JsonArgs, "event args")

	commitContractState(t, bs, scs)
	scs = openContractState(t, bs)
	blockInfo.No++
	txBody.Payload = buildNamePayload(name2, types.NameUpdate, buyer)
	event, err = ExecuteNameTx(bs, scs, txBody, sender, receiver, blockInfo)
	assert.NoError(t, err, "execute to update name")
	assert.Equal(t, "update name", event[0].EventName, "event name")
	assert.Equal(t, "[\"1234567890V2\",\"AmMSMkVHQ6qRVA7G7rqwjvv2NBwB48tTekJ2jFMrjfZrsofePgay\"]", event[0].JsonArgs, "event args")
}

func TestExcuteFailNameTx(t *testing.T) {
	initTest(t)
	defer deinitTest()
	txBody := &types.TxBody{}

	txBody.Account = types.ToAddress("AmMXVdJ8DnEFysN58cox9RADC74dF1CLrQimKCMdB4XXMkJeuQgL")
	txBody.Recipient = []byte(types.AergoName)

	name := "AB1234567890"
	txBody.Payload = buildNamePayload(name, types.NameCreate+"Broken", "")

	sender, _ := sdb.GetStateDB().GetAccountStateV(txBody.Account)
	receiver, _ := sdb.GetStateDB().GetAccountStateV(txBody.Recipient)
	bs := sdb.NewBlockState(sdb.GetRoot())
	scs := openContractState(t, bs)
	blockInfo := &types.BlockHeaderInfo{No: uint64(0), ForkVersion: 0}
	_, err := ExecuteNameTx(bs, scs, txBody, sender, receiver, blockInfo)
	assert.Error(t, err, "execute name tx")
}

func openContractState(t *testing.T, bs *state.BlockState) *state.ContractState {
	nameContractID := types.ToAccountID([]byte(types.AergoName))
	nameContract, err := bs.GetAccountState(nameContractID)
	assert.NoError(t, err, "could not get account state")
	scs, err := bs.OpenContractState(nameContractID, nameContract)
	assert.NoError(t, err, "could not open contract state")
	return scs
}

func openSystemContractState(t *testing.T, bs *state.BlockState) *state.ContractState {
	systemContractID := types.ToAccountID([]byte(types.AergoSystem))
	systemContract, err := bs.GetAccountState(systemContractID)
	assert.NoError(t, err, "could not get account state")
	scs, err := bs.OpenContractState(systemContractID, systemContract)
	assert.NoError(t, err, "could not open contract state")
	return scs
}

func commitContractState(t *testing.T, bs *state.BlockState, scs *state.ContractState) {
	bs.StageContractState(scs)
	bs.Update()
	bs.Commit()
	sdb.UpdateRoot(bs)
}

func nextBlockContractState(t *testing.T, bs *state.BlockState, scs *state.ContractState) *state.ContractState {
	commitContractState(t, bs, scs)
	return openContractState(t, bs)
}
