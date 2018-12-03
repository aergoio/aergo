#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "util.h"
#include "_cgo_export.h"

extern const bc_ctx_t *getLuaExecContext(lua_State *L);

static int systemPrint(lua_State *L)
{
    char *jsonValue;
	bc_ctx_t *exec = (bc_ctx_t *)getLuaExecContext(L);

    jsonValue = lua_util_get_json_from_stack (L, 1, lua_gettop(L), true);
    if (jsonValue == NULL) {
		lua_error(L);
	}
    LuaPrint(exec->contractId, jsonValue);
    free(jsonValue);
	return 0;
}

int setItem(lua_State *L)
{
	const char *key;
	char *jsonValue;
	char *dbKey;
	bc_ctx_t *exec = (bc_ctx_t *)getLuaExecContext(L);

	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}

	if (exec->isQuery) {
	    luaL_error(L, "set not permitted in query");
	}

	luaL_checkany(L, 2);
	key = luaL_checkstring(L, 1);

	jsonValue = lua_util_get_json (L, -1, false);
	if (jsonValue == NULL) {
		lua_error(L);
	}

	dbKey = lua_util_get_db_key(exec, key);
	if (LuaSetDB(L, exec->stateKey, dbKey, jsonValue) != 0) {
		free(jsonValue);
		free(dbKey);
		lua_error(L);
	}
	free(jsonValue);
	free(dbKey);

	return 0;
}

int getItem(lua_State *L)
{
	const char *key;
	char *dbKey;
	bc_ctx_t *exec = (bc_ctx_t *)getLuaExecContext(L);
	char *jsonValue;
	int ret;

	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	key = luaL_checkstring(L, 1);
	dbKey = lua_util_get_db_key(exec, key);

	ret = LuaGetDB(L, exec->stateKey, dbKey);

	free(dbKey);
	if (ret < 0) {
		lua_error(L);
	}
	if (ret == 0)
		return 0;
	jsonValue = (char *)luaL_checkstring(L, -1);
	lua_pop(L, 1);

	if (lua_util_json_to_lua(L, jsonValue, false) != 0) {
		luaL_error(L, "getItem error : can't convert %s", jsonValue);
	}
	return 1;
}

static int getSender(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	lua_pushstring(L, exec->sender);
	return 1;
}

static int getTxhash(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	lua_pushstring(L, exec->txHash);
	return 1;
}

static int getBlockHeight(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	lua_pushinteger(L, exec->blockHeight);
	return 1;
}

static int getTimestamp(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	lua_pushinteger(L, exec->timestamp);
	return 1;
}

static int getContractID(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	lua_pushstring(L, exec->contractId);
	return 1;
}

static int getCreator(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);
	int ret;

	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	ret = LuaGetDB(L, exec->stateKey, "Creator");
	if (ret < 0) {
		lua_error(L);
	}
	if (ret == 0)
		return 0;
	luaL_checkstring(L, -1);
	return 1;
}

static int getAmount(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);

	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	lua_pushstring(L, exec->amount);
	return 1;
}

static int getOrigin(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);

	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}
	lua_pushstring(L, exec->origin);
	return 1;
}

static const luaL_Reg sys_lib[] = {
	{"print", systemPrint},
	{"setItem", setItem},
	{"getItem", getItem},
	{"getSender", getSender},
	{"getCreator", getCreator},
	{"getTxhash", getTxhash},
	{"getBlockheight", getBlockHeight},
	{"getTimestamp", getTimestamp},
	{"getContractID", getContractID},
	{"getOrigin", getOrigin},
	{"getAmount", getAmount},
	{NULL, NULL}
};

int luaopen_system(lua_State *L)
{
	luaL_register(L, "system", sys_lib);
	lua_pop(L, 1);
	return 1;
}
