/**
 * @file    ast_blk.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"

#include "ast_id.h"
#include "ast_stmt.h"

#include "ast_blk.h"

static ast_blk_t *
ast_blk_new(blk_kind_t kind, trace_t *trc)
{
    ast_blk_t *blk = xcalloc(sizeof(ast_blk_t));

    ast_node_init(blk, trc);

    blk->kind = kind;

    array_init(&blk->ids);
    array_init(&blk->stmts);

    return blk;
}

ast_blk_t *
blk_anon_new(trace_t *trc)
{
    return ast_blk_new(BLK_ANON, trc);
}

ast_blk_t *
blk_root_new(trace_t *trc)
{
    return ast_blk_new(BLK_ROOT, trc);
}

ast_blk_t *
blk_loop_new(trace_t *trc)
{
    return ast_blk_new(BLK_LOOP, trc);
}

ast_blk_t *
blk_search_loop(ast_blk_t *blk)
{
    ASSERT(blk != NULL);

    do {
        if (blk->kind == BLK_LOOP)
            return blk;
    } while ((blk = blk->up) != NULL);

    return NULL;
}

void
ast_blk_dump(ast_blk_t *blk, int indent)
{
}

/* end of ast_blk.c */
