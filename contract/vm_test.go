package contract

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
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
	helloCode = "CNHAxLsQM73PTo14NxWGxfEDTjbxiZ2SQVAZRR7WeKkXFhu9LCCtqbZj2pFwkoE8cX28YWf8FcmioGRe4Mi7vhUTFHhadhAVpVSv2inngQJ3d3cRnnpo6UgP2VzWX52R5fpWyjg12EdRNK9K74aypS69uEqv68xibXavPdpPBAidPkVuJV5tkivZ3LYcVeQ4pn4L7QDqSyGrJmPy2ME7iKfawSZUmRGW1vERzjcKqH17SZAgrbdHG2KkmCVCHat2ogtJDyXE9b7F"
	/*function testState()
		string.format("creator: %s",system.getContractID())
		string.format("timestamp: %d",system.getTimestamp())
		string.format("blockheight: %d",system.getBlockheight())
		system.setItem("key1", 999)
		string.format("getitem : %s",system.getItem("key1"))
		return system.getSender(), system.getTxhash(),system.getContractID(), system.getTimestamp(), system.getBlockheight(), system.getItem("key1")
	  end abi.register(testState) */
	systemCode = "86VMDdo4twhyq7NpCkCqnn3uWVdL4EetXKSCKpVocGaShaZzeZnTbD8o6ikoLuz9XSJ3U4nyCDHMWa69Yf5YKyUciPTAfCHSmMXcrTpttEA3Gw2Ew6mZ6ceHGuYg8yaEA1vS7orTaCU6Lkdt1xTCmVoYfxXVFzB4jroCJvn6YpT5qba3mbgEsQ92zafTtGCbVjTgK7mBu2PTCgLLwfkR1G6aQy2QR851FHk5UqbZbioFUL5WhkpKpwXTZW78iwpgDgFzQsK56bYciv9MN26KYhLaM5KmJMjdRQdUi6z89JKFJxBAEf1wtZ389HCG4RzRHLiMUvqriDQtKqTJfCdP6wHERfdTi7SowvhPev2iqctS18hf5zFBWfAJZMh2Khd2wYbhSfXHcJDQ9ma2tnLR5ESbRiQAvsyu5aAFzkJH5sZvoNUHda4xyPuQE6wZLqnC6cDJQdWy1wXrHpBkwQ2tpATevya4j2e4KcM6TXert5jGKkSSe5tK3RTxW9LGYqkUnTKyMxtnGpfMMqPxvR7tGD8cZymf1he7xS8kt8aXr6MT9HQ3DdqJUvBpJ3GJEvxVa3Kf5kTfjKC3AaEwpdy4tUFLWHrxzzqKefw1k1rdQF6WUQCacZCFiswTvU1hpx6rdycDzxKeXYTyEgxWbb2L5th6ayd5VAezehDUHZLHeFwxc23fNqPrPsLQkZzxVxobf4Q6ckY9gBFzsmQqkQTeCQvckC7mZ9g33LNpLZ"
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
	queryCode = "52fg8Ywt6dwekcRC6KHx27QSg8umh2nF6bVdmYRp81s3eXW9fXP8bLVJqbjsqC3yhVFPFPQFVoQz1s77f9osQw4hRwSBtGRUbxTbxDtnJTxhP3oNgEAmFg7vujtRER7z5bH3P1nkoRgSaCX2A6qws4ueVZF1vrCbehfkYNcCedHDYFuJhpcRuWicL8Mfuj4aBLWsnXtTUGfN9xSKJWYYprAeUfDQAVSrPRS1TfmhwJhm3PJaxBGswSFcH1b3kCYDsXyj8mDjCPUKXdaMj7j5qnVg2FVXMiSDizeibjfrs3ySZ13dQwhWmn1NZvgxSirnW7bNYjsF39Hsm5iNBizLZWatBisVXrVpQPj2xAX7esLZMemFwxoJtmMZAwYHzKCvT1Sq5sJZgPBJ17mhwB2MTC1h6aH7y9bMcDtPQfxNoNsyJEqKG4kFTpwSrLJ1o2zDuHt1rYsz69ZYB1uvtdG8v5vDmSpZjhzgVmmQCEKwbMGYJAFQU65p3s9RvyaiFPNHMef91cgq5uXxQ"
)

func init() {
	sdb = state.NewChainStateDB()

	tmpDir, _ := ioutil.TempDir("", "vmtest")

	err := sdb.Init(path.Join(tmpDir, "testDB"))
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	TempReceiptDb = db.NewDB(db.BadgerImpl, path.Join(tmpDir, "receiptDB"))

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

	dir, err := ioutil.TempDir("", "SqlDB")
	if err != nil {
		log.Fatal(err)
	}

	LoadDatabase(dir)
}

func getContractState(t *testing.T, code string) *state.ContractState {
	contractState, err := sdb.OpenContractStateAccount(types.ToAccountID(aid))
	if err != nil {
		t.Fatalf("contract state open error : %s\n", err.Error())
	}
	rcode, err := util.DecodeCode(code)
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
	dbTx := TempReceiptDb.NewTx()
	err := Call(contractState, []byte(ci), aid, tid, bcCtx, dbTx)
	dbTx.Commit()
	if err != nil {
		t.Fatalf("contract Call error : %s\n", err.Error())
	}
}

func TestContractHello(t *testing.T) {
	callInfo := `{"Name":"hello", "Args":["World"]}`

	contractState := getContractState(t, helloCode)
	contractCall(t, contractState, callInfo, nil)
	receipt := types.NewReceiptFromBytes(TempReceiptDb.Get(tid))

	if receipt.GetRet() != `["Hello World"]` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
}

func TestContractSystem(t *testing.T) {
	callInfo := "{\"Name\":\"testState\", \"Args\":[]}"
	contractState := getContractState(t, systemCode)
	bcCtx := NewContext(sdb, nil, nil, contractState, "HNM6akcic1ou1fX", "c2b36750", 100, 1234,
		"node", 1, accountId, 0, nil, nil)

	contractCall(t, contractState, callInfo, bcCtx)
	receipt := types.NewReceiptFromBytes(TempReceiptDb.Get(tid))

	// sender, txhash, contractid, timestamp, blockheight, getItem("key1")
	if receipt.GetRet() != `["HNM6akcic1ou1fX","c2b36750","AnZNgPahfxH1pE8DkLFCt3uEQZBbivmxsF1s2QbHG7J6owEV2qh6",1234,100,999]` {
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
	queryInfo := `{"Name":"query", "Args":["key1"]}`
	setInfo := `{"Name":"inc", "Args":[]}`

	contractState := getContractState(t, queryCode)

	tx, err := BeginTx(types.ToAccountID(aid), 1)
	if err != nil {
		t.Error(err)
	}
	tx.Commit()

	_, err = Query(aid, contractState, []byte(setInfo))
	if err == nil || !strings.Contains(err.Error(), "not permitted set in query") {
		t.Errorf("failed check error: %s", err.Error())
	}

	bcCtx := NewContext(sdb, nil, nil, contractState, "", "", 100, 1234,
		"node", 1, accountId, 0, nil, nil)

	contractCall(t, contractState, setInfo, bcCtx)

	ret, err := Query(aid, contractState, []byte(queryInfo))
	if err != nil {
		t.Errorf("contract query error :%s\n", err.Error())
	}

	if string(ret) != "[1]" {
		t.Errorf("contract query ret error :%s\n", ret)
	}
}
