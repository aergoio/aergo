#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "util.h"
#include "_cgo_export.h"

extern const bc_ctx_t *getLuaExecContext(lua_State *L);
extern void LuaSetDB(char *key, char *value);

static int systemPrint(lua_State *L)
{
	printf ("systemPrinted");
	return 1;
}

static int setItem(lua_State *L)
{
    const char *key;
    char *jsonValue;
    char *dbKey;
	const bc_ctx_t *exec = getLuaExecContext(L);
 	if (exec == NULL) {
	    luaL_error(L, "cannot find execution context");
	}
	luaL_checkany(L, 2);
	key = luaL_checkstring(L, 1);

	jsonValue = lua_util_get_json (L, -1);

	dbKey = lua_util_get_db_key(exec, key);

    LuaSetDB(dbKey, jsonValue);
	free(jsonValue);
	free(dbKey);

	return 0;
}

static int getItem(lua_State *L)
{
    const char *key;
    char *dbKey;
	const bc_ctx_t *exec = getLuaExecContext(L);
	char *jsonValue;

 	if (exec == NULL) {
	    luaL_error(L, "cannot find execution context");
	}
	key = luaL_checkstring(L, 1);
	dbKey = lua_util_get_db_key(exec, key);

    jsonValue = LuaGetDB(dbKey);

    if (lua_util_json_to_lua(L, jsonValue) != 0) {
        luaL_error(L, "getItem error : can't convert %s", jsonValue);
    }
	free(dbKey);
	return 1;
}

static int delItem(lua_State *L)
{
    const char *key;
    char *dbKey;
	const bc_ctx_t *exec = getLuaExecContext(L);
	char *jsonValue;
 	if (exec == NULL) {
	    luaL_error(L, "cannot find execution context");
	}
	key = luaL_checkstring(L, 1);
	dbKey = lua_util_get_db_key(exec, key);

    LuaDelDB(dbKey);

	free(dbKey);
	return 0;
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

static int getBlockhash(lua_State *L)
{
	const bc_ctx_t *exec = getLuaExecContext(L);
	if (exec == NULL) {
	    luaL_error(L, "cannot find execution context");
	}
	lua_pushstring(L, exec->blockHash);
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
	{"delItem", delItem},
	{"getSender", getSender},
	{"getCreator", getContractID},
	{"getBlockhash", getBlockhash},
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
