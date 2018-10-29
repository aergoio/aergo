package contract

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aergoio/aergo-lib/db"
	"github.com/aergoio/aergo/cmd/aergocli/util"
	"github.com/aergoio/aergo/state"
	"github.com/aergoio/aergo/types"
	"github.com/minio/sha256-simd"
)

const (
	helloCode = `function hello(say) return "Hello " .. say end abi.register(hello)`

	systemCode = `function testState()
		string.format("creator: %s",system.getContractID())
		string.format("timestamp: %d",system.getTimestamp())
		string.format("blockheight: %d",system.getBlockheight())
		system.setItem("key1", 999)
		string.format("getitem : %s",system.getItem("key1"))
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
	abi.register(inc, query)`
)

func TestReturn(t *testing.T) {
	bc := loadBlockChain(t)

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
		newLuaTxDef("ktlee", "return_num", 10, "function return_num() return 10 end abi.register(return_num)"),
		newLuaTxCall("ktlee", "return_num", 10, `{"Name":"return_num", "Args":[]}`),
	)

	err := bc.query("return_num", `{"Name":"return_num", "Args":[]}`, "", "10")
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

	bc.connectBlock(
		newLuaTxDef("ktlee", "foo", 1, foo),
	)

	err = bc.query("foo", `{"Name":"foo", "Args":[]}`, "", "[1,2,3]")
	if err != nil {
		t.Error(err)
	}
	err = bc.query("foo", `{"Name":"foo2", "Args":["foo314"]}`, "", `"foo314"`)
	if err != nil {
		t.Error(err)
	}
}

func TestContractHello(t *testing.T) {
	bc := loadBlockChain(t)

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
	)
	bc.connectBlock(
		newLuaTxDef("ktlee", "hello", 1, helloCode),
	)
	tx := newLuaTxCall("ktlee", "hello", 1, `{"Name":"hello", "Args":["World"]}`)
	bc.connectBlock(tx)
	receipt := bc.getReceipt(tx.hash())
	if receipt.GetRet() != `"Hello World"` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
}

func TestContractSystem(t *testing.T) {
	bc := loadBlockChain(t)

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
	)
	bc.connectBlock(
		newLuaTxDef("ktlee", "system", 1, systemCode),
	)
	tx := newLuaTxCall("ktlee", "system", 1, `{"Name":"testState", "Args":[]}`)
	bc.connectBlock(tx)
	receipt := bc.getReceipt(tx.hash())
	exRv := fmt.Sprintf(`["Amg6nZWXKB6YpNgBPv9atcjdm6hnFvs5wMdRgb2e9DmaF5g9muF2","0c7902699be42c8a8e46fbbb4501726517e86b22c56a189f7625a6da49081b2451","AmhNNBNY7XFk4p5ym4CJf8nTcRTEHjWzAeXJfhP71244CjBCAQU3",%d,3,999]`, bc.cBlock.Header.Timestamp)
	if receipt.GetRet() != exRv {
		t.Errorf("expected: %s, but got: %s", exRv, receipt.GetRet())
	}
}

func TestGetABI(t *testing.T) {
	bc := loadBlockChain(t)

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
		newLuaTxDef("ktlee", "hello", 1,
			`function hello(say) return "Hello " .. say end abi.register(hello)`),
	)
	abi, err := bc.getABI("hello")
	if err != nil {
		t.Error(err)
	}
	b, err := json.Marshal(abi)
	if err != nil {
		t.Error(err)
	}
	if string(b) != `{"version":"0.1","language":"lua","functions":[{"name":"hello","arguments":[{"name":"say"}]}]}` {
		t.Error(string(b))
	}
}

func TestContractQuery(t *testing.T) {
	bc := loadBlockChain(t)

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
	)
	bc.connectBlock(
		newLuaTxDef("ktlee", "query", 1, queryCode),
		newLuaTxCall("ktlee", "query", 1, `{"Name":"inc", "Args":[]}`),
	)

	ktlee, err := bc.getAccountState("ktlee")
	if err != nil {
		t.Error(err)
	}
	if ktlee.Balance != uint64(98) {
		t.Error(ktlee.Balance)
	}
	query, err := bc.getAccountState("query")
	if err != nil {
		t.Error(err)
	}
	if query.Balance != uint64(2) {
		t.Error(query.Balance)
	}

	err = bc.query("query", `{"Name":"inc", "Args":[]}`, "set not permitted in query", "")
	if err != nil {
		t.Error(err)
	}

	err = bc.query("query", `{"Name":"query", "Args":["key1"]}`, "", "1")
	if err != nil {
		t.Error(err)
	}
}

func TestRollback(t *testing.T) {
	bc := loadBlockChain(t)

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
	)
	bc.connectBlock(
		newLuaTxDef("ktlee", "query", 1, queryCode),
		newLuaTxCall("ktlee", "query", 1, `{"Name":"inc", "Args":[]}`),
	)
	bc.connectBlock(
		newLuaTxCall("ktlee", "query", 1, `{"Name":"inc", "Args":[]}`),
		newLuaTxCall("ktlee", "query", 1, `{"Name":"inc", "Args":[]}`),
	)
	bc.connectBlock(
		newLuaTxCall("ktlee", "query", 1, `{"Name":"inc", "Args":[]}`),
		newLuaTxCall("ktlee", "query", 1, `{"Name":"inc", "Args":[]}`),
	)

	err := bc.query("query", `{"Name":"query", "Args":["key1"]}`, "", "5")
	if err != nil {
		t.Error(err)
	}

	err = bc.disconnectBlock()
	if err != nil {
		t.Error(err)
	}
	err = bc.query("query", `{"Name":"query", "Args":["key1"]}`, "", "3")
	if err != nil {
		t.Error(err)
	}

	err = bc.disconnectBlock()
	if err != nil {
		t.Error(err)
	}

	err = bc.query("query", `{"Name":"query", "Args":["key1"]}`, "", "1")
	if err != nil {
		t.Error(err)
	}

	bc.connectBlock(
		newLuaTxCall("ktlee", "query", 1, `{"Name":"inc", "Args":[]}`),
	)

	err = bc.query("query", `{"Name":"query", "Args":["key1"]}`, "", "2")
	if err != nil {
		t.Error(err)
	}
}

func TestVote(t *testing.T) {
	bc := loadBlockChain(t)

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

	bc.connectBlock(
		newLuaTxAccount("owner", 100),
		newLuaTxDef("owner", "vote", 1, definition),
		newLuaTxAccount("user1", 100),
	)

	err := bc.connectBlock(
		newLuaTxCall(
			"owner",
			"vote",
			1,
			`{"Name":"addCandidate", "Args":["candidate1"]}`,
		),
		newLuaTxCall(
			"owner",
			"vote",
			1,
			`{"Name":"addCandidate", "Args":["candidate2"]}`,
		),
		newLuaTxCall(
			"owner",
			"vote",
			1,
			`{"Name":"addCandidate", "Args":["candidate3"]}`,
		),
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.query(
		"vote",
		`{"Name":"getCandidates"}`,
		"",
		`[{"count":"0","name":"candidate1","id":0},{"count":"0","name":"candidate2","id":1},{"count":"0","name":"candidate3","id":2}]`,
	)
	if err != nil {
		t.Error(err)
	}

	bc.connectBlock(
		newLuaTxCall(
			"user1",
			"vote",
			1,
			`{"Name":"addCandidate", "Args":["candidate4"]}`,
		),
	)
	err = bc.query(
		"vote",
		`{"Name":"getCandidates"}`,
		"",
		`[{"count":"0","name":"candidate1","id":0},{"count":"0","name":"candidate2","id":1},{"count":"0","name":"candidate3","id":2}]`,
	)
	if err != nil {
		t.Error(err)
	}

	bc.connectBlock(
		// register voter
		newLuaTxCall(
			"owner",
			"vote",
			1,
			fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(strHash("user10"))),
		),
		newLuaTxCall(
			"owner",
			"vote",
			1,
			fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(strHash("user10"))),
		),
		newLuaTxCall(
			"owner",
			"vote",
			1,
			fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(strHash("user11"))),
		),
		newLuaTxCall(
			"owner",
			"vote",
			1,
			fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, types.EncodeAddress(strHash("user1"))),
		),
		// vote
		newLuaTxCall(
			"user1",
			"vote",
			1,
			`{"Name":"vote", "Args":["user1"]}`,
		),
		newLuaTxCall(
			"user1",
			"vote",
			1,
			`{"Name":"vote", "Args":["user1"]}`,
		),
		newLuaTxCall(
			"user1",
			"vote",
			1,
			`{"Name":"vote", "Args":["user2"]}`,
		),
		newLuaTxCall(
			"user1",
			"vote",
			1,
			`{"Name":"vote", "Args":["user2"]}`,
		),
		newLuaTxCall(
			"user1",
			"vote",
			1,
			`{"Name":"vote", "Args":["user3"]}`,
		),
	)

	err = bc.query(
		"vote",
		`{"Name":"getCandidates"}`,
		"",
		`[{"count":"0","name":"candidate1","id":0},{"count":"0","name":"candidate2","id":1},{"count":"0","name":"candidate3","id":2}]`,
	)
	if err != nil {
		t.Error(err)
	}

	bc.connectBlock(
		newLuaTxCall(
			"user11",
			"vote",
			1,
			`{"Name":"vote", "Args":["candidate1"]}`,
		),
		newLuaTxCall(
			"user10",
			"vote",
			1,
			`{"Name":"vote", "Args":["candidate1"]}`,
		),
	)

	err = bc.query(
		"vote",
		`{"Name":"getCandidates"}`,
		"",
		`[{"count":"2","name":"candidate1","id":0},{"count":"0","name":"candidate2","id":1},{"count":"0","name":"candidate3","id":2}]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestInfiniteLoop(t *testing.T) {
	bc := loadBlockChain(t)

	definition := `
function infiniteLoop()
    db.exec("create table if not exists dual(dummy int)")
	for i = 1, 100000000000000 do
		system.setItem("key_"..i, "value_"..i)
		db.exec("insert into dual values ("..tostring(i)..")")
	end
end
abi.register(infiniteLoop)`

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
		newLuaTxDef("ktlee", "loop", 1, definition),
		newLuaTxCall(
			"ktlee",
			"loop",
			1,
			`{"Name":"infiniteLoop"}`,
		).fail("exceeded the maximum instruction count"),
	)
}

func TestSqlVmSimple(t *testing.T) {
	bc := loadBlockChain(t)

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

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
		newLuaTxDef("ktlee", "simple-query", 1, definition),
	)
	bc.connectBlock(
		newLuaTxCall("ktlee", "simple-query", 1, `{"Name": "createAndInsert", "Args":[]}`),
	)
	err := bc.query(
		"simple-query",
		`{"Name": "query", "Args":[]}`,
		"",
		`[2,3.1,"X Hello Blockchain",2,3.1,"Y Hello Blockchain",2,3.1,"Z Hello Blockchain"]`,
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.query(
		"simple-query",
		`{"Name": "count", "Args":[]}`,
		"",
		`3`,
	)
	if err != nil {
		t.Error(err)
	}

	bc.connectBlock(
		newLuaTxCall("ktlee", "simple-query", 1, `{"Name": "createAndInsert", "Args":[]}`),
	)
	err = bc.query(
		"simple-query",
		`{"Name": "count", "Args":[]}`,
		"",
		`6`,
	)
	if err != nil {
		t.Error(err)
	}

	bc.disconnectBlock()

	err = bc.query(
		"simple-query",
		`{"Name": "count", "Args":[]}`,
		"",
		`3`,
	)
	if err != nil {
		t.Error(err)
	}

	err = bc.disconnectBlock()
	if err != nil {
		t.Error(err)
	}
	err = bc.disconnectBlock()
	if err != nil {
		t.Error(err)
	}

	// there is only a genesis block
	err = bc.query(
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
	bc := loadBlockChain(t)

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

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
		newLuaTxDef("ktlee", "fail", 1, definition),
		newLuaTxCall("ktlee", "fail", 1, `{"Name":"init"}`),
	)

	bc.connectBlock(
		newLuaTxCall("ktlee", "fail", 1, `{"Name":"add", "Args":[1]}`),
	)

	err := bc.connectBlock(
		newLuaTxCall("ktlee", "fail", 1, `{"Name":"add", "Args":[2]}`),
		newLuaTxCall("ktlee", "fail", 1, `{"Name":"addFail", "Args":[3]}`).
			fail(`near "set": syntax error`),
		newLuaTxCall("ktlee", "fail", 1, `{"Name":"add", "Args":[4]}`),
	)
	if err != nil {
		t.Error(err)
	}

	bc.connectBlock(
		newLuaTxCall("ktlee", "fail", 1, `{"Name":"add", "Args":[5]}`),
	)

	err = bc.query("fail", `{"Name":"get"}`, "", "12")
	if err != nil {
		t.Error(err)
	}

	bc.disconnectBlock()

	err = bc.query("fail", `{"Name":"get"}`, "", "7")
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmDateTime(t *testing.T) {
	bc := loadBlockChain(t)
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

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
		newLuaTxDef("ktlee", "datetime", 1, definition),
		newLuaTxCall("ktlee", "datetime", 1, `{"Name":"init"}`),
	)

	bc.connectBlock(
		newLuaTxCall("ktlee", "datetime", 1, `{"Name":"nowNull"}`),
	)

	bc.connectBlock(
		newLuaTxCall("ktlee", "datetime", 1, `{"Name":"localtimeNull"}`),
	)

	err := bc.query(
		"datetime",
		`{"Name":"get"}`,
		"",
		`[{"bool":0},{"bool":1},{"date":"1970-01-01 02:46:40","bool":1},{"date":"2004-11-23","bool":0}]`,
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlConstrains(t *testing.T) {
	bc := loadBlockChain(t)

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

	err := bc.connectBlock(
		newLuaTxAccount(
			"ktlee",
			100,
		),
		newLuaTxDef(
			"ktlee",
			"constraint",
			1,
			definition,
		),
		newLuaTxCall(
			"ktlee",
			"constraint",
			1,
			`{"Name":"init"}`,
		),
		newLuaTxCall(
			"ktlee",
			"constraint",
			1,
			`{"Name":"pkFail"}`,
		).fail("UNIQUE constraint failed: r.id"),
		newLuaTxCall(
			"ktlee",
			"constraint",
			1,
			`{"Name":"checkFail"}`,
		).fail("CHECK constraint failed: r"),
		newLuaTxCall(
			"ktlee",
			"constraint",
			1,
			`{"Name":"fkFail"}`,
		).fail("FOREIGN KEY constraint failed"),
		newLuaTxCall(
			"ktlee",
			"constraint",
			1,
			`{"Name":"notNullFail"}`,
		).fail("NOT NULL constraint failed: r.nonull"),
		newLuaTxCall(
			"ktlee",
			"constraint",
			1,
			`{"Name":"uniqueFail"}`,
		).fail("UNIQUE constraint failed: r.only"),
	)
	if err != nil {
		t.Error(err)
	}
}

func TestSqlVmCustomer(t *testing.T) {
	bc := loadBlockChain(t)

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

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
		newLuaTxDef("ktlee", "customer", 1, definition),
		newLuaTxCall(
			"ktlee",
			"customer",
			1,
			`{"Name":"createTable"}`,
		),
	)

	bc.connectBlock(
		newLuaTxCall(
			"ktlee",
			"customer",
			1,
			`{"Name":"insert", "Args":["id1","passwd1","name1","20180524","010-1234-5678"]}`,
		),
	)

	bc.connectBlock(
		newLuaTxCall(
			"ktlee",
			"customer",
			1,
			`{"Name":"insert", "Args":["id2","passwd2","name2","20180524","010-1234-5678"]}`,
		),
	)

	bc.connectBlock(
		newLuaTxCall(
			"ktlee",
			"customer",
			1,
			`{"Name":"update", "Args":["id2","passwd3"]}`,
		),
	)

	bc.query("customer", `{"Name":"count"}`, "", "[2]")

	bc.disconnectBlock()

	bc.query(
		"customer",
		`{"Name":"query", "Args":["id2"]}`,
		"",
		`[{"birth":"20180524","id":"id2","mobile":"010-1234-5678","name":"name2","passwd":"passwd2"}]`,
	)

	bc.connectBlock(
		newLuaTxCall(
			"ktlee",
			"customer",
			1,
			`{"Name":"delete", "Args":["id2"]}`,
		),
	)

	bc.query(
		"customer",
		`{"Name":"query", "Args":["id2"]}`,
		"",
		`{}`,
	)
}

func TestSqlVmDataType(t *testing.T) {
	bc := loadBlockChain(t)

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

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
		newLuaTxDef("ktlee", "datatype", 1, definition),
		newLuaTxCall(
			"ktlee",
			"datatype",
			1,
			`{"Name":"createDataTypeTable"}`,
		),
	)

	bc.connectBlock(
		newLuaTxCall(
			"ktlee",
			"datatype",
			1,
			`{"Name":"insertDataTypeTable"}`,
		),
		newLuaTxCall(
			"ktlee",
			"datatype",
			1,
			`{"Name":"insertDataTypeTable"}`,
		),
		newLuaTxCall(
			"ktlee",
			"datatype",
			1,
			`{"Name":"insertDataTypeTable"}`,
		),
	)

	bc.connectBlock(
		newLuaTxCall(
			"ktlee",
			"datatype",
			1,
			`{"Name":"insertDataTypeTable"}`,
		),
	)

	bc.query(
		"datatype",
		`{"Name":"queryOrderByDesc"}`,
		"",
		`[{"blockheight1":0,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"},{"blockheight1":0,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"},{"blockheight1":0,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"},{"blockheight1":0,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"}]`,
	)

	bc.query(
		"datatype",
		`{"Name":"queryGroupByBlockheight1"}`,
		"",
		`[{"avg_float1":3.14,"blockheight1":0,"count1":4,"sum_int1":4}]`,
	)
}

func TestSqlVmFunction(t *testing.T) {
	bc := loadBlockChain(t)

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

	bc.connectBlock(
		newLuaTxAccount("name", 100),
		newLuaTxDef("ktlee", "fns", 1, definition),
	)

	bc.query("fns", `{"Name":"sql_func"}`, "", `[3,1,6]`)

	bc.query("fns", `{"Name":"abs_func"}`, "", `[1,0,1]`)

	bc.query("fns", `{"Name":"typeof_func"}`,
		"", `["integer","text","real","null"]`,
	)
}

func TestSqlVmBook(t *testing.T) {
	bc := loadBlockChain(t)

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

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
		newLuaTxDef("ktlee", "book", 1, definition),
		newLuaTxCall(
			"ktlee",
			"book",
			1,
			`{"Name":"createTable"}`,
		),
	)

	bc.connectBlock(
		newLuaTxCall(
			"ktlee",
			"book",
			1,
			`{"Name":"makeBook"}`,
		),
	)

	bc.connectBlock(
		newLuaTxCall(
			"ktlee",
			"book",
			1,
			`{"Name":"copyBook"}`,
		),
	)

	bc.query(
		"book",
		`{"Name":"viewCopyBook"}`,
		"",
		`[100,"value=1"]`,
	)

	bc.query(
		"book",
		`{"Name":"viewCopyBook"}`,
		"",
		`[10,10,"value=100"]`,
	)
}

func TestSqlVmDateformat(t *testing.T) {
	bc := loadBlockChain(t)

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

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
		newLuaTxDef(
			"ktlee",
			"data_format",
			1,
			definition,
		),
		newLuaTxCall("ktlee", "data_format", 1, `{"Name":"init"}`),
	)

	bc.query(
		"data_format",
		`{"Name":"get"}`,
		"",
		`[["2004-10-24T00:00:00Z","2004-10-24T11:11:11Z","20041024111111"],["2018-05-28T00:00:00Z","2018-05-28T10:45:38Z","20180528104538"]]`,
	)
}

func TestSqlVmRecursiveData(t *testing.T) {
	bc := loadBlockChain(t)

	definition := `
function r()
	local t = {}
	t["name"] = "ktlee"
	t["self"] = t
	return t
end
abi.register(r)`

	tx := newLuaTxCall("ktlee", "r", 1, `{"Name":"r"}`)
	err := bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
		newLuaTxDef("ktlee", "r", 1, definition),
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

	bc := loadBlockChain(t)

	bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
		newLuaTxDef("ktlee", "counter", 10, definition1).constructor("[1]"),
		newLuaTxCall("ktlee", "counter", 10, `{"Name":"inc", "Args":[]}`),
	)

	err := bc.query("counter", `{"Name":"get", "Args":[]}`, "", "2")
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
	bc.connectBlock(
		newLuaTxDef("ktlee", "caller", 10, definition2).
			constructor(fmt.Sprintf(`["%s"]`, types.EncodeAddress(strHash("counter")))),
		newLuaTxCall("ktlee", "caller", 10, `{"Name":"add", "Args":[]}`),
	)
	err = bc.query("caller", `{"Name":"get", "Args":[]}`, "", "3")
	if err != nil {
		t.Error(err)
	}
	err = bc.query("caller", `{"Name":"dget", "Args":[]}`, "", "99")
	if err != nil {
		t.Error(err)
	}
	tx := newLuaTxCall("ktlee", "caller", 10, `{"Name":"dadd", "Args":[]}`)
	bc.connectBlock(tx)
	receipt := bc.getReceipt(tx.hash())
	if receipt.GetRet() != `99` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	tx = newLuaTxCall("ktlee", "caller", 10, `{"Name":"dadd", "Args":[]}`)
	bc.connectBlock(tx)
	receipt = bc.getReceipt(tx.hash())
	if receipt.GetRet() != `100` {
		t.Errorf("contract Call ret error :%s", receipt.GetRet())
	}
	err = bc.query("caller", `{"Name":"get", "Args":[]}`, "", "3")
	if err != nil {
		t.Error(err)
	}
}

func TestSparseTable(t *testing.T) {
	bc := loadBlockChain(t)

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

	tx := newLuaTxCall("ktlee", "r", 1, `{"Name":"r"}`)
	err := bc.connectBlock(
		newLuaTxAccount("ktlee", 100),
		newLuaTxDef("ktlee", "r", 1, definition),
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

// end of test-cases

// helper functions

type blockChain struct {
	sdb           *state.ChainStateDB
	bestBlock     *types.Block
	cBlock        *types.Block
	bestBlockNo   types.BlockNo
	bestBlockId   types.BlockID
	blockIds      []types.BlockID
	blocks        []*types.Block
	testReceiptDB db.DB
}

func loadBlockChain(t *testing.T) *blockChain {
	bc := &blockChain{sdb: state.NewChainStateDB()}
	dataPath, err := ioutil.TempDir("", "data")
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	err = bc.sdb.Init(string(db.BadgerImpl), dataPath, nil, false)
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	genesis := types.GetTestGenesis()
	bc.sdb.SetGenesis(genesis)
	bc.bestBlockNo = genesis.Block.BlockNo()
	bc.bestBlockId = genesis.Block.BlockID()
	bc.blockIds = append(bc.blockIds, bc.bestBlockId)
	bc.blocks = append(bc.blocks, genesis.Block)
	bc.testReceiptDB = db.NewDB(db.BadgerImpl, path.Join(dataPath, "receiptDB"))
	LoadDatabase(dataPath) // sql database
	return bc
}

func (bc *blockChain) newBState() *state.BlockState {
	b := types.Block{
		Header: &types.BlockHeader{
			PrevBlockHash: []byte(bc.bestBlockId.String()),
			BlockNo:       bc.bestBlockNo + 1,
			Timestamp:     time.Now().Unix(),
		},
	}
	bc.cBlock = &b
	// blockInfo := types.NewBlockInfo(b.BlockNo(), b.BlockID(), bc.bestBlockId)
	return state.NewBlockState(bc.sdb.OpenNewStateDB(bc.sdb.GetRoot()))
}

func (bc *blockChain) BeginReceiptTx() db.Transaction {
	return bc.testReceiptDB.NewTx()
}

func (bc *blockChain) getABI(contract string) (*types.ABI, error) {
	cState, err := bc.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID(strHash(contract)))
	if err != nil {
		return nil, err
	}
	return GetABI(cState)
}

func (bc *blockChain) getReceipt(txHash []byte) *types.Receipt {
	r := new(types.Receipt)
	r.UnmarshalBinary(bc.testReceiptDB.Get(txHash))
	return r
}

func (bc *blockChain) getAccountState(name string) (*types.State, error) {
	return bc.sdb.GetStateDB().GetAccountState(types.ToAccountID(strHash(name)))
}

type luaTx interface {
	run(bs *state.BlockState, blockNo uint64, ts int64, receiptTx db.Transaction) error
}

type luaTxAccount struct {
	name    []byte
	balance uint64
}

func newLuaTxAccount(name string, balance uint64) *luaTxAccount {
	return &luaTxAccount{
		name:    strHash(name),
		balance: balance,
	}
}

func (l *luaTxAccount) run(bs *state.BlockState, blockNo uint64, ts int64,
	receiptTx db.Transaction) error {

	id := types.ToAccountID(l.name)
	accountState, err := bs.GetAccountState(id)
	if err != nil {
		return err
	}
	updatedAccountState := types.State(*accountState)
	updatedAccountState.Balance = l.balance
	bs.PutState(id, &updatedAccountState)
	return nil
}

type luaTxCommon struct {
	sender   []byte
	contract []byte
	amount   uint64
	code     []byte
	id       uint64
}

type luaTxDef struct {
	luaTxCommon
}

func newLuaTxDef(sender, contract string, amount uint64, code string) *luaTxDef {
	luac := exec.Command("../bin/aergoluac", "--payload")
	stdin, err := luac.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, code)
	}()
	out, err := luac.Output()
	if err != nil {
		log.Fatal(err)
	}
	b, err := util.DecodeCode(string(out))
	if err != nil {
		log.Fatal(err)
	}
	codeWithInit := make([]byte, 4+len(b))
	binary.LittleEndian.PutUint32(codeWithInit, uint32(4+len(b)))
	copy(codeWithInit[4:], b)
	return &luaTxDef{
		luaTxCommon{
			sender:   strHash(sender),
			contract: strHash(contract),
			code:     codeWithInit,
			amount:   amount,
			id:       newTxId(),
		},
	}
}

func strHash(d string) []byte {
	h := sha256.New()
	h.Write([]byte(d))
	b := h.Sum(nil)
	b = append([]byte{0x0C}, b...)
	return b
}

var luaTxId uint64 = 0

func newTxId() uint64 {
	luaTxId++
	return luaTxId
}

func (l *luaTxDef) hash() []byte {
	h := sha256.New()
	h.Write([]byte(strconv.FormatUint(l.id, 10)))
	b := h.Sum(nil)
	b = append([]byte{0x0C}, b...)
	return b
}

func (l *luaTxDef) constructor(args string) *luaTxDef {
	argsLen := len([]byte(args))
	if argsLen == 0 {
		return l
	}

	code := make([]byte, len(l.code)+argsLen)
	codeLen := copy(code[0:], l.code)
	binary.LittleEndian.PutUint32(code[0:], uint32(codeLen))
	copy(code[codeLen:], []byte(args))

	l.code = code

	return l
}

func contractFrame(l *luaTxCommon, bs *state.BlockState,
	run func(s, c *types.State, id types.AccountID, cs *state.ContractState) error) error {

	creatorId := types.ToAccountID(l.sender)
	creatorState, err := bs.GetAccountState(creatorId)
	if err != nil {
		return err
	}

	contractId := types.ToAccountID(l.contract)
	contractState, err := bs.GetAccountState(contractId)
	if err != nil {
		return err
	}

	uContractState := types.State(*contractState)
	eContractState, err := bs.OpenContractState(&uContractState)
	if err != nil {
		return err
	}

	err = run(creatorState, &uContractState, contractId, eContractState)
	if err != nil {
		return err
	}

	uCallerState := types.State(*creatorState)
	uCallerState.Balance -= l.amount
	uContractState.Balance += l.amount

	bs.PutState(creatorId, &uCallerState)
	bs.PutState(contractId, &uContractState)
	return nil

}

func (l *luaTxDef) run(bs *state.BlockState, blockNo uint64, ts int64,
	receiptTx db.Transaction) error {

	return contractFrame(&l.luaTxCommon, bs,
		func(senderState, uContractState *types.State, contractId types.AccountID, eContractState *state.ContractState) error {
			uContractState.SqlRecoveryPoint = 1
			sqlTx, err := BeginTx(contractId, uContractState.SqlRecoveryPoint)
			if err != nil {
				return err
			}
			err = sqlTx.Savepoint()
			if err != nil {
				return err
			}

			bcCtx := NewContext(bs, senderState, eContractState,
				types.EncodeAddress(l.sender), hex.EncodeToString(l.hash()), blockNo, ts,
				"", 1, types.EncodeAddress(l.contract),
				0, nil, sqlTx.GetHandle(), ChainService)

			_, err = Create(eContractState, l.code, l.contract, bcCtx)
			if err != nil {
				_ = sqlTx.RollbackToSavepoint()
				return err
			}
			err = bs.CommitContractState(eContractState)
			if err != nil {
				_ = sqlTx.RollbackToSavepoint()
				return err
			}
			err = sqlTx.Release()
			if err != nil {
				return err
			}
			return nil
		},
	)
}

type luaTxCall struct {
	luaTxCommon
	expectedErr string
}

func newLuaTxCall(sender, contract string, amount uint64, code string) *luaTxCall {
	return &luaTxCall{
		luaTxCommon: luaTxCommon{
			sender:   strHash(sender),
			contract: strHash(contract),
			amount:   amount,
			code:     []byte(code),
			id:       newTxId(),
		},
	}
}

func (l *luaTxCall) hash() []byte {
	h := sha256.New()
	h.Write([]byte(strconv.FormatUint(l.id, 10)))
	b := h.Sum(nil)
	b = append([]byte{0x0C}, b...)
	return b
}

func (l *luaTxCall) fail(expectedErr string) *luaTxCall {
	l.expectedErr = expectedErr
	return l
}

func (l *luaTxCall) run(bs *state.BlockState, blockNo uint64, ts int64, receiptTx db.Transaction) error {

	err := contractFrame(&l.luaTxCommon, bs,
		func(senderState, uContractState *types.State, contractId types.AccountID, eContractState *state.ContractState) error {
			sqlTx, err := BeginTx(contractId, uContractState.SqlRecoveryPoint)
			if err != nil {
				return err
			}
			sqlTx.Savepoint()

			bcCtx := NewContext(bs, senderState, eContractState,
				types.EncodeAddress(l.sender), hex.EncodeToString(l.hash()), blockNo, ts,
				"", 1, types.EncodeAddress(l.contract),
				0, nil, sqlTx.GetHandle(), ChainService)

			rv, err := Call(eContractState, l.code, l.contract, bcCtx)
			if err != nil {
				_ = sqlTx.RollbackToSavepoint()
				return err
			}
			err = bs.CommitContractState(eContractState)
			if err != nil {
				r := types.NewReceipt(l.contract, err.Error(), "")
				b, _ := r.MarshalBinary()
				receiptTx.Set(l.hash(), b)
				_ = sqlTx.RollbackToSavepoint()
				return err
			}
			err = sqlTx.Release()
			if err != nil {
				return err
			}

			r := types.NewReceipt(l.contract, "SUCCESS", rv)
			b, _ := r.MarshalBinary()
			receiptTx.Set(l.hash(), b)

			return nil
		},
	)
	if l.expectedErr != "" {
		if err == nil || !strings.Contains(err.Error(), l.expectedErr) {
			return err
		}
		return nil
	}
	return err
}

func (bc *blockChain) connectBlock(txs ...luaTx) error {
	blockState := bc.newBState()
	tx := bc.BeginReceiptTx()
	defer tx.Commit()

	for _, x := range txs {
		if err := x.run(blockState, bc.cBlock.Header.BlockNo, bc.cBlock.Header.Timestamp, tx); err != nil {
			return err
		}
	}
	err := SaveRecoveryPoint(blockState)
	if err != nil {
		return err
	}
	err = bc.sdb.Apply(blockState)
	if err != nil {
		return err
	}
	//FIXME newblock must be created after sdb.apply()
	bc.cBlock.SetBlocksRootHash(bc.sdb.GetRoot())
	bc.bestBlockNo = bc.bestBlockNo + 1
	bc.bestBlockId = types.ToBlockID(bc.cBlock.GetHash())
	bc.blockIds = append(bc.blockIds, bc.bestBlockId)
	bc.blocks = append(bc.blocks, bc.cBlock)

	return nil
}

func (bc *blockChain) disconnectBlock() error {
	if len(bc.blockIds) == 1 {
		return errors.New("genesis block")
	}
	bc.bestBlockNo--
	bc.blockIds = bc.blockIds[0 : len(bc.blockIds)-1]
	bc.blocks = bc.blocks[0 : len(bc.blocks)-1]
	bc.bestBlockId = bc.blockIds[len(bc.blockIds)-1]

	bestBlock := bc.blocks[len(bc.blocks)-1]

	var sroot []byte
	if bestBlock != nil {
		sroot = bestBlock.GetHeader().GetBlocksRootHash()
	}
	return bc.sdb.Rollback(sroot)
}

func (bc *blockChain) query(contract, queryInfo string, expectedErr, expectedRv string) error {
	cState, err := bc.sdb.GetStateDB().OpenContractStateAccount(types.ToAccountID(strHash(contract)))
	if err != nil {
		return err
	}
	rv, err := Query(strHash(contract), bc.newBState(), cState, []byte(queryInfo))
	if expectedErr != "" {
		if err == nil || !strings.Contains(err.Error(), expectedErr) {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}

	if expectedRv != string(rv) {
		return fmt.Errorf("expected: %s, but got: %s", expectedRv, string(rv))
	}
	return nil
}
