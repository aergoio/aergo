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
    if (is_global_id(id))
        /* Initialization of the global variable will be done in the constructor */
        return;

    ASSERT(trans->fn != NULL);

    fn_add_local(trans->fn, id);
}

static void
id_trans_ctor(trans_t *trans, ast_id_t *id)
{
    int i, j;
    ast_id_t *addr_id;
    ast_exp_t *l_exp, *r_exp, *v_exp;
    array_t *stmts = array_new();
    ir_fn_t *fn = trans->fn;
    ir_t *ir = trans->ir;

    /* The parameter of the constructor is immutable */
    fn->abi = abi_lookup(&ir->abis, id);

    /* All global variables access memory by adding relative offset to this value */
    addr_id = id_new_tmp_var("cont$addr");
    addr_id->up = id;
    meta_set_int32(&addr_id->meta);

    fn_add_local(fn, addr_id);
    fn->cont_idx = addr_id->idx;

    l_exp = exp_new_local(TYPE_UINT32, addr_id->idx);
    r_exp = exp_new_global("heap$offset");

    stmt_add(stmts, stmt_new_assign(l_exp, r_exp, &id->pos));

    /* We use the "stmts" array to keep the declaration order of variables */
    array_foreach(&id->up->u_cont.blk->ids, i) {
        ast_id_t *var_id = array_get_id(&id->up->u_cont.blk->ids, i);
        ast_exp_t *dflt_exp = NULL;

        if (is_var_id(var_id)) {
            ir_add_heap(trans->ir, &var_id->meta, fn->cont_idx);

            dflt_exp = var_id->u_var.dflt_exp;
        }
        else if (is_tuple_id(var_id)) {
            array_foreach(var_id->u_tup.elem_ids, j) {
                ast_id_t *elem_id = array_get_id(var_id->u_tup.elem_ids, j);

                ir_add_heap(trans->ir, &elem_id->meta, fn->cont_idx);
            }

            dflt_exp = var_id->u_tup.dflt_exp;
        }

        if (dflt_exp != NULL)
            stmt_add(stmts, stmt_make_assign(var_id, dflt_exp));
    }

    /* Increase "heap$offset" by the amount of memory used by the global variables
     * defined in the contract */
    l_exp = exp_new_local(TYPE_UINT32, addr_id->idx);

    v_exp = exp_new_lit_i64(ir->offset, &id->pos);
    meta_set_int32(&v_exp->meta);

    r_exp = exp_new_binary(OP_ADD, l_exp, v_exp, &id->pos);

    stmt_add(stmts, stmt_new_assign(exp_new_global("heap$offset"), r_exp, &id->pos));

    if (!is_empty_array(stmts)) {
        if (id->u_fn.blk == NULL)
            id->u_fn.blk = blk_new_fn(&id->pos);

        array_join_first(&id->u_fn.blk->stmts, stmts);
    }
}

static void
id_trans_param(trans_t *trans, ast_id_t *id)
{
    ast_id_t *param_id;

    /* All functions that are not constructors must be added the contract address as
     * the first argument, and must also be added to the param_ids to reflect the abi */

    param_id = id_new_tmp_var("cont$addr");
    param_id->u_var.kind = PARAM_IN;
    param_id->up = id;
    meta_set_object(&param_id->meta, id->up);

    if (id->u_fn.param_ids == NULL)
        id->u_fn.param_ids = array_new();

    array_add_first(id->u_fn.param_ids, param_id);
}

static void
set_stack_addr(trans_t *trans, src_pos_t *pos)
{
    ast_exp_t *l_exp, *r_exp, *v_exp;
    ir_fn_t *fn = trans->fn;

    /* At the beginning of "entry_bb", set the current stack offset to the local
     * variable */
    l_exp = exp_new_local(TYPE_UINT32, fn->stack_idx);

    v_exp = exp_new_lit_i64(ALIGN64(fn->usage), pos);
    meta_set_int32(&v_exp->meta);

    r_exp = exp_new_binary(OP_SUB, exp_new_global("stack$offset"), v_exp, pos);

    array_add_first(&fn->entry_bb->stmts, stmt_new_assign(l_exp, r_exp, pos));

    /* If there is any stack variable in the function, it has to be restored to the
     * original value at the end of "exit_bb" because "stack$offset" has been changed */
    l_exp = exp_new_global("stack$offset");
    r_exp = exp_new_local(TYPE_UINT32, fn->stack_idx);

    array_add_last(&fn->exit_bb->stmts, stmt_new_assign(l_exp, r_exp, pos));
}

static void
id_trans_fn(trans_t *trans, ast_id_t *id)
{
    ast_id_t *ret_id = id->u_fn.ret_id;
    ir_fn_t *fn = fn_new(id);
    ir_t *ir = trans->ir;

    ASSERT(id->up != NULL);
    ASSERT1(is_cont_id(id->up), id->up->kind);

    trans->fn = fn;

    if (is_ctor_id(id)) {
        id_trans_ctor(trans, id);
    }
    else {
        id_trans_param(trans, id);

        /* The "cont_idx" is always 0 because it is prepended to parameters */
        fn->cont_idx = 0;
        fn->abi = abi_lookup(&ir->abis, id);
    }

    id->abi = fn->abi;

    /* All stack variables access memory by adding relative offset to this value */
    fn->stack_idx = fn_add_tmp_var(fn, "stack$addr", TYPE_UINT32);

    if (ret_id != NULL)
        /* All return values are stored in this variable */
        fn->ret_idx = fn_add_tmp_var(fn, "return$val", ret_id->meta.type);

    /* It is used internally for binaryen, not for us (see fn_gen()) */
    fn->reloop_idx = fn_add_tmp_var(fn, "relooper$helper", TYPE_INT32);

    trans->bb = fn->entry_bb;

    if (id->u_fn.blk != NULL)
        blk_trans(trans, id->u_fn.blk);

    if (fn->usage > 0)
        set_stack_addr(trans, &id->pos);

    if (trans->bb != NULL) {
        bb_add_branch(trans->bb, NULL, fn->exit_bb);
        fn_add_basic_blk(fn, trans->bb);
    }

    fn_add_basic_blk(fn, fn->exit_bb);

    if (ret_id != NULL) {
        /* The contract address or return value is returned at the end of "exit_bb" */
        ast_exp_t *arg_exp;
        ast_stmt_t *ret_stmt;

        if (is_ctor_id(id))
            arg_exp = exp_new_local(TYPE_UINT32, fn->cont_idx);
        else
            arg_exp = exp_new_local(ret_id->meta.type, fn->ret_idx);

        ret_stmt = stmt_new_return(arg_exp, &id->pos);
        ret_stmt->u_ret.ret_id = ret_id;

        array_add_last(&fn->exit_bb->stmts, ret_stmt);
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

    /* Move the constructor to the first position because it handles the memory
     * allocation of global variables. Even if the contract implements interface,
     * there is no problem because index 0 is empty. (see id_trans_interface()) */
    array_foreach(&blk->ids, i) {
        if (is_ctor_id(array_get_id(&blk->ids, i))) {
            array_move(&blk->ids, i, 0);
            break;
        }
    }

    /* Because the cross-reference is possible between functions, the index of the
     * function is numbered before transformation */
    array_foreach(&blk->ids, i) {
        ast_id_t *fn_id = array_get_id(&blk->ids, i);

        if (is_ctor_id(fn_id))
            /* Since the constructor can be called from any location (including another
             * contracts), it should always be accessed with an absolute index */
            fn_id->idx = array_size(&ir->fns);
        else if (is_fn_id(fn_id))
            /* The "idx" is the relative index within the contract */
            fn_id->idx = fn_idx++;
    }

    /* This value, like any other global variable, is stored in the heap area used by
     * the contract, and is stored in the first 4 bytes of the area. All functions
     * also access table by adding relative index to this value */
    idx_id = id_new_tmp_var("cont$idx");
    idx_id->up = id;
    idx_id->u_var.dflt_exp = exp_new_lit_i64(array_size(&ir->fns), &idx_id->pos);
    meta_set_int32(&idx_id->u_var.dflt_exp->meta);
    meta_set_int32(&idx_id->meta);

    array_add_first(&id->u_cont.blk->ids, idx_id);

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

    array_foreach(&blk->ids, i) {
        ast_id_t *fn_id = array_get_id(&blk->ids, i);

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

    array_foreach(id->u_tup.elem_ids, i) {
        id_trans_var(trans, array_get_id(id->u_tup.elem_ids, i));
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
