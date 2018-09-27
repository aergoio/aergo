/**
 * @file    ast.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast.h"

int node_num_ = 0;

ast_t *
ast_new(void)
{
    ast_t *ast = xmalloc(sizeof(ast_t));

    list_init(&ast->blk_l);

    return ast;
}

void 
ast_dump(ast_t *ast)
{
}

/* end of ast.c */
