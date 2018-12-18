/**
 * @file    check_blk.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "check_id.h"
#include "check_stmt.h"

#include "check_blk.h"

static void
add_label_ids(ast_blk_t *blk)
{
    int i;

    for (i = 0; i < array_size(&blk->stmts); i++) {
        ast_stmt_t *stmt = array_get(&blk->stmts, i, ast_stmt_t);

        if (stmt->label_id != NULL)
            id_add_last(&blk->ids, stmt->label_id);
    }
}

static void
check_unused_ids(array_t *ids)
{
    int i, j;

    for (i = 0; i < array_size(ids); i++) {
        ast_id_t *id = array_get(ids, i, ast_id_t);

        if (!is_ctor_id(id) && !id->is_used) {
            WARN(ERROR_UNUSED_ID, &id->pos, id->name);
        }
        else if (is_fn_id(id)) {
            array_t *param_ids = id->u_fn.param_ids;

            for (j = 0; j < array_size(param_ids); j++) {
                ast_id_t *param_id = array_get(param_ids, j, ast_id_t);

                if (!param_id->is_used)
                    WARN(ERROR_UNUSED_ID, &param_id->pos, param_id->name);
            }
        }
    }
}

void
blk_check(check_t *check, ast_blk_t *blk)
{
    int i;

    ASSERT(blk != NULL);

    blk->up = check->blk;
    check->blk = blk;

    add_label_ids(blk);

    for (i = 0; i < array_size(&blk->ids); i++) {
        ast_id_t *id = array_get(&blk->ids, i, ast_id_t);

        if (is_cont_blk(blk))
            id->scope = SCOPE_GLOBAL;

        id_check(check, id);
    }

    for (i = 0; i < array_size(&blk->stmts); i++) {
        stmt_check(check, array_get(&blk->stmts, i, ast_stmt_t));
    }

    check_unused_ids(&blk->ids);

    check->blk = blk->up;
}

/* end of check_blk.c */
