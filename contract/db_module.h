#ifndef _DB_MODULE_H
#define _DB_MODULE_H

#include <stdbool.h>
#include <stdint.h>

#include "sqlite3-binding.h"
#include "db_msg.h"

sqlite3 *vm_get_db(request *req);

void handle_db_exec(request *req, const char *sql, char *params_ptr, int params_len);
void handle_db_query(request *req, const char *sql, char *params_ptr, int params_len);
void handle_db_prepare(request *req, const char *sql);
void handle_db_get_snapshot(request *req);
void handle_db_open_with_snapshot(request *req, char *snapshot);
void handle_last_insert_rowid(request *req);
void handle_stmt_exec(request *req, int pstmt_id, char *params_ptr, int params_len);
void handle_stmt_query(request *req, int pstmt_id, char *params_ptr, int params_len);
void handle_stmt_column_info(request *req, int pstmt_id);
void handle_rs_get(request *req, int query_id);
void handle_rs_next(request *req, int query_id);

void lua_db_release_resource();

#endif /* _DB_MODULE_H */
