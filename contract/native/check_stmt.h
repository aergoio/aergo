/**
 * @file    check_stmt.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _CHECK_STMT_H
#define _CHECK_STMT_H

#include "common.h"

#include "ast_stmt.h"
#include "check.h"

int stmt_check(check_t *check, ast_stmt_t *stmt);

#endif /* ! _CHECK_STMT_H */
