/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

#include <string.h>
#include <stdlib.h>
#include "vm.h"
#include "system_module.h"

#define STATE_MAP_ID            "__state_map__"
#define STATE_ARRAY_ID          "__state_array__"
#define STATE_VALUE_ID          "__state_value__"

#define STATE_VAR_KEY_PREFIX    "_sv_"
#define STATE_VAR_META_LEN      "_sv_meta_len_"

#define TYPE_NAME               "_type_"

static int state_map_delete(lua_State *L);
static int state_array_append(lua_State *L);
static int state_array_pairs(lua_State *L);

/* map */

static int state_map(lua_State *L)
{
    lua_newtable(L);                                /* m */
    lua_pushstring(L, TYPE_NAME);                   /* m _type_ */
    lua_pushstring(L, "map");                       /* m _type_ map */
    lua_rawset(L, -3);                              /* m */
    lua_pushcfunction(L, state_map_delete);         /* m delete f */
    lua_setfield(L, -2, "delete");                  /* m */
    luaL_getmetatable(L, STATE_MAP_ID);             /* m mt */
    lua_setmetatable(L, -2);                        /* m */
    return 1;
}

static int state_map_check_index(lua_State *L, int index)
{
    int type = lua_type(L, index);
    luaL_argcheck(L, (type == LUA_TNUMBER || type == LUA_TSTRING),
                  index, "number or string expected");
    return type;
}

static void state_map_push_key(lua_State *L, int type)
{
    lua_getfield(L, 1, "id");                       /* m key value f id */
    if (!lua_isstring(L, -1)) {
        luaL_error(L, "the value is not a state.map type");
    }
    lua_pushvalue(L, 2);                            /* m key value f id key */
    if (type == LUA_TNUMBER) {
        lua_pushstring(L, "_n");
    } else {
        lua_pushstring(L, "_s");
    }
    lua_concat(L, 3);                               /* m key value f id.."_n"..key */
}

static int state_map_get(lua_State *L)
{
    int key_type;
    luaL_checktype(L, 1, LUA_TTABLE);               /* m key */
    key_type = state_map_check_index(L, 2);
    lua_pushcfunction(L, getItemWithPrefix);        /* m key f */
    state_map_push_key(L, key_type);                /* m key f id..key_type..key */
    lua_pushstring(L, STATE_VAR_KEY_PREFIX);        /* m key f id..key_type..key prefix */
    lua_call(L, 2, 1);                              /* m key rv */
    return 1;
}

static int state_map_set(lua_State *L)
{
    int key_type;
    luaL_checktype(L, 1, LUA_TTABLE);               /* m key value */
    key_type = state_map_check_index(L, 2);
    luaL_checkany(L, 3);
    lua_pushcfunction(L, setItemWithPrefix);        /* m key value f */
    state_map_push_key(L, key_type);                /* m key value f id..key_type..key */
    lua_pushvalue(L, 3);                            /* m key value f id..key_type..key value */
    lua_pushstring(L, STATE_VAR_KEY_PREFIX);        /* m key value f id..key_type..key value prefix */
    lua_call(L, 3, 0);                              /* t key value */
    return 0;
}

static int state_map_delete(lua_State *L)
{
    int key_type;
    luaL_checktype(L, 1, LUA_TTABLE);               /* m key */
    key_type = state_map_check_index(L, 2);
    lua_pushcfunction(L, delItemWithPrefix);        /* m key f */
    state_map_push_key(L, key_type);                /* m key f id..key_type..key */
    lua_pushstring(L, STATE_VAR_KEY_PREFIX);        /* m key f id..key_type..key prefix */
    lua_call(L, 2, 1);                              /* m key rv */
    return 0;
}

/* array */

typedef struct {
    char *id;
    int len;
    int is_fixed;
} state_array_t;

static int state_array(lua_State *L)
{
    int is_fixed;
    int len = 0;
    state_array_t *arr;

    is_fixed = lua_gettop(L) != 0;
    if (is_fixed) {
        len = luaL_checkint(L, 1);                       /* size */
        luaL_argcheck(L, (len > 0), 1, "the array length must be greater than zero");
    }
    arr = lua_newuserdata(L, sizeof(state_array_t));    /* size a */
    luaL_getmetatable(L, STATE_ARRAY_ID);               /* size a mt */
    lua_setmetatable(L, -2);                            /* size a */
    arr->len = len;
    arr->id = NULL;
    arr->is_fixed = is_fixed;
    return 1;
}

static int state_array_len(lua_State *L)
{
    state_array_t *arr = luaL_checkudata(L, 1, STATE_ARRAY_ID);
    lua_pushinteger(L, arr->len);
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
    const char *method;
    const char *idx;
    state_array_t *arr;

    method = lua_tostring(L, 2);
    if (method != NULL) {                           /* methods */
        if (strcmp(method, "append") == 0) {
            lua_pushcfunction(L, state_array_append);
            return 1;
        } else if (strcmp(method, "ipairs") == 0) {
            lua_pushcfunction(L, state_array_pairs);
            return 1;
        } else if (strcmp(method, "length") == 0) {
            lua_pushcfunction(L, state_array_len);
            return 1;
        }
    }
    arr = state_array_checkarg(L);                  /* a i */
    lua_pushcfunction(L, getItemWithPrefix);        /* a i f */
    lua_pushstring(L, arr->id);                     /* a i f id */
    if (!lua_isstring(L, -1)) {
        luaL_error(L, "the value is not a state.array type");
    }
    idx = lua_tostring(L, 2);                       /* FIXME NULL check ? */
    lua_pushstring(L, idx);                         /* a i f id i */
    lua_concat(L, 2);                               /* a i f idi */
    lua_pushstring(L, STATE_VAR_KEY_PREFIX);        /* a i f idi prefix */
    lua_call(L, 2, 1);                              /* a i rv */
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
    lua_pushcfunction(L, setItemWithPrefix);        /* a i v f */
    lua_pushstring(L, arr->id);                     /* a i v f id */
    if (!lua_isstring(L, -1)) {
        luaL_error(L, "the value is not a state.array type");
    }
    idx = lua_tostring(L, 2);                       /* FIXME NULL check ? */
    lua_pushstring(L, idx);                         /* a i v f id i */
    lua_concat(L, 2);                               /* a i v f idi */
    lua_pushvalue(L, 3);                            /* a i v f idi v */
    lua_pushstring(L, STATE_VAR_KEY_PREFIX);        /* a i v f idi v prefix */
    lua_call(L, 3, 0);                              /* a i v */
    return 0;
}

static int state_array_append(lua_State *L)
{
    state_array_t *arr = luaL_checkudata(L, 1, STATE_ARRAY_ID);
    luaL_checkany(L, 2);
    if (arr->is_fixed) {
        return luaL_error(L, "the fixed array cannot use " LUA_QL("append") " method");
    }
    arr->len++;
    lua_pushcfunction(L, state_array_set);          /* a v f */
    lua_pushvalue(L, 1);                            /* a v f a */
    lua_pushinteger(L, arr->len);                   /* a v f a i */
    lua_pushvalue(L, 2);                            /* a v f a i v */
    lua_call(L, 3, 0);
    lua_pushcfunction(L, setItemWithPrefix);
    lua_pushstring(L, arr->id);
    lua_pushinteger(L, arr->len);
    lua_pushstring(L, STATE_VAR_META_LEN);
    lua_call(L, 3, 0);
    return 0;
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
    lua_pushstring(L, TYPE_NAME);                   /* T _type_ */
    lua_pushstring(L, "value");                     /* T _type_ map */
    lua_rawset(L, -3);                              /* T */
    luaL_getmetatable(L, STATE_VALUE_ID);           /* T mt */
    lua_setmetatable(L, -2);                        /* T */
    return 1;
}

static int state_value_get(lua_State *L)
{
    luaL_checktype(L, 1, LUA_TTABLE);               /* t */
    lua_pushcfunction(L, getItemWithPrefix);        /* t f */
    lua_getfield(L, 1, "id");                       /* t f id */
    if (!lua_isstring(L, -1)) {
        luaL_error(L, "the value is not a state.value type");
    }
    lua_pushstring(L, STATE_VAR_KEY_PREFIX);        /* t f id prefix */
    lua_call(L, 2, 1);                              /* t rv */
    return 1;
}

static int state_value_set(lua_State *L)
{
    luaL_checktype(L, 1, LUA_TTABLE);               /* t */
    luaL_checkany(L, 2);
    lua_pushcfunction(L, setItemWithPrefix);        /* t f */
    lua_getfield(L, 1, "id");                       /* t f id */
    if (!lua_isstring(L, -1)) {
        luaL_error(L, "the value is not a state.value type");
    }
    lua_pushvalue(L, 2);                            /* t f id value */
    lua_pushstring(L, STATE_VAR_KEY_PREFIX);        /* t f id value prefix */
    lua_call(L, 3, 0);                              /* t */
    return 0;
}

/* global variable */

static void insert_var(lua_State *L, const char *var_name)
{
    lua_getglobal(L, "abi");                        /* "type_name" m */
    lua_getfield(L, -1, "register_var");            /* "type_name" m f */
    lua_pushstring(L, var_name);                    /* "type_name" m f var_name */
    lua_pushvalue(L, -4);                           /* "type_name" m f var_name "type_name" */
    lua_call(L, 2, 0);                              /* "type_name" m */
    lua_pop(L, 2);
}

static int state_var(lua_State *L)
{
    int t, i = 1;
    const char *var_name;

    luaL_checktype(L, 1, LUA_TTABLE);                                   /* T */
    lua_pushnil(L);                                                     /* T nil ; push the first key */
    while (lua_next(L, -2) != 0) {                                      /* T key value */
        var_name = luaL_checkstring(L, -2);
        t = lua_type(L, -1);
        if (LUA_TTABLE == t) {
            lua_pushstring(L, "id");                                    /* T key value id */
            lua_pushvalue(L, -3);                                       /* T key value id key */
            lua_rawset(L, -3);                                          /* T key value{id=key} */
            lua_pushstring(L, TYPE_NAME);                               /* T key value _type_ */
            lua_rawget(L, -2);                                          /* T key value "type_name" */
            if (lua_isnil(L, -1)) {
                lua_pushfstring(L, "bad argument " LUA_QL("%s") ": state.value, state.map or state.array expected, got %s",
                        var_name, lua_typename(L, t));
                lua_error(L);
            }
            insert_var(L, var_name);
            lua_setglobal(L, var_name);                                 /* T key */
        } else if (LUA_TUSERDATA == t) {
            state_array_t *arr = luaL_checkudata(L, -1, STATE_ARRAY_ID);
            arr->id = strdup((const char *)lua_tostring(L, -2));        /* T key value */
            lua_pushstring(L, "array");                                 /* T key "type_name" */
            insert_var(L, var_name);
            lua_setglobal(L, var_name);                                 /* T key */
            if (!arr->is_fixed) {
                lua_pushcfunction(L, getItemWithPrefix);
                lua_pushstring(L, arr->id);
                lua_pushstring(L, STATE_VAR_META_LEN);
                lua_call(L, 2, 1);
                arr->len = luaL_optinteger(L, -1, 0);
                lua_pop(L, 1);
            }
        } else {
            lua_pushfstring(L, "bad argument " LUA_QL("%s") ": state.value, state.map or state.array expected, got %s",
                            var_name, lua_typename(L, t));
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
