/**
 * @file    trans_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "array.h"
#include "ir_abi.h"
#include "ir_fn.h"
#include "ir_bb.h"
#include "trans_blk.h"
#include "trans_stmt.h"
#include "trans_exp.h"

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

    if (is_stack_id(id))
        fn_add_stack(trans->fn, id);
    else
        fn_add_local(trans->fn, id);
}

static void
trans_global(array_t *stmts, ast_id_t *id)
{
    meta_t *meta = &id->meta;
    ast_exp_t *dflt_exp = id->u_var.dflt_exp;
    ast_exp_t *id_exp;

    if (is_const_id(id))
        return;

    ASSERT(id->up != NULL);
    ASSERT1(is_global_id(id), id->up->kind);

    if (dflt_exp == NULL) {
        dflt_exp = exp_new_lit(meta->pos);

        if (is_array_type(meta)) {
            value_set_ptr(&dflt_exp->u_lit.val, xcalloc(meta->arr_size), meta->arr_size);
        }
        else if (is_bool_type(meta)) {
            value_set_bool(&dflt_exp->u_lit.val, false);
        }
        else if (is_fpoint_type(meta)) {
            value_set_f64(&dflt_exp->u_lit.val, 0.0);
        }
        else if (is_integer_type(meta) || is_pointer_type(meta)) {
            value_set_i64(&dflt_exp->u_lit.val, 0);
        }
        else {
            ASSERT1(is_struct_type(meta), meta->type);
            value_set_ptr(&dflt_exp->u_lit.val, xcalloc(meta_size(meta)),
                          meta_size(meta));
        }

        meta_copy(&dflt_exp->meta, meta);
        meta_set_undef(&dflt_exp->meta);
    }

    id_exp = exp_new_id(id->name, &dflt_exp->pos);

    id_exp->id = id;
    meta_copy(&id_exp->meta, meta);

    ASSERT2(meta_cmp(&id_exp->meta, &dflt_exp->meta), id_exp->meta.type,
            dflt_exp->meta.type);

    stmt_add(stmts, stmt_new_assign(id_exp, dflt_exp, &dflt_exp->pos));
}

static void
id_trans_fn(trans_t *trans, ast_id_t *id)
{
    ast_id_t *tmp_id;
    ast_exp_t *l_exp, *r_exp;
    ast_blk_t *blk;
    ir_fn_t *fn = fn_new(id);
    ast_id_t *cont_id = id->up;

    ASSERT(cont_id != NULL);
    ASSERT1(is_cont_id(cont_id), cont_id->kind);

    if (id->u_fn.blk == NULL)
        id->u_fn.blk = blk_new_fn(&id->pos);

    blk = id->u_fn.blk;

    tmp_id = id_new_var("stack$offset", MOD_PRIVATE, &id->pos);
    meta_set_int32(&tmp_id->meta);
    array_add_first(&blk->ids, tmp_id);

    l_exp = exp_new_id("stack$offset", &id->pos);
    l_exp->id = tmp_id;
    meta_set_int32(&l_exp->meta);

    r_exp = exp_new_global("stack$high");
    meta_set_int32(&r_exp->meta);

    array_add_first(&blk->stmts, stmt_new_assign(l_exp, r_exp, &id->pos));

    tmp_id = id_new_var("relooper$helper", MOD_PRIVATE, &id->pos);
    meta_set_int32(&tmp_id->meta);
    array_add_first(&blk->ids, tmp_id);

    if (is_ctor_id(id)) {
        int i, j;
        ast_exp_t *l_exp, *r_exp;
        ast_stmt_t *ret_stmt;
        array_t *stmts = array_new();

        ASSERT(id->u_fn.ret_id != NULL);
        ASSERT1(is_return_id(id->u_fn.ret_id), id->u_fn.ret_id->kind);

        tmp_id = id_new_var("object$addr", MOD_PRIVATE, &id->pos);
        meta_set_int32(&tmp_id->meta);
        array_add_last(&blk->ids, tmp_id);

        fn->obj_id = tmp_id;

        l_exp = exp_new_id("object$addr", &id->pos);
        l_exp->id = tmp_id;
        meta_set_int32(&l_exp->meta);

        r_exp = exp_new_global("heap$offset");
        meta_set_int32(&r_exp->meta);

        stmt_add(stmts, stmt_new_assign(l_exp, r_exp, &id->pos));

        /* constructor initializes global variables */
        array_foreach(&cont_id->u_cont.blk->ids, i) {
            ast_id_t *var_id = array_get_id(&cont_id->u_cont.blk->ids, i);

            if (is_var_id(var_id)) {
                trans_global(stmts, var_id);
            }
            else if (is_tuple_id(var_id)) {
                array_foreach(var_id->u_tup.elem_ids, j) {
                    trans_global(stmts, array_get_id(var_id->u_tup.elem_ids, j));
                }
            }
        }

        array_join_first(&blk->stmts, stmts);

        ret_stmt = stmt_new_return(l_exp, &id->pos);
        ret_stmt->u_ret.ret_id = id->u_fn.ret_id;

        stmt_add(&blk->stmts, ret_stmt);
    }
    else {
        tmp_id = id_new_var("object$addr", MOD_CONST, &id->pos);
        tmp_id->is_param = true;
        tmp_id->up = id;
        meta_set_object(&tmp_id->meta, cont_id);

        if (id->u_fn.param_ids == NULL)
            id->u_fn.param_ids = array_new();

        array_add_first(id->u_fn.param_ids, tmp_id);

        fn->obj_id = tmp_id;
    }

    fn->abi = abi_lookup(&trans->ir->abis, id);

    trans->fn = fn;
    trans->bb = fn->entry_bb;

    blk_trans(trans, id->u_fn.blk);

    if (trans->bb != NULL) {
        bb_add_branch(trans->bb, NULL, fn->exit_bb);
        fn_add_basic_blk(fn, trans->bb);
    }

    fn_add_basic_blk(fn, fn->exit_bb);

    trans->fn = NULL;
    trans->bb = NULL;

    ir_add_fn(trans->ir, fn);

    id->abi = fn->abi;
}

static void
id_trans_contract(trans_t *trans, ast_id_t *id)
{
    int i, j;
    ast_blk_t *blk = id->u_cont.blk;

    ASSERT(blk != NULL);

    if (id->u_cont.impl_exp != NULL) {
        /* rearrange functions according to the order in the interface */
        ast_id_t *itf_id = id->u_cont.impl_exp->id;

        array_foreach(&itf_id->u_itf.blk->ids, i) {
            ast_id_t *spec_id = array_get_id(&itf_id->u_itf.blk->ids, i);

            array_foreach(&blk->ids, j) {
                ast_id_t *fn_id = array_get_id(&blk->ids, j);

                if (is_fn_id(fn_id) && strcmp(spec_id->name, fn_id->name) == 0) {
                    array_move(&blk->ids, j, i);
                    break;
                }
            }
        }
    }

    id->idx = array_size(&trans->ir->fns);

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

    array_foreach(id->u_tup.elem_ids, i) {
        id_trans_var(trans, array_get_id(id->u_tup.elem_ids, i));
    }
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

    case ID_CONT:
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
    case ID_ITF:
        break;

    default:
        ASSERT1(!"invalid identifier", id->kind);
    }
}

/* end of trans_id.c */
