package contract

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/aergoio/aergo/types"
)

const (
	helloCode = `function hello(say) return "Hello " .. say end abi.register(hello)`

	systemCode = `function testState()
		system.setItem("key1", 999)
		return system.getSender(), system.getTxhash(),system.getContractID(), system.getTimestamp(), system.getBlockheight(), system.getItem("key1")
	  end 
abi.register(testState)`

	queryCode = `function inc()
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
	abi.register(query)
	abi.payable(inc)`
)

func TestReturn(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "return_num", 0, "function return_num() return 10 end abi.register(return_num)"),
		NewLuaTxCall("ktlee", "return_num", 0, `{"Name":"return_num", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("return_num", `{"Name":"return_num", "Args":[]}`, "", "10")
	if err != nil {
		t.Error(err)
	}

	foo := `function foo()
	return {1,2,3}
end
function foo2(bar)
	return bar
	end
abi.register(foo,foo2)`

	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "foo", 0, foo),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("foo", `{"Name":"foo", "Args":[]}`, "", "[1,2,3]")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("foo", `{"Name":"foo2", "Args":["foo314"]}`, "", `"foo314"`)
	if err != nil {
		t.Error(err)
	}
}

func TestContractHello(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
	)
	_ = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "hello", 0, helloCode),
	)
	tx := NewLuaTxCall("ktlee", "hello", 0, `{"Name":"hello", "Args":["World"]}`)
	_ = bc.ConnectBlock(tx)
	receipt := bc.getReceipt(tx.hash())
	if receipt.GetRet() != `"Hello World"` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
}

func TestContractSystem(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
	)
	_ = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "system", 0, systemCode),
	)
	tx := NewLuaTxCall("ktlee", "system", 0, `{"Name":"testState", "Args":[]}`)
	_ = bc.ConnectBlock(tx)
	receipt := bc.getReceipt(tx.hash())
	exRv := fmt.Sprintf(`["Amg6nZWXKB6YpNgBPv9atcjdm6hnFvs5wMdRgb2e9DmaF5g9muF2","99NTyZ796bpvwLLhMmsfwo8J3Wu3rUioUQsHE9CSYQKz","AmhNNBNY7XFk4p5ym4CJf8nTcRTEHjWzAeXJfhP71244CjBCAQU3",%d,3,999]`, bc.cBlock.Header.Timestamp/1e9)
	if receipt.GetRet() != exRv {
		t.Errorf("expected: %s, but got: %s", exRv, receipt.GetRet())
	}
}

func TestGetABI(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "hello", 0,
			`state.var {
	Say = state.value()
}

function hello(say) 
  return "Hello " .. say 
end 

abi.register(hello)`),
	)
	abi, err := bc.GetABI("hello")
	if err != nil {
		t.Error(err)
	}
	b, err := json.Marshal(abi)
	if err != nil {
		t.Error(err)
	}
	if string(b) != `{"version":"0.2","language":"lua","functions":[{"name":"hello","arguments":[{"name":"say"}]}],"state_variables":[{"name":"Say","type":"value"}]}` {
		t.Error(string(b))
	}
}

func TestContractQuery(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
	)
	_ = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "query", 0, queryCode),
		NewLuaTxCall("ktlee", "query", 2, `{"Name":"inc", "Args":[]}`),
	)

	ktlee, err := bc.GetAccountState("ktlee")
	if err != nil {
		t.Error(err)
	}
	if ktlee.GetBalanceBigInt().Uint64() != uint64(98) {
		t.Error(ktlee.Balance)
	}
	query, err := bc.GetAccountState("query")
	if err != nil {
		t.Error(err)
	}
	if query.GetBalanceBigInt().Uint64() != uint64(2) {
		t.Error(query.Balance)
	}

	err = bc.Query("query", `{"Name":"inc", "Args":[]}`, "set not permitted in query", "")
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "1")
	if err != nil {
		t.Error(err)
	}
}

func TestRollback(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
	)
	_ = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "query", 0, queryCode),
		NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`),
	)
	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`),
		NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`),
	)
	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`),
		NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`),
	)

	err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "5")
	if err != nil {
		t.Error(err)
	}

	err = bc.DisConnectBlock()
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "3")
	if err != nil {
		t.Error(err)
	}

	err = bc.DisConnectBlock()
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "1")
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "query", 0, `{"Name":"inc", "Args":[]}`),
	)

	err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "2")
	if err != nil {
		t.Error(err)
	}
}

func TestVote(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
function constructor()
	system.setItem("owner", system.getSender())
end

function addCandidate(name)
	if system.getSender() ~= system.getItem("owner") then
		return
	end

	if (system.getItem(name) ~= nil) then
		return
	end
	
	local numCandidates;
	if (system.getItem("numCandidates") == nil) then
		numCandidates = 0;
	else
		numCandidates = tonumber(system.getItem("numCandidates"))
	end

	system.setItem("candidate_list_" .. numCandidates, name)

	numCandidates = numCandidates + 1;
	system.setItem("numCandidates", tostring(numCandidates));
	system.setItem(name, tostring(0));
end

function getCandidates()
	local numCandidates;
	if (system.getItem("numCandidates") == nil) then
		return {};
	else
		numCandidates = tonumber(system.getItem("numCandidates"))
	end

	local candidates = {};
	local i = 0;

	while true do
		if (numCandidates == i) then
			break;
		end
		local candidate = system.getItem("candidate_list_" .. i)
		local count = system.getItem(candidate)
		if count == nil then
			count = 0
		end
		table.insert(candidates, {id = i, name = candidate, count = count});
		i = i + 1;
	end
	return candidates;
end

function registerVoter(address)
	if system.getSender() ~= system.getItem("owner") then
		return
	end
	
	system.setItem("voter_" .. address, "0");
end

function vote(candidateID)
	local totalVoted
	local voter = system.getItem("voter_" .. system.getSender())
	if voter == nil then
		return
	end
	totalVoted = tonumber(system.getItem("voter_" .. system.getSender()))
	if totalVoted > 3 then
		return
	end
	if system.getItem(candidateID) == nil then
		return
	end
	local currentVotes;
	if (system.getItem(candidateID) == nil) then
		currentVotes = 0;
	else
		currentVotes = tonumber(system.getItem(candidateID))
	end
	currentVotes = currentVotes + 1

	system.setItem(candidateID, tostring(currentVotes))
	totalVoted = totalVoted + 1
	system.setItem("voter_" .. system.getSender(), tostring(totalVoted));
end

abi.register(addCandidate, getCandidates, registerVoter, vote)`

	_ = bc.ConnectBlock(
		NewLuaTxAccount("owner", 100),
		NewLuaTxDef("owner", "vote", 0, definition),
		NewLuaTxAccount("user1", 100),
	)

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"owner",
			"vote",
			0,
			`{"Name":"addCandidate", "Args":["candidate1"]}`,
		),
		NewLuaTxCall(
			"owner",
			"vote",
			0,
			`{"Name":"addCandidate", "Args":["candidate2"]}`,
		),
		NewLuaTxCall(
			"owner",
			"vote",
			0,
			`{"Name":"addCandidate", "Args":["candidate3"]}`,
		),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query(
		"vote",
		`{"Name":"getCandidates"}`,
		"",
		`[{"count":"0","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`,
	)
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"user1",
			"vote",
			0,
			`{"Name":"addCandidate", "Args":["candidate4"]}`,
		),
	)
	err = bc.Query(
		"vote",
		`{"Name":"getCandidates"}`,
		"",
		`[{"count":"0","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`,
	)
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		// register voter
		NewLuaTxCall(
			"owner",
			"vote",
			0,
			fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(strHash("user10"))),
		),
		NewLuaTxCall(
			"owner",
			"vote",
			0,
			fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(strHash("user10"))),
		),
		NewLuaTxCall(
			"owner",
			"vote",
			0,
			fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(strHash("user11"))),
		),
		NewLuaTxCall(
			"owner",
			"vote",
			0,
			fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(strHash("user1"))),
		),
		// vote
		NewLuaTxCall(
			"user1",
			"vote",
			0,
			`{"Name":"vote", "Args":["user1"]}`,
		),
		NewLuaTxCall(
			"user1",
			"vote",
			0,
			`{"Name":"vote", "Args":["user1"]}`,
		),
		NewLuaTxCall(
			"user1",
			"vote",
			0,
			`{"Name":"vote", "Args":["user2"]}`,
		),
		NewLuaTxCall(
			"user1",
			"vote",
			0,
			`{"Name":"vote", "Args":["user2"]}`,
		),
		NewLuaTxCall(
			"user1",
			"vote",
			0,
			`{"Name":"vote", "Args":["user3"]}`,
		),
	)

	err = bc.Query(
		"vote",
		`{"Name":"getCandidates"}`,
		"",
		`[{"count":"0","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`,
	)
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"user11",
			"vote",
			0,
			`{"Name":"vote", "Args":["candidate1"]}`,
		),
		NewLuaTxCall(
			"user10",
			"vote",
			0,
			`{"Name":"vote", "Args":["candidate1"]}`,
		),
	)

	err = bc.Query(
		"vote",
		`{"Name":"getCandidates"}`,
		"",
		`[{"count":"2","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestInfiniteLoop(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
function infiniteLoop()
    local t = 0
	while true do
	    t = t + 1
	end
	return t
end
abi.register(infiniteLoop)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "loop", 0, definition),
		NewLuaTxCall(
			"ktlee",
			"loop",
			0,
			`{"Name":"infiniteLoop"}`,
		),
	)
	errMsg := "exceeded the maximum instruction count"
	if err == nil {
		t.Errorf("expected: %s", errMsg)
	}
	if err != nil && !strings.Contains(err.Error(), errMsg) {
		t.Error(err)
	}
}

func TestUpdateSize(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
function infiniteLoop()
	for i = 1, 100000000000000 do
		system.setItem("key_"..i, "value_"..i)
	end
end
abi.register(infiniteLoop)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "loop", 0, definition),
		NewLuaTxCall(
			"ktlee",
			"loop",
			0,
			`{"Name":"infiniteLoop"}`,
		),
	)
	errMsg := "exceeded size of updates in the state database"
	if err == nil {
		t.Errorf("expected: %s", errMsg)
	}
	if err != nil && !strings.Contains(err.Error(), errMsg) {
		t.Error(err)
	}
}

func TestSqlVmSimple(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
function createAndInsert()
    db.exec("create table if not exists dual(dummy char(1))")
	db.exec("insert into dual values ('X')")
    local insertYZ = db.prepare("insert into dual values (?),(?)")
    insertYZ:exec("Y", "Z")
end

function insertRollbackData()
	db.exec("insert into dual values ('A'),('B'),('C')")
end

function query()
    local rt = {}
    local stmt = db.prepare("select ?+1, round(?, 1), dummy || ? as col3 from dual order by col3")
    local rs = stmt:query(1, 3.14, " Hello Blockchain")
    while rs:next() do
        local col1, col2, col3 = rs:get()
        table.insert(rt, col1)
        table.insert(rt, col2)
        table.insert(rt, col3)
    end
    return rt
end

function count()
	local rs = db.query("select count(*) from dual")
	if rs:next() then
		local n = rs:get()
		--rs:next()
		return n
	else
		return "error in count()"
	end
end

function all()
    local rt = {}
    local rs = db.query("select dummy from dual order by 1")
    while rs:next() do
        local col = rs:get()
        table.insert(rt, col)
    end
    return rt
end

abi.register(createAndInsert, insertRollbackData, query, count, all)`

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "simple-query", 0, definition),
	)
	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "simple-query", 0, `{"Name": "createAndInsert", "Args":[]}`),
	)
	err = bc.Query(
		"simple-query",
		`{"Name": "query", "Args":[]}`,
		"",
		`[2,3.1,"X Hello Blockchain",2,3.1,"Y Hello Blockchain",2,3.1,"Z Hello Blockchain"]`,
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query(
		"simple-query",
		`{"Name": "count", "Args":[]}`,
		"",
		`3`,
	)
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "simple-query", 0, `{"Name": "createAndInsert", "Args":[]}`),
	)
	err = bc.Query(
		"simple-query",
		`{"Name": "count", "Args":[]}`,
		"",
		`6`,
	)
	if err != nil {
		t.Error(err)
	}

	_ = bc.DisConnectBlock()

	err = bc.Query(
		"simple-query",
		`{"Name": "count", "Args":[]}`,
		"",
		`3`,
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.DisConnectBlock()
	if err != nil {
		t.Error(err)
	}
	err = bc.DisConnectBlock()
	if err != nil {
		t.Error(err)
	}

	// there is only a genesis block
	err = bc.Query(
		"simple-query",
		`{"Name": "count", "Args":[]}`,
		"cannot find contract",
		"",
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmFail(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
function init()
    db.exec("create table if not exists total(n int)")
	db.exec("insert into total values (0)")
end

function add(n)
	local stmt = db.prepare("update total set n = n + ?")
	stmt:exec(n)
end

function addFail(n)
	local stmt = db.prepare("update set n = n + ?")
	stmt:exec(n)
end

function get()
	local rs = db.query("select n from total")
	rs:next()
	n = rs:get()
	return n
end
abi.register(init, add, addFail, get)`

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "fail", 0, definition),
		NewLuaTxCall("ktlee", "fail", 0, `{"Name":"init"}`),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "fail", 0, `{"Name":"add", "Args":[1]}`),
	)

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "fail", 0, `{"Name":"add", "Args":[2]}`),
		NewLuaTxCall("ktlee", "fail", 0, `{"Name":"addFail", "Args":[3]}`).
			Fail(`near "set": syntax error`),
		NewLuaTxCall("ktlee", "fail", 0, `{"Name":"add", "Args":[4]}`),
	)
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "fail", 0, `{"Name":"add", "Args":[5]}`),
	)

	err = bc.Query("fail", `{"Name":"get"}`, "", "12")
	if err != nil {
		t.Error(err)
	}

	_ = bc.DisConnectBlock()

	err = bc.Query("fail", `{"Name":"get"}`, "", "7")
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmDateTime(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	definition := `
function init()
    db.exec("create table if not exists dt_test (n datetime, b bool)")
	db.exec("insert into dt_test values (10000, 1),(date('2004-10-24', '+1 month', '-1 day'), 0)")
end

function nowNull()
	db.exec("insert into dt_test values (date('now'), 0)")
end

function localtimeNull()
	db.exec("insert into dt_test values (datetime('2018-05-25', 'localtime'), 1)")
end

function get()
	local rs = db.query("select n, b from dt_test order by 1, 2")
	local r = {}
	while rs:next() do
		local d, b = rs:get()
		table.insert(r, { date= d, bool= b })
	end
	return r
end
abi.register(init, nowNull, localtimeNull, get)`

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "datetime", 0, definition),
		NewLuaTxCall("ktlee", "datetime", 0, `{"Name":"init"}`),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "datetime", 0, `{"Name":"nowNull"}`),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "datetime", 0, `{"Name":"localtimeNull"}`),
	)

	err = bc.Query(
		"datetime",
		`{"Name":"get"}`,
		"",
		`[{"bool":0},{"bool":1},{"bool":1,"date":"1970-01-01 02:46:40"},{"bool":0,"date":"2004-11-23"}]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlConstrains(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
function init()
    db.exec([[create table if not exists r (
  id integer primary key
, n integer check(n >= 10)
, nonull text not null
, only integer unique)
]])
    db.exec("insert into r values (1, 11, 'text', 1)")
	db.exec("create table if not exists s (rid integer references r(id))")
end

function pkFail()
	db.exec("insert into r values (1, 12, 'text', 2)")
end

function checkFail()
	db.exec("insert into r values (2, 9, 'text', 3)")
end

function fkFail()
	db.exec("insert into s values (2)")
end

function notNullFail()
	db.exec("insert into r values (2, 13, null, 2)")
end

function uniqueFail()
	db.exec("insert into r values (2, 13, 'text', 1)")
end

abi.register(init, pkFail, checkFail, fkFail, notNullFail, uniqueFail)`

	err = bc.ConnectBlock(
		NewLuaTxAccount(
			"ktlee",
			100,
		),
		NewLuaTxDef(
			"ktlee",
			"constraint",
			0,
			definition,
		),
		NewLuaTxCall(
			"ktlee",
			"constraint",
			0,
			`{"Name":"init"}`,
		),
		NewLuaTxCall(
			"ktlee",
			"constraint",
			0,
			`{"Name":"pkFail"}`,
		).Fail("UNIQUE constraint failed: r.id"),
		NewLuaTxCall(
			"ktlee",
			"constraint",
			0,
			`{"Name":"checkFail"}`,
		).Fail("CHECK constraint failed: r"),
		NewLuaTxCall(
			"ktlee",
			"constraint",
			0,
			`{"Name":"fkFail"}`,
		).Fail("FOREIGN KEY constraint failed"),
		NewLuaTxCall(
			"ktlee",
			"constraint",
			0,
			`{"Name":"notNullFail"}`,
		).Fail("NOT NULL constraint failed: r.nonull"),
		NewLuaTxCall(
			"ktlee",
			"constraint",
			0,
			`{"Name":"uniqueFail"}`,
		).Fail("UNIQUE constraint failed: r.only"),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmCustomer(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
function createTable()
  db.exec([[create table if not exists customer(
        id varchar(10),
        passwd varchar(20),
        name varchar(30),
        birth char(8),
        mobile varchar(20)
    )]])
end

function query(id)
    local rt = {}
    local stmt = db.prepare("select * from customer where id like '%' || ? || '%'")
    local rs = stmt:query(id)
    while rs:next() do
        local col1, col2, col3, col4, col5 = rs:get()
        local item = {
                    id = col1,
                    passwd = col2,
                    name = col3,
                    birth = col4,
                    mobile = col5
            }
        table.insert(rt, item)
    end
    return rt
end

function insert(id , passwd, name, birth, mobile)
    local stmt = db.prepare("insert into customer values (?,?,?,?,?)")
    stmt:exec(id, passwd, name, birth, mobile)
end

function update(id , passwd)
    local stmt = db.prepare("update customer set passwd =? where id =?")
    stmt:exec(passwd, id)
end

function delete(id)
    local stmt = db.prepare("delete from customer where id =?")
    stmt:exec(id)
end

function count()
	local rs = db.query("select count(*) from customer")
	if rs:next() then
		local n = rs:get()
		return n
	else
		return "error in count()"
	end
end

abi.register(createTable, query, insert, update, delete, count)`

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "customer", 0, definition),
		NewLuaTxCall(
			"ktlee",
			"customer",
			0,
			`{"Name":"createTable"}`,
		),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"customer",
			0,
			`{"Name":"insert", "Args":["id1","passwd1","name1","20180524","010-1234-5678"]}`,
		),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"customer",
			0,
			`{"Name":"insert", "Args":["id2","passwd2","name2","20180524","010-1234-5678"]}`,
		),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"customer",
			0,
			`{"Name":"update", "Args":["id2","passwd3"]}`,
		),
	)

	err = bc.Query("customer", `{"Name":"count"}`, "", "2")
	if err != nil {
		t.Error(err)
	}

	_ = bc.DisConnectBlock()

	err = bc.Query(
		"customer",
		`{"Name":"query", "Args":["id2"]}`,
		"",
		`[{"birth":"20180524","id":"id2","mobile":"010-1234-5678","name":"name2","passwd":"passwd2"}]`,
	)
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"customer",
			0,
			`{"Name":"delete", "Args":["id2"]}`,
		),
	)

	err = bc.Query(
		"customer",
		`{"Name":"query", "Args":["id2"]}`,
		"",
		`{}`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmDataType(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
function createDataTypeTable()
  db.exec([[create table if not exists datatype_table(
        var1 varchar(10),
        char1 char(10),
        int1 int(5),
        float1 float(6),
        blockheight1 long
    )]])
end

function dropDataTypeTable()
   db.exec("drop table datatype_table")
end

function insertDataTypeTable()
    local stmt = db.prepare("insert into datatype_table values ('ABCD','fgh',1,3.14,?)")
    stmt:exec(system.getBlockheight())
end
function queryOrderByDesc()
    local rt = {}
    local rs = db.query("select * from datatype_table order by blockheight1 desc")
    while rs:next() do
        local col1, col2, col3, col4, col5 = rs:get()
        item = {
                    var1 = col1,
                    char1 = col2,
                    int1 = col3,
                    float1 = col4,
                    blockheight1 = col5
            }
        table.insert(rt, item)
    end
    return rt
end

function queryGroupByBlockheight1()
    local rt = {}
    local rs = db.query("select blockheight1, count(*), sum(int1), avg(float1) from datatype_table group by blockheight1")
    while rs:next() do
        local col1, col2, col3, col4 = rs:get()
        item = {
                    blockheight1 = col1,
                    count1 = col2,
                    sum_int1 = col3,
                    avg_float1 =col4
            }
        table.insert(rt, item)
    end
    return rt
end

abi.register(createDataTypeTable, dropDataTypeTable, insertDataTypeTable, queryOrderByDesc, queryGroupByBlockheight1)`

	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "datatype", 0, definition),
		NewLuaTxCall(
			"ktlee",
			"datatype",
			0,
			`{"Name":"createDataTypeTable"}`,
		),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"datatype",
			0,
			`{"Name":"insertDataTypeTable"}`,
		),
		NewLuaTxCall(
			"ktlee",
			"datatype",
			0,
			`{"Name":"insertDataTypeTable"}`,
		),
		NewLuaTxCall(
			"ktlee",
			"datatype",
			0,
			`{"Name":"insertDataTypeTable"}`,
		),
	)

	_ = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"datatype",
			0,
			`{"Name":"insertDataTypeTable"}`,
		),
	)

	err = bc.Query(
		"datatype",
		`{"Name":"queryOrderByDesc"}`,
		"",
		`[{"blockheight1":3,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"},{"blockheight1":2,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"},{"blockheight1":2,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"},{"blockheight1":2,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"}]`,
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query(
		"datatype",
		`{"Name":"queryGroupByBlockheight1"}`,
		"",
		`[{"avg_float1":3.14,"blockheight1":2,"count1":3,"sum_int1":3},{"avg_float1":3.14,"blockheight1":3,"count1":1,"sum_int1":1}]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmFunction(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
function sql_func()
    local rt = {}
    local rs = db.query("select round(3.14),min(1,2,3), max(4,5,6)")
	if rs:next() then
	    local col1, col2, col3 = rs:get()
        table.insert(rt, col1)
        table.insert(rt, col2)
        table.insert(rt, col3)
        return rt
	else
		return "error in func()"
	end
end

function abs_func()
    local rt = {}
    local rs = db.query("select abs(-1),abs(0), abs(1)")
	if rs:next() then
	    local col1, col2, col3 = rs:get()
        table.insert(rt, col1)
        table.insert(rt, col2)
        table.insert(rt, col3)
        return rt
	else
		return "error in abs()"
	end
end

function typeof_func()
    local rt = {}
    local rs = db.query("select typeof(-1), typeof('abc'), typeof(3.14), typeof(null)")
	if rs:next() then
	    local col1, col2, col3, col4 = rs:get()
        table.insert(rt, col1)
        table.insert(rt, col2)
        table.insert(rt, col3)
        table.insert(rt, col4)
        return rt
	else
		return "error in typeof()"
	end
end

abi.register(sql_func, abs_func, typeof_func)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("name", 100),
		NewLuaTxDef("ktlee", "fns", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("fns", `{"Name":"sql_func"}`, "", `[3,1,6]`)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("fns", `{"Name":"abs_func"}`, "", `[1,0,1]`)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("fns", `{"Name":"typeof_func"}`,
		"", `["integer","text","real","null"]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmBook(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
function createTable()
  db.exec([[create table if not exists book (
        page number,
        contents text
    )]])

  db.exec([[create table if not exists copy_book (
        page number,
        contents text
    )]])
end

function makeBook()
   	local stmt = db.prepare("insert into book values (?,?)")
	for i = 1, 100 do    
   		stmt:exec(i, "value=" .. i*i)
    end
end

function copyBook()
    local rs = db.query("select page, contents from book order by page asc")
    while rs:next() do
        local col1, col2 = rs:get()
        local stmt_t = db.prepare("insert into copy_book values (?,?)")
        stmt_t:exec(col1, col2)
    end
end


function viewCopyBook()
    local rt = {}
    local rs = db.query("select max(page), min(contents) from copy_book")
    while rs:next() do
        local col1, col2 = rs:get()
        table.insert(rt, col1)
		table.insert(rt, col2)
    end
    return rt
end

function viewJoinBook()
    local rt = {}
    local rs = db.query([[select c.page, b.page, c.contents  
							from copy_book c, book b 
							where c.page = b.page and c.page = 10 ]])
    while rs:next() do
        local col1, col2, col3 = rs:get()
        table.insert(rt, col1)
		table.insert(rt, col2)
		table.insert(rt, col3)
    end
    return rt
end

abi.register(createTable, makeBook, copyBook, viewCopyBook, viewJoinBook)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "book", 0, definition),
		NewLuaTxCall(
			"ktlee",
			"book",
			0,
			`{"Name":"createTable"}`,
		),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"book",
			0,
			`{"Name":"makeBook"}`,
		),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"book",
			0,
			`{"Name":"copyBook"}`,
		),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query(
		"book",
		`{"Name":"viewCopyBook"}`,
		"",
		`[100,"value=1"]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmDateformat(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
function init()
	db.exec("drop table if exists dateformat_test")
	db.exec([[create table if not exists dateformat_test
	(
		col1 date ,
		col2 datetime ,
		col3 text
	)]])
	db.exec("insert into dateformat_test values (date('2004-10-24 11:11:11'), datetime('2004-10-24 11:11:11'),strftime('%Y%m%d%H%M%S','2004-10-24 11:11:11'))")
	db.exec("insert into dateformat_test values (date(1527504338,'unixepoch'), datetime(1527504338,'unixepoch'), strftime('%Y%m%d%H%M%S',1527504338,'unixepoch') )")
end

function get()
    local rt = {}
    local rs = db.query([[select col1, col2, col3
                            from dateformat_test ]])
    while rs:next() do
        local col1, col2, col3 = rs:get()
        table.insert(rt, {col1,col2,col3} )
    end
    return rt
end

abi.register(init, get)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef(
			"ktlee",
			"data_format",
			0,
			definition,
		),
		NewLuaTxCall("ktlee", "data_format", 0, `{"Name":"init"}`),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query(
		"data_format",
		`{"Name":"get"}`,
		"",
		`[["2004-10-24","2004-10-24 11:11:11","20041024111111"],["2018-05-28","2018-05-28 10:45:38","20180528104538"]]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmRecursiveData(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
function r()
	local t = {}
	t["name"] = "ktlee"
	t["self"] = t
	return t
end
abi.register(r)`

	tx := NewLuaTxCall("ktlee", "r", 0, `{"Name":"r"}`)
	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "r", 0, definition),
		tx,
	)
	if err != nil {
		t.Error(err)
	}
	receipt := bc.getReceipt(tx.hash())
	if receipt.GetRet() != `nested table error` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
}

func TestContractCall(t *testing.T) {
	definition1 := `
	function constructor(init)
		system.setItem("count", init)
	end
	function inc()
		count = system.getItem("count")
		system.setItem("count", count + 1)
		return count
	end

	function get()
		return system.getItem("count")
	end

	function set(val)
		system.setItem("count", val)
	end
	abi.register(inc,get,set)
	`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "counter", 0, definition1).Constructor("[1]"),
		NewLuaTxCall("ktlee", "counter", 0, `{"Name":"inc", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "2")
	if err != nil {
		t.Error(err)
	}

	definition2 := `
	function constructor(addr)
		system.setItem("count", 99)
		system.setItem("addr", addr)
	end
	function add(amount)
		return contract.call.value(amount)(system.getItem("addr"), "inc")
	end
	function dadd()
		return contract.delegatecall(system.getItem("addr"), "inc")
	end
	function get()
		addr = system.getItem("addr")
		a = contract.call(addr, "get")
		return a
	end
	function dget()
		addr = system.getItem("addr")
		a = contract.delegatecall(addr, "get")
		return a
	end
	function set(val)
		contract.call(system.getItem("addr"), "set", val)
	end
	function dset(val)
		contract.delegatecall(system.getItem("addr"), "set", val)
	end
	abi.register(add,dadd, get, dget, set, dset)
	`
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "caller", 0, definition2).
			Constructor(fmt.Sprintf(`["%s"]`, types.EncodeAddress(strHash("counter")))),
		NewLuaTxCall("ktlee", "caller", 0, `{"Name":"add", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("caller", `{"Name":"get", "Args":[]}`, "", "3")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("caller", `{"Name":"dget", "Args":[]}`, "", "99")
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "caller", 0, `{"Name":"dadd", "Args":[]}`)
	_ = bc.ConnectBlock(tx)
	receipt := bc.getReceipt(tx.hash())
	if receipt.GetRet() != `99` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	tx = NewLuaTxCall("ktlee", "caller", 0, `{"Name":"dadd", "Args":[]}`)
	_ = bc.ConnectBlock(tx)
	receipt = bc.getReceipt(tx.hash())
	if receipt.GetRet() != `100` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	err = bc.Query("caller", `{"Name":"get", "Args":[]}`, "", "3")
	if err != nil {
		t.Error(err)
	}
}

func TestSparseTable(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
function is_table_equal(t1,t2,ignore_mt)
   local ty1 = type(t1)
   local ty2 = type(t2)
   if ty1 ~= ty2 then return false end
   -- non-table types can be directly compared
   if ty1 ~= 'table' and ty2 ~= 'table' then return t1 == t2 end
   -- as well as tables which have the metamethod __eq
   local mt = getmetatable(t1)
   if not ignore_mt and mt and mt.__eq then return t1 == t2 end
   for k1,v1 in pairs(t1) do
      local v2 = t2[k1]
      if v2 == nil or not is_table_equal(v1,v2) then return false end
   end
   for k2,v2 in pairs(t2) do
      local v1 = t1[k2]
      if v1 == nil or not is_table_equal(v1,v2) then return false end
   end
   return true
end

function r()
	local t = {}
	t[10000] = "1234"
	system.setItem("k", t)
	k = system.getItem("k")
	if is_table_equal(t, k, false) then
		return 1
    end
	return 0
end
abi.register(r)`

	tx := NewLuaTxCall("ktlee", "r", 0, `{"Name":"r"}`)
	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "r", 0, definition),
		tx,
	)
	if err != nil {
		t.Error(err)
	}
	receipt := bc.getReceipt(tx.hash())
	if receipt.GetRet() != `1` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
}

func TestKvstore(t *testing.T) {
	definition := `
	state.var{
		counts = state.map(),
		name = state.value()
	}

	function inc(key)
		if counts[key] == nil then
			counts[key] = 0
		end
		counts[key] = counts[key] + 1
	end

	function get(key)
		return counts[key]
	end

	function set(key,val)
		counts[key] = val
	end

	function setname(n)
		name:set(n)
	end

	function getname()
		return name:get()
	end

	abi.register(inc,get,set,setname,getname)`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "map", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "map", 0, `{"Name":"inc", "Args":["ktlee"]}`),
		NewLuaTxCall("ktlee", "map", 0, `{"Name":"setname", "Args":["eve2adam"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock()
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("map", `{"Name":"get", "Args":["ktlee"]}`, "", "1")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("map", `{"Name":"get", "Args":["htwo"]}`, "", "{}")
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "map", 0, `{"Name":"inc", "Args":["ktlee"]}`),
		NewLuaTxCall("ktlee", "map", 0, `{"Name":"inc", "Args":["htwo"]}`),
		NewLuaTxCall("ktlee", "map", 0, `{"Name":"set", "Args":["wook", 100]}`),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("map", `{"Name":"get", "Args":["ktlee"]}`, "", "2")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("map", `{"Name":"get", "Args":["htwo"]}`, "", "1")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("map", `{"Name":"get", "Args":["wook"]}`, "", "100")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("map", `{"Name":"getname"}`, "", `"eve2adam"`)
	if err != nil {
		t.Error(err)
	}
}

func TestJson(t *testing.T) {
	definition := `
	state.var{
		table = state.value()
	}

	function set(val)
		table:set(json.decode(val))
	end

	function get()
		return table:get()
	end

	function getenc()
		return json.encode(table:get())
	end
	
	function getlen()
		a = table:get()
		return a[1], string.len(a[1])
	end

	function getAmount()
		return system.getAmount()
	end

	abi.register(set, get, getenc, getlen)
	abi.payable(getAmount)`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "json", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "json", 0, `{"Name":"set", "Args":["[1,2,3]"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", "[1,2,3]")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"[1,2,3]"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "json", 0,
			`{"Name":"set", "Args":["{\"key1\":[1,2,3], \"run\", \"key2\":5, [4,5,6]}"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `{"1":"run","2":[4,5,6],"key1":[1,2,3],"key2":5}`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"{\"1\":\"run\",\"2\":[4,5,6],\"key1\":[1,2,3],\"key2\":5}"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "json", 0,
			`{"Name":"set", "Args":["{\"key1\":{\"arg1\": 1,\"arg2\":{}, \"arg3\":[]}, \"key2\":[5,4,3]}"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `{"key1":{"arg1":1,"arg2":{},"arg3":{}},"key2":[5,4,3]}`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"{\"key1\":{\"arg1\":1,\"arg2\":{},\"arg3\":{}},\"key2\":[5,4,3]}"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "json", 0,
			`{"Name":"set", "Args":["{\"key1\":[1,2,3], \"key1\":5}"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `{"key1":5}`)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "json", 0, `{"Name":"set", "Args":["[\"\\\"hh\\t\",\"2\",3]"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `["\"hh\u0009","2",3]`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"getlen", "Args":[]}`, "", `["\"hh\u0009",4]`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"[\"\\\"hh\\u0009\",\"2\",3]"`)
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "json", 100, `{"Name":"getAmount"}`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	receipt := bc.getReceipt(tx.hash())
	if receipt.GetRet() != `"100"` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "json", 0,
			`{"Name":"set", "Args":["{\"key1\":[1,2,3], \"key1\":5}}"]}`).Fail("not proper json format"),
	)
}

func TestArray(t *testing.T) {
	definition := `
	state.var{
		counts = state.array(10)
	}

	function inc(key)
		if counts[key] == nil then
			counts[key] = 0
		end
		counts[key] = counts[key] + 1
	end

	function get(key)
		return counts[key]
	end

	function set(key,val)
		counts[key] = val
	end

	function len()
		return counts:length()
	end

	function iter()
		local rv = {}
		for i, v in counts:ipairs() do 
			if v == nil then
				rv[i] = "nil"
			else
				rv[i] = v
			end
		end
		return rv
	end

	abi.register(inc,get,set,len,iter)`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "array", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[1]}`),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[0]}`).Fail("index out of range"),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[1]}`),
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"set", "Args":[2,"ktlee"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("array", `{"Name":"get", "Args":[11]}`, "index out of range", "")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("array", `{"Name":"get", "Args":[1]}`, "", "2")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("array", `{"Name":"get", "Args":[2]}`, "", `"ktlee"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("array", `{"Name":"len"}`, "", `10`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("array", `{"Name":"iter"}`, "", `[2,"ktlee","nil","nil","nil","nil","nil","nil","nil","nil"]`)
	if err != nil {
		t.Error(err)
	}
}

func TestPcall(t *testing.T) {
	definition1 := `
	function constructor(init)
		system.setItem("count", init)
	end

	function init()
		db.exec([[create table if not exists r (
	  id integer primary key
	, n integer check(n >= 10)
	, nonull text not null
	, only integer unique)
	]])
		db.exec("insert into r values (1, 11, 'text', 1)")
	end

	function pkins1()
		db.exec("insert into r values (3, 12, 'text', 2)")
		db.exec("insert into r values (1, 12, 'text', 2)")
	end

	function pkins2()
		db.exec("insert into r values (4, 12, 'text', 2)")
	end

	function pkget()
		local rs = db.query("select count(*) from r")
		if rs:next() then
			local n = rs:get()
			--rs:next()
			return n
		else
			return "error in count()"
		end
	end

	function inc()
		count = system.getItem("count")
		system.setItem("count", count + 1)
		return count
	end

	function get()
		return system.getItem("count")
	end

	function getOrigin()
		return system.getOrigin()
	end

	function set(val)
		system.setItem("count", val)
	end
	abi.register(inc,get,set, init, pkins1, pkins2, pkget, getOrigin)
	abi.payable(constructor, inc)
	`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "counter", 10, definition1).Constructor("[0]"),
		NewLuaTxCall("ktlee", "counter", 15, `{"Name":"inc", "Args":[]}`),
	)

	err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "1")
	if err != nil {
		t.Error(err)
	}

	definition2 := `
	function constructor(addr)
		system.setItem("count", 99)
		system.setItem("addr", addr)
	end
	function add(amount)
		first = contract.call.value(amount)(system.getItem("addr"), "inc")
		status, res = contract.pcall(contract.call.value(1000000), system.getItem("addr"), "inc")
		if status == false then
			return first
		end
		return res
	end
	function dadd()
		return contract.delegatecall(system.getItem("addr"), "inc")
	end
	function get()
		addr = system.getItem("addr")
		a = contract.call(addr, "get")
		return a
	end
	function dget()
		addr = system.getItem("addr")
		a = contract.delegatecall(addr, "get")
		return a
	end
	function send(addr, amount)
		contract.send(addr, amount)
		status, res = contract.pcall(contract.call.value(1000000000)(system.getItem("addr"), "inc"))
		return status
	end
	function sql()
		contract.call(system.getItem("addr"), "init")
		contract.pcall(contract.call, system.getItem("addr"), "pkins1")
		contract.call(system.getItem("addr"), "pkins2")
	end

	function sqlget()
		return contract.call(system.getItem("addr"), "pkget")
	end

	function getOrigin()
		return contract.call(system.getItem("addr"), "getOrigin")
	end
	abi.register(add, dadd, get, dget, send, sql, sqlget, getOrigin)
	abi.payable(constructor,add)
	`
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "caller", 10, definition2).
			Constructor(fmt.Sprintf(`["%s"]`, types.EncodeAddress(strHash("counter")))),
		NewLuaTxCall("ktlee", "caller", 15, `{"Name":"add", "Args":[]}`),
		NewLuaTxCall("ktlee", "caller", 0, `{"Name":"sql", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("caller", `{"Name":"get", "Args":[]}`, "", "2")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("caller", `{"Name":"sqlget", "Args":[]}`, "", "2")
	if err != nil {
		t.Error(err)
	}

	tx := NewLuaTxCall("ktlee", "caller", 0, `{"Name":"getOrigin", "Args":[]}`)
	_ = bc.ConnectBlock(tx)
	receipt := bc.getReceipt(tx.hash())
	if receipt.GetRet() != "\""+types.EncodeAddress(strHash("ktlee"))+"\"" {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}

	definition3 := `
	function pass(addr)
		contract.send(addr, 1)
	end

	function add(addr, a, b)
		system.setItem("arg", a)
		contract.pcall(pass, addr)
		return a+b
	end

	function set(addr)
		contract.send(addr, 1)
		system.setItem("arg", 2)
		status, ret  = contract.pcall(add, addr, 1, 2)
	end

	function set2(addr)
		contract.send(addr, 1)
		system.setItem("arg", 2)
		status, ret  = contract.pcall(add, addar, 1)
	end

	function get()
		return system.getItem("arg")
	end
	
	function getBalance()
		return contract.balance()
	end

	abi.register(set, set2, get, getBalance)
	abi.payable(set, set2)
	`

	bc, err = LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxAccount("bong", 0),
		NewLuaTxDef("ktlee", "counter", 0, definition3),
	)
	if err != nil {
		t.Error(err)
	}
	tx = NewLuaTxCall("ktlee", "counter", 20,
		fmt.Sprintf(`{"Name":"set", "Args":["%s"]}`, types.EncodeAddress(strHash("bong"))))
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "1")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("counter", `{"Name":"getBalance", "Args":[]}`, "", "\"18\"")
	if err != nil {
		t.Error(err)
	}
	state, err := bc.GetAccountState("bong")
	if state.GetBalanceBigInt().Uint64() != 2 {
		t.Error("balance error")
	}
	tx = NewLuaTxCall("ktlee", "counter", 10,
		fmt.Sprintf(`{"Name":"set2", "Args":["%s"]}`, types.EncodeAddress(strHash("bong"))))
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "2")
	if err != nil {
		t.Error(err)
	}
	state, err = bc.GetAccountState("bong")
	if state.GetBalanceBigInt().Uint64() != 3 {
		t.Error("balance error")
	}
}

func TestPingpongCall(t *testing.T) {
	definition1 := `
	function constructor()
		system.setItem("key",  "empty")
	end
	function start(addr)
		system.setItem("key",  "start")
		contract.call(addr, "called")
	end

	function callback()
		system.setItem("key",  "callback")
	end

	function get()
		return system.getItem("key")
	end

	abi.register(start, callback, get)
	`

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "a", 0, definition1),
	)

	definition2 := `
	function constructor(addr)
		system.setItem("key",  "empty")
		system.setItem("addr",  addr)
	end

	function called()
		system.setItem("key",  "called")
		contract.call(system.getItem("addr"), "callback")
	end

	function get()
		return system.getItem("key")
	end

	abi.register(called, get)
	`
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "b", 0, definition2).
			Constructor(fmt.Sprintf(`["%s"]`, types.EncodeAddress(strHash("a")))),
	)
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "a", 0,
		fmt.Sprintf(`{"Name":"start", "Args":["%s"]}`, types.EncodeAddress(strHash("b"))))
	_ = bc.ConnectBlock(tx)
	err = bc.Query("a", `{"Name":"get", "Args":[]}`, "", `"callback"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("b", `{"Name":"get", "Args":[]}`, "", `"called"`)
	if err != nil {
		t.Error(err)
	}
}

func TestArrayArg(t *testing.T) {
	definition1 := `
	function copy(arr)
		assert(type(arr) == "table", "table expected")
		local rv = {}
		for i, v in ipairs(arr) do
			table.insert(rv, i, v)
        end
		return rv
	end
	function two_arr(arr1, arr2)
		assert(type(arr1) == "table", "table expected")
		assert(type(arr2) == "table", "table expected")
		local rv = {}
		table.insert(rv, 1, #arr1)
		table.insert(rv, 2, #arr2)
		return rv
	end
	function mixed_args(arr1, map1, n)
		assert(type(arr1) == "table", "table expected")
		assert(type(map1) == "table", "table expected")
		local rv = {}
		table.insert(rv, 1, arr1)
		table.insert(rv, 2, map1)
		table.insert(rv, 3, n)
		return rv
	end

	abi.register(copy, two_arr, mixed_args)
	`

	bc, err := LoadDummyChain()
	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "a", 0, definition1),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name": "copy", "Args":[1, 2, 3]}`, "table expected", "")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name": "copy", "Args":[[1, 2, 3]]}`, "", "[1,2,3]")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name": "two_arr", "Args":[[1, 2, 3],[4, 5]]}`, "", "[3,2]")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name": "mixed_args", "Args":[[1, 2, 3], {"name": "kslee", "age": 39}, 7]}`,
		"",
		`[[1,2,3],{"age":39,"name":"kslee"},7]`,
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name": "mixed_args", "Args":[
[[1, 2, 3],["first", "second"]],
{"name": "kslee", "age": 39, "address": {"state": "XXX-do", "city": "YYY-si"}},
"end"
]}`,
		"",
		`[[[1,2,3],["first","second"]],{"address":{"city":"YYY-si","state":"XXX-do"},"age":39,"name":"kslee"},"end"]`,
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name": "mixed_args", "Args":[
[{"name": "wook", "age": 50}, {"name": "hook", "age": 42}],
{"name": "kslee", "age": 39, "scores": [10, 20, 30, 40, 50]},
"hmm..."
]}`,
		"",
		`[[{"age":50,"name":"wook"},{"age":42,"name":"hook"}],{"age":39,"name":"kslee","scores":[10,20,30,40,50]},"hmm..."]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestAbi(t *testing.T) {
	errMsg := "no exported functions"

	noAbi := `
	function dummy()
		system.print("dummy")
	end`

	bc, err := LoadDummyChain()
	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "a", 0, noAbi),
	)
	if err == nil {
		t.Errorf("expected: %s, but got: nil", errMsg)
	} else if !strings.Contains(err.Error(), "no exported functions") {
		t.Errorf("expected: %s, but got: %s", errMsg, err.Error())
	}

	empty := `
	function dummy()
		system.print("dummy")
	end
	abi.register()`

	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "a", 0, empty),
	)
	if err == nil {
		t.Errorf("expected: %s, but got: nil", errMsg)
	} else if !strings.Contains(err.Error(), "no exported functions.") {
		t.Errorf("expected: %s, but got: %s", errMsg, err.Error())
	}

	localFunc := `
	function dummy()
		system.print("dummy")
	end
	local function helper()
		system.print("helper")
	end
	abi.register(helper)`

	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "a", 0, localFunc),
	)
	if err == nil {
		t.Errorf("expected: %s, but got: nil", errMsg)
	} else if !strings.Contains(err.Error(), "global function expected") {
		t.Errorf("expected: %s, but got: %s", errMsg, err.Error())
	}
}

func TestMapKey(t *testing.T) {
	definition := `
	state.var{
		counts = state.map()
	}
	function setCount(key, value)
		counts[key] = value
	end
	function getCount(key)
		return counts[key]
	end
	function delCount(key)
		counts:delete(key)
	end
	abi.register(setCount, getCount, delCount)
`
	bc, _ := LoadDummyChain()
	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "a", 0, definition),
	)

	err := bc.Query("a", `{"Name":"getCount", "Args":[1]}`, "", "{}")
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "a", 0, `{"Name":"setCount", "Args":[1, 10]}`),
		NewLuaTxCall("ktlee", "a", 0, `{"Name":"setCount", "Args":["1", 20]}`).Fail("number expected, got string)"),
		NewLuaTxCall("ktlee", "a", 0, `{"Name":"setCount", "Args":[1.1, 30]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name":"getCount", "Args":["1"]}`, "(number expected, got string)", "")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name":"getCount", "Args":[1]}`, "", "10")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name":"getCount", "Args":[1.1]}`, "", "30")
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "a", 0,
			`{"Name":"setCount", "Args":[true, 40]}`,
		).Fail(`bad argument #2 to '__newindex' (number expected, got boolean)`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "a", 0, `{"Name":"delCount", "Args":[1.1]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name":"getCount", "Args":[1.1]}`, "", "{}")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("a", `{"Name":"getCount", "Args":[2]}`, "", "{}")
	if err != nil {
		t.Error(err)
	}

	_ = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "x", 0, definition),
	)
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "x", 0, `{"Name":"setCount", "Args":["1", 10]}`),
		NewLuaTxCall("ktlee", "x", 0, `{"Name":"setCount", "Args":[1, 20]}`).Fail("string expected, got number)"),
		NewLuaTxCall("ktlee", "x", 0, `{"Name":"setCount", "Args":["third", 30]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("x", `{"Name":"getCount", "Args":["1"]}`, "", "10")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("x", `{"Name":"getCount", "Args":["third"]}`, "", "30")
	if err != nil {
		t.Error(err)
	}
}

func TestStateVarFieldUpdate(t *testing.T) {
	src := `
state.var{
   Person = state.value()
}

function constructor()
  Person:set({ name = "kslee", age = 38, address = "blahblah..." })
end

function InvalidUpdateAge(age)
  Person:get().age = age
end

function ValidUpdateAge(age)
  local p = Person:get()
  p.age = age
  Person:set(p)
end

function GetPerson()
  return Person:get()
end

abi.register(InvalidUpdateAge, ValidUpdateAge, GetPerson)
`
	bc, _ := LoadDummyChain()
	err := bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "c", 0, src),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "c", 0, `{"Name":"InvalidUpdateAge", "Args":[10]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("c", `{"Name":"GetPerson"}`, "",
		`{"address":"blahblah...","age":38,"name":"kslee"}`,
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "c", 0, `{"Name":"ValidUpdateAge", "Args":[10]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("c", `{"Name":"GetPerson"}`, "",
		`{"address":"blahblah...","age":10,"name":"kslee"}`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestDatetime(t *testing.T) {
	src := `
state.var {
    cdate = state.value()
}

function constructor()
	cdate:set(906000490)
end

function CreateDate()
	return system.date("%c", cdate:get())
end

function Extract(fmt)
	return system.date(fmt, cdate:get())
end

function Difftime()
	system.print(system.date("%c", cdate:get()))
	s = system.date("*t", cdate:get())
	system.print(s)
	s.hour = 2 
	s.min = 0
	s.sec = 0
	system.print(system.date("*t", system.time(s)))
	return system.difftime(cdate:get(), system.time(s))
end

abi.register(CreateDate, Extract, Difftime)
`
	bc, _ := LoadDummyChain()
	err := bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "datetime", 0, src),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("datetime", `{"Name": "CreateDate"}`, "", `"1998-09-17 02:48:10"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("datetime", `{"Name": "Extract", "Args":["%x"]}`, "", `"09/17/98"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("datetime", `{"Name": "Extract", "Args":["%X"]}`, "", `"02:48:10"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("datetime", `{"Name": "Extract", "Args":["%A"]}`, "", `"Thursday"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("datetime", `{"Name": "Extract", "Args":["%I:%M:%S %p"]}`, "", `"02:48:10 AM"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("datetime", `{"Name": "Difftime"}`, "", `2890`)
	if err != nil {
		t.Error(err)
	}
}

func TestDynamicArray(t *testing.T) {
	zeroLen := `
state.var {
    fixedArray = state.array(0)
}

function Length()
	return fixedArray:length()
end

abi.register(Length)
`
	bc, _ := LoadDummyChain()
	_ = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
	)
	err := bc.ConnectBlock(
		NewLuaTxDef("ktlee", "zeroLen", 0, zeroLen),
	)
	if err == nil {
		t.Error("expected: the array length must be greater than zero")
	}
	if !strings.Contains(err.Error(), "the array length must be greater than zero") {
		t.Errorf(err.Error())
	}

	dArr := `
state.var {
    dArr = state.array()
}

function Append(val)
	dArr:append(val)
end

function Get(idx)
	return dArr[idx]
end

function Set(idx, val)
	dArr[idx] = val
end

function Length()
	return dArr:length()
end

abi.register(Append, Get, Set, Length)
`
	tx := NewLuaTxDef("ktlee", "dArr", 0, dArr)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("dArr", `{"Name": "Length"}`, "", "0")
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "dArr", 0, `{"Name": "Append", "Args": [10]}`),
		NewLuaTxCall("ktlee", "dArr", 0, `{"Name": "Append", "Args": [20]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("dArr", `{"Name": "Get", "Args": [1]}`, "", "10")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("dArr", `{"Name": "Get", "Args": [2]}`, "", "20")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("dArr", `{"Name": "Get", "Args": [3]}`, "index out of range", "")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("dArr", `{"Name": "Length"}`, "", "2")
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "dArr", 0, `{"Name": "Append", "Args": [30]}`),
		NewLuaTxCall("ktlee", "dArr", 0, `{"Name": "Append", "Args": [40]}`),
	)
	err = bc.Query("dArr", `{"Name": "Length"}`, "", "4")
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "dArr", 0, `{"Name": "Set", "Args": [3, 50]}`),
	)
	err = bc.Query("dArr", `{"Name": "Get", "Args": [3]}`, "", "50")
	if err != nil {
		t.Error(err)
	}
}

func TestDupVar(t *testing.T) {
	dupVar := `
state.var{
	Var1 = state.value(),
}
function GetVar1()
	return Var1:get()
end
state.var{
	Var1 = state.value(),
}
abi.register(GetVar1)
`
	bc, _ := LoadDummyChain()
	err := bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "dupVar", 0, dupVar),
	)
	if err == nil {
		t.Error("duplicated variable: 'Var1'")
	}
	if !strings.Contains(err.Error(), "duplicated variable: 'Var1'") {
		t.Error(err)
	}

	dupVar = `
state.var{
	Var1 = state.value(),
}
function GetVar1()
	return Var1:get()
end
function Work()
	state.var{
		Var1 = state.value(),
	}
end
abi.register(GetVar1, Work)
`
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "dupVar1", 0, dupVar),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "dupVar1", 0, `{"Name": "Work"}`).Fail("duplicated variable: 'Var1'"),
	)

	if err != nil {
		t.Error(err)
	}
}

func TestCrypto(t *testing.T) {
	src := `
function get(a)
	return crypto.sha256(a)
end

function checkEther()
	return crypto.ecverify("0xce0677bb30baa8cf067c88db9811f4333d131bf8bcf12fe7065d211dce971008",
"0x90f27b8b488db00b00606796d2987f6a5f59ae62ea05effe84fef5b8b0e549984a691139ad57a3f0b906637673aa2f63d1f55cb1a69199d4009eea23ceaddc9301",
"0xbcf9061f21320aa7e824b00d0152398b2d7a6e44")
end

function checkAergo()
	return crypto.ecverify("11e96f2b58622a0ce815b81f94da04ae7a17ba17602feb1fd5afa4b9f2467960",
"304402202e6d5664a87c2e29856bf8ff8b47caf44169a2a4a135edd459640be5b1b6ef8102200d8ea1f6f9ecdb7b520cdb3cc6816d773df47a1820d43adb4b74fb879fb27402",
"AmPbWrQbtQrCaJqLWdMtfk2KiN83m2HFpBbQQSTxqqchVv58o82i")
end
abi.register(get, checkEther, checkAergo)
`
	bc, _ := LoadDummyChain()
	err := bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "crypto", 0, src),
	)
	err = bc.Query("crypto", `{"Name": "get", "Args" : ["ab\u0000\u442a"]}`, "", `"0xc58f6dca13e4bba90a326d8605042862fe87c63a64a9dd0e95608a2ee68dc6f0"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("crypto", `{"Name": "get", "Args" : ["0x616200e490aa"]}`, "", `"0xc58f6dca13e4bba90a326d8605042862fe87c63a64a9dd0e95608a2ee68dc6f0"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("crypto", `{"Name": "checkEther", "Args" : []}`, "", `true`)
	if err != nil {
		t.Error(err)
	}

	err = bc.Query("crypto", `{"Name": "checkAergo", "Args" : []}`, "", `true`)
	if err != nil {
		t.Error(err)
	}
}

func TestPayable(t *testing.T) {
	src := `
state.var {
	Data = state.value()
}
function save(data)
	Data:set(data)
end
function load()
	return Data:get()
end

abi.register(load)
abi.payable(save)
`
	bc, _ := LoadDummyChain()
	err := bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
	)

	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "payable", 1, src),
	)
	if err == nil {
		t.Error("expected: " + errVmConstructorIsNotPayable.Error())
	} else {
		if !strings.Contains(err.Error(), errVmConstructorIsNotPayable.Error()) {
			t.Error(err)
		}
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "payable", 0, `{"Name":"save", "Args": ["blahblah"]}`).Fail("cannot find contract "),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "payable", 0, src),
		NewLuaTxCall("ktlee", "payable", 0, `{"Name":"save", "Args": ["blahblah"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("payable", `{"Name":"load"}`, "", `"blahblah"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "payable", 1, `{"Name":"save", "Args": ["payed"]}`),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("payable", `{"Name":"load"}`, "", `"payed"`)
	if err != nil {
		t.Error(err)
	}
}

func TestDefault(t *testing.T) {
	src := `
function default()
	return "default"
end
abi.register(default)
`
	bc, _ := LoadDummyChain()
	err := bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "default", 0, src),
	)
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "default", 0, "")
	err = bc.ConnectBlock(tx)
	receipt := bc.getReceipt(tx.hash())
	if receipt.GetRet() != `"default"` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "default", 1, "").Fail(`'default' is not payable`),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestBignum(t *testing.T) {
	bigNum := `
function test(addr)
	bal = contract.balance()
	contract.send(addr, bal / 2)
	return contract.balance()
end

function sendS(addr)
	contract.send(addr, "1 gaer 99999")
	return contract.balance()
end

function testBignum()
	bg = bignum.number("999999999999999999999999999999")
	system.setItem("big", bg)
	bi = system.getItem("big")
	return tostring(bi)
end

function argBignum(a)
	b = a + 1
	return tostring(b)
end

function calladdBignum(addr, a)
	return tostring(contract.call(addr, "add", a, 2) + 3)
end

function checkBignum()
	a = 1
	b = bignum.number(1)
	
	return bignum.isbignum(a), bignum.isbignum(b), bignum.isbignum("2333")
end

function constructor()
end

abi.register(test, sendS, testBignum, argBignum, calladdBignum, checkBignum)
abi.payable(constructor)
`
	callee := `
	function add(a, b)
		return a + b
	end
	abi.register(add)
	`
	bc, _ := LoadDummyChain()
	err := bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 1000000000000),
		NewLuaTxDef("ktlee", "bigNum", 50000000000, bigNum),
		NewLuaTxDef("ktlee", "add", 0, callee),
	)
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "bigNum", 0, fmt.Sprintf(`{"Name":"test", "Args":["%s"]}`, types.EncodeAddress(strHash("ktlee"))))
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	receipt := bc.getReceipt(tx.hash())
	if receipt.GetRet() != `"25000000000"` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	tx = NewLuaTxCall("ktlee", "bigNum", 0, fmt.Sprintf(`{"Name":"sendS", "Args":["%s"]}`, types.EncodeAddress(strHash("ktlee"))))
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	receipt = bc.getReceipt(tx.hash())
	if receipt.GetRet() != `"23999900001"` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	tx = NewLuaTxCall("ktlee", "bigNum", 0, `{"Name":"testBignum", "Args":[]}`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	receipt = bc.getReceipt(tx.hash())
	if receipt.GetRet() != `"999999999999999999999999999999"` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	err = bc.Query("bigNum", `{"Name":"argBignum", "Args":[{"_bignum":"99999999999999999999999999"}]}`, "", `"100000000000000000000000000"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("bigNum", fmt.Sprintf(`{"Name":"calladdBignum", "Args":["%s", {"_bignum":"999999999999999999"}]}`, types.EncodeAddress(strHash("add"))), "", `"1000000000000000004"`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("bigNum", `{"Name":"checkBignum"}`, "", `[false,true,false]`)
	if err != nil {
		t.Error(err)
	}
}

func TestDeploy(t *testing.T) {
	deploy := `
function hello()
	hello = [[
	function hello(say) 
		return "Hello " .. say 
	end 
	abi.register(hello)
	]]
	addr = contract.deploy(hello)
	ret = contract.call(addr, "hello", "world")
	return addr, ret
end

function helloQuery(addr)
	return contract.call(addr, "hello", "world")
end

function testConst()
	src = [[
		function hello(say, key) 
			return "Hello " .. say .. system.getItem(key) 
		end 
		function constructor(key, item) 
			system.setItem(key, item)
			return key, item
		end 
		abi.register(hello) 
		abi.payable(constructor)
	]]
	addr, key, item = contract.deploy.value(100)(src, "key", 2)
	ret = contract.call(addr, "hello", "world", "key")
	return addr, ret
end

function testFail()
	src = [[
		function hello(say, key) 
			return "Hello " .. say .. system.getItem(key) 
		end 
		function constructor()
		end 
		abi.register(hello) 
	]]
	addr = contract.deploy.value(100)(src)
	return addr
end
 
paddr = nil
function deploy()
	src = [[
		function hello(say, key) 
			return "Hello " .. say .. system.getItem(key) 
		end 
		function getcre()
			return system.getCreator()
		end
		function constructor()
		end 
		abi.register(hello, getcre) 
	]]
	paddr = contract.deploy(src)
	system.print("addr :", paddr)
	ret = contract.call(paddr, "hello", "world", "key")
end

function testPcall()
	ret = contract.pcall(deploy)
	return contract.call(paddr, "getcre")
end
function constructor()
end

abi.register(hello, helloQuery, testConst, testFail, testPcall)
abi.payable(constructor)
`
	bc, _ := LoadDummyChain()
	err := bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 1000000000000),
		NewLuaTxDef("ktlee", "deploy", 50000000000, deploy),
	)
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "deploy", 0, `{"Name":"hello"}`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	receipt := bc.getReceipt(tx.hash())
	if receipt.GetRet() != `["AmgKtCaGjH4XkXwny2Jb1YH5gdsJGJh78ibWEgLmRWBS5LMfQuTf","Hello world"]` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	err = bc.Query("deploy", `{"Name":"helloQuery", "Args":["AmgKtCaGjH4XkXwny2Jb1YH5gdsJGJh78ibWEgLmRWBS5LMfQuTf"]}`, "", `"Hello world"`)
	if err != nil {
		t.Error(err)
	}
	tx = NewLuaTxCall("ktlee", "deploy", 0, `{"Name":"testConst"}`)
	err = bc.ConnectBlock(tx)
	receipt = bc.getReceipt(tx.hash())
	if receipt.GetRet() != `["Amhmj6kKZz7mPstBAPJWRe1e8RHP7bZ5pV35XatqTHMWeAVSyMkc","Hello world2"]` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	deployAcc, err := bc.GetAccountState("deploy")
	if err != nil {
		t.Error(err)
	}
	if deployAcc.GetBalanceBigInt().Uint64() != uint64(49999999900) {
		t.Error(deployAcc.GetBalanceBigInt().Uint64())
	}
	tx = NewLuaTxCall("ktlee", "deploy", 0, `{"Name":"testFail"}`)
	err = bc.ConnectBlock(tx)
	deployAcc, err = bc.GetAccountState("deploy")
	if err != nil && deployAcc.Nonce == 2 {
		t.Error(err)
	}
	tx = NewLuaTxCall("ktlee", "deploy", 0, `{"Name":"testPcall"}`)
	err = bc.ConnectBlock(tx)
	deployAcc, err = bc.GetAccountState("deploy")
	if err != nil && deployAcc.Nonce == 2 {
		t.Error(err)
	}
	receipt = bc.getReceipt(tx.hash())
	if receipt.GetRet() != `` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
}

func TestSqlVmPubNet(t *testing.T) {
	flushLState := func() {
		for i := 0; i <= MAX_LSTATE_SIZE; i++ {
			s := GetLState()
			FreeLState(s)
		}
	}
	PubNet = true
	flushLState()
	defer func() {
		PubNet = false
		flushLState()
	}()

	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
function createAndInsert()
    db.exec("create table if not exists dual(dummy char(1))")
	db.exec("insert into dual values ('X')")
    local insertYZ = db.prepare("insert into dual values (?),(?)")
    insertYZ:exec("Y", "Z")
end
abi.register(createAndInsert)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "simple-query", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "simple-query", 0, `{"Name": "createAndInsert", "Args":[]}`).Fail(`attempt to index global 'db'`),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestReturnUData(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
	function test_die()
	return contract.call(system.getContractID(), "return_object")
	end
	function return_object()
	return db.query("select 1")
	end
	abi.register(test_die, return_object)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "rs-return", 0, definition),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "rs-return", 0, `{"Name": "test_die", "Args":[]}`).Fail(`unsupport type: userdata`),
	)
	if err != nil {
		t.Error(err)
	}
}

func checkRandomFloatValue(v string) error {
	n, _ := strconv.ParseFloat(v, 64)
	if n < 0.0 || n >= 1.0 {
		return errors.New("out of range")
	}
	return nil
}
func checkRandomIntValue(v string, min, max int) error {
	n, _ := strconv.Atoi(v)
	if n < min || n > max {
		return errors.New("out of range")
	}
	return nil
}

func TestRandom(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	random := `
function random(...)
	return system.random(...)
end
abi.register(random)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "random", 0, random),
	)
	if err != nil {
		t.Error(err)
	}
	tx := NewLuaTxCall("ktlee", "random", 0, `{"Name": "random", "Args":[]}`)
	tx1 := NewLuaTxCall("ktlee", "random", 0, `{"Name": "random", "Args":[]}`)
	err = bc.ConnectBlock(tx, tx1)
	if err != nil {
		t.Error(err)
	}
	receipt := bc.getReceipt(tx.hash())
	err = checkRandomFloatValue(receipt.GetRet())
	if err != nil {
		t.Errorf("error: %s, return value: %s", err.Error(), receipt.GetRet())
	}
	receipt = bc.getReceipt(tx1.hash())
	err = checkRandomFloatValue(receipt.GetRet())
	if err != nil {
		t.Errorf("error: %s, return value: %s", err.Error(), receipt.GetRet())
	}

	err = bc.ConnectBlock(
		NewLuaTxCall(
			"ktlee",
			"random",
			0,
			`{"Name": "random", "Args":[0]}`).Fail("the maximum value must be greater than zero"),
	)
	if err != nil {
		t.Error(err)
	}

	tx = NewLuaTxCall("ktlee", "random", 0, `{"Name": "random", "Args":[3]}`)
	err = bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
	}
	receipt = bc.getReceipt(tx.hash())
	err = checkRandomIntValue(receipt.GetRet(), 1, 3)
	if err != nil {
		t.Errorf("error: %s, return value: %s", err.Error(), receipt.GetRet())
	}

	tx = NewLuaTxCall("ktlee", "random", 0, `{"Name": "random", "Args":[3, 10]}`)
	err = bc.ConnectBlock(tx)
	receipt = bc.getReceipt(tx.hash())
	err = checkRandomIntValue(receipt.GetRet(), 3, 10)
	if err != nil {
		t.Errorf("error: %s, return value: %s", err.Error(), receipt.GetRet())
	}

	err = bc.Query("random", `{"Name": "random", "Args":[1]}`, "", "1")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("random", `{"Name": "random", "Args":[4,4]}`, "", "4")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("random", `{"Name": "random", "Args":[0,4]}`, "system.random: the minimum value must be greater than zero", "")
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("random", `{"Name": "random", "Args":[3,1]}`, "system.random: the maximum value must be greater than the minimum value", "")
	if err != nil {
		t.Error(err)
	}
}

func TestEvent(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	definition := `
	function test_ev()
		contract.event("ev1", 1,"local", 2, "form")
		contract.event("ev1", 3,"local", 4, "form")
	end
	abi.register(test_ev)`

	err = bc.ConnectBlock(
		NewLuaTxAccount("ktlee", 100),
		NewLuaTxDef("ktlee", "event", 0, definition),
	)
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "event", 0, `{"Name": "test_ev", "Args":[]}`),
	)
	if err != nil {
		t.Error(err)
	}

}

// end of test-cases
