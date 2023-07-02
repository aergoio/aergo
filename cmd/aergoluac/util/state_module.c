/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

#include <stdint.h>
#include <string.h>
#include <lualib.h>
#include <lauxlib.h>
#include <luajit.h>

#define TYPE_NAME "_type_"
#define TYPE_LEN  "_len_"
#define TYPE_DIMENSION  "_dimension_"

#define STATE_MAX_DIMENSION 5

static int state_map(lua_State *L) {
	int dimension = 1;

	if (luaL_isinteger(L, 1)) {
		dimension = luaL_checkint(L, 1);       /* m _type_ map dim*/
	} else if (lua_gettop(L) != 0) {
		luaL_typerror(L, 1, "integer");
	}

	if (dimension > STATE_MAX_DIMENSION) {
		luaL_error(L, "dimension over max limit(%d): %d, state.map",
		           STATE_MAX_DIMENSION, dimension);
	}
	lua_newtable(L);
	lua_pushstring(L, TYPE_NAME);   /* m _type_ */
	lua_pushstring(L, "map");       /* m _type_ map */
	lua_rawset(L, -3);
	lua_pushstring(L, TYPE_DIMENSION);       /* m _dimension_ */
	lua_pushinteger(L, dimension);       /* m _type_ map dim*/
	lua_rawset(L, -3);
	return 1;
}

static int state_array(lua_State *L) {
	int32_t len = 0;
	int argn = lua_gettop(L);

	if (argn > STATE_MAX_DIMENSION) {
		luaL_error(L, "dimension over max limit(%d): %d, state.array",
		           STATE_MAX_DIMENSION, argn);
	}
	lua_newtable(L);
	lua_pushstring(L, TYPE_NAME);   /* m _type_ */
	lua_pushstring(L, "array");     /* m _type_ array */
	lua_rawset(L, -3);
	lua_pushstring(L, TYPE_LEN);    /* m _len_ */
	if (argn == 0) {
		lua_pushinteger(L, len);        /* m _len_ len*/
	} else if (argn == 1) {
		if (!luaL_isinteger(L, 1)) {
			luaL_typerror(L, 1, "integer");
		}
		len = luaL_checkint(L, 1);      /* size */
		luaL_argcheck(L, (len > 0), 1, "the array length must be greater than zero");
		lua_pushinteger(L, len);        /* m _len_ len*/
	} else {
		int i;
		lua_pushstring(L, TYPE_DIMENSION);   /* m _len_ _dimension_ */
		for (i = 1; i <= argn; ++i) {
			if (!luaL_isinteger(L, i)) {
				luaL_typerror(L, i, "integer");
			}
			len = luaL_checkint(L, i);      /* size */
			luaL_argcheck(L, (len > 0), 1, "the array length must be greater than zero");
			lua_pushinteger(L, len);
			lua_pushstring(L, ",");
		}
		lua_pop(L, 1);
		lua_concat(L, argn * 2 - 1);
		lua_rawset(L, -4);      /* m _len_ */
		lua_pushinteger(L, 0);  /* m _len_ len*/
	}
	lua_rawset(L, -3);
	return 1;
}

static int state_value(lua_State *L) {
	lua_newtable(L);
	lua_pushstring(L, TYPE_NAME);   /* m _type_ */
	lua_pushstring(L, "value");     /* m _type_ value */
	lua_rawset(L, -3);
	return 1;
}

static int state_var(lua_State *L) {
	const char *var_name;
	int t;

	luaL_checktype(L, 1, LUA_TTABLE);           /* T */
	lua_pushnil(L);                             /* T nil ; push the first key */
	while (lua_next(L, -2) != 0) {              /* T key value */
		var_name = luaL_checkstring(L, -2);
		t = lua_type(L, -1);

		luaL_checktype(L, -1, LUA_TTABLE);
		lua_pushstring(L, "id");                /* T key value id */
		lua_pushvalue(L, -3);                   /* T key value id key */
		lua_rawset(L, -3);                      /* T key value{id=key} */

		lua_pushstring(L, TYPE_NAME);           /* T key value _type_ */
		lua_rawget(L, -2);                      /* T key value "type_name" */
		if (lua_isnil(L, -1)) {
			lua_pushfstring(L, "bad argument " LUA_QL("%s") ": state.value, state.map or state.array expected, got %s",
			                var_name, lua_typename(L, t));
			lua_error(L);
		}
		lua_getglobal(L, "abi");                /* T key value "type_name" m */
		lua_getfield(L, -1, "register_var");    /* T key value "type_name" m f */
		lua_pushstring(L, var_name);            /* T key value "type_name" m f var_name */
		lua_pushvalue(L, -5);                   /* T key value "type_name" m f var_name VT */
		lua_call(L, 2, 0);                      /* T key value "type_name" m */
		lua_pop(L, 2);                          /* T key value */
		lua_setglobal(L, var_name);             /* T key */
	}
	return 0;
}

int luac_open_state(lua_State *L) {
	static const luaL_Reg state_lib[] = {
		{"map", state_map},
		{"array", state_array},
		{"value", state_value},
		{"var", state_var},
		{NULL, NULL}
	};

	luaL_register(L, "state", state_lib);

	return 1;
}
