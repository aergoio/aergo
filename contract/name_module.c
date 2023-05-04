#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "util.h"
#include "_cgo_export.h"

extern int getLuaExecContext(lua_State *L);

static int resolve(lua_State *L)
{
	char *name, *ret;
	int service = getLuaExecContext(L);

	lua_gasuse(L, 100);

	name = (char *)luaL_checkstring(L, 1);
	ret = luaResolve(L, service, name);

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

static const luaL_Reg name_service_lib[] = {
	{"resolve", resolve},
	{NULL, NULL}
};

int luaopen_name(lua_State *L)
{
	luaL_register(L, "name_service", name_service_lib);
	lua_pop(L, 1);
	return 1;
}
