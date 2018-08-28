#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "system_module.h"
#include "util.h"

const char *luaExecContext= "__exec_context__";

static void preloadModules(lua_State *L)
{
	luaopen_system(L);
}

static void setLuaExecContext(lua_State *L, bc_ctx_t *bc_ctx)
{
	lua_pushlightuserdata(L, bc_ctx);
	lua_setglobal(L, luaExecContext);
}

const bc_ctx_t *getLuaExecContext(lua_State *L)
{
	bc_ctx_t *exec;
	lua_getglobal(L, luaExecContext);
	exec = (bc_ctx_t *)lua_touserdata(L, -1);
	lua_pop(L, 1);

	return exec;
}

lua_State *vm_newstate()
{
	lua_State *L = luaL_newstate();
	luaL_openlibs(L);
	preloadModules(L);
	return L;
}

const char *vm_loadbuff(lua_State *L, const char *code, size_t sz, const char *name, bc_ctx_t *bc_ctx)
{
	int err;
	const char *errMsg = NULL;

	setLuaExecContext(L, bc_ctx);

	err = luaL_loadbuffer(L, code, sz, name);
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

const char *vm_pcall(lua_State *L, int argc, int *nresult)
{
	int err;
	const char *errMsg = NULL;
	int nr = lua_gettop(L);

	err = lua_pcall(L, argc, LUA_MULTRET, 0);
	if (err != 0) {
		errMsg = strdup(lua_tostring(L, -1));
		return errMsg;
	}
	*nresult = lua_gettop(L) - nr + 1;
	return NULL;
}

const char *vm_get_json_ret(lua_State *L, int nresult)
{
	sbuff_t sbuf;
	lua_util_sbuf_init(&sbuf, 64);

	lua_pushstring(L, lua_util_get_json_from_ret(L, nresult, &sbuf));
	free(sbuf.buf);
	
	return lua_tostring(L, -1);
}

