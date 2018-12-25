/**
 * @file    trans_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ast_blk.h"

#include "trans_exp.h"

static void
exp_trans_id_ref(trans_t *trans, ast_exp_t *exp)
{
    ast_id_t *id = exp->id;

    if (!is_var_id(id) || is_global_id(id))
        /* nothing to do */
        return;

    if (is_local_id(id)) {
        exp->kind = EXP_LOCAL_REF;
        exp->u_lo.index = id->idx;
    }
    else {
        exp->kind = EXP_STACK_REF;
        exp->u_st.addr = id->addr;
        exp->u_st.offset = id->offset;
    }
}

static void
exp_trans_array(trans_t *trans, ast_exp_t *exp)
{
    ast_id_t *id = exp->id;

    exp_trans(trans, exp->u_arr.id_exp);
    exp_trans(trans, exp->u_arr.idx_exp);

    if (!is_array_type(&id->meta)) {
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
    exp_trans(trans, exp->u_un.val_exp);

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
    exp_trans(trans, exp->u_un.val_exp);

    switch (exp->u_un.kind) {
    case OP_INC:
    case OP_DEC:
        if (exp->u_un.is_prefix)
            bb_add_stmt(trans->bb, stmt_new_exp(exp, &exp->pos));
        else
            bb_set_piggyback(trans->bb, stmt_new_exp(exp, &exp->pos));

        *exp = *exp->u_un.val_exp;
        break;

    case OP_NEG:
    case OP_NOT:
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
        if (val_bool(&pre_exp->u_lit.val))
            *exp = *exp->u_tern.in_exp;
        else
            *exp = *exp->u_tern.post_exp;
    }
}

static void
exp_trans_access(trans_t *trans, ast_exp_t *exp)
{
    exp_trans(trans, exp->u_acc.id_exp);
    exp_trans(trans, exp->u_acc.fld_exp);

    if (is_fn_id(exp->id))
        return;

    /*
    ASSERT1(is_stack_exp(exp->u_acc.id_exp), exp->u_acc.id_exp->kind);

    exp->u_acc.id_exp->offset = exp->u_acc.fld_exp->offset;

    return exp->u_acc.id_exp;
    */

    return exp;
}

static void
exp_trans_call(trans_t *trans, ast_exp_t *exp)
{
    bb_add_stmt(trans->bb, stmt_new_exp(exp, &exp->pos));
}

static void
exp_trans_sql(trans_t *trans, ast_exp_t *exp)
{
}

static void
exp_trans_tuple(trans_t *trans, ast_exp_t *exp)
{
}

static void
exp_trans_init(trans_t *trans, ast_exp_t *exp)
{
}

void
exp_trans(trans_t *trans, ast_exp_t *exp)
{
    ASSERT(exp != NULL);

    switch (exp->kind) {
    case EXP_NULL:
        return;

    case EXP_ID_REF:
        exp_trans_id_ref(trans, exp);

    case EXP_LIT:
        exp;

    case EXP_ARRAY:
        exp_trans_array(trans, exp);

    case EXP_CAST:
        exp_trans_cast(trans, exp);

    case EXP_UNARY:
        exp_trans_unary(trans, exp);

    case EXP_BINARY:
        exp_trans_binary(trans, exp);

    case EXP_TERNARY:
        exp_trans_ternary(trans, exp);

    case EXP_ACCESS:
        exp_trans_access(trans, exp);

    case EXP_CALL:
        exp_trans_call(trans, exp);

    case EXP_SQL:
        exp_trans_sql(trans, exp);

    case EXP_TUPLE:
        exp_trans_tuple(trans, exp);

    case EXP_INIT:
        exp_trans_init(trans, exp);

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }
}

/* end of trans_exp.c */
