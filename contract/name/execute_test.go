package name

import (
	"testing"

	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

func TestExcuteNameTx(t *testing.T) {
	initTest(t)
	defer deinitTest()
	txBody := &types.TxBody{}

	txBody.Account = types.ToAddress("AmMXVdJ8DnEFysN58cox9RADC74dF1CLrQimKCMdB4XXMkJeuQgL")
	txBody.Recipient = []byte(types.AergoName)

	name := "AB1234567890"
	txBody.Payload = buildNamePayload(name, 'c', nil)

	bs := sdb.NewBlockState(sdb.GetRoot())
	scs := openContractState(t, bs)

	err := ExecuteNameTx(scs, txBody)
	assert.NoError(t, err, "execute name tx")

	commitContractState(t, bs, scs)

	scs = openContractState(t, bs)
	ret := GetAddress(scs, []byte(name))
	assert.Equal(t, txBody.Account, ret, "pubkey address")

}

func openContractState(t *testing.T, bs *state.BlockState) *state.ContractState {
	nameContractID := types.ToAccountID([]byte(types.AergoName))
	nameContract, err := bs.GetAccountState(nameContractID)
	assert.NoError(t, err, "could not get account state")
	scs, err := bs.OpenContractState(nameContractID, nameContract)
	assert.NoError(t, err, "could not open contract state")
	return scs
}
func commitContractState(t *testing.T, bs *state.BlockState, scs *state.ContractState) {
	bs.StageContractState(scs)
	bs.Update()
	bs.Commit()
}
func nextBlockContractState(t *testing.T, bs *state.BlockState, scs *state.ContractState) *state.ContractState {
	commitContractState(t, bs, scs)
	return openContractState(t, bs)
}
