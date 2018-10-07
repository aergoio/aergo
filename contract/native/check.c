/**
 * @file    check.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_blk.h"

#include "check_id.h"

#include "check.h"

static void
check_init(check_t *check, ast_t *ast)
{
    ast_blk_t *root;

    root = ast->root;

    ASSERT(root != NULL);
    ASSERT(is_empty_array(&root->stmts));
    ASSERT(root->up == NULL);

    check->root = root;
    check->blk = root;
    check->aq_id = NULL;
    check->fn_id = NULL;
}

void
check(ast_t *ast, flag_t flag)
{
    int i;
    check_t check;

    if (ast == NULL)
        /* empty contract can be null */
        return;

    check_init(&check, ast);

    for (i = 0; i < array_size(&check.root->ids); i++) {
        check_id(&check, array_item(&check.root->ids, i, ast_id_t));
    }
}

/* end of check.c */
