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
    ast_blk_t *up = trans->blk;

    trans->blk = blk;
    //trans->fn_idx = 0;

    /*
    array_foreach(&blk->ids, i) {
        id_trans(trans, array_get_id(&blk->ids, i));
    }
    */

    array_foreach(&blk->nodes, i) {
        ast_node_t *node = array_get(&blk->nodes, i, ast_node_t);

        if (is_id_node(node))
            id_trans(trans, (ast_id_t *)node);
        else
            stmt_trans(trans, (ast_stmt_t *)node);
    }

    trans->blk = up;

    /* Since there is no relation between general blocks and basic blocks,
     * we deliberately do not do anything about basic blocks here. */
}

/* end of trans_blk.c */
