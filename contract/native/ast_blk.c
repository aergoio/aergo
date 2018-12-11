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
ast_blk_new(blk_kind_t kind, char *prefix, src_pos_t *pos)
{
    ast_blk_t *blk = xcalloc(sizeof(ast_blk_t));

    ast_node_init(blk, pos);

    blk->kind = kind;
    snprintf(blk->name, sizeof(blk->name), "%s_blk_%d", prefix, blk->num);

    array_init(&blk->ids);
    array_init(&blk->stmts);

    return blk;
}

ast_blk_t *
blk_new_normal(src_pos_t *pos)
{
    return ast_blk_new(BLK_NORMAL, "normal", pos);
}

ast_blk_t *
blk_new_root(src_pos_t *pos)
{
    return ast_blk_new(BLK_ROOT, "root", pos);
}

ast_blk_t *
blk_new_loop(src_pos_t *pos)
{
    return ast_blk_new(BLK_LOOP, "loop", pos);
}

void
blk_set_loop(ast_blk_t *blk)
{
    blk->kind = BLK_LOOP;

    snprintf(blk->name, sizeof(blk->name), "loop_blk_%d", blk->num);
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
