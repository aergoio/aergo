/**
 * @file    check.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _CHECK_H
#define _CHECK_H

#include "common.h"

#include "ast.h"
#include "ast_blk.h"
#include "ast_exp.h"
#include "ast_stmt.h"
#include "ast_id.h"
#include "ast_meta.h"

#define RC_OK           0
#define RC_ERROR        (-1)

#define TRY(stmt)                                                              \
    do {                                                                       \
        int rc = (stmt);                                                       \
        if (rc != RC_OK)                                                       \
            return RC_ERROR;                                                   \
    } while (0)

typedef struct check_s {
    ast_blk_t *root;

    ast_blk_t *blk;     // current block
} check_t;

void check(ast_t *ast, flag_t flag);

#endif /* ! _CHECK_H */
