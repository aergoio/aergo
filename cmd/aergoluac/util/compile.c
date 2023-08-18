
#include <stdlib.h>
#include <string.h>
#include <lualib.h>
#include <lauxlib.h>
#include <luajit.h>
#include "state_module.h"
#include "_cgo_export.h"

lua_State *luac_vm_newstate() {
	lua_State *L = luaL_newstate(3);
	if (L == NULL) {
		return NULL;
	}
	luaL_openlibs(L);
	luac_open_state(L);
	return L;
}

void luac_vm_close(lua_State *L) {
	if (L != NULL) {
		lua_close(L);
	}
}

static int kpt_lua_Writer(struct lua_State *L, const void *p, size_t sz, void *u) {
	return (fwrite(p, sz, 1, (FILE *)u) != 1) && (sz != 0);
}

#define GEN_ABI() \
	do { \
		if (lua_pcall(L, 0, 0, 0) != 0) { \
			return lua_tostring(L, -1); \
		} \
		lua_getfield(L, LUA_GLOBALSINDEX, "abi"); \
		lua_getfield(L, -1, "autoload"); \
		if (lua_pcall(L, 0, 0, 0) != 0) { \
			return lua_tostring(L, -1); \
		} \
		lua_getfield(L, -1, "generate"); \
		if (lua_pcall(L, 0, 1, 0) != 0) { \
			return lua_tostring(L, -1); \
		} \
		if (!lua_isstring(L, -1)) { \
			return "empty ABI string"; \
		} \
	} while(0)

const char *vm_compile(lua_State *L, const char *code, const char *byte, const char *abi) {
	FILE *f = NULL;

	if (luaL_loadfile(L, code) != 0) {
		return lua_tostring(L, -1);
	}
	f = fopen(byte, "wb");
	if (f == NULL) {
		return "cannot open a bytecode file";
	}
	if (lua_dump(L, kpt_lua_Writer, f) != 0) {
		fclose(f);
		return lua_tostring(L, -1);
	}
	fclose(f);

	if (abi != NULL && strlen(abi) > 0) {
		f = fopen(abi, "wb");
		if (f == NULL) {
			return "cannot open a abi file";
		}
		GEN_ABI();
		fwrite((char *)lua_tostring(L, -1), 1, lua_strlen(L, -1), f);
		fclose(f);
	}

	return NULL;
}

static int writer_buf(lua_State *L, const void *p, size_t size, void *b) {
	luaL_addlstring((luaL_Buffer *)b, (const char *)p, size);
	return 0;
}

const char *vm_loadfile(lua_State *L, const char *code) {
	if (luaL_loadfile(L, code) != 0) {
		return lua_tostring(L, -1);
	}
	return NULL;
}

const char *vm_loadstring(lua_State *L, const char *code) {
	// use loadbufferx with mode="t" to limit input to text (no bytecode)
	if (luaL_loadbufferx(L, code, strlen(code), "code", "t") != 0) {
		return lua_tostring(L, -1);
	}
	return NULL;
}

const char *vm_stringdump(lua_State *L) {
	luaL_Buffer b;

	luaL_buffinit(L, &b);
	if (lua_dump(L, writer_buf, &b) != 0) {
		return lua_tostring(L, -1);
	}
	luaL_pushresult(&b);    /* code dump */
	if (!lua_isstring(L, -1)) {
		return "empty bytecode";
	}
	lua_pushvalue(L, -2);   /* code dump code */
	GEN_ABI();          /* code dump code abi */
	lua_remove(L, -2);  /* code dump abi */
	return NULL;
}

