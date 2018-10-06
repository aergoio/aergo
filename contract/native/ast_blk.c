/**
 * @file    ast_blk.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"

#include "ast_id.h"
#include "ast_stmt.h"

#include "ast_blk.h"

ast_blk_t *
ast_blk_new(trace_t *trc)
{
    ast_blk_t *blk = xcalloc(sizeof(ast_blk_t));

    ast_node_init(blk, trc);

    array_init(&blk->ids);
    array_init(&blk->stmts);

    return blk;
}

void
ast_blk_dump(ast_blk_t *blk, int indent)
{
}

/* end of ast_blk.c */
