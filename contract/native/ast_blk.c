/**
 * @file    ast_blk.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_var.h"
#include "ast_struct.h"
#include "ast_stmt.h"
#include "ast_func.h"

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

    return blk;
}

void
ast_blk_dump(ast_blk_t *blk)
{
    list_node_t *node;

    list_foreach(node, &blk->var_l) {
        ast_var_dump((ast_var_t *)node->item);
    }

    list_foreach(node, &blk->struct_l) {
        ast_struct_dump((ast_struct_t *)node->item);
    }

    list_foreach(node, &blk->stmt_l) {
        ast_stmt_dump((ast_stmt_t *)node->item);
    }

    list_foreach(node, &blk->func_l) {
        ast_func_dump((ast_func_t *)node->item);
    }
}

/* end of ast_blk.c */
