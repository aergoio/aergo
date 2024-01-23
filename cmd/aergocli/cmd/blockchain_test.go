package cmd

import (
	"encoding/json"
	"testing"

	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/aergoio/aergo/v2/internal/enc/hex"
	"github.com/aergoio/aergo/v2/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBlockchainWithMock(t *testing.T) {
	mock := initMock(t)
	defer deinitMock()

	testBlockHashString := "56Qy6MQei9KM13rqEq1jiJ7Da21Kcq9KdmYWcnPLtxS3"
	testBlockHash, _ := base58.Decode(testBlockHashString)

	mock.EXPECT().Blockchain(
		gomock.Any(), // expect any value for first parameter
		gomock.Any(), // expect any value for second parameter
	).Return(
		&types.BlockchainStatus{BestBlockHash: testBlockHash, BestHeight: 1, ConsensusInfo: ""},
		nil,
	).MaxTimes(3)

	output, err := executeCommand(rootCmd, "blockchain")
	assert.NoError(t, err, "should be success")
	t.Log(output)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, testBlockHashString, result["hash"])
	assert.Equal(t, float64(1), result["height"])

	output, err = executeCommand(rootCmd, "blockchain", "trashargs")
	assert.NoError(t, err, "should be success")

	output, err = executeCommand(rootCmd, "blockchain", "--hex")
	assert.NoError(t, err, "should be success")
	t.Log(output)

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatal(err)
	}
	testBlockHashByte, _ := base58.Decode(testBlockHashString)
	assert.Equal(t, hex.Encode(testBlockHashByte), result["hash"])
	assert.Equal(t, float64(1), result["height"])
}
