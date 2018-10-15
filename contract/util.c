#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <ctype.h>
#include "util.h"
#include "vm.h"
#include "math.h"

typedef struct tcall {
    void **ptrs;
    int curidx;
    int size;
} callinfo_t;

void lua_util_sbuf_init(sbuff_t *sbuf, int len)
{
	sbuf->idx = 0;
	sbuf->buf_len = len;
	sbuf->buf = malloc (len);
}

static void copy_to_buffer(char *src, int len, sbuff_t *sbuf)
{
	int orig_buf_len = sbuf->buf_len;
	while (len + sbuf->idx >= sbuf->buf_len) {
		sbuf->buf_len *= 2; 
	}
	if (sbuf->buf_len != orig_buf_len) {
		sbuf->buf = realloc (sbuf->buf, sbuf->buf_len);
	}
	memcpy (sbuf->buf + sbuf->idx, src, len);
	sbuf->idx += len;
}

static callinfo_t *callinfo_new()
{
    callinfo_t *callinfo = malloc(sizeof(callinfo_t));
    callinfo->size = 4;
    callinfo->ptrs = malloc(sizeof(void *) * callinfo->size);
    callinfo->curidx = 0;

    return callinfo;
}

static void callinfo_del(callinfo_t *callinfo)
{
    free (callinfo->ptrs);
    free (callinfo);
}

static bool register_tcall(callinfo_t *callinfo, void *ptr)
{
    int i;

    for(i = 0; i < callinfo->curidx; i++) {
        if (callinfo->ptrs[i] == ptr)
            return false;
    }
    if (callinfo->curidx == callinfo->size) {
        callinfo->size *= 2;
        callinfo->ptrs = realloc(callinfo->ptrs, sizeof(void *) * callinfo->size);
    }
    callinfo->ptrs[callinfo->curidx++] = ptr;
    return true;
}

static void unregister_tcall(callinfo_t *callinfo)
{
    callinfo->curidx--;
}

static bool lua_util_dump_json (lua_State *L, int idx, sbuff_t *sbuf, bool json_form, bool iskey,
                         callinfo_t *callinfo)
{
	int len;
	char *src_val;
	char tmp[128];

	switch (lua_type(L, idx)) {
	case LUA_TNUMBER: {
	    if (json_form && iskey) {
	    	if (luaL_isinteger(L, idx))
                len = sprintf (tmp, "\"%ld\",", lua_tointeger(L, idx));
            else
                len = sprintf (tmp, "\"%g\",", lua_tonumber(L, idx));
	    }
	    else {
	    	if (luaL_isinteger(L, idx))
                len = sprintf (tmp, "%ld,", lua_tointeger(L, idx));
            else
                len = sprintf (tmp, "%g,", lua_tonumber(L, idx));
	    }
		src_val = tmp;
		break;
	}
	case LUA_TBOOLEAN: {
		if (lua_toboolean(L, idx))
			src_val = "true,";
		else
			src_val = "false,";
		break;
	}
	case LUA_TNIL:
	    if (json_form)
	        src_val = "{},";
	    else
		    src_val = "null,";
		break;
	case LUA_TSTRING: {
		src_val = (char *)lua_tostring(L, idx);
		copy_to_buffer ("\"", 1, sbuf);
		copy_to_buffer (src_val, strlen (src_val), sbuf);
		src_val = "\",";
		break;
	}
	case LUA_TTABLE: {
		int orig_bidx;
		int table_idx = idx;
		int tbl_len;
		int key_idx;
		bool tcall_make = (callinfo == NULL);

		bool is_array = false;
        if (tcall_make) {
            callinfo = callinfo_new();
        }
        if (!register_tcall(callinfo, (void *)lua_topointer(L, idx))) {
            callinfo_del(callinfo);
            lua_pushstring(L, "nested table error");
            return false;
        }

		if (table_idx < 0)
			table_idx = lua_gettop(L) + idx + 1;
		tbl_len = lua_objlen(L, table_idx);
		if (json_form && tbl_len > 0) {
		    double number;
            char *check_array = calloc(tbl_len, sizeof(char));
		    is_array = true;
            lua_pushnil(L);
            while (lua_next(L, table_idx) != 0) {
                lua_pop (L ,1);
                if (!lua_isnumber(L, -1) || lua_tonumber(L, -1) != round(lua_tonumber(L, -1))) {
                    is_array = false;
                    lua_pop (L ,1);
                    break;
                }
                key_idx = lua_tointeger(L, -1) - 1;
                if (key_idx >= tbl_len || key_idx < 0) {
                    is_array = false;
                    lua_pop (L ,1);
                    break;
                }
                check_array[key_idx] = 1;
            }
            if (is_array) {
                for (key_idx = 0; key_idx < tbl_len; ++key_idx) {
                    if (check_array[key_idx] != 1) {
                        is_array = false;
                        break;
                    }
                }
            }
        }

	    if (is_array) {
	        copy_to_buffer ("[", 1, sbuf);
	        for (key_idx = 1; key_idx <= tbl_len; ++ key_idx) {
	            lua_rawgeti(L, table_idx, key_idx);
			    if (!lua_util_dump_json (L, -1, sbuf, true, false, callinfo))
			        return false;
			    lua_pop(L, 1);
	        }
			--(sbuf->idx);
	        src_val = "],";
	    }
	    else {
            copy_to_buffer ("{", 1, sbuf);
            orig_bidx = (sbuf->idx);
            lua_pushnil(L);
            while (lua_next(L, table_idx) != 0) {
                if (!lua_util_dump_json (L, -2, sbuf, json_form, true, callinfo))
                    return false;
                --(sbuf->idx);
                copy_to_buffer (":", 1, sbuf);
                if (!lua_util_dump_json (L, -1, sbuf, json_form, false, callinfo))
                    return false;
                lua_pop(L, 1);
            }
            if (orig_bidx != sbuf->idx)
                --(sbuf->idx);
            src_val = "},";
        }
		if (tcall_make)
		    callinfo_del(callinfo);
		else
		    unregister_tcall(callinfo);
		break;
	}
	default:
	    lua_pushstring(L, "unsupport type:");
	    lua_pushstring(L, lua_typename (L, lua_type(L, idx)));
	    lua_concat(L, 2);
	    return false;
	}

	len = strlen (src_val);
	copy_to_buffer (src_val, len, sbuf);
	return true;

}

static int json_to_lua_table(lua_State *L, char *json) {
	char *token_end;
	bool end = false;

	if (*json != '{')
		return -1;

	lua_newtable(L);
	json++;
	while(*json != '}' && !end) {
	   token_end = strchr(json, ':');
	   *token_end = '\0';
	   if (lua_util_json_to_lua (L, json) != 0)
		   return -1;
	   json = token_end + 1;
	   token_end = json;
	   while(*token_end != ',' && *token_end != '}') ++token_end;
	   if (*token_end == '}')
			end = true;
	   *token_end = '\0';
	   if (lua_util_json_to_lua (L, json) != 0)
		   return -1;
	   lua_rawset(L, -3);
	   json = token_end + 1;
	}
	return 0;
}

int lua_util_json_to_lua (lua_State *L, char *json)
{
	if (*json == '"') {
		int len = strlen(json);
		json[len - 1] = '\0';
		lua_pushstring(L, json + 1);
	} else if (isdigit(*json)) {
		double d;
		sscanf(json, "%lf", &d);
		lua_pushnumber(L, d);
	} else if (*json == '{') {
		return json_to_lua_table(L, json);
	} else if (strcmp(json, "true") == 0) {
		lua_pushboolean(L, 1);
	} else if (strcmp(json, "false") == 0) {
		lua_pushboolean(L, 0);
	} else if (strcmp(json, "null") == 0) {
		lua_pushnil(L);
	} else {
		return -1;
	}
	return 0;
}

char *lua_util_get_json_from_stack (lua_State *L, int start, int end, bool json_form)
{
	int i;
	sbuff_t sbuf;
	int start_idx;
	lua_util_sbuf_init (&sbuf, 64);

    if (!json_form || start < end)
	    copy_to_buffer ("[", 1, &sbuf);
	start_idx = sbuf.idx;
	for (i = start; i <= end; ++i) {
		if (!lua_util_dump_json (L, i, &sbuf, json_form, false, NULL)) {
		    free(sbuf.buf);
		    return NULL;
		}
	}
	if (sbuf.idx != start_idx)
	    sbuf.idx--;
	if (!json_form || start < end) {
	    copy_to_buffer ("]", 2, &sbuf);
	}
	else {
		sbuf.buf[sbuf.idx] = '\0';
    }

	return sbuf.buf;
}

char *lua_util_get_json (lua_State *L, int idx, bool json_form)
{
	sbuff_t sbuf;
	lua_util_sbuf_init (&sbuf, 64);

	if(!lua_util_dump_json (L, idx, &sbuf, json_form, false, NULL)) {
	    free(sbuf.buf);
	    return NULL;
	}

	if (sbuf.idx != 0)
		sbuf.buf[sbuf.idx - 1] = '\0';

	return sbuf.buf;
}

char *lua_util_get_db_key(const bc_ctx_t *bc_ctx, const char *key)
{

	char *dbKey = malloc(sizeof(char) * (strlen(key) + strlen(bc_ctx->contractId) + 2));
	sprintf(dbKey, "%s_%s", bc_ctx->contractId, key);

	return dbKey;
}


