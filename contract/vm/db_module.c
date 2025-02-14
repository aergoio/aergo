#include <string.h>
#include <stdlib.h>
#include <ctype.h>
#include <time.h>
#include "vm.h"
//#include "sqlcheck.h"
#include "bignum_module.h"
#include "util.h"
#include "../db_msg.h"
#include "_cgo_export.h"

#define RESOURCE_PSTMT_KEY "_RESOURCE_PSTMT_KEY_"
#define RESOURCE_RS_KEY "_RESOURCE_RS_KEY_"

extern void checkLuaExecContext(lua_State *L);

static int append_resource(lua_State *L, const char *key, void *data) {
	int refno;
	if (luaL_findtable(L, LUA_REGISTRYINDEX, key, 0) != NULL) {
		luaL_error(L, "cannot find the environment of the db module");
	}
	/* tab */
	lua_pushlightuserdata(L, data); /* tab pstmt */
	refno = luaL_ref(L, -2);        /* tab */
	lua_pop(L, 1);                  /* remove tab */
	return refno;
}

#define DB_PSTMT_ID "__db_pstmt__"

typedef struct {
	int id;
	int closed;
	int colcnt;
	int refno;
} db_pstmt_t;

#define DB_RS_ID "__db_rs__"

typedef struct {
	int query_id;
	int closed;
	int nc;
	int refno;
} db_rs_t;


static void send_vm_api_request(lua_State *L, char *method, buffer *args, rresponse *resp) {
	luaSendRequest(L, method, args, resp);
}

static db_rs_t *get_db_rs(lua_State *L, int pos) {
	db_rs_t *rs = luaL_checkudata(L, pos, DB_RS_ID);
	if (rs->closed) {
		luaL_error(L, "resultset is closed");
	}
	return rs;
}

static int db_rs_tostr(lua_State *L) {
	db_rs_t *rs = luaL_checkudata(L, 1, DB_RS_ID);
	if (rs->closed) {
		lua_pushfstring(L, "resultset is closed");
	} else {
		lua_pushfstring(L, "resultset{query_id=%d}", rs->query_id);
	}
	return 1;
}

static int db_rs_get(lua_State *L) {
	buffer buf = {0}, *req = &buf;
	rresponse resp = {0}, *response = &resp;
	db_rs_t *rs = get_db_rs(L, 1);
	int count=0;

	add_int(req, rs->query_id);

	send_vm_api_request(L, "rsGet", req, response);
	free_buffer(req);
	if (response->error) {
		lua_pushfstring(L, "get failed: %s", response->error);
		free_response(response);
		lua_error(L);
	}

	char *ptr = NULL;
	int len;
	while (ptr = get_next_item(&response->result, ptr, &len)) {
		char type = get_type(ptr, len);
		ptr += 1; len -= 1;
		switch (type) {
		case 'b':
			lua_pushboolean(L, read_bool(ptr));
			break;
		case 'i':
			lua_pushinteger(L, read_int(ptr));
			break;
		case 'l':
			lua_pushinteger(L, read_int64(ptr));
			break;
		case 'd':
			lua_pushnumber(L, read_double(ptr));
			break;
		case 's':
			lua_pushlstring(L, ptr, len - 1);
			break;
		case 'n':
			lua_pushnil(L);
			break;
		default:
			lua_pushnil(L);
		}
		count++;
	}

	free_response(response);
	return count;
}

static int db_rs_colcnt(lua_State *L) {
	db_rs_t *rs = get_db_rs(L, 1);
	lua_pushinteger(L, rs->nc);
	return 1;
}

static void db_rs_close(lua_State *L, db_rs_t *rs, int remove) {
	if (rs->closed) {
		return;
	}
	rs->closed = 1;
	if (remove) {
		if (luaL_findtable(L, LUA_REGISTRYINDEX, RESOURCE_RS_KEY, 0) != NULL) {
			luaL_error(L, "cannot find the environment of the db module");
		}
		luaL_unref(L, -1, rs->refno);
		lua_pop(L, 1);
	}
}

static int db_rs_next(lua_State *L) {
	buffer buf = {0}, *req = &buf;
	rresponse resp = {0}, *response = &resp;
	db_rs_t *rs = get_db_rs(L, 1);
	int rc;

	add_int(req, rs->query_id);

	send_vm_api_request(L, "rsNext", req, response);
	free_buffer(req);
	if (response->error) {
		db_rs_close(L, rs, 1);
		lua_pushfstring(L, "next failed: %s", response->error);
		free_response(response);
		lua_error(L);
	}

	bool has_more = get_bool(&response->result, 1);

	if (has_more) {
		lua_pushboolean(L, 1);
	} else {
		db_rs_close(L, rs, 1);
		lua_pushboolean(L, 0);
	}

	free_response(response);
	return 1;
}

static int db_rs_gc(lua_State *L) {
	db_rs_close(L, luaL_checkudata(L, 1, DB_RS_ID), 1);
	return 0;
}

static db_pstmt_t *get_db_pstmt(lua_State *L, int pos) {
	db_pstmt_t *pstmt = luaL_checkudata(L, pos, DB_PSTMT_ID);
	if (pstmt->closed) {
		luaL_error(L, "prepared statement is closed");
	}
	return pstmt;
}

static int db_pstmt_tostr(lua_State *L) {
	db_pstmt_t *pstmt = luaL_checkudata(L, 1, DB_PSTMT_ID);
	if (pstmt->closed) {
		lua_pushfstring(L, "prepared statement is closed");
	} else {
		lua_pushfstring(L, "prepared statement{id=%d}", pstmt->id);
	}
	return 1;
}

static int add_parameters(lua_State *L, buffer *req) {
	buffer buf = {0}, *params = &buf;
	int i;
	int argc = lua_gettop(L) - 1;

	for (i = 1; i <= argc; i++) {
		int t, b, n = i + 1;
		const char *s;
		size_t l;

		luaL_checkany(L, n);
		t = lua_type(L, n);

		switch (t) {
		case LUA_TNUMBER:
			if (luaL_isinteger(L, n)) {
				lua_Integer d = lua_tointeger(L, n);
				add_int64(params, d);
			} else {
				lua_Number d = lua_tonumber(L, n);
				add_double(params, d);
			}
			break;
		case LUA_TSTRING:
			s = lua_tolstring(L, n, &l);
			add_string_ex(params, s, l);
			break;
		case LUA_TBOOLEAN:
			b = lua_toboolean(L, n);
			if (b) {
				add_int(params, 1);
			} else {
				add_int(params, 0);
			}
			break;
		case LUA_TNIL:
			add_null(params);
			break;
		case LUA_TUSERDATA:
			if (lua_isbignumber(L, n)) {
				long int d = lua_get_bignum_si(L, n);
				if (d == 0 && lua_bignum_is_zero(L, n) != 0) {
					char *s = lua_get_bignum_str(L, n);
					lua_pushfstring(L, "bignum value overflow for binding %s", s);
					free(s);
					free_buffer(params);
					return -1;
				}
				add_int64(params, d);
				break;
			}
		default:
			lua_pushfstring(L, "unsupported type: %s", lua_typename(L, n));
			free_buffer(params);
			return -1;
		}
	}

	add_bytes(req, params->ptr, params->len);
	free_buffer(params);
	return 0;
}

static int db_pstmt_exec(lua_State *L) {
	buffer buf = {0}, *args = &buf;
	rresponse resp = {0}, *response = &resp;
	int rc;
	db_pstmt_t *pstmt = get_db_pstmt(L, 1);

	if (!pstmt || pstmt->id == 0) {
		luaL_error(L, "invalid prepared statement");
	}

	add_int(args, pstmt->id);

	rc = add_parameters(L, args);
	if (rc == -1) {
		free_buffer(args);
		lua_error(L);
	}

	send_vm_api_request(L, "stmtExec", args, response);
	free_buffer(args);
	if (response->error) {
		lua_pushfstring(L, "exec failed: %s", response->error);
		free_response(response);
		lua_error(L);
	}

	lua_Integer changes = get_int64(&response->result, 1);
	lua_pushinteger(L, changes);
	free_response(response);
	return 1;
}

static int db_pstmt_query(lua_State *L) {
	buffer buf = {0}, *args = &buf;
	rresponse resp = {0}, *response = &resp;
	int rc;
	db_pstmt_t *pstmt = get_db_pstmt(L, 1);
	db_rs_t *rs;

	if (!pstmt || pstmt->id == 0) {
		luaL_error(L, "invalid prepared statement");
	}

	checkLuaExecContext(L);

	add_int(args, pstmt->id);

	rc = add_parameters(L, args);
	if (rc != 0) {
		free_buffer(args);
		lua_error(L);
	}

	send_vm_api_request(L, "stmtQuery", args, response);
	free_buffer(args);
	if (response->error) {
		lua_pushfstring(L, "query failed: %s", response->error);
		free_response(response);
		lua_error(L);
	}

	// store the query id on the structure
	rs = (db_rs_t *) lua_newuserdata(L, sizeof(db_rs_t));
	luaL_getmetatable(L, DB_RS_ID);
	lua_setmetatable(L, -2);
	rs->query_id = get_int(&response->result, 1);
	rs->nc = get_int(&response->result, 2);
	rs->closed = 0;
	rs->refno = append_resource(L, RESOURCE_RS_KEY, (void *)rs);

	free_response(response);
	return 1;
}

static void get_column_meta(lua_State *L, bytes *result) {
	bytes names, types;
	get_bytes(result, 1, &names);
	get_bytes(result, 2, &types);
	int colcnt = get_count(&names);
	int i;

	lua_createtable(L, 0, 2);
	lua_pushinteger(L, colcnt);
	lua_setfield(L, -2, "colcnt");
	if (colcnt > 0) {
		lua_createtable(L, colcnt, 0); /* colinfos names */
		lua_createtable(L, colcnt, 0); /* colinfos names decltypes */
	} else {
		lua_pushnil(L);
		lua_pushnil(L);
	}

	for (i = 1; i <= colcnt; i++) {
		char *name = get_string(&names, i);
		lua_pushstring(L, name);
		lua_rawseti(L, -3, i);

		char *decltype = get_string(&types, i);
		lua_pushstring(L, decltype);
		lua_rawseti(L, -2, i);
	}

	lua_setfield(L, -3, "decltypes");
	lua_setfield(L, -2, "names");
}

static int db_pstmt_column_info(lua_State *L) {
	buffer buf = {0}, *args = &buf;
	rresponse resp = {0}, *response = &resp;
	db_pstmt_t *pstmt = get_db_pstmt(L, 1);

	checkLuaExecContext(L);

	add_int(args, pstmt->id);

	send_vm_api_request(L, "stmtColumnInfo", args, response);
	free_buffer(args);
	if (response->error) {
		lua_pushfstring(L, "column_info failed: %s", response->error);
		free_response(response);
		lua_error(L);
	}

	get_column_meta(L, &response->result);
	free_response(response);
	return 1;
}

static int db_pstmt_bind_param_cnt(lua_State *L) {
	db_pstmt_t *pstmt = get_db_pstmt(L, 1);
	checkLuaExecContext(L);
	lua_pushinteger(L, pstmt->colcnt);
	return 1;
}

static void db_pstmt_close(lua_State *L, db_pstmt_t *pstmt, int remove) {
	if (pstmt->closed) {
		return;
	}
	pstmt->closed = 1;
	if (remove) {
		if (luaL_findtable(L, LUA_REGISTRYINDEX, RESOURCE_PSTMT_KEY, 0) != NULL) {
			luaL_error(L, "cannot find the environment of the db module");
		}
		luaL_unref(L, -1, pstmt->refno);
		lua_pop(L, 1);
	}
}

static int db_pstmt_gc(lua_State *L) {
	db_pstmt_close(L, luaL_checkudata(L, 1, DB_PSTMT_ID), 1);
	return 0;
}

static int db_exec(lua_State *L) {
	buffer buf = {0}, *args = &buf;
	rresponse resp = {0}, *response = &resp;
	const char *sql;
	int rc;

	sql = luaL_checkstring(L, 1);
	add_string(args, sql);

	rc = add_parameters(L, args);
	if (rc == -1) {
		free_buffer(args);
		lua_error(L);
	}

	send_vm_api_request(L, "dbExec", args, response);
	free_buffer(args);
	if (response->error) {
		lua_pushfstring(L, "exec failed: %s", response->error);
		free_response(response);
		lua_error(L);
	}

	lua_Integer changes = get_int64(&response->result, 1);
	lua_pushinteger(L, changes);
	free_response(response);
	return 1;
}

static int db_query(lua_State *L) {
	buffer buf = {0}, *args = &buf;
	rresponse resp = {0}, *response = &resp;
	db_rs_t *rs;
	const char *sql;
	int rc;

	checkLuaExecContext(L);

	sql = luaL_checkstring(L, 1);
	add_string(args, sql);

	rc = add_parameters(L, args);
	if (rc == -1) {
		free_buffer(args);
		lua_error(L);
	}

	send_vm_api_request(L, "dbQuery", args, response);  // it could release args
	free_buffer(args);
	if (response->error) {
		lua_pushfstring(L, "query failed: %s", response->error);
		free_response(response);
		lua_error(L);
	}

	// store the query id on the structure
	rs = (db_rs_t *) lua_newuserdata(L, sizeof(db_rs_t));
	luaL_getmetatable(L, DB_RS_ID);
	lua_setmetatable(L, -2);
	rs->query_id = get_int(&response->result, 1);
	rs->nc = get_int(&response->result, 2);
	rs->closed = 0;
	rs->refno = append_resource(L, RESOURCE_RS_KEY, (void *)rs);

	free_response(response);
	return 1;
}

static int db_prepare(lua_State *L) {
	buffer buf = {0}, *args = &buf;
	rresponse resp = {0}, *response = &resp;
	const char *sql;
	db_pstmt_t *pstmt;

	checkLuaExecContext(L);

	sql = luaL_checkstring(L, 1);
	add_string(args, sql);

	send_vm_api_request(L, "dbPrepare", args, response);
	free_buffer(args);
	if (response->error) {
		lua_pushfstring(L, "prepare failed: %s", response->error);
		free_response(response);
		lua_error(L);
	}

	// save the prepared statement id on the structure
	pstmt = (db_pstmt_t *) lua_newuserdata(L, sizeof(db_pstmt_t));
	luaL_getmetatable(L, DB_PSTMT_ID);
	lua_setmetatable(L, -2);
	pstmt->id = get_int(&response->result, 1);
	pstmt->colcnt = get_int(&response->result, 2);
	pstmt->closed = 0;
	pstmt->refno = append_resource(L, RESOURCE_PSTMT_KEY, (void *)pstmt);

	return 1;
}

static int db_get_snapshot(lua_State *L) {
	rresponse resp = {0}, *response = &resp;

	checkLuaExecContext(L);

	send_vm_api_request(L, "dbGetSnapshot", NULL, response);
	if (response->error) {
		lua_pushfstring(L, "get_snapshot failed: %s", response->error);
		free_response(response);
		lua_error(L);
	}

	lua_pushstring(L, get_string(&response->result, 1));
	free_response(response);
	return 1;
}

static int db_open_with_snapshot(lua_State *L) {
	buffer buf = {0}, *args = &buf;
	rresponse resp = {0}, *response = &resp;
	char *snapshot = (char *) luaL_checkstring(L, 1);

	checkLuaExecContext(L);

	add_string(args, snapshot);
	send_vm_api_request(L, "dbOpenWithSnapshot", args, response);
	free_buffer(args);
	if (response->error) {
		lua_pushfstring(L, "open_with_snapshot failed: %s", response->error);
		free_response(response);
		lua_error(L);
	}
	free_response(response);

	return 0;
}

static int db_last_insert_rowid(lua_State *L) {
	rresponse resp = {0}, *response = &resp;

	checkLuaExecContext(L);

	send_vm_api_request(L, "lastInsertRowid", NULL, response);
	if (response->error) {
		lua_pushfstring(L, "last_insert_rowid failed: %s", response->error);
		free_response(response);
		lua_error(L);
	}

	lua_Integer id = get_int64(&response->result, 1);

	lua_pushinteger(L, id);
	return 1;
}

int lua_db_release_resource(lua_State *L) {
	lua_getfield(L, LUA_REGISTRYINDEX, RESOURCE_RS_KEY);
	if (lua_istable(L, -1)) {
		/* T */
		lua_pushnil(L); /* T nil(key) */
		while (lua_next(L, -2)) {
			if (lua_islightuserdata(L, -1)) {
				db_rs_close(L, (db_rs_t *) lua_topointer(L, -1), 0);
			}
			lua_pop(L, 1);
		}
		lua_pop(L, 1);
	}
	lua_getfield(L, LUA_REGISTRYINDEX, RESOURCE_PSTMT_KEY);
	if (lua_istable(L, -1)) {
		/* T */
		lua_pushnil(L); /* T nil(key) */
		while (lua_next(L, -2)) {
			if (lua_islightuserdata(L, -1)) {
				db_pstmt_close(L, (db_pstmt_t *) lua_topointer(L, -1), 0);
			}
			lua_pop(L, 1);
		}
		lua_pop(L, 1);
	}
	return 0;
}

static const luaL_Reg rs_methods[] = {
	{"next",  db_rs_next},
	{"get", db_rs_get},
	{"colcnt", db_rs_colcnt},
	{"__tostring", db_rs_tostr},
	{"__gc", db_rs_gc},
	{NULL, NULL}
};

static const luaL_Reg pstmt_methods[] = {
	{"exec",  db_pstmt_exec},
	{"query", db_pstmt_query},
	{"column_info", db_pstmt_column_info},
	{"bind_param_cnt", db_pstmt_bind_param_cnt},
	{"__tostring", db_pstmt_tostr},
	{"__gc", db_pstmt_gc},
	{NULL, NULL}
};

static const luaL_Reg db_lib[] = {
	{"exec", db_exec},
	{"query", db_query},
	{"prepare", db_prepare},
	{"getsnap", db_get_snapshot},
	{"open_with_snapshot", db_open_with_snapshot},
	{"last_insert_rowid", db_last_insert_rowid},
	{NULL, NULL}
};

int luaopen_db(lua_State *L) {

	luaL_newmetatable(L, DB_RS_ID);
	lua_pushvalue(L, -1);
	lua_setfield(L, -2, "__index");
	luaL_register(L, NULL, rs_methods);

	luaL_newmetatable(L, DB_PSTMT_ID);
	lua_pushvalue(L, -1);
	lua_setfield(L, -2, "__index");
	luaL_register(L, NULL, pstmt_methods);

	luaL_register(L, "db", db_lib);

	lua_pop(L, 3);
	return 1;
}
