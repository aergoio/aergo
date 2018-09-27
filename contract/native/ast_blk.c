/**
 * @file    ast_blk.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_blk.h"

ast_blk_t *
ast_blk_new(errpos_t *pos)
{
    ast_blk_t *blk = xmalloc(sizeof(ast_blk_t));

    blk->pos = *pos;
    blk->name = NULL;

    list_init(&blk->var_l);
    list_init(&blk->struct_l);
    list_init(&blk->stmt_l);
    list_init(&blk->func_l);

    blk->up = NULL;

    return blk;
}

/* end of ast_blk.c */
