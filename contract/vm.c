#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "system_module.h"
#include "contract_module.h"
#include "db_module.h"
#include "state_module.h"
#include "crypto_module.h"
#include "util.h"
#include "lbc.h"
#include "_cgo_export.h"

const char *luaExecContext= "__exec_context__";
const char *construct_name= "constructor";

static void preloadModules(lua_State *L)
{
    int status;

	luaopen_system(L);
	luaopen_contract(L);
	luaopen_state(L);
	luaopen_json(L);
	luaopen_crypto(L);
	luaopen_bc(L);
	if (!IsPublic()) {
        luaopen_db(L);
	}
#ifdef MEASURE
    lua_register(L, "nsec", lj_cf_nsec);
    luaopen_jit(L);
    lua_getfield(L, LUA_REGISTRYINDEX, "_LOADED");
    lua_getfield(L, -1, "jit");
    lua_remove(L, -2);
    lua_getfield(L, -1, "off");
    status = lua_pcall(L, 0, 0, 0);
    if (status != LUA_OK) {
        lua_pushstring(L, "cannot load the `jit` module");
        lua_error(L);
    }
    lua_remove(L, -1); /* remove jit.* */
#endif
}

static void setLuaExecContext(lua_State *L, int *service)
{
	lua_pushlightuserdata(L, service);
	lua_setglobal(L, luaExecContext);
}

const int *getLuaExecContext(lua_State *L)
{
	int *service;
	lua_getglobal(L, luaExecContext);
	service = (int *)lua_touserdata(L, -1);
	lua_pop(L, 1);

	return service;
}

lua_State *vm_newstate()
{
	lua_State *L = luaL_newstate();
	if (L == NULL)
		return NULL;
	luaL_openlibs(L);
	preloadModules(L);
	return L;
}

const char *vm_loadbuff(lua_State *L, const char *code, size_t sz, char *hex_id, int *service)
{
	int err;
	const char *errMsg = NULL;

	setLuaExecContext(L, service);

	err = luaL_loadbuffer(L, code, sz, hex_id);
	if (err != 0) {
		errMsg = strdup(lua_tostring(L, -1));
		return errMsg;
	}
	err = lua_pcall(L, 0, 0, 0);
	if (err != 0) {
		errMsg = strdup(lua_tostring(L, -1));
		return errMsg;
	}
	return NULL;
}

void vm_getfield(lua_State *L, const char *name)
{
	lua_getfield(L, LUA_GLOBALSINDEX, name);
}

int vm_isnil(lua_State *L, int idx)
{
	return lua_isnil(L, idx);
}

void vm_get_constructor(lua_State *L)
{
    lua_getfield(L, LUA_GLOBALSINDEX, construct_name);
}

void vm_remove_constructor(lua_State *L)
{
	lua_pushnil(L);
	lua_setfield(L, LUA_GLOBALSINDEX, construct_name);
}

static void count_hook(lua_State *L, lua_Debug *ar)
{
	lua_pushstring(L, "exceeded the maximum instruction count");
	lua_error(L);
}

void vm_set_count_hook(lua_State *L, int limit)
{
	lua_sethook (L, count_hook, LUA_MASKCOUNT, limit);
}

const char *vm_pcall(lua_State *L, int argc, int *nresult)
{
	int err;
	const char *errMsg = NULL;
	int nr = lua_gettop(L) - argc - 1;

	err = lua_pcall(L, argc, LUA_MULTRET, 0);
	if (err != 0) {
		errMsg = strdup(lua_tostring(L, -1));
		return errMsg;
	}
	*nresult = lua_gettop(L) - nr;
	return NULL;
}

const char *vm_get_json_ret(lua_State *L, int nresult)
{
	int top = lua_gettop(L);
	char *json_ret = lua_util_get_json_from_stack(L, top - nresult + 1, top, true);

	if (json_ret == NULL)
		return lua_tostring(L, -1);

	lua_pushstring(L, json_ret);
	free(json_ret);
	
	return lua_tostring(L, -1);
}

const char *vm_tostring(lua_State *L, int idx)
{
	return lua_tolstring(L, idx, NULL);
}

const char *vm_copy_result(lua_State *L, lua_State *target, int cnt)
{
	int i;
	int top = lua_gettop(L);
	char *json;

	for (i = top - cnt + 1; i <= top; ++i) {
		json = lua_util_get_json (L, i, false);
		if (json == NULL)
			return lua_tostring(L, -1);

		lua_util_json_to_lua(target, json, false);
		free (json);
	}
	return NULL;
}

sqlite3 *vm_get_db(lua_State *L)
{
    int *service;
    sqlite3 *db;

    service = (int *)getLuaExecContext(L);
    db = LuaGetDbHandle(service);
    if (db == NULL) {
        lua_pushstring(L, "can't open a database connection");
        lua_error(L);
    }
    return db;
}

void vm_get_abi_function(lua_State *L, char *fname)
{
	lua_getfield(L, LUA_GLOBALSINDEX, "abi");
	lua_getfield(L, -1, "call");
	lua_pushstring(L, fname);
}

int vm_is_payable_function(lua_State *L, char *fname)
{
    int err;
	lua_getfield(L, LUA_GLOBALSINDEX, "abi");
	lua_getfield(L, -1, "is_payable");
	lua_pushstring(L, fname);
	err = lua_pcall(L, 1, 1, 0);
	if (err != 0) {
	    return 0;
	}
	return lua_tointeger(L, -1);
}

char *vm_resolve_function(lua_State *L, char *fname, int *viewflag, int *payflag)
{
    int err;

	lua_getfield(L, LUA_GLOBALSINDEX, "abi");
	lua_getfield(L, -1, "resolve");
	lua_pushstring(L, fname);
	err = lua_pcall(L, 1, 3, 0);
	if (err != 0) {
		return NULL;
	}
    fname = (char *)lua_tostring(L, -3);
	if (fname == NULL)
	    return fname;
	*payflag = lua_tointeger(L, -2);
	*viewflag = lua_tointeger(L, -1);

	return fname;
}
