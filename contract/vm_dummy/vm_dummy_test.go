// The tests in vm_dummy_test.go are architecture independent tests.
package vm_dummy

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/aergoio/aergo/v2/contract"
	"github.com/aergoio/aergo/v2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/aergoio/aergo/v2/cmd/aergoluac/util"
)

const min_version int32 = 2
const max_version int32 = 4
const min_version_multicall int32 = 4

func TestDisabledFunctions(t *testing.T) {
	code := readLuaCode(t, "disabled-functions.lua")

	for version := int32(4); version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version), SetPubNet())
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user", 1, types.Aergo),
			NewLuaTxDeploy("user", "test", 0, code),
		)
		assert.NoErrorf(t, err, "failed to deploy contract")

		err = bc.ConnectBlock(
			NewLuaTxCall("user", "test", 0, `{"Name":"check_disabled_functions","Args":[]}`),
		)
		assert.NoErrorf(t, err, "failed execution")
	}
}

func TestMaxCallDepth(t *testing.T) {
	//code := readLuaCode(t, "maxcalldepth_1.lua")
	// this contract receives a list of contract IDs to be called
	code2 := readLuaCode(t, "maxcalldepth_2.lua")
	// this contract stores the address of the next contract to be called
	code3 := readLuaCode(t, "maxcalldepth_3.lua")

	for version := int32(3); version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version), SetPubNet())
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user", 1, types.Aergo),
		)
		if err != nil {
			t.Error(err)
		}

		/*
			// deploy 2 identical contracts
			err = bc.ConnectBlock(
				NewLuaTxDeploy("user", "c1", 0, definition1),
				NewLuaTxDeploy("user", "c2", 0, definition1),
			)
			if err != nil {
				t.Error(err)
			}

			// call first contract - recursion depth 64
			err = bc.ConnectBlock(
				NewLuaTxCall("user", "c1", 0, `{"Name":"call_me", "Args":[1, 64]}`),
			)
			if err != nil {
				t.Error(err)
			}
			// check state
			err = bc.Query("c1", `{"Name":"check_state"}`, "", "true")
			if err != nil {
				t.Error(err)
			}
			// query view
			err = bc.Query("c1", `{"Name":"get_total_calls"}`, "", "[64,64]")
			if err != nil {
				t.Error(err)
			}
			for i := 1; i <= 64; i++ {
				err = bc.Query("c1", fmt.Sprintf(`{"Name":"get_call_info", "Args":["%d"]}`, i), "", fmt.Sprintf("%d", i))
				if err != nil {
					t.Error(err)
				}
			}

			// call second contract - recursion depth 66
			err = bc.ConnectBlock(
				NewLuaTxCall("user", "c2", 0, `{"Name":"call_me", "Args":[1, 66]}`).
					Fail("exceeded the maximum call depth"),
			)
			if err != nil {
				t.Error(err)
			}
			// check state - should fail
			err = bc.Query("c2", `{"Name":"check_state"}`, "", "")
			if err == nil {
				t.Error("should fail")
			}
			// query view - must return nil
			err = bc.Query("c2", `{"Name":"get_total_calls"}`, "", "[null,null]")
			if err != nil {
				t.Error(err)
			}
			for i := 1; i <= 64; i++ {
				err = bc.Query("c2", fmt.Sprintf(`{"Name":"get_call_info", "Args":["%d"]}`, i), "", "null")
				if err != nil {
					t.Error(err)
				}
			}
		*/

		// deploy 66 identical contracts using definition2
		for i := 1; i <= 66; i++ {
			err = bc.ConnectBlock(
				NewLuaTxDeploy("user", fmt.Sprintf("c2%d", i), 0, code2),
			)
			if err != nil {
				t.Error(err)
			}
		}
		// deploy 66 identical contracts using definition3
		for i := 1; i <= 66; i++ {
			err = bc.ConnectBlock(
				NewLuaTxDeploy("user", fmt.Sprintf("c3%d", i), 0, code3),
			)
			if err != nil {
				t.Error(err)
			}
		}

		// build a list of contract IDs, used to call the first contract
		contracts := make([]string, 64)
		contracts_str := []byte("")
		for i := 1; i <= 64; i++ {
			contracts[i-1] = StrToAddress(fmt.Sprintf("c2%d", i))
		}
		contracts_str, err = json.Marshal(contracts)
		if err != nil {
			t.Error(err)
		}
		// call first contract - recursion depth 64
		err = bc.ConnectBlock(
			NewLuaTxCall("user", "c2"+fmt.Sprintf("%d", 1), 0, fmt.Sprintf(`{"Name":"call_me", "Args":[%s, 1, 64]}`, string(contracts_str))),
		)
		if err != nil {
			t.Error(err)
		}
		// check state on all the 64 contracts (query total calls and call info)
		for i := 1; i <= 64; i++ {
			err = bc.Query(fmt.Sprintf("c2%d", i), `{"Name":"get_total_calls"}`, "", "1")
			if err != nil {
				t.Error(err)
			}
			//err = bc.Query(fmt.Sprintf("c2%d", i), fmt.Sprintf(`{"Name":"get_call_info", "Args":["%d"]}`, i), "", fmt.Sprintf("%d", i))
			err = bc.Query(fmt.Sprintf("c2%d", i), `{"Name":"get_call_info", "Args":["1"]}`, "", fmt.Sprintf("%d", i))
			if err != nil {
				t.Error(err)
			}
		}

		// add the 66th contract to the list
		contracts = append(contracts, StrToAddress(fmt.Sprintf("c2%d", 6)))
		contracts_str, err = json.Marshal(contracts)
		if err != nil {
			t.Error(err)
		}
		// call first contract - recursion depth 66
		err = bc.ConnectBlock(
			NewLuaTxCall("user", "c2"+fmt.Sprintf("%d", 1), 0, fmt.Sprintf(`{"Name":"call_me", "Args":[%s, 1, 66]}`, string(contracts_str))).Fail("exceeded the maximum call depth"),
		)
		if err != nil {
			t.Error(err)
		}
		// check state on all the 64 contracts (query total calls and call info)
		for i := 1; i <= 64; i++ {
			err = bc.Query(fmt.Sprintf("c2%d", i), `{"Name":"get_total_calls"}`, "", "1")
			if err != nil {
				t.Error(err)
			}
			err = bc.Query(fmt.Sprintf("c2%d", i), `{"Name":"get_call_info", "Args":["1"]}`, "", fmt.Sprintf("%d", i))
			if err != nil {
				t.Error(err)
			}
		}
		// check state on the 66th contract (query total calls and call info)
		err = bc.Query("c2"+fmt.Sprintf("%d", 66), `{"Name":"get_total_calls"}`, "", "null")
		if err != nil {
			t.Error(err)
		}
		err = bc.Query("c2"+fmt.Sprintf("%d", 66), `{"Name":"get_call_info", "Args":["1"]}`, "", "null")
		if err != nil {
			t.Error(err)
		}

		// set next_contract for each contract
		for i := 1; i <= 66; i++ {
			err = bc.ConnectBlock(
				NewLuaTxCall("user", fmt.Sprintf("c3%d", i), 0, fmt.Sprintf(`{"Name":"set_next_contract", "Args":["%s"]}`, StrToAddress(fmt.Sprintf("c3%d", i+1)))),
			)
			if err != nil {
				t.Error(err)
			}
		}
		// call first contract - recursion depth 64
		err = bc.ConnectBlock(
			NewLuaTxCall("user", "c3"+fmt.Sprintf("%d", 1), 0, `{"Name":"call_me", "Args":[1, 64]}`),
		)
		if err != nil {
			t.Error(err)
		}
		// check state on all the 64 contracts (query total calls and call info)
		for i := 1; i <= 64; i++ {
			err = bc.Query(fmt.Sprintf("c3%d", i), `{"Name":"get_total_calls"}`, "", "1")
			if err != nil {
				t.Error(err)
			}
			err = bc.Query(fmt.Sprintf("c3%d", i), `{"Name":"get_call_info", "Args":["1"]}`, "", fmt.Sprintf("%d", i))
			if err != nil {
				t.Error(err)
			}
		}

		// call first contract - recursion depth 66
		err = bc.ConnectBlock(
			NewLuaTxCall("user", "c3"+fmt.Sprintf("%d", 1), 0, `{"Name":"call_me", "Args":[1, 66]}`).Fail("exceeded the maximum call depth"),
		)
		if err != nil {
			t.Error(err)
		}
		// check state on all the 64 contracts (query total calls and call info)
		for i := 1; i <= 64; i++ {
			err = bc.Query(fmt.Sprintf("c3%d", i), `{"Name":"get_total_calls"}`, "", "1")
			if err != nil {
				t.Error(err)
			}
			err = bc.Query(fmt.Sprintf("c3%d", i), `{"Name":"get_call_info", "Args":["1"]}`, "", fmt.Sprintf("%d", i))
			if err != nil {
				t.Error(err)
			}
		}
		// check state on the 66th contract (query total calls and call info)
		err = bc.Query("c3"+fmt.Sprintf("%d", 66), `{"Name":"get_total_calls"}`, "", "null")
		if err != nil {
			t.Error(err)
		}
		err = bc.Query("c3"+fmt.Sprintf("%d", 66), `{"Name":"get_call_info", "Args":["1"]}`, "", "null")
		if err != nil {
			t.Error(err)
		}

		// Circle: contract 1 calls contract 2, contract 2 calls contract 3, contract 3 calls contract 1...

		// deploy 4 identical contracts using definition2
		for i := 1; i <= 4; i++ {
			err = bc.ConnectBlock(
				NewLuaTxDeploy("user", fmt.Sprintf("c4%d", i), 0, code2),
			)
			if err != nil {
				t.Error(err)
			}
		}
		// build a list of contract IDs, used to call the first contract
		contracts = make([]string, 4)
		for i := 1; i <= 4; i++ {
			contracts[i-1] = StrToAddress(fmt.Sprintf("c4%d", i))
		}
		contracts_str, err = json.Marshal(contracts)
		if err != nil {
			t.Error(err)
		}
		// call first contract - recursion depth 64
		err = bc.ConnectBlock(
			NewLuaTxCall("user", "c4"+fmt.Sprintf("%d", 1), 0, fmt.Sprintf(`{"Name":"call_me", "Args":[%s, 1, 64]}`, string(contracts_str))),
		)
		if err != nil {
			t.Error(err)
		}
		// check state on all the 4 contracts
		// each contract should have (64 / 4) = 16 calls
		for i := 1; i <= 4; i++ {
			err = bc.Query(fmt.Sprintf("c4%d", i), `{"Name":"get_total_calls"}`, "", "16")
			if err != nil {
				t.Error(err)
			}
			for j := 1; j <= 16; j++ {
				err = bc.Query(fmt.Sprintf("c4%d", i), fmt.Sprintf(`{"Name":"get_call_info", "Args":["%d"]}`, j), "", fmt.Sprintf("%d", i+4*(j-1)))
				if err != nil {
					t.Error(err)
				}
			}
		}

		// call first contract - recursion depth 66
		err = bc.ConnectBlock(
			NewLuaTxCall("user", "c4"+fmt.Sprintf("%d", 1), 0, fmt.Sprintf(`{"Name":"call_me", "Args":[%s, 1, 66]}`, string(contracts_str))).Fail("exceeded the maximum call depth"),
		)
		if err != nil {
			t.Error(err)
		}
		// check state on all the 4 contracts
		// each contract should have (64 / 4) = 16 calls
		for i := 1; i <= 4; i++ {
			err = bc.Query(fmt.Sprintf("c4%d", i), `{"Name":"get_total_calls"}`, "", "16")
			if err != nil {
				t.Error(err)
			}
			for j := 1; j <= 16; j++ {
				err = bc.Query(fmt.Sprintf("c4%d", i), fmt.Sprintf(`{"Name":"get_call_info", "Args":["%d"]}`, j), "", fmt.Sprintf("%d", i+4*(j-1)))
				if err != nil {
					t.Error(err)
				}
			}
		}

		// ZigZag: contract 1 calls contract 2, contract 2 calls contract 1...

		// deploy 2 identical contracts using definition2
		for i := 1; i <= 2; i++ {
			err = bc.ConnectBlock(
				NewLuaTxDeploy("user", fmt.Sprintf("c5%d", i), 0, code2),
			)
			if err != nil {
				t.Error(err)
			}
		}
		// build a list of contract IDs, used to call the first contract
		contracts = make([]string, 2)
		for i := 1; i <= 2; i++ {
			contracts[i-1] = StrToAddress(fmt.Sprintf("c5%d", i))
		}
		contracts_str, err = json.Marshal(contracts)
		if err != nil {
			t.Error(err)
		}
		// call first contract - recursion depth 64
		err = bc.ConnectBlock(
			NewLuaTxCall("user", "c5"+fmt.Sprintf("%d", 1), 0, fmt.Sprintf(`{"Name":"call_me", "Args":[%s, 1, 64]}`, string(contracts_str))),
		)
		if err != nil {
			t.Error(err)
		}
		// check state on all the 2 contracts
		// each contract should have (64 / 2) = 32 calls
		for i := 1; i <= 2; i++ {
			err = bc.Query(fmt.Sprintf("c5%d", i), `{"Name":"get_total_calls"}`, "", "32")
			if err != nil {
				t.Error(err)
			}
			for j := 1; j <= 32; j++ {
				err = bc.Query(fmt.Sprintf("c5%d", i), fmt.Sprintf(`{"Name":"get_call_info", "Args":["%d"]}`, j), "", fmt.Sprintf("%d", i+2*(j-1)))
				if err != nil {
					t.Error(err)
				}
			}
		}

		// call first contract - recursion depth 66
		err = bc.ConnectBlock(
			NewLuaTxCall("user", "c5"+fmt.Sprintf("%d", 1), 0, fmt.Sprintf(`{"Name":"call_me", "Args":[%s, 1, 66]}`, string(contracts_str))).Fail("exceeded the maximum call depth"),
		)
		if err != nil {
			t.Error(err)
		}
		// check state on all the 2 contracts
		// each contract should have (64 / 2) = 32 calls
		for i := 1; i <= 2; i++ {
			err = bc.Query(fmt.Sprintf("c5%d", i), `{"Name":"get_total_calls"}`, "", "32")
			if err != nil {
				t.Error(err)
			}
			for j := 1; j <= 32; j++ {
				err = bc.Query(fmt.Sprintf("c5%d", i), fmt.Sprintf(`{"Name":"get_call_info", "Args":["%d"]}`, j), "", fmt.Sprintf("%d", i+2*(j-1)))
				if err != nil {
					t.Error(err)
				}
			}
		}

	}
}

func TestContractSystem(t *testing.T) {
	code := readLuaCode(t, "contract_system.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo))
		require.NoErrorf(t, err, "failed to new account")

		err = bc.ConnectBlock(NewLuaTxDeploy("user1", "system", 0, code))
		require.NoErrorf(t, err, "failed to deploy contract")

		tx := NewLuaTxCall("user1", "system", 0, `{"Name":"testState", "Args":[]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")

		receipt := bc.GetReceipt(tx.Hash())
		exRv := fmt.Sprintf(`["%s","6FbDRScGruVdATaNWzD51xJkTfYCVwxSZDb7gzqCLzwf","AmhNNBNY7XFk4p5ym4CJf8nTcRTEHjWzAeXJfhP71244CjBCAQU3",%d,3,999]`, StrToAddress("user1"), bc.cBlock.Header.Timestamp/1e9)
		assert.Equal(t, exRv, receipt.GetRet(), "receipt ret error")

		if version >= 4 {

      // system.version()

			tx = NewLuaTxCall("user1", "system", 0, `{"Name":"get_version", "Args":[]}`)
			err = bc.ConnectBlock(tx)
			require.NoErrorf(t, err, "failed to call tx")

			receipt = bc.GetReceipt(tx.Hash())
			expected := fmt.Sprintf(`%d`, version)
			assert.Equal(t, expected, receipt.GetRet(), "receipt ret error")

			err = bc.Query("system", `{"Name":"get_version", "Args":[]}`, "", expected)
			require.NoErrorf(t, err, "failed to query")

			// system.toPubKey()

			err = bc.Query("system", `{"Name":"to_pubkey", "Args":["AmgKtCaGjH4XkXwny2Jb1YH5gdsJGJh78ibWEgLmRWBS5LMfQuTf"]}`, "", `"0x0c3270bb25fea5bf0029b57e78581647a143265810b84940dd24e543ddc618ab91"`)
			require.NoErrorf(t, err, "failed to query")

			err = bc.Query("system", `{"Name":"to_pubkey", "Args":["Amhmj6kKZz7mPstBAPJWRe1e8RHP7bZ5pV35XatqTHMWeAVSyMkc"]}`, "", `"0x0cf0d0fd04f44db75d66409346102167d67c40a5d76d46748fc4533f0265d0f83f"`)
			require.NoErrorf(t, err, "failed to query")

			err = bc.Query("system", `{"Name":"to_pubkey", "Args":["6FbDRScGruVdATaNWzD51xJkTfYCVwxSZDb7gzqCLzwf"]}`, "invalid address length", "")
			require.NoErrorf(t, err, "failed to query")

			err = bc.Query("system", `{"Name":"to_pubkey", "Args":["0x0c3270bb25fea5bf0029b57e78581647a143265810b84940dd24e543ddc618ab91"]}`, "invalid address length", "")
			require.NoErrorf(t, err, "failed to query")

			err = bc.Query("system", `{"Name":"to_pubkey", "Args":[""]}`, "invalid address length", "")
			require.NoErrorf(t, err, "failed to query")

			err = bc.Query("system", `{"Name":"to_pubkey", "Args":[]}`, "string expected, got nil", "")
			require.NoErrorf(t, err, "failed to query")

			// system.toAddress()

			err = bc.Query("system", `{"Name":"to_address", "Args":["0x0c3270bb25fea5bf0029b57e78581647a143265810b84940dd24e543ddc618ab91"]}`, "", `"AmgKtCaGjH4XkXwny2Jb1YH5gdsJGJh78ibWEgLmRWBS5LMfQuTf"`)
			require.NoErrorf(t, err, "failed to query")

			err = bc.Query("system", `{"Name":"to_address", "Args":["0x0cf0d0fd04f44db75d66409346102167d67c40a5d76d46748fc4533f0265d0f83f"]}`, "", `"Amhmj6kKZz7mPstBAPJWRe1e8RHP7bZ5pV35XatqTHMWeAVSyMkc"`)
			require.NoErrorf(t, err, "failed to query")

			err = bc.Query("system", `{"Name":"to_address", "Args":["0cf0d0fd04f44db75d66409346102167d67c40a5d76d46748fc4533f0265d0f83f"]}`, "", `"Amhmj6kKZz7mPstBAPJWRe1e8RHP7bZ5pV35XatqTHMWeAVSyMkc"`)
			require.NoErrorf(t, err, "failed to query")

			err = bc.Query("system", `{"Name":"to_address", "Args":["AmhNNBNY7XFk4p5ym4CJf8nTcRTEHjWzAeXJfhP71244CjBCAQU3"]}`, "invalid public key", "")
			require.NoErrorf(t, err, "failed to query")

			err = bc.Query("system", `{"Name":"to_address", "Args":["6FbDRScGruVdATaNWzD51xJkTfYCVwxSZDb7gzqCLzwf"]}`, "invalid public key", "")
			require.NoErrorf(t, err, "failed to query")

			err = bc.Query("system", `{"Name":"to_address", "Args":[""]}`, "invalid public key", "")
			require.NoErrorf(t, err, "failed to query")

			err = bc.Query("system", `{"Name":"to_address", "Args":[]}`, "string expected, got nil", "")
			require.NoErrorf(t, err, "failed to query")

		}

	}
}

func TestContractHello(t *testing.T) {
	code := readLuaCode(t, "contract_hello.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create test database")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo))
		require.NoErrorf(t, err, "failed to new account")

		err = bc.ConnectBlock(NewLuaTxDeploy("user1", "hello", 0, code))
		require.NoErrorf(t, err, "failed to deploy contract")

		tx := NewLuaTxCall("user1", "hello", 0, `{"Name":"hello", "Args":["World"]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")

		receipt := bc.GetReceipt(tx.Hash())
		assert.Equal(t, `"Hello World"`, receipt.GetRet(), "receipt ret error")

	}
}

func TestContractSend(t *testing.T) {
	code1 := readLuaCode(t, "contract_send_1.lua")
	code2 := readLuaCode(t, "contract_send_2.lua")
	code3 := readLuaCode(t, "contract_send_3.lua")
	code4 := readLuaCode(t, "contract_send_4.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "test1", 50, code1),
			NewLuaTxDeploy("user1", "test2", 0, code2),
			NewLuaTxDeploy("user1", "test3", 0, code3),
			NewLuaTxDeploy("user1", "test4", 0, code4),
		)
		assert.NoErrorf(t, err, "failed to deploy contract")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "test1", 0, fmt.Sprintf(`{"Name":"send", "Args":["%s"]}`, nameToAddress("test2"))),
		)
		assert.NoErrorf(t, err, "failed to call tx")

		state, err := bc.GetAccountState("test2")
		assert.Equalf(t, int64(2), state.GetBalanceBigInt().Int64(), "balance error")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "test1", 0, fmt.Sprintf(`{"Name":"send", "Args":["%s"]}`, nameToAddress("test3"))).Fail(`[Contract.LuaSendAmount] call err: not found function: default`),
		)
		assert.NoErrorf(t, err, "failed to connect new block")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "test1", 0, fmt.Sprintf(`{"Name":"send", "Args":["%s"]}`, nameToAddress("test4"))).Fail(`[Contract.LuaSendAmount] call err: 'default' is not payable`),
		)
		assert.NoErrorf(t, err, "failed to connect new block")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "test1", 0, fmt.Sprintf(`{"Name":"send", "Args":["%s"]}`, nameToAddress("user1"))),
		)
		assert.NoErrorf(t, err, "failed to connect new block")

	}
}

func TestContractQuery(t *testing.T) {
	code := readLuaCode(t, "contract_query.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo))
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.ConnectBlock(
			NewLuaTxDeploy("user1", "query", 0, code),
			NewLuaTxCall("user1", "query", 2, `{"Name":"inc", "Args":[]}`),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		query, err := bc.GetAccountState("query")
		require.NoErrorf(t, err, "failed to get account state")
		assert.Equalf(t, int64(2), query.GetBalanceBigInt().Int64(), "not equal balance")

		err = bc.Query("query", `{"Name":"inc", "Args":[]}`, "set not permitted in query", "")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "1")
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestContractCall(t *testing.T) {
	code1 := readLuaCode(t, "contract_call_1.lua")
	code2 := readLuaCode(t, "contract_call_2.lua")
	code3 := readLuaCode(t, "contract_call_3.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			// deploy the counter contract
			NewLuaTxDeploy("user1", "counter", 0, code1).Constructor("[1]"),
			// increment the value
			NewLuaTxCall("user1", "counter", 0, `{"Name":"inc", "Args":[]}`),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		// check the value

		err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "2")
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(
			// deploy the caller contract
			NewLuaTxDeploy("user1", "caller", 0, code2).Constructor(fmt.Sprintf(`["%s"]`, nameToAddress("counter"))),
			// indirectly increment the value on the counter contract
			NewLuaTxCall("user1", "caller", 0, `{"Name":"cinc", "Args":[]}`),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		// check the value on both contracts

		err = bc.Query("caller", `{"Name":"cget", "Args":[]}`, "", "3")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "3")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("caller", `{"Name":"dget", "Args":[]}`, "", "99")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("caller", `{"Name":"get", "Args":[]}`, "", "99")
		require.NoErrorf(t, err, "failed to query")

		// use delegate call to increment the value on the same contract

		tx := NewLuaTxCall("user1", "caller", 0, `{"Name":"dinc", "Args":[]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")
		receipt := bc.GetReceipt(tx.Hash())
		assert.Equalf(t, `99`, receipt.GetRet(), "contract Call ret error")

		// do it again

		tx = NewLuaTxCall("user1", "caller", 0, `{"Name":"dinc", "Args":[]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")
		receipt = bc.GetReceipt(tx.Hash())
		assert.Equalf(t, `100`, receipt.GetRet(), "contract Call ret error")

		// check the value on both contracts

		err = bc.Query("caller", `{"Name":"cget", "Args":[]}`, "", "3")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "3")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("caller", `{"Name":"dget", "Args":[]}`, "", "101")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("caller", `{"Name":"get", "Args":[]}`, "", "101")
		require.NoErrorf(t, err, "failed to query")

		// use delegate call to set the value on the same contract

		tx = NewLuaTxCall("user1", "caller", 0, `{"Name":"dset", "Args":[500]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")
		receipt = bc.GetReceipt(tx.Hash())
		assert.Equalf(t, ``, receipt.GetRet(), "contract Call ret error")

		// check the value on both contracts

		err = bc.Query("caller", `{"Name":"cget", "Args":[]}`, "", "3")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "3")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("caller", `{"Name":"dget", "Args":[]}`, "", "500")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("caller", `{"Name":"get", "Args":[]}`, "", "500")
		require.NoErrorf(t, err, "failed to query")

		// indirectly set the value on the counter contract

		tx = NewLuaTxCall("user1", "caller", 0, `{"Name":"cset", "Args":[750]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")
		receipt = bc.GetReceipt(tx.Hash())
		assert.Equalf(t, ``, receipt.GetRet(), "contract Call ret error")

		// check the value on both contracts

		err = bc.Query("caller", `{"Name":"cget", "Args":[]}`, "", "750")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "750")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("caller", `{"Name":"dget", "Args":[]}`, "", "500")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("caller", `{"Name":"get", "Args":[]}`, "", "500")
		require.NoErrorf(t, err, "failed to query")

		// collect call info using delegate call: A -> delegate_call(B) -> A

		tx = NewLuaTxCall("user1", "caller", 0, `{"Name":"get_call_info", "Args":["AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn","get_call_info2"]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")
		receipt = bc.GetReceipt(tx.Hash())
		expected := `[{"ctr_id":"AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn","origin":"Amg25cfD4ibjmjPYbtWnMKocrF147gJJxKy5uuFymEBNF2YiPwzr","sender":"Amg25cfD4ibjmjPYbtWnMKocrF147gJJxKy5uuFymEBNF2YiPwzr"},{"ctr_id":"AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn","origin":"Amg25cfD4ibjmjPYbtWnMKocrF147gJJxKy5uuFymEBNF2YiPwzr","sender":"Amg25cfD4ibjmjPYbtWnMKocrF147gJJxKy5uuFymEBNF2YiPwzr"},{"ctr_id":"AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn","origin":"Amg25cfD4ibjmjPYbtWnMKocrF147gJJxKy5uuFymEBNF2YiPwzr","sender":"AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn"}]`
		assert.Equalf(t, expected, receipt.GetRet(), "contract Call ret error")

		// collect call info via delegate call using query

		expected = `[{"ctr_id":"AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn","origin":"","sender":""},{"ctr_id":"AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn","origin":"","sender":""},{"ctr_id":"AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn","origin":"","sender":"AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn"}]`
		err = bc.Query("caller", `{"Name":"get_call_info", "Args":["AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn","get_call_info2"]}`, "", expected)
		require.NoErrorf(t, err, "failed to query")

		// deploy the third contract

		err = bc.ConnectBlock(
			NewLuaTxDeploy("user1", "third", 0, code3),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		// collect call info using delegate call: A -> delegate_call(B) -> C

		tx = NewLuaTxCall("user1", "caller", 0, `{"Name":"get_call_info", "Args":["AmhJ2JWVSDeXxYrMRtH38hjnGDLVkLJCLD1XCTGZSjoQV2xCQUEg","get_call_info"]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")
		receipt = bc.GetReceipt(tx.Hash())
		expected = `[{"ctr_id":"AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn","origin":"Amg25cfD4ibjmjPYbtWnMKocrF147gJJxKy5uuFymEBNF2YiPwzr","sender":"Amg25cfD4ibjmjPYbtWnMKocrF147gJJxKy5uuFymEBNF2YiPwzr"},{"ctr_id":"AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn","origin":"Amg25cfD4ibjmjPYbtWnMKocrF147gJJxKy5uuFymEBNF2YiPwzr","sender":"Amg25cfD4ibjmjPYbtWnMKocrF147gJJxKy5uuFymEBNF2YiPwzr"},{"ctr_id":"AmhJ2JWVSDeXxYrMRtH38hjnGDLVkLJCLD1XCTGZSjoQV2xCQUEg","origin":"Amg25cfD4ibjmjPYbtWnMKocrF147gJJxKy5uuFymEBNF2YiPwzr","sender":"AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn"}]`
		assert.Equalf(t, expected, receipt.GetRet(), "contract Call ret error")

		// collect call info via delegate call using query

		expected = `[{"ctr_id":"AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn","origin":"","sender":""},{"ctr_id":"AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn","origin":"","sender":""},{"ctr_id":"AmhJ2JWVSDeXxYrMRtH38hjnGDLVkLJCLD1XCTGZSjoQV2xCQUEg","origin":"","sender":"AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn"}]`
		err = bc.Query("caller", `{"Name":"get_call_info", "Args":["AmhJ2JWVSDeXxYrMRtH38hjnGDLVkLJCLD1XCTGZSjoQV2xCQUEg","get_call_info"]}`, "", expected)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestContractCallSelf(t *testing.T) {
	code := readLuaCode(t, "contract_call_self.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "A", 0, code),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		tx := NewLuaTxCall("user1", "A", 0, `{"Name":"call_myself", "Args":[]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")
		receipt := bc.GetReceipt(tx.Hash())
		require.Equalf(t, `123`, receipt.GetRet(), "contract call ret error")

		// make a recursive call like this: A -> A -> A -> A -> A
		tx = NewLuaTxCall("user1", "A", 0, `{"Name":"call_me_again", "Args":[0,5]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")
		// make sure the first instance can read the updated state variable
		receipt = bc.GetReceipt(tx.Hash())
		require.Equalf(t, `5`, receipt.GetRet(), "contract call ret error")

	}
}

func TestContractPingPongCall(t *testing.T) {
	code1 := readLuaCode(t, "contract_pingpongcall_1.lua")
	code2 := readLuaCode(t, "contract_pingpongcall_2.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "A", 0, code1),
			NewLuaTxDeploy("user1", "B", 0, code2).Constructor(fmt.Sprintf(`["%s"]`, nameToAddress("A"))),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		// make a ping pong call like this: A -> B -> A
		tx := NewLuaTxCall("user1", "A", 0, fmt.Sprintf(`{"Name":"start", "Args":["%s"]}`, nameToAddress("B")))
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")

		// make sure the first instance can read the updated state variable
		receipt := bc.GetReceipt(tx.Hash())
		require.Equalf(t, `"callback"`, receipt.GetRet(), "contract call ret error")

		err = bc.Query("A", `{"Name":"get"}`, "", `"callback"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("B", `{"Name":"get"}`, "", `"called"`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestRollback(t *testing.T) {
	code := readLuaCode(t, "rollback.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo))
		require.NoErrorf(t, err, "failed to connect new block")
		err = bc.ConnectBlock(NewLuaTxDeploy("user1", "query", 0, code), NewLuaTxCall("user1", "query", 0, `{"Name":"inc", "Args":[]}`))
		require.NoErrorf(t, err, "failed to connect new block")
		err = bc.ConnectBlock(NewLuaTxCall("user1", "query", 0, `{"Name":"inc", "Args":[]}`), NewLuaTxCall("user1", "query", 0, `{"Name":"inc", "Args":[]}`))
		require.NoErrorf(t, err, "failed to connect new block")
		err = bc.ConnectBlock(NewLuaTxCall("user1", "query", 0, `{"Name":"inc", "Args":[]}`), NewLuaTxCall("user1", "query", 0, `{"Name":"inc", "Args":[]}`))
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "5")
		require.NoErrorf(t, err, "failed to query")

		err = bc.DisConnectBlock()
		require.NoErrorf(t, err, "failed to disconnect block")

		err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "3")
		require.NoErrorf(t, err, "failed to query")

		err = bc.DisConnectBlock()
		require.NoErrorf(t, err, "failed to disconnect block")

		err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "1")
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "query", 0, `{"Name":"inc", "Args":[]}`))
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.Query("query", `{"Name":"query", "Args":["key1"]}`, "", "2")
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestAbi(t *testing.T) {
	codeNoAbi := readLuaCode(t, "abi_no.lua")
	codeEmpty := readLuaCode(t, "abi_empty.lua")
	codeLocalFunc := readLuaCode(t, "abi_localfunc.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "a", 0, codeNoAbi))
		require.Errorf(t, err, fmt.Sprintf("expected err : %s, buf got nil", "no exported functions"))
		require.Containsf(t, err.Error(), "no exported functions", "not contains error message")

		err = bc.ConnectBlock(NewLuaTxDeploy("user1", "a", 0, codeEmpty))
		require.Errorf(t, err, fmt.Sprintf("expected err : %s, buf got nil", "no exported functions."))
		require.Containsf(t, err.Error(), "no exported functions.", "not contains error message")

		err = bc.ConnectBlock(NewLuaTxDeploy("user1", "a", 0, codeLocalFunc))
		require.Errorf(t, err, fmt.Sprintf("expected err : %s, buf got nil", "global function expected"))
		require.Containsf(t, err.Error(), "global function expected", "not contains error message")

	}
}

func TestGetABI(t *testing.T) {
	code := readLuaCode(t, "getabi.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "hello", 0, code))
		require.NoErrorf(t, err, "failed to connect new block")

		abi, err := bc.GetABI("hello")
		require.NoErrorf(t, err, "failed to get abi")

		jsonAbi, err := json.Marshal(abi)
		require.NoErrorf(t, err, "failed to marshal abi")
		require.Equalf(t, `{"version":"0.2","language":"lua","functions":[{"name":"hello","arguments":[{"name":"say"}]}],"state_variables":[{"name":"Say","type":"value"}]}`, string(jsonAbi), "not equal abi")

	}
}

func TestPayable(t *testing.T) {
	code := readLuaCode(t, "payable.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo))
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.ConnectBlock(NewLuaTxDeploy("user1", "payable", 1, code))
		require.Errorf(t, err, "expected: 'constructor' is not payable")
		require.Containsf(t, err.Error(), "'constructor' is not payable", "not contains error message")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "payable", 0, `{"Name":"save", "Args": ["blahblah"]}`).Fail("not found contract"))
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.ConnectBlock(NewLuaTxDeploy("user1", "payable", 0, code), NewLuaTxCall("user1", "payable", 0, `{"Name":"save", "Args": ["blahblah"]}`))
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.Query("payable", `{"Name":"load"}`, "", `"blahblah"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "payable", 1, `{"Name":"save", "Args": ["payed"]}`))
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.Query("payable", `{"Name":"load"}`, "", `"payed"`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestDefault(t *testing.T) {
	code := readLuaCode(t, "default.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "default", 0, code),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		tx := NewLuaTxCall("user1", "default", 0, "")
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")

		receipt := bc.GetReceipt(tx.Hash())
		require.Equalf(t, `"default"`, receipt.GetRet(), "contract Call ret error")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "default", 1, "").Fail(`'default' is not payable`))
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.Query("default", `{"Name":"a"}`, "not found function: a", "")
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestReturn(t *testing.T) {
	code := readLuaCode(t, "return_1.lua")
	code2 := readLuaCode(t, "return_2.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "return_num", 0, code),
			NewLuaTxCall("user1", "return_num", 0, `{"Name":"return_num", "Args":[]}`),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.Query("return_num", `{"Name":"return_num", "Args":[]}`, "", "10")
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxDeploy("user1", "foo", 0, code2))
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.Query("foo", `{"Name":"foo", "Args":[]}`, "", "[1,2,3]")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("foo", `{"Name":"foo2", "Args":["foo314"]}`, "", `"foo314"`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestReturnUData(t *testing.T) {
	code := readLuaCode(t, "return_udata.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "rs-return", 0, code),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "rs-return", 0, `{"Name": "test_die", "Args":[]}`).Fail(`unsupport type: userdata`))
		require.NoErrorf(t, err, "failed to connect new block")

	}
}

func TestEvent(t *testing.T) {
	code := readLuaCode(t, "event.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "event", 0, code),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "event", 0, `{"Name": "test_ev", "Args":[]}`))
		require.NoErrorf(t, err, "failed to connect new block")

	}

}

func TestView(t *testing.T) {
	code := readLuaCode(t, "view.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "view", 0, code),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "view", 0, `{"Name": "test_view", "Args":[]}`).Fail("[Contract.Event] event not permitted in query"))
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.Query("view", `{"Name":"k", "Args":[10]}`, "", "10")
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "view", 0, `{"Name": "tx_in_view_function", "Args":[]}`).Fail("[Contract.Event] event not permitted in query"))
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "view", 0, `{"Name": "tx_after_view_function", "Args":[]}`))
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "view", 0, `{"Name": "k2", "Args":[]}`).Fail("[Contract.Event] event not permitted in query"))
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "view", 0, `{"Name": "k3", "Args":[]}`))
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "view", 0, `{"Name": "sqltest", "Args":[]}`).Fail("not permitted in view function"))
		require.NoErrorf(t, err, "failed to connect new block")

	}
}

func TestDeploy(t *testing.T) {
	code := readLuaCode(t, "deploy.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "deploy", uint64(types.Aergo/2), code),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		tx := NewLuaTxCall("user1", "deploy", 0, `{"Name":"hello"}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")

		receipt := bc.GetReceipt(tx.Hash())
		assert.Equalf(t, `["AmgKtCaGjH4XkXwny2Jb1YH5gdsJGJh78ibWEgLmRWBS5LMfQuTf","Hello world"]`, receipt.GetRet(), "contract Call ret error")

		err = bc.Query("deploy", `{"Name":"helloQuery", "Args":["AmgKtCaGjH4XkXwny2Jb1YH5gdsJGJh78ibWEgLmRWBS5LMfQuTf"]}`, "", `"Hello world"`)
		require.NoErrorf(t, err, "failed to query")

		tx = NewLuaTxCall("user1", "deploy", 0, `{"Name":"testConst"}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")

		receipt = bc.GetReceipt(tx.Hash())
		assert.Equalf(t, `["Amhmj6kKZz7mPstBAPJWRe1e8RHP7bZ5pV35XatqTHMWeAVSyMkc","Hello world2"]`, receipt.GetRet(), "contract Call ret error")

		deployAcc, err := bc.GetAccountState("deploy")
		require.NoErrorf(t, err, "failed to get account state")
		assert.Equalf(t, int64(types.Aergo/2-100), deployAcc.GetBalanceBigInt().Int64(), "not same balance")

		deployAcc, err = bc.GetAccountState("deploy")
		require.NoErrorf(t, err, "failed to get account state")

		tx = NewLuaTxCall("user1", "deploy", 0, `{"Name":"testFail"}`)
		err = bc.ConnectBlock(tx)
		require.Errorf(t, err, "expect err : `constructor` is not payable")

		deployAcc, err = bc.GetAccountState("deploy")
		require.NoErrorf(t, err, "failed to get account state")
		assert.Equalf(t, int64(2), int64(deployAcc.Nonce), "not same nonce")

		tx = NewLuaTxCall("user1", "deploy", 0, `{"Name":"testPcall"}`)
		err = bc.ConnectBlock(tx)
		require.Errorf(t, err, "expect err : cannot find contract Amhs9v8EeAAWrrvEFrvMng4UksHRsR7wN1iLqKkXw5bqMV18JP3h")

		deployAcc, err = bc.GetAccountState("deploy")
		require.NoErrorf(t, err, "failed to get account state")
		assert.Equalf(t, int64(2), int64(deployAcc.Nonce), "nonce rollback failed")

		receipt = bc.GetReceipt(tx.Hash())
		assert.Containsf(t, receipt.GetRet(), "cannot find contract", "contract Call ret error")

	}
}

func TestDeploy2(t *testing.T) {
	code := readLuaCode(t, "deploy2.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		oneAergo := types.NewAmount(1, types.Aergo)
		halfAergo := new(big.Int).Div(oneAergo, big.NewInt(2))

		err = bc.ConnectBlock(
			NewLuaTxAccountBig("user1", oneAergo),
			NewLuaTxDeployBig("user1", "deploy", halfAergo, code),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		tx := NewLuaTxCall("user1", "deploy", 0, `{"Name":"hello"}`).Fail(`not permitted state referencing at global scope`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")

	}

}

func TestNDeploy(t *testing.T) {
	code := readLuaCode(t, "deployn.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "n-deploy", 100000, code),
			NewLuaTxCall("user1", "n-deploy", 200000, `{"Name":"testall"}`),
		)
		require.NoErrorf(t, err, "failed to connect new block")

	}
}

func xestInfiniteLoop(t *testing.T) {
	code := readLuaCode(t, "infiniteloop.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetTimeout(50), SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "loop", 0, code),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		errTimeout := "exceeded the maximum instruction count"

		err = bc.ConnectBlock(NewLuaTxCall("user1", "loop", 0, `{"Name":"infiniteLoop"}`))
		require.Errorf(t, err, "expected: %v", errTimeout)
		require.Containsf(t, err.Error(), errTimeout, "not contain timeout error")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "loop", 0, `{"Name":"catch"}`))
		require.Errorf(t, err, "expected: %v", errTimeout)
		require.Containsf(t, err.Error(), errTimeout, "not contain timeout error")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "loop", 0, `{"Name":"contract_catch"}`))
		require.Errorf(t, err, "expected: %v", errTimeout)
		require.Containsf(t, err.Error(), errTimeout, "not contain timeout error")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "loop", 0, `{"Name":"infiniteCall"}`).Fail("stack overflow"))
		require.NoErrorf(t, err, "failed to connect new block")

	}
}

func TestInfiniteLoopOnPubNet(t *testing.T) {
	code := readLuaCode(t, "infiniteloop.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetTimeout(50), SetPubNet(), SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "loop", 0, code),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		errTimeout := contract.VmTimeoutError{}

		err = bc.ConnectBlock(NewLuaTxCall("user1", "loop", 0, `{"Name":"infiniteLoop"}`))
		require.Errorf(t, err, "expected: %v", errTimeout)
		require.Containsf(t, err.Error(), errTimeout.Error(), "not contain timeout error")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "loop", 0, `{"Name":"catch"}`))
		require.Errorf(t, err, "expected: %v", errTimeout)
		require.Containsf(t, err.Error(), errTimeout.Error(), "not contain timeout error")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "loop", 0, `{"Name":"contract_catch"}`))
		require.Errorf(t, err, "expected: %v", errTimeout)
		require.Containsf(t, err.Error(), errTimeout.Error(), "not contain timeout error")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "loop", 0, `{"Name":"infiniteCall"}`).Fail("stack overflow"))
		require.NoErrorf(t, err, "failed to connect new block")

	}
}

func TestUpdateSize(t *testing.T) {
	code := readLuaCode(t, "updatesize.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "loop", 0, code),
			NewLuaTxCall("user1", "loop", 0, `{"Name":"infiniteLoop"}`),
		)
		errMsg := "exceeded size of updates in the state database"
		require.Errorf(t, err, "expected: %s", errMsg)
		require.Containsf(t, err.Error(), errMsg, "error message not same as expected")

	}
}

func TestTimeoutCnt(t *testing.T) {
	// FIXME delete skip after gas limit patch
	t.Skip("disabled until gas limit check is added")
	code := readLuaCode(t, "timeout_1.lua")
	code2 := readLuaCode(t, "timeout_2.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetTimeout(500), SetPubNet(), SetHardForkVersion(version)) // timeout 500 milliseconds
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "timeout-cnt", 0, code),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "timeout-cnt", 0, `{"Name": "infinite_loop"}`).Fail("contract timeout"))
		require.NoErrorf(t, err, "failed to connect new block")

		err = bc.Query("timeout-cnt", `{"Name": "infinite_loop"}`, "exceeded the maximum instruction count")
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxDeploy("user1", "timeout-cnt2", 0, code2))
		require.NoErrorf(t, err, "failed to deploy new tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "timeout-cnt2", 0, `{"Name": "a"}`).Fail("contract timeout"))
		require.NoErrorf(t, err, "failed to call tx")

	}
}

func TestSnapshot(t *testing.T) {
	code := readLuaCode(t, "snapshot.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "snap", 0, code),
		)
		require.NoErrorf(t, err, "failed to deploy contract")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "snap", 0, `{"Name": "inc", "Args":[]}`))
		assert.NoErrorf(t, err, "failed to call contract")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "snap", 0, `{"Name": "inc", "Args":[]}`))
		assert.NoErrorf(t, err, "failed to call contract")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "snap", 0, `{"Name": "inc", "Args":[]}`))
		assert.NoErrorf(t, err, "failed to call contract")

		err = bc.Query("snap", `{"Name":"query"}`, "", "[3,3,3,3]")
		assert.NoErrorf(t, err, "failed to query")

		err = bc.Query("snap", `{"Name":"query", "Args":[2]}`, "", "[1,null,null,null]")
		assert.NoErrorf(t, err, "failed to query")

		err = bc.Query("snap", `{"Name":"query", "Args":[3]}`, "", "[2,2,2,2]")
		assert.NoErrorf(t, err, "failed to query")

		err = bc.Query("snap", `{"Name":"query2", "Args":[]}`, "invalid argument at getsnap, need (state.array, index, blockheight)", "")
		assert.NoErrorf(t, err, "failed to query")

	}
}

func TestKvstore(t *testing.T) {
	code := readLuaCode(t, "kvstore.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "map", 0, code),
		)
		require.NoErrorf(t, err, "failed to deploy contract")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "map", 0, `{"Name":"inc", "Args":["user1"]}`),
			NewLuaTxCall("user1", "map", 0, `{"Name":"setname", "Args":["eve2adam"]}`),
		)
		require.NoErrorf(t, err, "failed to call contract")

		err = bc.ConnectBlock()
		require.NoErrorf(t, err, "failed to new block")

		err = bc.Query("map", `{"Name":"get", "Args":["user1"]}`, "", "1")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("map", `{"Name":"get", "Args":["htwo"]}`, "", "null")
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "map", 0, `{"Name":"inc", "Args":["user1"]}`),
			NewLuaTxCall("user1", "map", 0, `{"Name":"inc", "Args":["htwo"]}`),
			NewLuaTxCall("user1", "map", 0, `{"Name":"set", "Args":["wook", 100]}`),
		)
		require.NoErrorf(t, err, "failed to call contract")

		err = bc.Query("map", `{"Name":"get", "Args":["user1"]}`, "", "2")
		assert.NoErrorf(t, err, "failed to query")

		err = bc.Query("map", `{"Name":"get", "Args":["htwo"]}`, "", "1")
		assert.NoErrorf(t, err, "failed to query")

		err = bc.Query("map", `{"Name":"get", "Args":["wook"]}`, "", "100")
		assert.NoErrorf(t, err, "failed to query")

		err = bc.Query("map", `{"Name":"getname"}`, "", `"eve2adam"`)
		assert.NoErrorf(t, err, "failed to query")

	}
}

// sql tests
func TestSqlConstrains(t *testing.T) {
	code := readLuaCode(t, "sql_constrains.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "constraint", 0, code),
			NewLuaTxCall("user1", "constraint", 0, `{"Name":"init"}`),
			NewLuaTxCall("user1", "constraint", 0, `{"Name":"pkFail"}`).Fail("UNIQUE constraint failed: r.id"),
			NewLuaTxCall("user1", "constraint", 0, `{"Name":"checkFail"}`).Fail("CHECK constraint failed: r"),
			NewLuaTxCall("user1", "constraint", 0, `{"Name":"fkFail"}`).Fail("FOREIGN KEY constraint failed"),
			NewLuaTxCall("user1", "constraint", 0, `{"Name":"notNullFail"}`).Fail("NOT NULL constraint failed: r.nonull"),
			NewLuaTxCall("user1", "constraint", 0, `{"Name":"uniqueFail"}`).Fail("UNIQUE constraint failed: r.only"),
		)
		require.NoErrorf(t, err, "failed to call contract")

	}
}

func TestSqlAutoincrement(t *testing.T) {
	code := readLuaCode(t, "sql_autoincrement.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "auto", 0, code),
			NewLuaTxCall("user1", "auto", 0, `{"Name":"init"}`),
		)
		require.NoErrorf(t, err, "failed to deploy")

		tx := NewLuaTxCall("user1", "auto", 0, `{"Name":"get"}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")

	}
}

func TestSqlOnConflict(t *testing.T) {
	code := readLuaCode(t, "sql_onconflict.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "on_conflict", 0, code),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "on_conflict", 0, `{"name":"stmt_exec", "args": ["insert into t values (2)"]}`),
			NewLuaTxCall("user1", "on_conflict", 0, `{"name":"stmt_exec", "args": ["insert into t values (3),(2),(4)"]}`).Fail(`UNIQUE constraint failed: t.col`),
		)
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("on_conflict", `{"name":"get"}`, "", `[1,2]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "on_conflict", 0, `{"name":"stmt_exec", "args": ["replace into t values (2)"]}`),
			NewLuaTxCall("user1", "on_conflict", 0, `{"name":"stmt_exec", "args": ["insert or ignore into t values (3),(2),(4)"]}`),
		)
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("on_conflict", `{"name":"get"}`, "", `[1,2,3,4]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "on_conflict", 0, `{"name":"stmt_exec", "args": ["insert into t values (5)"]}`),
			NewLuaTxCall("user1", "on_conflict", 0, `{"name":"stmt_exec", "args": ["insert or rollback into t values (6),(5),(7)"]}`).Fail("syntax error"),
		)
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("on_conflict", `{"name":"get"}`, "", `[1,2,3,4,5]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "on_conflict", 0, `{"name":"stmt_exec", "args": ["insert or abort into t values (6),(7),(5),(8),(9)"]}`).Fail("UNIQUE constraint failed"))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("on_conflict", `{"name":"get"}`, "", `[1,2,3,4,5]`)
		require.NoErrorf(t, err, "failed to query")

		// successful pcall
		err = bc.ConnectBlock(NewLuaTxCall("user1", "on_conflict", 0, `{"name":"stmt_exec_pcall", "args": ["insert into t values (6)"]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("on_conflict", `{"name":"get"}`, "", `[1,2,3,4,5,6]`)
		require.NoErrorf(t, err, "failed to query")

		// pcall fails but the tx succeeds
		err = bc.ConnectBlock(NewLuaTxCall("user1", "on_conflict", 0, `{"name":"stmt_exec_pcall", "args": ["insert or fail into t values (7),(5),(8)"]}`))
		require.NoErrorf(t, err, "failed to call tx")

		var expected string
		if version >= 4 {
			// pcall reverts the changes
			expected = `[1,2,3,4,5,6]`
		} else {
			// pcall does not revert the changes
			expected = `[1,2,3,4,5,6,7]`
		}

		err = bc.Query("on_conflict", `{"name":"get"}`, "", expected)
		require.NoErrorf(t, err, "failed to query")

		// here the tx is reverted
		err = bc.ConnectBlock(NewLuaTxCall("user1", "on_conflict", 0, `{"name":"stmt_exec", "args": ["insert or fail into t values (7),(5),(8)"]}`).Fail("UNIQUE constraint failed"))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("on_conflict", `{"name":"get"}`, "", expected)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestSqlDupCol(t *testing.T) {
	code := readLuaCode(t, "sql_dupcol.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "dup_col", 0, code),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.Query("dup_col", `{"name":"get"}`, `too many duplicate column name "1+1", max: 5`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestSqlVmSimple(t *testing.T) {
	code := readLuaCode(t, "sql_vm_simple.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "simple-query", 0, code),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "simple-query", 0, `{"Name": "createAndInsert", "Args":[]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("simple-query", `{"Name": "query", "Args":[]}`, "", `[2,3.1,"X Hello Blockchain",2,3.1,"Y Hello Blockchain",2,3.1,"Z Hello Blockchain"]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("simple-query", `{"Name": "count", "Args":[]}`, "", `3`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "simple-query", 0, `{"Name": "createAndInsert", "Args":[]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("simple-query", `{"Name": "count", "Args":[]}`, "", `6`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.DisConnectBlock()
		require.NoErrorf(t, err, "failed to disconnect block")

		err = bc.Query("simple-query", `{"Name": "count", "Args":[]}`, "", `3`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.DisConnectBlock()
		require.NoErrorf(t, err, "failed to disconnect block")

		err = bc.DisConnectBlock()
		require.NoErrorf(t, err, "failed to disconnect block")

		// there is only a genesis block
		err = bc.Query("simple-query", `{"Name": "count", "Args":[]}`, "not found contract", "")
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestSqlVmFail(t *testing.T) {
	code := readLuaCode(t, "sql_vm_fail.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "fail", 0, code),
			NewLuaTxCall("user1", "fail", 0, `{"Name":"init"}`),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "fail", 0, `{"Name":"add", "Args":[1]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "fail", 0, `{"Name":"add", "Args":[2]}`),
			NewLuaTxCall("user1", "fail", 0, `{"Name":"addFail", "Args":[3]}`).Fail(`near "set": syntax error`),
			NewLuaTxCall("user1", "fail", 0, `{"Name":"add", "Args":[4]}`),
		)
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "fail", 0, `{"Name":"add", "Args":[5]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("fail", `{"Name":"get"}`, "", "12")
		require.NoErrorf(t, err, "failed to query")

		err = bc.DisConnectBlock()
		require.NoErrorf(t, err, "failed to disconnect block")

		err = bc.Query("fail", `{"Name":"get"}`, "", "7")
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestSqlVmPubNet(t *testing.T) {
	code := readLuaCode(t, "sql_vm_pubnet.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetPubNet(), SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "simple-query", 0, code),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "simple-query", 0, `{"Name": "createAndInsert", "Args":[]}`).Fail(`attempt to index global 'db'`))
		require.NoErrorf(t, err, "failed to call tx")

	}
}

func TestSqlVmDateTime(t *testing.T) {
	code := readLuaCode(t, "sql_vm_datetime.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "datetime", 0, code),
			NewLuaTxCall("user1", "datetime", 0, `{"Name":"init"}`),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "datetime", 0, `{"Name":"nowNull"}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "datetime", 0, `{"Name":"localtimeNull"}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("datetime", `{"Name":"get"}`, "", `[{"bool":0},{"bool":1},{"bool":1,"date":"1970-01-01 02:46:40"},{"bool":0,"date":"2004-11-23"}]`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestSqlVmCustomer(t *testing.T) {
	code := readLuaCode(t, "sql_vm_customer.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "customer", 0, code),
			NewLuaTxCall("user1", "customer", 0, `{"Name":"createTable"}`),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "customer", 0, `{"Name":"insert", "Args":["id1","passwd1","name1","20180524","010-1234-5678"]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "customer", 0, `{"Name":"insert", "Args":["id2","passwd2","name2","20180524","010-1234-5678"]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "customer", 0, `{"Name":"update", "Args":["id2","passwd3"]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("customer", `{"Name":"count"}`, "", "2")
		require.NoErrorf(t, err, "failed to query")

		err = bc.DisConnectBlock()
		require.NoErrorf(t, err, "failed to disconnect block")

		err = bc.Query("customer", `{"Name":"query", "Args":["id2"]}`, "", `[{"birth":"20180524","id":"id2","mobile":"010-1234-5678","name":"name2","passwd":"passwd2"}]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "customer", 0, `{"Name":"delete", "Args":["id2"]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("customer", `{"Name":"query", "Args":["id2"]}`, "", `{}`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestSqlVmDataType(t *testing.T) {
	code := readLuaCode(t, "sql_vm_datatype.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "datatype", 0, code),
			NewLuaTxCall("user1", "datatype", 0, `{"Name":"createDataTypeTable"}`),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "datatype", 0, `{"Name":"insertDataTypeTable"}`),
			NewLuaTxCall("user1", "datatype", 0, `{"Name":"insertDataTypeTable"}`),
			NewLuaTxCall("user1", "datatype", 0, `{"Name":"insertDataTypeTable"}`),
		)
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "datatype", 0, `{"Name":"insertDataTypeTable"}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("datatype", `{"Name":"queryOrderByDesc"}`, "", `[{"blockheight1":3,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"},{"blockheight1":2,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"},{"blockheight1":2,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"},{"blockheight1":2,"char1":"fgh","float1":3.14,"int1":1,"var1":"ABCD"}]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("datatype", `{"Name":"queryGroupByBlockheight1"}`, "", `[{"avg_float1":3.14,"blockheight1":2,"count1":3,"sum_int1":3},{"avg_float1":3.14,"blockheight1":3,"count1":1,"sum_int1":1}]`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestSqlVmFunction(t *testing.T) {
	code := readLuaCode(t, "sql_vm_function.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "fns", 0, code),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.Query("fns", `{"Name":"sql_func"}`, "", `[3,1,6]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("fns", `{"Name":"abs_func"}`, "", `[1,0,1]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("fns", `{"Name":"typeof_func"}`, "", `["integer","text","real","null"]`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestSqlVmBook(t *testing.T) {
	code := readLuaCode(t, "sql_vm_book.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "book", 0, code),
			NewLuaTxCall("user1", "book", 0, `{"Name":"createTable"}`),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "book", 0, `{"Name":"makeBook"}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "book", 0, `{"Name":"copyBook"}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("book", `{"Name":"viewCopyBook"}`, "", `[100,"value=1"]`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestSqlVmDateformat(t *testing.T) {
	code := readLuaCode(t, "sql_vm_dateformat.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "data_format", 0, code),
			NewLuaTxCall("user1", "data_format", 0, `{"Name":"init"}`),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.Query("data_format", `{"Name":"get"}`, "", `[["2004-10-24","2004-10-24 11:11:11","20041024111111"],["2018-05-28","2018-05-28 10:45:38","20180528104538"]]`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestSqlVmRecursiveData(t *testing.T) {
	code := readLuaCode(t, "sql_vm_recursivedata.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		tx := NewLuaTxCall("user1", "r", 0, `{"Name":"r"}`)
		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "r", 0, code),
			tx,
		)
		require.Errorf(t, err, "expect err")
		require.Equalf(t, "nested table error", err.Error(), "expect err")

	}
}

func TestSqlJdbc(t *testing.T) {
	code := readLuaCode(t, "sql_jdbc.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "jdbc", 0, code),
			NewLuaTxCall("user1", "jdbc", 0, `{"Name":"init"}`),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.Query("jdbc", `{"Name":"query", "Args":["select a,b,c from total"]}`, "",
			`{"colcnt":3,"colmetas":{"colcnt":3,"decltypes":["int","int","text"],"names":["a","b","c"]},"data":[[1,{},"2"],[2,2,"3"],[3,2,"3"],[4,2,"3"],[5,2,"3"],[6,2,"3"],[7,2,"3"]],"rowcnt":7,"snap":"2"}`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("jdbc", `{"Name":"getmeta", "Args":["select a,b,?+1 from total"]}`, "",
			`[{"colcnt":3,"decltypes":["int","int",""],"names":["a","b","?+1"]},1]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "jdbc", 0, `{"Name": "exec", "Args":["insert into total values (3,4,5)"]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("jdbc", `{"Name":"query", "Args":["select a,b,c from total"]}`, "",
			`{"colcnt":3,"colmetas":{"colcnt":3,"decltypes":["int","int","text"],"names":["a","b","c"]},"data":[[1,{},"2"],[2,2,"3"],[3,2,"3"],[4,2,"3"],[5,2,"3"],[6,2,"3"],[7,2,"3"],[3,4,"5"]],"rowcnt":8,"snap":"3"}`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("jdbc", `{"Name":"queryS", "Args":["2", "select a,b,c from total"]}`, "",
			`{"colcnt":3,"colmetas":{"colcnt":3,"decltypes":["int","int","text"],"names":["a","b","c"]},"data":[[1,{},"2"],[2,2,"3"],[3,2,"3"],[4,2,"3"],[5,2,"3"],[6,2,"3"],[7,2,"3"]],"rowcnt":7,"snap":"3"}`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestTypeMaxString(t *testing.T) {
	code := readLuaCode(t, "type_maxstring.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "oom", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		errMsg := "not enough memory"
		err = bc.ConnectBlock(NewLuaTxCall("user1", "oom", 0, `{"Name":"oom"}`).Fail(errMsg))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "oom", 0, `{"Name":"p"}`).Fail(errMsg))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "oom", 0, `{"Name":"cp"}`).Fail(errMsg))
		require.NoErrorf(t, err, "failed to call tx")

	}
}

func TestTypeMaxStringOnPubNet(t *testing.T) {
	code := readLuaCode(t, "type_maxstring.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version), SetPubNet())
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "oom", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		errMsg := "string length overflow"
		errMsg1 := "not enough memory"
		var travis bool
		if os.Getenv("TRAVIS") == "true" {
			travis = true
		}
		err = bc.ConnectBlock(NewLuaTxCall("user1", "oom", 0, `{"Name":"oom"}`))
		require.Errorf(t, err, "expected: %s", errMsg)
		if !strings.Contains(err.Error(), errMsg) && !strings.Contains(err.Error(), errMsg1) {
			t.Error(err)
		}
		err = bc.ConnectBlock(NewLuaTxCall("user1", "oom", 0, `{"Name":"p"}`))
		if err != nil && (!travis || !strings.Contains(err.Error(), errMsg1)) {
			t.Error(err)
		}
		err = bc.ConnectBlock(NewLuaTxCall("user1", "oom", 0, `{"Name":"cp"}`))
		if err != nil && (!travis || !strings.Contains(err.Error(), errMsg1)) {
			t.Error(err)
		}

	}
}

func TestTypeNsec(t *testing.T) {
	code := readLuaCode(t, "type_nsec.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "nsec", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "nsec", 0, `{"Name": "test_nsec"}`).Fail(`attempt to call global 'nsec' (a nil value)`))
		require.NoErrorf(t, err, "failed to call tx")

	}
}

func TestTypeUtf(t *testing.T) {
	code := readLuaCode(t, "type_utf.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "utf", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.Query("utf", `{"Name":"query"}`, "", "")
		assert.NoErrorf(t, err, "failed to query")

		err = bc.Query("utf", `{"Name":"query2"}`, "", `["E8D4A51000","00"]`)
		assert.NoErrorf(t, err, "failed to query")

		err = bc.Query("utf", `{"Name":"query3"}`, "bignum not allowed negative value", "")
		assert.NoErrorf(t, err, "failed to query")

	}
}

func TestTypeDupVar(t *testing.T) {
	code := readLuaCode(t, "type_dupvar_1.lua")
	code2 := readLuaCode(t, "type_dupvar_2.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo))
		require.NoErrorf(t, err, "failed to new tx")

		err = bc.ConnectBlock(NewLuaTxDeploy("user1", "dupVar", 0, code))
		require.Errorf(t, err, "error expect | duplicated variable: 'Var1'")
		if !strings.Contains(err.Error(), "duplicated variable: 'Var1'") {
			t.Error(err)
		}

		err = bc.ConnectBlock(NewLuaTxDeploy("user1", "dupVar1", 0, code2))
		require.NoErrorf(t, err, "failed to deploy")
		err = bc.ConnectBlock(NewLuaTxCall("user1", "dupVar1", 0, `{"Name": "Work"}`).Fail("duplicated variable: 'Var1'"))
		require.NoErrorf(t, err, "failed to call tx")

	}
}

func TestTypeByteKey(t *testing.T) {
	code := readLuaCode(t, "type_bytekey.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "bk", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.Query("bk", `{"Name":"get"}`, "", `["kk","kk"]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("bk", `{"Name":"getcre"}`, "", fmt.Sprintf(`"%s"`, nameToAddress("user1")))
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestTypeArray(t *testing.T) {
	code := readLuaCode(t, "type_array.lua")

	code2 := readLuaCode(t, "type_array_overflow.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "array", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "array", 0, `{"Name":"inc", "Args":[1]}`),
			NewLuaTxCall("user1", "array", 0, `{"Name":"inc", "Args":[0]}`).Fail("index out of range"),
			NewLuaTxCall("user1", "array", 0, `{"Name":"inc", "Args":[1]}`),
			NewLuaTxCall("user1", "array", 0, `{"Name":"inc", "Args":[1.00000001]}`).Fail("integer expected, got number"),
			NewLuaTxCall("user1", "array", 0, `{"Name":"inc", "Args":["1"]}`).Fail("integer expected, got string)"),
			NewLuaTxCall("user1", "array", 0, `{"Name":"inc", "Args":[true]}`).Fail("integer expected, got boolean"),
			NewLuaTxCall("user1", "array", 0, `{"Name":"inc", "Args":[[1, 2]]}`).Fail("integer expected, got table"),
			NewLuaTxCall("user1", "array", 0, `{"Name":"inc", "Args":[null]}`).Fail("integer expected, got nil)"),
			NewLuaTxCall("user1", "array", 0, `{"Name":"inc", "Args":[{}]}`).Fail("integer expected, got table)"),
			NewLuaTxCall("user1", "array", 0, `{"Name":"inc", "Args":[""]}`).Fail("integer expected, got string)"),
			NewLuaTxCall("user1", "array", 0, `{"Name":"set", "Args":[2,"user1"]}`),
		)
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("array", `{"Name":"get", "Args":[11]}`, "index out of range", "")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("array", `{"Name":"get", "Args":[1]}`, "", "2")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("array", `{"Name":"get", "Args":[2]}`, "", `"user1"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("array", `{"Name":"len"}`, "", `10`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("array", `{"Name":"iter"}`, "", `[2,"user1","nil","nil","nil","nil","nil","nil","nil","nil"]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxDeploy("user1", "overflow", 0, code2))
		errMsg := "integer expected, got number"
		require.Errorf(t, err, "expect no error")
		require.Containsf(t, err.Error(), errMsg, "err not match")

	}
}

func TestTypeMultiArray(t *testing.T) {
	code := readLuaCode(t, "type_multiarray_1.lua")
	code2 := readLuaCode(t, "type_multiarray_2.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "ma", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "ma", 0, `{"Name": "inc", "Args":[]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "ma", 0, `{"Name": "inc", "Args":[]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("ma", fmt.Sprintf(`{"Name":"query", "Args":["%s"]}`, nameToAddress("user1")), "", "[2,2,2,null,10,11]")
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "ma", 0, `{"Name": "del", "Args":[]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("ma", fmt.Sprintf(`{"Name":"query", "Args":["%s"]}`, nameToAddress("user1")), "", "[2,2,null,null,10,11]")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("ma", `{"Name":"iter"}`, "", `{"1,10":"k","10,5":"l"}`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("ma", `{"Name":"seterror"}`, "", ``)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "ma", 0, code2))
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.Query("ma", `{"Name":"query", "Args":[]}`, "", `["A","B",null,null,"A","B","v1"]`)
		require.NoErrorf(t, err, "failed to query")

		tx := NewLuaTxCall("user1", "ma", 0, `{"Name": "abc", "Args":[]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")

		receipt := bc.GetReceipt(tx.Hash())
		require.Equalf(t, `["C","D","A","B","v3"]`, receipt.GetRet(), "contract Call ret error")

		err = bc.Query("ma", `{"Name":"query", "Args":[]}`, "", `["A","B","C","D","A","B","v3"]`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestTypeArrayArg(t *testing.T) {
	code := readLuaCode(t, "type_arrayarg.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "a", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.Query("a", `{"Name": "copy", "Args":[1, 2, 3]}`, "table expected", "")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("a", `{"Name": "copy", "Args":[[1, 2, 3]]}`, "", "[1,2,3]")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("a", `{"Name": "two_arr", "Args":[[1, 2, 3],[4, 5]]}`, "", "[3,2]")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("a", `{"Name": "mixed_args", "Args":[[1, 2, 3], {"name": "user2", "age": 39}, 7]}`, "", `[[1,2,3],{"age":39,"name":"user2"},7]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("a", `{"Name": "mixed_args", "Args":[
[[1, 2, 3],["first", "second"]],
{"name": "user2", "age": 39, "address": {"state": "XXX-do", "city": "YYY-si"}},
"end"
]}`, "", `[[[1,2,3],["first","second"]],{"address":{"city":"YYY-si","state":"XXX-do"},"age":39,"name":"user2"},"end"]`,
		)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("a", `{"Name": "mixed_args", "Args":[
[{"name": "wook", "age": 50}, {"name": "hook", "age": 42}],
{"name": "user2", "age": 39, "scores": [10, 20, 30, 40, 50]},
"hmm..."
]}`, "", `[[{"age":50,"name":"wook"},{"age":42,"name":"hook"}],{"age":39,"name":"user2","scores":[10,20,30,40,50]},"hmm..."]`,
		)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestTypeMapKey(t *testing.T) {
	code := readLuaCode(t, "type_mapkey.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "a", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.Query("a", `{"Name":"getCount", "Args":[1]}`, "", "null")
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "a", 0, `{"Name":"setCount", "Args":[1, 10]}`),
			NewLuaTxCall("user1", "a", 0, `{"Name":"setCount", "Args":["1", 20]}`).Fail("(number expected, got string)"),
			NewLuaTxCall("user1", "a", 0, `{"Name":"setCount", "Args":[1.1, 30]}`),
		)
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("a", `{"Name":"getCount", "Args":["1"]}`, "(number expected, got string)", "")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("a", `{"Name":"getCount", "Args":[1]}`, "", "10")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("a", `{"Name":"getCount", "Args":[1.1]}`, "", "30")
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "a", 0, `{"Name":"setCount", "Args":[true, 40]}`).Fail(`invalid key type: 'boolean', state.map: 'counts'`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "a", 0, `{"Name":"delCount", "Args":[1.1]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("a", `{"Name":"getCount", "Args":[1.1]}`, "", "null")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("a", `{"Name":"getCount", "Args":[2]}`, "", "null")
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxDeploy("user1", "x", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "x", 0, `{"Name":"setCount", "Args":["1", 10]}`),
			NewLuaTxCall("user1", "x", 0, `{"Name":"setCount", "Args":[1, 20]}`).Fail("string expected, got number)"),
			NewLuaTxCall("user1", "x", 0, `{"Name":"setCount", "Args":["third", 30]}`),
		)
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("x", `{"Name":"getCount", "Args":["1"]}`, "", "10")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("x", `{"Name":"getCount", "Args":["third"]}`, "", "30")
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestTypeStateVarFieldUpdate(t *testing.T) {
	code := readLuaCode(t, "type_statevarfieldupdate.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "c", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "c", 0, `{"Name":"InvalidUpdateAge", "Args":[10]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("c", `{"Name":"GetPerson"}`, "", `{"address":"blahblah...","age":38,"name":"user2"}`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "c", 0, `{"Name":"ValidUpdateAge", "Args":[10]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("c", `{"Name":"GetPerson"}`, "", `{"address":"blahblah...","age":10,"name":"user2"}`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestTypeDatetime(t *testing.T) {
	code := readLuaCode(t, "type_datetime.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "datetime", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		// not allowed specifiers

		err = bc.Query("datetime", `{"Name": "Extract", "Args":["%a%A%b%B%h%I%n%p%r%t%x%X%z%Z"]}`, "", `"%a%A%b%B%h%I%n%p%r%t%x%X%z%Z"`)
		require.NoErrorf(t, err, "failed to query")

		// allowed specifiers

		specifiers := map[string]string{
			"%c": "1998-09-10 22:05:18",
			"%C": "19",
			"%d": "10",
			"%D": "09/10/98",
			"%e": "10",
			"%F": "1998-09-10",
			"%g": "98",
			"%G": "1998",
			"%H": "22",
			"%j": "253", // Day of the year [001,366]
			"%m": "09",
			"%M": "05",
			"%R": "22:05",
			"%S": "18",
			"%T": "22:05:18",
			"%u": "4",  // Monday as 1 through Sunday as 7
			"%U": "36", // Week number of the year (Sunday as the first day of the week)
			"%V": "37", // ISO 8601 week number
			"%w": "4",  // Sunday as 0, Saturday as 6
			"%W": "36", // Week number of the year (Monday as the first day of the week)
			"%y": "98",
			"%Y": "1998",
			"%%": "%",
		}

		for specifier, expected := range specifiers {
			err := bc.Query("datetime", `{"Name": "Extract", "Args":["`+specifier+`"]}`, "", `"`+expected+`"`)
			require.NoErrorf(t, err, "failed to query with specifier %s", specifier)
		}

		err = bc.Query("datetime", `{"Name": "Extract", "Args":["%FT%T"]}`, "", `"1998-09-10T22:05:18"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("datetime", `{"Name": "Extract", "Args":["%Y-%m-%d %H:%M:%S"]}`, "", `"1998-09-10 22:05:18"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("datetime", `{"Name": "Difftime"}`, "", `[72318,"20:05:18"]`)
		require.NoErrorf(t, err, "failed to query")

		// set a fixed timestamp for the next block
		bc.SetTimestamp(false, 1696286666)
		// need to create the block for the next queries to use the value
		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "datetime", 0, `{"Name": "SetTimestamp", "Args": [2527491900]}`),
		)
		require.NoErrorf(t, err, "failed to call tx")

		// use the block timestamp

		err = bc.Query("datetime", `{"Name": "CreateDate", "Args":["%Y-%m-%d %H:%M:%S"]}`, "", `"2023-10-02 22:44:26"`)
		require.NoErrorf(t, err, "failed to query")

		// used the new stored timestamp

		specifiers = map[string]string{
			"%c": "2050-02-03 09:05:00",
			"%C": "20",
			"%d": "03",
			"%D": "02/03/50",
			"%e": " 3", // Space-padded day of the month
			"%F": "2050-02-03",
			"%g": "50",
			"%G": "2050",
			"%H": "09",
			"%j": "034", // Day of the year [001,366]
			"%m": "02",
			"%M": "05",
			"%R": "09:05",
			"%S": "00",
			"%T": "09:05:00",
			"%u": "4",  // Thursday (Monday as 1, Sunday as 7)
			"%U": "05", // Week number of the year (Sunday as the first day of the week)
			"%V": "05", // ISO 8601 week number
			"%w": "4",  // Sunday as 0, Saturday as 6
			"%W": "05", // Week number of the year (Monday as the first day of the week)
			"%y": "50",
			"%Y": "2050",
			"%%": "%",
		}

		for specifier, expected := range specifiers {
			err := bc.Query("datetime", `{"Name": "Extract", "Args":["`+specifier+`"]}`, "", `"`+expected+`"`)
			require.NoErrorf(t, err, "failed to query with specifier %s", specifier)
		}

		err = bc.Query("datetime", `{"Name": "Extract", "Args":["%FT%T"]}`, "", `"2050-02-03T09:05:00"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("datetime", `{"Name": "Extract", "Args":["%Y-%m-%d %H:%M:%S"]}`, "", `"2050-02-03 09:05:00"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("datetime", `{"Name": "Difftime"}`, "", `[25500,"07:05:00"]`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestTypeDynamicArray(t *testing.T) {
	code := readLuaCode(t, "type_dynamicarray_zerolen.lua")
	code2 := readLuaCode(t, "type_dynamicarray.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo))
		require.NoErrorf(t, err, "failed to new account")
		err = bc.ConnectBlock(NewLuaTxDeploy("user1", "zeroLen", 0, code))
		require.Errorf(t, err, "no error | expected: the array length must be greater than zero")
		require.Containsf(t, err.Error(), "the array length must be greater than zero", "wrong error message")

		tx := NewLuaTxDeploy("user1", "dArr", 0, code2)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.Query("dArr", `{"Name": "Length"}`, "", "0")
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "dArr", 0, `{"Name": "Append", "Args": [10]}`),
			NewLuaTxCall("user1", "dArr", 0, `{"Name": "Append", "Args": [20]}`),
		)
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("dArr", `{"Name": "Get", "Args": [1]}`, "", "10")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("dArr", `{"Name": "Get", "Args": [2]}`, "", "20")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("dArr", `{"Name": "Get", "Args": [3]}`, "index out of range", "")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("dArr", `{"Name": "Length"}`, "", "2")
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "dArr", 0, `{"Name": "Append", "Args": [30]}`),
			NewLuaTxCall("user1", "dArr", 0, `{"Name": "Append", "Args": [40]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("dArr", `{"Name": "Length"}`, "", "4")
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "dArr", 0, `{"Name": "Set", "Args": [3, 50]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("dArr", `{"Name": "Get", "Args": [3]}`, "", "50")
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestTypeCrypto(t *testing.T) {
	code := readLuaCode(t, "type_crypto.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "crypto", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.Query("crypto", `{"Name": "get", "Args" : ["ab\u0000\u442a"]}`, "", `"0xc58f6dca13e4bba90a326d8605042862fe87c63a64a9dd0e95608a2ee68dc6f0"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("crypto", `{"Name": "get", "Args" : ["0x616200e490aa"]}`, "", `"0xc58f6dca13e4bba90a326d8605042862fe87c63a64a9dd0e95608a2ee68dc6f0"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("crypto", `{"Name": "checkEther", "Args" : []}`, "", `true`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("crypto", `{"Name": "checkAergo", "Args" : []}`, "", `true`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("crypto", `{"Name": "keccak256", "Args" : ["0x616263"]}`, "", `"0x4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("crypto", `{"Name": "keccak256", "Args" : ["0x616572676F"]}`, "", `"0xe98bb03ab37161f8bbfe131f711dcccf3002a9cd9ec31bbd52edf181f7ab09a0"`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestTypeBignum(t *testing.T) {
	bignum := readLuaCode(t, "type_bignum.lua")
	callee := readLuaCode(t, "type_bignum_callee.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "bigNum", uint64(types.Gaer)*50, bignum),
			NewLuaTxDeploy("user1", "add", 0, callee),
		)
		require.NoErrorf(t, err, "failed to deploy")

		tx := NewLuaTxCall("user1", "bigNum", 0, fmt.Sprintf(`{"Name":"test", "Args":["%s"]}`, nameToAddress("user1")))
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")

		receipt := bc.GetReceipt(tx.Hash())
		assert.Equalf(t, `"25000000000"`, receipt.GetRet(), "contract Call ret error")

		tx = NewLuaTxCall("user1", "bigNum", 0, fmt.Sprintf(`{"Name":"sendS", "Args":["%s"]}`, nameToAddress("user1")))
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")

		receipt = bc.GetReceipt(tx.Hash())
		assert.Equalf(t, `"23999900001"`, receipt.GetRet(), "contract Call ret error")

		tx = NewLuaTxCall("user1", "bigNum", 0, `{"Name":"testBignum", "Args":[]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")

		receipt = bc.GetReceipt(tx.Hash())
		assert.Equalf(t, `"999999999999999999999999999999"`, receipt.GetRet(), "contract Call ret error")

		err = bc.Query("bigNum", `{"Name":"argBignum", "Args":[{"_bignum":"99999999999999999999999999"}]}`, "", `"100000000000000000000000000"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("bigNum", fmt.Sprintf(`{"Name":"calladdBignum", "Args":["%s", {"_bignum":"999999999999999999"}]}`, nameToAddress("add")), "", `"1000000000000000004"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("bigNum", `{"Name":"checkBignum"}`, "", `[false,true,false]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("bigNum", `{"Name":"calcBignum"}`, "bignum divide by zero", "")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("bigNum", `{"Name":"negativeBignum"}`, "bignum not allowed negative value", "")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("bigNum", `{"Name":"byteBignum"}`, "", `{"_bignum":"177"}`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestBignumValues(t *testing.T) {
	code := readLuaCode(t, "bignum_values.lua")

	bc, err := LoadDummyChain(SetHardForkVersion(2))
	require.NoErrorf(t, err, "failed to create dummy chain")
	defer bc.Release()

	err = bc.ConnectBlock(
		NewLuaTxAccount("user1", 1, types.Aergo),
		NewLuaTxDeploy("user1", "contract1", 0, code),
	)
	require.NoErrorf(t, err, "failed to deploy")

	// hardfork 2

	// process octal, hex, binary

	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["0"]}`, "", `"0"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["9"]}`, "", `"9"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["0055"]}`, "", `"45"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["01234567"]}`, "", `"342391"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["0x123456789abcdef"]}`, "", `"81985529216486895"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["0b1010101010101"]}`, "", `"5461"`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":[{"_bignum":"0"}]}`, "", `"0"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":[{"_bignum":"9"}]}`, "", `"9"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":[{"_bignum":"01234567"}]}`, "", `"342391"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":[{"_bignum":"0x123456789abcdef"}]}`, "", `"81985529216486895"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":[{"_bignum":"0b1010101010101"}]}`, "", `"5461"`)
	require.NoErrorf(t, err, "failed to query")


	// hardfork 3
	bc.HardforkVersion = 3

	// block octal, hex and binary

	tx := NewLuaTxCall("user1", "contract1", 0, `{"Name":"parse_bignum", "Args":["01234567"]}`)
	err = bc.ConnectBlock(tx)
	require.NoErrorf(t, err, "failed to call tx")
	receipt := bc.GetReceipt(tx.Hash())
	assert.Equalf(t, `"1234567"`, receipt.GetRet(), "contract Call ret error")

	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["0"]}`, "", `"0"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["9"]}`, "", `"9"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["0055"]}`, "", `"55"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["01234567"]}`, "", `"1234567"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["0x123456789abcdef"]}`, "bignum invalid number string", `""`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["0b1010101010101"]}`, "bignum invalid number string", `""`)
	require.NoErrorf(t, err, "failed to query")

	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":[{"_bignum":"0"}]}`, "", `"0"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":[{"_bignum":"9"}]}`, "", `"9"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":[{"_bignum":"01234567"}]}`, "", `"1234567"`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":[{"_bignum":"0x123456789abcdef"}]}`, "bignum invalid number string", `""`)
	require.NoErrorf(t, err, "failed to query")
	err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":[{"_bignum":"0b1010101010101"}]}`, "bignum invalid number string", `""`)
	require.NoErrorf(t, err, "failed to query")


	// hardfork 4 and after

	for version := int32(4); version <= max_version; version++ {
		bc, err = LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "contract1", 0, code),
		)
		require.NoErrorf(t, err, "failed to deploy")

		// process hex, binary. block octal

		tx = NewLuaTxCall("user1", "contract1", 0, `{"Name":"parse_bignum", "Args":["01234567"]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")
		receipt = bc.GetReceipt(tx.Hash())
		assert.Equalf(t, `"1234567"`, receipt.GetRet(), "contract Call ret error")

		err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["0"]}`, "", `"0"`)
		require.NoErrorf(t, err, "failed to query")
		err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["9"]}`, "", `"9"`)
		require.NoErrorf(t, err, "failed to query")
		err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["0055"]}`, "", `"55"`)
		require.NoErrorf(t, err, "failed to query")
		err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["01234567"]}`, "", `"1234567"`)
		require.NoErrorf(t, err, "failed to query")
		err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["0x123456789abcdef"]}`, "", `"81985529216486895"`)
		require.NoErrorf(t, err, "failed to query")
		err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":["0b1010101010101"]}`, "", `"5461"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":[{"_bignum":"0"}]}`, "", `"0"`)
		require.NoErrorf(t, err, "failed to query")
		err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":[{"_bignum":"9"}]}`, "", `"9"`)
		require.NoErrorf(t, err, "failed to query")
		err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":[{"_bignum":"01234567"}]}`, "", `"1234567"`)
		require.NoErrorf(t, err, "failed to query")
		err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":[{"_bignum":"0x123456789abcdef"}]}`, "", `"81985529216486895"`)
		require.NoErrorf(t, err, "failed to query")
		err = bc.Query("contract1", `{"Name":"parse_bignum", "Args":[{"_bignum":"0b1010101010101"}]}`, "", `"5461"`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func checkRandomIntValue(v string, min, max int) error {
	n, _ := strconv.Atoi(v)
	if n < min || n > max {
		return errors.New("out of range")
	}
	return nil
}

func TestTypeRandom(t *testing.T) {
	code1 := readLuaCode(t, "type_random.lua")
	code2 := readLuaCode(t, "type_random_caller.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "random", 0, code1),
			NewLuaTxDeploy("user1", "caller", 0, code2),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "random", 0, `{"Name": "random", "Args":[]}`).Fail("1 or 2 arguments required"))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "random", 0, `{"Name": "random", "Args":[0]}`).Fail("the maximum value must be greater than zero"))
		require.NoErrorf(t, err, "failed to call tx")

		tx := NewLuaTxCall("user1", "random", 0, `{"Name": "random", "Args":[3]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")

		receipt := bc.GetReceipt(tx.Hash())
		err = checkRandomIntValue(receipt.GetRet(), 1, 3)
		require.NoErrorf(t, err, "failed to check random value")

		tx = NewLuaTxCall("user1", "random", 0, `{"Name": "random", "Args":[3, 10]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")

		receipt = bc.GetReceipt(tx.Hash())
		err = checkRandomIntValue(receipt.GetRet(), 3, 10)
		require.NoErrorf(t, err, "failed to check random value")

		err = bc.Query("random", `{"Name": "random", "Args":[1]}`, "", "1")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("random", `{"Name": "random", "Args":[4,4]}`, "", "4")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("random", `{"Name": "random", "Args":[0,4]}`, "system.random: the minimum value must be greater than zero", "")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("random", `{"Name": "random", "Args":[3,1]}`, "system.random: the maximum value must be greater than the minimum value", "")
		require.NoErrorf(t, err, "failed to query")

		tx = NewLuaTxCall("user1", "caller", 0, `{"Name": "check_if_equal", "Args":["`+nameToAddress("random")+`"]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")
		receipt = bc.GetReceipt(tx.Hash())
		assert.Equalf(t, `false`, receipt.GetRet(), "random numbers are the same on the same transaction")

	}
}

func TestTypeSparseTable(t *testing.T) {
	code := readLuaCode(t, "type_sparsetable.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		tx := NewLuaTxCall("user1", "r", 0, `{"Name":"r"}`)
		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "r", 0, code), tx)
		require.NoErrorf(t, err, "failed to new account, deploy, call")

		receipt := bc.GetReceipt(tx.Hash())
		require.Equalf(t, `1`, receipt.GetRet(), "contract Call ret error")

	}
}

func TestTypeJson(t *testing.T) {
	code := readLuaCode(t, "type_json.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "json", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "json", 0, `{"Name":"set", "Args":["[1,2,3]"]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", "[1,2,3]")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"[1,2,3]"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "json", 0, `{"Name":"set", "Args":["{\"key1\":[1,2,3], \"run\", \"key2\":5, [4,5,6]}"]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `{"1":"run","2":[4,5,6],"key1":[1,2,3],"key2":5}`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"{\"1\":\"run\",\"2\":[4,5,6],\"key1\":[1,2,3],\"key2\":5}"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "json", 0, `{"Name":"set", "Args":["{\"key1\":{\"arg1\": 1,\"arg2\":null, \"arg3\":[]}, \"key2\":[5,4,3]}"]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `{"key1":{"arg1":1,"arg3":{}},"key2":[5,4,3]}`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"{\"key1\":{\"arg1\":1,\"arg3\":{}},\"key2\":[5,4,3]}"`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "json", 0, `{"Name":"set", "Args":["{\"key1\":[1,2,3], \"key1\":5}"]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `{"key1":5}`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "json", 0, `{"Name":"set", "Args":["[\"\\\"hh\\t\",\"2\",3]"]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("json", `{"Name":"get", "Args":[]}`, "", `["\"hh\u0009","2",3]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("json", `{"Name":"getlen", "Args":[]}`, "", `["\"hh\u0009",4]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("json", `{"Name":"getenc", "Args":[]}`, "", `"[\"\\\"hh\\u0009\",\"2\",3]"`)
		require.NoErrorf(t, err, "failed to query")

		tx := NewLuaTxCall("user1", "json", 100, `{"Name":"getAmount"}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")
		receipt := bc.GetReceipt(tx.Hash())
		require.Equalf(t, `"100"`, receipt.GetRet(), "contract Call ret error")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "json", 0, `{"Name":"set", "Args":["{\"key1\":[1,2,3], \"key1\":5}}"]}`).Fail("not proper json format"))
		require.NoErrorf(t, err, "failed to call tx")

	}
}

// feature tests
func TestFeatureVote(t *testing.T) {
	code := readLuaCode(t, "feature_vote.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("owner", 1, types.Aergo),
			NewLuaTxDeploy("owner", "vote", 0, code),
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxAccount("user10", 1, types.Aergo),
			NewLuaTxAccount("user11", 1, types.Aergo),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(
			NewLuaTxCall("owner", "vote", 0, `{"Name":"addCandidate", "Args":["candidate1"]}`),
			NewLuaTxCall("owner", "vote", 0, `{"Name":"addCandidate", "Args":["candidate2"]}`),
			NewLuaTxCall("owner", "vote", 0, `{"Name":"addCandidate", "Args":["candidate3"]}`),
		)
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("vote", `{"Name":"getCandidates"}`, "", `[{"count":"0","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "vote", 0, `{"Name":"addCandidate", "Args":["candidate4"]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("vote", `{"Name":"getCandidates"}`, "", `[{"count":"0","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(
			// register voter
			NewLuaTxCall("owner", "vote", 0, fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, nameToAddress("user10"))),
			NewLuaTxCall("owner", "vote", 0, fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, nameToAddress("user10"))),
			NewLuaTxCall("owner", "vote", 0, fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, nameToAddress("user11"))),
			NewLuaTxCall("owner", "vote", 0, fmt.Sprintf(`{"Name":"registerVoter", "Args":["%s"]}`, nameToAddress("user1"))),
			// vote
			NewLuaTxCall("user1", "vote", 0, `{"Name":"vote", "Args":["user1"]}`),
			NewLuaTxCall("user1", "vote", 0, `{"Name":"vote", "Args":["user1"]}`),
			NewLuaTxCall("user1", "vote", 0, `{"Name":"vote", "Args":["user2"]}`),
			NewLuaTxCall("user1", "vote", 0, `{"Name":"vote", "Args":["user2"]}`),
			NewLuaTxCall("user1", "vote", 0, `{"Name":"vote", "Args":["user3"]}`),
		)
		require.NoErrorf(t, err, "failed to call tx | vote error")

		err = bc.Query("vote", `{"Name":"getCandidates"}`, "", `[{"count":"0","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(
			NewLuaTxCall("user11", "vote", 0, `{"Name":"vote", "Args":["candidate1"]}`),
			NewLuaTxCall("user10", "vote", 0, `{"Name":"vote", "Args":["candidate1"]}`),
		)
		require.NoErrorf(t, err, "failed to call tx | vote error")

		err = bc.Query("vote", `{"Name":"getCandidates"}`, "", `[{"count":"2","id":0,"name":"candidate1"},{"count":"0","id":1,"name":"candidate2"},{"count":"0","id":2,"name":"candidate3"}]`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestFeatureGovernance(t *testing.T) {
	code := readLuaCode(t, "feature_governance.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 40000, types.Aergo), NewLuaTxDeploy("user1", "gov", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		amount := types.NewAmount(40000, types.Aergo) // 40,000 aergo
		err = bc.ConnectBlock(NewLuaTxCallBig("user1", "gov", amount, `{"Name": "test_gov", "Args":[]}`))
		require.NoErrorf(t, err, "failed to call tx")

		oldstaking, err := bc.GetStaking("gov")
		require.NoErrorf(t, err, "failed to get staking")

		oldgov, err := bc.GetAccountState("gov")
		require.NoErrorf(t, err, "failed to get gov account state")

		tx := NewLuaTxCall("user1", "gov", 0, `{"Name": "test_pcall", "Args":[]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")

		staking, err := bc.GetStaking("gov")
		require.NoErrorf(t, err, "failed to get staking")

		gov, err := bc.GetAccountState("gov")
		require.NoErrorf(t, err, "failed to get gov account state")
		require.Equalf(t, oldstaking.Amount, staking.Amount, "pcall error, staking amount should be same")
		require.Equalf(t, oldgov.GetBalance(), gov.GetBalance(), "pcall error, gov balance should be same")

		tx = NewLuaTxCall("user1", "gov", 0, `{"Name": "error_case", "Args":[]}`)
		err = bc.ConnectBlock(tx)
		require.Errorf(t, err, "expect error | less time has passed")

		newstaking, err := bc.GetStaking("gov")
		require.NoErrorf(t, err, "failed to get staking")

		newgov, err := bc.GetAccountState("gov")
		require.NoErrorf(t, err, "failed to get gov account state")

		require.Equalf(t, oldstaking.Amount, newstaking.Amount, "pcall error, staking amount should be same")
		require.Equalf(t, oldgov.GetBalance(), newgov.GetBalance(), "pcall error, gov balance should be same")

	}
}

func TestFeaturePcallRollback(t *testing.T) {
	code1 := readLuaCode(t, "feature_pcall_rollback_1.lua")
	code2 := readLuaCode(t, "feature_pcall_rollback_2.lua")
	code3 := readLuaCode(t, "feature_pcall_rollback_3.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "counter", 10, code1).Constructor("[0]"),
			NewLuaTxCall("user1", "counter", 15, `{"Name":"inc", "Args":[]}`),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "1")
		require.NoErrorf(t, err, "failed to query")

		err = bc.ConnectBlock(
			NewLuaTxDeploy("user1", "caller", 10, code2).Constructor(fmt.Sprintf(`["%s"]`, nameToAddress("counter"))),
			NewLuaTxCall("user1", "caller", 15, `{"Name":"add", "Args":[]}`),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(NewLuaTxCall("user1", "caller", 0, `{"Name":"sql", "Args":[]}`))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("caller", `{"Name":"get", "Args":[]}`, "", "2")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("caller", `{"Name":"sqlget", "Args":[]}`, "", "2")
		require.NoErrorf(t, err, "failed to query")

		tx := NewLuaTxCall("user1", "caller", 0, `{"Name":"getOrigin", "Args":[]}`)
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")

		receipt := bc.GetReceipt(tx.Hash())
		require.Equalf(t, "\""+nameToAddress("user1")+"\"", receipt.GetRet(), "contract Call ret error")

		// create new dummy chain

		bc, err = LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxAccount("bong", 0, 0),
			NewLuaTxDeploy("user1", "counter", 0, code3),
		)
		require.NoErrorf(t, err, "failed to deploy")

		tx = NewLuaTxCall("user1", "counter", 20, fmt.Sprintf(`{"Name":"set", "Args":["%s"]}`, nameToAddress("bong")))
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "1")
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("counter", `{"Name":"getBalance", "Args":[]}`, "", "\"18\"")
		require.NoErrorf(t, err, "failed to query")

		state, err := bc.GetAccountState("bong")
		require.NoErrorf(t, err, "failed to get account state")
		assert.Equal(t, int64(2), state.GetBalanceBigInt().Int64(), "balance error")

		tx = NewLuaTxCall("user1", "counter", 10, fmt.Sprintf(`{"Name":"set2", "Args":["%s"]}`, nameToAddress("bong")))
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("counter", `{"Name":"get", "Args":[]}`, "", "2")
		require.NoErrorf(t, err, "failed to query")

		state, err = bc.GetAccountState("bong")
		require.NoErrorf(t, err, "failed to get account state")
		assert.Equal(t, int64(3), state.GetBalanceBigInt().Int64(), "balance error")

	}
}

func TestFeaturePcallNested(t *testing.T) {
	code := readLuaCode(t, "feature_pcall_nested.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 10, types.Aergo),
			NewLuaTxAccount("bong", 0, 0),
			NewLuaTxDeploy("user1", "pcall", uint64(types.Aergo)*10, code),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "pcall", 0, fmt.Sprintf(`{"Name":"pcall1", "Args":["%s", "%s"]}`,
				nameToAddress("pcall"), nameToAddress("bong"))),
		)
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.Query("pcall", fmt.Sprintf(`{"Name":"map", "Args":["%s"]}`, nameToAddress("pcall")), "", "2")
		require.NoErrorf(t, err, "failed to query")

		state, err := bc.GetAccountState("bong")
		require.NoErrorf(t, err, "failed to get account state")
		assert.Equal(t, int64(types.Aergo), state.GetBalanceBigInt().Int64(), "balance error")

	}
}

// test rollback of state variable and balance
func TestPcallStateRollback1(t *testing.T) {
	resolver := readLuaCode(t, "resolver.lua")

	for version := min_version; version <= max_version; version++ {

		files := make([]string, 0)
		files = append(files, "feature_pcall_rollback_4a.lua")   // contract.pcall
		if version >= 4 {
			files = append(files, "feature_pcall_rollback_4b.lua") // pcall
			files = append(files, "feature_pcall_rollback_4c.lua") // xpcall
		}

		// iterate over all files
		for _, file := range files {

			code := readLuaCode(t, file)

			bc, err := LoadDummyChain(SetHardForkVersion(version))
			require.NoErrorf(t, err, "failed to create dummy chain")
			defer bc.Release()

			// deploy and setup the name resolver
			err = bc.ConnectBlock(
				NewLuaTxAccount("user", 10, types.Aergo),
				NewLuaTxDeploy("user", "resolver", 0, resolver),
				NewLuaTxCall("user", "resolver", 0, fmt.Sprintf(`{"Name":"set","Args":["A","%s"]}`, nameToAddress("A"))),
				NewLuaTxCall("user", "resolver", 0, fmt.Sprintf(`{"Name":"set","Args":["B","%s"]}`, nameToAddress("B"))),
				NewLuaTxCall("user", "resolver", 0, fmt.Sprintf(`{"Name":"set","Args":["C","%s"]}`, nameToAddress("C"))),
			)
			require.NoErrorf(t, err, "failed to deploy and setup resolver")

			// deploy the contracts
			err = bc.ConnectBlock(
				NewLuaTxDeploy("user", "A", 3, code).Constructor(fmt.Sprintf(`["%s","A"]`, nameToAddress("resolver"))),
				NewLuaTxDeploy("user", "B", 0, code).Constructor(fmt.Sprintf(`["%s","B"]`, nameToAddress("resolver"))),
				NewLuaTxDeploy("user", "C", 0, code).Constructor(fmt.Sprintf(`["%s","C"]`, nameToAddress("resolver"))),
			)
			require.NoErrorf(t, err, "failed to deploy the contracts")

			// A -> A -> A (3 calls on the same contract)

			script := `[[
				['set','x',111],
				['pcall','A']
			],[
				['set','x',222],
				['pcall','A']
			],[
				['set','x',333]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 333}, nil)

			script = `[[
				['set','x',111],
				['pcall','A']
			],[
				['set','x',222],
				['pcall','A']
			],[
				['set','x',333],
				['fail']
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 222}, nil)

			script = `[[
				['set','x',111],
				['pcall','A']
			],[
				['set','x',222],
				['pcall','A'],
				['fail']
			],[
				['set','x',333]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111}, nil)

			script = `[[
				['set','x',111],
				['pcall','A'],
				['fail']
			],[
				['set','x',222],
				['pcall','A']
			],[
				['set','x',333]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 0}, nil)

			// A -> B -> C (3 different contracts)

			script = `[[
				['set','x',111],
				['pcall','B',2]
			],[
				['set','x',222],
				['pcall','C',1]
			],[
				['set','x',333]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 222, "C": 333},
				map[string]int64{"A": 1, "B": 1, "C": 1})

			script = `[[
				['set','x',111],
				['pcall','B',2]
			],[
				['set','x',222],
				['pcall','C',1]
			],[
				['set','x',333],
				['fail']
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 222, "C": 0},
				map[string]int64{"A": 1, "B": 2, "C": 0})

			script = `[[
				['set','x',111],
				['pcall','B',2]
			],[
				['set','x',222],
				['pcall','C',1],
				['fail']
			],[
				['set','x',333]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 0, "C": 0},
				map[string]int64{"A": 3, "B": 0, "C": 0})

			script = `[[
				['set','x',111],
				['pcall','B',2],
				['fail']
			],[
				['set','x',222],
				['pcall','C',1]
			],[
				['set','x',333]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 0, "B": 0, "C": 0},
				map[string]int64{"A": 3, "B": 0, "C": 0})

			// A -> B -> A (call back to original contract)

			script = `[[
				['set','x',111],
				['pcall','B',2]
			],[
				['set','x',222],
				['pcall','A',1]
			],[
				['set','x',333]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 333, "B": 222},
				map[string]int64{"A": 2, "B": 1})

			script = `[[
				['set','x',111],
				['pcall','B',2]
			],[
				['set','x',222],
				['pcall','A',1]
			],[
				['set','x',333],
				['fail']
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 222},
				map[string]int64{"A": 1, "B": 2})

			script = `[[
				['set','x',111],
				['pcall','B',2]
			],[
				['set','x',222],
				['pcall','A',1],
				['fail']
			],[
				['set','x',333]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 0},
				map[string]int64{"A": 3, "B": 0})

			script = `[[
				['set','x',111],
				['pcall','B',2],
				['fail']
			],[
				['set','x',222],
				['pcall','A',1]
			],[
				['set','x',333]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 0, "B": 0},
				map[string]int64{"A": 3, "B": 0})

			// A -> B -> B

			script = `[[
				['set','x',111],
				['pcall','B',3]
			],[
				['set','x',222],
				['pcall','B']
			],[
				['set','x',333]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 333},
				map[string]int64{"A": 0, "B": 3})

			script = `[[
				['set','x',111],
				['pcall','B',3]
			],[
				['set','x',222],
				['pcall','B']
			],[
				['set','x',333],
				['fail']
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 222},
				map[string]int64{"A": 0, "B": 3})

			script = `[[
				['set','x',111],
				['pcall','B',3]
			],[
				['set','x',222],
				['pcall','B'],
				['fail']
			],[
				['set','x',333]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 0},
				map[string]int64{"A": 3, "B": 0})

			script = `[[
				['set','x',111],
				['pcall','B',3],
				['fail']
			],[
				['set','x',222],
				['pcall','B']
			],[
				['set','x',333]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 0, "B": 0},
				map[string]int64{"A": 3, "B": 0})

			// A -> A -> B

			script = `[[
				['set','x',111],
				['pcall','A']
			],[
				['set','x',222],
				['pcall','B',3]
			],[
				['set','x',333]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 222, "B": 333},
				map[string]int64{"A": 0, "B": 3})

			script = `[[
				['set','x',111],
				['pcall','A']
			],[
				['set','x',222],
				['pcall','B',3]
			],[
				['set','x',333],
				['fail']
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 222, "B": 0},
				map[string]int64{"A": 3, "B": 0})

			script = `[[
				['set','x',111],
				['pcall','A']
			],[
				['set','x',222],
				['pcall','B',3],
				['fail']
			],[
				['set','x',333]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 0},
				map[string]int64{"A": 3, "B": 0})

			script = `[[
				['set','x',111],
				['pcall','A'],
				['fail']
			],[
				['set','x',222],
				['pcall','B',3]
			],[
				['set','x',333]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 0, "B": 0},
				map[string]int64{"A": 3, "B": 0})

			// A -> B -> A -> B -> A  (zigzag)

			script = `[[
				['set','x',111],
				['pcall','B',1]
			],[
				['set','x',222],
				['pcall','A',1]
			],[
				['set','x',333],
				['pcall','B',1]
			],[
				['set','x',444],
				['pcall','A',1]
			],[
				['set','x',555]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 555, "B": 444},
				map[string]int64{"A": 3, "B": 0})

			script = `[[
				['set','x',111],
				['pcall','B',1]
			],[
				['set','x',222],
				['pcall','A',1]
			],[
				['set','x',333],
				['pcall','B',1]
			],[
				['set','x',444],
				['pcall','A',1]
			],[
				['set','x',555],
				['fail']
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 333, "B": 444},
				map[string]int64{"A": 2, "B": 1})

			script = `[[
				['set','x',111],
				['pcall','B',1]
			],[
				['set','x',222],
				['pcall','A',1]
			],[
				['set','x',333],
				['pcall','B',1]
			],[
				['set','x',444],
				['pcall','A',1],
				['fail']
			],[
				['set','x',555]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 333, "B": 222},
				map[string]int64{"A": 3, "B": 0})

			script = `[[
				['set','x',111],
				['pcall','B',1]
			],[
				['set','x',222],
				['pcall','A',1]
			],[
				['set','x',333],
				['pcall','B',1],
				['fail']
			],[
				['set','x',444],
				['pcall','A',1]
			],[
				['set','x',555]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 222},
				map[string]int64{"A": 2, "B": 1})

			script = `[[
				['set','x',111],
				['pcall','B',1]
			],[
				['set','x',222],
				['pcall','A',1],
				['fail']
			],[
				['set','x',333],
				['pcall','B',1]
			],[
				['set','x',444],
				['pcall','A',1]
			],[
				['set','x',555]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 0},
				map[string]int64{"A": 3, "B": 0})

			script = `[[
				['set','x',111],
				['pcall','B',1],
				['fail']
			],[
				['set','x',222],
				['pcall','A',1]
			],[
				['set','x',333],
				['pcall','B',1]
			],[
				['set','x',444],
				['pcall','A',1]
			],[
				['set','x',555]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 0, "B": 0},
				map[string]int64{"A": 3, "B": 0})

		}
	}
}

// test rollback of state variable and balance - send separate from call
func TestPcallStateRollback2(t *testing.T) {
	t.Skip("disabled until bug with test is fixed")
	resolver := readLuaCode(t, "resolver.lua")

	for version := min_version; version <= max_version; version++ {
		files := make([]string, 0)
		files = append(files, "feature_pcall_rollback_4a.lua")   // contract.pcall
		if version >= 4 {
			files = append(files, "feature_pcall_rollback_4b.lua") // pcall
			files = append(files, "feature_pcall_rollback_4c.lua") // xpcall
		}

		// iterate over all files
		for _, file := range files {

			code := readLuaCode(t, file)

			bc, err := LoadDummyChain(SetHardForkVersion(version))
			require.NoErrorf(t, err, "failed to create dummy chain")
			defer bc.Release()

			// deploy and setup the name resolver
			err = bc.ConnectBlock(
				NewLuaTxAccount("user", 10, types.Aergo),
				NewLuaTxDeploy("user", "resolver", 0, resolver),
				NewLuaTxCall("user", "resolver", 0, fmt.Sprintf(`{"Name":"set","Args":["A","%s"]}`, nameToAddress("A"))),
				NewLuaTxCall("user", "resolver", 0, fmt.Sprintf(`{"Name":"set","Args":["B","%s"]}`, nameToAddress("B"))),
				NewLuaTxCall("user", "resolver", 0, fmt.Sprintf(`{"Name":"set","Args":["C","%s"]}`, nameToAddress("C"))),
				NewLuaTxCall("user", "resolver", 0, fmt.Sprintf(`{"Name":"set","Args":["D","%s"]}`, nameToAddress("D"))),
				NewLuaTxCall("user", "resolver", 0, fmt.Sprintf(`{"Name":"set","Args":["E","%s"]}`, nameToAddress("E"))),
			)
			require.NoErrorf(t, err, "failed to deploy and setup resolver")

			// deploy the contracts
			err = bc.ConnectBlock(
				NewLuaTxDeploy("user", "A", 3, code).Constructor(fmt.Sprintf(`["%s","A"]`, nameToAddress("resolver"))),
				NewLuaTxDeploy("user", "B", 0, code).Constructor(fmt.Sprintf(`["%s","B"]`, nameToAddress("resolver"))),
				NewLuaTxDeploy("user", "C", 0, code).Constructor(fmt.Sprintf(`["%s","C"]`, nameToAddress("resolver"))),
			)
			require.NoErrorf(t, err, "failed to deploy the contracts")

			// A -> A -> A (3 calls on the same contract)

			script := `[[
				['set','x',111],
				['send','B',1],
				['pcall','A']
			],[
				['set','x',222],
				['send','C',1],
				['pcall','A']
			],[
				['set','x',333],
				['send','E',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 333},
				map[string]int64{"A": 0, "B": 1, "C": 1, "E": 1})

			script = `[[
				['set','x',111],
				['send','B',1],
				['pcall','A']
			],[
				['set','x',222],
				['send','C',1],
				['pcall','A']
			],[
				['set','x',333],
				['send','D',1],
				['fail']
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 222},
				map[string]int64{"A": 1, "B": 1, "C": 1, "D": 0})

			script = `[[
				['set','x',111],
				['send','B',1],
				['pcall','A']
			],[
				['set','x',222],
				['send','C',1],
				['pcall','A'],
				['fail']
			],[
				['set','x',333],
				['send','D',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111},
				map[string]int64{"A": 2, "B": 1, "C": 0, "D": 0})

			script = `[[
				['set','x',111],
				['send','B',1],
				['pcall','A'],
				['fail']
			],[
				['set','x',222],
				['send','C',1],
				['pcall','A']
			],[
				['set','x',333],
				['send','D',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 0},
				map[string]int64{"A": 3, "B": 0, "C": 0, "D": 0})

			// A -> B -> C (3 different contracts)

			script = `[[
				['set','x',111],
				['send','B',3],
				['pcall','B']
			],[
				['set','x',222],
				['send','C',2],
				['pcall','C']
			],[
				['set','x',333],
				['send','A',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 222, "C": 333},
				map[string]int64{"A": 1, "B": 1, "C": 1})

			script = `[[
				['set','x',111],
				['send','B',3],
				['pcall','B']
			],[
				['set','x',222],
				['send','C',2],
				['pcall','C']
			],[
				['set','x',333],
				['send','A',1],
				['fail']
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 222, "C": 0},
				map[string]int64{"A": 0, "B": 1, "C": 2})

			script = `[[
				['set','x',111],
				['send','B',3],
				['pcall','B']
			],[
				['set','x',222],
				['send','C',2],
				['pcall','C'],
				['fail']
			],[
				['set','x',333],
				['send','A',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 0, "C": 0},
				map[string]int64{"A": 0, "B": 3, "C": 0})

			script = `[[
				['set','x',111],
				['send','B',3],
				['pcall','B'],
				['fail']
			],[
				['set','x',222],
				['send','C',2],
				['pcall','C']
			],[
				['set','x',333],
				['send','A',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 0, "B": 0, "C": 0},
				map[string]int64{"A": 3, "B": 0, "C": 0})

			// A -> B -> A (call back to original contract)

			script = `[[
				['set','x',111],
				['send','B',3],
				['pcall','B']
			],[
				['set','x',222],
				['send','A',2],
				['pcall','A']
			],[
				['set','x',333],
				['send','B',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 333, "B": 222},
				map[string]int64{"A": 1, "B": 2})

			script = `[[
				['set','x',111],
				['send','B',3],
				['pcall','B']
			],[
				['set','x',222],
				['send','A',2],
				['pcall','A']
			],[
				['set','x',333],
				['send','B',1],
				['fail']
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 222},
				map[string]int64{"A": 2, "B": 1})

			script = `[[
				['set','x',111],
				['send','B',3],
				['pcall','B']
			],[
				['set','x',222],
				['send','A',2],
				['pcall','A'],
				['fail']
			],[
				['set','x',333],
				['send','B',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 0},
				map[string]int64{"A": 0, "B": 3})

			script = `[[
				['set','x',111],
				['send','B',3],
				['pcall','B'],
				['fail']
			],[
				['set','x',222],
				['send','A',2],
				['pcall','A']
			],[
				['set','x',333],
				['send','B',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 0, "B": 0},
				map[string]int64{"A": 3, "B": 0})

			// A -> B -> B

			script = `[[
				['set','x',111],
				['send','B',3],
				['pcall','B']
			],[
				['set','x',222],
				['send','C',1],
				['pcall','B']
			],[
				['set','x',333],
				['send','A',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 333},
				map[string]int64{"A": 1, "B": 1, "C": 1})

			script = `[[
				['set','x',111],
				['send','B',3],
				['pcall','B']
			],[
				['set','x',222],
				['send','C',1],
				['pcall','B']
			],[
				['set','x',333],
				['send','A',1],
				['fail']
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 222},
				map[string]int64{"A": 0, "B": 2, "C": 1})

			script = `[[
				['set','x',111],
				['send','B',3],
				['pcall','B']
			],[
				['set','x',222],
				['send','C',1],
				['pcall','B'],
				['fail']
			],[
				['set','x',333],
				['send','A',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 0},
				map[string]int64{"A": 0, "B": 3, "C": 0})

			script = `[[
				['set','x',111],
				['send','B',3],
				['pcall','B'],
				['fail']
			],[
				['set','x',222],
				['send','C',1],
				['pcall','B']
			],[
				['set','x',333],
				['send','A',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 0, "B": 0},
				map[string]int64{"A": 3, "B": 0, "C": 0})

			// A -> A -> B

			script = `[[
				['set','x',111],
				['send','B',2],
				['pcall','A']
			],[
				['set','x',222],
				['send','C',1],
				['pcall','B']
			],[
				['set','x',333],
				['send','A',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 222, "B": 333},
				map[string]int64{"A": 1, "B": 1, "C": 1})

			script = `[[
				['set','x',111],
				['send','B',2],
				['pcall','A']
			],[
				['set','x',222],
				['send','C',1],
				['pcall','B']
			],[
				['set','x',333],
				['send','A',1],
				['fail']
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 222, "B": 0},
				map[string]int64{"A": 0, "B": 2, "C": 1})

			script = `[[
				['set','x',111],
				['send','B',2],
				['pcall','A']
			],[
				['set','x',222],
				['send','C',1],
				['pcall','B'],
				['fail']
			],[
				['set','x',333],
				['send','A',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 0},
				map[string]int64{"A": 1, "B": 2, "C": 0})

			script = `[[
				['set','x',111],
				['send','B',2],
				['pcall','A'],
				['fail']
			],[
				['set','x',222],
				['send','C',1],
				['pcall','B']
			],[
				['set','x',333],
				['send','A',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 0, "B": 0},
				map[string]int64{"A": 3, "B": 0, "C": 0})

			// A -> B -> A -> B -> A  (zigzag)

			script = `[[
				['set','x',111],
				['send','B',1],
				['pcall','B']
			],[
				['set','x',222],
				['send','A',1],
				['pcall','A']
			],[
				['set','x',333],
				['send','B',1],
				['pcall','B']
			],[
				['set','x',444],
				['send','A',1],
				['pcall','A']
			],[
				['set','x',555],
				['send','B',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 555, "B": 444},
				map[string]int64{"A": 2, "B": 1})

			script = `[[
				['set','x',111],
				['send','B',1],
				['pcall','B']
			],[
				['set','x',222],
				['send','A',1],
				['pcall','A']
			],[
				['set','x',333],
				['send','B',1],
				['pcall','B']
			],[
				['set','x',444],
				['send','A',1],
				['pcall','A']
			],[
				['set','x',555],
				['send','B',1],
				['fail']
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 333, "B": 444},
				map[string]int64{"A": 3, "B": 0})

			script = `[[
				['set','x',111],
				['send','B',1],
				['pcall','B']
			],[
				['set','x',222],
				['send','A',1],
				['pcall','A']
			],[
				['set','x',333],
				['send','B',1],
				['pcall','B']
			],[
				['set','x',444],
				['send','A',1],
				['pcall','A'],
				['fail']
			],[
				['set','x',555],
				['send','B',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 333, "B": 222},
				map[string]int64{"A": 2, "B": 1})

			script = `[[
				['set','x',111],
				['send','B',1],
				['pcall','B']
			],[
				['set','x',222],
				['send','A',1],
				['pcall','A']
			],[
				['set','x',333],
				['send','B',1],
				['pcall','B'],
				['fail']
			],[
				['set','x',444],
				['send','A',1],
				['pcall','A']
			],[
				['set','x',555],
				['send','B',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 222},
				map[string]int64{"A": 3, "B": 0})

			script = `[[
				['set','x',111],
				['send','B',1],
				['pcall','B']
			],[
				['set','x',222],
				['send','A',1],
				['pcall','A'],
				['fail']
			],[
				['set','x',333],
				['send','B',1],
				['pcall','B']
			],[
				['set','x',444],
				['send','A',1],
				['pcall','A']
			],[
				['set','x',555],
				['send','B',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 0},
				map[string]int64{"A": 2, "B": 1})

			script = `[[
				['set','x',111],
				['send','B',1],
				['pcall','B'],
				['fail']
			],[
				['set','x',222],
				['send','A',1],
				['pcall','A']
			],[
				['set','x',333],
				['send','B',1],
				['pcall','B']
			],[
				['set','x',444],
				['send','A',1],
				['pcall','A']
			],[
				['set','x',555],
				['send','B',1]
			]]`
			testStateRollback(t, bc, script,
				map[string]int{"A": 0, "B": 0},
				map[string]int64{"A": 3, "B": 0})

		}
	}
}

// test rollback of db
func TestPcallStateRollback3(t *testing.T) {
	t.Skip("disabled until bug with test is fixed")
	resolver := readLuaCode(t, "resolver.lua")

	for version := min_version; version <= max_version; version++ {
		files := make([]string, 0)
		files = append(files, "feature_pcall_rollback_4a.lua")   // contract.pcall
		if version >= 4 {
			files = append(files, "feature_pcall_rollback_4b.lua") // pcall
			files = append(files, "feature_pcall_rollback_4c.lua") // xpcall
		}

		// iterate over all files
		for _, file := range files {

			code := readLuaCode(t, file)

			bc, err := LoadDummyChain(SetHardForkVersion(version))
			require.NoErrorf(t, err, "failed to create dummy chain")
			defer bc.Release()

			err = bc.ConnectBlock(
				NewLuaTxAccount("user", 1, types.Aergo),
				NewLuaTxDeploy("user", "resolver", 0, resolver),
				NewLuaTxDeploy("user", "A", 0, code).Constructor(fmt.Sprintf(`["%s","A"]`, nameToAddress("resolver"))),
				NewLuaTxDeploy("user", "B", 0, code).Constructor(fmt.Sprintf(`["%s","B"]`, nameToAddress("resolver"))),
				NewLuaTxDeploy("user", "C", 0, code).Constructor(fmt.Sprintf(`["%s","C"]`, nameToAddress("resolver"))),
			)
			require.NoErrorf(t, err, "failed to deploy")

			err = bc.ConnectBlock(
				NewLuaTxCall("user", "resolver", 0, fmt.Sprintf(`{"Name":"set","Args":["A","%s"]}`, nameToAddress("A"))),
				NewLuaTxCall("user", "resolver", 0, fmt.Sprintf(`{"Name":"set","Args":["B","%s"]}`, nameToAddress("B"))),
				NewLuaTxCall("user", "resolver", 0, fmt.Sprintf(`{"Name":"set","Args":["C","%s"]}`, nameToAddress("C"))),
			)
			require.NoErrorf(t, err, "failed to call resolver contract")

			// A -> A -> A (3 calls on the same contract)

			script := `[[
				['db.set',111],
				['pcall','A']
			],[
				['db.set',222],
				['pcall','A']
			],[
				['db.set',333]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 333})

			script = `[[
				['db.set',111],
				['pcall','A']
			],[
				['db.set',222],
				['pcall','A']
			],[
				['db.set',333],
				['fail']
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 222})

			script = `[[
				['db.set',111],
				['pcall','A']
			],[
				['db.set',222],
				['pcall','A'],
				['fail']
			],[
				['db.set',333]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 111})

			script = `[[
				['db.set',111],
				['pcall','A'],
				['fail']
			],[
				['db.set',222],
				['pcall','A']
			],[
				['db.set',333]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 0})

			// A -> B -> C (3 different contracts)

			script = `[[
				['db.set',111],
				['pcall','B']
			],[
				['db.set',222],
				['pcall','C']
			],[
				['db.set',333]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 222, "C": 333})

			script = `[[
				['db.set',111],
				['pcall','B']
			],[
				['db.set',222],
				['pcall','C']
			],[
				['db.set',333],
				['fail']
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 222, "C": 0})

			script = `[[
				['db.set',111],
				['pcall','B']
			],[
				['db.set',222],
				['pcall','C'],
				['fail']
			],[
				['db.set',333]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 0, "C": 0})

			script = `[[
				['db.set',111],
				['pcall','B'],
				['fail']
			],[
				['db.set',222],
				['pcall','C']
			],[
				['db.set',333]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 0, "B": 0, "C": 0})

			// A -> B -> A (call back to original contract)

			script = `[[
				['db.set',111],
				['pcall','B']
			],[
				['db.set',222],
				['pcall','A']
			],[
				['db.set',333]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 333, "B": 222})

			script = `[[
				['db.set',111],
				['pcall','B']
			],[
				['db.set',222],
				['pcall','A']
			],[
				['db.set',333],
				['fail']
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 222})

			script = `[[
				['db.set',111],
				['pcall','B']
			],[
				['db.set',222],
				['pcall','A'],
				['fail']
			],[
				['db.set',333]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 0})

			script = `[[
				['db.set',111],
				['pcall','B'],
				['fail']
			],[
				['db.set',222],
				['pcall','A']
			],[
				['db.set',333]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 0, "B": 0})

			// A -> B -> B

			script = `[[
				['db.set',111],
				['pcall','B']
			],[
				['db.set',222],
				['pcall','B']
			],[
				['db.set',333]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 333})

			script = `[[
				['db.set',111],
				['pcall','B']
			],[
				['db.set',222],
				['pcall','B']
			],[
				['db.set',333],
				['fail']
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 222})

			script = `[[
				['db.set',111],
				['pcall','B']
			],[
				['db.set',222],
				['pcall','B'],
				['fail']
			],[
				['db.set',333]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 0})

			script = `[[
				['db.set',111],
				['pcall','B'],
				['fail']
			],[
				['db.set',222],
				['pcall','B']
			],[
				['db.set',333]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 0, "B": 0})

			// A -> A -> B

			script = `[[
				['db.set',111],
				['pcall','A']
			],[
				['db.set',222],
				['pcall','B']
			],[
				['db.set',333]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 222, "B": 333})

			script = `[[
				['db.set',111],
				['pcall','A']
			],[
				['db.set',222],
				['pcall','B']
			],[
				['db.set',333],
				['fail']
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 222, "B": 0})

			script = `[[
				['db.set',111],
				['pcall','A']
			],[
				['db.set',222],
				['pcall','B'],
				['fail']
			],[
				['db.set',333]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 0})

			script = `[[
				['db.set',111],
				['pcall','A'],
				['fail']
			],[
				['db.set',222],
				['pcall','B']
			],[
				['db.set',333]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 0, "B": 0})

			// A -> B -> A -> B -> A  (zigzag)

			script = `[[
				['db.set',111],
				['pcall','B']
			],[
				['db.set',222],
				['pcall','A']
			],[
				['db.set',333],
				['pcall','B']
			],[
				['db.set',444],
				['pcall','A']
			],[
				['db.set',555]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 555, "B": 444})

			script = `[[
				['db.set',111],
				['pcall','B']
			],[
				['db.set',222],
				['pcall','A']
			],[
				['db.set',333],
				['pcall','B']
			],[
				['db.set',444],
				['pcall','A']
			],[
				['db.set',555],
				['fail']
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 333, "B": 444})

			script = `[[
				['db.set',111],
				['pcall','B']
			],[
				['db.set',222],
				['pcall','A']
			],[
				['db.set',333],
				['pcall','B']
			],[
				['db.set',444],
				['pcall','A'],
				['fail']
			],[
				['db.set',555]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 333, "B": 222})

			script = `[[
				['db.set',111],
				['pcall','B']
			],[
				['db.set',222],
				['pcall','A']
			],[
				['db.set',333],
				['pcall','B'],
				['fail']
			],[
				['db.set',444],
				['pcall','A']
			],[
				['db.set',555]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 222})

			script = `[[
				['db.set',111],
				['pcall','B']
			],[
				['db.set',222],
				['pcall','A'],
				['fail']
			],[
				['db.set',333],
				['pcall','B']
			],[
				['db.set',444],
				['pcall','A']
			],[
				['db.set',555]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 111, "B": 0})

			script = `[[
				['db.set',111],
				['pcall','B'],
				['fail']
			],[
				['db.set',222],
				['pcall','A']
			],[
				['db.set',333],
				['pcall','B']
			],[
				['db.set',444],
				['pcall','A']
			],[
				['db.set',555]
			]]`
			testDbStateRollback(t, bc, script,
				map[string]int{"A": 0, "B": 0})

		}
	}
}

func testStateRollback(t *testing.T, bc *DummyChain, script string, expected_state map[string]int, expected_amount map[string]int64) {
	t.Helper()

	for n := 1; n <= 2; n++ {

		err := bc.ConnectBlock(
			NewLuaTxCall("user", "A", 0, `{"Name":"set","Args":["x",0]}`),
			NewLuaTxCall("user", "B", 0, `{"Name":"set","Args":["x",0]}`),
			NewLuaTxCall("user", "C", 0, `{"Name":"set","Args":["x",0]}`),
			NewLuaTxCall("user", "B", 0, `{"Name":"send_back","Args":["A"]}`),
			NewLuaTxCall("user", "C", 0, `{"Name":"send_back","Args":["A"]}`),
		)
		require.NoErrorf(t, err, "failed to reset")

		account, _ := bc.GetAccountState("A")
		if account.GetBalanceBigInt().Int64() != 3 {
			amount := uint64(3 - account.GetBalanceBigInt().Int64())
			err = bc.ConnectBlock(
				NewLuaTxSendBig("user", "A", types.NewAmount(amount, types.Aer)),
			)
			require.NoErrorf(t, err, "failed to send")
		}

		names := make(map[string]int64)
		names["A"] = 3
		names["B"] = 0
		names["C"] = 0
		for name, amount := range names {
			account, _ := bc.GetAccountState(name)
			assert.Equal(t, amount, account.GetBalanceBigInt().Int64(), "balance of "+name+" is not reset")
			err = bc.Query(name, `{"Name":"get", "Args":["x"]}`, "", "0")
			require.NoErrorf(t, err, "failed to query on reset")
		}

		script = strings.ReplaceAll(script, "'", "\"")
		var tx LuaTxTester
		if n == 1 {
			tx = NewLuaTxCall("user", "A", 0, fmt.Sprintf(`{"Name":"test", "Args":[%s]}`, script))
		} else {
			tx = NewLuaTxMultiCall("user", fmt.Sprintf(`[["call","%s","test",%s]]`, nameToAddress("A"), script))
		}
		err = bc.ConnectBlock(tx)
		//require.NoErrorf(t, err, "failed to call tx")
		//receipt := bc.GetReceipt(tx.Hash())
		//assert.Equal(t, ``, receipt.GetRet(), "receipt ret error")
		//fmt.Printf("events: %v\n", receipt.GetEvents())

		for contract, value := range expected_state {
			err = bc.Query(contract, `{"Name":"get", "Args":["x"]}`, "", fmt.Sprintf("%d", value))
			require.NoErrorf(t, err, "query failed")
		}

		for name, amount := range expected_amount {
			account, err := bc.GetAccountState(name)
			require.NoErrorf(t, err, "failed to get account state")
			assert.Equal(t, amount, account.GetBalanceBigInt().Int64(), "balance is different")
		}
	}
}

func testDbStateRollback(t *testing.T, bc *DummyChain, script string, expected map[string]int) {
	t.Helper()

	for n := 1; n <= 2; n++ {

		err := bc.ConnectBlock(
			NewLuaTxCall("user", "A", 0, `{"Name":"db_reset"}`),
			NewLuaTxCall("user", "B", 0, `{"Name":"db_reset"}`),
			NewLuaTxCall("user", "C", 0, `{"Name":"db_reset"}`),
		)
		require.NoErrorf(t, err, "failed to reset")

		names := []string{"A", "B", "C"}
		for _, name := range names {
			err = bc.Query(name, `{"Name":"db_get"}`, "", "0")
			require.NoErrorf(t, err, "failed to query on reset")
		}

		script = strings.ReplaceAll(script, "'", "\"")
		var tx LuaTxTester
		if n == 1 {
			tx = NewLuaTxCall("user", "A", 0, fmt.Sprintf(`{"Name":"test", "Args":[%s]}`, script))
		} else {
			tx = NewLuaTxMultiCall("user", fmt.Sprintf(`[["call","%s","test",%s]]`, nameToAddress("A"), script))
		}
		err = bc.ConnectBlock(tx)
		//require.NoErrorf(t, err, "failed to call tx")
		//receipt := bc.GetReceipt(tx.Hash())
		//assert.Equal(t, ``, receipt.GetRet(), "receipt ret error")
		//fmt.Printf("events: %v\n", receipt.GetEvents())

		for contract, value := range expected {
			err = bc.Query(contract, `{"Name":"db_get"}`, "", fmt.Sprintf("%d", value))
			require.NoErrorf(t, err, "query failed")
		}
	}
}

func TestFeatureLuaCryptoVerifyProof(t *testing.T) {
	code := readLuaCode(t, "feature_crypto_verify_proof.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(NewLuaTxAccount("user1", 1, types.Aergo), NewLuaTxDeploy("user1", "eth", 0, code))
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.Query("eth", `{"Name":"verifyProofRaw"}`, "", `true`)
		require.NoErrorf(t, err, "failed to query")

		err = bc.Query("eth", `{"Name":"verifyProofHex"}`, "", `true`)
		require.NoErrorf(t, err, "failed to query")

	}
}

func TestFeatureFeeDelegation(t *testing.T) {
	code := readLuaCode(t, "feature_feedelegation_1.lua")
	code2 := readLuaCode(t, "feature_feedelegation_2.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetPubNet(), SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccountBig("user1", types.NewAmount(100, types.Aergo)),
			NewLuaTxAccount("user2", 0, 0),
			NewLuaTxDeploy("user1", "fd", 0, code),
			NewLuaTxSendBig("user1", "fd", types.NewAmount(50, types.Aergo)),
		)
		require.NoErrorf(t, err, "failed to deploy")

		err = bc.ConnectBlock(NewLuaTxCallFeeDelegate("user2", "fd", 0, `{"Name": "check_delegation", "Args":[]}`).Fail("check_delegation function is not declared of fee delegation"))
		require.NoErrorf(t, err, "failed to call check_delegation")

		err = bc.ConnectBlock(NewLuaTxCall("user2", "fd", 0, `{"Name": "query", "Args":[]}`).Fail("not enough balance"))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxCallFeeDelegate("user2", "fd", 0, `{"Name": "query", "Args":[]}`).Fail("fee delegation is not allowed"))
		require.NoErrorf(t, err, "failed to call tx")

		contract1, err := bc.GetAccountState("fd")
		require.NoErrorf(t, err, "failed to get contract")

		tx := NewLuaTxCallFeeDelegate("user2", "fd", 0, `{"Name": "query", "Args":["arg"]}`)
		err = bc.ConnectBlock(
			NewLuaTxCall("user1", "fd", 0, fmt.Sprintf(`{"Name":"reg", "Args":["%s"]}`, nameToAddress("user2"))),
			tx,
		)
		require.NoErrorf(t, err, "failed to call tx")

		contract2, err := bc.GetAccountState("fd")
		require.NoErrorf(t, err, "failed to get contract")
		require.NotEqualf(t, contract1.GetBalanceBigInt().Int64(), contract2.GetBalanceBigInt().Int64(), "balance is not changed")

		err = bc.ConnectBlock(tx.Fail("fee delegation is not allowed"))
		require.NoErrorf(t, err, "failed to call tx")

		err = bc.ConnectBlock(NewLuaTxDeploy("user1", "fd2", 0, code2))
		require.Errorf(t, err, "expect error")
		require.Containsf(t, err.Error(), "no 'check_delegation' function", "invalid error message")

	}
}

/*
func TestFeatureFeeDelegationLoop(t *testing.T) {
	definition := `
	state.var{
        whitelist = state.map(),
    }

    function query_no(a)
		if (system.isFeeDelegation() == true) then
        	whitelist[system.getSender()] = false
		end
        return 1,2,3,4,5
    end
	function default()
	end
    function check_delegation(fname,k)
		return true
    end
    abi.payable(default)
	abi.fee_delegation(query_no)
`
	for version := min_version; version <= max_version; version++ {
	bc, err := LoadDummyChain(OnPubNet, SetHardForkVersion(version))
	if err != nil {
		t.Errorf("failed to create test database: %v", err)
	}
	defer bc.Release()

	balance, _ := new(big.Int).SetString("1000000000000000000000", 10)
	send, _ := new(big.Int).SetString("500000000000000000000", 10)

	err = bc.ConnectBlock(
		NewLuaTxAccountBig("user1", balance),
		NewLuaTxAccount("user1", 0, types.Aer),
		NewLuaTxDeploy("user1", "fd", 0, definition),
		NewLuaTxSendBig("user1", "fd", send),
	)

	err = bc.ConnectBlock(
		NewLuaTxCall("user1", "fd", 0, `{"Name": "query_no", "Args":[]}`).
			Fail("not enough balance"),
	)
	if err != nil {
		t.Error(err)
	}
	txs := make([]LuaTxTester, 10000)

	for i:=0; i < 10000; i++ {
		txs[i] =
			NewLuaTxCallFeeDelegate("user1", "fd", 0, `{"Name": "query_no", "Args":[]}`)
	}
	err = bc.ConnectBlock(txs...)
	if err != nil {
		t.Error(err)
	}
}
*/

// Make sure that changes made on one contract Lua VM do not affect other called contracts
func TestContractIsolation(t *testing.T) {
	code := readLuaCode(t, "feature_isolation.lua")

	for version := min_version; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("user1", 1, types.Aergo),
			NewLuaTxDeploy("user1", "A", 0, code),
			NewLuaTxDeploy("user1", "B", 0, code),
		)
		require.NoErrorf(t, err, "failed to connect new block")

		// forward order
		tx := NewLuaTxCall("user1", "A", 0, fmt.Sprintf(`{"Name":"test_vm_isolation_forward", "Args":["%s"]}`, nameToAddress("B")))
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")
		receipt := bc.GetReceipt(tx.Hash())
		require.Equalf(t, ``, receipt.GetRet(), "contract call ret error")

		// reverse order using A -> A
		tx = NewLuaTxCall("user1", "A", 0, fmt.Sprintf(`{"Name":"test_vm_isolation_reverse", "Args":["%s"]}`, nameToAddress("A")))
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")
		receipt = bc.GetReceipt(tx.Hash())
		require.Equalf(t, ``, receipt.GetRet(), "contract call ret error")

		// reverse order using A -> B
		tx = NewLuaTxCall("user1", "A", 0, fmt.Sprintf(`{"Name":"test_vm_isolation_reverse", "Args":["%s"]}`, nameToAddress("B")))
		err = bc.ConnectBlock(tx)
		require.NoErrorf(t, err, "failed to connect new block")
		receipt = bc.GetReceipt(tx.Hash())
		require.Equalf(t, ``, receipt.GetRet(), "contract call ret error")

	}
}

//////////////////////////////////////////////////////////////////////
// COMPOSABLE TRANSACTIONS
//////////////////////////////////////////////////////////////////////

func multicall(t *testing.T, bc *DummyChain, params ...string) {
	t.Helper()

	var expectedError, expectedResult string
	account := params[0]
	payload := params[1]
	if len(params) > 2 {
		expectedError = params[2]
	}
	if len(params) > 3 {
		expectedResult = params[3]
	}

	tx := NewLuaTxMultiCall(account, payload).Fail(expectedError)

	err := bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
		return
	}

	if expectedError == "" && expectedResult != "" {
		receipt := bc.GetReceipt(tx.Hash())
		if receipt.GetRet() != expectedResult {
			t.Errorf("multicall invalid result - expected: %s	got: %s", expectedResult, receipt.GetRet())
		}
	}

}

func call(t *testing.T, bc *DummyChain,
          account string, amount uint64,
          contract string, function string, args string,
          expectedError string, expectedResult string) {

	t.Helper()

	callinfo := fmt.Sprintf(`{"Name":"%s", "Args":%s}`, function, args)

	tx := NewLuaTxCall(account, contract, amount, callinfo).Fail(expectedError)

	err := bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
		return
	}

	if expectedError == "" && expectedResult != "" {
		receipt := bc.GetReceipt(tx.Hash())
		if receipt.GetRet() != expectedResult {
			t.Errorf("call invalid result - expected: %s  got: %s", expectedResult, receipt.GetRet())
		}
	}

}

func contract_multicall(t *testing.T, bc *DummyChain,
          account string, contract string, function string, script string,
          expectedError string, expectedResult string) {

	t.Helper()

	// escape the script JSON string
	callinfo := fmt.Sprintf(`{"Name":"%s", "Args":[%s]}`, function, strconv.Quote(script))

	tx := NewLuaTxCall(account, contract, 0, callinfo).Fail(expectedError)

	err := bc.ConnectBlock(tx)
	if err != nil {
		t.Error(err)
		return
	}

	if expectedError == "" && expectedResult != "" {
		receipt := bc.GetReceipt(tx.Hash())
		if receipt.GetRet() != expectedResult {
			t.Errorf("call invalid result - expected: %s  got: %s", expectedResult, receipt.GetRet())
		}
	}

}

func build_call_tx(account string, contract string, function string, args string) *luaTxCall {
	callinfo := fmt.Sprintf(`{"Name":"%s", "Args":%s}`, function, args)
	return NewLuaTxCall(account, contract, 0, callinfo)
}

func build_contract_multicall_tx(account string, contract string, function string, script string) *luaTxCall {
	callinfo := fmt.Sprintf(`{"Name":"%s", "Args":[%s]}`, function, strconv.Quote(script))
	return NewLuaTxCall(account, contract, 0, callinfo)
}

func execute_block(t *testing.T, bc *DummyChain, txns []*luaTxCall, expectedResults []string) {
	t.Helper()

	// Convert []*luaTxCall to []LuaTxTester
	luaTxTesters := make([]LuaTxTester, len(txns))
	for i, tx := range txns {
		luaTxTesters[i] = LuaTxTester(tx)
	}

	err := bc.ConnectBlock(luaTxTesters...)
	if err != nil {
		t.Error(err)
		return
	}

	for i, tx := range txns {
		receipt := bc.GetReceipt(tx.Hash())
		if receipt.GetRet() != expectedResults[i] {
			t.Errorf("call invalid result - expected: %s  got: %s", expectedResults[i], receipt.GetRet())
		}
	}

}

func TestComposableTransactions(t *testing.T) {
	code := readLuaCode(t, "feature_multicall.lua")

	for version := min_version_multicall; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("ac0", 10, types.Aergo),
			NewLuaTxAccount("ac1", 10, types.Aergo),
			NewLuaTxAccount("ac2", 10, types.Aergo),
			NewLuaTxAccount("ac3", 10, types.Aergo),
			NewLuaTxAccount("ac4", 10, types.Aergo),
			NewLuaTxAccount("ac5", 10, types.Aergo),
			NewLuaTxDeploy("ac0", "tables", 0, code),
			NewLuaTxDeploy("ac0", "c1", 0, code),
			NewLuaTxDeploy("ac0", "c2", 0, code),
			NewLuaTxDeploy("ac0", "c3", 0, code),
		)
		if err != nil {
			t.Error(err)
		}


		multicall(t, bc, "ac1", `[
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","get_dict"],
		 ["store result as","dict"],
		 ["set","%dict%","two",22],
		 ["set","%dict%","four",4],
		 ["set","%dict%","one",null],
		 ["get","%dict%","two"],
		 ["set","%dict%","copy","%last result%"],
		 ["return","%dict%"]
		]`, ``, `{"copy":22,"four":4,"three":3,"two":22}`)

		multicall(t, bc, "ac1", `[
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","get_list"],
		 ["store result as","array"],
		 ["set","%array%",2,"2nd"],
		 ["insert","%array%",1,"zero"],
		 ["insert","%array%","last"],
		 ["return","%array%"]
		]`, ``, `["zero","first","2nd","third",123,12.5,true,"last"]`)

		multicall(t, bc, "ac1", `[
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","get_list"],
		 ["store result as","array"],
		 ["remove","%array%",3],
		 ["return","%array%","%last result%"]
		]`, ``, `[["first","second",123,12.5,true],"third"]`)


		// create new dict or array using fromjson

		multicall(t, bc, "ac1", `[
		 ["from json","{\"one\":1,\"two\":2}"],
		 ["set","%last result%","three",3],
		 ["return","%last result%"]
		]`, ``, `{"one":1,"three":3,"two":2}`)


		// define dict or list using let

		multicall(t, bc, "ac1", `[
		 ["let","obj",{"one":1,"two":2}],
		 ["set","%obj%","three",3],
		 ["return","%obj%"]
		]`, ``, `{"one":1,"three":3,"two":2}`)

		multicall(t, bc, "ac1", `[
		 ["let","list",["one",1,"two",2,2.5,true,false]],
		 ["set","%list%",4,"three"],
		 ["insert","%list%",1,"first"],
		 ["insert","%list%","last"],
		 ["return","%list%"]
		]`, ``, `["first","one",1,"two","three",2.5,true,false,"last"]`)

		multicall(t, bc, "ac1", `[
		 ["let","list",["one",22,3.3,true,false]],
		 ["get","%list%",1],
		 ["assert","%last result%","=","one"],
		 ["get","%list%",2],
		 ["assert","%last result%","=",22],
		 ["get","%list%",3],
		 ["assert","%last result%","=",3.3],
		 ["get","%list%",4],
		 ["assert","%last result%","=",true],
		 ["get","%list%",5],
		 ["assert","%last result%","=",false],
		 ["return","%list%"]
		]`, ``, `["one",22,3.3,true,false]`)


		// get size

		multicall(t, bc, "ac1", `[
		 ["let","str","this is a string"],
		 ["get size","%str%"],
		 ["return","%last result%"]
		]`, ``, `16`)

		multicall(t, bc, "ac1", `[
		 ["let","list",["one",1,"two",2,2.5,true,false]],
		 ["get size","%list%"],
		 ["return","%last result%"]
		]`, ``, `7`)

		multicall(t, bc, "ac1", `[
		 ["let","obj",{"one":1,"two":2,"three":3}],
		 ["get size","%obj%"],
		 ["return","%last result%"]
		]`, ``, `0`)




		// BIGNUM

		multicall(t, bc, "ac1", `[
		 ["to big number",123],
		 ["store result as","a"],
		 ["to big number",123],
		 ["store result as","b"],
		 ["multiply","%a%","%b%"],
		 ["return","%last result%"]
		]`, ``, `{"_bignum":"15129"}`)

		multicall(t, bc, "ac1", `[
		 ["to big number","500000000000000000000"],
		 ["store result as","a"],
		 ["to big number","100000"],
		 ["store result as","b"],
		 ["divide","%a%","%b%"],
		 ["return","%last result%"]
		]`, ``, `{"_bignum":"5000000000000000"}`)

		multicall(t, bc, "ac1", `[
		 ["to big number","500000000000000000000"],
		 ["store result as","a"],
		 ["to big number","100000"],
		 ["store result as","b"],
		 ["divide","%a%","%b%"],
		 ["to string","%last result%"],
		 ["return","%last result%"]
		]`, ``, `"5000000000000000"`)

		multicall(t, bc, "ac1", `[
		 ["to big number","500000000000000000000"],
		 ["store result as","a"],

		 ["to big number","100000"],
		 ["divide","%a%","%last result%"],
		 ["store result as","a"],

		 ["to big number","1000000000000000"],
		 ["subtract","%a%","%last result%"],
		 ["store result as","a"],

		 ["to big number","1234"],
		 ["add","%a%","%last result%"],
		 ["store result as","a"],

		 ["to big number","2"],
		 ["remainder","%a%","10000"],

		 ["return","%last result%"]
		]`, ``, `{"_bignum":"1234"}`)



		// STRINGS

		multicall(t, bc, "ac1", `[
		 ["format","%s%s%s","hello"," ","world"],
		 ["return","%last result%"]
		]`, ``, `"hello world"`)

		multicall(t, bc, "ac1", `[
		 ["let","s","hello world"],
		 ["extract","%s%",1,4],
		 ["return","%last result%"]
		]`, ``, `"hell"`)

		multicall(t, bc, "ac1", `[
		 ["let","s","hello world"],
		 ["extract","%s%",-2,-1],
		 ["return","%last result%"]
		]`, ``, `"ld"`)

		multicall(t, bc, "ac1", `[
		 ["let","s","the amount is 12345"],
		 ["find","%s%","%d+"],
		 ["to number"],
		 ["return","%last result%"]
		]`, ``, `12345`)

		multicall(t, bc, "ac1", `[
		 ["let","s","rate: 55 10%"],
		 ["find","%s%","(%d+)%%"],
		 ["to number"],
		 ["return","%last result%"]
		]`, ``, `10`)

		multicall(t, bc, "ac1", `[
		 ["let","s","rate: 12%"],
		 ["find","%s%","%s*(%d+)%%"],
		 ["to number"],
		 ["return","%last result%"]
		]`, ``, `12`)

		multicall(t, bc, "ac1", `[
		 ["let","s","hello world"],
		 ["replace","%s%","hello","good bye"],
		 ["return","%last result%"]
		]`, ``, `"good bye world"`)

		multicall(t, bc, "ac1", `[
		 ["from json","{\"name\":\"ticket\",\"value\":12.5,\"amount\":10}"],
		 ["replace","name = $name, value = $value, amount = $amount","%$(%w+)","%last result%"],
		 ["return","%last result%"]
		]`, ``, `"name = ticket, value = 12.5, amount = 10"`)


		// IF THEN ELSE

		multicall(t, bc, "ac1", `[
		 ["let","s",20],
		 ["if","%s%",">=",20],
		 ["let","b","big"],
		 ["else if","%s%",">=",10],
		 ["let","b","medium"],
		 ["else"],
		 ["let","b","low"],
		 ["end if"],
		 ["let","c","after"],
		 ["return","%b%","%c%"]
		]`, ``, `["big","after"]`)

		multicall(t, bc, "ac1", `[
		 ["let","s",10],
		 ["if","%s%",">=",20],
		 ["let","b","big"],
		 ["else if","%s%",">=",10],
		 ["let","b","medium"],
		 ["else"],
		 ["let","b","low"],
		 ["end if"],
		 ["let","c","after"],
		 ["return","%b%","%c%"]
		]`, ``, `["medium","after"]`)

		multicall(t, bc, "ac1", `[
		 ["let","s",5],
		 ["if","%s%",">=",20],
		 ["let","b","big"],
		 ["else if","%s%",">=",10],
		 ["let","b","medium"],
		 ["else"],
		 ["let","b","low"],
		 ["end if"],
		 ["let","c","after"],
		 ["return","%b%","%c%"]
		]`, ``, `["low","after"]`)

		multicall(t, bc, "ac1", `[
		 ["let","s",20],
		 ["if","%s%",">=",20],
		 ["return","big"],
		 ["else if","%s%",">=",10],
		 ["return","medium"],
		 ["else"],
		 ["return","low"],
		 ["end if"],
		 ["return","after"]
		]`, ``, `"big"`)

		multicall(t, bc, "ac1", `[
		 ["let","s",10],
		 ["if","%s%",">=",20],
		 ["return","big"],
		 ["else if","%s%",">=",10],
		 ["return","medium"],
		 ["else"],
		 ["return","low"],
		 ["end if"],
		 ["return","after"]
		]`, ``, `"medium"`)

		multicall(t, bc, "ac1", `[
		 ["let","s",5],
		 ["if","%s%",">=",20],
		 ["return","big"],
		 ["else if","%s%",">=",10],
		 ["return","medium"],
		 ["else"],
		 ["return","low"],
		 ["end if"],
		 ["return","after"]
		]`, ``, `"low"`)


		multicall(t, bc, "ac1", `[
		 ["to big number","500000000000000000000"],
		 ["store result as","a"],
		 ["to big number","500000000000000000000"],
		 ["store result as","b"],
		 ["if","%a%","=","%b%"],
		 ["let","b","equal"],
		 ["else"],
		 ["let","b","diff"],
		 ["end if"],
		 ["return","%b%"]
		]`, ``, `"equal"`)

		multicall(t, bc, "ac1", `[
		 ["to big number","500000000000000000000"],
		 ["store result as","a"],
		 ["to big number","500000000000000000001"],
		 ["store result as","b"],
		 ["if","%a%","=","%b%"],
		 ["let","b","equal"],
		 ["else"],
		 ["let","b","diff"],
		 ["end if"],
		 ["return","%b%"]
		]`, ``, `"diff"`)

		multicall(t, bc, "ac1", `[
		 ["to big number","500000000000000000001"],
		 ["store result as","a"],
		 ["to big number","500000000000000000000"],
		 ["store result as","b"],
		 ["if","%a%",">","%b%"],
		 ["let","b","bigger"],
		 ["else"],
		 ["let","b","lower"],
		 ["end if"],
		 ["return","%b%"]
		]`, ``, `"bigger"`)

		multicall(t, bc, "ac1", `[
		 ["to big number","500000000000000000000"],
		 ["store result as","a"],
		 ["to big number","500000000000000000001"],
		 ["store result as","b"],
		 ["if","%a%",">","%b%"],
		 ["let","b","bigger"],
		 ["else"],
		 ["let","b","lower"],
		 ["end if"],
		 ["return","%b%"]
		]`, ``, `"lower"`)


		multicall(t, bc, "ac1", `[
		 ["to big number","500000000000000000000"],
		 ["store result as","a"],
		 ["to big number","500000000000000000001"],
		 ["store result as","b"],
		 ["if","%a%","<","%b%","and","1","=","0"],
		 ["let","b","wrong 1"],
		 ["else if","%a%","<","%b%","and","1","=","1"],
		 ["let","b","correct"],
		 ["else"],
		 ["let","b","wrong 2"],
		 ["end if"],
		 ["return","%b%"]
		]`, ``, `"correct"`)

		multicall(t, bc, "ac1", `[
		 ["to big number","500000000000000000000"],
		 ["store result as","a"],
		 ["to big number","500000000000000000001"],
		 ["store result as","b"],
		 ["if","%a%","<","%b%","and",1,"=",0],
		 ["let","b","wrong 1"],
		 ["else if","%a%","<","%b%","and",1,"=",1],
		 ["let","b","correct"],
		 ["else"],
		 ["let","b","wrong 2"],
		 ["end if"],
		 ["return","%b%"]
		]`, ``, `"correct"`)

		multicall(t, bc, "ac1", `[
		 ["to big number","500000000000000000000"],
		 ["store result as","a"],
		 ["to big number","500000000000000000001"],
		 ["store result as","b"],
		 ["to big number","400000000000000000000"],
		 ["store result as","c"],
		 ["if","%a%","<","%b%","and","%a%","<","%c%"],
		 ["let","b","wrong 1"],
		 ["else if","%a%",">","%b%","and","%a%",">","%c%"],
		 ["let","b","wrong 2"],
		 ["else if","%a%","<","%b%","and","%a%",">","%c%"],
		 ["let","b","correct"],
		 ["else"],
		 ["let","b","wrong 3"],
		 ["end if"],
		 ["return","%b%"]
		]`, ``, `"correct"`)

		multicall(t, bc, "ac1", `[
		 ["to big number","500000000000000000000"],
		 ["store result as","a"],
		 ["to big number","500000000000000000001"],
		 ["store result as","b"],
		 ["to big number","400000000000000000000"],
		 ["store result as","c"],
		 ["to string",0],

		 ["if","%a%",">","%b%","or","%a%","<","%c%"],
		 ["format","%s%s","%last result%","1"],
		 ["end if"],

		 ["if","%a%","=","%b%","or","%a%","=","%c%"],
		 ["format","%s%s","%last result%","2"],
		 ["end if"],

		 ["if","%a%","<","%b%","or","%a%","<","%c%"],
		 ["format","%s%s","%last result%","3"],
		 ["end if"],

		 ["if","%a%",">","%b%","or","%a%",">","%c%"],
		 ["format","%s%s","%last result%","4"],
		 ["end if"],

		 ["if","%a%","<","%b%","or","%a%",">","%c%"],
		 ["format","%s%s","%last result%","5"],
		 ["end if"],

		 ["if","%a%","!=","%b%","or","%a%","=","%c%"],
		 ["format","%s%s","%last result%","6"],
		 ["end if"],

		 ["if","%a%","=","%b%","or","%a%","!=","%c%"],
		 ["format","%s%s","%last result%","7"],
		 ["end if"],


		 ["if","%a%",">=","%b%","and","%a%","<=","%c%"],
		 ["format","%s%s","%last result%","8"],
		 ["end if"],

		 ["if","%a%",">=","%c%","and","%a%","<=","%b%"],
		 ["format","%s%s","%last result%","9"],
		 ["end if"],

		 ["if","%b%",">=","%a%","and","%b%","<=","%c%"],
		 ["format","%s%s","%last result%","A"],
		 ["end if"],

		 ["if","%b%",">=","%c%","and","%b%","<=","%a%"],
		 ["format","%s%s","%last result%","B"],
		 ["end if"],

		 ["if","%c%",">=","%a%","and","%c%","<=","%b%"],
		 ["format","%s%s","%last result%","C"],
		 ["end if"],

		 ["if","%c%",">=","%b%","and","%c%","<=","%a%"],
		 ["format","%s%s","%last result%","D"],
		 ["end if"],


		 ["if","%a%",">=","%b%","and","%a%","<=","%c%","or",1,"=",0],
		 ["format","%s%s","%last result%","E"],
		 ["end if"],

		 ["if","%a%",">=","%b%","and","%a%","<=","%c%","or",1,"=",1],
		 ["format","%s%s","%last result%","F"],
		 ["end if"],

		 ["if","%a%",">=","%b%","and","%a%","<=","%c%","and",1,"=",0],
		 ["format","%s%s","%last result%","G"],
		 ["end if"],

		 ["if","%a%",">=","%b%","and","%a%","<=","%c%","and",1,"=",1],
		 ["format","%s%s","%last result%","H"],
		 ["end if"],


		 ["if","%a%",">=","%c%","and","%a%","<=","%b%","or",1,"=",0],
		 ["format","%s%s","%last result%","I"],
		 ["end if"],

		 ["if","%a%",">=","%c%","and","%a%","<=","%b%","or",1,"=",1],
		 ["format","%s%s","%last result%","J"],
		 ["end if"],

		 ["if","%a%",">=","%c%","and","%a%","<=","%b%","and",1,"=",0],
		 ["format","%s%s","%last result%","K"],
		 ["end if"],

		 ["if","%a%",">=","%c%","and","%a%","<=","%b%","and",1,"=",1],
		 ["format","%s%s","%last result%","L"],
		 ["end if"],


		 ["if",1,"=",0,"or","%a%",">=","%b%","and","%a%","<=","%c%"],
		 ["format","%s%s","%last result%","M"],
		 ["end if"],

		 ["if",1,"=",1,"or","%a%",">=","%b%","and","%a%","<=","%c%"],
		 ["format","%s%s","%last result%","N"],
		 ["end if"],

		 ["if",1,"=",0,"or","%a%",">=","%c%","and","%a%","<=","%b%"],
		 ["format","%s%s","%last result%","O"],
		 ["end if"],

		 ["if",1,"=",1,"or","%a%",">=","%c%","and","%a%","<=","%b%"],
		 ["format","%s%s","%last result%","P"],
		 ["end if"],


		 ["return","%last result%"]
		]`, ``, `"0345679IJLNOP"`)




		// FOR

		multicall(t, bc, "ac1", `[
		 ["for","n",1,5],
		 ["loop"],
		 ["return","%n%"]
		]`, ``, `6`)

		multicall(t, bc, "ac1", `[
		 ["to number","0"],
		 ["for","n",1,5],
		 ["add","%last result%",1],
		 ["loop"],
		 ["return","%last result%"]
		]`, ``, `5`)

		multicall(t, bc, "ac1", `[
		 ["to big number","10000000000000000001"],
		 ["store result as","to_add"],
		 ["to big number","100000000000000000000"],

		 ["for","n",1,3],
		 ["add","%last result%","%to_add%"],
		 ["loop"],

		 ["to string","%last result%"],
		 ["return","%last result%"]
		]`, ``, `"130000000000000000003"`)


		multicall(t, bc, "ac1", `[
		 ["to number","0"],
		 ["for","n",500,10,-5],
		 ["add","%last result%",1],
		 ["loop"],
		 ["return","%last result%"]
		]`, ``, `99`)


		multicall(t, bc, "ac1", `[
		 ["to number","0"],
		 ["for","n",5,1,-1],
		 ["add","%last result%",1],
		 ["loop"],
		 ["return","%last result%"]
		]`, ``, `5`)

		multicall(t, bc, "ac1", `[
		 ["to number","0"],
		 ["for","n",5,1],
		 ["add","%last result%",1],
		 ["loop"],
		 ["return","%last result%"]
		]`, ``, `0`)

		multicall(t, bc, "ac1", `[
		 ["to number","0"],
		 ["for","n",1,5],
		 ["add","%last result%",1],
		 ["loop"],
		 ["return","%last result%"]
		]`, ``, `5`)

		multicall(t, bc, "ac1", `[
		 ["to number","0"],
		 ["for","n",1,5,-1],
		 ["add","%last result%",1],
		 ["loop"],
		 ["return","%last result%"]
		]`, ``, `0`)



		// FOREACH

		multicall(t, bc, "ac1", `[
		 ["let","list",[11,22,33]],
		 ["let","r",0],
		 ["for each","item","in","%list%"],
		 ["add","%r%","%item%"],
		 ["store result as","r"],
		 ["loop"],
		 ["return","%r%"]
		]`, ``, `66`)

		multicall(t, bc, "ac1", `[
		 ["let","list",[11,22,33]],
		 ["let","counter",0],
		 ["for each","item","in","%list%"],
		 ["add","%counter%",1],
		 ["store result as","counter"],
		 ["loop"],
		 ["return","%counter%"]
		]`, ``, `3`)

		multicall(t, bc, "ac1", `[
		 ["let","list",[]],
		 ["let","counter",0],
		 ["for each","item","in","%list%"],
		 ["add","%counter%",1],
		 ["store result as","counter"],
		 ["loop"],
		 ["return","%counter%"]
		]`, ``, `0`)

		multicall(t, bc, "ac1", `[
		 ["let","list",["one",1,"two",2,2.5,true,false]],
		 ["let","counter",0],
		 ["for each","item","in","%list%"],
		 ["add","%counter%",1],
		 ["store result as","counter"],
		 ["loop"],
		 ["return","%counter%"]
		]`, ``, `7`)

		multicall(t, bc, "ac1", `[
		 ["let","list",[10,21,32]],
		 ["let","r",0],
		 ["for each","item","in","%list%"],
		 ["if","%item%","<",30],
		 ["add","%r%","%item%"],
		 ["store result as","r"],
		 ["end if"],
		 ["loop"],
		 ["return","%r%"]
		]`, ``, `31`)



		// FORPAIR

		multicall(t, bc, "ac1", `[
		 ["let","str",""],
		 ["let","sum",0],
		 ["let","obj",{"one":1,"two":2,"three":3}],
		 ["for each","key","value","in","%obj%"],
		 ["combine","%str%","%key%"],
		 ["store result as","str"],
		 ["add","%sum%","%value%"],
		 ["store result as","sum"],
		 ["loop"],
		 ["return","%str%","%sum%"]
		]`, ``, `["onethreetwo",6]`)

		multicall(t, bc, "ac1", `[
		 ["let","str",""],
		 ["let","sum",0],
		 ["let","obj",{"one":1.5,"two":2.5,"three":3.5,"four":4.5}],
		 ["for each","key","value","in","%obj%"],
		 ["combine","%str%","%key%"],
		 ["store result as","str"],
		 ["add","%sum%","%value%"],
		 ["store result as","sum"],
		 ["loop"],
		 ["return","%str%","%sum%"]
		]`, ``, `["fouronethreetwo",12]`)

		multicall(t, bc, "ac1", `[
		 ["let","names",[]],
		 ["let","values",[]],
		 ["let","obj",{"one":1.5,"two":2.5,"three":3.5,"four":4.5}],
		 ["for each","key","value","in","%obj%"],
		 ["insert","%names%","%key%"],
		 ["insert","%values%","%value%"],
		 ["loop"],
		 ["return","%names%","%values%"]
		]`, ``, `[["four","one","three","two"],[4.5,1.5,3.5,2.5]]`)

		multicall(t, bc, "ac1", `[
		 ["let","names",[]],
		 ["let","values",[]],
		 ["let","obj",{"one":1.5,"two":2.5,"three":3.5,"four":4.5}],
		 ["for each","key","value","in","%obj%"],
		 ["insert","%names%","%key%"],
		 ["insert","%values%","%value%"],
		 ["loop"],
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","sort","%values%"],
		 ["store result as","values"],
		 ["return","%names%","%values%"]
		]`, ``, `[["four","one","three","two"],[1.5,2.5,3.5,4.5]]`)

		multicall(t, bc, "ac1", `[
		 ["let","obj",{}],
		 ["let","counter",0],
		 ["for each","key","value","in","%obj%"],
		 ["add","%counter%",1],
		 ["store result as","counter"],
		 ["loop"],
		 ["return","%counter%"]
		]`, ``, `0`)



		// FOR "BREAK"

		multicall(t, bc, "ac1", `[
		 ["let","c",0],
		 ["for","n",1,10],
		 ["add","%c%",1],
		 ["store result as","c"],
		 ["if","%n%","=",5],
		 ["let","n",500],
		 ["end if"],
		 ["loop"],
		 ["return","%c%"]
		]`, ``, `5`)

		multicall(t, bc, "ac1", `[
		 ["to number","0"],
		 ["for","n",500,10,-5],
		 ["add","%last result%",1],
		 ["if","%n%","=",475],
		 ["let","n",2],
		 ["end if"],
		 ["loop"],
		 ["return","%last result%"]
		]`, ``, `6`)

		multicall(t, bc, "ac1", `[
		 ["let","c",0],
		 ["for","n",1,10],
		 ["add","%c%",1],
		 ["store result as","c"],
		 ["if","%n%","=",5],
		 ["break"],
		 ["end if"],
		 ["loop"],
		 ["return","%c%"]
		]`, ``, `5`)

		multicall(t, bc, "ac1", `[
		 ["let","c",0],
		 ["for","n",1,10],
		 ["add","%c%",1],
		 ["store result as","c"],
		 ["break","if","%n%","=",5],
		 ["loop"],
		 ["return","%c%"]
		]`, ``, `5`)

		multicall(t, bc, "ac1", `[
		 ["to number","0"],
		 ["for","n",500,10,-5],
		 ["add","%last result%",1],
		 ["if","%n%","=",475],
		 ["break"],
		 ["end if"],
		 ["loop"],
		 ["return","%last result%"]
		]`, ``, `6`)

		multicall(t, bc, "ac1", `[
		 ["to number","0"],
		 ["for","n",500,10,-5],
		 ["add","%last result%",1],
		 ["break","if","%n%","=",475],
		 ["loop"],
		 ["return","%last result%"]
		]`, ``, `6`)

		multicall(t, bc, "ac1", `[
		 ["for","n",1,5],
		 ["loop"],
		 ["return","%n%"]
		]`, ``, `6`)

		multicall(t, bc, "ac1", `[
		 ["for","n",1,5],
		 ["break"],
		 ["loop"],
		 ["return","%n%"]
		]`, ``, `1`)

		multicall(t, bc, "ac1", `[
		 ["let","names",[]],
		 ["let","list",["one","two","three","four"]],
		 ["for each","item","in","%list%"],
		 ["if","%item%","=","three"],
		 ["break"],
		 ["end if"],
		 ["insert","%names%","%item%"],
		 ["loop"],
		 ["return","%names%"]
		]`, ``, `["one","two"]`)

		multicall(t, bc, "ac1", `[
		 ["let","names",[]],
		 ["let","list",["one","two","three","four"]],
		 ["for each","item","in","%list%"],
		 ["break","if","%item%","=","three"],
		 ["insert","%names%","%item%"],
		 ["loop"],
		 ["return","%names%"]
		]`, ``, `["one","two"]`)

		multicall(t, bc, "ac1", `[
		 ["let","names",[]],
		 ["let","obj",{"one":true,"two":false,"three":false,"four":true}],
		 ["for each","key","value","in","%obj%"],
		 ["if","%value%","=",false],
		 ["break"],
		 ["end if"],
		 ["insert","%names%","%key%"],
		 ["loop"],
		 ["return","%names%"]
		]`, ``, `["four","one"]`)

		multicall(t, bc, "ac1", `[
		 ["let","names",[]],
		 ["let","obj",{"one":true,"two":false,"three":false,"four":true}],
		 ["for each","key","value","in","%obj%"],
		 ["break","if","%value%","=",false],
		 ["insert","%names%","%key%"],
		 ["loop"],
		 ["return","%names%"]
		]`, ``, `["four","one"]`)



		// RETURN before the end

		multicall(t, bc, "ac1", `[
		 ["let","v",123],
		 ["if","%v%",">",100],
		 ["return"],
		 ["end if"],
		 ["let","v",500],
		 ["return","%v%"]
		]`, ``, ``)

		multicall(t, bc, "ac1", `[
		 ["let","v",123],
		 ["if","%v%",">",200],
		 ["return"],
		 ["end if"],
		 ["let","v",500],
		 ["return","%v%"]
		]`, ``, `500`)



		// FULL LOOPS

		multicall(t, bc, "ac1", `[
		 ["let","c","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA"],
		 ["call","%c%","inc","n"],
		 ["call","%c%","get","n"],
		 ["if","%last result%",">=",5],
		 ["return","%last result%"],
		 ["end if"],
		 ["loop"]
		]`, ``, `5`)




		// CALLS

		multicall(t, bc, "ac1", `[
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","works"],
		 ["assert","%last result%","=",123],
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","works"],
		 ["return","%last result%"]
		]`, ``, `123`)

		multicall(t, bc, "ac1", `[
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","works"],
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","fails"]
		]`, `this call should fail`)


		multicall(t, bc, "ac3", `[
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","set_name","test"],
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","get_name"],
		 ["assert","%last result%","=","test"]
		]`)

		multicall(t, bc, "ac3", `[
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","set_name","wrong"],
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","get_name"],
		 ["assert","%last result%","=","wrong"],
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","set_name",123]
		]`, `must be string`)

		multicall(t, bc, "ac3", `[
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","get_name"],
		 ["assert","%last result%","=","test"],
		 ["return","%last result%"]
		]`, ``, `"test"`)


		multicall(t, bc, "ac3", `[
		 ["let","c","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA"],
		 ["call","%c%","set_name","test2"],
		 ["call","%c%","get_name"],
		 ["assert","%last result%","=","test2"]
		]`)

		multicall(t, bc, "ac3", `[
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","sender"],
		 ["assert","%last result%","=","%my account address%"],
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","is_contract","%my account address%"],
		 ["assert","%last result%","=",false],
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","set","account","%my account address%"],
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","get","account"],
		 ["assert","%last result%","=","%my account address%"]
		]`)

		// CALL + SEND

		multicall(t, bc, "ac3", `[
		 ["get aergo balance","%my account address%"],
		 ["store result as","my balance before"],
		 ["get aergo balance","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA"],
		 ["store result as","contract balance before"],
		 ["call + send","0.25 aergo","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","resend_to","%my account address%"],
		 ["assert","%last result%","=","250000000000000000"],
		 ["get aergo balance","%my account address%"],
		 ["assert","%last result%","=","%my balance before%"],
		 ["get aergo balance","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA"],
		 ["assert","%last result%","=","%contract balance before%"]
		]`)

		multicall(t, bc, "ac3", `[
		 ["get aergo balance","%my account address%"],
		 ["store result as","my balance before"],
		 ["get aergo balance","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA"],
		 ["store result as","contract balance before"],

		 ["let","amount","1.5","aergo"],
		 ["call + send","%amount%","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","recv_aergo"],

		 ["assert","%my aergo balance%","<","%my balance before%"],
		 ["get aergo balance","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA"],
		 ["assert","%last result%",">","%contract balance before%"],

		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","send_to","%my account address%","%amount%"],

		 ["assert","%my aergo balance%","=","%my balance before%"],
		 ["get aergo balance","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA"],
		 ["assert","%last result%","=","%contract balance before%"]
		]`)


		// CALL LOOP

		multicall(t, bc, "ac3", `[
		 ["let","list",["first","second","third"]],
		 ["for each","item","in","%list%"],
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","add","%item%"],
		 ["loop"]
		]`)

		multicall(t, bc, "ac1", `[
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","get","1"],
		 ["assert","%last result%","=","first"],
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","get","2"],
		 ["assert","%last result%","=","second"],
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","get","3"],
		 ["assert","%last result%","=","third"]
		]`)

		multicall(t, bc, "ac3", `[
		 ["let","list",["1st","2nd","3rd"]],
		 ["let","n",1],
		 ["for each","item","in","%list%"],
		 ["to string","%n%"],
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","set","%last result%","%item%"],
		 ["add","%n%",1],
		 ["store result as","n"],
		 ["loop"]
		]`)

		multicall(t, bc, "ac1", `[
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","get","1"],
		 ["assert","%last result%","=","1st"],
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","get","2"],
		 ["assert","%last result%","=","2nd"],
		 ["call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","get","3"],
		 ["assert","%last result%","=","3rd"]
		]`)



		// TRY CALL

		multicall(t, bc, "ac1", `[
		 ["try call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","works"],
		 ["assert","%call succeeded%","=",true],
		 ["try call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","fails"],
		 ["assert","%call succeeded%","=",false]
		]`)

		multicall(t, bc, "ac3", `[
		 ["try call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","set_name","1st"],
		 ["assert","%call succeeded%"],

		 ["try call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","get_name"],
		 ["assert","%call succeeded%","=",true],
		 ["assert","%last result%","=","1st"],

		 ["try call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","set_name",22],
		 ["assert","%call succeeded%","=",false],

		 ["try call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","get_name"],
		 ["assert","%call succeeded%","=",true],
		 ["assert","%last result%","=","1st"],

		 ["return","%last result%"]
		]`, ``, `"1st"`)


		// TRY CALL + SEND

		multicall(t, bc, "ac1", `[
		 ["get aergo balance","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA"],
		 ["assert","%last result%","=","0"],
		 ["try call + send","0.25 aergo","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","recv_aergo"],
		 ["assert","%call succeeded%"],
		 ["try call + send","1 aergo","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","recv_aergo"],
		 ["assert","%call succeeded%","=",true],
		 ["get aergo balance","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA"],
		 ["assert","%last result%","=","1.25 aergo"],
		 ["try call + send","1 aergo","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRB","recv_aergo"],
		 ["assert","%call succeeded%","=",false]
		]`)

		// ac1: AmgMPiyZYr19kQ1kHFNiGenez1CRTBqNWqppj6gGZGEP6qszDGe1
		// ac2: AmgeSw3M3V3orBMjf1j98kGne4WycnmQWVTJe6MYNrQ2wuVz3Li2

		multicall(t, bc, "ac1", `[
		 ["let","balance before","%my aergo balance%"],
		 ["get aergo balance","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA"],
		 ["assert","%last result%","=","1.25 aergo"],

		 ["try call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","send_and_fail","AmgMPiyZYr19kQ1kHFNiGenez1CRTBqNWqppj6gGZGEP6qszDGe1","0.5 aergo"],
		 ["assert","%call succeeded%","=",false],
		 ["get aergo balance","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA"],
		 ["assert","%last result%","=","1.25 aergo"],
		 ["assert","%my aergo balance%","=","%balance before%"],

		 ["try call + send","0.25 aergo","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","resend_and_fail","AmgMPiyZYr19kQ1kHFNiGenez1CRTBqNWqppj6gGZGEP6qszDGe1"],
		 ["assert","%call succeeded%","=",false],
		 ["get aergo balance","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA"],
		 ["assert","%last result%","=","1.25 aergo"],
		 ["assert","%my aergo balance%","=","%balance before%"],

		 ["try call + send","0.25 aergo","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","resend_and_fail","AmgeSw3M3V3orBMjf1j98kGne4WycnmQWVTJe6MYNrQ2wuVz3Li2"],
		 ["assert","%call succeeded%","=",false],
		 ["get aergo balance","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA"],
		 ["assert","%last result%","=","1.25 aergo"],
		 ["assert","%my aergo balance%","=","%balance before%"],

		 ["try call","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA","send_to","AmgMPiyZYr19kQ1kHFNiGenez1CRTBqNWqppj6gGZGEP6qszDGe1","0.5 aergo"],
		 ["assert","%call succeeded%"],
		 ["get aergo balance","AmhbUWkqenFtgKLnbDd1NXHce7hn35pcHWYRWBnq5vauLfEQXXRA"],
		 ["assert","%last result%","=","0.75 aergo"],
		 ["subtract","%my aergo balance%","%balance before%"],
		 ["assert","%last result%","=","0.5 aergo"]
		]`)


		// MULTICALL ON ACCOUNT ------------------------------------------


		//deploy ac0 0 c1 test.lua
		//deploy ac0 0 c2 test.lua
		//deploy ac0 0 c3 test.lua

		// c1: AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9
		// c2: Amh8PekqkDmLiwE6FUX6JejjWk3R54cmTaa1Tc1VHZmTRJMruWe4
		// c3: AmgtL32d1M56xGENKDnDqXFzkrYJwWidzSMtay3F8fFDU1VAEdvK

		multicall(t, bc, "ac0", `[
		 ["call","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9","set_name","testing multicall"],
		 ["call","Amh8PekqkDmLiwE6FUX6JejjWk3R54cmTaa1Tc1VHZmTRJMruWe4","set_name","contract 2"],
		 ["call","AmgtL32d1M56xGENKDnDqXFzkrYJwWidzSMtay3F8fFDU1VAEdvK","set_name","third one"]
		]`)

		multicall(t, bc, "ac0", `[
		 ["call","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9","get_name"],
		 ["assert","%last result%","=","testing multicall"],
		 ["store result as","r1"],
		 ["call","Amh8PekqkDmLiwE6FUX6JejjWk3R54cmTaa1Tc1VHZmTRJMruWe4","get_name"],
		 ["assert","%last result%","=","contract 2"],
		 ["store result as","r2"],
		 ["call","AmgtL32d1M56xGENKDnDqXFzkrYJwWidzSMtay3F8fFDU1VAEdvK","get_name"],
		 ["assert","%last result%","=","third one"],
		 ["store result as","r3"],
		 ["return","%r1%","%r2%","%r3%"]
		]`, ``, `["testing multicall","contract 2","third one"]`)

		multicall(t, bc, "ac0", `[
		 ["from json","{}"],
		 ["store result as","res"],
		 ["call","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9","get_name"],
		 ["set","%res%","r1","%last result%"],
		 ["call","Amh8PekqkDmLiwE6FUX6JejjWk3R54cmTaa1Tc1VHZmTRJMruWe4","get_name"],
		 ["set","%res%","r2","%last result%"],
		 ["call","AmgtL32d1M56xGENKDnDqXFzkrYJwWidzSMtay3F8fFDU1VAEdvK","get_name"],
		 ["set","%res%","r3","%last result%"],
		 ["return","%res%"]
		]`, ``, `{"r1":"testing multicall","r2":"contract 2","r3":"third one"}`)


		multicall(t, bc, "ac0", `[
		 ["call","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9","set_name","wohooooooo"],
		 ["call","Amh8PekqkDmLiwE6FUX6JejjWk3R54cmTaa1Tc1VHZmTRJMruWe4","set_name","it works!"],
		 ["call","AmgtL32d1M56xGENKDnDqXFzkrYJwWidzSMtay3F8fFDU1VAEdvK","set_name","it really works!"]
		]`)

		multicall(t, bc, "ac0", `[
		 ["call","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9","get_name"],
		 ["assert","%last result%","=","wohooooooo"],
		 ["store result as","r1"],
		 ["call","Amh8PekqkDmLiwE6FUX6JejjWk3R54cmTaa1Tc1VHZmTRJMruWe4","get_name"],
		 ["assert","%last result%","=","it works!"],
		 ["store result as","r2"],
		 ["call","AmgtL32d1M56xGENKDnDqXFzkrYJwWidzSMtay3F8fFDU1VAEdvK","get_name"],
		 ["assert","%last result%","=","it really works!"],
		 ["store result as","r3"],
		 ["return","%r1%","%r2%","%r3%"]
		]`, ``, `["wohooooooo","it works!","it really works!"]`)

		multicall(t, bc, "ac0", `[
		 ["from json","{}"],
		 ["store result as","res"],
		 ["call","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9","get_name"],
		 ["set","%res%","r1","%last result%"],
		 ["call","Amh8PekqkDmLiwE6FUX6JejjWk3R54cmTaa1Tc1VHZmTRJMruWe4","get_name"],
		 ["set","%res%","r2","%last result%"],
		 ["call","AmgtL32d1M56xGENKDnDqXFzkrYJwWidzSMtay3F8fFDU1VAEdvK","get_name"],
		 ["set","%res%","r3","%last result%"],
		 ["return","%res%"]
		]`, ``, `{"r1":"wohooooooo","r2":"it works!","r3":"it really works!"}`)



		// aergo BALANCE and SEND

		multicall(t, bc, "ac5", `[
		 ["assert","%my aergo balance%","=","10 aergo"],

		 ["send","AmhC1V24v3MM6EzLaypudkrEP4NyX3D44aU9vmzV9cAxXUjbRTK6","0.5 aergo"],
		 ["assert","%my aergo balance%","=","9.5 aergo"],

		 ["let","amount","1.5","aergo"],
		 ["send","AmhC1V24v3MM6EzLaypudkrEP4NyX3D44aU9vmzV9cAxXUjbRTK6","%amount%"],
		 ["assert","%my aergo balance%","=","8000000000000000000"],

		 ["let","amount","2","aergo"],
		 ["send","AmhC1V24v3MM6EzLaypudkrEP4NyX3D44aU9vmzV9cAxXUjbRTK6","%amount%"],
		 ["assert","%my aergo balance%","=","6 aergo"],

		 ["return","%my aergo balance%"]
		]`, ``, `{"_bignum":"6000000000000000000"}`)

		multicall(t, bc, "ac2", `[
		 ["get aergo balance"],
		 ["assert","%last result%","=","10000000000000000000"],

		 ["get aergo balance","AmgHyfkUt5iuXJKZNTrdthtXWLLJCrKWdJ6H6Yshn6ZR285Wr2Hc"],
		 ["assert","%last result%","=","0"],

		 ["send","AmgHyfkUt5iuXJKZNTrdthtXWLLJCrKWdJ6H6Yshn6ZR285Wr2Hc","3000000000000000000"],

		 ["get aergo balance","AmgHyfkUt5iuXJKZNTrdthtXWLLJCrKWdJ6H6Yshn6ZR285Wr2Hc"],
		 ["assert","%last result%","=","3000000000000000000"],

		 ["get aergo balance"],
		 ["assert","%last result%","=","7000000000000000000"],

		 ["to string"],
		 ["return","%last result%"]
		]`, ``, `"7000000000000000000"`)



		// SECURITY CHECKS

		// it should not be possible to call the code from another account

		// a. from an account (via multicall, using the 'call' command)

		multicall(t, bc, "ac1", `[
		 ["call","AmgeSw3M3V3orBMjf1j98kGne4WycnmQWVTJe6MYNrQ2wuVz3Li2","execute",[["add",11,22],["return","%last result%"]]],
		 ["return","%last result%"]
		]`, `nd contract`)


		// b. from an account (via a call tx)

		call(t, bc, "ac1", 0, "ac1", "execute", `[[["add",11,22],["return","%last result%"]]]`, `nd contract`, ``)

		call(t, bc, "ac1", 0, "ac2", "execute", `[[["add",11,22],["return","%last result%"]]]`, `nd contract`, ``)


		// c. from a contract (calling back)

		multicall(t, bc, "ac1", `[
		 ["call","Amh8PekqkDmLiwE6FUX6JejjWk3R54cmTaa1Tc1VHZmTRJMruWe4","call","AmgMPiyZYr19kQ1kHFNiGenez1CRTBqNWqppj6gGZGEP6qszDGe1","execute",[["add",11,22],["return","%last result%"]]],
		 ["return","%last result%"]
		]`, `nd contract`)


		// d. from a contract (calling another account)

		multicall(t, bc, "ac1", `[
		 ["call","Amh8PekqkDmLiwE6FUX6JejjWk3R54cmTaa1Tc1VHZmTRJMruWe4","call","AmgeSw3M3V3orBMjf1j98kGne4WycnmQWVTJe6MYNrQ2wuVz3Li2","execute",[["add",11,22],["return","%last result%"]]],
		 ["return","%last result%"]
		]`, `nd contract`)


		// e. from a contract (via a call txn)

		call(t, bc, "ac1", 0, "c2", "call", `["AmgMPiyZYr19kQ1kHFNiGenez1CRTBqNWqppj6gGZGEP6qszDGe1","execute",[["add",11,22],["return","%last result%"]]]`, `nd contract`, ``)

		call(t, bc, "ac1", 0, "c2", "call", `["AmgeSw3M3V3orBMjf1j98kGne4WycnmQWVTJe6MYNrQ2wuVz3Li2","execute",[["add",11,22],["return","%last result%"]]]`, `nd contract`, ``)


		// system.isContract() should return false on user accounts

		call(t, bc, "ac1", 0, "c2", "is_contract", `["AmgMPiyZYr19kQ1kHFNiGenez1CRTBqNWqppj6gGZGEP6qszDGe1"]`, ``, `false`)

		call(t, bc, "ac1", 0, "c2", "is_contract", `["AmgeSw3M3V3orBMjf1j98kGne4WycnmQWVTJe6MYNrQ2wuVz3Li2"]`, ``, `false`)

		multicall(t, bc, "ac1", `[
		 ["call","Amh8PekqkDmLiwE6FUX6JejjWk3R54cmTaa1Tc1VHZmTRJMruWe4","is_contract","AmgMPiyZYr19kQ1kHFNiGenez1CRTBqNWqppj6gGZGEP6qszDGe1"],
		 ["assert","%last result%","=",false],
		 ["return","%last result%"]
		]`, ``, `false`)

		multicall(t, bc, "ac1", `[
		 ["call","Amh8PekqkDmLiwE6FUX6JejjWk3R54cmTaa1Tc1VHZmTRJMruWe4","is_contract","AmgeSw3M3V3orBMjf1j98kGne4WycnmQWVTJe6MYNrQ2wuVz3Li2"],
		 ["assert","%last result%","=",false],
		 ["return","%last result%"]
		]`, ``, `false`)


		// on a contract called by multicall, the system.getSender() and system.getOrigin() must be the same

		multicall(t, bc, "ac1", `[
		 ["call","Amh8PekqkDmLiwE6FUX6JejjWk3R54cmTaa1Tc1VHZmTRJMruWe4","sender"],
		 ["store result as","sender"],
		 ["call","Amh8PekqkDmLiwE6FUX6JejjWk3R54cmTaa1Tc1VHZmTRJMruWe4","origin"],
		 ["store result as","origin"],
		 ["assert","%sender%","=","%origin%"],
		 ["return","%sender%"]
		]`, ``, `"AmgMPiyZYr19kQ1kHFNiGenez1CRTBqNWqppj6gGZGEP6qszDGe1"`)

	}
}

func TestContractMulticall(t *testing.T) {
	code1 := readLuaCode(t, "feature_multicall_contract.lua")
	code2 := readLuaCode(t, "feature_multicall.lua")

	for version := min_version_multicall; version <= max_version; version++ {
		bc, err := LoadDummyChain(SetHardForkVersion(version))
		require.NoErrorf(t, err, "failed to create dummy chain")
		defer bc.Release()

		err = bc.ConnectBlock(
			NewLuaTxAccount("ac0", 10, types.Aergo),
			NewLuaTxAccount("ac1", 10, types.Aergo),
			NewLuaTxAccount("ac2", 10, types.Aergo),
			NewLuaTxAccount("ac3", 10, types.Aergo),

			NewLuaTxDeploy("ac0", "caller", 0, code1),
			NewLuaTxDeploy("ac0", "c1", 0, code2),
			NewLuaTxDeploy("ac0", "c2", 0, code2),
			NewLuaTxDeploy("ac0", "c3", 0, code2),
		)
		if err != nil {
			t.Error(err)
		}

		// ac1: AmgMPiyZYr19kQ1kHFNiGenez1CRTBqNWqppj6gGZGEP6qszDGe1
		// ac2: AmgeSw3M3V3orBMjf1j98kGne4WycnmQWVTJe6MYNrQ2wuVz3Li2

		// caller: AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn
		// c1: AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9
		// c2: Amh8PekqkDmLiwE6FUX6JejjWk3R54cmTaa1Tc1VHZmTRJMruWe4
		// c3: AmgtL32d1M56xGENKDnDqXFzkrYJwWidzSMtay3F8fFDU1VAEdvK

		// simple script, no state change
		contract_multicall(t, bc, "ac1", "caller", "multicall", `[
			["add",11,22],
			["return","%last result%"]
		]`, ``, `33`)

		// state change
		call(t, bc, "ac1", 0, "caller", "set_value", `["num",111]`, ``, ``)
		call(t, bc, "ac1", 0, "caller", "get_value", `["num"]`, ``, `111`)

		// check balance
		state, err := bc.GetAccountState("caller")
		assert.Equalf(t, uint64(0), state.GetBalanceBigInt().Uint64(), "balance error")

		contract_multicall(t, bc, "ac1", "caller", "multicall", `[
			["return","%my aergo balance%"]
		]`, ``, `{"_bignum":"0"}`)

		// transfer aergo to the 'caller' contract
		call(t, bc, "ac1", types.NewAmount(1,types.Aergo).Uint64(), "caller", "recv_aergo", `[]`, ``, ``)

		// use the multicall contract in a delegated call
		contract_multicall(t, bc, "ac1", "caller", "multicall_and_check", `[
			["assert","%my aergo balance%","=","1 aergo"],
			["call + send","0.125 aergo","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9","recv_aergo"],
			["assert","%my aergo balance%","=","0.875 aergo"],
			["get aergo balance","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9"],
			["assert","%last result%","=","0.125 aergo"],
			["return","%my aergo balance%","%last result%"]
		]`, ``, `[{"_bignum":"875000000000000000"},{"_bignum":"125000000000000000"}]`)

		// check the state of the 2 contracts

		state, err = bc.GetAccountState("caller")
		assert.Equalf(t, int64(875000000000000000), state.GetBalanceBigInt().Int64(), "balance error")

		state, err = bc.GetAccountState("AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9")
		assert.Equalf(t, int64(125000000000000000), state.GetBalanceBigInt().Int64(), "balance error")

		contract_multicall(t, bc, "ac1", "caller", "multicall", `[
			["assert","%my aergo balance%","=","0.875 aergo"],
			["get aergo balance","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9"],
			["assert","%last result%","=","0.125 aergo"],
			["return","%my aergo balance%","%last result%"]
		]`, ``, `[{"_bignum":"875000000000000000"},{"_bignum":"125000000000000000"}]`)

		// send the aergo back to the 'caller' contract
		call(t, bc, "ac1", 0, "c1", "send_to", `["AmggmgtWPXtsDkC5hkYYx2iYaWfGs8D4ZvZNwxwdm4gxGSDaCqKn","0.125 aergo"]`, ``, ``)

		contract_multicall(t, bc, "ac1", "caller", "multicall", `[
			["assert","%my aergo balance%","=","1 aergo"],
			["get aergo balance","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9"],
			["assert","%last result%","=","0 aergo"],
			["return","%my aergo balance%","%last result%"]
		]`, ``, `[{"_bignum":"1000000000000000000"},{"_bignum":"0"}]`)


		// make sure that the called contract state is not replaced by the multicall contract

		// transfer tokens to other contract and:
		// - check amounts on the same transaction (in the multicall contract)
		// - check amounts on the same transaction (in the caller contract, after returning from the multicall contract)
		// - check amounts on the same block (another transaction)
		// - check amounts on the next block
		// and different ABI on all 3 contracts

		tx1 := build_contract_multicall_tx("ac1", "caller", "multicall_and_check", `[
			["send","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9","0.125 aergo"],
			["assert","%my aergo balance%","=","0.875 aergo"],
			["get aergo balance","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9"],
			["assert","%last result%","=","0.125 aergo"],
			["return","%my aergo balance%","%last result%"]
		]`)
		tx2 := build_call_tx("ac1", "caller", "get_balance", `[]`)
		tx3 := build_call_tx("ac1", "c1", "get_aergo_balance", `[]`)

		execute_block(t, bc, []*luaTxCall{tx1, tx2, tx3}, []string{`[{"_bignum":"875000000000000000"},{"_bignum":"125000000000000000"}]`, `"875000000000000000"`, `"125000000000000000"`})


		// now the same with just delegate call (no multicall)

		call(t, bc, "ac1", 0, "caller", "get_value", `["num"]`, ``, `111`)

		tx1 = build_call_tx("ac1", "caller", "delegate_call", `["AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9","set","num",222]`)
		tx2 = build_call_tx("ac1", "caller", "get_value", `["num"]`)
		tx3 = build_call_tx("ac1", "c1", "get", `["num"]`)

		execute_block(t, bc, []*luaTxCall{tx1, tx2, tx3}, []string{``, `222`, `null`})

		call(t, bc, "ac1", 0, "caller", "get_value", `["num"]`, ``, `222`)


		// test state recovery - normal atomic txn

		tx1 = build_contract_multicall_tx("ac1", "caller", "multicall_and_check", `[
			["assert","%my aergo balance%","=","0.875 aergo"],
			["get aergo balance","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9"],
			["assert","%last result%","=","0.125 aergo"],

			["send","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9","0.125 aergo"],
			["assert","%my aergo balance%","=","0.75 aergo"],
			["get aergo balance","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9"],
			["assert","%last result%","=","0.25 aergo"],

			["call","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9","fails"],

			["return","%my aergo balance%","%last result%"]
		]`)
		tx2 = build_call_tx("ac1", "caller", "get_balance", `[]`)
		tx3 = build_call_tx("ac1", "c1", "get_aergo_balance", `[]`)

		tx1 = tx1.Fail("this call should fail")

		execute_block(t, bc, []*luaTxCall{tx1, tx2, tx3}, []string{`[Contract.LuaDelegateCallContract] call error: [Contract.LuaCallContract] call err: AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9:0: this call should fail`, `"875000000000000000"`, `"125000000000000000"`})


		// test state recovery - pcall

		tx1 = build_contract_multicall_tx("ac1", "caller", "multicall_and_check", `[
			["assert","%my aergo balance%","=","0.875 aergo"],
			["get aergo balance","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9"],
			["assert","%last result%","=","0.125 aergo"],

			["try call","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9","send_and_fail","AmgtL32d1M56xGENKDnDqXFzkrYJwWidzSMtay3F8fFDU1VAEdvK","0.125 aergo"],

			["assert","%my aergo balance%","=","0.875 aergo"],
			["get aergo balance","AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9"],
			["assert","%last result%","=","0.125 aergo"],

			["return","%my aergo balance%","%last result%"]
		]`)
		tx2 = build_call_tx("ac1", "caller", "get_balance", `[]`)
		tx3 = build_call_tx("ac1", "c1", "get_aergo_balance", `[]`)
		tx4 := build_call_tx("ac1", "c3", "get_aergo_balance", `[]`)

		execute_block(t, bc, []*luaTxCall{tx1, tx2, tx3, tx4}, []string{`[{"_bignum":"875000000000000000"},{"_bignum":"125000000000000000"}]`, `"875000000000000000"`, `"125000000000000000"`, `"0"`})

		state, err = bc.GetAccountState("caller")
		assert.Equalf(t, int64(875000000000000000), state.GetBalanceBigInt().Int64(), "balance error")

		state, err = bc.GetAccountState("c1")
		assert.Equalf(t, int64(125000000000000000), state.GetBalanceBigInt().Int64(), "balance error")

		state, err = bc.GetAccountState("c3")
		assert.Equalf(t, int64(0), state.GetBalanceBigInt().Int64(), "balance error")

	}

}

// ----------------------------------------------------------------------------

const (
	DEF_TEST_CONTRACT = "testcontract"
	DEF_TEST_ACCOUNT  = "testaccount"
)

// utility function for tests
func readLuaCode(t *testing.T, file string) (luaCode string) {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if ok != true {
		return ""
	}
	filePath := filepath.Join(filepath.Dir(filename), "test_files", file)
	raw, err := util.ReadContract(filePath)
	require.NoErrorf(t, err, "failed to read "+file)
	require.NotEmpty(t, raw, "failed to read "+file)
	return string(raw)
}

func nameToAddress(name string) (address string) {
	return types.EncodeAddress(contract.StrHash(name))
}
