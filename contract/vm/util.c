#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <ctype.h>
#include <stdint.h>
#include "util.h"
#include "vm.h"
#include "math.h"
#include "bignum_module.h"

typedef struct sbuff {
	char *buf;
	int idx;
	int buf_len;
} sbuff_t;

typedef struct tcall {
	void **ptrs;
	int curidx;
	int size;
} callinfo_t;

typedef struct sort_key {
	char *elem;
	int start_idx;
	int key_len;
} sort_key_t;

static void lua_util_sbuf_init(sbuff_t *sbuf, int len) {
	sbuf->idx = 0;
	sbuf->buf_len = len;
	sbuf->buf = malloc(len);
}

static void copy_to_buffer(char *src, int len, sbuff_t *sbuf) {
	int orig_buf_len = sbuf->buf_len;
	while (len + sbuf->idx >= sbuf->buf_len) {
		sbuf->buf_len *= 2;
	}
	if (sbuf->buf_len != orig_buf_len) {
		sbuf->buf = realloc(sbuf->buf, sbuf->buf_len);
	}
	memcpy(sbuf->buf + sbuf->idx, src, len);
	sbuf->idx += len;
}

static void add_escape(sbuff_t *sbuf, char ch) {
	if (sbuf->idx + 2 >= sbuf->buf_len) {
		sbuf->buf_len *= 2;
		sbuf->buf = realloc(sbuf->buf, sbuf->buf_len);
	}
	sbuf->buf[sbuf->idx++] = '\\';
	sbuf->buf[sbuf->idx++] = ch;
}

static void copy_str_to_buffer(char *src, int len, sbuff_t *sbuf) {
	int i;

	char *end = src + len;

	for (; src < end; ++src) {
		if (*src >= 0x00 && *src <= 0x1f) {
			if (sbuf->idx + 6 >= sbuf->buf_len) {
				sbuf->buf_len = sbuf->buf_len * 2 + 6;
				sbuf->buf = realloc(sbuf->buf, sbuf->buf_len);
			}
			sprintf(sbuf->buf + sbuf->idx, "\\u00%02x", *src);
			sbuf->idx = sbuf->idx + 6;
			continue;
		}
		switch(*src) {
		case '"':
		case '\\':
			add_escape(sbuf, *src);
			break;
		case '\t':
			add_escape(sbuf, 't');
			break;
		case '\n':
			add_escape(sbuf, 'n');
			break;
		case '\b':
			add_escape(sbuf, 'b');
			break;
		case '\f':
			add_escape(sbuf, 'f');
			break;
		case '\r':
			add_escape(sbuf, 'r');
			break;
		default:
			if (sbuf->idx + 1 >= sbuf->buf_len) {
				sbuf->buf_len *= 2;
				sbuf->buf = realloc(sbuf->buf, sbuf->buf_len);
			}
			sbuf->buf[sbuf->idx++] = *src;
		}
	}
}
static callinfo_t *callinfo_new() {
	callinfo_t *callinfo = malloc(sizeof(callinfo_t));
	callinfo->size = 4;
	callinfo->ptrs = malloc(sizeof(void *) * callinfo->size);
	callinfo->curidx = 0;

	return callinfo;
}

static void callinfo_del(callinfo_t *callinfo) {
	if (callinfo == NULL) {
		return;
	}
	free(callinfo->ptrs);
	free(callinfo);
}

static bool register_tcall(callinfo_t *callinfo, void *ptr) {
	int i;

	for(i = 0; i < callinfo->curidx; i++) {
		if (callinfo->ptrs[i] == ptr) {
			return false;
		}
	}
	if (callinfo->curidx == callinfo->size) {
		callinfo->size *= 2;
		callinfo->ptrs = realloc(callinfo->ptrs, sizeof(void *) * callinfo->size);
	}
	callinfo->ptrs[callinfo->curidx++] = ptr;
	return true;
}

static void unregister_tcall(callinfo_t *callinfo) {
	callinfo->curidx--;
}

static int sort_key_compare(const void *first, const void*second) {
	sort_key_t *key1 = (sort_key_t *)first, *key2 = (sort_key_t *)second;
	if (key1->key_len == key2->key_len) {
		return strcmp(key1->elem, key2->elem);
	}
	else {
		int comp_len = (key1->key_len > key2->key_len ? key2->key_len : key1->key_len);
		int ret = strncmp(key1->elem, key2->elem, comp_len);
		if (ret == 0) {
			return (key1->key_len > key2->key_len ? 1 : -1);
		}
		return ret;
	}
}

char *bignum_str = "{\"_bignum\":\"";

static bool lua_util_dump_json(lua_State *L, int idx, sbuff_t *sbuf, bool json_form, bool iskey,
                               callinfo_t **pcallinfo) {
	int len;
	char *src_val;
	char tmp[128];

	lua_gasuse(L, GAS_MID);

	switch (lua_type(L, idx)) {
	case LUA_TNUMBER: {
		if (json_form && iskey) {
			if (luaL_isinteger(L, idx)) {
				len = sprintf(tmp, "\"%ld\",", lua_tointeger(L, idx));
			}
			else {
				double d = lua_tonumber(L, idx);
				if (isinf(d) || isnan(d)) {
					lua_pushstring(L, "not support nan or infinity");
					return false;
				}
				len = sprintf(tmp, "\"%.14g\",", d);
			}
		}
		else {
			if (luaL_isinteger(L, idx)) {
				len = sprintf(tmp, "%ld,", lua_tointeger(L, idx));
			}
			else {
				double d = lua_tonumber(L, idx);
				if (isinf(d) || isnan(d)) {
					lua_pushstring(L, "not support nan or infinity");
					return false;
				}
				len = sprintf(tmp, "%.14g,", lua_tonumber(L, idx));
			}
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
		if (json_form && !vm_is_hardfork(L, 2))
			src_val = "{},";
		else
			src_val = "null,";
		break;
	case LUA_TSTRING: {
		size_t len;
		src_val = (char *)lua_tolstring(L, idx, &len);
		copy_to_buffer("\"", 1, sbuf);
		copy_str_to_buffer(src_val, len, sbuf);
		src_val = "\",";
		break;
	}
	case LUA_TTABLE: {
		int orig_bidx;
		int table_idx = idx;
		int tbl_len;
		int key_idx;
		bool is_array = false;
		callinfo_t *callinfo = *pcallinfo;
		if (callinfo == NULL) {
			callinfo = callinfo_new();
			*pcallinfo = callinfo;
		}
		if (!register_tcall(callinfo, (void *)lua_topointer(L, idx))) {
			lua_pushstring(L, "nested table error");
			return false;
		}

		if (table_idx < 0) {
			table_idx = lua_gettop(L) + idx + 1;
		}
		tbl_len = lua_objlen(L, table_idx);
		if ((json_form || vm_is_hardfork(L, 2)) && tbl_len > 0) {
			double number;
			char *check_array = calloc(tbl_len, sizeof(char));
			is_array = true;
			lua_pushnil(L);
			while (lua_next(L, table_idx) != 0) {
				lua_pop(L,1);
				if (!lua_isnumber(L, -1) || lua_tonumber(L, -1) != round(lua_tonumber(L, -1))) {
					is_array = false;
					lua_pop(L,1);
					break;
				}
				key_idx = lua_tointeger(L, -1) - 1;
				if (key_idx >= tbl_len || key_idx < 0) {
					is_array = false;
					lua_pop(L,1);
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
			free(check_array);
		}

		if (is_array) {
			copy_to_buffer("[", 1, sbuf);
			for (key_idx = 1; key_idx <= tbl_len; ++key_idx) {
				lua_rawgeti(L, table_idx, key_idx);
				if (!lua_util_dump_json(L, -1, sbuf, true, false, pcallinfo)) {
					return false;
				}
				lua_pop(L, 1);
			}
			--(sbuf->idx);
			src_val = "],";
		}
		else {
			sbuff_t sort_buf;
			int idx = 0, max_key = 5;
			int i;
			sort_key_t *sort_keys = malloc(sizeof(sort_key_t) * max_key);
			lua_util_sbuf_init(&sort_buf, 20);

			copy_to_buffer("{", 1, sbuf);
			orig_bidx = (sbuf->idx);
			lua_pushnil(L);
			while (lua_next(L, table_idx) != 0) {
				if (idx == max_key) {
					max_key *= 2;
					sort_keys = realloc(sort_keys, sizeof(sort_key_t) * max_key);
				}
				sort_keys[idx].start_idx = sort_buf.idx;
				if (!lua_util_dump_json(L, -2, &sort_buf, json_form, true, pcallinfo)) {
					free(sort_keys);
					free(sort_buf.buf);
					return false;
				}
				sort_keys[idx].key_len = sort_buf.idx - sort_keys[idx].start_idx - 1;
				sort_buf.buf[sort_buf.idx - 1]=':';
				if (!lua_util_dump_json(L, -1, &sort_buf, json_form, false, pcallinfo)) {
					free(sort_keys);
					free(sort_buf.buf);
					return false;
				}
				sort_buf.buf[sort_buf.idx - 1] = '\0';
				lua_pop(L, 1);
				++idx;

			}
			if (idx > 0) {
				for (i = 0; i < idx; ++i) {
					sort_keys[i].elem = sort_buf.buf + sort_keys[i].start_idx;
				}
				qsort(sort_keys, idx, sizeof(sort_key_t), sort_key_compare);
				for (i = 0; i < idx; ++i) {
					copy_to_buffer(sort_keys[i].elem, strlen(sort_keys[i].elem), sbuf);
					copy_to_buffer(",", 1, sbuf);
				}
			}
			free(sort_keys);
			free(sort_buf.buf);
			if (orig_bidx != sbuf->idx) {
				--(sbuf->idx);
			}
			src_val = "},";
		}
		unregister_tcall(callinfo);
		break;
	}
	case LUA_TUSERDATA: {
		if (lua_isbignumber(L, idx)) {
			char *s;
			copy_to_buffer(bignum_str,strlen(bignum_str), sbuf);
			s = lua_get_bignum_str(L, idx);
			if (s != NULL) {
				copy_str_to_buffer(s, strlen(s), sbuf);
				free(s);
			}
			src_val = "\"},";
			break;
		}
	}
	default:
		lua_pushfstring(L, "\"unsupport type: %s\"", lua_typename(L, lua_type(L, idx)));
		return false;
	}

	len = strlen(src_val);
	copy_to_buffer(src_val, len, sbuf);
	return true;

}

static int json_to_lua(lua_State *L, char **start, bool check, bool is_bignum);

// the input is like an array but without the [] characters
static int json_args_to_lua(lua_State *L, char *json, bool check) {
	int count = 0;
	while(*json != '\0') {
		if (json_to_lua(L, &json, check, false) != 0) {
			return -1;
		}
		if (*json == ',') {
			++json;
		} else if(*json != '\0') {
			return -1;
		}
		++count;
	}
	return count;
}

int lua_util_json_array_to_lua(lua_State *L, char *json, bool check) {
	int count = 0;
	if (*json != '[') {
		return -1;
	}
	++json;
	while(*json != ']') {
		if (json_to_lua(L, &json, check, false) != 0) {
			return -1;
		}
		if (*json == ',') {
			++json;
		} else if(*json != ']') {
			return -1;
		}
		++count;
	}
	return count;
}

static int json_array_to_lua_table(lua_State *L, char **start, bool check) {
	char *json = (*start) + 1;
	int index = 1;

	lua_newtable(L);
	while(*json != ']') {
		lua_pushnumber(L, index++);
		if (json_to_lua(L, &json, check, false) != 0) {
			return -1;
		}
		if (*json == ',') {
			++json;
		} else if(*json != ']') {
			return -1;
		}
		lua_rawset(L, -3);
	}
	*start = json + 1;
	return 0;
}

static int json_to_lua_table(lua_State *L, char **start, bool check) {
	char *json = (*start) + 1;
	int index = 1;
	bool is_bignum = false;
	int elem_cnt = 0;

	lua_newtable(L);
	while(*json != '}') {
		lua_pushnumber(L, index++);
		if (json_to_lua(L, &json, check, false) != 0) {
			return -1;
		}
		if (*json == ':') {
			lua_remove(L, -2);
			--index;
			if (check && !lua_isstring(L, -1)) {
				return -1;
			}
			if (elem_cnt == 0 && lua_type(L, -1) == LUA_TSTRING &&
			    strcmp(lua_tostring(L, -1), "_bignum") == 0) {
				is_bignum = true;
			}
			++json;
			if (json_to_lua(L, &json, check, is_bignum) != 0) {
				return -1;
			}
		}
		if (*json == ',') {
			if (is_bignum) {
				return -1;
			}
			++json;
		} else if (*json != '}') {
			return -1;
		}

		if (is_bignum) {
			if (!lua_isbignumber(L, -1)) {
				return -1;
			}
			lua_replace(L, -3);
			lua_pop(L, 1);
		} else {
			lua_rawset(L, -3);
		}
	}
	*start = json + 1;
	return 0;
}


int lua_util_utf8_encode(char *s, unsigned ch) {
	if (ch < 0x80) {
		s[0] = (char)ch;
		return 1;
	}
	if (ch <= 0x7FF) {
		s[1] = (char) ((ch | 0x80) & 0xBF);
		s[0] = (char) ((ch >> 6) | 0xC0);
		return 2;
	}
	if (ch <= 0xFFFF) {
		s[2] = (char) ((ch | 0x80) & 0xBF);
		s[1] = (char) (((ch >> 6) | 0x80) & 0xBF);
		s[0] = (char) ((ch >> 12) | 0xE0);
		return 3;
	}
	{
		char buff[UTF8_MAX];
		unsigned mfb = 0x3F;
		int n = 1;
		do {
			if (n > 6) {
				return -1;
			}
			buff[UTF8_MAX - (n++)] = 0x80 | (ch&0x3F);
			ch >>= 6;
			mfb >>= 1;
		} while (ch > mfb);
		buff[UTF8_MAX - n] = (~mfb << 1) | ch;
		memcpy(s, &buff[UTF8_MAX - n], n);
		return n;
	}
}

static int json_to_lua(lua_State *L, char **start, bool check, bool is_bignum) {
	char *json = *start;
	char special[5];

	lua_gasuse(L, GAS_MID);

	special[4] = '\0';
	while(isspace(*json)) ++json;
	if (*json == '"') {
		if (is_bignum) {
			char *end = strchr(json + 1, '"');
			if (end != NULL) {
				*end = '\0';
				lua_set_bignum(L, json + 1);
				*end = '"';
				json = end + 1;
			} else {
				return -1;
			}
		} else {
			char *end = json + 1;
			char *target = end;
			while ((*end) != '"') {
				if (*end == '\0') {
					return -1;
				}
				if ((*end) == '\\') {
					end++;
					switch(*end) {
					case 't':
						*target = '\t';
						break;
					case 'n':
						*target = '\n';
						break;
					case 'b':
						*target = '\b';
						break;
					case 'f':
						*target = '\f';
						break;
					case 'r':
						*target = '\r';
						break;
					case 'u': {
						int i;
						unsigned ch;
						int out;
						for (i = 1; i < 5; ++i) {
							if (!isxdigit(*(end + i))) {
								return -1;
							}
						}
						memcpy(special, end+1, 4);
						ch = strtol(special, NULL, 16);
						out = lua_util_utf8_encode(target, ch);
						if (out < 0) {
							return -1;
						}
						target = target + out - 1;
						end += 4;
						break;
					}
					default:
						*target = *end;
					}
				} else if (end != target) {
					*target = *end;
				}
				end++;
				target++;
			}
			*target = '\0';
			lua_pushlstring(L, json + 1, target - json - 1);
			json = end + 1;
		}
	} else if (isdigit(*json) || *json == '-' || *json == '+') {
		double d;
		char *end = json + 1;

		if (is_bignum) {
			return -1;
		}
		while(*end != '\0') {
			if (!isdigit(*end) && *end != '-' && *end != '.' &&
			    *end != 'e' && *end != 'E' && *end != '+') {
				break;
			}
			++end;
		}
		sscanf(json, "%lf", &d);
		if (vm_is_hardfork(L, 2) && d == (int64_t)d) {
			lua_pushinteger(L, (int64_t)d);
		} else {
			lua_pushnumber(L, d);
		}
		json = end;
	} else if (*json == '{') {
		if (json_to_lua_table(L, &json, check) != 0) {
			return -1;
		}
	} else if (*json == '[') {
		if (json_array_to_lua_table(L, &json, check) != 0) {
			return -1;
		}
	} else if (strncasecmp(json, "true", 4) == 0) {
		lua_pushboolean(L, 1);
		json = json + 4;
	} else if (strncasecmp(json, "false", 5) == 0) {
		lua_pushboolean(L, 0);
		json = json + 5;
	} else if (strncasecmp(json, "null", 4) == 0) {
		lua_pushnil(L);
		json = json + 4;
	} else {
		return -1;
	}
	while(isspace(*json)) ++json;
	*start = json;
	return 0;
}

void minus_inst_count(lua_State *L, int count) {
	if (!lua_usegas(L)) {
		int cnt = vm_instcount(L);
		cnt -= count;
		if (cnt <= 0) {
			cnt = 1;
		}
		vm_setinstcount(L, cnt);
	}
}

int lua_util_json_value_to_lua(lua_State *L, char *json, bool check) {
	if (json_to_lua(L, &json, check, false) != 0) {
		return -1;
	}
	if (check && *json != '\0') {
		return -1;
	}
	return 0;
}

char *lua_util_get_json_from_stack(lua_State *L, int start, int end, bool json_form) {
	int i;
	sbuff_t sbuf;
	int start_idx;
	callinfo_t *callinfo = NULL;
	lua_util_sbuf_init(&sbuf, 64);

	if (!json_form || start < end) {
		copy_to_buffer("[", 1, &sbuf);
	}
	start_idx = sbuf.idx;
	for (i = start; i <= end; ++i) {
		if (!lua_util_dump_json(L, i, &sbuf, json_form, false, &callinfo)) {
			callinfo_del(callinfo);
			free(sbuf.buf);
			return NULL;
		}
	}
	callinfo_del(callinfo);
	if (sbuf.idx != start_idx) {
		sbuf.idx--;
	}
	if (!json_form || start < end) {
		copy_to_buffer("]", 2, &sbuf);
	} else {
		sbuf.buf[sbuf.idx] = '\0';
	}

	minus_inst_count(L, strlen(sbuf.buf));
	return sbuf.buf;
}

char *lua_util_get_json_array_from_stack(lua_State *L, int start, int end, bool json_form) {
	int i;
	sbuff_t sbuf;
	int start_idx;
	callinfo_t *callinfo = NULL;
	lua_util_sbuf_init(&sbuf, 64);

	copy_to_buffer("[", 1, &sbuf);
	start_idx = sbuf.idx;
	for (i = start; i <= end; ++i) {
		if (!lua_util_dump_json(L, i, &sbuf, json_form, false, &callinfo)) {
			callinfo_del(callinfo);
			free(sbuf.buf);
			return NULL;
		}
	}
	callinfo_del(callinfo);
	if (sbuf.idx != start_idx) {
		sbuf.idx--;
	}
	copy_to_buffer("]", 2, &sbuf);

	minus_inst_count(L, strlen(sbuf.buf));
	return sbuf.buf;
}

char *lua_util_get_json(lua_State *L, int idx, bool json_form) {
	sbuff_t sbuf;
	callinfo_t *callinfo = NULL;
	lua_util_sbuf_init(&sbuf, 64);

	if(!lua_util_dump_json(L, idx, &sbuf, json_form, false, &callinfo)) {
		callinfo_del(callinfo);
		free(sbuf.buf);
		return NULL;
	}
	callinfo_del(callinfo);

	if (sbuf.idx != 0) {
		sbuf.buf[sbuf.idx - 1] = '\0';
	}

	minus_inst_count(L, strlen(sbuf.buf));
	return sbuf.buf;
}

static int lua_json_encode(lua_State *L) {
	char *json;

	lua_gasuse(L, 50);
	json = lua_util_get_json(L, -1, true);
	if (json == NULL) {
		luaL_throwerror(L);
	}
	lua_pushstring(L, json);
	free(json);
	return 1;
}

static int lua_json_decode(lua_State *L) {
	char *org = (char *)luaL_checkstring(L, -1);
	char *json = strdup(org);

	lua_gasuse(L, 50);
	minus_inst_count(L, strlen(json));
	if (lua_util_json_value_to_lua(L, json, true) != 0) {
		free(json);
		luaL_error(L, "not proper json format");
	}
	free(json);
	return 1;
}

static const luaL_Reg json_lib[] = {
	{"encode", lua_json_encode},
	{"decode", lua_json_decode},
	{NULL, NULL}
};

int luaopen_json(lua_State *L) {
	luaL_register(L, "json", json_lib);
	lua_pop(L, 1);
	return 1;
}
