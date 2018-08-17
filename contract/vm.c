#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "system_module.h"

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
	exec = (bc_ctx_t *)lua_touserdata(L, 1);
	lua_pop(L, 1);

	return exec;
}

const char *vm_loadbuff(const char *code, size_t sz, const char *name, bc_ctx_t *bc_ctx, lua_State **p)
{
	int err;
	lua_State *L = luaL_newstate();
	const char *errMsg = NULL;

    luaL_openlibs(L);
	preloadModules(L);
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
	*p = L;
	return NULL;
}

void vm_getfield(lua_State *L, const char *name)
{
	lua_getfield(L, LUA_GLOBALSINDEX, name);
}

const char *vm_pcall(lua_State *L, int argc)
{
	int err;
	const char *errMsg = NULL;

	err = lua_pcall(L, argc, 0, 0);
	if (err != 0) {
		errMsg = strdup(lua_tostring(L, -1));
		return errMsg;
	}
	return NULL;
}
