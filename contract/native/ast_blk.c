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
ast_blk_new(blk_kind_t kind, src_pos_t *pos)
{
    ast_blk_t *blk = xcalloc(sizeof(ast_blk_t));

    ast_node_init(blk, pos);

    blk->kind = kind;

    array_init(&blk->ids);
    array_init(&blk->stmts);

    return blk;
}

ast_blk_t *
blk_new_anon(src_pos_t *pos)
{
    return ast_blk_new(BLK_ANON, pos);
}

ast_blk_t *
blk_new_root(src_pos_t *pos)
{
    return ast_blk_new(BLK_ROOT, pos);
}

ast_blk_t *
blk_new_loop(src_pos_t *pos)
{
    return ast_blk_new(BLK_LOOP, pos);
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
