#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "util.h"
#include "_cgo_export.h"

extern const bc_ctx_t *getLuaExecContext(lua_State *L);

static int systemPrint(lua_State *L)
{
	return 0;
}

static int setItem(lua_State *L)
{
	const char *key;
	char *jsonValue;
	char *dbKey;
	bc_ctx_t *exec = (bc_ctx_t *)getLuaExecContext(L);

	if (exec == NULL) {
		luaL_error(L, "cannot find execution context");
	}

	if (exec->isQuery) {
	    luaL_error(L, "not permitted set in query");
	}

	luaL_checkany(L, 2);
	key = luaL_checkstring(L, 1);

	jsonValue = lua_util_get_json (L, -1, false);
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

static int getItem(lua_State *L)
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

	if (lua_util_json_to_lua(L, jsonValue) != 0) {
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

static const luaL_Reg sys_lib[] = {
	{"print", systemPrint},
	{"setItem", setItem},
	{"getItem", getItem},
	{"getSender", getSender},
	{"getCreator", getContractID},
	{"getTxhash", getTxhash},
	{"getBlockheight", getBlockHeight},
	{"getTimestamp", getTimestamp},
	{"getContractID", getContractID},
	{NULL, NULL}
};

int luaopen_system(lua_State *L)
{
	luaL_register(L, "system", sys_lib);
	return 1;
}
