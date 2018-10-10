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
    trace_t trc;
    ast_t *ast = xmalloc(sizeof(ast_t));

    trace_init(&trc, NULL, NULL);

    ast->root = blk_root_new(&trc);

    return ast;
}

void 
ast_dump(ast_t *ast)
{
}

/* end of ast.c */
