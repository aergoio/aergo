package contract

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
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
	aid []byte
	tid []byte
)

const (
	accountId = "AnZNgPahfxH1pE8DkLFCt3uEQZBbivmxsF1s2QbHG7J6owEV2qh6"
	txId      = "c2b36750"
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

	var err error
	aid, err = types.DecodeAddress(accountId)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	tid, err = hex.DecodeString(txId)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func getContractState(t *testing.T, code string) *state.ContractState {
	contractState, err := sdb.OpenContractStateAccount(types.ToAccountID(aid))
	if err != nil {
		t.Fatalf("contract state open error : %s\n", err.Error())
	}
	rcode, err := base58.Decode(code)
	if err != nil {
		t.Fatalf("contract SetCode error : %s\n", err.Error())
	}

	err = contractState.SetCode(rcode)
	if err != nil {
		t.Fatalf("contract SetCode error : %s\n", err.Error())
	}
	return contractState
}

func contractCall(t *testing.T, contractState *state.ContractState, ci string,
	bcCtx *LBlockchainCtx) {
	dbTx := DB.NewTx(true)
	err := Call(contractState, []byte(ci), aid, tid, bcCtx, dbTx)
	dbTx.Commit()
	if err != nil {
		t.Fatalf("contract Call error : %s\n", err.Error())
	}
}

func TestContractHello(t *testing.T) {
	callInfo := "{\"Name\":\"hello\", \"Args\":[\"World\"]}"

	contractState := getContractState(t, helloCode)
	contractCall(t, contractState, callInfo, nil)
	receipt := types.NewReceiptFromBytes(DB.Get(tid))

	if receipt.GetRet() != "[\"Hello World\"]" {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
}

func TestContractSystem(t *testing.T) {
	callInfo := "{\"Name\":\"testState\", \"Args\":[]}"
	sender, _ := base58.Decode("sender2")
	contractState := getContractState(t, systemCode)
	bcCtx := NewContext(contractState, sender, tid, 100, 1234,
		"node", true, aid, false)

	contractCall(t, contractState, callInfo, bcCtx)
	receipt := types.NewReceiptFromBytes(DB.Get(tid))

	// sender, txhash, contractid, timestamp, blockheight, getItem("key1")
	if receipt.GetRet() != "[\"HNM6akcic1ou1fX\",\"c2b36750\",\"AnZNgPahfxH1pE8DkLFCt3uEQZBbivmxsF1s2QbHG7J6owEV2qh6\",1234,100,999]" {
		t.Errorf("contract Call ret error: %s\n", receipt.GetRet())
	}

}

func TestGetABI(t *testing.T) {
	contractState := getContractState(t, helloCode)
	abi, err := GetABI(contractState, aid)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(abi)
}

func TestContractQuery(t *testing.T) {
	queryInfo := "{\"Name\":\"query\", \"Args\":[\"key1\"]}"
	setInfo := "{\"Name\":\"inc\", \"Args\":[]}"

	contractState := getContractState(t, queryCode)

	ret, err := Query(aid, contractState, []byte(setInfo))
	if err == nil || !strings.Contains(err.Error(), "not permitted set in query") {
		t.Errorf("failed check error: %s", err.Error())
	}

	bcCtx := NewContext(contractState, nil, nil, 100, 1234,
		"node", true, aid, false)

	contractCall(t, contractState, setInfo, bcCtx)

	ret, err = Query(aid, contractState, []byte(queryInfo))
	if err != nil {
		t.Errorf("contract query error :%s\n", err.Error())
	}

	if string(ret) != "[1]" {
		t.Errorf("contract query ret error :%s\n", ret)
	}
}
