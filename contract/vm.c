#include <string.h>
#include <stdlib.h>
#include <stdint.h>
#include "vm.h"
#include "system_module.h"
#include "contract_module.h"
#include "name_module.h"
#include "db_module.h"
#include "state_module.h"
#include "crypto_module.h"
#include "util.h"
#include "bignum_module.h"
#include "_cgo_export.h"

const char *luaExecContext = "__exec_context__";
const char *construct_name = "constructor";
const char *VM_INST_LIMIT = "__INST_LIMIT__";
const char *VM_INST_COUNT = "__INST_COUNT_";
const int VM_TIMEOUT_INST_COUNT = 200;

extern int luaopen_utf8(lua_State *L);
extern void (*lj_internal_view_start)(lua_State *);
extern void (*lj_internal_view_end)(lua_State *);
void vm_internal_view_start(lua_State *L);
void vm_internal_view_end(lua_State *L);

int getLuaExecContext(lua_State *L) {
	int service = luaL_service(L);
	if (service < 0)
		luaL_error(L, "not permitted state referencing at global scope");
	return service;
}

#ifdef MEASURE
static int nsec(lua_State *L) {
	lua_pushnumber(L, luaL_nanosecond(L));
	return 1;
}
#endif

static void preloadModules(lua_State *L) {
	int status;

	luaopen_system(L);
	luaopen_contract(L);
	luaopen_state(L);
	luaopen_json(L);
	luaopen_crypto(L);
	luaopen_bignum(L);
	luaopen_utf8(L);

	if (vm_is_hardfork(L, 4)) {
		luaopen_name(L);
	}

	if (!isPublic()) {
		luaopen_db(L);
	}

#ifndef DEBUG
	if (vm_is_hardfork(L, 4)) {
		lua_getglobal(L, "_G");
		// disable getmetatable
		lua_pushnil(L);
		lua_setfield(L, -2, "getmetatable");
		// disable setmetatable
		lua_pushnil(L);
		lua_setfield(L, -2, "setmetatable");
		// disable rawget
		lua_pushnil(L);
		lua_setfield(L, -2, "rawget");
		// disable rawset
		lua_pushnil(L);
		lua_setfield(L, -2, "rawset");
		// disable rawequal
		lua_pushnil(L);
		lua_setfield(L, -2, "rawequal");
		lua_pop(L, 1);
		// disable string.dump
		lua_getglobal(L, "string");
		lua_pushnil(L);
		lua_setfield(L, -2, "dump");
		lua_pop(L, 1);
	}
#endif

#ifdef MEASURE
	lua_register(L, "nsec", nsec);
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

// overridden version of pcall
// used to rollback state and drop events upon error
static int pcall(lua_State *L) {
	int argc = lua_gettop(L);
	int service = getLuaExecContext(L);
	int num_events = luaGetEventCount(L, service);
	struct luaSetRecoveryPoint_return start_seq;
	int ret;

	if (argc < 1) {
		return luaL_error(L, "pcall: not enough arguments");
	}

	lua_gasuse(L, 300);

	start_seq = luaSetRecoveryPoint(L, service);
	if (start_seq.r0 < 0) {
		strPushAndRelease(L, start_seq.r1);
		luaL_throwerror(L);
	}

	// the stack is like this:
	//   func arg1 arg2 ... argn

	// call the function
	ret = lua_pcall(L, argc - 1, LUA_MULTRET, 0);

	// if failed, drop the events
	if (ret != 0) {
		if (vm_is_hardfork(L, 4)) {
			luaDropEvent(L, service, num_events);
		}
	}

	// throw the error if out of memory
	if (ret == LUA_ERRMEM) {
		luaL_throwerror(L);
	}

	// insert the status at the bottom of the stack
	lua_pushboolean(L, ret == 0);
	lua_insert(L, 1);

	// release the recovery point or revert the contract state
	if (start_seq.r0 > 0) {
		bool is_error = (ret != 0);
		char *errStr = luaClearRecovery(L, service, start_seq.r0, is_error);
		if (errStr != NULL) {
			if (vm_is_hardfork(L, 4)) {
				luaDropEvent(L, service, num_events);
			}
			strPushAndRelease(L, errStr);
			luaL_throwerror(L);
		}
	}

	// return the number of items in the stack
	return lua_gettop(L);
}

// overridden version of xpcall
// used to rollback state and drop events upon error
static int xpcall(lua_State *L) {
	int argc = lua_gettop(L);
	int service = getLuaExecContext(L);
	int num_events = luaGetEventCount(L, service);
	struct luaSetRecoveryPoint_return start_seq;
	int ret, errfunc;

	if (argc < 2) {
		return luaL_error(L, "xpcall: not enough arguments");
	}

	lua_gasuse(L, 300);

	start_seq = luaSetRecoveryPoint(L, service);
	if (start_seq.r0 < 0) {
		strPushAndRelease(L, start_seq.r1);
		luaL_throwerror(L);
	}

	// the stack is like this:
	//   func errfunc arg1 arg2 ... argn

	// check the error handler
	errfunc = 2;
	if (!lua_isfunction(L, errfunc)) {
		return luaL_error(L, "xpcall: error handler is not a function");
	}

	// move the error handler to the first position
	lua_pushvalue(L, 1);  // function
	lua_pushvalue(L, 2);  // error handler
	lua_replace(L, 1);    // 1: error handler
	lua_replace(L, 2);    // 2: function

	// now the stack is like this:
	//   errfunc func arg1 arg2 ... argn

	// update the error handler position
	errfunc = 1;

	// call the function
	ret = lua_pcall(L, argc - 2, LUA_MULTRET, errfunc);

	// if failed, drop the events
	if (ret != 0) {
		if (vm_is_hardfork(L, 4)) {
			luaDropEvent(L, service, num_events);
		}
	}

	// throw the error if out of memory
	if (ret == LUA_ERRMEM) {
		luaL_throwerror(L);
	}

	// ensure the stack has 1 free slot
	if (!lua_checkstack(L, 1)) {
		// return: false, "stack overflow"
		lua_settop(L, 0);
		lua_pushboolean(L, 0);
		lua_pushliteral(L, "stack overflow");
		return 2;
	}

	// store the status at the bottom of the stack, replacing the error handler
	lua_pushboolean(L, ret == 0);
	lua_replace(L, 1);

	// release the recovery point or revert the contract state
	if (start_seq.r0 > 0) {
		bool is_error = (ret != 0);
		char *errStr = luaClearRecovery(L, service, start_seq.r0, is_error);
		if (errStr != NULL) {
			if (vm_is_hardfork(L, 4)) {
				luaDropEvent(L, service, num_events);
			}
			strPushAndRelease(L, errStr);
			luaL_throwerror(L);
		}
	}

	// return the number of items in the stack
	return lua_gettop(L);
}

static const struct luaL_Reg _basefuncs[] = {
	{"pcall", pcall},
	{"xpcall", xpcall},
	{NULL, NULL}};

static void override_basefuncs(lua_State *L) {
	// override Lua builtin functions
	lua_getglobal(L, "_G");
	luaL_register(L, NULL, _basefuncs);
	lua_pop(L, 1);
}

static int loadLibs(lua_State *L) {
	luaL_openlibs(L);
	preloadModules(L);
	if (vm_is_hardfork(L, 4)) {
		// override pcall to drop events upon error
		override_basefuncs(L);
	}
	return 0;
}

lua_State *vm_newstate(int hardfork_version) {
	lua_State *L = luaL_newstate(hardfork_version);
	if (L == NULL)
		return NULL;
	// hardfork version set before loading modules
	int status = lua_cpcall(L, loadLibs, NULL);
	if (status != 0)
		return NULL;
	return L;
}

void vm_closestates(lua_State *s[], int count) {
	int i;

	for (i = 0; i < count; ++i)
		if (s[i] != NULL)
			lua_close(s[i]);
}

void initViewFunction() {
	lj_internal_view_start = vm_internal_view_start;
	lj_internal_view_end = vm_internal_view_end;
}

bool vm_is_hardfork(lua_State *L, int version) {
	int v = luaL_hardforkversion(L);
	return v >= version;
}

const char *vm_loadcall(lua_State *L) {
	int err;

	if (lua_usegas(L)) {
		lua_enablegas(L);
		vm_set_timeout_hook(L);
	} else {
		if (vm_is_hardfork(L, 2)) {
			vm_set_timeout_count_hook(L, 5000000);
		} else {
			vm_set_count_hook(L, 5000000);
		}
		luaL_enablemaxmem(L);
	}

	err = lua_pcall(L, 0, 0, 0);

	if (lua_usegas(L)) {
		lua_disablegas(L);
	} else {
		luaL_disablemaxmem(L);
	}

	lua_sethook(L, NULL, 0, 0);

	if (err != 0) {
		return lua_tostring(L, -1);
	}
	return NULL;
}

static int cp_getLuaExecContext(lua_State *L) {
	int *service = (int *)lua_topointer(L, 1);
	*service = getLuaExecContext(L);
	return 0;
}

const char *vm_copy_service(lua_State *L, lua_State *main) {
	int service;
	service = luaL_service(main);
	if (service < 0) {
		return "not permitted state referencing at global scope";
	}
	luaL_set_service(L, service);
	return NULL;
}

const char *vm_loadbuff(lua_State *L, const char *code, size_t sz, char *hex_id, int service) {
	int err;

	luaL_set_service(L, service);
	err = luaL_loadbuffer(L, code, sz, hex_id);
	if (err != 0) {
		return lua_tostring(L, -1);
	}

	return NULL;
}

int vm_autoload(lua_State *L, char *fname) {
	lua_getfield(L, LUA_GLOBALSINDEX, fname);
	return lua_isnil(L, -1) == 0;
}

void vm_remove_constructor(lua_State *L) {
	lua_pushnil(L);
	lua_setfield(L, LUA_GLOBALSINDEX, construct_name);
}

static void count_hook(lua_State *L, lua_Debug *ar) {
	luaL_setuncatchablerror(L);
	lua_pushstring(L, "exceeded the maximum instruction count");
	luaL_throwerror(L);
}

void vm_set_count_hook(lua_State *L, int limit) {
	lua_sethook(L, count_hook, LUA_MASKCOUNT, limit);
}

static void timeout_hook(lua_State *L, lua_Debug *ar) {
	int errCode = luaCheckTimeout(luaL_service(L));
	if (errCode == 1) {
		luaL_setuncatchablerror(L);
		lua_pushstring(L, ERR_BF_TIMEOUT);
		luaL_throwerror(L);
	} else if (errCode == -1) {
		luaL_error(L, "cannot find execution context");
	}
}

void vm_set_timeout_hook(lua_State *L) {
	if (vm_is_hardfork(L, 2)) {
		lua_sethook(L, timeout_hook, LUA_MASKCOUNT, VM_TIMEOUT_INST_COUNT);
	}
}

static void timeout_count_hook(lua_State *L, lua_Debug *ar) {
	int errCode;
	int inst_count, new_inst_count, inst_limit;

	timeout_hook(L, ar);

	inst_count = luaL_tminstcount(L);
	inst_limit = luaL_tminstlimit(L);
	new_inst_count = inst_count + VM_TIMEOUT_INST_COUNT;
	if (new_inst_count <= 0 || new_inst_count > inst_limit) {
		luaL_setuncatchablerror(L);
		lua_pushstring(L, "exceeded the maximum instruction count");
		luaL_throwerror(L);
	}
	luaL_set_tminstcount(L, new_inst_count);
}

void vm_set_timeout_count_hook(lua_State *L, int limit) {
	luaL_set_tminstlimit(L, limit);
	luaL_set_tminstcount(L, 0);
	lua_sethook(L, timeout_count_hook, LUA_MASKCOUNT, VM_TIMEOUT_INST_COUNT);
}

const char *vm_pcall(lua_State *L, int argc, int *nresult) {
	int err;
	int nr = lua_gettop(L) - argc - 1;

	if (lua_usegas(L)) {
		lua_enablegas(L);
	} else {
		luaL_enablemaxmem(L);
	}

	err = lua_pcall(L, argc, LUA_MULTRET, 0);

	if (lua_usegas(L)) {
		lua_disablegas(L);
	} else {
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

const char *vm_get_json_ret(lua_State *L, int nresult, int *err) {
	int top;
	char *json_ret;

	top = lua_gettop(L);
	json_ret = lua_util_get_json_from_stack(L, top - nresult + 1, top, true);

	if (json_ret == NULL) {
		*err = 1;
		return lua_tostring(L, -1);
	}

	lua_pushstring(L, json_ret);
	free(json_ret);

	return lua_tostring(L, -1);
}

const char *vm_copy_result(lua_State *L, lua_State *target, int cnt) {
	int i;
	int top;
	char *json;

	if (lua_usegas(L)) {
		lua_disablegas(target);
	} else {
		luaL_disablemaxmem(target);
	}

	top = lua_gettop(L);
	for (i = top - cnt + 1; i <= top; ++i) {
		json = lua_util_get_json(L, i, false);
		if (json == NULL) {
			if (lua_usegas(L)) {
				lua_enablegas(target);
			} else {
				luaL_enablemaxmem(target);
			}
			return lua_tostring(L, -1);
		}

		minus_inst_count(L, strlen(json));
		lua_util_json_to_lua(target, json, false);
		free(json);
	}

	if (lua_usegas(L)) {
		lua_enablegas(target);
	} else {
		luaL_enablemaxmem(target);
	}

	return NULL;
}

sqlite3 *vm_get_db(lua_State *L) {
	int service;
	sqlite3 *db;

	service = getLuaExecContext(L);
	db = luaGetDbHandle(service);
	if (db == NULL) {
		lua_pushstring(L, "can't open a database connection");
		luaL_throwerror(L);
	}
	return db;
}

void vm_get_abi_function(lua_State *L, char *fname) {
	lua_getfield(L, LUA_GLOBALSINDEX, "abi");
	lua_getfield(L, -1, "call");
	lua_pushstring(L, fname);
}

void vm_internal_view_start(lua_State *L) {
	luaViewStart(getLuaExecContext(L));
}

void vm_internal_view_end(lua_State *L) {
	luaViewEnd(getLuaExecContext(L));
}

int vm_instcount(lua_State *L) {
	if (lua_usegas(L)) {
		return 0;
	}
	if (vm_is_hardfork(L, 2)) {
		return luaL_tminstlimit(L);
	} else {
		return luaL_instcount(L);
	}
}

void vm_setinstcount(lua_State *L, int count) {
	if (lua_usegas(L)) {
		return;
	}
	if (vm_is_hardfork(L, 2)) {
		luaL_set_tminstlimit(L, count);
	} else {
		luaL_setinstcount(L, count);
	}
}
