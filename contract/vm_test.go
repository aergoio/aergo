package contract

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"path"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/mr-tron/base58/base58"
)

var (
	sdb *state.ChainStateDB
)

const (
	accountId = "31KcyXb99xYD5tQ9Jpx4BMnhVh9a"
	/*function hello(say) return "Hello " .. say end abi.register(hello)*/
	helloCode = "C34wJetPFqYBV8bpSao36R2NxFcW5X4ZGZoPcnfJvHkvvBR4PcbsZ5xki8nSHubPMnNusFPpn19x7myhR8baq12RvoufQ8z2DR1PyvGfYf6VmAzhF5rg8F7mvVEBRdqnAMnFmb5E6E2iey4wEjXNrjJ8RfsPEfottZ2umDN5WFc8egeydXesa1a59QLBp926MoETbDwRJMeFHeGHuvQ4bikCXhXaQS6Vi73y4Xpcu3Bv1rMnfBxEX5ZehtcuYDFUoZzn8"
	/*function testState()
		print(string.format("creator: %s",system.getContractID()))
		print(string.format("timestamp: %d",system.getTimestamp()))
		print(string.format("blockheight: %d",system.getBlockheight()))
		system.setItem("key1", 999)
		print(string.format("getitem : %s",system.getItem("key1")))
		return system.getSender(), system.getTxHash(),system.getContractID(), system.getTimestamp(), system.getBlockheight(), system.getItem("key1")
	  end abi.register(testState) */
	systemCode = "CCPAqwHFihSv6zD8ri2HiE1nqhxojJUywAwgU1aA78ZQthLAPFUbF3dtMDa1Ti79SdXu853dGSrejzuvxkgm3GPXD6rUgx9HvD7AhcRRzwcon2LmvnMgzCf8F2jPVHxquP9NU7DyWPbdBiMi99ByUFQePHwQ6dE2n3wLNaqVPs3uAapDWzHZBPR3EbhxjnQUfrRv8YzFXMeVYbWxgoWwivpZTFwgNyWcbJcEkXZWwYdfmuC5v2bbTpsWtvPP8Qw95F4XGui1mZ2FLawFoCMG3NA5twDyyfeQ2CHhjCK3GzfoXCBKKQJSfdfRsHrgdukpU3aB9R3QKiGbZRd8MV6uxmrRAcwm6Uu2Xj1dJMUrbYE1iQU6SJ9tV6X44NXafvtuyRrLvpvkoyU95m2F1tQRtLVtFV75S72W5gGMgeAF4RjJWWTdh4dNoDy4LcPNH86aT6CDE876vMLLGhwnAXKG8EU7W3TD14aoGZEVhymMHbu6Afrpk1sXD5hiw39iKUKTH5m8AzQKJEQJT2tfN5pu3Q1SZsLqtF9uuSKcAostds34czYkm6BY9BuNL6v3GR4T5uiobKWy9VdhRsq7NENGTLYUTgNFjKGVrEHQbVtqBCJVJbpLjhQkSL2ztemcjC243JjnsyRLVQfZijLWPj1oWrsNzdVvgNvZEJi7M5czxVuRZBepTF28KwQHCftvEtoovfGgjmU2gTrf3nCiJbzhCKKNzwUaS8LwkhuCgf1tw34zmH36LPtrCz1kYF4SLEVtts8dmYAJYSErDmhNYV2"
)

func init() {
	sdb = state.NewStateDB()

	tmpDir, _ := ioutil.TempDir("", "vmtest")

	sdb.Init(path.Join(tmpDir, "testDB"))
	DB = db.NewDB(db.BadgerImpl, path.Join(tmpDir, "receiptDB"))
}

func getContractState(t *testing.T) *state.ContractState {
	aid, _ := base58.Decode(accountId)
	accountState, err := sdb.GetAccountStateClone(types.ToAccountID(aid))
	if err != nil {
		t.Errorf("getAccount error : %s\n", err.Error())
	}

	stateChange := types.Clone(*accountState).(types.State)
	contractState, err := sdb.OpenContractState(&stateChange)
	if err != nil {
		t.Errorf("contract open error : %s\n", err.Error())
	}
	return contractState
}

func contractCall(t *testing.T, contractState *state.ContractState, code string, ci *types.CallInfo,
	bcCtx *LBlockchainCtx, txId string) {
	rcode, _ := base58.Decode(code)

	err := contractState.SetCode(rcode)
	if err != nil {
		t.Errorf("contract SetCode error : %s\n", err.Error())
	}

	payload, _ := json.Marshal(*ci)

	err = Call(contractState, payload, []byte(accountId), []byte(txId), bcCtx)
	if err != nil {
		t.Errorf("contract Call error : %s\n", err.Error())
	}
}

func TestContractHello(t *testing.T) {
	var ci types.CallInfo

	txId := "c2b36745"
	ci.Name = "hello"
	json.Unmarshal([]byte("[\"World\"]"), &ci.Args)

	contractStatus := getContractState(t)
	contractCall(t, contractStatus, helloCode, &ci, nil, txId)
	receipt := types.NewReceiptFromBytes(DB.Get([]byte(txId)))

	if receipt.GetRet() != "[\"Hello World\"]" {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
}

func TestContractSystem(t *testing.T) {
	var ci types.CallInfo
	txId := "c2b36750"

	ci.Name = "testState"
	contractId, _ := base58.Decode(accountId)
	txhash, _ := hex.DecodeString("c2b367")
	sender, _ := base58.Decode("sender2")
	contractStatus := getContractState(t)
	bcCtx := NewContext(contractStatus, sender, txhash, 100, 1234,
		"node", true, contractId)

	contractCall(t, contractStatus, systemCode, &ci, bcCtx, txId)
	receipt := types.NewReceiptFromBytes(DB.Get([]byte(txId)))

	if receipt.GetRet() != "[\"sender2\",\"c2b367\",\"31KcyXb99xYD5tQ9Jpx4BMnhVh9a\",1234,100,999]" {
		t.Errorf("contract Call ret error :%s\n", receipt.GetRet())
	}

}
