#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "util.h"
#include "_cgo_export.h"

extern int getLuaExecContext(lua_State *L);

// READ FUNCTIONS ///////////////////////////////////////////////////

static int name_resolve(lua_State *L) {
	char *name, *ret;
	int service = getLuaExecContext(L);

	lua_gasuse(L, 100);

	name = (char *)luaL_checkstring(L, 1);
	ret = luaNameResolve(L, service, name);

	if (ret == NULL) {
		lua_pushnil(L);
	} else {
		strPushAndRelease(L, ret);
		// if the returned string starts with `[`, it's an error
		if (ret[0] == '[') {
			luaL_throwerror(L);
		}
	}

	return 1;
}

static int name_owner(lua_State *L) {
	char *name, *ret;
	int service = getLuaExecContext(L);

	lua_gasuse(L, 100);

	name = (char *)luaL_checkstring(L, 1);
	ret = luaNameOwner(L, service, name);

	if (ret == NULL) {
		lua_pushnil(L);
	} else {
		strPushAndRelease(L, ret);
		// if the returned string starts with `[`, it's an error
		if (ret[0] == '[') {
			luaL_throwerror(L);
		}
	}

	return 1;
}

static int name_operator(lua_State *L) {
	char *name, *ret;
	int service = getLuaExecContext(L);

	lua_gasuse(L, 100);

	name = (char *)luaL_checkstring(L, 1);
	ret = luaNameOperator(L, service, name);

	if (ret == NULL) {
		lua_pushnil(L);
	} else {
		strPushAndRelease(L, ret);
		// if the returned string starts with `[`, it's an error
		if (ret[0] == '[') {
			luaL_throwerror(L);
		}
	}

	return 1;
}

// WRITE FUNCTIONS //////////////////////////////////////////////////

static int name_service_exec(lua_State *L, char type) {
	int service = getLuaExecContext(L);
	char *args, *ret;

	lua_gasuse(L, 500);  // ?????????

	args = lua_util_get_json_from_stack(L, 1, lua_gettop(L), false);
	if (args == NULL) {
		luaL_throwerror(L);
	}
	if (strlen(args) == 0) {
		free(args);
		luaL_error(L, "invalid input");
	}

	ret = luaNameService(L, service, type, args);

	free(args);

	if (ret != NULL) {
		strPushAndRelease(L, ret);
		luaL_throwerror(L);
	}
	return 0;
}

static int name_register(lua_State *L) {
  return name_service_exec(L, 'C');
}

static int name_transfer(lua_State *L) {
  return name_service_exec(L, 'U');
}

static const luaL_Reg name_service_lib[] = {
	{"resolve",  name_resolve},
	{"owner",    name_owner},
	{"operator", name_operator},
	{"register", name_register},
	{"transfer", name_transfer},
	{NULL, NULL}
};

int luaopen_name(lua_State *L) {
	luaL_register(L, "name_service", name_service_lib);
	lua_pop(L, 1);
	return 1;
}
