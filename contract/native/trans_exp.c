/**
 * @file    trans_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ast_blk.h"

#include "trans_exp.h"

static ast_exp_t *
exp_trans_id_ref(trans_t *trans, ast_exp_t *exp)
{
    ast_id_t *id = NULL;

    ASSERT1(is_id_ref_exp(exp), exp->kind);
    ASSERT(exp->u_id.name != NULL);

    if (strcmp(exp->u_id.name, "this") == 0) {
        id = trans->cont_id;
    }
    else if (trans->qual_id != NULL) {
        id = id_search_fld(trans->qual_id, exp->u_id.name,
                           trans->cont_id == trans->qual_id);
    }
    else {
        if (trans->fn_id != NULL)
            id = id_search_param(trans->fn_id, exp->u_id.name);

        if (id == NULL) {
            id = blk_search_id(trans->blk, exp->u_id.name, exp->num);

            if (id != NULL && is_cont_id(id))
                /* search constructor */
                id = blk_search_id(id->u_cont.blk, exp->u_id.name, exp->num);
        }
    }

    if (id == NULL)
        RETURN(ERROR_UNDEFINED_ID, &exp->pos, exp->u_id.name);

    ASSERT(id->is_transed);

    id->is_used = true;

    if (is_const_id(id) && id->val != NULL) {
        exp->kind = EXP_LIT;
        exp->u_lit.val = *id->val;
    }
    else {
        exp->id = id;
    }

    meta_copy(&exp->meta, &id->meta);
}

static ast_exp_t *
exp_trans_array(trans_t *trans, ast_exp_t *exp)
{
    ast_id_t *id = exp->id;

    exp->u_arr.id_exp = exp_trans(trans, exp->u_arr.id_exp);
    exp->u_arr.idx_exp = exp_trans(trans, exp->u_arr.idx_exp);

    if (is_array_type(&id->meta)) {
    }
    else {
        /* int addr = fn_add_stack_var(trans->fn);
         * ast_exp_t *call_exp = exp_new_call("$map_get", &exp->pos);
         *
         * bb_add_stmt(trans->bb, stmt_new_exp(call_exp, &exp->pos));
         *
         * return <return address of call>; */
    }
}

static ast_exp_t *
exp_trans_cast(trans_t *trans, ast_exp_t *exp)
{
    exp->u_cast.val_exp = exp_trans(trans, exp->u_un.val_exp);

    if (!is_primitive_type(&exp->meta) || !is_primitive_type(&exp->u_cast.to_meta)) {
        /* int addr = fn_add_stack_var(trans->fn);
         * ast_exp_t *call_exp = exp_new_call("$concat", &exp->pos);
         *
         * bb_add_stmt(trans->bb, stmt_new_exp(call_exp, &exp->pos));
         *
         * return <return address of call>; */
    }

    return exp;
}

static ast_exp_t *
exp_trans_unary(trans_t *trans, ast_exp_t *exp)
{
    exp->u_un.val_exp = exp_trans(trans, exp->u_un.val_exp);

    switch (exp->u_un.kind) {
    case OP_INC:
    case OP_DEC:
        if (exp->u_un.is_prefix)
            bb_add_stmt(trans->bb, stmt_new_exp(exp, &exp->pos));
        else
            bb_set_piggyback(trans->bb, stmt_new_exp(exp, &exp->pos));

        return exp->u_un.val_exp;

    case OP_NEG:
    case OP_NOT:
        break;

    default:
        ASSERT1(!"invalid operator", exp->u_un.kind);
    }

    return exp;
}

static ast_exp_t *
exp_trans_binary(trans_t *trans, ast_exp_t *exp)
{
    exp->u_bin.l_exp = exp_trans(trans, exp->u_bin.l_exp);
    exp->u_bin.r_exp = exp_trans(trans, exp->u_bin.r_exp);

    if (exp->u_bin.kind == OP_ADD && is_string_type(&exp->meta)) {
        /*
         * int addr = fn_add_stack_var();
         * ast_exp_t *call exp = exp_new_call("$concat", &exp->pos);
         * bb_add_stmt(trans->bb, stmt_new_exp(call_exp, &exp->pos));
         * return exp_new_stack(addr, offset, &exp->pos);
        */
    }

    return exp;
}

static ast_exp_t *
exp_trans_ternary(trans_t *trans, ast_exp_t *exp)
{
    ast_exp_t *pre_exp = exp->u_tern.pre_exp;
    ast_exp_t *in_exp = exp->u_tern.in_exp;
    ast_exp_t *post_exp = exp->u_tern.post_exp;

    if (is_lit_exp(pre_exp)) {
        if (val_bool(&pre_exp->u_lit.val))
            return exp_trans(trans, exp->u_tern.in_exp);
        else
            return exp_trans(trans, exp->u_tern.post_exp);
    }

    exp->u_tern.pre_exp = exp_trans(trans, pre_exp);
    exp->u_tern.in_exp = exp_trans(trans, in_exp);
    exp->u_tern.post_exp = exp_trans(trans, post_exp);

    return exp;
}

static ast_exp_t *
exp_trans_access(trans_t *trans, ast_exp_t *exp)
{
    ast_exp_t *id_exp = exp->u_acc.id_exp;
    ast_exp_t *fld_exp = exp->u_acc.fld_exp;

    exp->u_acc.id_exp = exp_trans(trans, id_exp);
    exp->u_acc.fld_exp = exp_trans(trans, fld_exp);

    /*
    if (is_fn_id(exp->id))
        return exp;

    ASSERT1(is_stack_exp(exp->u_acc.id_exp), exp->u_acc.id_exp->kind);

    exp->u_acc.id_exp->offset = exp->u_acc.fld_exp->offset;

    return exp->u_acc.id_exp;
    */

    return exp;
}

static ast_exp_t *
exp_trans_call(trans_t *trans, ast_exp_t *exp)
{
    bb_add_stmt(trans->bb, stmt_new_exp(exp, &exp->pos));

    /* ??? */
    return exp;
}

static ast_exp_t *
exp_trans_sql(trans_t *trans, ast_exp_t *exp)
{
    return exp;
}

static ast_exp_t *
exp_trans_tuple(trans_t *trans, ast_exp_t *exp)
{
    return exp;
}

static ast_exp_t *
exp_trans_init(trans_t *trans, ast_exp_t *exp)
{
    return exp;
}

ast_exp_t *
exp_trans(trans_t *trans, ast_exp_t *exp)
{
    ASSERT(exp != NULL);

    switch (exp->kind) {
    case EXP_NULL:
        return exp;

    case EXP_ID_REF:
        return exp_trans_id_ref(trans, exp);

    case EXP_LIT:
        return exp;

    case EXP_ARRAY:
        return exp_trans_array(trans, exp);

    case EXP_CAST:
        return exp_trans_cast(trans, exp);

    case EXP_UNARY:
        return exp_trans_unary(trans, exp);

    case EXP_BINARY:
        return exp_trans_binary(trans, exp);

    case EXP_TERNARY:
        return exp_trans_ternary(trans, exp);

    case EXP_ACCESS:
        return exp_trans_access(trans, exp);

    case EXP_CALL:
        return exp_trans_call(trans, exp);

    case EXP_SQL:
        return exp_trans_sql(trans, exp);

    case EXP_TUPLE:
        return exp_trans_tuple(trans, exp);

    case EXP_INIT:
        return exp_trans_init(trans, exp);

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }
}

/* end of trans_exp.c */
