package enterprise

import (
	"encoding/pem"
	"strings"
	"testing"

	"github.com/aergoio/aergo/consensus"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

type MockCCC struct {
	consensus.ChainConsensusCluster
}

func (ccc *MockCCC) MakeConfChangeProposal(req *types.MembershipChange) (*consensus.ConfChangePropose, error) {
	return &consensus.ConfChangePropose{}, nil
}

var (
	ccc = &MockCCC{}
)

func TestBasicFailEnterprise(t *testing.T) {
	scs, sender, _ := initTest(t)
	defer deinitTest()

	txIn := &types.Tx{Body: &types.TxBody{}}
	tx := types.NewTransaction(txIn)

	_, err := ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "empty body")
	txIn.Body.Payload = []byte("invalid")
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "invalid body")
	txIn.Body.Payload = []byte("{}")
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "empty json")
	txIn.Body.Payload = []byte(`{"name":"enableConf"}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "empty arg in enable conf")
	txIn.Body.Payload = []byte(`{"name":"setConf"}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "empty arg in set conf")
	txIn.Body.Payload = []byte(`{"name":"enableConf", "args":["raft",true]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "admin is not set when enble conf")
	txIn.Body.Payload = []byte(`{"name":"setConf", "args":["raft","thisisraftid1", "thisisraftid2"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "admin is not set when set conf")
	txIn.Body.Payload = []byte(`{"name":"setAdmin", "args":["AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "invalid arg in set admin")
	txIn.Body.Payload = []byte(`{"name":"setAdmin", "args":[]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "invalid arg in set admin")

	txIn.Body.Payload = []byte(`{"name":"appendAdmin", "args":["AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.NoError(t, err, "set admin")
	txIn.Body.Payload = []byte(`{"name":"appendAdmin", "args":["AmLqZFnwMLqLg5fMshgzmfvwBP8uiYGgfV3tBZAm36Tv7jFYcs4f"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.NoError(t, err, "set admin")
	txIn.Body.Payload = []byte(`{"name":"appendAdmin", "args":["AmLqZFnwMLqLg5fMshgzmfvwBP8uiYGgfV3tBZAm36Tv7jFYcs4f"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "set same admin permission")

	txIn.Body.Payload = []byte(`{"name":"appendConf", "args":["admins", "AmLqZFnwMLqLg5fMshgzmfvwBP8uiYGgfV3tBZAm36Tv7jFYcs4f"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "not allowed key")

	txIn.Body.Payload = []byte(`{"name":"appendConf", "args":["abc", "AmLqZ\FnwMLqLg5fMshgzmfvwBP8uiYGgfV3tBZAm36Tv7jFYcs4f"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "not allowed char")

	txIn.Body.Payload = []byte(`{"name":"setConf", "args":["p2pwhite","16Uiu2HAmAokYAtLbZxJAPRgp2jCc4bD35cJD921trqUANh59Rc4n", "16Uiu2HAmAokYAtLbZxJAPRgp2jCc4bD35cJD921trqUANh59Rc4n", "16Uiu2HAmGiJ2QgVAWHMUtzLKKNM5eFUJ3Ds3FN7nYJq1mHN5ZPj9"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "duplicate arguments")

	txIn.Body.Payload = []byte(`{"name":"setConf", "args":["p2pwhite","16Uiu2HAmAokYAtLbZxJAPRgp2jCc4bD35cJD921trqUANh59Rc4n", "16Uiu2HAm4xYtGsqk7WGKUxr8prfVpJ25hD23AQ3Be6anEL9Kxkgw", "16Uiu2HAmGiJ2QgVAWHMUtzLKKNM5eFUJ3Ds3FN7nYJq1mHN5ZPj9"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.NoError(t, err, "set conf")

	txIn.Body.Payload = []byte(`{"name":"appendConf", "args":["p2pwhite","16Uiu2HAmAokYAtLbZxJAPRgp2jCc4bD35cJD921trqUANh59Rc4n"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "duplicated set conf")

	txIn.Body.Payload = []byte(`{"name":"setConf", "args":["rpcpermissions","abc:R", "bcd:S", "cde:C"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.NoError(t, err, "set conf")

	txIn.Body.Payload = []byte(`{"name":"enableConf", "args":["rpcpermissions",true]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "enable conf")

	txIn.Body.Payload = []byte(`{"name":"appendConf", "args":["rpcpermissions","abc:WR"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.NoError(t, err, "append conf")

	txIn.Body.Payload = []byte(`{"name":"enableConf", "args":["rpcpermissions",true]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.NoError(t, err, "enable conf")

	txIn.Body.Payload = []byte(`{"name":"removeConf", "args":["rpcpermissions","abc:WR"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "remove conf")
}

func TestBasicEnterprise(t *testing.T) {
	scs, sender, _ := initTest(t)
	defer deinitTest()

	txIn := &types.Tx{Body: &types.TxBody{}}
	tx := types.NewTransaction(txIn)

	txIn.Body.Payload = []byte(`{"name":"appendAdmin", "args":["AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"]}`)
	event, err := ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.NoError(t, err, "add admin")
	txIn.Body.Payload = []byte(`{"name":"appendAdmin", "args":["AmLt7Z3y2XTu7YS8KHNuyKM2QAszpFHSX77FLKEt7FAuRW7GEhj7"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.NoError(t, err, "add admin")
	admins, err := getAdmins(scs)
	assert.NoError(t, err, "remove admin")
	assert.Equal(t, 2, len(admins), "check admin")
	assert.Equal(t, "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4", types.EncodeAddress(admins[0]), "check admin")
	assert.Equal(t, "AmLt7Z3y2XTu7YS8KHNuyKM2QAszpFHSX77FLKEt7FAuRW7GEhj7", types.EncodeAddress(admins[1]), "check admin")
	txIn.Body.Payload = []byte(`{"name":"removeAdmin", "args":["AmLt7Z3y2XTu7YS8KHNuyKM2QAszpFHSX77FLKEt7FAuRW7GEhj7"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.NoError(t, err, "remove admin")
	admins, err = getAdmins(scs)
	assert.NoError(t, err, "remove admin")
	assert.Equal(t, 1, len(admins), "check admin")
	assert.Equal(t, "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4", types.EncodeAddress(admins[0]), "check admin")

	txIn.Body.Payload = []byte(`{"name":"setConf", "args":["p2pwhite","16Uiu2HAmAokYAtLbZxJAPRgp2jCc4bD35cJD921trqUANh59Rc4n", "16Uiu2HAm4xYtGsqk7WGKUxr8prfVpJ25hD23AQ3Be6anEL9Kxkgw", "16Uiu2HAmGiJ2QgVAWHMUtzLKKNM5eFUJ3Ds3FN7nYJq1mHN5ZPj9"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.NoError(t, err, "set conf")
	conf, err := getConf(scs, []byte("P2PWhite")) //key is ignore case
	assert.Equal(t, false, conf.On, "conf on")
	assert.Equal(t, 3, len(conf.Values), "conf values length")
	assert.Equal(t, "16Uiu2HAmAokYAtLbZxJAPRgp2jCc4bD35cJD921trqUANh59Rc4n", conf.Values[0], "conf value 0")
	assert.Equal(t, "16Uiu2HAm4xYtGsqk7WGKUxr8prfVpJ25hD23AQ3Be6anEL9Kxkgw", conf.Values[1], "conf value 1")
	assert.Equal(t, "16Uiu2HAmGiJ2QgVAWHMUtzLKKNM5eFUJ3Ds3FN7nYJq1mHN5ZPj9", conf.Values[2], "conf value 2")

	txIn.Body.Payload = []byte(`{"name":"appendConf", "args":["p2pwhite","16Uiu2HAmAAtqye6QQbeG9EZnrWJbGK8Xw74cZxpnGGEAZAB3zJ8B"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.NoError(t, err, "set conf")
	conf, err = getConf(scs, []byte("p2pwhite"))
	assert.Equal(t, false, conf.On, "conf on")
	assert.Equal(t, 4, len(conf.Values), "conf values length")
	assert.Equal(t, "16Uiu2HAmAAtqye6QQbeG9EZnrWJbGK8Xw74cZxpnGGEAZAB3zJ8B", conf.Values[3], "conf value 3")

	txIn.Body.Payload = []byte(`{"name":"enableConf", "args":["p2pwhite",true]}`)
	event, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	//t.Log(event)
	assert.NoError(t, err, "enable conf")
	conf, err = getConf(scs, []byte("p2pwhite"))
	assert.Equal(t, true, conf.On, "conf on")

	block, _ := pem.Decode([]byte(testCert))
	assert.NotNil(t, block, "parse value 0")
	cert := types.EncodeB64(block.Bytes)
	txIn.Body.Payload = []byte(`{"name":"appendConf", "args":["rpcpermissions","` + cert + `:RWCS"]}`)
	event, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.NoError(t, err, "add conf")
	conf, err = getConf(scs, []byte("rpcpermissions"))
	assert.Equal(t, false, conf.On, "conf on")
	assert.Equal(t, 1, len(conf.Values), "conf values length")
	assert.Equal(t, cert, strings.Split(conf.Values[0], ":")[0], "conf value 0")
	assert.Equal(t, "RWCS", strings.Split(conf.Values[0], ":")[1], "conf value 1")

	txIn.Body.Payload = []byte(`{"name":"appendConf", "args":["rpcpermissions","` + strings.Split(conf.Values[0], ":")[0] + `:RWCS"]}`)
	event, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.Error(t, err, "dup add conf")
	t.Log(event)

	txIn.Body.Payload = []byte(`{"name":"enableConf", "args":["p2pwhite",false]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender)
	assert.NoError(t, err, "enable conf")
	conf, err = getConf(scs, []byte("p2pwhite"))
	assert.Equal(t, false, conf.On, "conf on")

	bs := state.NewBlockState(&state.StateDB{})
	txIn.Body.Payload = []byte(`{"name":"changeCluster", "args":[{"command" : "add", "name": "aergonew", "url": "http://127.0.0.1:13000", "peerid":"16Uiu2HAmAAtqye6QQbeG9EZnrWJbGK8Xw74cZxpnGGEAZAB3zJ8B"}]}`)
	_, err = ExecuteEnterpriseTx(bs, ccc, scs, tx, sender)
	assert.NoError(t, err)
	assert.NotNil(t, bs.CCProposal)

	bs = state.NewBlockState(&state.StateDB{})
	txIn.Body.Payload = []byte(`{"name":"changeCluster", "args":[{"command" : "remove", "id": "1234"}]}`)
	_, err = ExecuteEnterpriseTx(bs, ccc, scs, tx, sender)
	assert.NoError(t, err)
	assert.NotNil(t, bs.CCProposal)

	bs = state.NewBlockState(&state.StateDB{})
	txIn.Body.Payload = []byte(`{"name":"changeCluster", "args":[{"command" : "nocmd", "name": "aergonew", "url": "http://127.0.0.1:13000", "peerID":"16Uiu2HAmAAtqye6QQbeG9EZnrWJbGK8Xw74cZxpnGGEAZAB3zJ8B"}]}`)
	_, err = ExecuteEnterpriseTx(bs, ccc, scs, tx, sender)
	assert.Error(t, err)
	assert.Nil(t, bs.CCProposal)
}
