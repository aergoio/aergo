package contract

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"path"
	"strings"
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
	/*function inc()
		a = system.getItem("key1")
		if (a == nil) then
			system.setItem("key1", 1)
			return
		end
		system.setItem("key1", a + 1)
	end
	function query(a)
			return system.getItem(a)
	end
	abi.register(inc)
	abi.register(query) */
	queryCode = "Bc9qiEuzkF76vQfsmDgW15NCgEVNpQvJcEXLD5Ug6afZkfawubznZxgCwTpGtDttsntWgogxAfd7h1Ki6hWd6YhSdzoeJWQ3biTpkFuiB8t15vgx6FKQur3q9zwkQcgN2BGEAjsrKE47j1fVDBGcYPEbZFgWn5nbet3R3NHa3dfT7RMZNVSR1NKLdKXvbh81vHAKbsERsMsNRHg1E8fLs4dJR4yPvJy2rteC4JcmZ5EuVLeier8WMNxadgVsLhcdRbgj2j9A4vW2AUawvFFrZPG4mkVNJ2Z51KcCvFecEXXZN4bpvhihA6RbX6s8DwxQt3ajfXyS5ge79CXcq1nbq53SQPP8arEox1tdHqsVZ39X6dr4juEtWNe8d1uu1Stsb7DdSfqDv9CwPmE5oeMXwurCqBRYBy5n6vrG7Bqtbxjhv6UX7jzWRttDzHCSiq5pRtzgsMkUjcZ6d9q9R3bYPQm6uRNi9rHpEf6AhHTiWiLvafPrD2iTvizwWJqeobjzBNAhoz"
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

	contractState := getContractState(t)
	contractCall(t, contractState, helloCode, &ci, nil, txId)
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
	contractState := getContractState(t)
	bcCtx := NewContext(contractState, sender, txhash, 100, 1234,
		"node", true, contractId, false)

	contractCall(t, contractState, systemCode, &ci, bcCtx, txId)
	receipt := types.NewReceiptFromBytes(DB.Get([]byte(txId)))

	if receipt.GetRet() != "[\"sender2\",\"c2b367\",\"31KcyXb99xYD5tQ9Jpx4BMnhVh9a\",1234,100,999]" {
		t.Errorf("contract Call ret error :%s\n", receipt.GetRet())
	}

}

func TestGetABI(t *testing.T) {
	contractState := getContractState(t)
	contractId, err := base58.Decode(accountId)
	if err != nil {
		t.Fatal(err)
	}
	code, err := base58.Decode(helloCode)
	if err != nil {
		t.Fatal(err)
	}
	err = contractState.SetCode(code)
	if err != nil {
		t.Fatal(err)
	}
	abi, err := GetABI(contractState, contractId)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(abi)
}

func TestContractQuery(t *testing.T) {
	var ci types.CallInfo
	txId := "c2b36750"

	ci.Name = "inc"
	queryInfo := "{\"Name\":\"query\", \"Args\":[\"key1\"]}"
	setInfo := "{\"Name\":\"inc\", \"Args\":[]}"

	contractId, _ := base58.Decode(accountId)
	contractState, err := sdb.OpenContractStateAccount(types.ToAccountID(contractId))
	if err != nil {
		t.Errorf("contract open error :%s\n", err.Error())
	}
	code, _ := base58.Decode(queryCode)
	contractState.SetCode(code)

	ret, err := Query(contractId, contractState, []byte(setInfo))
	if err == nil || !strings.Contains(err.Error(), "not permitted set in query") {
		t.Errorf("failed check error: %s", err.Error())
	}

	bcCtx := NewContext(contractState, nil, nil, 100, 1234,
		"node", true, contractId, false)

	err = Call(contractState, []byte(setInfo), contractId, []byte(txId), bcCtx)
	if err != nil {
		t.Fatalf("contract inc error :%s", err.Error())
	}

	ret, err = Query(contractId, contractState, []byte(queryInfo))
	if err != nil {
		t.Errorf("contract query error :%s\n", err.Error())
	}

	if string(ret) != "[1]" {
		t.Errorf("contract query ret error :%s\n", ret)
	}
}
