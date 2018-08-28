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

static int kpt_lua_Writer(struct lua_State *L, const void *p, size_t sz, void *u)
{
	return (fwrite(p, sz, 1, (FILE *)u) != 1) && (sz != 0);
}

const char *vm_compile(lua_State *L, const char *code, const char *byte, const char *abi)
{
	const char *errMsg = NULL;
	FILE *f = NULL;

	if (luaL_loadfile(L, code) != 0) {
		errMsg = strdup(lua_tostring(L, -1));
		lua_close(L);
		return errMsg;
	}
	f = fopen(byte, "wb");
	if (f == NULL) {
		return "cannot open a bytecode file";
	}
	if (lua_dump(L, kpt_lua_Writer, f) != 0) {
		errMsg = strdup(lua_tostring(L, -1));
	}
	fclose(f);

	if (abi != NULL && strlen(abi) > 0) {
		const char *r;
		if (lua_pcall(L, 0, 0, 0) != 0) {
		   errMsg = strdup(lua_tostring(L, -1));
		   return errMsg;
		}
		lua_getfield(L, LUA_GLOBALSINDEX, "abi");
		lua_getfield(L, -1, "generate");
		if (lua_pcall(L, 0, 1, 0) != 0) {
		    errMsg = strdup(lua_tostring(L, -1));
		    return errMsg;
		}
		if (!lua_isstring(L, -1)) {
		    return "cannot create a abi file";
		}
		r = lua_tostring(L, -1);
		f = fopen(abi, "wb");
		if (f == NULL) {
		    return "cannot open a abi file";
		}
		fwrite(r, 1, strlen(r), f);
		fclose(f);
	}

	return errMsg;
}
