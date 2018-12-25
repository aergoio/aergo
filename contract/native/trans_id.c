/**
 * @file    trans_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "array.h"
#include "ir_fn.h"
#include "ir_bb.h"
#include "trans_blk.h"
#include "trans_stmt.h"

#include "trans_id.h"

static void
id_trans_var(trans_t *trans, ast_id_t *id)
{
    if (is_global_id(id)) {
        /* Initialization of the global variable will be done in the constructor */
        ir_add_global(trans->ir, id);
        return;
    }

    ASSERT(trans->fn != NULL);

    if (is_local_id(id))
        fn_add_local(trans->fn, id);
    else
        fn_add_stack(trans->fn, id);

    if (id->u_var.dflt_stmt != NULL)
        stmt_trans(trans, id->u_var.dflt_stmt);
}

static void
id_trans_fn(trans_t *trans, ast_id_t *id)
{
    ir_fn_t *fn = fn_new(id);

    trans->fn = fn;
    trans->bb = fn->entry_bb;

    if (is_ctor_id(id)) {
        int i;
        ast_id_t *cont_id = id->u_fn.cont_id;

        ASSERT(cont_id != NULL);
        ASSERT1(is_cont_id(cont_id), cont_id->kind);

        for (i = 0; i < array_size(&cont_id->u_cont.blk->ids); i++) {
            ast_id_t *fld_id = array_get_id(&cont_id->u_cont.blk->ids, i);

            if (is_var_id(fld_id) && fld_id->u_var.dflt_stmt != NULL)
                stmt_trans(trans, fld_id->u_var.dflt_stmt);
        }

        ASSERT(trans->bb == fn->entry_bb);
    }

    if (id->u_fn.blk != NULL) {
        blk_trans(trans, id->u_fn.blk);

        if (trans->bb != NULL) {
            bb_add_branch(trans->bb, NULL, fn->exit_bb);
            fn_add_basic_blk(fn, trans->bb);
        }

        fn_add_basic_blk(fn, fn->exit_bb);
    }
    else {
        fn_add_basic_blk(fn, fn->entry_bb);
    }

    trans->fn = NULL;
    trans->bb = NULL;

    ir_add_fn(trans->ir, fn);
}

static void
id_trans_contract(trans_t *trans, ast_id_t *id)
{
    ASSERT(id->u_cont.blk != NULL);

    blk_trans(trans, id->u_cont.blk);
}

static void
id_trans_label(trans_t *trans, ast_id_t *id)
{
    id->u_lab.stmt->label_bb = bb_new();
}

static void
id_trans_tuple(trans_t *trans, ast_id_t *id)
{
    int i;

    for (i = 0; i < array_size(&id->u_tup.var_ids); i++) {
        id_trans_var(trans, array_get_id(&id->u_tup.var_ids, i));
    }

    if (id->u_tup.dflt_stmt != NULL)
        stmt_trans(trans, id->u_tup.dflt_stmt);
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
        id_trans_contract(trans, id);
        break;

    case ID_LABEL:
        id_trans_label(trans, id);
        break;

    case ID_TUPLE:
        id_trans_tuple(trans, id);
        break;

    case ID_STRUCT:
    case ID_ENUM:
        break;

    default:
        ASSERT1(!"invalid identifier", id->kind);
    }
}

/* end of trans_id.c */
