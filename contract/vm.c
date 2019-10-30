#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "system_module.h"
#include "contract_module.h"
#include "db_module.h"
#include "state_module.h"
#include "crypto_module.h"
#include "util.h"
#include "lgmp.h"
#include "_cgo_export.h"

const char *luaExecContext= "__exec_context__";
const char *construct_name= "constructor";
const char *VM_INST_LIMIT = "__INST_LIMIT__";
const char *VM_INST_COUNT = "__INST_COUNT_";
const int VM_TIMEOUT_INST_COUNT = 200;
extern int luaopen_utf8 (lua_State *L);
extern void (*lj_internal_view_start)(lua_State *);
extern void (*lj_internal_view_end)(lua_State *);
void vm_internal_view_start(lua_State *L);
void vm_internal_view_end(lua_State *L);

static void preloadModules(lua_State *L)
{
    int status;

	luaopen_system(L);
	luaopen_contract(L);
	luaopen_state(L);
	luaopen_json(L);
	luaopen_crypto(L);
	luaopen_gmp(L);
    luaopen_utf8(L);

	if (!isPublic()) {
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

void initViewFunction()
{
    lj_internal_view_start = vm_internal_view_start;
    lj_internal_view_end = vm_internal_view_end;
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
	if (*service == -1)
	    luaL_error(L, "not permitted state referencing at global scope");

	return service;
}

void setHardforkV2(lua_State *L)
{
    lua_pushboolean(L, true);
	lua_setfield (L, LUA_REGISTRYINDEX, FORK_V2);
}

int isHardfork(lua_State *L, char *forkname)
{
	lua_getfield (L, LUA_REGISTRYINDEX, forkname);
	if (lua_isnil(L, -1)) {
	    lua_pop(L, 1);
	    return 0;
	}
	lua_pop(L, 1);
    return 1;
}

const char *VM_RESOURCE_LIMIT = "__VM_RESOURCE_LIMIT__";

void vm_set_resource_limit(lua_State *L)
{
    lua_pushboolean(L, true);
	lua_setfield (L, LUA_REGISTRYINDEX, VM_RESOURCE_LIMIT);
}

int vm_need_resource_limit(lua_State *L)
{
	lua_getfield (L, LUA_REGISTRYINDEX, VM_RESOURCE_LIMIT);
	if (lua_isnil(L, -1)) {
	    lua_pop(L, 1);
	    return 0;
	}
	lua_pop(L, 1);
    return 1;
}

static int loadLibs(lua_State *L)
{
	luaL_openlibs(L);
	preloadModules(L);
	return 0;
}

lua_State *vm_newstate()
{
	lua_State *L = luaL_newstate();
	int status;
	if (L == NULL)
		return NULL;
	status = lua_cpcall(L, loadLibs, NULL);
	if (status != 0)
	    return NULL;
	return L;
}

static void set_loadbuf_rlimit(lua_State *L)
{
    if (vm_need_resource_limit(L)) {
        vm_set_count_hook(L, 5000000);
        luaL_enablemaxmem(L);
    }
}

static void unset_loadbuf_rlimit(lua_State *L)
{
    if (vm_need_resource_limit(L)) {
        luaL_disablemaxmem(L);
        lua_sethook(L, NULL, 0, 0);
    }
}

static int pcall(lua_State *L, int narg, int nret)
{
    int err;
    set_loadbuf_rlimit(L);
    err = lua_pcall(L, narg, nret, 0);
    unset_loadbuf_rlimit(L);
    return err;
}

static int cp_setLuaExecContext(lua_State *L)
{
    int *service = (int *)lua_topointer(L, 1);
    setLuaExecContext(L, service);
    return 0;
}

const char *vm_loadbuff(lua_State *L, const char *code, size_t sz, char *hex_id, int *service)
{
	int err;

	err = lua_cpcall(L, cp_setLuaExecContext, service);
    if (err != 0) {
	    return lua_tostring(L, -1);
	}

    err = luaL_loadbuffer(L, code, sz, hex_id) || pcall(L, 0, 0);
    if (err != 0) {
	    return lua_tostring(L, -1);
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

void vm_get_autoload(lua_State *L, char *fname)
{
    lua_getfield(L, LUA_GLOBALSINDEX, fname);
}

void vm_remove_constructor(lua_State *L)
{
	lua_pushnil(L);
	lua_setfield(L, LUA_GLOBALSINDEX, construct_name);
}

static void count_hook(lua_State *L, lua_Debug *ar)
{
    luaL_setuncatchablerror(L);
	lua_pushstring(L, "exceeded the maximum instruction count");
	luaL_throwerror(L);
}

void vm_set_count_hook(lua_State *L, int limit)
{
	lua_sethook(L, count_hook, LUA_MASKCOUNT, limit);
}

static void timeout_hook(lua_State *L, lua_Debug *ar)
{
	int errCode = luaCheckTimeout(*getLuaExecContext(L));
    if (errCode == 1) {
        luaL_setuncatchablerror(L);
        lua_pushstring(L, ERR_BF_TIMEOUT);
        luaL_throwerror(L);
    } else if (errCode == -1) {
		luaL_error(L, "cannot find execution context");
    }
}

void vm_set_timeout_hook(lua_State *L)
{
    if (isHardfork(L, FORK_V2)) {
        lua_sethook(L, timeout_hook, LUA_MASKCOUNT, VM_TIMEOUT_INST_COUNT);
    }
}

static void timeout_count_hook(lua_State *L, lua_Debug *ar)
{
	int errCode;
	int inst_cnt, new_inst_cnt, inst_limit;

    timeout_hook(L, ar);

    lua_getfield(L, LUA_REGISTRYINDEX, VM_INST_COUNT);
    inst_cnt = lua_tointeger(L, -1);
    lua_getfield(L, LUA_REGISTRYINDEX, VM_INST_LIMIT);
    inst_limit = lua_tointeger(L, -1);
    lua_pop(L, 2);
    new_inst_cnt = inst_cnt + VM_TIMEOUT_INST_COUNT;
    if (new_inst_cnt <= 0 || new_inst_cnt > inst_limit) {
        luaL_setuncatchablerror(L);
        lua_pushstring(L, "exceeded the maximum instruction count");
        luaL_throwerror(L);
    }
    lua_pushinteger(L, new_inst_cnt);
    lua_setfield(L, LUA_REGISTRYINDEX, VM_INST_COUNT);
}

void vm_set_timeout_count_hook(lua_State *L, int limit)
{
    lua_pushinteger(L, limit);
    lua_setfield (L, LUA_REGISTRYINDEX, VM_INST_LIMIT);
    lua_pushinteger(L, 0);
    lua_setfield (L, LUA_REGISTRYINDEX, VM_INST_COUNT);

    lua_sethook(L, timeout_count_hook, LUA_MASKCOUNT, VM_TIMEOUT_INST_COUNT);
}

const char *vm_pcall(lua_State *L, int argc, int *nresult)
{
	int err;
	int nr = lua_gettop(L) - argc - 1;

    if (vm_need_resource_limit(L)) {
        luaL_enablemaxmem(L);
    }

	err = lua_pcall(L, argc, LUA_MULTRET, 0);

    if (vm_need_resource_limit(L)) {
        luaL_disablemaxmem(L);
    }

	if (err != 0) {
        lua_cpcall(L, lua_db_release_resource, NULL);
		return lua_tostring(L, -1);
	}
    err = lua_cpcall(L, lua_db_release_resource, NULL);
    if (err != 0) {
		return lua_tostring(L, -1);
    }
	*nresult = lua_gettop(L) - nr;
	return NULL;
}

const char *vm_get_json_ret(lua_State *L, int nresult, int* err)
{
	int top = lua_gettop(L);
	char *json_ret = lua_util_get_json_from_stack(L, top - nresult + 1, top, true);

	if (json_ret == NULL) {
	    *err = 1;
		return lua_tostring(L, -1);
    }

	lua_pushstring(L, json_ret);
	free(json_ret);
	
	return lua_tostring(L, -1);
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

		minus_inst_count(L, strlen(json));
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
    db = luaGetDbHandle(service);
    if (db == NULL) {
        luaL_setsyserror(L);
        lua_pushstring(L, "can't open a database connection");
        luaL_throwerror(L);
    }
    return db;
}

void vm_get_abi_function(lua_State *L, char *fname)
{
	lua_getfield(L, LUA_GLOBALSINDEX, "abi");
	lua_getfield(L, -1, "call");
	lua_pushstring(L, fname);
}

void vm_internal_view_start(lua_State *L)
{
    luaViewStart((int *)getLuaExecContext(L));
}

void vm_internal_view_end(lua_State *L)
{
    luaViewEnd((int *)getLuaExecContext(L));
}

int vm_instcount(lua_State *L)
{
    if (isHardfork(L, FORK_V2)) {
        int n;
        lua_getfield(L, LUA_REGISTRYINDEX, VM_INST_LIMIT);
        n = lua_tointeger(L, -1);
        lua_pop(L, 1);
        return n;
    } else {
        return luaL_instcount(L);
    }
}

void vm_setinstcount(lua_State *L, int cnt)
{
    if (isHardfork(L, FORK_V2)) {
        lua_pushinteger(L, cnt);
        lua_setfield(L, LUA_REGISTRYINDEX, VM_INST_LIMIT);
    } else {
        luaL_setinstcount(L, cnt);
    }
}
