package contract

import (
	"encoding/json"
	"fmt"
	"github.com/aergoio/aergo/types"
	"strings"
	"testing"
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
	exRv := fmt.Sprintf(`["Amg6nZWXKB6YpNgBPv9atcjdm6hnFvs5wMdRgb2e9DmaF5g9muF2","4huAuw28LdAg9nKji5t1EGSkZ3ScvnyZwH2KBZCKejqHJ","AmhNNBNY7XFk4p5ym4CJf8nTcRTEHjWzAeXJfhP71244CjBCAQU3",%d,3,999]`, bc.cBlock.Header.Timestamp/1e9)
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
		`[{"count":"0","name":"candidate1","id":0},{"count":"0","name":"candidate2","id":1},{"count":"0","name":"candidate3","id":2}]`,
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
		`[{"count":"0","name":"candidate1","id":0},{"count":"0","name":"candidate2","id":1},{"count":"0","name":"candidate3","id":2}]`,
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
		`[{"count":"0","name":"candidate1","id":0},{"count":"0","name":"candidate2","id":1},{"count":"0","name":"candidate3","id":2}]`,
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
		`[{"count":"2","name":"candidate1","id":0},{"count":"0","name":"candidate2","id":1},{"count":"0","name":"candidate3","id":2}]`,
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

func TestMemSize(t *testing.T) {
	bc, err := LoadDummyChain()
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}

	definition := `
function infiniteLoop()
	local s = "hello world"
	while true do
		s = s .. s
	end
	return s
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
	errMsg := "not enough memory"
	if err == nil {
		t.Errorf("expected: %s", errMsg)
	}
	if err != nil && !strings.Contains(err.Error(), errMsg) {
		t.Error(err)
	}
}

func TestRecursiveData(t *testing.T) {
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
	err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `{"key1":{"arg2":{},"arg3":{},"arg1":1},"key2":[5,4,3]}`)
	if err != nil {
		t.Error(err)
	}
	err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"{\"key1\":{\"arg2\":{},\"arg3\":{},\"arg1\":1},\"key2\":[5,4,3]}"`)
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
			`{"Name":"set", "Args":["{\"key1\":[1,2,3], \"key1\":5}}"]}`).fail("not proper json format"),
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
		NewLuaTxCall("ktlee", "array", 0, `{"Name":"inc", "Args":[0]}`).fail("index out of range"),
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
	end

	function pkins1()
	end

	function pkins2()
	end

	function pkget()
		return 2
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
		`[[1,2,3],{"name":"kslee","age":39},7]`,
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
		`[[[1,2,3],["first","second"]],{"address":{"state":"XXX-do","city":"YYY-si"},"age":39,"name":"kslee"},"end"]`,
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
		`[[{"name":"wook","age":50},{"name":"hook","age":42}],{"scores":[10,20,30,40,50],"age":39,"name":"kslee"},"hmm..."]`,
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
		NewLuaTxCall("ktlee", "a", 0, `{"Name":"setCount", "Args":["1", 20]}`).fail("number expected, got string)"),
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
		).fail(`bad argument #2 to '__newindex' (number expected, got boolean)`),
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
		NewLuaTxCall("ktlee", "x", 0, `{"Name":"setCount", "Args":[1, 20]}`).fail("string expected, got number)"),
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
		NewLuaTxCall("ktlee", "dupVar1", 0, `{"Name": "Work"}`).fail("duplicated variable: 'Var1'"),
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
	err = bc.Query("crypto", `{"Name": "get", "Args" : ["ab\u0000\u442a"]}`, "", `"c58f6dca13e4bba90a326d8605042862fe87c63a64a9dd0e95608a2ee68dc6f0"`)
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
		t.Error(err)
	} else {
		if !strings.Contains(err.Error(), types.ErrVmConstructorIsNotPayable.Error()) {
			t.Error(err)
		}
	}
	err = bc.ConnectBlock(
		NewLuaTxCall("ktlee", "payable", 0, `{"Name":"save", "Args": ["blahblah"]}`).fail("cannot find contract "),
	)
	if err != nil {
		t.Error(err)
	}
	err = bc.ConnectBlock(
		NewLuaTxDef("ktlee", "payable", 0, src),
		NewLuaTxCall("ktlee", "payable", 0, `{"Name":"save", "Args": ["blahblah"]}`).fail("cannot find contract "),
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
		NewLuaTxCall("ktlee", "default", 1, "").fail(`'default' is not payable`),
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

function constructor()
end

abi.register(test, sendS, testBignum, argBignum, calladdBignum)
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
}

func TestSqlVmPubNet(t *testing.T) {
	PubNet = true
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
		NewLuaTxCall("ktlee", "simple-query", 0, `{"Name": "createAndInsert", "Args":[]}`).fail(`attempt to index global 'db'`),
	)
	if err != nil {
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

// end of test-cases
