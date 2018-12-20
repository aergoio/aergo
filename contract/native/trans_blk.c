/**
 * @file    trans_blk.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "trans_id.h"
#include "trans_stmt.h"

#include "trans_blk.h"

void
blk_trans(trans_t *trans, ast_blk_t *blk)
{
    int i;

    for (i = 0; i < array_size(&blk->ids); i++) {
        id_trans(trans, array_get(&blk->ids, i, ast_id_t));
    }

    for (i = 0; i < array_size(&blk->stmts); i++) {
        stmt_trans(trans, array_get(&blk->stmts, i, ast_stmt_t));
    }

    /* Since there is no relation between general blocks and basic blocks,
     * we deliberately do not do anything about basic blocks here. */
}

/* end of trans_blk.c */
