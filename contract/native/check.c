/**
 * @file    check.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "check_id.h"

#include "check.h"

static void
check_init(check_t *ctx, ast_t *ast)
{
    ast_blk_t *root;

    ASSERT(ast != NULL);

    root = ast->root;

    ASSERT(root != NULL);
    ASSERT(list_empty(&root->stmt_l));
    ASSERT(root->up == NULL);

    ctx->root = root;
    ctx->blk = NULL;
}

void
check(ast_t *ast, flag_t flag)
{
    check_t ctx;
    list_node_t *node;

    check_init(&ctx, ast);

    list_foreach(node, &ctx.root->id_l) {
        check_id(&ctx, (ast_id_t *)node->item);
    }
}

ast_id_t *
check_search_id(check_t *ctx, int num, char *name)
{
    list_node_t *node;
    ast_blk_t *blk = ctx->blk;

    if (blk == NULL)
        return NULL;

    // XXX: need to check siblings of blk

    do {
        list_foreach(node, &blk->id_l) {
            ast_id_t *id = (ast_id_t *)node->item;
            if (id->num > num)
                break;

            ASSERT(id->name != NULL);

            if (strcmp(id->name, name) == 0)
                return id;
        }
    } while ((blk = blk->up) != NULL);
}

/* end of check.c */
