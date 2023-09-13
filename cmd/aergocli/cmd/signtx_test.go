package cmd

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/cmd/aergocli/util/encoding/json"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
	"github.com/stretchr/testify/assert"
)

func TestSignWithKey(t *testing.T) {
	const testAddr = "AmNBjtxomk1uaFrwj8rEKVxYEJ1nzy73dsGrNZzkqs88q8Mkv8GN"
	const signLength = 71
	output, err := executeCommand(rootCmd, "signtx", "--key", "12345678", "--jsontx", "{}")
	assert.NoError(t, err, "should be success")

	outputline := strings.Split(output, "\n")

	addr, err := types.DecodeAddress(outputline[0])
	assert.NoError(t, err, "should be success")
	assert.Equalf(t, types.AddressLength, len(addr), "wrong address length value = %s", output)

	ouputjson := strings.Join(outputline[1:], "")
	var tx util.InOutTx
	err = json.Unmarshal([]byte(ouputjson), &tx)
	assert.NoError(t, err, "should be success")

	sign, err := base58.Decode(tx.Body.Sign)
	assert.NoError(t, err, "should be success")
	assert.Equalf(t, len(sign), signLength, "wrong sign length value = %s", tx.Body.Sign)
}

func TestSignWithPath(t *testing.T) {
	const testDir = "signwithpathtest"

	addr, err := executeCommand(rootCmd, "account", "new", "--password", "1", "--keystore", testDir)
	assert.NoError(t, err, "should be success")

	re := regexp.MustCompile(`\r?\n`)
	addr = re.ReplaceAllString(addr, "")

	rawaddr, err := types.DecodeAddress(addr)
	assert.NoError(t, err, "should be success")
	assert.Equalf(t, types.AddressLength, len(rawaddr), "wrong address length from %s", addr)

	_, err = executeCommand(rootCmd, "signtx", "--keystore", testDir, "--jsontx", "{}", "--password", "1", "--address", addr)
	assert.NoError(t, err, "should be success")

	os.RemoveAll(testDir)
}
// aergocli signtx --from AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R --to AmM1M4jAxeTxfh9CHUmKXJwTLxGJ6Qe5urEeFXqbqkKnZ73Uvx4y --amount 1aergo --password bmttest
// aergocli signtx --keystore . --password bmttest --address AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R --jsontx "{\"from\": \"AmNdzAYv3dYKFtPRgfUMGppGwBJS2JvZLRTF9gRruF49vppEepgj\", \"to\": \"AmM1M4jAxeTxfh9CHUmKXJwTLxGJ6Qe5urEeFXqbqkKnZ73Uvx4y\", \"amount\", \"1aergo\"}"
aergocli signtx --keystore . --password bmttest --address AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R --jsontx "{\"Account\": \"AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R\", \"Recipient\": \"AmM1M4jAxeTxfh9CHUmKXJwTLxGJ6Qe5urEeFXqbqkKnZ73Uvx4y\", \"Amount\": \"1aergo\", \"Type\":4, \"Nonce\":0, \"GasLimit\":0, \"chainIdHash\": \"BNVSYKKqSR78hTSrxkaroFK2S3mgsoocpWg9HYtb1Zhn\"}"

{"from": "AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R", "to": "AmM1M4jAxeTxfh9CHUmKXJwTLxGJ6Qe5urEeFXqbqkKnZ73Uvx4y", "amount": "1aergo"}
"{\"Account\": \"AmPpcKvToDCUkhT1FJjdbNvR4kNDhLFJGHkSqfjWe3QmHm96qv4R\", \"Recipient\": \"AmM1M4jAxeTxfh9CHUmKXJwTLxGJ6Qe5urEeFXqbqkKnZ73Uvx4y\", \"Amount\": \"1aergo\", \"Type\":4, \"Nonce\":1, \"GasLimit\":0, \"chainIdHash\": \"BNVSYKKqSR78hTSrxkaroFK2S3mgsoocpWg9HYtb1Zhn\"}"
