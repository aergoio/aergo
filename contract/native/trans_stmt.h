/**
 * @file    trans_stmt.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _TRANS_STMT_H
#define _TRANS_STMT_H

#include "common.h"

#include "ast_stmt.h"
#include "trans.h"

void stmt_trans_alloc(trans_t *trans, uint32_t reg_idx, bool is_heap, meta_t *meta);
void stmt_trans_initz(trans_t *trans, ast_exp_t *var_exp);

void stmt_trans(trans_t *trans, ast_stmt_t *stmt);

#endif /* ! _TRANS_STMT_H */
