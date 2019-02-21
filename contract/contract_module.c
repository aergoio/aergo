#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "util.h"
#include "lbc.h"
#include "config.h"
#include "_cgo_export.h"

extern const int *getLuaExecContext(lua_State *L);

static const char *contract_str = "contract";
static const char *call_str = "call";
static const char *delegatecall_str = "delegatecall";
static const char *deploy_str = "deploy";
static const char *amount_str = "amount";
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
	    char *str = bc_num2str(*((void **)luaL_checkudata(L, 1, MYTYPE)));
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
	int ret;
	int *service = (int *)getLuaExecContext(L);
	lua_Integer gas;
	char *amount;

	if (service == NULL) {
		luaL_error(L, "cannot find execution context");
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
	fname = (char *)luaL_checkstring(L, 3);
	json_args = lua_util_get_json_from_stack (L, 4, lua_gettop(L), false);
	if (json_args == NULL) {
		lua_error(L);
	}

	if ((ret = LuaCallContract(L, service, contract, fname, json_args, amount, gas)) < 0) {
		free(json_args);
		lua_error(L);
	}
	free(json_args);
	reset_amount_info(L);
	return ret;
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
	int ret;
	int *service = (int *)getLuaExecContext(L);
	lua_Integer gas;

	if (service == NULL) {
		luaL_error(L, "cannot find execution context");
	}

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
		lua_error(L);
	}
	if ((ret = LuaDelegateCallContract(L, service, contract, fname, json_args, gas)) < 0) {
		free(json_args);
		lua_error(L);
	}
	free(json_args);
	reset_amount_info(L);

	return ret;
}

static int moduleSend(lua_State *L)
{
	char *contract;
	int ret;
	int *service = (int *)getLuaExecContext(L);
	char *amount;
	bool needfree = false;

	if (service == NULL) {
		luaL_error(L, "cannot find execution context");
	}
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
        amount = bc_num2str(*((void **)luaL_checkudata(L, 2, MYTYPE)));
        needfree = true;
        break;
    default:
		luaL_error(L, "invalid input");
	}
	ret = LuaSendAmount(L, service, contract, amount);
	if (needfree)
	    free(amount);
	if (ret < 0)
		lua_error(L);
	return 0;
}

static int moduleBalance(lua_State *L)
{
	char *contract;
	int ret;
	int *service = (int *)getLuaExecContext(L);
	lua_Integer amount;

	if (service == NULL) {
		luaL_error(L, "cannot find execution context");
	}

    if (lua_gettop(L) == 0 || lua_isnil(L, 1))
        contract = NULL;
    else {
	    contract = (char *)luaL_checkstring(L, 1);
	}
	if ((ret = LuaGetBalance(L, service, contract)) < 0) {
		lua_error(L);
	}

	return 1;
}

static int modulePcall(lua_State *L)
{
	int argc = lua_gettop(L) - 1;
	int *service = (int *)getLuaExecContext(L);
	int start_seq = -1;

	if (service == NULL) {
		luaL_error(L, "cannot find execution context");
	}

	start_seq = LuaSetRecoveryPoint(L, service);
	if (start_seq < 0)
	    lua_error(L);

	if (lua_pcall(L, argc, LUA_MULTRET, 0) != 0) {
		lua_pushboolean(L, false);
		lua_insert(L, 1);
		if (start_seq > 0) {
			if (LuaClearRecovery(L, service, start_seq, true) < 0)
				lua_error(L);
		}
		return 2;
	}
	lua_pushboolean(L, true);
	lua_insert(L, 1);
	if (start_seq > 0) {
		if (LuaClearRecovery(L, service, start_seq, false) < 0)
			lua_error(L);
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
	int ret;
	int *service = (int *)getLuaExecContext(L);
	char *amount;

	if (service == NULL) {
		luaL_error(L, "cannot find execution context");
	}

	lua_getfield(L, 1, amount_str);
	if (lua_isnil(L, -1))
		amount = NULL;
	else
		amount = (char *)luaL_checkstring(L, -1);
	lua_pop(L, 1);
	contract = (char *)luaL_checkstring(L, 2);
	json_args = lua_util_get_json_from_stack (L, 3, lua_gettop(L), false);
	if (json_args == NULL) {
		lua_error(L);
	}
	if ((ret = LuaDeployContract(L, service, contract, json_args, amount)) < 0) {
		free(json_args);
		lua_error(L);
	}
	free(json_args);
	reset_amount_info(L);
	return ret;
}

static int moduleEvent(lua_State *L)
{
	char *event_name;
	char *json_args;
	int ret;
	int *service = (int *)getLuaExecContext(L);

	if (service == NULL) {
		luaL_error(L, "cannot find execution context");
	}

	event_name = (char *)luaL_checkstring(L, 1);
	json_args = lua_util_get_json_from_stack (L, 2, lua_gettop(L), false);
	if (json_args == NULL) {
		lua_error(L);
	}
	LuaEvent(L, service, event_name, json_args);
	free(json_args);
	return ret;
}

static const luaL_Reg call_methods[] = {
	{"value", call_value},
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
	{NULL, NULL}
};

int luaopen_contract(lua_State *L)
{
	luaL_register(L, contract_str, contract_lib);
	lua_createtable(L, 0, 2);
	luaL_register(L, NULL, call_methods);
	lua_createtable(L, 0, 1);
	luaL_register(L, NULL, call_meta);
	lua_setmetatable(L, -2);
	lua_setfield(L, -2, call_str);

	lua_createtable(L, 0, 2);
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