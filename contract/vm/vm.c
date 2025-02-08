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

void checkLuaExecContext(lua_State *L) {
	if (luaL_is_loading(L)) {
		luaL_error(L, "state referencing not permitted at global scope");
	}
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
	checkLuaExecContext(L);
	struct luaSetRecoveryPoint_return start_seq;
	int ret;

	if (argc < 1) {
		return luaL_error(L, "pcall: not enough arguments");
	}

	lua_gasuse(L, 300);

	start_seq = luaSetRecoveryPoint(L);
	if (start_seq.r0 < 0) {
		strPushAndRelease(L, start_seq.r1);
		luaL_throwerror(L);
	}

	// the stack is like this:
	//   func arg1 arg2 ... argn

	// call the function
	ret = lua_pcall(L, argc - 1, LUA_MULTRET, 0);

	// release the recovery point (on success) or revert the contract state (on error)
	if (start_seq.r0 > 0) {
		bool is_error = (ret != 0);
		char *errStr = luaClearRecovery(L, start_seq.r0, is_error);
		if (errStr != NULL) {
			strPushAndRelease(L, errStr);
			luaL_throwerror(L);
		}
	}

	// throw the error if out of memory
	if (ret == LUA_ERRMEM) {
		luaL_throwerror(L);
	}

	// insert the status at the bottom of the stack
	lua_pushboolean(L, ret == 0);
	lua_insert(L, 1);

	// return the number of items in the stack
	return lua_gettop(L);
}

// overridden version of xpcall
// used to rollback state and drop events upon error
static int xpcall(lua_State *L) {
	int argc = lua_gettop(L);
	checkLuaExecContext(L);
	struct luaSetRecoveryPoint_return start_seq;
	int ret, errfunc;

	if (argc < 2) {
		return luaL_error(L, "xpcall: not enough arguments");
	}

	lua_gasuse(L, 300);

	start_seq = luaSetRecoveryPoint(L);
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

	// release the recovery point (on success) or revert the contract state (on error)
	if (start_seq.r0 > 0) {
		bool is_error = (ret != 0);
		char *errStr = luaClearRecovery(L, start_seq.r0, is_error);
		if (errStr != NULL) {
			strPushAndRelease(L, errStr);
			luaL_throwerror(L);
		}
	}

	// throw the error if out of memory
	if (ret == LUA_ERRMEM) {
		luaL_throwerror(L);
	}

/*
	// ensure the stack has 1 free slot
	if (!lua_checkstack(L, 1)) {
		// return: false, "stack overflow"
		lua_settop(L, 0);
		lua_pushboolean(L, 0);
		lua_pushliteral(L, "stack overflow");
		return 2;
	}
*/

	// store the status at the bottom of the stack, replacing the error handler
	lua_pushboolean(L, ret == 0);
	lua_replace(L, 1);

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

bool vm_is_hardfork(lua_State *L, int version) {
	int v = luaL_hardforkversion(L);
	return v >= version;
}

// execute code from the global scope, like declaring state variables and functions
// as well as abi.register, abi.register_view, abi.payable, etc.
const char *vm_pre_run(lua_State *L) {
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

	// remove hook
	lua_sethook(L, NULL, 0, 0);

	if (err != 0) {
		return lua_tostring(L, -1);
	}
	return NULL;
}

// load the code into the Lua state
const char *vm_load_code(lua_State *L, const char *code, size_t sz, char *hex_id) {
	int err;

	// mark as running on global scope
	luaL_set_loading(L, true);

	err = luaL_loadbuffer(L, code, sz, hex_id);
	if (err != 0) {
		return lua_tostring(L, -1);
	}

	return NULL;
}

void vm_push_abi_function(lua_State *L, char *fname) {
	lua_getfield(L, LUA_GLOBALSINDEX, "abi");
	lua_getfield(L, -1, "call");
	lua_pushstring(L, fname);
}

int vm_push_global_function(lua_State *L, char *fname) {
	lua_getfield(L, LUA_GLOBALSINDEX, fname);
	return lua_isnil(L, -1) == 0;
}

void vm_remove_constructor(lua_State *L) {
	lua_pushnil(L);
	lua_setfield(L, LUA_GLOBALSINDEX, construct_name);
}

const char *vm_call(lua_State *L, int argc, int *nresult) {
	int err;
	int nr = lua_gettop(L) - argc - 1;

	// mark as no longer loading, now running the function call
	luaL_set_loading(L, false);

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

const char *vm_get_json_ret(lua_State *L, int nresult, bool has_parent, int *err) {
	int top;
	char *json_ret;

	top = lua_gettop(L);
	json_ret = lua_util_get_json_from_stack(L, top - nresult + 1, top, !has_parent);

	if (json_ret == NULL) {
		*err = 1;
		return lua_tostring(L, -1);
	}

	lua_pushstring(L, json_ret);
	free(json_ret);

	return lua_tostring(L, -1);
}

// VIEW FUNCTIONS

void vm_internal_view_start(lua_State *L) {
	luaViewStart();
}

void vm_internal_view_end(lua_State *L) {
	luaViewEnd();
}

void initViewFunction() {
	lj_internal_view_start = vm_internal_view_start;
	lj_internal_view_end = vm_internal_view_end;
}

// INSTRUCTION COUNT

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

// INSTRUCTION COUNT

// this function is called at every N instructions
static void count_hook(lua_State *L, lua_Debug *ar) {
	luaL_setuncatchablerror(L);
	lua_pushstring(L, "exceeded the maximum instruction count");
	luaL_throwerror(L);
}

// set instruction count hook
void vm_set_count_hook(lua_State *L, int limit) {
	lua_sethook(L, count_hook, LUA_MASKCOUNT, limit);
}

// TIMEOUT

// this function is called at every N instructions
static void timeout_hook(lua_State *L, lua_Debug *ar) {
	int errCode = luaCheckTimeout();
	if (errCode == 1) {
		luaL_setuncatchablerror(L);
		lua_pushstring(L, ERR_BF_TIMEOUT);
		luaL_throwerror(L);
	}
}

// set timeout hook
void vm_set_timeout_hook(lua_State *L) {
	lua_sethook(L, timeout_hook, LUA_MASKCOUNT, VM_TIMEOUT_INST_COUNT);
}

// timeout and instruction count hook
// this function is called at every N instructions
static void timeout_count_hook(lua_State *L, lua_Debug *ar) {
	int inst_count, new_inst_count, inst_limit;

	// check for timeout
	timeout_hook(L, ar);

	// check instruction count
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

// set timeout and instruction count hook
void vm_set_timeout_count_hook(lua_State *L, int limit) {
	luaL_set_tminstlimit(L, limit);
	luaL_set_tminstcount(L, 0);
	lua_sethook(L, timeout_count_hook, LUA_MASKCOUNT, VM_TIMEOUT_INST_COUNT);
}
