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
    ASSERT(array_empty(&root->stmts));
    ASSERT(root->up == NULL);

    ctx->root = root;
    ctx->blk = root;
}

void
check(ast_t *ast, flag_t flag)
{
    int i;
    check_t ctx;

    check_init(&ctx, ast);

    for (i = 0; i < array_size(&ctx.root->ids); i++) {
        check_id(&ctx, array_item(&ctx.root->ids, i, ast_id_t));
    }
}

/* end of check.c */
