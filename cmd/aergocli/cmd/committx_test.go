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

func TestCommitTxWithMock(t *testing.T) {
	mock := initMock(t)
	defer deinitMock()

	testTxHashString := "HB44gJvHhVoEfgiGq3VZmV9VUXfBXhHjcEvroBMkJGnY"
	testTxHash, _ := base58.Decode(testTxHashString)
	output, err := executeCommand(rootCmd, "committx", "--jsontx", "{}")
	assert.Error(t, err, "should occur error when empty json")

	mock.EXPECT().CommitTX(
		gomock.Any(), // expect any value for first parameter
		gomock.Any(), // expect any value for second parameter
	).Return(
		&types.CommitResultList{Results: []*types.CommitResult{
			{
				Hash:   nil,
				Error:  types.CommitStatus_TX_INVALID_FORMAT,
				Detail: "tx invalid format",
			},
		}},
		nil,
	).MaxTimes(1)

	output, err = executeCommand(rootCmd, "committx", "--jsontx", "{\"Body\":{}}")
	out := &jsonrpc.InOutCommitResultList{}
	err = json.Unmarshal([]byte(output), out)
	assert.NoError(t, err, "commit output is invalid")
	assert.Equal(t, "tx invalid format", out.Results[0].Detail)

	mock.EXPECT().CommitTX(
		gomock.Any(), // expect any value for first parameter
		gomock.Any(), // expect any value for second parameter
	).Return(
		&types.CommitResultList{Results: []*types.CommitResult{
			{
				Hash:   testTxHash,
				Error:  types.CommitStatus_TX_OK,
				Detail: "",
			},
		}},
		nil,
	).MaxTimes(1)

	output, err = executeCommand(rootCmd, "committx", "--jsontx", "{ \"Hash\": \"HB44gJvHhVoEfgiGq3VZmV9VUXfBXhHjcEvroBMkJGnY\", \"Body\": {\"Nonce\": 2, \"Account\": \"AmNBZ8WQKP8DbuP9Q9W9vGFhiT8vQNcuSZ2SbBbVvbJWGV3Wh1mn\", \"Recipient\": \"AmLnVfGwq49etaa7dnzfGJTbaZWV7aVmrxFes4KmWukXwtooVZPJ\", \"Amount\": \"25000\", \"Payload\": \"\", \"Limit\": 100, \"Price\": \"1\", \"Type\": 0, \"Sign\": \"381yXYxTtq2tRPRQPF7tHH6Cq3y8PvcsFWztPwCRmmYfqnK83Z3a6Yj9fyy8Rpvrrw76Y52SNAP6Th3BYQjX1Bcmf6NQrDHQ\"}}")
	err = json.Unmarshal([]byte(output), out)
	assert.NoError(t, err, "should no error")
	assert.Equal(t, "HB44gJvHhVoEfgiGq3VZmV9VUXfBXhHjcEvroBMkJGnY", out.Results[0].Hash)
}
