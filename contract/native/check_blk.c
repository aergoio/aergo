/**
 * @file    check_blk.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "check_id.h"
#include "check_stmt.h"

#include "check_blk.h"

static void
check_unused_ids(array_t *ids)
{
    int i, j;

    for (i = 0; i < array_size(ids); i++) {
        ast_id_t *id = array_item(ids, i, ast_id_t);

        if (id->mod != MOD_INITIAL && !id->is_used) {
            WARN(ERROR_UNUSED_ID, &id->trc, id->name);
        }
        else if (is_func_id(id)) {
            array_t *param_ids = id->u_func.param_ids;

            for (j = 0; j < array_size(param_ids); j++) {
                ast_id_t *param_id = array_item(param_ids, j, ast_id_t);

                if (!param_id->is_used)
                    WARN(ERROR_UNUSED_ID, &param_id->trc, param_id->name);
            }
        }
    }
}

void
check_blk(check_t *check, ast_blk_t *blk)
{
    int i, j;

    ASSERT(blk != NULL);

    blk->up = check->blk;
    check->blk = blk;

    for (i = 0; i < array_size(&blk->ids); i++) {
        check_id(check, array_item(&blk->ids, i, ast_id_t));
    }

    for (i = 0; i < array_size(&blk->stmts); i++) {
        check_stmt(check, array_item(&blk->stmts, i, ast_stmt_t));
    }

    check_unused_ids(&blk->ids);

    check->blk = blk->up;
}

/* end of check_blk.c */
