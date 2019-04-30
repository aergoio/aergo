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
#include "syslib.h"

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

    /* All functions that are not constructors must be added the contract address as the first
     * argument, and must also be added to the param_ids to reflect the abi */

    param_id = id_new_tmp_var("cont$addr", TYPE_OBJECT, &id->pos);
    param_id->u_var.is_param = true;
    param_id->up = id;
    meta_set_object(&param_id->meta, id->up);

    if (id->u_fn.param_ids == NULL)
        id->u_fn.param_ids = vector_new();

    vector_add_first(id->u_fn.param_ids, param_id);
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

    if (fn->heap_usage > 0)
        stmt_trans(trans, stmt_make_malloc(fn->heap_idx, fn->heap_usage, &id->pos));

    vector_foreach(stmts, i) {
        stmt_trans(trans, vector_get_stmt(stmts, i));
    }

    trans->bb = NULL;
}

static void
set_stack_addr(trans_t *trans, ast_id_t *id)
{
    ir_fn_t *fn = trans->fn;
    src_pos_t *pos = &id->pos;
    ast_exp_t *stk_exp, *reg_exp;
    ast_exp_t *val_exp, *bin_exp;
    ast_exp_t *max_exp, *cond_exp, *call_exp;
    ast_blk_t *if_blk;

    if (fn->stack_usage == 0)
        return;

    /* At the beginning of "entry_bb", set the current stack offset to the register */
    reg_exp = exp_new_reg(fn->stack_idx);
    meta_set_int32(&reg_exp->meta);

    stk_exp = exp_new_global("stack_top");

    /* Make "r1 = stack_top" */
    vector_add(&fn->entry_bb->stmts, 0, stmt_new_assign(reg_exp, stk_exp, pos));

    val_exp = exp_new_lit_int(fn->stack_usage, pos);
    meta_set_int32(&val_exp->meta);

    bin_exp = exp_new_binary(OP_ADD, reg_exp, val_exp, pos);
    meta_set_int32(&bin_exp->meta);

    /* Make "stack_top = r1 + usage" */
    vector_add(&fn->entry_bb->stmts, 1, stmt_new_assign(stk_exp, bin_exp, pos));

    max_exp = exp_new_global("stack_max");
    meta_set_int32(&max_exp->meta);

    cond_exp = exp_new_binary(OP_GE, stk_exp, max_exp, pos);
    meta_set_int32(&cond_exp->meta);

    if_blk = blk_new_normal(&id->pos);

    call_exp = exp_new_call(FN_STACK_OVF, NULL, NULL, pos);
    meta_set_void(&call_exp->meta);

    exp_trans(trans, call_exp);

    stmt_add(&if_blk->stmts, stmt_new_exp(call_exp, pos));

    /* Make "if (stack_top >= stack_max) __stack_overflow();" */
    vector_add(&fn->entry_bb->stmts, 2, stmt_new_if(cond_exp, if_blk, pos));

    /* If there is any stack variable in the function, it has to be restored to the original value
     * at the end of "exit_bb" because "stack_top" has been changed */
    vector_add_last(&fn->exit_bb->stmts, stmt_new_assign(stk_exp, reg_exp, pos));
}

static void
check_return_stmt(ir_fn_t *fn, src_pos_t *pos)
{
    int i, j;

    vector_foreach(&fn->bbs, i) {
        ir_bb_t *bb = vector_get_bb(&fn->bbs, i);

        /* except unreachable block */
        if (bb->ref_cnt > 0) {
            vector_foreach(&bb->brs, j) {
                ast_stmt_t *stmt;
                ir_br_t *br = vector_get_br(&bb->brs, j);

                if (br->bb != fn->exit_bb)
                    continue;

                stmt = vector_get_last(&bb->stmts, ast_stmt_t);
                if (stmt == NULL || !is_return_stmt(stmt)) {
                    ERROR(ERROR_MISSING_RETURN, pos);
                    return;
                }
            }
        }
    }
}

static void
id_trans_fn(trans_t *trans, ast_id_t *id)
{
    ast_id_t *ret_id = id->u_fn.ret_id;
    meta_t addr_meta;
    ir_md_t *md = trans->md;
    ir_fn_t *fn;

    ASSERT(id->up != NULL);
    ASSERT1(is_cont_id(id->up), id->up->kind);
    ASSERT(md != NULL);

    id_trans_param(trans, id);

    fn = fn_new(id);

    meta_set_int32(&addr_meta);

    /* All heap variables access memory by adding relative offset to this register */
    fn->heap_idx = fn_add_register(fn, &addr_meta);

    /* All stack variables access memory by adding relative offset to this register */
    fn->stack_idx = fn_add_register(fn, &addr_meta);

    if (ret_id != NULL)
        /* All return values are stored in this register */
        fn->ret_idx = fn_add_register(fn, &ret_id->meta);

    /* It is used internally for binaryen, not for us (see fn_gen()) */
    fn->reloop_idx = fn_add_register(fn, &addr_meta);

    trans->fn = fn;

    if (is_ctor_id(id))
        id_trans_ctor(trans, id);
    else
        /* The "cont_idx" is always 0 because it is prepended to parameters */
        fn->cont_idx = 0;

    trans->bb = fn->entry_bb;

    if (id->u_fn.blk != NULL)
        blk_trans(trans, id->u_fn.blk);

    set_stack_addr(trans, id);

    if (trans->bb != NULL) {
        bb_add_branch(trans->bb, NULL, fn->exit_bb);
        fn_add_basic_blk(fn, trans->bb);
    }

    fn_add_basic_blk(fn, fn->exit_bb);

    if (ret_id != NULL) {
        ast_exp_t *arg_exp;
        ast_stmt_t *ret_stmt;

        if (is_ctor_id(id)) {
            /* The contract address is returned at the end of "exit_bb" */
            arg_exp = exp_new_reg(fn->cont_idx);
            meta_set_int32(&arg_exp->meta);

            ret_stmt = stmt_new_return(arg_exp, &id->pos);
            ret_stmt->u_ret.ret_id = ret_id;

            vector_add_last(&fn->exit_bb->stmts, ret_stmt);
        }
        else {
            check_return_stmt(fn, &id->pos);
        }
    }

    trans->fn = NULL;
    trans->bb = NULL;

    if (is_public_id(id) || id->is_used)
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

        /* If the interface type is used as a parameter, we can invoke it with the interface
         * function, so transform the parameter here */

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
    case ID_LIB:
        break;

    default:
        ASSERT1(!"invalid identifier", id->kind);
    }
}

/* end of trans_id.c */
