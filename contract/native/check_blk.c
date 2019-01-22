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

    array_foreach(ids, i) {
        ast_id_t *id = array_get_id(ids, i);

        if (is_tuple_id(id)) {
            array_t *elem_ids = id->u_tup.elem_ids;

            array_foreach(elem_ids, j) {
                ast_id_t *elem_id = array_get_id(elem_ids, j);

                if (!is_public_id(elem_id) && !elem_id->is_used)
                    WARN(ERROR_UNUSED_ID, &elem_id->pos, elem_id->name);
            }
        }
        else {
            if (is_fn_id(id)) {
                array_t *param_ids = id->u_fn.param_ids;

                array_foreach(param_ids, j) {
                    ast_id_t *param_id = array_get_id(param_ids, j);

                    if (!param_id->is_used)
                        WARN(ERROR_UNUSED_ID, &param_id->pos, param_id->name);
                }
            }

            if (!is_public_id(id) && !id->is_used)
                WARN(ERROR_UNUSED_ID, &id->pos, id->name);
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

    array_foreach(&blk->ids, i) {
        id_check(check, array_get_id(&blk->ids, i));
    }

    /* TODO: Checking the return statement on a block for ignore warnings is not
     * possible at this time because there is a return statement on a particular label
     * in a switch-case statement (See 006_statement/switch_case) */

    array_foreach(&blk->stmts, i) {
        stmt_check(check, array_get_stmt(&blk->stmts, i));
    }

    if (!is_itf_blk(blk))
        check_unused_ids(&blk->ids);

    check->blk = blk->up;
}

/* end of check_blk.c */
