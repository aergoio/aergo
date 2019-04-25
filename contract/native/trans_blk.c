/**
 * @file    trans_blk.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ir_bb.h"
#include "ir_fn.h"
#include "trans_id.h"
#include "trans_stmt.h"

#include "trans_blk.h"

void
blk_trans(trans_t *trans, ast_blk_t *blk)
{
    int i;
    ast_blk_t *up = trans->blk;

    ASSERT(blk != NULL);

    trans->blk = blk;

    if (is_loop_blk(blk)) {
        ASSERT(trans->bb != NULL);
        ASSERT(trans->cont_bb != NULL);
        ASSERT(trans->break_bb != NULL);

        bb_add_branch(trans->bb, NULL, trans->cont_bb);

        fn_add_basic_blk(trans->fn, trans->bb);

        trans->bb = trans->cont_bb;
    }

    vector_foreach(&blk->ids, i) {
        id_trans(trans, vector_get_id(&blk->ids, i));
    }

    vector_foreach(&blk->stmts, i) {
        stmt_trans(trans, vector_get_stmt(&blk->stmts, i));
    }

    trans->blk = up;
}

/* end of trans_blk.c */
