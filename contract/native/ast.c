/**
 * @file    ast.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_blk.h"

#include "ast.h"

int node_num_ = 0;

ast_t *
ast_new(void)
{
    errpos_t pos;
    ast_t *ast = xmalloc(sizeof(ast_t));

    errpos_init(&pos, NULL);

    ast->root = ast_blk_new(&pos);

    return ast;
}

void 
ast_dump(ast_t *ast)
{
}

/* end of ast.c */
