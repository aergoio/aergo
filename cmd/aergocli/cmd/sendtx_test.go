package cmd

import (
	"encoding/json"
	"testing"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/types"
	"github.com/aergoio/aergo/v2/types/jsonrpc"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSendTxWithMock(t *testing.T) {
	mock := initMock(t)
	defer deinitMock()

	testTxHashString := "BdAoKcLSsrscjdpTPGe9DoFsz4mP9ezbc4Dk5fuBTT4e"
	testTxHash, _ := base58.Decode(testTxHashString)

	mock.EXPECT().SendTX(
		gomock.Any(), // expect any value for first parameter
		gomock.Any(), // expect any value for second parameter
	).Return(
		&types.CommitResult{
			Hash:   testTxHash,
			Error:  types.CommitStatus_TX_OK,
			Detail: "",
		},
		nil,
	).MaxTimes(1)

	output, err := executeCommand(rootCmd, "sendtx", "--from", "AmNL5neKQS2ZwRuBeqfcfHMLg3aSmGoefEh5bW8ozWxrtmxaGHZ3", "--to", "AmNfacq5A3orqn3MhgkHSncufXEP8gVJgqDy8jTgBphXQeuuaHHF", "--amount", "1000", "--keystore", "")
	assert.NoError(t, err, "should no error")
	t.Log(output)
	out := &jsonrpc.InOutCommitResult{}
	err = json.Unmarshal([]byte(output), out)
	assert.Equal(t, testTxHashString, out.Hash)
}

func TestSendTxFromToValidation(t *testing.T) {
	_, err := executeCommand(rootCmd, "sendtx", "--from", "InvalidKQS2ZwRuBeqfcfHMLg3aSmGoefEh5bW8ozWxrtmxaGHZ3", "--to", "AmNfacq5A3orqn3MhgkHSncufXEP8gVJgqDy8jTgBphXQeuuaHHF", "--amount", "1000", "--keystore", "")
	assert.Error(t, err, "should error when wrong --from flag")

	_, err = executeCommand(rootCmd, "sendtx", "--from", "AmNL5neKQS2ZwRuBeqfcfHMLg3aSmGoefEh5bW8ozWxrtmxaGHZ3", "--to", "AmNfacq5A3orqn3MhgkHSncufXEP8gVJgqDy8jTgBphXQInvalid", "--amount", "1000", "--keystore", "")
	assert.Error(t, err, "should error when wrong --to flag")
}
