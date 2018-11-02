/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

#include <lualib.h>
#include <lauxlib.h>
#include <luajit.h>

static int state_map(lua_State *L)
{
    lua_newtable(L);
    return 1;
}

static int state_value(lua_State *L)
{
    lua_newtable(L);
    return 1;
}

static int state_var(lua_State *L)
{
    luaL_checktype(L, 1, LUA_TTABLE);           /* T */
    lua_pushnil(L);                             /* T nil ; push the first key */
    while (lua_next(L, -2) != 0) {              /* T key value */
        luaL_checkstring(L, -2);
        luaL_checktype(L, -1, LUA_TTABLE);
        lua_pushstring(L, "id");                /* T key value id */
        lua_pushvalue(L, -3);                   /* T key value id key */
        lua_rawset(L, -3);                      /* T key value{id=key} */
        lua_setglobal(L, lua_tostring(L, -2));  /* T key */
    }
    return 0;
}

int luaopen_state(lua_State *L)
{
    static const luaL_Reg state_lib[] = {
        {"map", state_map},
        {"value", state_value},
        {"var", state_var},
        {NULL, NULL}
    };

    luaL_register(L, "state", state_lib);

    return 1;
}
