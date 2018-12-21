/**
 * @file    check.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_blk.h"
#include "check_id.h"

#include "check.h"

static void
check_init(check_t *check, ast_t *ast, flag_t flag)
{
    ast_blk_t *root;

    check->flag = flag;

    root = ast->root;
    ASSERT(root != NULL);
    ASSERT(is_empty_array(&root->stmts));
    ASSERT(root->up == NULL);

    check->ast = ast;

    check->blk = root;

    check->cont_id = NULL;
    check->qual_id = NULL;
    check->fn_id = NULL;
}

void
check(ast_t *ast, flag_t flag)
{
    int i;
    check_t check;

    check_init(&check, ast, flag);

    for (i = 0; i < array_size(&ast->root->ids); i++) {
        ast_id_t *id = array_get_id(&ast->root->ids, i);

        ASSERT1(is_cont_id(id), id->kind);

        id_check(&check, id);
    }
}

/* end of check.c */
