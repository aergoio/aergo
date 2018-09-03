
#include <stdlib.h>
#include <string.h>
#include <lualib.h>
#include <lauxlib.h>
#include <luajit.h>
#include "_cgo_export.h"

lua_State *vm_newstate()
{
	lua_State *L = luaL_newstate();
	luaL_openlibs(L);
	return L;
}

void vm_close(lua_State *L)
{
	if (L != NULL)
		lua_close(L);
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

static int writer_buf(lua_State *L, const void *p, size_t size, void *b)
{
  	luaL_addlstring((luaL_Buffer *)b, (const char *)p, size);
  	return 0;
}

const char *vm_loadfile(lua_State *L, const char *code)
{
	const char *errMsg = NULL;
	if (luaL_loadfile(L, code) != 0) {
		errMsg = strdup(lua_tostring(L, -1));
		lua_close(L);
		return errMsg;
	}
	return NULL;
}

const char *vm_loadstring(lua_State *L, const char *code)
{
	const char *errMsg = NULL;
	if (luaL_loadstring(L, code) != 0) {
		errMsg = strdup(lua_tostring(L, -1));
		lua_close(L);
		return errMsg;
	}
	return NULL;
}

const char *vm_stringdump(lua_State *L)
{
	const char *errMsg = NULL;
	luaL_Buffer b;

	luaL_buffinit(L, &b);
	if (lua_dump(L, writer_buf, &b) != 0) {
		errMsg = strdup(lua_tostring(L, -1));
	}
	luaL_pushresult(&b);
	if (!lua_isstring(L, -1)) {
		return "compile failed";
	}
	addLen(lua_strlen(L, -1));
	addByteN((char *)lua_tostring(L, -1), lua_strlen(L, -1));
	lua_pop(L, 1);

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
		return "abi generation is failed";
	}
	addByteN((char *)lua_tostring(L, -1), lua_strlen(L, -1));

	return errMsg;
}

