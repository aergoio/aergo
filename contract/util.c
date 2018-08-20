#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include "util.h"

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

static void dump_json (lua_State *L, int idx, sbuff_t *sbuf) 
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
		src_val = "{},";
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
			dump_json (L, -2, sbuf);
			--(sbuf->idx);
			copy_to_buffer (":", 1, sbuf);
			dump_json (L, -1, sbuf);
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

char *lua_util_get_json_from_ret (lua_State *L, int nresult, sbuff_t *sbuf)
{
	int i;

    int nr = lua_gettop(L);
    int start = nr - nresult;
	copy_to_buffer ("{", 1, sbuf);
	for (i = nr - nresult; i < nr; ++i) {
		dump_json (L, i + 1, sbuf);
	}
	lua_pop(L, nresult);
	if (sbuf->idx != 1)
		--sbuf->idx;
	copy_to_buffer ("}", 1, sbuf);

	return sbuf->buf;
}


