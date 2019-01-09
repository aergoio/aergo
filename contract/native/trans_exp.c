/**
 * @file    trans_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ast_blk.h"
#include "ast_stmt.h"
#include "ir_bb.h"
#include "ir_fn.h"
#include "ir_sgmt.h"

#include "trans_exp.h"

static void
exp_trans_id_ref(trans_t *trans, ast_exp_t *exp)
{
    ast_id_t *id = exp->id;

    ASSERT(id != NULL);

    if (is_var_id(id)) {
        if (id->is_param || (!is_global_id(id) && !is_stack_id(id)))
            exp_set_local_ref(exp, id->idx);
        else
            /* the global variable always refers to the stack */
            exp_set_stack_ref(exp, id->addr, id->offset);
    }
    else if (is_fn_id(id)) {
        exp_set_lit(exp, NULL);

        value_set_i64(&exp->u_lit.val, id->idx);
    }
}

static void
exp_trans_lit(trans_t *trans, ast_exp_t *exp)
{
    int addr;
    value_t *val = &exp->u_lit.val;
    ir_sgmt_t *sgmt = &trans->ir->sgmt;

    switch (val->type) {
    case TYPE_BOOL:
    case TYPE_UINT64:
    case TYPE_DOUBLE:
        break;

    case TYPE_STRING:
        addr = sgmt_add_raw(sgmt, val_ptr(val), val_size(val) + 1);
        value_set_i64(val, addr);
        break;

    case TYPE_OBJECT:
        if (is_null_val(val))
            addr = 0;
        else
            addr = sgmt_add_raw(sgmt, val_ptr(val), val_size(val) + 1);
        value_set_i64(val, addr);
        break;

    default:
        ASSERT1(!"invalid value", val->type);
    }
}

static void
exp_trans_array(trans_t *trans, ast_exp_t *exp)
{
    ast_id_t *id = exp->id;

    exp_trans(trans, exp->u_arr.id_exp);
    exp_trans(trans, exp->u_arr.idx_exp);

    if (is_array_type(&id->meta)) {
        uint32_t offset;
        ast_exp_t *id_exp = exp->u_arr.id_exp;
        ast_exp_t *idx_exp = exp->u_arr.idx_exp;

        if (!is_stack_ref_exp(id_exp) || !is_lit_exp(idx_exp))
            return;

        /* The following arr_size is stripped arr_size */
        offset = val_i64(&idx_exp->u_lit.val) * exp->meta.arr_size;

        exp_set_stack_ref(exp, id_exp->u_stk.addr, id_exp->u_stk.offset + offset);
    }
    else {
        /* TODO
         * int addr = fn_add_stack_var(trans->fn);
         * ast_exp_t *call_exp = exp_new_call("$map_get", &exp->pos);
         *
         * bb_add_stmt(trans->bb, stmt_new_exp(call_exp, &exp->pos));
         *
         * return <return address of call>; */
    }
}

static void
exp_trans_cast(trans_t *trans, ast_exp_t *exp)
{
    exp_trans(trans, exp->u_cast.val_exp);

    if (!is_primitive_type(&exp->meta) || !is_primitive_type(&exp->u_cast.to_meta)) {
        /* TODO
         * int addr = fn_add_stack_var(trans->fn);
         * ast_exp_t *call_exp = exp_new_call("$concat", &exp->pos);
         *
         * bb_add_stmt(trans->bb, stmt_new_exp(call_exp, &exp->pos));
         *
         * return <return address of call>; */
    }
}

static void
exp_trans_unary(trans_t *trans, ast_exp_t *exp)
{
    ast_exp_t *val_exp = exp->u_un.val_exp;
    ast_exp_t *var_exp, *bi_exp, *lit_exp;

    switch (exp->u_un.kind) {
    case OP_INC:
    case OP_DEC:
        var_exp = exp_clone(val_exp);

        exp_trans(trans, var_exp);
        exp_trans(trans, val_exp);

        lit_exp = exp_new_lit(&exp->pos);
        value_set_i64(&lit_exp->u_lit.val, 1);

        bi_exp = exp_new_binary(exp->u_un.kind == OP_INC ? OP_ADD : OP_SUB, val_exp,
                                lit_exp, &exp->pos);

        meta_copy(&lit_exp->meta, &val_exp->meta);
        meta_copy(&bi_exp->meta, &val_exp->meta);

        bb_add_piggyback(trans->bb, stmt_new_assign(var_exp, bi_exp, &exp->pos));

        if (exp->u_un.is_prefix)
            *exp = *bi_exp;
        else
            *exp = *val_exp;
        break;

    case OP_NEG:
    case OP_NOT:
        exp_trans(trans, val_exp);
        break;

    default:
        ASSERT1(!"invalid operator", exp->u_un.kind);
    }
}

static void
exp_trans_binary(trans_t *trans, ast_exp_t *exp)
{
    exp_trans(trans, exp->u_bin.l_exp);
    exp_trans(trans, exp->u_bin.r_exp);

    if (exp->u_bin.kind == OP_ADD && is_string_type(&exp->meta)) {
        /* TODO
         * int addr = fn_add_stack();
         * ast_exp_t *call exp = exp_new_call("$concat", &exp->pos);
         * bb_add_stmt(trans->bb, stmt_new_exp(call_exp, &exp->pos));
         * return exp_new_stack(addr, offset, &exp->pos);
        */
    }
}

static void
exp_trans_ternary(trans_t *trans, ast_exp_t *exp)
{
    exp_trans(trans, exp->u_tern.pre_exp);
    exp_trans(trans, exp->u_tern.in_exp);
    exp_trans(trans, exp->u_tern.post_exp);

    if (is_lit_exp(exp->u_tern.pre_exp)) {
        /* maybe we should do this in optimizer */
        meta_t meta = exp->meta;

        if (val_bool(&exp->u_tern.pre_exp->u_lit.val))
            *exp = *exp->u_tern.in_exp;
        else
            *exp = *exp->u_tern.post_exp;

        meta_copy(&exp->meta, &meta);
    }
}

static void
exp_trans_access(trans_t *trans, ast_exp_t *exp)
{
    ast_exp_t *id_exp = exp->u_acc.id_exp;
    ast_id_t *qual_id = id_exp->id;
    ast_id_t *fld_id = exp->id;

    ASSERT(!is_fn_id(fld_id));

    exp_trans(trans, id_exp);

    if (is_fn_id(fld_id)) {
        if (is_itf_id(fld_id->up)) {
            int i;
            ast_id_t *cont_id = qual_id->meta.type_id;

            ASSERT1(is_object_type(&qual_id->meta), qual_id->meta.type);

            ASSERT(cont_id != NULL);
            ASSERT1(is_cont_id(cont_id), cont_id->kind);

            array_foreach(&cont_id->u_cont.blk->ids, i) {
                ast_id_t *fn_id = array_get_id(&cont_id->u_cont.blk->ids, i);

                if (is_fn_id(fn_id) && strcmp(fld_id->name, fn_id->name) == 0) {
                    ASSERT2(id_cmp(fld_id, fn_id), fld_id->kind, fn_id->kind);
                    exp->id = fn_id;
                    break;
                }
            }
        }
    }
    else if (is_stack_ref_exp(id_exp)) {
        exp_set_stack_ref(exp, id_exp->u_stk.addr, id_exp->u_stk.offset + fld_id->offset);
    }
}

static void
exp_trans_call(trans_t *trans, ast_exp_t *exp)
{
    int i;

    if (is_map_type(&exp->meta))
        return;

    /* since all return values of a function are treated as parameters,
     * the call expression is added as a separate statement.
     *
     * as a result, the "v = f();" statement changes to "f(); v = r;" */

    //exp_trans(trans, exp->u_call.id_exp);

    array_foreach(exp->u_call.param_exps, i) {
        exp_trans(trans, array_get_exp(exp->u_call.param_exps, i));
    }

    if (exp->id->u_fn.ret_id != NULL) {
        ast_id_t *ret_id = exp->id->u_fn.ret_id;

        /* if there is a return value,
         * we have to clone it because the expression itself is transformed */
        bb_add_stmt(trans->bb, stmt_new_exp(exp_clone(exp), &exp->pos));

        if (is_tuple_id(ret_id)) {
            int i;
            array_t *elem_exps = array_new();
            array_t *elem_ids = ret_id->u_tup.elem_ids;

            ASSERT(trans->fn != NULL);

            array_foreach(elem_ids, i) {
                ast_id_t *elem_id = array_get_id(elem_ids, i);
                ast_exp_t *ref_exp;

                ASSERT1(elem_id->offset == 0, elem_id->offset);

                fn_add_stack(trans->fn, elem_id);

                ref_exp = exp_new_stack_ref(elem_id->addr, 0, &exp->pos);
                meta_copy(&ref_exp->meta, &elem_id->meta);

                array_add_last(elem_exps, ref_exp);
            }

            exp->kind = EXP_TUPLE;
            exp->u_tup.elem_exps = elem_exps;
        }
        else {
            fn_add_stack(trans->fn, ret_id);

            ASSERT1(ret_id->offset == 0, ret_id->offset);

            exp_set_stack_ref(exp, ret_id->addr, 0);
        }
    }
    else {
        bb_add_stmt(trans->bb, stmt_new_exp(exp, &exp->pos));
    }
}

static void
exp_trans_sql(trans_t *trans, ast_exp_t *exp)
{
    /* TODO */
}

static void
exp_trans_tuple(trans_t *trans, ast_exp_t *exp)
{
    int i;
    array_t *elem_exps = exp->u_tup.elem_exps;

    array_foreach(elem_exps, i) {
        exp_trans(trans, array_get_exp(elem_exps, i));
    }
}

static void
exp_trans_init(trans_t *trans, ast_exp_t *exp)
{
    int i;
    array_t *elem_exps = exp->u_init.elem_exps;

    array_foreach(elem_exps, i) {
        exp_trans(trans, array_get_exp(elem_exps, i));
    }
}

void
exp_trans(trans_t *trans, ast_exp_t *exp)
{
    ASSERT(exp != NULL);

    switch (exp->kind) {
    case EXP_NULL:
        break;

    case EXP_ID_REF:
        exp_trans_id_ref(trans, exp);
        break;

    case EXP_LOCAL_REF:
    case EXP_STACK_REF:
        break;

    case EXP_LIT:
        exp_trans_lit(trans, exp);
        break;

    case EXP_ARRAY:
        exp_trans_array(trans, exp);
        break;

    case EXP_CAST:
        exp_trans_cast(trans, exp);
        break;

    case EXP_UNARY:
        exp_trans_unary(trans, exp);
        break;

    case EXP_BINARY:
        exp_trans_binary(trans, exp);
        break;

    case EXP_TERNARY:
        exp_trans_ternary(trans, exp);
        break;

    case EXP_ACCESS:
        exp_trans_access(trans, exp);
        break;

    case EXP_CALL:
        exp_trans_call(trans, exp);
        break;

    case EXP_SQL:
        exp_trans_sql(trans, exp);
        break;

    case EXP_TUPLE:
        exp_trans_tuple(trans, exp);
        break;

    case EXP_INIT:
        exp_trans_init(trans, exp);
        break;

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }
}

/* end of trans_exp.c */
