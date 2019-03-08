/**
 * @file    ast.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_blk.h"

#include "ast.h"

ast_t *
ast_new(void)
{
    src_pos_t pos;
    ast_t *ast = xmalloc(sizeof(ast_t));

    src_pos_init(&pos, NULL, NULL);

    ast->root = blk_new_root(&pos);

    return ast;
}

/* end of ast.c */
