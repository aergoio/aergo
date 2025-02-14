// vm_dummy has two types of unit tests; tests that can be run only in the Linux/amd64
// environment where offical public chains such as Aergo mainnet and testnet are running,
// and tests that can be run regardless of architecture.
// The tests in vm_dummy_pub_test.go are architecture dependent tests.
package vm_dummy

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/aergoio/aergo/v2/contract"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipNotOnAmd64 check if test is run on amd64 architecture, otherwise skip test
func skipNotOnAmd64(t *testing.T) {
	if runtime.GOARCH != "amd64" {
		t.Skipf("%s: skip architecture dependent test", t.Name())
	}
}

func TestContractSendF(t *testing.T) {
	skipNotOnAmd64(t)

	code := readLuaCode(t, "contract_sendf_1.lua")
	code2 := readLuaCode(t, "contract_sendf_2.lua")

	// skip if current version is less than 3
	if currentVersion < 3 {
		t.Skipf("%s: skip version less than 3", t.Name())
	}

		bc, err := LoadDummyChain(SetHardForkVersion(currentVersion), RunOnPubNet())
		require.NoErrorf(t, err, "failed to create dummy chain")
		if bc == nil {
			t.Skip("skipping test")
		}
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "test1", 50000000000000000, code),
			NewLuaTxDeploy("user1", "test2", 0, code2),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		tx := NewLuaTxCall("user1", "test1", 0, fmt.Sprintf(`{"Name":"send", "Args":["%s"]}`, nameToAddress("test2")))
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")

		r := bc.GetReceipt(tx.Hash())
		expectedGas := map[int32]int64{3: 105087, 4: 105087}[currentVersion]
		assert.Equalf(t, expectedGas, int64(r.GetGasUsed()), "gas used not equal")

		state, err := bc.GetAccountState("test2")
		assert.Equalf(t, int64(2), state.GetBalanceBigInt().Int64(), "balance state not equal")

		tx = NewLuaTxCall("user1", "test1", 0, fmt.Sprintf(`{"Name":"send2", "Args":["%s"]}`, nameToAddress("test2")))
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")

		r = bc.GetReceipt(tx.Hash())
		expectedGas = map[int32]int64{3: 105179, 4: 105755}[currentVersion]
		assert.Equalf(t, expectedGas, int64(r.GetGasUsed()), "gas used not equal")

		state, err = bc.GetAccountState("test2")
		assert.Equalf(t, int64(6), state.GetBalanceBigInt().Int64(), "balance state not equal")

}

func TestGasPerFunction(t *testing.T) {
	skipNotOnAmd64(t)

	var err error
	code := readLuaCode(t, "gas_per_function.lua")

	bc, err := LoadDummyChain(SetHardForkVersion(currentVersion), RunOnPubNet())
	require.NoErrorf(t, err, "failed to create dummy chain")
	if bc == nil {
		t.Skip("skipping test")
	}
	defer bc.Release()

	err = bc.ConnectBlock(
		// add funds to account
		NewLuaTxAccount("user", 100, types.Aergo),
		// deploy 2 copies of the contract
		NewLuaTxDeploy("user", "contract_v2", 0, code),
		NewLuaTxDeploy("user", "contract_v3", 0, code),
		NewLuaTxDeploy("user", "contract_v4", 0, code),
	)
	require.NoError(t, err)

	// transfer funds to the contracts
	err = bc.ConnectBlock(
		NewLuaTxCall("user", "contract_v2", uint64(10e18), `{"Name":"default"}`),
		NewLuaTxCall("user", "contract_v3", uint64(10e18), `{"Name":"default"}`),
		NewLuaTxCall("user", "contract_v4", uint64(10e18), `{"Name":"default"}`),
	)
	require.NoError(t, err, "sending funds to contracts")

	tests_v2 := []struct {
		funcName    string
		funcArgs    string
		amount      int64
		expectedGas int64
	}{
		{"comp_ops", "", 0, 134635},
		{"unarytest_n_copy_ops", "", 0, 134548},
		{"unary_ops", "", 0, 134947},
		{"binary_ops", "", 0, 136470},
		{"constant_ops", "", 0, 134463},
		{"upvalue_n_func_ops", "", 0, 135742},
		{"table_ops", "", 0, 135733},
		{"call_n_vararg_ops", "", 0, 136396},
		{"return_ops", "", 0, 134468},
		{"loop_n_branche_ops", "", 0, 137803},
		{"function_header_ops", "", 0, 134447},

		{"assert", "", 0, 134577},
		{"getfenv", "", 0, 134472},
		{"metatable", "", 0, 135383},
		{"ipairs", "", 0, 134470},
		{"pairs", "", 0, 134470},
		{"next", "", 0, 134518},
		{"rawequal", "", 0, 134647},
		{"rawget", "", 0, 134518},
		{"rawset", "", 0, 135336},
		{"select", "", 0, 134597},
		{"setfenv", "", 0, 134507},
		{"tonumber", "", 0, 134581},
		{"tostring", "", 0, 134852},
		{"type", "", 0, 134680},
		{"unpack", "", 0, 142140},
		{"pcall", "", 0, 138169},
		{"xpcall", "", 0, 138441},

		{"string.byte", "", 0, 148435},
		{"string.char", "", 0, 151792},
		{"string.dump", "", 0, 138300},
		{"string.find", "", 0, 139239},
		{"string.format", "", 0, 135159},
		{"string.gmatch", "", 0, 135194},
		{"string.gsub", "", 0, 136338},
		{"string.len", "", 0, 134528},
		{"string.lower", "", 0, 139746},
		{"string.match", "", 0, 134708},
		{"string.rep", "", 0, 213323},
		{"string.reverse", "", 0, 139746},
		{"string.sub", "", 0, 136600},
		{"string.upper", "", 0, 139746},

		{"table.concat", "", 0, 155263},
		{"table.insert", "", 0, 288649},
		{"table.remove", "", 0, 148059},
		{"table.maxn", "", 0, 139357},
		{"table.sort", "", 0, 151261},

		{"math.abs", "", 0, 134615},
		{"math.ceil", "", 0, 134615},
		{"math.floor", "", 0, 134615},
		{"math.max", "", 0, 134987},
		{"math.min", "", 0, 134987},
		{"math.pow", "", 0, 134975},

		{"bit.tobit", "", 0, 134510},
		{"bit.tohex", "", 0, 134985},
		{"bit.bnot", "", 0, 134487},
		{"bit.bor", "", 0, 134561},
		{"bit.band", "", 0, 134537},
		{"bit.xor", "", 0, 134537},
		{"bit.lshift", "", 0, 134510},
		{"bit.rshift", "", 0, 134510},
		{"bit.ashift", "", 0, 134510},
		{"bit.rol", "", 0, 134510},
		{"bit.ror", "", 0, 134510},
		{"bit.bswap", "", 0, 134467},

		{"bignum.number", "", 0, 136307},
		{"bignum.isneg", "", 0, 136539},
		{"bignum.iszero", "", 0, 136539},
		{"bignum.tonumber", "", 0, 136859},
		{"bignum.tostring", "", 0, 137150},
		{"bignum.neg", "", 0, 138603},
		{"bignum.sqrt", "", 0, 139479},
		{"bignum.compare", "", 0, 136804},
		{"bignum.add", "", 0, 138145},
		{"bignum.sub", "", 0, 138090},
		{"bignum.mul", "", 0, 140468},
		{"bignum.div", "", 0, 139958},
		{"bignum.mod", "", 0, 141893},
		{"bignum.pow", "", 0, 140887},
		{"bignum.divmod", "", 0, 146193},
		{"bignum.powmod", "", 0, 145559},
		{"bignum.operators", "", 0, 138811},

		{"json", "", 0, 142320},

		{"crypto.sha256", "", 0, 137578},
		{"crypto.ecverify", "", 0, 139467},

		{"state.set", "", 0, 137310},
		{"state.get", "", 0, 137115},
		{"state.delete", "", 0, 137122},

		{"system.getSender", "", 0, 135656},
		{"system.getBlockheight", "", 0, 134761},
		{"system.getTxhash", "", 0, 135132},
		{"system.getTimestamp", "", 0, 134761},
		{"system.getContractID", "", 0, 135656},
		{"system.setItem", "", 0, 135589},
		{"system.getItem", "", 0, 135898},
		{"system.getAmount", "", 0, 134803},
		{"system.getCreator", "", 0, 135156},
		{"system.getOrigin", "", 0, 135656},
		// as the returned value differs in length (43 or 44)
		// due to base58, the computed gas is different.
		//{ "system.getPrevBlockHash", "", 0, 135132 },

		{"contract.send", "", 0, 135716},
		{"contract.balance", "", 0, 135605},
		{"contract.deploy", "", 0, 158752},
		{"contract.call", "", 0, 149642},
		{"contract.pcall", "", 0, 150563},
		{"contract.delegatecall", "", 0, 144902},
		{"contract.event", "", 0, 153263},
	}

	tests_v3 := []struct {
		funcName    string
		funcArgs    string
		amount      int64
		expectedGas int64
	}{
		{"comp_ops", "", 0, 134635},
		{"unarytest_n_copy_ops", "", 0, 134548},
		{"unary_ops", "", 0, 134947},
		{"binary_ops", "", 0, 136470},
		{"constant_ops", "", 0, 134463},
		{"upvalue_n_func_ops", "", 0, 135742},
		{"table_ops", "", 0, 135733},
		{"call_n_vararg_ops", "", 0, 136396},
		{"return_ops", "", 0, 134468},
		{"loop_n_branche_ops", "", 0, 137803},
		{"function_header_ops", "", 0, 134447},

		{"assert", "", 0, 134577},
		{"getfenv", "", 0, 134472},
		{"metatable", "", 0, 135383},
		{"ipairs", "", 0, 134470},
		{"pairs", "", 0, 134470},
		{"next", "", 0, 134518},
		{"rawequal", "", 0, 134647},
		{"rawget", "", 0, 134518},
		{"rawset", "", 0, 135336},
		{"select", "", 0, 134597},
		{"setfenv", "", 0, 134507},
		{"tonumber", "", 0, 134581},
		{"tostring", "", 0, 134852},
		{"type", "", 0, 134680},
		{"unpack", "", 0, 142140},
		{"pcall", "", 0, 137560},
		{"xpcall", "", 0, 137832},

		{"string.byte", "", 0, 148435},
		{"string.char", "", 0, 151792},
		{"string.dump", "", 0, 138261},
		{"string.find", "", 0, 139239},
		{"string.format", "", 0, 135159},
		{"string.gmatch", "", 0, 135194},
		{"string.gsub", "", 0, 136338},
		{"string.len", "", 0, 134528},
		{"string.lower", "", 0, 139746},
		{"string.match", "", 0, 134708},
		{"string.rep", "", 0, 213323},
		{"string.reverse", "", 0, 139746},
		{"string.sub", "", 0, 136600},
		{"string.upper", "", 0, 139746},

		{"table.concat", "", 0, 155263},
		{"table.insert", "", 0, 288649},
		{"table.remove", "", 0, 148059},
		{"table.maxn", "", 0, 139357},
		{"table.sort", "", 0, 151261},

		{"math.abs", "", 0, 134615},
		{"math.ceil", "", 0, 134615},
		{"math.floor", "", 0, 134615},
		{"math.max", "", 0, 134987},
		{"math.min", "", 0, 134987},
		{"math.pow", "", 0, 134975},

		{"bit.tobit", "", 0, 134510},
		{"bit.tohex", "", 0, 134985},
		{"bit.bnot", "", 0, 134487},
		{"bit.bor", "", 0, 134561},
		{"bit.band", "", 0, 134537},
		{"bit.xor", "", 0, 134537},
		{"bit.lshift", "", 0, 134510},
		{"bit.rshift", "", 0, 134510},
		{"bit.ashift", "", 0, 134510},
		{"bit.rol", "", 0, 134510},
		{"bit.ror", "", 0, 134510},
		{"bit.bswap", "", 0, 134467},

		{"bignum.number", "", 0, 136307},
		{"bignum.isneg", "", 0, 136539},
		{"bignum.iszero", "", 0, 136539},
		{"bignum.tonumber", "", 0, 136859},
		{"bignum.tostring", "", 0, 137150},
		{"bignum.neg", "", 0, 138603},
		{"bignum.sqrt", "", 0, 139479},
		{"bignum.compare", "", 0, 136804},
		{"bignum.add", "", 0, 138145},
		{"bignum.sub", "", 0, 138090},
		{"bignum.mul", "", 0, 140468},
		{"bignum.div", "", 0, 139958},
		{"bignum.mod", "", 0, 141893},
		{"bignum.pow", "", 0, 140887},
		{"bignum.divmod", "", 0, 146193},
		{"bignum.powmod", "", 0, 145559},
		{"bignum.operators", "", 0, 138811},

		{"json", "", 0, 142320},

		{"crypto.sha256", "", 0, 137578},
		{"crypto.ecverify", "", 0, 139467},

		{"state.set", "", 0, 137310},
		{"state.get", "", 0, 137115},
		{"state.delete", "", 0, 137122},

		{"system.getSender", "", 0, 135656},
		{"system.getBlockheight", "", 0, 134761},
		{"system.getTxhash", "", 0, 135132},
		{"system.getTimestamp", "", 0, 134761},
		{"system.getContractID", "", 0, 135656},
		{"system.setItem", "", 0, 135589},
		{"system.getItem", "", 0, 135898},
		{"system.getAmount", "", 0, 134803},
		{"system.getCreator", "", 0, 135156},
		{"system.getOrigin", "", 0, 135656},
		// as the returned value differs in length (43 or 44)
		// due to base58, the computed gas is different.
		//{ "system.getPrevBlockHash", "", 0, 135132 },

		{"contract.send", "", 0, 135716},
		{"contract.balance", "", 0, 135605},
		{"contract.deploy", "", 0, 158752},
		{"contract.call", "", 0, 149642},
		{"contract.pcall", "", 0, 150563},
		{"contract.delegatecall", "", 0, 144902},
		{"contract.event", "", 0, 153263},
	}

	tests_v4 := []struct {
		funcName   string
		funcArgs   string
		amount     int64
		expectedGas int64
	}{
		{"comp_ops", "", 0, 143204},
		{"unarytest_n_copy_ops", "", 0, 143117},
		{"unary_ops", "", 0, 143552},
		{"binary_ops", "", 0, 145075},
		{"constant_ops", "", 0, 143032},
		{"upvalue_n_func_ops", "", 0, 144347},
		{"table_ops", "", 0, 144482},
		{"call_n_vararg_ops", "", 0, 145001},
		{"return_ops", "", 0, 143037},
		{"loop_n_branche_ops", "", 0, 146372},
		{"function_header_ops", "", 0, 143016},

		{"assert", "", 0, 143146},
		{"ipairs", "", 0, 143039},
		{"pairs", "", 0, 143039},
		{"next", "", 0, 143087},
		{"select", "", 0, 143166},
		{"tonumber", "", 0, 143186},
		{"tostring", "", 0, 143457},
		{"type", "", 0, 143285},
		{"unpack", "", 0, 150745},
		{"pcall", "", 0, 147905},
		{"xpcall", "", 0, 148177},

		{"string.byte", "", 0, 157040},
		{"string.char", "", 0, 160397},
		{"string.find", "", 0, 147808},
		{"string.format", "", 0, 143764},
		{"string.gmatch", "", 0, 143799},
		{"string.gsub", "", 0, 144943},
		{"string.len", "", 0, 143097},
		{"string.lower", "", 0, 148351},
		{"string.match", "", 0, 143313},
		{"string.rep", "", 0, 221928},
		{"string.reverse", "", 0, 148351},
		{"string.sub", "", 0, 145205},
		{"string.upper", "", 0, 148351},

		{"table.concat", "", 0, 163868},
		{"table.insert", "", 0, 297254},
		{"table.remove", "", 0, 156664},
		{"table.maxn", "", 0, 147962},
		{"table.sort", "", 0, 159866},

		{"math.abs", "", 0, 143184},
		{"math.ceil", "", 0, 143184},
		{"math.floor", "", 0, 143184},
		{"math.max", "", 0, 143556},
		{"math.min", "", 0, 143556},
		{"math.pow", "", 0, 143544},

		{"bit.tobit", "", 0, 143079},
		{"bit.tohex", "", 0, 143590},
		{"bit.bnot", "", 0, 143056},
		{"bit.bor", "", 0, 143130},
		{"bit.band", "", 0, 143106},
		{"bit.xor", "", 0, 143106},
		{"bit.lshift", "", 0, 143079},
		{"bit.rshift", "", 0, 143079},
		{"bit.ashift", "", 0, 143079},
		{"bit.rol", "", 0, 143079},
		{"bit.ror", "", 0, 143079},
		{"bit.bswap", "", 0, 143036},

		{"bignum.number", "", 0, 144912},
		{"bignum.isneg", "", 0, 145144},
		{"bignum.iszero", "", 0, 145144},
		{"bignum.tonumber", "", 0, 145464},
		{"bignum.tostring", "", 0, 145755},
		{"bignum.neg", "", 0, 147208},
		{"bignum.sqrt", "", 0, 148084},
		{"bignum.compare", "", 0, 145409},
		{"bignum.add", "", 0, 146750},
		{"bignum.sub", "", 0, 146695},
		{"bignum.mul", "", 0, 149073},
		{"bignum.div", "", 0, 148563},
		{"bignum.mod", "", 0, 150498},
		{"bignum.pow", "", 0, 149492},
		{"bignum.divmod", "", 0, 154798},
		{"bignum.powmod", "", 0, 154164},
		{"bignum.operators", "", 0, 147416},

		{"json", "", 0, 151357},

		{"crypto.sha256", "", 0, 146183},
		{"crypto.ecverify", "", 0, 148036},

		{"state.set", "", 0, 145915},
		{"state.get", "", 0, 145720},
		{"state.delete", "", 0, 145727},

		{"system.getSender", "", 0, 144261},
		{"system.getBlockheight", "", 0, 143330},
		//{"system.getTxhash", "", 0, 143734},
		{"system.getTimestamp", "", 0, 143330},
		{"system.getContractID", "", 0, 144261},
		{"system.setItem", "", 0, 144194},
		{"system.getItem", "", 0, 144503},
		{"system.getAmount", "", 0, 143408},
		{"system.getCreator", "", 0, 143761},
		{"system.getOrigin", "", 0, 144261},

		{"contract.send", "", 0, 144321},
		{"contract.balance", "", 0, 144210},
		{"contract.deploy", "", 0, 168092},
		{"contract.call", "", 0, 159738},
		{"contract.pcall", "", 0, 160659},
		{"contract.delegatecall", "", 0, 153795},
		{"contract.event", "", 0, 163452},
	}

	if currentVersion == 2 {

	// iterate over the tests
	for _, test := range tests_v2 {
		funcName := test.funcName
		funcArgs := test.funcArgs
		amount := test.amount
		expectedGas := test.expectedGas

		var payload string
		if len(funcArgs) == 0 {
			payload = fmt.Sprintf(`{"Name":"run_test", "Args":["%s"]}`, funcName)
		} else {
			payload = fmt.Sprintf(`{"Name":"run_test", "Args":["%s",%s]}`, funcName, funcArgs)
		}
		tx := NewLuaTxCall("user", "contract_v2", uint64(amount), payload)
		err = bc.ConnectBlock(tx)
		require.NoError(t, err, "while executing %s", funcName)

		usedGas := bc.GetReceipt(tx.Hash()).GetGasUsed()
		assert.Equal(t, expectedGas, int64(usedGas), "wrong used gas for %s", funcName)

		// print the function name and the used gas
		// for this test:
		//fmt.Printf("		{ \"%s\", \"\", 0, %d },\n", funcName, usedGas)
		// for integration tests (tests/test-gas-per-function-v2.sh):
		//fmt.Printf("add_test \"%s\" %d\n", funcName, usedGas)
	}

	} else if currentVersion == 3 {

	// iterate over the tests
	for _, test := range tests_v3 {
		funcName := test.funcName
		funcArgs := test.funcArgs
		amount := test.amount
		expectedGas := test.expectedGas

		var payload string
		if len(funcArgs) == 0 {
			payload = fmt.Sprintf(`{"Name":"run_test", "Args":["%s"]}`, funcName)
		} else {
			payload = fmt.Sprintf(`{"Name":"run_test", "Args":["%s",%s]}`, funcName, funcArgs)
		}
		tx := NewLuaTxCall("user", "contract_v3", uint64(amount), payload)
		err = bc.ConnectBlock(tx)
		require.NoError(t, err, "while executing %s", funcName)

		usedGas := bc.GetReceipt(tx.Hash()).GetGasUsed()
		assert.Equal(t, expectedGas, int64(usedGas), "wrong used gas for %s", funcName)

		// print the function name and the used gas
		// for this test:
		//fmt.Printf("		{ \"%s\", \"\", 0, %d },\n", funcName, usedGas)
		// for integration tests (tests/test-gas-per-function-v3.sh):
		//fmt.Printf("add_test \"%s\" %d\n", funcName, usedGas)
	}

	} else if currentVersion == 4 {

	// iterate over the tests
	for _, test := range tests_v4 {
		funcName := test.funcName
		funcArgs := test.funcArgs
		amount := test.amount
		expectedGas := test.expectedGas

		var payload string
		if len(funcArgs) == 0 {
			payload = fmt.Sprintf(`{"Name":"run_test", "Args":["%s"]}`, funcName)
		} else {
			payload = fmt.Sprintf(`{"Name":"run_test", "Args":["%s",%s]}`, funcName, funcArgs)
		}
		tx := NewLuaTxCall("user", "contract_v4", uint64(amount), payload)
		err = bc.ConnectBlock(tx)
		require.NoError(t, err, "while executing %s", funcName)

		usedGas := bc.GetReceipt(tx.Hash()).GetGasUsed()
		assert.Equal(t, expectedGas, int64(usedGas), "wrong used gas for %s", funcName)

		// print the function name and the used gas
		// for this test:
		//fmt.Printf("		{\"%s\", \"\", 0, %d},\n", funcName, usedGas)
		// for integration tests (tests/test-gas-per-function-v4.sh):
		//fmt.Printf("add_test \"%s\" %d\n", funcName, usedGas)
	}

	}

}

func TestGasHello(t *testing.T) {
	skipNotOnAmd64(t)

	var err error
	code := readLuaCode(t, "contract_hello.lua")

	if currentVersion == 1 {

		err = expectGas(code, 0, `"hello"`, `"world"`, 100000)
		assert.NoError(t, err)

	} else if currentVersion == 2 {

		err = expectGas(code, 0, `"hello"`, `"w"`, 101203+3*1)
		assert.NoError(t, err)
		err = expectGas(code, 0, `"hello"`, `"wo"`, 101203+3*2)
		assert.NoError(t, err)
		err = expectGas(code, 0, `"hello"`, `"wor"`, 101203+3*3)
		assert.NoError(t, err)
		err = expectGas(code, 0, `"hello"`, `"worl"`, 101203+3*4)
		assert.NoError(t, err)
		err = expectGas(code, 0, `"hello"`, `"world"`, 101203+3*5)
		assert.NoError(t, err)

	} else if currentVersion >= 3 {

		err = expectGas(code, 0, `"hello"`, `"world"`, 101203+3*5)
		assert.NoError(t, err)

	}
}

func TestGasDeploy(t *testing.T) {
	skipNotOnAmd64(t)

	var err error
	code := readLuaCode(t, "gas_deploy.lua")

	if currentVersion <= 1 {

		err = expectGas(code, 0, `"testPcall"`, ``, 0)
		assert.NoError(t, err)

	} else if currentVersion == 2 {

		err = expectGas(code, 0, `"testPcall"`, ``, 117861)
		assert.NoError(t, err)

	} else if currentVersion == 3 {

		err = expectGas(code, 0, `"testPcall"`, ``, 117861)
		assert.NoError(t, err)

	} else if currentVersion == 4 {

		err = expectGas(code, 0, `"testPcall"`, ``, 118350)
		assert.NoError(t, err)

	}

}

func TestGasOp(t *testing.T) {
	skipNotOnAmd64(t)

	var err error
	code := readLuaCode(t, "gas_op.lua")

	if currentVersion <= 1 {

		err = expectGas(string(code), 0, `"main"`, ``, 100000)
		assert.NoError(t, err)

	} else if currentVersion == 2 {

		err = expectGas(string(code), 0, `"main"`, ``, 117610)
		assert.NoError(t, err)

	} else if currentVersion == 3 {

		err = expectGas(string(code), 0, `"main"`, ``, 117610)
		assert.NoError(t, err)

	} else if currentVersion == 4 {

		err = expectGas(string(code), 0, `"main"`, ``, 120832)
		assert.NoError(t, err)

	}
}

func TestGasBF(t *testing.T) {
	skipNotOnAmd64(t)

	var err error
	code2 := readLuaCode(t, "gas_bf_v2.lua")
	code4 := readLuaCode(t, "gas_bf_v4.lua")

	if currentVersion <= 1 {

		// err = expectGas(t, string(code), 0, `"main"`, ``, 100000, SetTimeout(500))
		// assert.NoError(t, err)

	} else if currentVersion == 2 {

		err = expectGas(string(code2), 0, `"main"`, ``, 47456244, SetTimeout(500))
		assert.NoError(t, err)

	} else if currentVersion == 3 {

		err = expectGas(string(code2), 0, `"main"`, ``, 47456046, SetTimeout(500))
		assert.NoError(t, err)

	} else if currentVersion == 4 {

		err = expectGas(string(code4), 0, `"main"`, ``, 47342481, SetTimeout(500))
		assert.NoError(t, err)

	}
}

func TestGasLuaCryptoVerifyProof(t *testing.T) {
	skipNotOnAmd64(t)

	code := readLuaCode(t, "feature_crypto_verify_proof.lua")

	if currentVersion == 2 {

		// v2 raw
		err := expectGas(string(code), 0, `"verifyProofRaw"`, ``, 154137)
		assert.NoError(t, err)

		// v2 hex
		err = expectGas(string(code), 0, `"verifyProofHex"`, ``, 108404)
		assert.NoError(t, err)

	} else if currentVersion == 3 {

		// v3 raw
		err := expectGas(string(code), 0, `"verifyProofRaw"`, ``, 154137)
		assert.NoError(t, err)

		// v3 hex
		err = expectGas(string(code), 0, `"verifyProofHex"`, ``, 108404)
		assert.NoError(t, err)

	} else if currentVersion == 4 {

		// v4 raw
		err := expectGas(string(code), 0, `"verifyProofRaw"`, ``, 160281)
		assert.NoError(t, err)

		// v4 hex
		err = expectGas(string(code), 0, `"verifyProofHex"`, ``, 108404)
		assert.NoError(t, err)

	}
}

func expectGas(contractCode string, amount int64, funcName, funcArgs string, expectGas int64, opt ...DummyChainOptions) error {
	// append set pubnet
	bc, err := LoadDummyChain(append(opt, RunOnPubNet(), SetHardForkVersion(currentVersion))...)
	if err != nil {
		return err
	}
	if bc == nil {
		return nil
	}
	defer bc.Release()

	if err = bc.ConnectBlock(
		NewLuaTxAccount(DEF_TEST_ACCOUNT, 1, types.Aergo),
		NewLuaTxDeploy(DEF_TEST_ACCOUNT, DEF_TEST_CONTRACT, 0, contractCode),
	); err != nil {
		return err
	}

	var payload string
	if len(funcArgs) == 0 {
		payload = fmt.Sprintf(`{"Name":%s}`, funcName)
	} else {
		payload = fmt.Sprintf(`{"Name":%s, "Args":[%s]}`, funcName, funcArgs)
	}

	var balanceBefore, balanceAfter int64
	// get before balance
	if state, err := bc.GetAccountState(DEF_TEST_ACCOUNT); err != nil {
		return fmt.Errorf("failed to get account state: %s", err)
	} else {
		balanceBefore = state.GetBalanceBigInt().Int64()
	}
	// execute tx in block
	tx := NewLuaTxCall(DEF_TEST_ACCOUNT, DEF_TEST_CONTRACT, uint64(amount), payload)
	if err = bc.ConnectBlock(tx); err != nil {
		return err
	}
	// get after balance
	if state, err := bc.GetAccountState(DEF_TEST_ACCOUNT); err != nil {
		return fmt.Errorf("failed to get account state: %s", err)
	} else {
		balanceAfter = state.GetBalanceBigInt().Int64()
	}

	usedGas := bc.GetReceipt(tx.Hash()).GetGasUsed()
	if expectGas != int64(usedGas) {
		return fmt.Errorf("wrong used gas, expected: %d, but got: %d", expectGas, usedGas)
	}
	if balanceBefore-expectGas != balanceAfter {
		return fmt.Errorf("wrong balance status, expected: %d, but got: %d", expectGas, balanceBefore-balanceAfter)
	}

	return nil
}

func TestTypeInvalidKey(t *testing.T) {
	skipNotOnAmd64(t)

	code := readLuaCode(t, "type_invalidkey.lua")

	if currentVersion < 3 {
		t.Skip()
	}

		bc, err := LoadDummyChain(SetHardForkVersion(currentVersion), RunOnAllNets())
		require.NoErrorf(t, err, "failed to create dummy chain")
		if bc == nil {
			t.Skip("skipping test")
		}
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "invalidkey", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "invalidkey", 0, `{"Name":"key_table"}`).Fail("cannot use 'table' as a key"))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "invalidkey", 0, `{"Name":"key_func"}`).Fail("cannot use 'function' as a key"))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "invalidkey", 0, `{"Name":"key_statemap"}`).Fail("cannot use 'userdata' as a key"))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "invalidkey", 0, `{"Name":"key_statearray"}`).Fail("cannot use 'userdata' as a key"))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "invalidkey", 0, `{"Name":"key_statevalue"}`).Fail("cannot use 'userdata' as a key"))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "invalidkey", 0, `{"Name":"key_upval"}`).Fail("cannot use 'table' as a key"))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "invalidkey", 0, `{"Name":"key_nil"}`).Fail("invalid key type: 'nil', state.map: 'h'"))
		require.NoErrorf(t, err, "failed to call tx")

}

func TestTypeBigTable(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	skipNotOnAmd64(t)

	code := readLuaCode(t, "type_bigtable_1.lua")
	code2 := readLuaCode(t, "type_bigtable_2.lua")

	if currentVersion < 3 {
		t.Skip()
	}

		bc, err := LoadDummyChain(SetHardForkVersion(currentVersion), RunOnPrivNet())
		require.NoErrorf(t, err, "failed to create dummy chain")
		if bc == nil {
			t.Skip("skipping test")
		}
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "big", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		// About 900MB
		err = bc.ConnectBlock(NewLuaTxCall("user1", "big", 0, `{"Name": "inserts", "Args":[25]}`))
		require.NoErrorf(t, err, "failed to call tx")

		contract.SetStateSQLMaxDBSize(20)
		err = bc.ConnectBlock(NewLuaTxAccount("user1", 100, types.Aer), NewLuaTxDeploy("user1", "big20", 0, code2))
		require.NoErrorf(t, err, "failed to deploy")

		for i := 0; i < 17; i++ {
			err = bc.ConnectBlock(NewLuaTxCall("user1", "big20", 0, `{"Name": "inserts"}`))
			require.NoErrorf(t, err, "failed to call tx")
		}
		err = bc.ConnectBlock(NewLuaTxCall("user1", "big20", 0, `{"Name": "inserts"}`).Fail("database or disk is full"))
		require.NoErrorf(t, err, "failed to call tx")

}
