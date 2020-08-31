#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "util.h"
#include "lgmp.h"
#include "_cgo_export.h"

extern int getLuaExecContext(lua_State *L);

static const char *contract_str = "contract";
static const char *call_str = "call";
static const char *delegatecall_str = "delegatecall";
static const char *deploy_str = "deploy";
static const char *amount_str = "amount_value";
static const char *fee_str = "fee";

static void set_call_obj(lua_State *L, const char* obj_name)
{
	lua_getglobal(L, contract_str);
	lua_getfield(L, -1, obj_name);
}

static void reset_amount_info (lua_State *L)
{
	lua_pushnil(L);
	lua_setfield(L, 1, amount_str);
	lua_pushnil(L);
	lua_setfield(L, 1, fee_str);
}

static int set_value(lua_State *L, const char *str)
{
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

static int set_gas(lua_State *L, const char *str)
{
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

static int call_value(lua_State *L)
{
    return set_value(L, call_str);
}

static int call_gas(lua_State *L)
{
    return set_gas(L, call_str);
}

static int moduleCall(lua_State *L)
{
	char *contract;
	char *fname;
	char *json_args;
	struct luaCallContract_return ret;
	int service = getLuaExecContext(L);
	lua_Integer gas;
	char *amount;

	if (lua_gettop(L) == 2) {
        lua_gasuse(L, 300);
    } else {
        lua_gasuse(L, 2000);
    }

	lua_getfield(L, 1, amount_str);
	if (lua_isnil(L, -1))
		amount = NULL;
	else
		amount = (char *)luaL_checkstring(L, -1);

	lua_getfield(L, 1, fee_str);
	if (lua_isnil(L, -1))
		gas = 0;
	else
		gas = luaL_checkinteger(L, -1);

	lua_pop(L, 2);
	contract = (char *)luaL_checkstring(L, 2);
	if (lua_gettop(L) == 2) {
	    char *errStr = luaSendAmount(L, service, contract, amount);
	    reset_amount_info(L);
	    if (errStr != NULL) {
            strPushAndRelease(L, errStr);
            luaL_throwerror(L);
        }
        return 0;
	}
	fname = (char *)luaL_checkstring(L, 3);
	json_args = lua_util_get_json_from_stack (L, 4, lua_gettop(L), false);
	if (json_args == NULL) {
	    reset_amount_info(L);
		luaL_throwerror(L);
	}

    ret = luaCallContract(L, service, contract, fname, json_args, amount, gas);
	if (ret.r1 != NULL) {
		free(json_args);
	    reset_amount_info(L);
		strPushAndRelease(L, ret.r1);
		luaL_throwerror(L);
	}
	free(json_args);
	reset_amount_info(L);
	return ret.r0;
}

static int delegate_call_gas(lua_State *L)
{
    return set_gas(L, delegatecall_str);
}

static int moduleDelegateCall(lua_State *L)
{
	char *contract;
	char *fname;
	char *json_args;
	struct luaDelegateCallContract_return ret;
	int service = getLuaExecContext(L);
	lua_Integer gas;

    lua_gasuse(L, 2000);

	lua_getfield(L, 1, fee_str);
	if (lua_isnil(L, -1))
		gas = 0;
	else
		gas = luaL_checkinteger(L, -1);

	lua_pop(L, 1);
	contract = (char *)luaL_checkstring(L, 2);
	fname = (char *)luaL_checkstring(L, 3);
	json_args = lua_util_get_json_from_stack (L, 4, lua_gettop(L), false);
	if (json_args == NULL) {
	    reset_amount_info(L);
		luaL_throwerror(L);
	}
	ret = luaDelegateCallContract(L, service, contract, fname, json_args, gas);
	if (ret.r1 != NULL) {
		free(json_args);
	    reset_amount_info(L);
		strPushAndRelease(L, ret.r1);
		luaL_throwerror(L);
	}
	free(json_args);
	reset_amount_info(L);

	return ret.r0;
}

static int moduleSend(lua_State *L)
{
	char *contract;
	char *errStr;
	int service = getLuaExecContext(L);
	char *amount;
	bool needfree = false;

    lua_gasuse(L, 300);

	contract = (char *)luaL_checkstring(L, 1);
	if (lua_isnil(L, 2))
	    return 0;

	switch(lua_type(L, 2)) {
    case LUA_TNUMBER:
        amount = (char *)lua_tostring(L, 2);
        break;
    case LUA_TSTRING:
        amount = (char *)lua_tostring(L, 2);
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
	errStr = luaSendAmount(L, service, contract, amount);
	if (needfree)
	    free(amount);
	if (errStr != NULL) {
        strPushAndRelease(L, errStr);
		luaL_throwerror(L);
    }
	return 0;
}

static int moduleBalance(lua_State *L)
{
	char *contract;
	int service = getLuaExecContext(L);
	lua_Integer amount;
	struct luaGetBalance_return balance;
    int nArg;

    lua_gasuse(L, 300);

    nArg = lua_gettop(L);
    if (nArg== 0 || lua_isnil(L, 1)) {
        contract = NULL;
    }
    else {
	    contract = (char *)luaL_checkstring(L, 1);
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
            ret = luaGetStaking(service, contract);
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

	balance = luaGetBalance(L, service, contract);
	if (balance.r1 != NULL) {
	    strPushAndRelease(L, balance.r1);
		luaL_throwerror(L);
	}

	strPushAndRelease(L, balance.r0);
	return 1;
}

static int modulePcall(lua_State *L)
{
	int argc = lua_gettop(L) - 1;
	int service = getLuaExecContext(L);
	struct luaSetRecoveryPoint_return start_seq;
	int ret;

    lua_gasuse(L, 300);

	start_seq = luaSetRecoveryPoint(L, service);
	if (start_seq.r0 < 0) {
	    strPushAndRelease(L, start_seq.r1);
	    luaL_throwerror(L);
    }

	if ((ret = lua_pcall(L, argc, LUA_MULTRET, 0)) != 0) {
	    if (ret == LUA_ERRMEM) {
			luaL_throwerror(L);
	    }
		lua_pushboolean(L, false);
		lua_insert(L, 1);
		if (start_seq.r0 > 0) {
		    char *errStr = luaClearRecovery(L, service, start_seq.r0, true);
			if (errStr != NULL) {
			    strPushAndRelease(L, errStr);
				luaL_throwerror(L);
            }
		}
		return 2;
	}
	lua_pushboolean(L, true);
	lua_insert(L, 1);
	if (start_seq.r0 == 1) {
        char *errStr = luaClearRecovery(L, service, start_seq.r0, false);
		if (errStr != NULL) {
			strPushAndRelease(L, errStr);
			luaL_throwerror(L);
        }
	}
	return lua_gettop(L);
}

static int deploy_value(lua_State *L)
{
    return set_value(L, deploy_str);
}

static int moduleDeploy(lua_State *L)
{
	char *contract;
	char *fname;
	char *json_args;
	struct luaDeployContract_return ret;
	int service = getLuaExecContext(L);
	char *amount;

    lua_gasuse(L, 5000);

	lua_getfield(L, 1, amount_str);
	if (lua_isnil(L, -1))
		amount = NULL;
	else
		amount = (char *)luaL_checkstring(L, -1);
	lua_pop(L, 1);
	contract = (char *)luaL_checkstring(L, 2);
	json_args = lua_util_get_json_from_stack (L, 3, lua_gettop(L), false);
	if (json_args == NULL) {
	    reset_amount_info(L);
		luaL_throwerror(L);
	}

	ret = luaDeployContract(L, service, contract, json_args, amount);
	if (ret.r0 < 0) {
		free(json_args);
	    reset_amount_info(L);
		strPushAndRelease(L, ret.r1);
		luaL_throwerror(L);
	}
	free(json_args);
	reset_amount_info(L);
	strPushAndRelease(L, ret.r1);
	if (ret.r0 > 1)
        lua_insert(L, -ret.r0);

	return ret.r0;
}

static int moduleEvent(lua_State *L)
{
	char *event_name;
	char *json_args;
	int service = getLuaExecContext(L);
	char *errStr;

    lua_gasuse(L, 500);

	event_name = (char *)luaL_checkstring(L, 1);
	if (vm_is_hardfork(L, 2)) {
	    json_args = lua_util_get_json_array_from_stack (L, 2, lua_gettop(L), true);
	}
	else {
	    json_args = lua_util_get_json_from_stack (L, 2, lua_gettop(L), false);
	}
	if (json_args == NULL) {
		luaL_throwerror(L);
	}
	errStr = luaEvent(L, service, event_name, json_args);
	if (errStr != NULL) {
	    strPushAndRelease(L, errStr);
	    luaL_throwerror(L);
	}
	free(json_args);
	return 0;
}

static int governance(lua_State *L, char type) {
	char *ret;
	int service = getLuaExecContext(L);
	char *arg;
	bool needfree = false;

    lua_gasuse(L, 500);

    if (type == 'S' || type == 'U') {
    	if (lua_isnil(L, 1))
	    return 0;

        switch(lua_type(L, 1)) {
        case LUA_TNUMBER:
            arg = (char *)lua_tostring(L, 1);
            break;
        case LUA_TSTRING:
            arg = (char *)lua_tostring(L, 1);
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
    }
    else {
	    arg = lua_util_get_json_from_stack (L, 1, lua_gettop(L), false);
	    if (arg == NULL)
		    luaL_throwerror(L);
		if (strlen(arg) == 0) {
		    free(arg);
		    luaL_error(L, "invalid input");
		}
		needfree = true;
    }
	ret = luaGovernance(L, service, type, arg);
	if (needfree)
	    free(arg);
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

int luaopen_contract(lua_State *L)
{
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
