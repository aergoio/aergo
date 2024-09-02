#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include <time.h>
#include <sqlite3-binding.h>
#include "sqlcheck.h"
#include "db_module.h"
#include "linkedlist.h"
#include "_cgo_export.h"

typedef struct stmt_t stmt_t;
struct stmt_t{
	stmt_t *next;
	int id;
	sqlite3 *db;
	sqlite3_stmt *s;
	int closed;
};

typedef struct rs_t rs_t;
struct rs_t{
	rs_t *next;
	int id;
	sqlite3 *db;
	sqlite3_stmt *s;
	int closed;
	int nc;
	int shared_stmt;
	char **decltypes;
};

// list of stmt_t
stmt_t *pstmt_list = NULL;
// list of rs_t
rs_t *rs_list = NULL;

int last_id = 0;


static void *malloc_zero(request *req, size_t size) {
	void *ptr = malloc(size);
	if (ptr == NULL) {
		set_error(req, "out of memory");
		return NULL;
	}
	memset(ptr, 0, size);
	return ptr;
}

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

void handle_rs_get(request *req, char *args_ptr, int args_len) {
	bytes args = {args_ptr, args_len};
	rs_t *rs;
	int query_id, i;
	sqlite3_int64 d;
	double f;
	int n;
	const unsigned char *s;

	if (!checkDbExecContext(req->service)) {
		set_error(req, "invalid db context");
		return;
	}

	query_id = get_int(&args, 1);
	rs = get_rs(query_id);
	if (rs == NULL || rs->decltypes == NULL) {
		set_error(req, "'get' called without calling 'next'");
		return;
	}
	if (sqlite3_column_count(rs->s) != rs->nc) {
		set_error(req, "column count mismatch - expected %d got %d", rs->nc, sqlite3_column_count(rs->s));
		return;
	}

	for (i = 0; i < rs->nc; i++) {
		switch (sqlite3_column_type(rs->s, i)) {
		case SQLITE_INTEGER:
			d = sqlite3_column_int64(rs->s, i);
			char *decltype = rs->decltypes[i];
			if (decltype && strcmp(decltype, "boolean") == 0) {
				add_bool(&req->result, d != 0);
			} else if (decltype &&
				        (strcmp(decltype, "date") == 0 ||
				         strcmp(decltype, "datetime") == 0 ||
				         strcmp(decltype, "timestamp") == 0)) {
				char buf[80];
				strftime(buf, 80, "%Y-%m-%d %H:%M:%S", gmtime((time_t *)&d));
				add_string(&req->result, buf);
			} else {
				add_int64(&req->result, d);
			}
			break;
		case SQLITE_FLOAT:
			f = sqlite3_column_double(rs->s, i);
			add_double(&req->result, f);
			break;
		case SQLITE_TEXT:
			n = sqlite3_column_bytes(rs->s, i);
			s = sqlite3_column_text(rs->s, i);
			add_string(&req->result, s);
			break;
		case SQLITE_NULL:
		/* fallthrough */
		default: /* unsupported types */
			add_null(&req->result);
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

void handle_rs_next(request *req, char *args_ptr, int args_len) {
	bytes args = {args_ptr, args_len};
	int query_id, rc;
	rs_t *rs;

	if (!checkDbExecContext(req->service)) {
		set_error(req, "invalid db context");
		return;
	}

	query_id = get_int(&args, 1);
	rs = get_rs(query_id);
	if (!rs) {
		set_error(req, "invalid query id");
		return;
	}

	rc = sqlite3_step(rs->s);

	if (rc == SQLITE_ROW) {
		add_bool(&req->result, true);
	} else if (rc == SQLITE_DONE) {
		add_bool(&req->result, false);
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

	for (int i = 0; i < column_count; i++) {
		char *decltype = dup_decltype(sqlite3_column_decltype(stmt, i));
		rs->decltypes[i] = decltype;
	}

	add_int(&req->result, column_count);

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

	if (param_count == 0) {
		return 0;
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
			rc = sqlite3_bind_text(pstmt, i, ptr, len-1, SQLITE_TRANSIENT);
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

void handle_stmt_exec(request *req, char *args_ptr, int args_len) {
	bytes args = {args_ptr, args_len};
	bytes params;
	int pstmt_id, rc;
	stmt_t *pstmt;
	bool success;

	if (!checkDbExecContext(req->service)) {
		set_error(req, "invalid db context");
		return;
	}
	if (luaIsView(req->service)) {
		set_error(req, "not permitted in view function");
		return;
	}

	pstmt_id = get_int(&args, 1);
	success = get_bytes(&args, 2, &params);
	if (!success) {
		set_error(req, "invalid parameters");  //FIXME: remove later
		return;
	}

	pstmt = get_pstmt(pstmt_id);
	if (pstmt == NULL) {
		set_error(req, "invalid pstmt id");
		return;
	}

	rc = bind_parameters(req, pstmt->db, pstmt->s, &params);
	if (rc == -1) {
		sqlite3_reset(pstmt->s);
		sqlite3_clear_bindings(pstmt->s);
		return;
	}

	rc = sqlite3_step(pstmt->s);
	if (rc != SQLITE_ROW && rc != SQLITE_OK && rc != SQLITE_DONE) {
		set_error(req, sqlite3_errmsg(pstmt->db));
		sqlite3_reset(pstmt->s);
		sqlite3_clear_bindings(pstmt->s);
		return;
	}

	add_int64(&req->result, sqlite3_changes(pstmt->db));
}

void handle_stmt_query(request *req, char *args_ptr, int args_len) {
	bytes args = {args_ptr, args_len};
	bytes params;
	int pstmt_id, rc;
	stmt_t *pstmt;
	rs_t *rs;
	bool success;

	if (!checkDbExecContext(req->service)) {
		set_error(req, "invalid db context");
		return;
	}

	pstmt_id = get_int(&args, 1);
	success = get_bytes(&args, 2, &params);
	if (!success) {
		set_error(req, "invalid parameters");  //FIXME: remove later
		return;
	}

	pstmt = get_pstmt(pstmt_id);
	if (pstmt == NULL) {
		set_error(req, "invalid pstmt id");
		return;
	}
	if (!sqlite3_stmt_readonly(pstmt->s)) {
		set_error(req, "invalid sql command (only read permitted)");
		return;
	}

	rc = bind_parameters(req, pstmt->db, pstmt->s, &params);
	if (rc != 0) {
		sqlite3_reset(pstmt->s);
		sqlite3_clear_bindings(pstmt->s);
		return;
	}

	rs = (rs_t *) malloc_zero(req, sizeof(rs_t));
	if (rs == NULL) {
		return;
	}
	rs->id = get_next_id();
	rs->db = pstmt->db;
	rs->s = pstmt->s;
	rs->closed = 0;
	rs->shared_stmt = 1;
	llist_add(&rs_list, rs);

	add_int(&req->result, rs->id);

	process_columns(req, pstmt->s, rs);

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
			add_string(&names, "");
		} else {
			add_string(&names, name);
		}

		decltype = sqlite3_column_decltype(stmt, i);
		if (decltype == NULL) {
			add_string(&types, "");
		} else {
			add_string(&types, decltype);
		}
	}

	add_bytes(&req->result, names.ptr, names.len);
	add_bytes(&req->result, types.ptr, types.len);
	free(names.ptr);
	free(types.ptr);
}

void handle_stmt_column_info(request *req, char *args_ptr, int args_len) {
	bytes args = {args_ptr, args_len};
	int pstmt_id;
	stmt_t *pstmt;

	if (!checkDbExecContext(req->service)) {
		set_error(req, "invalid db context");
		return;
	}

	pstmt_id = get_int(&args, 1);
	pstmt = get_pstmt(pstmt_id);
	if (pstmt == NULL) {
		set_error(req, "invalid pstmt id");
		return;
	}

	get_column_meta(req, pstmt->s);
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

void handle_db_exec(request *req, char *args_ptr, int args_len) {
	sqlite3 *db;
	sqlite3_stmt *s;
	bytes args = {args_ptr, args_len};
	bytes params;
	char *sql;
	int rc;
	bool success;

	if (!checkDbExecContext(req->service)) {
		set_error(req, "invalid db context");
		return;
	}
	if (luaIsView(req->service)) {
		set_error(req, "not permitted in view function");
		return;
	}

	sql = get_string(&args, 1);
	success = get_bytes(&args, 2, &params);
	if (!success) {
		set_error(req, "invalid parameters");  //FIXME: remove later
		return;
	}

	if (!sqlcheck_is_permitted_sql(sql)) {
		set_error(req, "invalid sql command: %s", sql);
		return;
	}

	db = vm_get_db(req);
	if (db == NULL) {
		// error already set by vm_get_db
		return;
	}

	rc = sqlite3_prepare_v2(db, sql, -1, &s, NULL);
	if (rc != SQLITE_OK) {
		set_error(req, sqlite3_errmsg(db));
		return;
	}

	rc = bind_parameters(req, db, s, &params);
	if (rc == -1) {
		sqlite3_finalize(s);
		return;
	}

	rc = sqlite3_step(s);
	if (rc != SQLITE_ROW && rc != SQLITE_OK && rc != SQLITE_DONE) {
		set_error(req, sqlite3_errmsg(db));
		sqlite3_finalize(s);
		return;
	}
	sqlite3_finalize(s);

	add_int64(&req->result, sqlite3_changes(db));

}

void handle_db_query(request *req, char *args_ptr, int args_len) {
	bytes args = {args_ptr, args_len};
	bytes params;
	char *sql;
	int rc;
	sqlite3 *db;
	sqlite3_stmt *s;
	rs_t *rs;
	bool success;

	if (!checkDbExecContext(req->service)) {
		set_error(req, "invalid db context");
		return;
	}
	sql = get_string(&args, 1);
	success = get_bytes(&args, 2, &params);
	if (!success) {
		set_error(req, "invalid parameters");  //FIXME: remove later
		return;
	}
	if (!sqlcheck_is_readonly_sql(sql)) {
		set_error(req, "invalid sql command (only read permitted)");
		return;
	}

	db = vm_get_db(req);
	if (db == NULL) {
		// error already set by vm_get_db
		return;
	}

	rc = sqlite3_prepare_v2(db, sql, -1, &s, NULL);
	if (rc != SQLITE_OK) {
		set_error(req, sqlite3_errmsg(db));
		return;
	}

	rc = bind_parameters(req, db, s, &params);
	if (rc == -1) {
		sqlite3_finalize(s);
		return;
	}

	rs = (rs_t *) malloc_zero(req, sizeof(rs_t));
	if (rs == NULL) {
		sqlite3_finalize(s);
		set_error(req, "out of memory");
		return;
	}
	rs->id = get_next_id();
	rs->db = db;
	rs->s = s;
	rs->closed = 0;
	rs->shared_stmt = 0;
	llist_add(&rs_list, rs);

	add_int(&req->result, rs->id);

	process_columns(req, s, rs);

}

void handle_db_prepare(request *req, char *args_ptr, int args_len) {
	bytes args = {args_ptr, args_len};
	char *sql;
	int rc;
	sqlite3 *db;
	sqlite3_stmt *s;
	stmt_t *pstmt;

	if (!checkDbExecContext(req->service)) {
		set_error(req, "invalid db context");
		return;
	}
	sql = get_string(&args, 1);
	if (!sqlcheck_is_permitted_sql(sql)) {
		set_error(req, "invalid sql command: %s", sql);
		return;
	}

	db = vm_get_db(req);
	if (db == NULL) {
		// error already set by vm_get_db
		return;
	}

	rc = sqlite3_prepare_v2(db, sql, -1, &s, NULL);
	if (rc != SQLITE_OK) {
		set_error(req, sqlite3_errmsg(db));
		return;
	}

	pstmt = (stmt_t *) malloc_zero(req, sizeof(stmt_t));
	if (pstmt == NULL) {
		sqlite3_finalize(s);
		set_error(req, "out of memory");
		return;
	}
	pstmt->id = get_next_id();
	pstmt->db = db;
	pstmt->s = s;
	pstmt->closed = 0;
	llist_add(&pstmt_list, pstmt);

	add_int(&req->result, pstmt->id);
	add_int(&req->result, sqlite3_bind_parameter_count(pstmt->s));

}

sqlite3 *vm_get_db(request *req) {
	sqlite3 *db;
	db = luaGetDbHandle(req->service);
	if (db == NULL) {
		set_error(req, "can't open a connection to the contract's database");
	}
	return db;
}

void handle_db_get_snapshot(request *req) {
	char *snapshot;
	if (!checkDbExecContext(req->service)) {
		set_error(req, "invalid db context");
		return;
	}
	snapshot = LuaGetDbSnapshot(req->service);
	add_string(&req->result, snapshot);
}

void handle_db_open_with_snapshot(request *req, char *args_ptr, int args_len) {
	bytes args = {args_ptr, args_len};
	char *snapshot;
	char *errStr;

	if (!checkDbExecContext(req->service)) {
		set_error(req, "invalid db context");
		return;
	}

	snapshot = get_string(&args, 1);
	if (snapshot == NULL) {
		set_error(req, "invalid snapshot");
		return;
	}

	errStr = LuaGetDbHandleSnap(req->service, snapshot);
	if (errStr != NULL) {
		set_error(req, errStr);
		free(errStr);
		return;
	}

	add_string(&req->result, "ok");
}

void handle_last_insert_rowid(request *req) {
	if (!checkDbExecContext(req->service)) {
		set_error(req, "invalid db context");
		return;
	}
	sqlite3 *db = vm_get_db(req);
	if (db == NULL) {
		// error already set by vm_get_db
		return;
	}
	sqlite3_int64 id = sqlite3_last_insert_rowid(db);
	add_int64(&req->result, id);
}

void db_release_resource() {

	rs_t *rs = rs_list, *rs_next;
	while (rs != NULL) {
		rs_next = rs->next;
		rs_close(rs, 0);
		free(rs);
		rs = rs_next;
	}
	rs_list = NULL;

	stmt_t *pstmt = pstmt_list, *pstmt_next;
	while (pstmt != NULL) {
		pstmt_next = pstmt->next;
		stmt_close(pstmt, 0);
		free(pstmt);
		pstmt = pstmt_next;
	}
	pstmt_list = NULL;

}
