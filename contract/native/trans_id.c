/**
 * @file    trans_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "vector.h"
#include "ir_md.h"
#include "ir_fn.h"
#include "ir_bb.h"
#include "trans_blk.h"
#include "trans_stmt.h"
#include "trans_exp.h"

#include "trans_id.h"

static void
id_trans_var(trans_t *trans, ast_id_t *id)
{
    if (is_global_id(id))
        /* Initialization of the global variable will be done in the constructor */
        return;

    ASSERT(trans->fn != NULL);

    id->idx = fn_add_register(trans->fn, &id->meta);
}

static void
id_trans_param(trans_t *trans, ast_id_t *id)
{
    ast_id_t *param_id;

    if (is_ctor_id(id))
        return;

    /* All functions that are not constructors must be added the contract address as
     * the first argument, and must also be added to the param_ids to reflect the abi */

    param_id = id_new_tmp_var("cont$addr");
    param_id->u_var.is_param = true;
    param_id->up = id;
    meta_set_object(&param_id->meta, id->up);

    if (id->u_fn.param_ids == NULL)
        id->u_fn.param_ids = vector_new();

    vector_add_first(id->u_fn.param_ids, param_id);
}

static void
update_heap_offset(ir_fn_t *fn, vector_t *stmts, src_pos_t *pos)
{
    ast_exp_t *l_exp, *r_exp, *v_exp;

    /* Increase "heap$offset" by the amount of memory used by initializer or allocator
     * expressions defined in the function */
    l_exp = exp_new_reg(fn->heap_idx);
    meta_set_uint32(&l_exp->meta);

    v_exp = exp_new_lit_i64(fn->heap_usage, pos);
    meta_set_uint32(&v_exp->meta);

    r_exp = exp_new_binary(OP_ADD, l_exp, v_exp, pos);

    vector_add_last(stmts, stmt_new_assign(exp_new_global("heap$offset"), r_exp, pos));
}

static void
id_trans_ctor(trans_t *trans, ast_id_t *id)
{
    int i, j;
    ir_fn_t *fn = trans->fn;
    vector_t *stmts = vector_new();

    ASSERT1(fn->heap_usage == 0, fn->heap_usage);

    fn->cont_idx = fn->heap_idx;

    /* We use the "stmts" vector to keep the declaration order of variables */
    vector_foreach(&id->up->u_cont.blk->ids, i) {
        ast_id_t *elem_id = vector_get_id(&id->up->u_cont.blk->ids, i);
        ast_exp_t *dflt_exp = NULL;

        if (is_var_id(elem_id)) {
            fn_add_global(fn, &elem_id->meta);

            dflt_exp = elem_id->u_var.dflt_exp;
        }
        else if (is_tuple_id(elem_id)) {
            vector_foreach(elem_id->u_tup.elem_ids, j) {
                fn_add_global(fn, &vector_get_id(elem_id->u_tup.elem_ids, j)->meta);
            }

            dflt_exp = elem_id->u_tup.dflt_exp;
        }

        if (dflt_exp != NULL)
            stmt_add(stmts, stmt_make_assign(elem_id, dflt_exp));
    }

    if (fn->heap_usage > 0)
        /* Update "heap$offset" to prevent the heap from being overwritten by another
         * constructor */
        update_heap_offset(fn, &fn->entry_bb->stmts, &id->pos);

    trans->bb = fn->entry_bb;

    vector_foreach(stmts, i) {
        stmt_trans(trans, vector_get_stmt(stmts, i));
    }

    trans->bb = NULL;
}

static void
set_memory_addr(ir_fn_t *fn, ast_id_t *id, uint32_t heap_start)
{
    src_pos_t *pos = &id->pos;
    ast_exp_t *glob_exp, *reg_exp;

    if (fn->stack_usage > 0) {
        ast_exp_t *val_exp, *bin_exp;

        glob_exp = exp_new_global("stack$top");

        reg_exp = exp_new_reg(fn->stack_idx);
        meta_set_uint32(&reg_exp->meta);

        val_exp = exp_new_lit_i64(fn->stack_usage, pos);
        meta_set_uint32(&val_exp->meta);

        bin_exp = exp_new_binary(OP_ADD, reg_exp, val_exp, pos);

        /* At the beginning of "entry_bb", set the current stack offset to the register */
        vector_add_first(&fn->entry_bb->stmts, stmt_new_assign(glob_exp, bin_exp, pos));
        vector_add_first(&fn->entry_bb->stmts, stmt_new_assign(reg_exp, glob_exp, pos));

        /* TODO: checking stack overflow */

        /* If there is any stack variable in the function, it has to be restored to the
         * original value at the end of "exit_bb" because "stack$top" has been changed */
        vector_add_last(&fn->exit_bb->stmts, stmt_new_assign(glob_exp, reg_exp, pos));
    }

    if (fn->heap_usage > 0 || is_ctor_id(id)) {
        /* At the beginning of "entry_bb", set the current heap offset to the register */
        glob_exp = exp_new_global("heap$offset");

        reg_exp = exp_new_reg(fn->heap_idx);
        meta_set_uint32(&reg_exp->meta);

        vector_add_first(&fn->entry_bb->stmts, stmt_new_assign(reg_exp, glob_exp, pos));

        if (fn->heap_usage - heap_start > 0)
            /* Increase "heap$offset" by the amount of memory used by initializer or
             * allocator expressions defined in the function */
            update_heap_offset(fn, &fn->exit_bb->stmts, pos);
    }
}

static void
id_trans_fn(trans_t *trans, ast_id_t *id)
{
    uint32_t heap_start = 0;
    ast_id_t *ret_id = id->u_fn.ret_id;
    ir_md_t *md = trans->md;
    ir_fn_t *fn;

    ASSERT(id->up != NULL);
    ASSERT1(is_cont_id(id->up), id->up->kind);
    ASSERT(md != NULL);

    if (!is_public_id(id) && !id->is_used)
        return;

    id_trans_param(trans, id);

    fn = fn_new(id);

    /* All heap variables access memory by adding relative offset to this register */
    fn->heap_idx = fn_add_register(fn, &addr_meta_);

    /* All stack variables access memory by adding relative offset to this register */
    fn->stack_idx = fn_add_register(fn, &addr_meta_);

    if (ret_id != NULL)
        /* All return values are stored in this register */
        fn->ret_idx = fn_add_register(fn, &ret_id->meta);

    /* It is used internally for binaryen, not for us (see fn_gen()) */
    fn->reloop_idx = fn_add_register(fn, &addr_meta_);

    trans->fn = fn;

    if (is_ctor_id(id)) {
        id_trans_ctor(trans, id);

        /* To check the net usage of the function body */
        heap_start = fn->heap_usage;
    }
    else {
        /* The "cont_idx" is always 0 because it is prepended to parameters */
        fn->cont_idx = 0;
    }

    trans->bb = fn->entry_bb;

    if (id->u_fn.blk != NULL)
        blk_trans(trans, id->u_fn.blk);

    set_memory_addr(fn, id, heap_start);

    if (trans->bb != NULL) {
        bb_add_branch(trans->bb, NULL, fn->exit_bb);
        fn_add_basic_blk(fn, trans->bb);
    }

    fn_add_basic_blk(fn, fn->exit_bb);

    if (ret_id != NULL) {
        /* The contract address or return value is returned at the end of "exit_bb" */
        ast_exp_t *arg_exp;
        ast_stmt_t *ret_stmt;

        if (is_ctor_id(id)) {
            arg_exp = exp_new_reg(fn->cont_idx);
            meta_set_uint32(&arg_exp->meta);
        }
        else {
            arg_exp = exp_new_reg(fn->ret_idx);
            meta_copy(&arg_exp->meta, &ret_id->meta);
        }

        ret_stmt = stmt_new_return(arg_exp, &id->pos);
        ret_stmt->u_ret.ret_id = ret_id;

        vector_add_last(&fn->exit_bb->stmts, ret_stmt);
    }

    trans->fn = NULL;
    trans->bb = NULL;

    md_add_fn(md, fn);
}

static void
id_trans_contract(trans_t *trans, ast_id_t *id)
{
    ast_blk_t *blk = id->u_cont.blk;

    ASSERT(blk != NULL);

    if (id->is_imported) {
        int i;

        vector_foreach(&blk->ids, i) {
            ast_id_t *elem_id = vector_get_id(&blk->ids, i);

            if (is_fn_id(elem_id))
                id_trans_param(trans, elem_id);
        }
        return;
    }

    trans->id = id;
    trans->md = md_new(id->name);

    blk_trans(trans, blk);

    ir_add_md(trans->ir, trans->md);

    trans->md = NULL;
    trans->id = NULL;
}

static void
id_trans_interface(trans_t *trans, ast_id_t *id)
{
    int i;
    ast_blk_t *blk = id->u_itf.blk;

    ASSERT(blk != NULL);

    vector_foreach(&blk->ids, i) {
        ast_id_t *elem_id = vector_get_id(&blk->ids, i);

        ASSERT1(is_fn_id(elem_id), elem_id->kind);
        ASSERT(!is_ctor_id(elem_id));

        /* If the interface type is used as a parameter, we can invoke it with the
         * interface function, so transform the parameter here */

        id_trans_param(trans, elem_id);
    }
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

    vector_foreach(id->u_tup.elem_ids, i) {
        id_trans_var(trans, vector_get_id(id->u_tup.elem_ids, i));
    }
}

void
id_trans(trans_t *trans, ast_id_t *id)
{
    ASSERT(id != NULL);

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

    case ID_ITF:
        id_trans_interface(trans, id);
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
