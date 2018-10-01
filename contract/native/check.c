/**
 * @file    check.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "check_blk.h"

#include "check.h"

static void
check_init(check_t *ctx, ast_t *ast)
{
    ASSERT(ast != NULL);

    ctx->ast = ast;
    ctx->blk = NULL;
}

void
check(ast_t *ast, flag_t flag)
{
    check_t ctx;
    list_node_t *node;

    check_init(&ctx, ast);

    list_foreach(node, &ast->blk_l) {
        check_blk(&ctx, (ast_blk_t *)node->item);
    }
}

ast_var_t *
check_search_var(check_t *ctx, int num, char *name)
{
    list_node_t *node;
    ast_blk_t *blk = ctx->blk;

    if (blk == NULL)
        return NULL;

    do {
        list_foreach(node, &blk->var_l) {
            ast_var_t *var = (ast_var_t *)node->item;
            if (var->num > num)
                break;

            ASSERT(var->name != NULL);

            if (strcmp(var->name, name) == 0)
                return var;
        }
    } while ((blk = blk->up) != NULL);
}

ast_struct_t *
check_search_struct(check_t *ctx, int num, char *name)
{
    list_node_t *node;
    ast_blk_t *blk = ctx->blk;

    if (blk == NULL)
        return NULL;

    do {
        list_foreach(node, &blk->struct_l) {
            ast_struct_t *struc = (ast_struct_t *)node->item;
            if (struc->num > num)
                break;

            if (strcmp(struc->name, name) == 0)
                return struc;
        }
    } while ((blk = blk->up) != NULL);
}

/* end of check.c */
