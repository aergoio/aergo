package enterprise

import (
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

func TestBasicFailEnterprise(t *testing.T) {
	scs, sender, _ := initTest(t)
	defer deinitTest()
	tx := &types.TxBody{}
	_, err := ExecuteEnterpriseTx(scs, tx, sender)
	assert.Error(t, err, "empty body")
	tx.Payload = []byte("invalid")
	_, err = ExecuteEnterpriseTx(scs, tx, sender)
	assert.Error(t, err, "invalid body")
	tx.Payload = []byte("{}")
	_, err = ExecuteEnterpriseTx(scs, tx, sender)
	assert.Error(t, err, "empty json")
	tx.Payload = []byte(`{"name":"enableConf"}`)
	_, err = ExecuteEnterpriseTx(scs, tx, sender)
	assert.Error(t, err, "empty arg in enable conf")
	tx.Payload = []byte(`{"name":"setConf"}`)
	_, err = ExecuteEnterpriseTx(scs, tx, sender)
	assert.Error(t, err, "empty arg in set conf")
	tx.Payload = []byte(`{"name":"enableConf", "args":["raft",true]}`)
	_, err = ExecuteEnterpriseTx(scs, tx, sender)
	assert.Error(t, err, "admin is not set when enble conf")
	tx.Payload = []byte(`{"name":"setConf", "args":["raft","thisisraftid1", "thisisraftid2"]}`)
	_, err = ExecuteEnterpriseTx(scs, tx, sender)
	assert.Error(t, err, "admin is not set when set conf")
	tx.Payload = []byte(`{"name":"setAdmin", "args":["AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr"]}`)
	_, err = ExecuteEnterpriseTx(scs, tx, sender)
	assert.Error(t, err, "invalid arg in set admin")
	tx.Payload = []byte(`{"name":"setAdmin", "args":[]}`)
	_, err = ExecuteEnterpriseTx(scs, tx, sender)
	assert.Error(t, err, "invalid arg in set admin")

	tx.Payload = []byte(`{"name":"setAdmin", "args":["AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"]}`)
	_, err = ExecuteEnterpriseTx(scs, tx, sender)
	assert.NoError(t, err, "set admin")
	tx.Payload = []byte(`{"name":"setAdmin", "args":["AmLqZFnwMLqLg5fMshgzmfvwBP8uiYGgfV3tBZAm36Tv7jFYcs4f"]}`)
	_, err = ExecuteEnterpriseTx(scs, tx, sender)
	assert.NoError(t, err, "set admin")
	tx.Payload = []byte(`{"name":"setAdmin", "args":["AmLqZFnwMLqLg5fMshgzmfvwBP8uiYGgfV3tBZAm36Tv7jFYcs4f"]}`)
	_, err = ExecuteEnterpriseTx(scs, tx, sender)
	assert.Error(t, err, "set admin without permission")
}

func TestBasicEnterprise(t *testing.T) {
	scs, sender, _ := initTest(t)
	defer deinitTest()
	tx := &types.TxBody{}
	tx.Payload = []byte(`{"name":"setAdmin", "args":["AmPNYHyzyh9zweLwDyuoiUuTVCdrdksxkRWDjVJS76WQLExa2Jr4"]}`)
	_, err := ExecuteEnterpriseTx(scs, tx, sender)
	assert.NoError(t, err, "set admin")
	tx.Payload = []byte(`{"name":"setConf", "args":["p2pwhite","16Uiu2HAmAokYAtLbZxJAPRgp2jCc4bD35cJD921trqUANh59Rc4n", "16Uiu2HAm4xYtGsqk7WGKUxr8prfVpJ25hD23AQ3Be6anEL9Kxkgw", "16Uiu2HAmGiJ2QgVAWHMUtzLKKNM5eFUJ3Ds3FN7nYJq1mHN5ZPj9"]}`)
	_, err = ExecuteEnterpriseTx(scs, tx, sender)
	assert.NoError(t, err, "set conf")
	conf, err := getConf(scs, []byte("p2pwhite"))
	assert.Equal(t, false, conf.On, "conf on")
	assert.Equal(t, 3, len(conf.Values), "conf values length")
	assert.Equal(t, "16Uiu2HAmAokYAtLbZxJAPRgp2jCc4bD35cJD921trqUANh59Rc4n", conf.Values[0], "conf value 0")
	assert.Equal(t, "16Uiu2HAm4xYtGsqk7WGKUxr8prfVpJ25hD23AQ3Be6anEL9Kxkgw", conf.Values[1], "conf value 1")
	assert.Equal(t, "16Uiu2HAmGiJ2QgVAWHMUtzLKKNM5eFUJ3Ds3FN7nYJq1mHN5ZPj9", conf.Values[2], "conf value 2")

	tx.Payload = []byte(`{"name":"enableConf", "args":["p2pwhite",true]}`)
	_, err = ExecuteEnterpriseTx(scs, tx, sender)
	assert.NoError(t, err, "enable conf")
	conf, err = getConf(scs, []byte("p2pwhite"))
	assert.Equal(t, true, conf.On, "conf on")

	tx.Payload = []byte(`{"name":"enableConf", "args":["p2pwhite",false]}`)
	_, err = ExecuteEnterpriseTx(scs, tx, sender)
	assert.NoError(t, err, "enable conf")
	conf, err = getConf(scs, []byte("p2pwhite"))
	assert.Equal(t, false, conf.On, "conf on")
}
