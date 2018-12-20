/**
 * @file    trans_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "array.h"
#include "ir_fn.h"
#include "ir_bb.h"
#include "trans_blk.h"

#include "trans_id.h"

static void
id_trans_var(trans_t *trans, ast_id_t *id)
{
    if (is_global_id(id)) {
        ir_add_global(trans->ir, id);
    }
    else {
        ASSERT(trans->fn != NULL);
        fn_add_local(trans->fn, id);
    }
}

static void
id_trans_fn(trans_t *trans, ast_id_t *id)
{
    ir_fn_t *fn = fn_new(id);

    /* TODO: we have to use this */
    array_join_last(&fn->params, id->u_fn.param_ids);
    array_join_last(&fn->params, id->u_fn.ret_ids);

    if (id->u_fn.blk != NULL) {
        trans->fn = fn;
        trans->bb = fn->entry_bb;

        blk_trans(trans, id->u_fn.blk);

        if (trans->bb != NULL)
            fn_add_basic_blk(fn, trans->bb);

        trans->fn = NULL;
        trans->bb = NULL;
    }

    fn_add_basic_blk(fn, fn->exit_bb);

    ir_add_fn(trans->ir, fn);
}

static void
id_trans_cont(trans_t *trans, ast_id_t *id)
{
    if (id->u_cont.blk != NULL)
        blk_trans(trans, id->u_cont.blk);
}

static void
id_trans_label(trans_t *trans, ast_id_t *id)
{
    id->u_label.stmt->label_bb = bb_new();
}

void
id_trans(trans_t *trans, ast_id_t *id)
{
    switch (id->kind) {
    case ID_VAR:
        id_trans_var(trans, id);
        break;

    case ID_FN:
        id_trans_fn(trans, id);
        break;

    case ID_CONTRACT:
        id_trans_cont(trans, id);
        break;

    case ID_LABEL:
        id_trans_label(trans, id);
        break;

    case ID_STRUCT:
    case ID_ENUM:
        break;

    default:
        ASSERT1(!"invalid identifier", id->kind);
    }
}

/* end of trans_id.c */
