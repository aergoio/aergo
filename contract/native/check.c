/**
 * @file    check.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_blk.h"
#include "check_id.h"

#include "check.h"

static void
check_init(check_t *check, ast_blk_t *root, flag_t flag)
{
    ASSERT(root != NULL);
    ASSERT(root->up == NULL);
    ASSERT1(is_empty_vector(&root->stmts), vector_size(&root->stmts));

    check->flag = flag;

    check->blk = root;
    check->id = NULL;

    check->cont_id = NULL;
    check->impl_id = NULL;

    check->qual_id = NULL;
    check->fn_id = NULL;
}

void
check(ast_t *ast, flag_t flag)
{
    int i;
    check_t check;
    ast_blk_t *root = ast->root;

    check_init(&check, root, flag);

    vector_foreach(&root->ids, i) {
        ast_id_t *id = vector_get_id(&root->ids, i);

        ASSERT1(is_root_id(id), id->kind);

        id_check(&check, id);
    }
}

/* end of check.c */
