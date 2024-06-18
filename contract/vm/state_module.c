/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

#include <string.h>
#include <stdlib.h>
#include <stdint.h>
#include "vm.h"
#include "system_module.h"

#define STATE_MAP_ID            "__state_map__"
#define STATE_ARRAY_ID          "__state_array__"
#define STATE_VALUE_ID          "__state_value__"

#define STATE_VAR_KEY_PREFIX    "_sv_"
#define STATE_VAR_META_LEN      "_sv_meta-len_"
#define STATE_VAR_META_TYPE     "_sv_meta-type_"

#define STATE_MAX_DIMENSION 5

static int state_map_delete(lua_State *L);
static int state_array_append(lua_State *L);
static int state_array_pairs(lua_State *L);

/*
** If this is a sub-element, update the index at position 2 with
** the parent's element key.
*/
static void update_key_with_parent(lua_State *L, char *parent_key) {
	if (parent_key != NULL) {
		// concatenate the key at this level with the given index
		lua_pushstring(L, parent_key);
		lua_pushstring(L, "-");
		lua_pushvalue(L, 2);
		lua_concat(L, 3);
		// update the index at position 2
		lua_replace(L, 2);
	}
}

/* map */

typedef struct {
	char *id;
	int key_type;
	int dimension;
	char *key;
} state_map_t;

static int state_map(lua_State *L) {
	int nargs = lua_gettop(L);

	state_map_t *m = lua_newuserdata(L, sizeof(state_map_t)); /* m */
	m->id = NULL;
	m->key_type = LUA_TNONE;
	m->key = NULL;

	if (luaL_isinteger(L, 1))
		m->dimension = luaL_checkint(L, 1);
	else if (nargs == 0)
		m->dimension = 1;
	else
		luaL_typerror(L, 1, "integer");

	if (m->dimension > STATE_MAX_DIMENSION) {
		luaL_error(L, "dimension over max limit(%d): %d, state.map",
		           STATE_MAX_DIMENSION, m->dimension);
	}

	luaL_getmetatable(L, STATE_MAP_ID);                     /* m mt */
	lua_setmetatable(L, -2);                                /* m */
	return 1;
}

/*
** Once a state map is accessed with one type, it can only
** be used with that type (string or number).
** Probably to make it more compatible with Lua tables,
** where m[1] ~= m['1']
*/
static void state_map_check_index_type(lua_State *L, state_map_t *m) {
	// get the type used for the key on the current access
	int key_type = lua_type(L, 2);             /* m key */
	// get the type used on the given map
	int stored_type = m->key_type;
	// maps can be accessed with only string and number
	if (key_type != LUA_TNUMBER && key_type != LUA_TSTRING) {
		luaL_error(L, "invalid key type: " LUA_QS ", state.map: " LUA_QS,
		           lua_typename(L, key_type), m->id);
	}
	// if the key type for this map is not loaded yet...
	if (stored_type == LUA_TNONE) {
		// load it from state
		lua_pushcfunction(L, getItemWithPrefix); /* m key f */
		lua_pushstring(L, m->id);                /* m key f id */
		lua_pushstring(L, STATE_VAR_META_TYPE);  /* m key f id prefix */
		lua_call(L, 2, 1);                       /* m key t */
		// is any type stored for this map?
		if (!lua_isnil(L, -1)) {
			// if yes, check the type
			stored_type = luaL_checkint(L, -1);
			if (stored_type != LUA_TNUMBER && stored_type != LUA_TSTRING) {
				luaL_error(L, "invalid stored key type: " LUA_QS ", state.map:",
				           lua_typename(L, stored_type), m->id);
			}
		}
		// store the type on the map
		if (vm_is_hardfork(L, 2)) {
			m->key_type = stored_type;
		}
		// remove it from stack
		lua_pop(L, 1);
	}
	// make sure the map is accessed with the same type
	if (stored_type != LUA_TNONE && key_type != stored_type) {
		luaL_typerror(L, 2, lua_typename(L, stored_type));
	}
}

static void state_map_push_key(lua_State *L, state_map_t *m) {
	lua_pushstring(L, m->id);                   /* m key value f id */
	lua_pushstring(L, "-");                     /* m key value f id '-' */
	lua_pushvalue(L, 2);                        /* m key value f id '-' key */
	lua_concat(L, 3);                           /* m key value f id-key */
}

static int state_map_get(lua_State *L) {
	int key_type = LUA_TNONE;
	int nargs = lua_gettop(L);  /* map key [blockheight] */
	state_map_t *m = luaL_checkudata(L, 1, STATE_MAP_ID);

	key_type = lua_type(L, 2);
	if (key_type == LUA_TSTRING) {
		const char *method = lua_tostring(L, 2);
		if (method != NULL && strcmp(method, "delete") == 0) {
			lua_pushcfunction(L, state_map_delete);
			return 1;
		}
	}

	state_map_check_index_type(L, m);

	if (m->dimension > 1) {
		state_map_t *subm = lua_newuserdata(L, sizeof(state_map_t)); /* m */
		subm->id = strdup(m->id);
		subm->key_type = m->key_type;
		subm->dimension = m->dimension - 1;

		luaL_getmetatable(L, STATE_MAP_ID);  /* m mt */
		lua_setmetatable(L, -2);             /* m */

		if (m->key == NULL) {                /* m key */
			subm->key = strdup(lua_tostring(L, 2));
		} else {
			lua_pushstring(L, m->key);   /* m key id */
			lua_pushstring(L, "-");      /* m key id '-' */
			lua_pushvalue(L, 2);         /* m key id '-' key */
			lua_concat(L, 3);            /* m key id-key */
			subm->key = strdup(lua_tostring(L, -1));
			lua_pop(L, 1);               /* m key */
		}
		return 1;
	}

	// read the value associated with the given key
	lua_pushcfunction(L, getItemWithPrefix);    /* m key f */
	// if this is a sub-map, update the index with the parent's map key
	update_key_with_parent(L, m->key);
	state_map_push_key(L, m);                   /* m key f id-key */
	if (nargs == 3) {                           /* m key h f id-key */
		lua_pushvalue(L, 3);                /* m key h f id-key h */
	}
	lua_pushstring(L, STATE_VAR_KEY_PREFIX);    /* m key f id-key prefix */
	lua_call(L, nargs, 1);                      /* m key value */
	return 1;
}

static int state_map_set(lua_State *L) {
	/* m key value */
	int key_type = LUA_TNONE;
	state_map_t *m = luaL_checkudata(L, 1, STATE_MAP_ID);

	key_type = lua_type(L, 2);

	if (m->dimension > 1) {
		luaL_error(L, "not permitted to set intermediate dimension of map");
	}
	if (key_type == LUA_TSTRING) {
		const char *method = lua_tostring(L, 2);
		if (method != NULL && strcmp(method, "delete") == 0) {
			luaL_error(L, "can't use " LUA_QL("delete") " as a key");
		}
	}

	state_map_check_index_type(L, m);

	// store the data type used for keys on this map
	if (m->key_type == LUA_TNONE) {
		lua_pushcfunction(L, setItemWithPrefix);  /* m key value f */
		lua_pushstring(L, m->id);                 /* m key value f id */
		lua_pushinteger(L, key_type);             /* m key value f id type */
		lua_pushstring(L, STATE_VAR_META_TYPE);   /* m key value f id type prefix */
		lua_call(L, 3, 0);                        /* m key value */
		m->key_type = key_type;
	}

	luaL_checkany(L, 3);

	// save the value with the given key
	lua_pushcfunction(L, setItemWithPrefix);    /* m key value f */
	// if this is a sub-map, update the index with the parent's map key
	update_key_with_parent(L, m->key);
	state_map_push_key(L, m);                   /* m key value f id-key */
	lua_pushvalue(L, 3);                        /* m key value f id-key value */
	lua_pushstring(L, STATE_VAR_KEY_PREFIX);    /* m key value f id-key value prefix */
	lua_call(L, 3, 0);                          /* t key value */
	return 0;
}

static int state_map_delete(lua_State *L) {
	/* m key */
	state_map_t *m = luaL_checkudata(L, 1, STATE_MAP_ID);

	if (m->dimension > 1) {
		luaL_error(L, "not permitted to set intermediate dimension of map");
	}

	state_map_check_index_type(L, m);

	// delete the item with the given key
	lua_pushcfunction(L, delItemWithPrefix);    /* m key f */
	// if this is a sub-map, update the index with the parent's map key
	update_key_with_parent(L, m->key);
	state_map_push_key(L, m);                   /* m key f id-key */
	lua_pushstring(L, STATE_VAR_KEY_PREFIX);    /* m key f id-key prefix */
	lua_call(L, 2, 1);                          /* m key rv */
	return 0;
}

static int state_map_gc(lua_State *L) {
	state_map_t *m = luaL_checkudata(L, 1, STATE_MAP_ID);
	if (m->id) {
		free(m->id);
		m->id = NULL;
	}
	if (m->key) {
		free(m->key);
		m->key = NULL;
	}
	return 0;
}

/* array */

typedef struct {
	char *id;
	int is_fixed;
	int dimension;
	int32_t *lens;
	char *key;
} state_array_t;

static int state_array(lua_State *L) {
	int is_fixed;
	state_array_t *arr;
	int dimension = lua_gettop(L);
	int32_t *lens = NULL;

	is_fixed = dimension != 0;

	if (dimension > STATE_MAX_DIMENSION) {
		luaL_error(L, "dimension over max limit(%d): %d, state.array",
		           STATE_MAX_DIMENSION, dimension);
	}

	if (is_fixed) {
		int i;
		lens = malloc(sizeof(int32_t) * dimension);
		for (i = 1; i <= dimension; i++) {
			if (!luaL_isinteger(L, i)) {
				luaL_typerror(L, i, "integer");
			}
			lens[i -1] = luaL_checkint(L, i);          /* size */
			luaL_argcheck(L, (lens[i - 1] > 0), i, "the array length must be greater than zero");
		}
	}

	arr = lua_newuserdata(L, sizeof(state_array_t)); /* size a */
	luaL_getmetatable(L, STATE_ARRAY_ID);           /* size a mt */
	lua_setmetatable(L, -2);                        /* size a */
	arr->lens = lens;
	arr->id = NULL;
	arr->dimension = dimension;
	arr->is_fixed = is_fixed;
	arr->key = NULL;
	return 1;
}

static int state_array_len(lua_State *L) {
	state_array_t *arr = luaL_checkudata(L, 1, STATE_ARRAY_ID);
	if (arr->lens == NULL)
		lua_pushinteger(L, 0);
	else
		lua_pushinteger(L, arr->lens[0]);
	return 1;
}

static void state_array_load_len(lua_State *L, state_array_t *arr) {
	if (!arr->is_fixed && arr->lens == NULL) {
		lua_pushcfunction(L, getItemWithPrefix); /* a i f */
		lua_pushstring(L, arr->id);              /* a i f id */
		lua_pushstring(L, STATE_VAR_META_LEN);   /* a i f id prefix */
		lua_call(L, 2, 1);                       /* a i n */
		arr->lens = malloc(sizeof(int32_t));
		arr->lens[0] = luaL_optinteger(L, -1, 0);
		lua_pop(L, 1);
	}
}

static void state_array_checkarg(lua_State *L, state_array_t *arr) {
	int idx;
	if (!luaL_isinteger(L, 2)) {
		luaL_typerror(L, 2, "integer");
	}
	idx = luaL_checkint(L, 2);
	luaL_argcheck(L, idx >= 1 && idx <= arr->lens[0], 2, "index out of range");
}

static void state_array_push_key(lua_State *L, const char *id) {
	lua_pushstring(L, id);  /* a i value f id */
	lua_pushstring(L, "-"); /* a i value f id '-' */
	lua_pushvalue(L, 2);    /* a i value f id '-' key */
	lua_concat(L, 3);       /* a i value f id-key */
}

static int state_array_get(lua_State *L) {
	state_array_t *arr;
	int nargs = lua_gettop(L);    /* arr index [blockheight] */
	int key_type = LUA_TNONE;

	arr = luaL_checkudata(L, 1, STATE_ARRAY_ID);
	state_array_load_len(L, arr);

	if (lua_type(L, 2) == LUA_TSTRING) {        /* methods */
		const char *method = lua_tostring(L, 2);
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
		luaL_typerror(L, 2, "integer");
	}

	if (arr->dimension > 1) {
		state_array_t *suba = lua_newuserdata(L, sizeof(state_array_t)); /* m */
		suba->id = strdup(arr->id);
		suba->dimension = arr->dimension - 1;
		suba->lens = malloc(sizeof(int32_t) * suba->dimension);
		suba->is_fixed = arr->is_fixed;
		memcpy(suba->lens, arr->lens + 1, sizeof(int32_t) * suba->dimension);

		luaL_getmetatable(L, STATE_ARRAY_ID);                 /* m mt */
		lua_setmetatable(L, -2);                              /* m */

		if (arr->key == NULL) {              /* a i */
			suba->key = strdup(lua_tostring(L, 2));
		} else {
			lua_pushstring(L, arr->key); /* a i key */
			lua_pushstring(L, "-");      /* a i key '-' */
			lua_pushvalue(L, 2);         /* a i key '-' i */
			lua_concat(L, 3);            /* a i key-i */
			suba->key = strdup(lua_tostring(L, -1));
			lua_pop(L, 1);               /* a i */
		}
		return 1;
	}

	if (nargs == 3) {
		lua_pushvalue(L, 2);  //? why?
	}
	state_array_checkarg(L, arr);               /* a i */

	// read the value at the given position
	lua_pushcfunction(L, getItemWithPrefix);    /* a i f */
	// if this is a sub-array, update the index with the parent's array key
	update_key_with_parent(L, arr->key);
	state_array_push_key(L, arr->id);           /* a i [h] f id-i */
	if (nargs == 3) {                           /* a i h f id-i */
		lua_pushvalue(L, 3);                /* a i h f id-i h */
	}
	lua_pushstring(L, STATE_VAR_KEY_PREFIX);    /* a i [h] f id-i [h] prefix */
	lua_call(L, nargs, 1);                      /* a i value */
	return 1;
}

static int state_array_set(lua_State *L) {
	state_array_t *arr;

	arr = luaL_checkudata(L, 1, STATE_ARRAY_ID);

	if (arr->dimension > 1) {
		luaL_error(L, "not permitted to set intermediate dimension of array");
	}
	state_array_load_len(L, arr);

	state_array_checkarg(L, arr);               /* a i v */

	// if this is a sub-array, update the index with the parent's array key
	update_key_with_parent(L, arr->key);

	// save the value at the given position
	lua_pushcfunction(L, setItemWithPrefix);    /* a i v f */
	state_array_push_key(L, arr->id);           /* a i v f id-i */
	lua_pushvalue(L, 3);                        /* a i v f id-i v */
	lua_pushstring(L, STATE_VAR_KEY_PREFIX);    /* a i v f id-i v prefix */
	lua_call(L, 3, 0);                          /* a i v */

	return 0;
}

static int state_array_append(lua_State *L) {
	// get reference to the array
	state_array_t *arr = luaL_checkudata(L, 1, STATE_ARRAY_ID);
	// ensure there is a value of any type
	luaL_checkany(L, 2);                        /* a v */
	// append() function only allowed on variable size arrays
	if (arr->is_fixed) {
		luaL_error(L, "the fixed array cannot use " LUA_QL("append") " method");
	}
	// check for overflow in size
	if (arr->lens[0] + 1 <= 0) {
		luaL_error(L, "state.array " LUA_QS " overflow", arr->id);
	}
	// increment the size of the array
	arr->lens[0]++;
	// save the value at the given position
	lua_pushcfunction(L, state_array_set);      /* a v f */
	lua_pushvalue(L, 1);                        /* a v f a */
	lua_pushinteger(L, arr->lens[0]);           /* a v f a i */
	lua_pushvalue(L, 2);                        /* a v f a i v */
	lua_call(L, 3, 0);
	// save the new array size on the state
	lua_pushcfunction(L, setItemWithPrefix);    /* a v f */
	lua_pushstring(L, arr->id);                 /* a v f id */
	lua_pushinteger(L, arr->lens[0]);           /* a v f id size */
	lua_pushstring(L, STATE_VAR_META_LEN);      /* a v f id size prefix */
	lua_call(L, 3, 0);
	// nothing to return
	return 0;
}

static int state_array_gc(lua_State *L) {
	state_array_t *arr = luaL_checkudata(L, 1, STATE_ARRAY_ID);
	if (arr->id) {
		free(arr->id);
		arr->id = NULL;
	}
	if (arr->lens) {
		free(arr->lens);
		arr->lens = NULL;
	}
	if (arr->key) {
		free(arr->key);
		arr->key = NULL;
	}
	return 0;
}

static int state_array_iter(lua_State *L) {
	state_array_t *arr = luaL_checkudata(L, 1, STATE_ARRAY_ID);
	int i = luaL_checkint(L, 2);
	i = i + 1;
	if (i <= arr->lens[0]) {
		lua_pushinteger(L, i);
		lua_pushcfunction(L, state_array_get);
		lua_pushvalue(L, 1);
		lua_pushinteger(L, i);
		lua_call(L, 2, 1);
		return 2;
	}
	return 0;
}

static int state_array_pairs(lua_State *L) {
	luaL_checkudata(L, 1, STATE_ARRAY_ID);
	lua_pushcfunction(L, state_array_iter);
	lua_pushvalue(L, 1);
	lua_pushinteger(L, 0);
	return 3;
}

/* scalar value */

typedef struct {
	char *id;
} state_value_t;

static int state_value(lua_State *L) {
	state_value_t *v = lua_newuserdata(L, sizeof(state_value_t)); /* v */
	v->id = NULL;
	luaL_getmetatable(L, STATE_VALUE_ID);                       /* v mt */
	lua_setmetatable(L, -2);                                    /* v */
	return 1;
}

static int state_value_get(lua_State *L) {
	state_value_t *v = luaL_checkudata(L, 1, STATE_VALUE_ID);
	lua_pushcfunction(L, getItemWithPrefix);    /* v f */
	lua_pushstring(L, v->id);                   /* v f id */
	lua_pushstring(L, STATE_VAR_KEY_PREFIX);    /* v f id prefix */
	lua_call(L, 2, 1);                          /* v rv */
	return 1;
}

static int state_value_snapget(lua_State *L) {
	int nargs = lua_gettop(L);
	state_value_t *v = luaL_checkudata(L, 1, STATE_VALUE_ID); /* v */
	lua_pushcfunction(L, getItemWithPrefix);                /* v f */
	lua_pushstring(L, v->id);                               /* v f id */
	if (nargs == 2) {
		lua_pushvalue(L, 2);
	}
	lua_pushstring(L, STATE_VAR_KEY_PREFIX);                /* v f id prefix */
	lua_call(L, nargs + 1, 1);                              /* v rv */
	return 1;
}

static int state_value_set(lua_State *L) {
	state_value_t *v = luaL_checkudata(L, 1, STATE_VALUE_ID);
	luaL_checkany(L, 2);                        /* s v */
	lua_pushcfunction(L, setItemWithPrefix);    /* s v f */
	if (v->id == NULL) {
		luaL_error(L, "invalid state.value: (nil)");
	}
	lua_pushstring(L, v->id);                   /* s v f id */
	lua_pushvalue(L, 2);                        /* s v f id value */
	lua_pushstring(L, STATE_VAR_KEY_PREFIX);    /* s v f id value prefix */
	lua_call(L, 3, 0);                          /* s v */
	return 0;
}

static int state_value_gc(lua_State *L) {
	state_value_t *v = luaL_checkudata(L, 1, STATE_VALUE_ID);
	if (v->id) {
		free(v->id);
		v->id = NULL;
	}
	return 0;
}

/* global variable */

static void insert_var(lua_State *L, const char *var_name) {
	lua_getglobal(L, "abi");                    /* "type_name" m */
	lua_getfield(L, -1, "register_var");        /* "type_name" m f */
	lua_pushstring(L, var_name);                /* "type_name" m f var_name */
	lua_pushvalue(L, -4);                       /* "type_name" m f var_name "type_name" */
	lua_call(L, 2, 0);                          /* "type_name" m */
	lua_pop(L, 2);
}

static int state_var(lua_State *L) {
	const char *var_name;
	state_map_t *m = NULL;
	state_array_t *arr = NULL;
	state_value_t *v = NULL;

	luaL_checktype(L, 1, LUA_TTABLE);               /* T */
	lua_pushnil(L);                                 /* T nil ; push the first key */
	while (lua_next(L, -2) != 0) {                  /* T key value */
		var_name = luaL_checkstring(L, -2);
		if (!lua_isuserdata(L, -1)) {
			lua_pushfstring(L, "bad argument " LUA_QS ": state.value, state.map or state.array expected, got %s",
			                var_name, lua_typename(L, lua_type(L, -1)));
			lua_error(L);
		}

		m = luaL_testudata(L, -1, STATE_MAP_ID);
		if (m != NULL) {
			m->id = strdup(var_name);
			goto found;
		}

		arr = luaL_testudata(L, -1, STATE_ARRAY_ID);
		if (arr != NULL) {
			arr->id = strdup(var_name);     /* T key value */
			goto found;
		}

		v = luaL_testudata(L, -1, STATE_VALUE_ID);
		if (v != NULL) {
			v->id = strdup(var_name);
		} else {
			lua_pushfstring(L, "bad argument " LUA_QS ": state.value, state.map or state.array expected", var_name);
			lua_error(L);
		}

found:
		lua_newtable(L);
		insert_var(L, var_name);
		lua_setglobal(L, var_name);         /* T key */
	}
	return 0;
}

static int state_get_snap(lua_State *L) {
	state_map_t *m = NULL;
	state_array_t *arr = NULL;
	state_value_t *v = NULL;

	if (!lua_isuserdata(L, 1)) {
		luaL_typerror(L, 1, "state.value, state.map or state.array");
	}

	m = luaL_testudata(L, 1, STATE_MAP_ID);
	if (m != NULL) {
		if (lua_gettop(L) != 3) {
			luaL_error(L, "invalid argument at getsnap, need (state.map, key, blockheight)");
		}
		return state_map_get(L);
	}

	arr = luaL_testudata(L, 1, STATE_ARRAY_ID);
	if (arr != NULL) {
		if (lua_gettop(L) != 3) {
			luaL_error(L, "invalid argument at getsnap, need (state.array, index, blockheight)");
		}
		return state_array_get(L);
	}

	v = luaL_testudata(L, 1, STATE_VALUE_ID);
	if (v != NULL) {
		if (lua_gettop(L) != 2) {
			luaL_error(L, "invalid argument at getsnap, need (state.value, blockheight)");
		}
		return state_value_snapget(L);
	}

	luaL_typerror(L, 1, "state.value, state.map or state.array");
	return 0;
}

int luaopen_state(lua_State *L) {
	static const luaL_Reg state_map_metas[] = {
		{"__index",  state_map_get},
		{"__newindex", state_map_set},
		{"__gc", state_map_gc},
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
		{"getsnap", state_get_snap},
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
	lua_pushcfunction(L, state_value_gc);
	lua_setfield(L, -2, "__gc");

	luaL_register(L, "state", state_lib);

	lua_pop(L, 1);
	return 1;
}
