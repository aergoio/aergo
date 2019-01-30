/**
 * @file    trans_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "vector.h"
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
    if (is_global_id(id))
        /* Initialization of the global variable will be done in the constructor */
        return;

    ASSERT(trans->fn != NULL);

    id->idx = fn_add_register(trans->fn, &id->meta);

    if (id->u_var.dflt_exp != NULL)
        stmt_trans(trans, stmt_make_assign(id, id->u_var.dflt_exp));
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
    l_exp = exp_new_register(fn->heap_idx);
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

    trans->bb = fn->entry_bb;

    vector_foreach(stmts, i) {
        stmt_trans(trans, vector_get_stmt(stmts, i));
    }

    trans->bb = NULL;

    if (fn->heap_usage > 0)
        /* Update "heap$offset" to prevent the heap from being overwritten by another
         * constructor */
        update_heap_offset(fn, &fn->entry_bb->stmts, &id->pos);
}

static void
set_memory_addr(ir_fn_t *fn, uint32_t heap_start, src_pos_t *pos)
{
    ast_exp_t *glob_exp, *reg_exp;

    if (fn->stack_usage > 0) {
        ast_exp_t *val_exp, *bin_exp;

        glob_exp = exp_new_global("stack$offset");

        reg_exp = exp_new_register(fn->stack_idx);
        meta_set_uint32(&reg_exp->meta);

        /* At the beginning of "entry_bb", set the current stack offset to the register */
        val_exp = exp_new_lit_i64(fn->stack_usage, pos);
        meta_set_uint32(&val_exp->meta);

        bin_exp = exp_new_binary(OP_SUB, glob_exp, val_exp, pos);

        vector_add_first(&fn->entry_bb->stmts, stmt_new_assign(glob_exp, reg_exp, pos));
        vector_add_first(&fn->entry_bb->stmts, stmt_new_assign(reg_exp, bin_exp, pos));

        /* If there is any stack variable in the function, it has to be restored to the
         * original value at the end of "exit_bb" because "stack$offset" has been
         * changed */
        bin_exp = exp_new_binary(OP_ADD, reg_exp, val_exp, pos);

        vector_add_last(&fn->exit_bb->stmts, stmt_new_assign(glob_exp, bin_exp, pos));
    }

    if (fn->heap_usage > 0) {
        /* At the beginning of "entry_bb", set the current heap offset to the register */
        glob_exp = exp_new_global("heap$offset");

        reg_exp = exp_new_register(fn->heap_idx);
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
    ir_t *ir = trans->ir;
    ir_fn_t *fn = fn_new(id);

    ASSERT(id->up != NULL);
    ASSERT1(is_cont_id(id->up), id->up->kind);

    trans->fn = fn;

    id_trans_param(trans, id);

    fn->abi = abi_lookup(&ir->abis, id);
    id->abi = fn->abi;

    /* All heap variables access memory by adding relative offset to this register */
    fn->heap_idx = fn_add_register(fn, &addr_meta_);

    /* All stack variables access memory by adding relative offset to this register */
    fn->stack_idx = fn_add_register(fn, &addr_meta_);

    if (ret_id != NULL)
        /* All return values are stored in this register */
        fn->ret_idx = fn_add_register(fn, &ret_id->meta);

    /* It is used internally for binaryen, not for us (see fn_gen()) */
    fn->reloop_idx = fn_add_register(fn, &addr_meta_);

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

    set_memory_addr(fn, heap_start, &id->pos);

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
            arg_exp = exp_new_register(fn->cont_idx);
            meta_set_uint32(&arg_exp->meta);
        }
        else {
            arg_exp = exp_new_register(fn->ret_idx);
            meta_copy(&arg_exp->meta, &ret_id->meta);
        }

        ret_stmt = stmt_new_return(arg_exp, &id->pos);
        ret_stmt->u_ret.ret_id = ret_id;

        vector_add_last(&fn->exit_bb->stmts, ret_stmt);
    }

    trans->fn = NULL;
    trans->bb = NULL;

    ir_add_fn(ir, fn);
}

static void
id_trans_contract(trans_t *trans, ast_id_t *id)
{
    int i, j;
    int fn_idx = 1;
    ast_id_t *idx_id;
    ast_blk_t *blk = id->u_cont.blk;
    ir_t *ir = trans->ir;

    ASSERT(blk != NULL);
    ASSERT1(ir->offset == 0, ir->offset);

    if (id->u_cont.impl_exp != NULL) {
        /* Reorder functions according to the order in the interface
         *
         * The reason for reordering the functions is that if the parameter of the
         * function is an interface and there are several contracts that implement the
         * interface, there is no way to know the location of the function belonging to
         * the contract used as the argument.
         *
         * By making the index of the function equal between the interface and the
         * contract, the function can be called through the index of the interface
         * function */
        ast_id_t *itf_id = id->u_cont.impl_exp->id;

        ASSERT1(is_itf_id(itf_id), itf_id->kind);

        vector_foreach(&itf_id->u_itf.blk->ids, i) {
            ast_id_t *spec_id = vector_get_id(&itf_id->u_itf.blk->ids, i);

            vector_foreach(&blk->ids, j) {
                ast_id_t *fn_id = vector_get_id(&blk->ids, j);

                if (is_fn_id(fn_id) && strcmp(spec_id->name, fn_id->name) == 0) {
                    vector_move(&blk->ids, j, i);
                    break;
                }
            }
        }
    }

    /* Move the constructor to the first position because it handles the memory
     * allocation of global variables. Even if the contract implements interface,
     * there is no problem because index 0 is empty. (see id_trans_interface()) */
    vector_foreach(&blk->ids, i) {
        if (is_ctor_id(vector_get_id(&blk->ids, i))) {
            vector_move(&blk->ids, i, 0);
            break;
        }
    }

    /* Because the cross-reference is possible between functions, the index of the
     * function is numbered before transformation */
    vector_foreach(&blk->ids, i) {
        ast_id_t *fn_id = vector_get_id(&blk->ids, i);

        if (is_ctor_id(fn_id))
            /* Since the constructor can be called from any location (including another
             * contracts), it should always be accessed with an absolute index */
            fn_id->idx = vector_size(&ir->fns);
        else if (is_fn_id(fn_id))
            /* The "idx" is the relative index within the contract */
            fn_id->idx = fn_idx++;
    }

    /* This value, like any other global variable, is stored in the heap area used by
     * the contract, and is stored in the first 4 bytes of the area. All functions
     * also access table by adding relative index to this value */
    idx_id = id_new_tmp_var("cont$idx");
    idx_id->up = id;
    idx_id->u_var.dflt_exp = exp_new_lit_i64(vector_size(&ir->fns), &idx_id->pos);
    meta_set_int32(&idx_id->u_var.dflt_exp->meta);
    meta_set_int32(&idx_id->meta);

    vector_add_first(&id->u_cont.blk->ids, idx_id);

    blk_trans(trans, id->u_cont.blk);

    ir->offset = 0;
}

static void
id_trans_interface(trans_t *trans, ast_id_t *id)
{
    int i;
    /* Index 0 is reserved for the constructor */
    int fn_idx = 1;
    ast_blk_t *blk = id->u_itf.blk;
    ir_t *ir = trans->ir;

    ASSERT(blk != NULL);

    vector_foreach(&blk->ids, i) {
        ast_id_t *fn_id = vector_get_id(&blk->ids, i);

        ASSERT1(is_fn_id(fn_id), fn_id->kind);
        ASSERT(!is_ctor_id(fn_id));

        /* If the interface type is used as a parameter, we can invoke it with the
         * interface function, so transform the parameter here and set abi */

        id_trans_param(trans, fn_id);

        fn_id->idx = fn_idx++;
        fn_id->abi = abi_lookup(&ir->abis, fn_id);
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
