/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

#include "vm.h"
#include "system_module.h"

#define STATE_MAP_ID "__state_map__"
#define STATE_VALUE_ID "__state_value__"

static int state_map(lua_State *L)
{
    lua_newtable(L);                        /* T */
    luaL_getmetatable(L, STATE_MAP_ID);     /* T map_mt */
    lua_setmetatable(L, -2);                /* T{mt=map_mt} */
    return 1;
}

static int state_map_get(lua_State *L)
{
    luaL_checktype(L, 1, LUA_TTABLE);
    luaL_checkstring(L, 2);
    lua_pushcfunction(L, getItem);                  /* t f */
    lua_getfield(L, 1, "id");                       /* t f id */
    if (!lua_isstring(L, -1)) {
        luaL_error(L, "the value is not a state.map type");
    }
    lua_pushvalue(L, 2);                            /* t f id key */
    lua_concat(L, 2);                               /* t f id+key */
    lua_call(L, 1, 1);                              /* t rv */
    return 1;
}

static int state_map_set(lua_State *L)
{
    luaL_checktype(L, 1, LUA_TTABLE);
    luaL_checkstring(L, 2);
    luaL_checkany(L, 3);
    lua_pushcfunction(L, setItem);                  /* t f */
    lua_getfield(L, 1, "id");                       /* t f id */
    if (!lua_isstring(L, -1)) {
        luaL_error(L, "the value is not a state.map type");
    }
    lua_pushvalue(L, 2);                            /* t f id key */
    lua_concat(L, 2);                               /* t f id+key */
    lua_pushvalue(L, 3);                            /* t f id+key value */
    lua_call(L, 2, 0);                              /* t */
    return 0;
}

static int state_value(lua_State *L)
{
    lua_newtable(L);                            /* T */
    luaL_getmetatable(L, STATE_VALUE_ID);       /* T value_mt */
    lua_setmetatable(L, -2);                    /* T{mt=value_mt} */
    return 1;
}

static int state_value_get(lua_State *L)
{
    luaL_checktype(L, 1, LUA_TTABLE);               /* t */
    lua_pushcfunction(L, getItem);                  /* t f */
    lua_getfield(L, 1, "id");                       /* t f id */
    if (!lua_isstring(L, -1)) {
        luaL_error(L, "the value is not a state.value type");
    }
    lua_call(L, 1, 1);                              /* t system rv */
    return 1;
}

static int state_value_set(lua_State *L)
{
    luaL_checktype(L, 1, LUA_TTABLE);               /* t */
    luaL_checkany(L, 2);
    lua_pushcfunction(L, setItem);                  /* t f */
    lua_getfield(L, 1, "id");                       /* t f id */
    if (!lua_isstring(L, -1)) {
        luaL_error(L, "the value is not a state.value type");
    }
    lua_pushvalue(L, 2);                            /* t f id value */
    lua_call(L, 2, 0);                              /* t rv */
    return 0;
}

static int state_var(lua_State *L)
{
    luaL_checktype(L, 1, LUA_TTABLE);               /* T */
    lua_pushnil(L);                                 /* T nil ; push the first key */
    while (lua_next(L, -2) != 0) {                  /* T key value */
        luaL_checkstring(L, -2);
        luaL_checktype(L, -1, LUA_TTABLE);
        lua_pushstring(L, "id");                    /* T key value id */
        lua_pushvalue(L, -3);                       /* T key value id key */
        lua_rawset(L, -3);                          /* T key value{id=key} */
        lua_setglobal(L, lua_tostring(L, -2));      /* T key */
    }
    return 0;
}

int luaopen_state(lua_State *L)
{
    static const luaL_Reg state_map_metas[] = {
        {"__index",  state_map_get},
        {"__newindex", state_map_set},
        {NULL, NULL}
    };

    static const luaL_Reg state_value_methods[] = {
        {"get", state_value_get},
        {"set", state_value_set},
        {NULL, NULL}
    };

    static const luaL_Reg state_lib[] = {
        {"map", state_map},
        {"value", state_value},
        {"var", state_var},
        {NULL, NULL}
    };

    luaL_newmetatable(L, STATE_MAP_ID);
    luaL_register(L, NULL, state_map_metas);

    luaL_newmetatable(L, STATE_VALUE_ID);
    lua_pushvalue(L, -1);
    lua_setfield(L, -2, "__index");
    luaL_register(L, NULL, state_value_methods);

    luaL_register(L, "state", state_lib);

    return 1;
}
