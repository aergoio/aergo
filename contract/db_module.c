#include <string.h>
#include <stdlib.h>
#include <ctype.h>
#include <time.h>
#include <sqlite3-binding.h>
#include "vm.h"
#include "sqlcheck.h"
#include "bignum_module.h"
#include "util.h"
#include "_cgo_export.h"
#include "db_msg.h"

extern void checkLuaExecContext(int service);

typedef struct {
	stmt_t *next;
	int id;
	sqlite3 *db;
	sqlite3_stmt *s;
	int closed;
} stmt_t;

typedef struct {
	rs_t *next;
	int id;
	sqlite3 *db;
	sqlite3_stmt *s;
	int closed;
	int nc;
	int shared_stmt;
	char **decltypes;
} rs_t;

// list of stmt_t
stmt_t *pstmt_list = NULL;
// list of rs_t
rs_t *rs_list = NULL;

int last_id = 0;


static int get_next_id() {
	return ++last_id;
}

static rs_t *get_rs(int id) {
	rs_t *rs = rs_list;
	while (rs != NULL && rs->id != id) {
		rs = rs->next;
	}
	return rs;
}

static char *dup_decltype(const char *decltype) {
	char *p, *c;
	if (decltype == NULL) {
		return NULL;
	}
	p = c = malloc(strlen(decltype)+1);
	if (p == NULL) {
		return NULL;
	}
	while ((*c++ = tolower(*decltype++)));
	return p;
}

static void free_decltypes(rs_t *rs) {
	int i;
	for (i = 0; i < rs->nc; i++) {
		if (rs->decltypes[i] != NULL) {
			free(rs->decltypes[i]);
		}
	}
	free(rs->decltypes);
	rs->decltypes = NULL;
}

void handle_rs_get(request *req, int query_id) {
	rs_t *rs = get_rs(query_id);
	int i;
	sqlite3_int64 d;
	double f;
	int n;
	const unsigned char *s;

	checkLuaExecContext(req->service);

	if (rs == NULL || rs->decltypes == NULL) {
		set_error(req, "`get' called without calling `next'");
		return;
	}

	for (i = 0; i < rs->nc; i++) {
		switch (sqlite3_column_type(rs->s, i)) {
		case SQLITE_INTEGER:
			d = sqlite3_column_int64(rs->s, i);
			if (strcmp(rs->decltypes[i], "boolean") == 0) {
				add_bool(req->result, d != 0);
			} else if (strcmp(rs->decltypes[i], "date") == 0 ||
				         strcmp(rs->decltypes[i], "datetime") == 0 ||
				         strcmp(rs->decltypes[i], "timestamp") == 0) {
				char buf[80];
				strftime(buf, 80, "%Y-%m-%d %H:%M:%S", gmtime((time_t *)&d));
				add_string(req->result, buf);
			} else {
				add_int64(req->result, d);
			}
			break;
		case SQLITE_FLOAT:
			f = sqlite3_column_double(rs->s, i);
			add_double(req->result, f);
			break;
		case SQLITE_TEXT:
			n = sqlite3_column_bytes(rs->s, i);
			s = sqlite3_column_text(rs->s, i);
			add_string(req->result, s);
			break;
		case SQLITE_NULL:
		/* fallthrough */
		default: /* unsupported types */
			add_null(req->result);
		}
	}

}

static void rs_close(rs_t *rs, int remove) {
	if (rs->closed) {
		return;
	}
	rs->closed = 1;
	if (rs->decltypes) {
		free_decltypes(rs);
	}
	if (rs->shared_stmt == 0) {
		sqlite3_finalize(rs->s);
	}
	if (remove) {
		llist_remove(&rs_list, rs);
	}
}

void handle_rs_next(request *req, int query_id) {
	int rc;
	rs_t *rs = get_rs(query_id);

	checkLuaExecContext(req->service);

	if (!rs) {
		set_error(req, "invalid query id");
		return;
	}

	rc = sqlite3_step(rs->s);

	if (rc == SQLITE_ROW) {
		add_int(req->result, 1);
	} else if (rc == SQLITE_DONE) {
		add_int(req->result, 0);
		rs_close(rs, 1);
	} else {
		rs_close(rs, 1);
		if (rc != SQLITE_OK) {
			set_error(req, sqlite3_errmsg(rs->db));
		} else {
			set_error(req, "unknown error");
		}
	}

}

static void process_columns(request *req, sqlite3_stmt *stmt, rs_t *rs) {

	int column_count = sqlite3_column_count(stmt);

	rs->nc = column_count;
	rs->decltypes = malloc(sizeof(char *) * column_count);
	if (rs->decltypes == NULL) {
		set_error(req, "out of memory");
		return;
	}

	add_int(req->result, column_count);

	for (int i = 0; i < column_count; i++) {
		char *decltype = dup_decltype(sqlite3_column_decltype(stmt, i));
		rs->decltypes[i] = decltype;
		add_string(req->result, decltype);
	}

}

static stmt_t *get_pstmt(int id) {
	stmt_t *pstmt = pstmt_list;
	while (pstmt != NULL && pstmt->id != id) {
		pstmt = pstmt->next;
	}
	return pstmt;
}

static int bind_parameters(request *req, sqlite3 *db, sqlite3_stmt *pstmt, bytes *params) {
	int rc, i;
	int param_count = get_count(params);
	int bind_count;

	bind_count = sqlite3_bind_parameter_count(pstmt);
	if (param_count != bind_count) {
		set_error(req, "parameter count mismatch: want %d got %d", bind_count, param_count);
		return -1;
	}

	rc = sqlite3_reset(pstmt);
	sqlite3_clear_bindings(pstmt);
	if (rc != SQLITE_ROW && rc != SQLITE_OK && rc != SQLITE_DONE) {
		set_error(req, sqlite3_errmsg(db));
		return -1;
	}

	i = 0;
	char *ptr = NULL;
	int len;
	while (ptr = get_next_item(params, ptr, &len)) {
		char type = get_type(ptr, len);
		ptr += 1; len -= 1;
		i++;
		switch (type) {
		case 'l':
			rc = sqlite3_bind_int64(pstmt, i, read_int64(ptr));
			break;
		case 'd':
			rc = sqlite3_bind_double(pstmt, i, read_double(ptr));
			break;
		case 's':
			rc = sqlite3_bind_text(pstmt, i, ptr, len, SQLITE_TRANSIENT);
			break;
		case 'b':
			if (read_bool(ptr)) {
				rc = sqlite3_bind_int(pstmt, i, 1);
			} else {
				rc = sqlite3_bind_int(pstmt, i, 0);
			}
			break;
		case 'n':
			rc = sqlite3_bind_null(pstmt, i);
			break;
		default:
			set_error(req, "unsupported type: %c", type);
			return -1;
		}
		if (rc != SQLITE_OK) {
			set_error(req, sqlite3_errmsg(db));
			return -1;
		}
	}

	return 0;
}

int handle_stmt_exec(request *req, int pstmt_id, char *params_ptr, int params_len) {
	bytes params = {params_ptr, params_len};
	sqlite3_stmt *pstmt = get_pstmt(pstmt_id);
	int rc;

	checkLuaExecContext(req->service);
	if (luaIsView(req->service)) {
		set_error(req, "not permitted in view function");
		return -1;
	}

	if (pstmt == NULL) {
		set_error(req, "invalid pstmt id");
		return -1;
	}

	rc = bind_parameters(req, pstmt->db, pstmt->s, &params);
	if (rc == -1) {
		sqlite3_reset(pstmt->s);
		sqlite3_clear_bindings(pstmt->s);
		return -1;
	}

	rc = sqlite3_step(pstmt->s);
	if (rc != SQLITE_ROW && rc != SQLITE_OK && rc != SQLITE_DONE) {
		set_error(req, sqlite3_errmsg(pstmt->db));
		sqlite3_reset(pstmt->s);
		sqlite3_clear_bindings(pstmt->s);
		return -1;
	}

	add_int64(req->result, sqlite3_changes(pstmt->db));
	return 0;
}

int handle_stmt_query(request *req, int pstmt_id, char *params_ptr, int params_len) {
	bytes params = {params_ptr, params_len};
	stmt_t *pstmt = get_pstmt(pstmt_id);
	rs_t *rs;
	int rc;

	checkLuaExecContext(req->service);

	if (!sqlite3_stmt_readonly(pstmt->s)) {
		set_error(req, "invalid sql command(permitted readonly)");
		return -1;
	}

	if (pstmt == NULL) {
		set_error(req, "invalid pstmt id");
		return -1;
	}

	rc = bind_parameters(req, pstmt->db, pstmt->s, &params);
	if (rc != 0) {
		sqlite3_reset(pstmt->s);
		sqlite3_clear_bindings(pstmt->s);
		return -1;
	}

	rs = (rs_t *) malloc_zero(sizeof(rs_t));
	rs->id = get_next_id();
	rs->db = pstmt->db;
	rs->s = pstmt->s;
	rs->closed = 0;
	rs->shared_stmt = 1;
	llist_add(&rs_list, rs);

	add_int(req->result, rs->id);

	process_columns(req, pstmt->s, rs);

	return 0;
}

static void get_column_meta(request *req, sqlite3_stmt* stmt) {
	buffer names = {0};
	buffer types = {0};
	const char *name, *decltype;
	int colcnt = sqlite3_column_count(stmt);
	int i;

	for (i = 0; i < colcnt; i++) {
		name = sqlite3_column_name(stmt, i);
		if (name == NULL) {
			add_string(names, "");
		} else {
			add_string(names, name);
		}

		decltype = sqlite3_column_decltype(stmt, i);
		if (decltype == NULL) {
			add_string(types, "");
		} else {
			add_string(types, decltype);
		}
	}

	add_bytes(req->result, names.ptr, names.len);
	add_bytes(req->result, types.ptr, types.len);
	free(names.ptr);
	free(types.ptr);
}

static int handle_stmt_column_info(request *req, int pstmt_id) {
	checkLuaExecContext(req->service);
	stmt_t *pstmt = get_pstmt(pstmt_id);
	if (pstmt == NULL) {
		set_error(req, "invalid pstmt id");
		return -1;
	}
	get_column_meta(req, pstmt->s);
	return 0;
}

static void stmt_close(stmt_t *pstmt, int remove) {
	if (pstmt->closed) {
		return;
	}
	pstmt->closed = 1;
	sqlite3_finalize(pstmt->s);
	if (remove) {
		llist_remove(&pstmt_list, pstmt);
	}
}

int handle_db_exec(request *req, const char *sql, char *params_ptr, int params_len) {
	bytes params = {params_ptr, params_len};
	sqlite3 *db;
	sqlite3_stmt *s;
	int rc;

	checkLuaExecContext(req->service);
	if (luaIsView(req->service)) {
		set_error(req, "not permitted in view function");
		return -1;
	}

	if (!sqlcheck_is_permitted_sql(sql)) {
		set_error(req, "invalid sql command: %s", sql);
		return -1;
	}

	db = vm_get_db(req->service);

	rc = sqlite3_prepare_v2(db, sql, -1, &s, NULL);
	if (rc != SQLITE_OK) {
		set_error(req, sqlite3_errmsg(db));
		return -1;
	}

	rc = bind_parameters(req, db, s, &params);
	if (rc == -1) {
		sqlite3_finalize(s);
		return -1;
	}

	rc = sqlite3_step(s);
	if (rc != SQLITE_ROW && rc != SQLITE_OK && rc != SQLITE_DONE) {
		set_error(req, sqlite3_errmsg(db));
		sqlite3_finalize(s);
		return -1;
	}
	sqlite3_finalize(s);

	add_int64(req->result, sqlite3_changes(db));

	return 0;
}

void handle_db_query(request *req, const char *sql, char *params_ptr, int params_len) {
	bytes params = {params_ptr, params_len};
	int rc;
	sqlite3 *db;
	sqlite3_stmt *s;
	rs_t *rs;

	checkLuaExecContext(req->service);

	if (!sqlcheck_is_readonly_sql(sql)) {
		set_error(req, "invalid sql command(permitted readonly)");
		return;
	}

	db = vm_get_db(req->service);
	rc = sqlite3_prepare_v2(db, sql, -1, &s, NULL);
	if (rc != SQLITE_OK) {
		set_error(req, sqlite3_errmsg(db));
		return;
	}

	rc = bind_parameters(req, db, s, &params);
	if (rc == -1) {
		sqlite3_finalize(s);
		set_error(req, lua_tostring(L, -1));
		return;
	}

	rs = (rs_t *) malloc_zero(sizeof(rs_t));
	rs->id = get_next_id();
	rs->db = db;
	rs->s = s;
	rs->closed = 0;
	rs->shared_stmt = 0;
	llist_add(&rs_list, rs);

	add_int(req->result, rs->id);

	process_columns(req, s, rs);

}

void handle_db_prepare(request *req, const char *sql) {
	int rc;
	sqlite3 *db;
	sqlite3_stmt *s;
	stmt_t *pstmt;

	checkLuaExecContext(req->service);

	if (!sqlcheck_is_permitted_sql(sql)) {
		set_error(req, "invalid sql commond: %s", sql);
		return;
	}

	db = vm_get_db(req->service);
	rc = sqlite3_prepare_v2(db, sql, -1, &s, NULL);
	if (rc != SQLITE_OK) {
		set_error(req, sqlite3_errmsg(db));
		return;
	}

	pstmt = (stmt_t *) malloc_zero(sizeof(stmt_t));
	pstmt->id = get_next_id();
	pstmt->db = db;
	pstmt->s = s;
	pstmt->closed = 0;
	llist_add(&pstmt_list, pstmt);

	add_int(req->result, pstmt->id);
	add_int(req->result, sqlite3_bind_parameter_count(pstmt->s));

}

sqlite3 *vm_get_db(int service) {
	sqlite3 *db;
	checkLuaExecContext(service);
	db = luaGetDbHandle(service);
	if (db == NULL) {
		lua_pushstring(L, "can't open a database connection");
		luaL_throwerror(L);
	}
	return db;
}

/*
static int db_get_snapshot(lua_State *L) {
	char *snapshot;
	int service = checkLuaExecContext(L);

	snapshot = LuaGetDbSnapshot(service);
	strPushAndRelease(L, snapshot);

	return 1;
}

static int db_open_with_snapshot(lua_State *L) {
	char *snapshot = (char *) luaL_checkstring(L, 1);
	char *errStr;
	int service = checkLuaExecContext(L);

	errStr = LuaGetDbHandleSnap(service, snapshot);
	if (errStr != NULL) {
		strPushAndRelease(L, errStr);
		luaL_throwerror(L);
	}
	return 1;
}
*/

void handle_last_insert_rowid(request *req) {
	checkLuaExecContext(req->service);
	sqlite3 *db = vm_get_db(req->service);
	sqlite3_int64 id = sqlite3_last_insert_rowid(db);
	add_int64(req->result, id);
}

// TODO: this must be called when the VM call is done
// the db must also be closed
void lua_db_release_resource() {

	rs_t *rs = rs_list;
	while (rs != NULL) {
		rs_close(rs, 0);
		rs = rs->next;
	}

	stmt_t *pstmt = pstmt_list;
	while (pstmt != NULL) {
		stmt_close(pstmt, 0);
		pstmt = pstmt->next;
	}

}
