/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

#include <lualib.h>
#include <lauxlib.h>
#include <luajit.h>

#define TYPE_NAME "_type_"

static int state_map(lua_State *L)
{
    lua_newtable(L);
    lua_pushstring(L, TYPE_NAME);    /* m _type_ */
    lua_pushstring(L, "map");       /* m _type_ map */
    lua_rawset(L, -3);
    return 1;
}

static int state_array(lua_State *L)
{
    luaL_checkint(L, 1);
    lua_newtable(L);
    lua_pushstring(L, TYPE_NAME);    /* m _type_ */
    lua_pushstring(L, "array");     /* m _type_ array */
    lua_rawset(L, -3);
    return 1;
}

static int state_value(lua_State *L)
{
    lua_newtable(L);
    lua_pushstring(L, TYPE_NAME);    /* m _type_ */
    lua_pushstring(L, "value");     /* m _type_ value */
    lua_rawset(L, -3);
    return 1;
}

static int state_var(lua_State *L)
{
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
            lua_pushfstring(L, "bad argument " LUA_QL("%s") ": state_value, state_map or state_array expected, got %s",
                            var_name, lua_typename(L, t));
            lua_error(L);
        }
        lua_getglobal(L, "abi");                /* T key value "type_name" m f */
        lua_getfield(L, -1, "register_var");    /* T key value "type_name" m f */
        lua_pushstring(L, var_name);            /* T key value "type_name" m f var_name */
        lua_pushvalue(L, -4);                   /* T key value "type_name" m f var_name "type_name" */
        lua_call(L, 2, 0);                      /* T key value "type_name" m */
        lua_pop(L, 2);                          /* T key value */
        lua_setglobal(L, var_name);             /* T key */
    }
    return 0;
}

int luac_open_state(lua_State *L)
{
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
