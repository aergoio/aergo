#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "util.h"
#include "_cgo_export.h"

extern void checkLuaExecContext(lua_State *L);

static int resolve(lua_State *L) {
	struct luaNameResolve_return ret;
	char *name;

	checkLuaExecContext(L);
	lua_gasuse(L, 100);

	name = (char *)luaL_checkstring(L, 1);

	ret = luaNameResolve(L, name);
	if (ret.r1 != NULL) {
		strPushAndRelease(L, ret.r1);
		luaL_throwerror(L);
	}
	strPushAndRelease(L, ret.r0);
	return 1;
}

static const luaL_Reg name_service_lib[] = {
	{"resolve", resolve},
	{NULL, NULL}
};

int luaopen_name(lua_State *L) {
	luaL_register(L, "name_service", name_service_lib);
	lua_pop(L, 1);
	return 1;
}
