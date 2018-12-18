/**
 * @file    trans_stmt.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _TRANS_STMT_H
#define _TRANS_STMT_H

#include "common.h"

#include "ast_stmt.h"
#include "trans.h"

void stmt_trans(trans_t *trans, ast_stmt_t *stmt);

#endif /* ! _TRANS_STMT_H */
