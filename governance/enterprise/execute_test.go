package enterprise

import (
	"encoding/pem"
	"strings"
	"testing"

	"github.com/aergoio/aergo/v2/consensus"
	"github.com/aergoio/aergo/v2/state"
	"github.com/aergoio/aergo/v2/types"
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
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	tx := &types.TxBody{}
	testBlockNo := types.BlockNo(1)

	_, err := ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "empty body")
	tx.Payload = []byte("invalid")
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "invalid body")
	tx.Payload = []byte("{}")
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "empty json")
	tx.Payload = []byte(`{"name":"enableConf"}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "empty arg in enable conf")
	tx.Payload = []byte(`{"name":"setConf"}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "empty arg in set conf")
	tx.Payload = []byte(`{"name":"enableConf", "args":["raft",true]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "admin is not set when enble conf")
	tx.Payload = []byte(`{"name":"setConf", "args":["raft","thisisraftid1", "thisisraftid2"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "admin is not set when set conf")
	tx.Payload = []byte(`{"name":"setAdmin", "args":["AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "invalid arg in set admin")
	tx.Payload = []byte(`{"name":"setAdmin", "args":[]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "invalid arg in set admin")

	tx.Payload = []byte(`{"name":"appendAdmin", "args":["AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "set admin")
	tx.Payload = []byte(`{"name":"appendAdmin", "args":["AmLqZFnwMLqLg5fMshgzmfvwBP8uiYGgfV3tBZAm36Tv7jFYcs4f"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "set admin")
	tx.Payload = []byte(`{"name":"appendAdmin", "args":["AmLqZFnwMLqLg5fMshgzmfvwBP8uiYGgfV3tBZAm36Tv7jFYcs4f"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "set same admin permission")

	tx.Payload = []byte(`{"name":"appendConf", "args":["admins", "AmLqZFnwMLqLg5fMshgzmfvwBP8uiYGgfV3tBZAm36Tv7jFYcs4f"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "not allowed key")

	tx.Payload = []byte(`{"name":"appendConf", "args":["rpcpermissions", "AmLqZ\FnwMLqLg5fMshgzmfvwBP8uiYGgfV3tBZAm36Tv7jFYcs4f"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "not allowed char")

	tx.Payload = []byte(`{"name":"setConf", "args":["p2pwhite","{\"peerid\":\"16Uiu2HAmAokYAtLbZxJAPRgp2jCc4bD35cJD921trqUANh59Rc4n\"}", "{\"peerid\":\"16Uiu2HAmAokYAtLbZxJAPRgp2jCc4bD35cJD921trqUANh59Rc4n\"}", "{\"peerid\":\"16Uiu2HAmGiJ2QgVAWHMUtzLKKNM5eFUJ3Ds3FN7nYJq1mHN5ZPj9\"}"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "duplicate arguments")

	tx.Payload = []byte(`{"name":"setConf", "args":["p2pwhite","{\"peerid\":\"16Uiu2HAmAokYAtLbZxJAPRgp2jCc4bD35cJD921trqUANh59Rc4n\"}", "{\"peerid\":\"16Uiu2HAm4xYtGsqk7WGKUxr8prfVpJ25hD23AQ3Be6anEL9Kxkgw\"}", "{\"peerid\":\"16Uiu2HAmGiJ2QgVAWHMUtzLKKNM5eFUJ3Ds3FN7nYJq1mHN5ZPj9\"}"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "set conf")

	tx.Payload = []byte(`{"name":"appendConf", "args":["p2pwhite","16Uiu2HAmAokYAtLbZxJAPRgp2jCc4bD35cJD921trqUANh59Rc4n"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "duplicated set conf")

	tx.Payload = []byte(`{"name":"setConf", "args":["rpcpermissions","dGVzdAo=:R", "dGVzdDIK:S", "dGVzdDMK:C"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "set conf")

	tx.Payload = []byte(`{"name":"enableConf", "args":["rpcpermissions",true]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "enable conf")

	tx.Payload = []byte(`{"name":"appendConf", "args":["rpcpermissions","dGVzdAo=:WR"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "append conf")

	tx.Payload = []byte(`{"name":"enableConf", "args":["rpcpermissions",true]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "enable conf")

	tx.Payload = []byte(`{"name":"removeConf", "args":["rpcpermissions","dGVzdAo=:WR"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "remove conf")
}

func TestBasicEnterprise(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	tx := &types.TxBody{}
	testBlockNo := types.BlockNo(1)

	tx.Payload = []byte(`{"name":"appendAdmin", "args":["AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"]}`)
	event, err := ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "add admin")
	assert.Equal(t, "Append ADMIN", event[0].EventName, "append admin event")
	assert.Equal(t, "\"AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4\"", event[0].JsonArgs, "append admin event")
	tx.Payload = []byte(`{"name":"appendAdmin", "args":["AmLt7Z3y2XTu7YS8KHNuyKM2QAszpFHSX77FLKEt7FAuRW7GEhj7"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "add admin")
	admins, err := getAdmins(scs)
	assert.NoError(t, err, "get after appending admin")
	assert.Equal(t, 2, len(admins), "check admin")
	assert.Equal(t, "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4", types.EncodeAddress(admins[0]), "check admin")
	assert.Equal(t, "AmLt7Z3y2XTu7YS8KHNuyKM2QAszpFHSX77FLKEt7FAuRW7GEhj7", types.EncodeAddress(admins[1]), "check admin")

	tx.Payload = []byte(`{"name":"removeAdmin", "args":["AmLt7Z3y2XTu7YS8KHNuyKM2QAszpFHSX77FLKEt7FAuRW7GEhj7"]}`)
	event, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "remove admin")
	assert.Equal(t, "Remove ADMIN", event[0].EventName, "append admin event")
	assert.Equal(t, "\"AmLt7Z3y2XTu7YS8KHNuyKM2QAszpFHSX77FLKEt7FAuRW7GEhj7\"", event[0].JsonArgs, "append admin event")
	admins, err = getAdmins(scs)
	assert.NoError(t, err, "get after removing admin")
	assert.Equal(t, 1, len(admins), "check admin")
	assert.Equal(t, "AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4", types.EncodeAddress(admins[0]), "check admin")

	tx.Payload = []byte(`{"name":"setConf", "args":["p2pwhite","{\"peerid\":\"16Uiu2HAmAokYAtLbZxJAPRgp2jCc4bD35cJD921trqUANh59Rc4n\"}", "{\"peerid\":\"16Uiu2HAm4xYtGsqk7WGKUxr8prfVpJ25hD23AQ3Be6anEL9Kxkgw\"}", "{\"peerid\":\"16Uiu2HAmGiJ2QgVAWHMUtzLKKNM5eFUJ3Ds3FN7nYJq1mHN5ZPj9\"}"]}`)
	event, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "set conf")
	assert.Equal(t, "Set P2PWHITE", event[0].EventName, "append admin event")
	conf, err := getConf(scs, []byte("P2PWhite")) //key is ignore case
	assert.Equal(t, false, conf.On, "conf on")
	assert.Equal(t, 3, len(conf.Values), "conf values length")
	assert.Equal(t, `{"peerid":"16Uiu2HAmAokYAtLbZxJAPRgp2jCc4bD35cJD921trqUANh59Rc4n"}`, conf.Values[0], "conf value 0")
	assert.Equal(t, `{"peerid":"16Uiu2HAm4xYtGsqk7WGKUxr8prfVpJ25hD23AQ3Be6anEL9Kxkgw"}`, conf.Values[1], "conf value 1")
	assert.Equal(t, `{"peerid":"16Uiu2HAmGiJ2QgVAWHMUtzLKKNM5eFUJ3Ds3FN7nYJq1mHN5ZPj9"}`, conf.Values[2], "conf value 2")

	tx.Payload = []byte(`{"name":"appendConf", "args":["p2pwhite","{\"peerid\":\"16Uiu2HAmAAtqye6QQbeG9EZnrWJbGK8Xw74cZxpnGGEAZAB3zJ8B\"}"]}`)
	event, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	t.Log(event)
	assert.NoError(t, err, "set conf")
	assert.Equal(t, "Set P2PWHITE", event[0].EventName, "append admin event")
	conf, err = getConf(scs, []byte("p2pwhite"))
	assert.Equal(t, false, conf.On, "conf on")
	assert.Equal(t, 4, len(conf.Values), "conf values length")
	assert.Equal(t, `{"peerid":"16Uiu2HAmAAtqye6QQbeG9EZnrWJbGK8Xw74cZxpnGGEAZAB3zJ8B"}`, conf.Values[3], "conf value 3")

	tx.Payload = []byte(`{"name":"enableConf", "args":["p2pwhite",true]}`)
	event, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	t.Log(event)
	assert.NoError(t, err, "enable conf")
	conf, err = getConf(scs, []byte("p2pwhite"))
	assert.Equal(t, true, conf.On, "conf on")

	block, _ := pem.Decode([]byte(testCert))
	assert.NotNil(t, block, "parse value 0")
	cert := types.EncodeB64(block.Bytes)
	tx.Payload = []byte(`{"name":"appendConf", "args":["rpcpermissions","` + cert + `:RWCS"]}`)
	event, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "add conf")
	conf, err = getConf(scs, []byte("rpcpermissions"))
	assert.Equal(t, false, conf.On, "conf on")
	assert.Equal(t, 1, len(conf.Values), "conf values length")
	assert.Equal(t, cert, strings.Split(conf.Values[0], ":")[0], "conf value 0")
	assert.Equal(t, "RWCS", strings.Split(conf.Values[0], ":")[1], "conf value 1")

	tx.Payload = []byte(`{"name":"appendConf", "args":["rpcpermissions","` + strings.Split(conf.Values[0], ":")[0] + `:RWCS"]}`)
	event, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, "dup add conf")
	t.Log(event)

	tx.Payload = []byte(`{"name":"enableConf", "args":["p2pwhite",false]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "enable conf")
	conf, err = getConf(scs, []byte("p2pwhite"))
	assert.Equal(t, false, conf.On, "conf on")
}

func TestEnterpriseChangeCluster(t *testing.T) {
	consensus.SetCurConsensus("raft")

	scs, sender, receiver := initTest(t)
	defer deinitTest()

	tx := &types.TxBody{}
	testBlockNo := types.BlockNo(1)

	tx.Payload = []byte(`{"name":"appendAdmin", "args":["AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"]}`)
	_, err := ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "add admin")

	bs := state.NewBlockState(&state.StateDB{})
	tx.Payload = []byte(`{"name":"changeCluster", "args":[{"command" : "add", "name": "aergonew", "address": "/ip4/127.0.0.1/tcp/11001", "peerid":"16Uiu2HAmAAtqye6QQbeG9EZnrWJbGK8Xw74cZxpnGGEAZAB3zJ8B"}]}`)
	_, err = ExecuteEnterpriseTx(bs, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err)
	assert.NotNil(t, bs.CCProposal)

	bs = state.NewBlockState(&state.StateDB{})
	tx.Payload = []byte(`{"name":"changeCluster", "args":[{"command" : "remove", "id": "1234"}]}`)
	_, err = ExecuteEnterpriseTx(bs, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err)
	assert.NotNil(t, bs.CCProposal)

	bs = state.NewBlockState(&state.StateDB{})
	tx.Payload = []byte(`{"name":"changeCluster", "args":[{"command" : "nocmd", "name": "aergonew", "address": "/ip4/127.0.0.1/tcp/11001", "PeerID":"16Uiu2HAmAAtqye6QQbeG9EZnrWJbGK8Xw74cZxpnGGEAZAB3zJ8B"}]}`)
	_, err = ExecuteEnterpriseTx(bs, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err)
	assert.Nil(t, bs.CCProposal)

	bs = state.NewBlockState(&state.StateDB{})
	tx.Payload = []byte(`{"name":"changeCluster", "args":[{"command" : "add", "name": "aergonew", "address": "http://127.0.0.1:1001", "peerid":"16Uiu2HAmAAtqye6QQbeG9EZnrWJbGK8Xw74cZxpnGGEAZAB3zJ8B"}]}`)
	_, err = ExecuteEnterpriseTx(bs, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err)
	assert.Nil(t, bs.CCProposal)
}

func TestCheckArgs(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	tx := &types.TxBody{}
	testBlockNo := types.BlockNo(1)

	tx.Payload = []byte(`{"name":"appendAdmin", "args":["AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"]}`)
	_, err := ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "add admin")

	block, _ := pem.Decode([]byte(testCert))
	assert.NotNil(t, block, "parse value 0")
	cert := types.EncodeB64(block.Bytes)
	tx.Payload = []byte(`{"name":"appendConf", "args":["rpcpermissions","` + cert + `:RWCS"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, RPCPermissions)

	//missing permission string
	tx.Payload = []byte(`{"name":"appendConf", "args":["rpcpermissions","` + cert + `"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, RPCPermissions)

	//invalid rpc cert
	tx.Payload = []byte(`{"name":"appendConf", "args":["rpcpermissions","-+TEST+-:RWCS"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, RPCPermissions)

	tx.Payload = []byte(`{"name":"appendConf", "args":["accountwhite","AmMMFgzR14wdQBTCCuyXQj3NYrBenecCmurutTqPqqBZ9TEY2z7c"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, AccountWhite)

	//invalid account address
	tx.Payload = []byte(`{"name":"appendConf", "args":["accountwhite","BmMMFgzR14wdQBTCCuyXQj3NYrBenecCmurutTqPqqBZ9TEY2z7c"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.Error(t, err, AccountWhite)
}

func TestEnterpriseAdminAccountWhitelist(t *testing.T) {
	scs, sender, receiver := initTest(t)
	defer deinitTest()

	tx := &types.TxBody{}
	testBlockNo := types.BlockNo(1)

	tx.Payload = []byte(`{"name":"appendAdmin", "args":["AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"]}`)
	_, err := ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "add admin")
	tx.Payload = []byte(`{"name":"appendAdmin", "args":["AmLt7Z3y2XTu7YS8KHNuyKM2QAszpFHSX77FLKEt7FAuRW7GEhj7"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "add admin")

	tx.Payload = []byte(`{"name":"appendConf", "args":["accountwhite","AmMMFgzR14wdQBTCCuyXQj3NYrBenecCmurutTqPqqBZ9TEY2z7c"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, AccountWhite)

	tx.Payload = []byte(`{"name":"enableConf", "args":["accountwhite",true]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.EqualError(t, err, "the values of ACCOUNTWHITE should have at least one admin address", AccountWhite)

	tx.Payload = []byte(`{"name":"appendConf", "args":["accountwhite","AmLt7Z3y2XTu7YS8KHNuyKM2QAszpFHSX77FLKEt7FAuRW7GEhj7"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, AccountWhite)

	tx.Payload = []byte(`{"name":"removeAdmin", "args":["AmLt7Z3y2XTu7YS8KHNuyKM2QAszpFHSX77FLKEt7FAuRW7GEhj7"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "remove admin")

	tx.Payload = []byte(`{"name":"enableConf", "args":["accountwhite",true]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.EqualError(t, err, "the values of ACCOUNTWHITE should have at least one admin address", AccountWhite)

	tx.Payload = []byte(`{"name":"appendAdmin", "args":["AmLt7Z3y2XTu7YS8KHNuyKM2QAszpFHSX77FLKEt7FAuRW7GEhj7"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, "add admin")

	tx.Payload = []byte(`{"name":"enableConf", "args":["accountwhite",true]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.NoError(t, err, AccountWhite)

	tx.Payload = []byte(`{"name":"removeAdmin", "args":["AmLt7Z3y2XTu7YS8KHNuyKM2QAszpFHSX77FLKEt7FAuRW7GEhj7"]}`)
	_, err = ExecuteEnterpriseTx(nil, ccc, scs, tx, sender, receiver, testBlockNo)
	assert.EqualError(t, err, "admin is in the account whitelist: AmLt7Z3y2XTu7YS8KHNuyKM2QAszpFHSX77FLKEt7FAuRW7GEhj7", AccountWhite)
}
