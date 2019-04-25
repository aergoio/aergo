/**
 * @file    check_blk.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "check_id.h"
#include "check_stmt.h"

#include "check_blk.h"

static void
check_unused_ids(vector_t *ids)
{
    int i, j;

    vector_foreach(ids, i) {
        ast_id_t *id = vector_get_id(ids, i);

        if (is_tuple_id(id)) {
            vector_t *elem_ids = id->u_tup.elem_ids;

            vector_foreach(elem_ids, j) {
                ast_id_t *elem_id = vector_get_id(elem_ids, j);

                if (!is_public_id(elem_id) && !elem_id->is_used)
                    WARN(ERROR_UNUSED_ID, &elem_id->pos, elem_id->name);
            }
        }
        else {
            if (is_fn_id(id)) {
                vector_t *param_ids = id->u_fn.param_ids;

                vector_foreach(param_ids, j) {
                    ast_id_t *param_id = vector_get_id(param_ids, j);

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

    /* In the case of a function, the specification and the body are checked separately because
     * they can be called regardless of the declared position. */

    vector_foreach(&blk->ids, i) {
        id_check(check, vector_get_id(&blk->ids, i));
    }

    vector_foreach(&blk->stmts, i) {
        stmt_check(check, vector_get_stmt(&blk->stmts, i));
    }

    vector_foreach(&blk->ids, i) {
        ast_id_t *id = vector_get_id(&blk->ids, i);

        if (!is_fn_id(id))
            continue;

        if (id->u_fn.blk != NULL) {
            check->id = id;
            check->fn_id = id;

            blk_check(check, id->u_fn.blk);

            check->fn_id = NULL;
            check->id = id->up;
        }
    }

    if (!is_itf_blk(blk) && !is_lib_blk(blk))
        check_unused_ids(&blk->ids);

    check->blk = blk->up;
}

/* end of check_blk.c */
