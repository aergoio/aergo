/**
 * @file    check_stmt.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _CHECK_STMT_H
#define _CHECK_STMT_H

#include "common.h"

#include "check.h"
#include "ast_stmt.h"

void check_stmt(check_t *check, ast_stmt_t *stmt);

#endif /* ! _CHECK_STMT_H */
