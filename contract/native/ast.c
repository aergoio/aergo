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
    ast_t *ast = xmalloc(sizeof(ast_t));

    list_init(&ast->blk_l);

    return ast;
}

void 
ast_dump(ast_t *ast)
{
    list_node_t *node;

    list_foreach(node, &ast->blk_l) {
        ast_blk_dump((ast_blk_t *)node->item);
    }
}

/* end of ast.c */
