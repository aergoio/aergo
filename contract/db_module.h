#ifndef _DB_MODULE_H
#define _DB_MODULE_H

#include <stdbool.h>
#include <stdint.h>

#include "sqlite3-binding.h"
#include "db_msg.h"

sqlite3 *vm_get_db(request *req);

void handle_db_exec(request *req, char *args_ptr, int args_len);
void handle_db_query(request *req, char *args_ptr, int args_len);
void handle_db_prepare(request *req, char *args_ptr, int args_len);
void handle_db_get_snapshot(request *req);
void handle_db_open_with_snapshot(request *req, char *args_ptr, int args_len);
void handle_last_insert_rowid(request *req);
void handle_stmt_exec(request *req, char *args_ptr, int args_len);
void handle_stmt_query(request *req, char *args_ptr, int args_len);
void handle_stmt_column_info(request *req, char *args_ptr, int args_len);
void handle_rs_get(request *req, char *args_ptr, int args_len);
void handle_rs_next(request *req, char *args_ptr, int args_len);

void db_release_resource();

#endif /* _DB_MODULE_H */
