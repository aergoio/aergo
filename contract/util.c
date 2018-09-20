#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <stdbool.h>
#include <ctype.h>
#include "util.h"
#include "vm.h"

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

void lua_util_dump_json (lua_State *L, int idx, sbuff_t *sbuf)
{
	int len;
	char *src_val;

	switch (lua_type(L, idx)) {
	case LUA_TNUMBER: {
		char tmp[128];
		len = sprintf (tmp, "%g,", lua_tonumber(L, idx));
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
		copy_to_buffer ("{", 1, sbuf);
		orig_bidx = (sbuf->idx);
		if (table_idx < 0)
			table_idx = lua_gettop(L) + idx + 1;
		lua_pushnil(L);
		while (lua_next(L, table_idx) != 0) {
			lua_util_dump_json (L, -2, sbuf);
			--(sbuf->idx);
			copy_to_buffer (":", 1, sbuf);
			lua_util_dump_json (L, -1, sbuf);
			lua_pop(L, 1);
		}
		if (orig_bidx != sbuf->idx) 
			--(sbuf->idx);
		src_val = "},";
		break;
	}
	default:
		src_val = (char *)lua_typename (L, lua_type(L, idx));
		copy_to_buffer ("\"unsupport type:", 16, sbuf);
		copy_to_buffer (src_val, strlen (src_val), sbuf);
		src_val = "\",";

		break;
	}

	len = strlen (src_val);
	copy_to_buffer (src_val, len, sbuf);
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

char *lua_util_get_json_from_ret (lua_State *L, int nresult, sbuff_t *sbuf)
{
	int i;

	int nr = lua_gettop(L);
	copy_to_buffer ("[", 1, sbuf);
	for (i = nr - nresult; i < nr; ++i) {
		lua_util_dump_json (L, i + 1, sbuf);
	}
	lua_pop(L, nresult);
	if (sbuf->idx != 1)
		--sbuf->idx;
	copy_to_buffer ("]", 2, sbuf);

	return sbuf->buf;
}

char *lua_util_get_json_from_args (lua_State *L, int start)
{
	int i;
	sbuff_t sbuf;
	lua_util_sbuf_init (&sbuf, 64);

	int argc = lua_gettop(L);
	copy_to_buffer ("[", 1, &sbuf);
	for (i = start; i <= argc; ++i) {
		lua_util_dump_json (L, i, &sbuf);
	}
	lua_pop(L, argc - start);
	if (sbuf.idx != 1)
		--sbuf.idx;
	copy_to_buffer ("]", 2, &sbuf);

	return sbuf.buf;
}

char *lua_util_get_json (lua_State *L, int idx)
{
	sbuff_t sbuf;
	lua_util_sbuf_init (&sbuf, 64);

	lua_util_dump_json (L, idx, &sbuf);
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


