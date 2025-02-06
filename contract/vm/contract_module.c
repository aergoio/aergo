#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "util.h"
#include "bignum_module.h"
#include "_cgo_export.h"

extern void checkLuaExecContext(lua_State *L);

static const char *contract_str = "contract";
static const char *call_str = "call";
static const char *delegatecall_str = "delegatecall";
static const char *deploy_str = "deploy";
static const char *amount_str = "amount_value";
static const char *fee_str = "fee";

static void set_call_obj(lua_State *L, const char* obj_name) {
	lua_getglobal(L, contract_str);
	lua_getfield(L, -1, obj_name);
}

static void reset_amount_info (lua_State *L) {
	lua_pushnil(L);
	lua_setfield(L, 1, amount_str);
	lua_pushnil(L);
	lua_setfield(L, 1, fee_str);
}

static int set_value(lua_State *L, const char *str) {

	set_call_obj(L, str);
	if (lua_isnil(L, 1)) {
		return 1;
	}

	switch(lua_type(L, 1)) {
	case LUA_TNUMBER: {
		const char *str = lua_tostring(L, 1);
		lua_pushstring(L, str);
		break;
	}
	case LUA_TSTRING:
		lua_pushvalue(L, 1);
		break;
	case LUA_TUSERDATA: {
		char *str = lua_get_bignum_str(L, 1);
		if (str == NULL) {
			luaL_error(L, "not enough memory");
		}
		lua_pushstring(L, str);
		free (str);
		break;
	}
	default:
		luaL_error(L, "invalid input");
	}

	lua_setfield(L, -2, amount_str);

	return 1;
}

static int set_gas(lua_State *L, const char *str) {
	lua_Integer gas;

	set_call_obj(L, str);
	if (lua_isnil(L, 1)) {
		return 1;
	}
	gas = luaL_checkinteger(L, 1);
	if (gas < 0) {
		luaL_error(L, "invalid number");
	}
	lua_pushinteger(L, gas);
	lua_setfield(L, -2, fee_str);

	return 1;
}

static int call_value(lua_State *L) {
	return set_value(L, call_str);
}

static int call_gas(lua_State *L) {
	return set_gas(L, call_str);
}

static int moduleCall(lua_State *L) {
	char *contract;
	char *fname;
	char *json_args;
	struct luaCallContract_return ret;
	lua_Integer gas;
	char *amount;

	checkLuaExecContext(L);

	if (lua_gettop(L) == 2) {
		lua_gasuse(L, 300);
	} else {
		lua_gasuse(L, 2000);
	}

	lua_getfield(L, 1, amount_str);
	if (lua_isnil(L, -1)) {
		amount = NULL;
	} else {
		amount = (char *)luaL_checkstring(L, -1);
	}

	lua_getfield(L, 1, fee_str);
	if (lua_isnil(L, -1)) {
		gas = 0;
	} else {
		gas = luaL_checkinteger(L, -1);
	}

	lua_pop(L, 2);

	// get the contract address
	contract = (char *)luaL_checkstring(L, 2);

	// when called with contract.call.value(amount)(address) - this triggers a call to the default() function
	if (lua_gettop(L) == 2) {
		char *errStr = luaSendAmount(L, contract, amount);
		reset_amount_info(L);
		if (errStr != NULL) {
			strPushAndRelease(L, errStr);
			luaL_throwerror(L);
		}
		return 0;
	}

	// get the function name
	fname = (char *)luaL_checkstring(L, 3);
	// get the arguments
	json_args = lua_util_get_json_from_stack (L, 4, lua_gettop(L), false);
	if (json_args == NULL) {
		reset_amount_info(L);
		luaL_throwerror(L);
	}

	// call the function on the contract
	ret = luaCallContract(L, contract, fname, json_args, amount, gas);
	free(json_args);
	reset_amount_info(L);

	// if it returned an error message, push it to the stack and throw an error
	if (ret.r1 != NULL) {
		strPushAndRelease(L, ret.r1);
		luaL_throwerror(L);
	}

	// disable gas while importing the returned values
	if (lua_usegas(L)) {
		lua_disablegas(L);
	}
	// push the returned values to the stack
	int count = lua_util_json_array_to_lua(L, ret.r0, true);
	free(ret.r0);
	// enable gas again
	if (lua_usegas(L)) {
		lua_enablegas(L);
	}
	// check for invalid result format
	if (count == -1) {
		luaL_setuncatchablerror(L);
		lua_pushstring(L, "internal error: result from call is not a valid JSON array");
		luaL_throwerror(L);
	}
	// return the number of items in the stack
	return count;
}

static int delegate_call_gas(lua_State *L) {
	return set_gas(L, delegatecall_str);
}

static int moduleDelegateCall(lua_State *L) {
	char *contract;
	char *fname;
	char *json_args;
	struct luaDelegateCallContract_return ret;
	lua_Integer gas;

	checkLuaExecContext(L);

	lua_gasuse(L, 2000);

	lua_getfield(L, 1, fee_str);
	if (lua_isnil(L, -1)) {
		gas = 0;
	} else {
		gas = luaL_checkinteger(L, -1);
	}
	lua_pop(L, 1);

	contract = (char *) luaL_checkstring(L, 2);
	fname = (char *) luaL_checkstring(L, 3);
	json_args = lua_util_get_json_from_stack(L, 4, lua_gettop(L), false);
	if (json_args == NULL) {
		reset_amount_info(L);
		luaL_throwerror(L);
	}

	ret = luaDelegateCallContract(L, contract, fname, json_args, gas);
	free(json_args);
	reset_amount_info(L);

	// if it returned an error message, push it to the stack and throw an error
	if (ret.r1 != NULL) {
		strPushAndRelease(L, ret.r1);
		luaL_throwerror(L);
	}

	// disable gas while importing the returned values
	if (lua_usegas(L)) {
		lua_disablegas(L);
	}
	// push the returned values to the stack
	int count = lua_util_json_array_to_lua(L, ret.r0, true);
	free(ret.r0);
	// enable gas again
	if (lua_usegas(L)) {
		lua_enablegas(L);
	}
	// check for invalid result format
	if (count == -1) {
		luaL_setuncatchablerror(L);
		lua_pushstring(L, "internal error: result from call is not a valid JSON array");
		luaL_throwerror(L);
	}
	// return the number of items in the stack
	return count;
}

static int moduleSend(lua_State *L) {
	char *contract;
	char *errStr;
	char *amount;
	bool needfree = false;

	checkLuaExecContext(L);

	lua_gasuse(L, 300);

	contract = (char *) luaL_checkstring(L, 1);
	if (lua_isnil(L, 2)) {
		return 0;
	}

	switch(lua_type(L, 2)) {
	case LUA_TNUMBER:
		amount = (char *) lua_tostring(L, 2);
		break;
	case LUA_TSTRING:
		amount = (char *) lua_tostring(L, 2);
		break;
	case LUA_TUSERDATA:
		amount = lua_get_bignum_str(L, 2);
		if (amount == NULL) {
			luaL_error(L, "not enough memory");
		}
		needfree = true;
		break;
	default:
		luaL_error(L, "invalid input");
	}

	errStr = luaSendAmount(L, contract, amount);

	if (needfree) {
		free(amount);
	}
	if (errStr != NULL) {
		strPushAndRelease(L, errStr);
		luaL_throwerror(L);
	}
	return 0;
}

static int moduleBalance(lua_State *L) {
	char *contract;
	struct luaGetBalance_return balance;
	int nArg;

	checkLuaExecContext(L);

	lua_gasuse(L, 300);

	nArg = lua_gettop(L);
	if (nArg== 0 || lua_isnil(L, 1)) {
		contract = NULL;
	} else {
		contract = (char *) luaL_checkstring(L, 1);
	}
	if (contract != NULL && nArg == 2) {
		const char *modeArg = luaL_checkstring(L, 2);
		int mode = -1;
		if (strcmp(modeArg, "staking") == 0) {
			mode = 0;
		} else if (strcmp(modeArg, "stakingandwhen") == 0) {
			mode = 1;
		}
		if (mode != -1) {
			struct luaGetStaking_return ret;
			const char *errMsg;
			ret = luaGetStaking(L, contract);
			if (ret.r2 != NULL) {
				strPushAndRelease(L, ret.r2);
				luaL_throwerror(L);
			}
			errMsg = lua_set_bignum(L, ret.r0);
			free(ret.r0);
			if (errMsg != NULL) {
				luaL_error(L, errMsg);
			}
			if (mode == 1) {
				lua_pushinteger(L, ret.r1);
				return 2;
			}
			return 1;
		}
	}

	balance = luaGetBalance(L, contract);
	if (balance.r1 != NULL) {
		strPushAndRelease(L, balance.r1);
		luaL_throwerror(L);
	}

	strPushAndRelease(L, balance.r0);
	return 1;
}

static int modulePcall(lua_State *L) {
	struct luaSetRecoveryPoint_return start_seq;
	int argc;
	int ret;

	checkLuaExecContext(L);

	argc = lua_gettop(L) - 1;

	lua_gasuse(L, 300);

	// create a recovery point
	start_seq = luaSetRecoveryPoint(L);
	if (start_seq.r0 < 0) {
		strPushAndRelease(L, start_seq.r1);
		luaL_throwerror(L);
	}

	// call the function
	ret = lua_pcall(L, argc, LUA_MULTRET, 0);
	if (ret != 0) {
		// revert the contract state
		if (start_seq.r0 > 0) {
			char *errStr = luaClearRecovery(L, start_seq.r0, true);
			if (errStr != NULL) {
				strPushAndRelease(L, errStr);
				luaL_throwerror(L);
			}
		}
		// if out of memory, throw error
		if (ret == LUA_ERRMEM) {
			luaL_throwerror(L);
		}
		// add 'success = false' as the first returned value
		lua_pushboolean(L, false);
		lua_insert(L, 1);
		// return the 2 values
		return 2;
	}

	// add 'success = true' as the first returned value
	lua_pushboolean(L, true);
	lua_insert(L, 1);

	// release the recovery point
	if (start_seq.r0 == 1) {
		char *errStr = luaClearRecovery(L, start_seq.r0, false);
		if (errStr != NULL) {
			strPushAndRelease(L, errStr);
			luaL_throwerror(L);
		}
	}

	// return the number of items in the stack
	return lua_gettop(L);
}

static int deploy_value(lua_State *L) {
	return set_value(L, deploy_str);
}

static int moduleDeploy(lua_State *L) {
	char *contract;
	char *fname;
	char *json_args;
	struct luaDeployContract_return ret;
	char *amount;

	checkLuaExecContext(L);

	lua_gasuse(L, 5000);

	// get the amount to transfer to the new contract
	lua_getfield(L, 1, amount_str);
	if (lua_isnil(L, -1)) {
		amount = NULL;
	} else {
		amount = (char *) luaL_checkstring(L, -1);
	}
	lua_pop(L, 1);
	// get the contract source code or the address to an existing contract
	contract = (char *) luaL_checkstring(L, 2);
	// get the deploy arguments to the constructor function
	json_args = lua_util_get_json_from_stack(L, 3, lua_gettop(L), false);
	if (json_args == NULL) {
		reset_amount_info(L);
		luaL_throwerror(L);
	}

	ret = luaDeployContract(L, contract, json_args, amount);
	free(json_args);
	reset_amount_info(L);

	// if it returned an error message, push it to the stack and throw an error
	if (ret.r1 != NULL) {
		strPushAndRelease(L, ret.r1);
		luaL_throwerror(L);
	}

	// disable gas while importing the returned values
	if (lua_usegas(L)) {
		lua_disablegas(L);
	}
	// push the returned values to the stack
	int count = lua_util_json_array_to_lua(L, ret.r0, true);
	free(ret.r0);
	// enable gas again
	if (lua_usegas(L)) {
		lua_enablegas(L);
	}
	// check for invalid result format
	if (count == -1) {
		luaL_setuncatchablerror(L);
		lua_pushstring(L, "internal error: result from call is not a valid JSON array");
		luaL_throwerror(L);
	}
	// return the number of items in the stack
	return count;
}

static int moduleEvent(lua_State *L) {
	char *event_name;
	char *json_args;
	char *errStr;

	checkLuaExecContext(L);

	lua_gasuse(L, 500);

	event_name = (char *) luaL_checkstring(L, 1);
	if (vm_is_hardfork(L, 2)) {
		json_args = lua_util_get_json_array_from_stack(L, 2, lua_gettop(L), true);
	} else {
		json_args = lua_util_get_json_from_stack(L, 2, lua_gettop(L), false);
	}
	if (json_args == NULL) {
		luaL_throwerror(L);
	}

	errStr = luaEvent(L, event_name, json_args);
	free(json_args);
	if (errStr != NULL) {
		strPushAndRelease(L, errStr);
		luaL_throwerror(L);
	}
	return 0;
}

static int governance(lua_State *L, char type) {
	char *ret;
	char *arg;
	bool needfree = false;

	checkLuaExecContext(L);

	lua_gasuse(L, 500);

	if (type == 'S' || type == 'U') {
		if (lua_isnil(L, 1)) {
			return 0;
		}
		switch(lua_type(L, 1)) {
		case LUA_TNUMBER:
			arg = (char *) lua_tostring(L, 1);
			break;
		case LUA_TSTRING:
			arg = (char *) lua_tostring(L, 1);
			break;
		case LUA_TUSERDATA:
			arg = lua_get_bignum_str(L, 1);
			if (arg == NULL) {
				luaL_error(L, "not enough memory");
			}
			needfree = true;
			break;
		default:
			luaL_error(L, "invalid input");
		}
	} else {
		arg = lua_util_get_json_from_stack(L, 1, lua_gettop(L), false);
		if (arg == NULL) {
			luaL_throwerror(L);
		}
		if (strlen(arg) == 0) {
			free(arg);
			luaL_error(L, "invalid input");
		}
		needfree = true;
	}

	ret = luaGovernance(L, type, arg);
	if (needfree) {
		free(arg);
	}
	if (ret != NULL) {
		strPushAndRelease(L, ret);
		luaL_throwerror(L);
	}
	return 0;
}

static int moduleStake(lua_State *L) {
	return governance(L, 'S');
}

static int moduleUnstake(lua_State *L) {
	return governance(L, 'U');
}

static int moduleVote(lua_State *L) {
	return governance(L, 'V');
}

static int moduleVoteDao(lua_State *L) {
	return governance(L, 'D');
}

static const luaL_Reg call_methods[] = {
	{"value", call_value},
	{"amount", call_value},
	{"gas", call_gas},
	{NULL, NULL}
};

static const luaL_Reg call_meta[] = {
	{"__call", moduleCall},
	{NULL, NULL}
};

static const luaL_Reg delegate_call_methods[] = {
	{"gas", delegate_call_gas},
	{NULL, NULL}
};

static const luaL_Reg delegate_call_meta[] = {
	{"__call", moduleDelegateCall},
	{NULL, NULL}
};

static const luaL_Reg deploy_call_methods[] = {
	{"value", deploy_value},
	{"amount", deploy_value},
	{NULL, NULL}
};

static const luaL_Reg deploy_call_meta[] = {
	{"__call", moduleDeploy},
	{NULL, NULL}
};

static const luaL_Reg contract_lib[] = {
	{"balance", moduleBalance},
	{"send", moduleSend},
	{"pcall", modulePcall},
	{"event", moduleEvent},
	{"stake", moduleStake},
	{"unstake", moduleUnstake},
	{"vote", moduleVote},
	{"voteDao", moduleVoteDao},
	{NULL, NULL}
};

int luaopen_contract(lua_State *L) {

	luaL_register(L, contract_str, contract_lib);

	lua_createtable(L, 0, 3);
	luaL_register(L, NULL, call_methods);
	lua_createtable(L, 0, 1);
	luaL_register(L, NULL, call_meta);
	lua_setmetatable(L, -2);
	lua_setfield(L, -2, call_str);

	lua_createtable(L, 0, 1);
	luaL_register(L, NULL, delegate_call_methods);
	lua_createtable(L, 0, 1);
	luaL_register(L, NULL, delegate_call_meta);
	lua_setmetatable(L, -2);
	lua_setfield(L, -2, delegatecall_str);

	lua_createtable(L, 0, 2);
	luaL_register(L, NULL, deploy_call_methods);
	lua_createtable(L, 0, 1);
	luaL_register(L, NULL, deploy_call_meta);
	lua_setmetatable(L, -2);
	lua_setfield(L, -2, deploy_str);

	lua_pop(L, 1);
	return 1;
}
