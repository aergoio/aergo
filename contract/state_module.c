/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "system_module.h"

#define STATE_MAP_ID "__state_map__"
#define STATE_ARRAY_ID "__state_array__"
#define STATE_VALUE_ID "__state_value__"

/* map */

static int state_map(lua_State *L)
{
    lua_newtable(L);                                /* m */
    luaL_getmetatable(L, STATE_MAP_ID);             /* m mt */
    lua_setmetatable(L, -2);                        /* m */
    return 1;
}

static int state_map_get(lua_State *L)
{
    luaL_checktype(L, 1, LUA_TTABLE);               /* m key */
    luaL_checkstring(L, 2);
    lua_pushcfunction(L, getItem);                  /* m key f */
    lua_getfield(L, 1, "id");                       /* m key f id */
    if (!lua_isstring(L, -1)) {
        luaL_error(L, "the value is not a state.map type");
    }
    lua_pushvalue(L, 2);                            /* m key f id key */
    lua_concat(L, 2);                               /* m key f idkey */
    lua_call(L, 1, 1);                              /* m key rv */
    return 1;
}

static int state_map_set(lua_State *L)
{
    luaL_checktype(L, 1, LUA_TTABLE);               /* m key value */
    luaL_checkstring(L, 2);
    luaL_checkany(L, 3);
    lua_pushcfunction(L, setItem);                  /* m key value f */
    lua_getfield(L, 1, "id");                       /* m key value f id */
    if (!lua_isstring(L, -1)) {
        luaL_error(L, "the value is not a state.map type");
    }
    lua_pushvalue(L, 2);                            /* m key value f id key */
    lua_concat(L, 2);                               /* m key value f idkey */
    lua_pushvalue(L, 3);                            /* m key value f idkey value */
    lua_call(L, 2, 0);                              /* t key value rv */
    return 0;
}

/* array */

typedef struct {
    char *id;
    int len;
} state_array_t;

static int state_array(lua_State *L)
{
    int len;
    state_array_t *arr;
    len = luaL_checkint(L, 1);                          /* size */
    arr = lua_newuserdata(L, sizeof(state_array_t));    /* size a */
    luaL_getmetatable(L, STATE_ARRAY_ID);               /* size a mt */
    lua_setmetatable(L, -2);                            /* size a */
    arr->len = len;
    arr->id = NULL;
    return 1;
}

static state_array_t *state_array_checkarg(lua_State *L)
{
    int idx = luaL_checkint(L, -1) - 1;             /* a i */
    state_array_t *arr = luaL_checkudata(L, -2, STATE_ARRAY_ID);
    idx = luaL_checkint(L, -1) - 1;                 /* a i */
    luaL_argcheck(L, 0 <= idx && idx < arr->len, 2, "index out of range");
    return arr;
}

static int state_array_get(lua_State *L)
{
    const char *idx;
    state_array_t *arr;

    arr = state_array_checkarg(L);                  /* a i */
    lua_pushcfunction(L, getItem);                  /* a i f */
    lua_pushstring(L, arr->id);                     /* a i f id */
    if (!lua_isstring(L, -1)) {
        luaL_error(L, "the value is not a state.array type");
    }
    idx = lua_tostring(L, 2);                       /* FIXME NULL check ? */
    lua_pushstring(L, idx);                         /* a i f id i */
    lua_concat(L, 2);                               /* a i f idi */
    lua_call(L, 1, 1);                              /* a i rv */
    return 1;
}

static int state_array_set(lua_State *L)
{
    const char *idx;
    state_array_t *arr;

    lua_pushvalue(L, 1);                            /* a i v a */
    lua_pushvalue(L, 2);                            /* a i v a i */
    arr = state_array_checkarg(L);                  /* a i v a i */
    lua_pop(L, 2);                                  /* a i v */
    lua_pushcfunction(L, setItem);                  /* a i v f */
    lua_pushstring(L, arr->id);                     /* a i v f id */
    if (!lua_isstring(L, -1)) {
        luaL_error(L, "the value is not a state.array type");
    }
    idx = lua_tostring(L, 2);                       /* FIXME NULL check ? */
    lua_pushstring(L, idx);                         /* a i v f id i */
    lua_concat(L, 2);                               /* a i v f idi */
    lua_pushvalue(L, 3);                            /* a i v f idi v */
    lua_call(L, 2, 0);                              /* a i v */
    return 0;
}

static int state_array_len(lua_State *L)
{
    state_array_t *arr = luaL_checkudata(L, 1, STATE_ARRAY_ID);
    lua_pushinteger(L, arr->len);
    return 1;
}

static int state_array_gc(lua_State *L)
{
    state_array_t *arr = luaL_checkudata(L, 1, STATE_ARRAY_ID);
    if (arr->id) {
        free(arr->id);
        arr->id = NULL;
    }
    return 0;
}

static int state_array_iter(lua_State *L)
{
    state_array_t *arr = luaL_checkudata(L, 1, STATE_ARRAY_ID);
    int i = luaL_checkint(L, 2);
    i = i + 1;
    if (i <= arr->len) {
        lua_pushinteger(L, i);
        lua_pushcfunction(L, state_array_get);
        lua_pushvalue(L, 1);
        lua_pushinteger(L, i);
        lua_call(L, 2, 1);
        return 2;
    }
    return 0;
}

static int state_array_pairs(lua_State *L)
{
    luaL_checkudata(L, 1, STATE_ARRAY_ID);
    lua_pushcfunction(L, state_array_iter);
    lua_pushvalue(L, 1);
    lua_pushinteger(L, 0);
    return 3;
}

/* scalar value */

static int state_value(lua_State *L)
{
    lua_newtable(L);                                /* T */
    luaL_getmetatable(L, STATE_VALUE_ID);           /* T mt */
    lua_setmetatable(L, -2);                        /* T */
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

/* global variable */

static int state_var(lua_State *L)
{
    int t, i = 1;
    luaL_checktype(L, 1, LUA_TTABLE);                                   /* T */
    lua_pushnil(L);                                                     /* T nil ; push the first key */
    while (lua_next(L, -2) != 0) {                                      /* T key value */
        luaL_checkstring(L, -2);
        t = lua_type(L, -1);
        if (LUA_TTABLE == t) {
            lua_pushstring(L, "id");                                    /* T key value id */
            lua_pushvalue(L, -3);                                       /* T key value id key */
            lua_rawset(L, -3);                                          /* T key value{id=key} */
            lua_setglobal(L, lua_tostring(L, -2));                      /* T key */
        } else if (LUA_TUSERDATA == t) {
            state_array_t *arr = luaL_checkudata(L, -1, STATE_ARRAY_ID);
            arr->id = strdup((const char *)lua_tostring(L, -2));        /* T key value */
            lua_setglobal(L, lua_tostring(L, -2));                      /* T key */
        } else {
            lua_pushfstring(L, "bad argument %d to 'state_var' (state_value, state_map or state_array expected, got %s", i, lua_typename(L, t));
            lua_error(L);
        }
        i++;
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

    static const luaL_Reg state_array_metas[] = {
        {"__index",  state_array_get},
        {"__newindex", state_array_set},
        {"__len", state_array_len},
        {"__gc", state_array_gc},
        {NULL, NULL}
    };

    static const luaL_Reg state_value_methods[] = {
        {"get", state_value_get},
        {"set", state_value_set},
        {NULL, NULL}
    };

    static const luaL_Reg state_lib[] = {
        {"map", state_map},
        {"array", state_array},
        {"value", state_value},
        {"var", state_var},
        {"array_pairs", state_array_pairs},
        {NULL, NULL}
    };

    luaL_newmetatable(L, STATE_MAP_ID);
    luaL_register(L, NULL, state_map_metas);

    luaL_newmetatable(L, STATE_ARRAY_ID);
    luaL_register(L, NULL, state_array_metas);

    luaL_newmetatable(L, STATE_VALUE_ID);
    lua_pushvalue(L, -1);
    lua_setfield(L, -2, "__index");
    luaL_register(L, NULL, state_value_methods);

    luaL_register(L, "state", state_lib);

    return 1;
}
