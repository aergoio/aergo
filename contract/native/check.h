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

#define TRY(stmt)                                                              \
    do {                                                                       \
        ec_t ec = (stmt);                                                      \
        if (ec != NO_ERROR)                                                    \
            return ec;                                                         \
    } while (0)

#define THROW(ec, pos, ...)                                                    \
    do {                                                                       \
        error_push((ec), LVL_ERROR, (pos), ## __VA_ARGS__);                    \
        return (ec);                                                           \
    } while (0)

typedef struct check_s {
    ast_blk_t *root;

    /* temporary context */
    ast_blk_t *blk;         /* current block */
    ast_id_t *aq_id;        /* access qualifier */
    ast_id_t *fn_id;        /* function identifier */
} check_t;

void check(ast_t *ast, flag_t flag);

#endif /* ! _CHECK_H */
