/**
 * @file    trans_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ast_blk.h"

#include "trans_exp.h"

static ast_exp_t *
exp_trans_ref(trans_t *trans, ast_exp_t *exp)
{
    ast_id_t *id = NULL;

    ASSERT1(is_ref_exp(exp), exp->kind);
    ASSERT(exp->u_ref.name != NULL);

    if (strcmp(exp->u_ref.name, "this") == 0) {
        id = trans->cont_id;
    }
    else if (trans->qual_id != NULL) {
        id = id_search_fld(trans->qual_id, exp->u_ref.name,
                           trans->cont_id == trans->qual_id);
    }
    else {
        if (trans->fn_id != NULL)
            id = id_search_param(trans->fn_id, exp->u_ref.name);

        if (id == NULL) {
            id = blk_search_id(trans->blk, exp->u_ref.name, exp->num);

            if (id != NULL && is_cont_id(id))
                /* search constructor */
                id = blk_search_id(id->u_cont.blk, exp->u_ref.name, exp->num);
        }
    }

    if (id == NULL)
        RETURN(ERROR_UNDEFINED_ID, &exp->pos, exp->u_ref.name);

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
    ast_exp_t *id_exp;
    meta_t *id_meta;
    ast_exp_t *idx_exp;
    meta_t *idx_meta;

    ASSERT1(is_array_exp(exp), exp->kind);
    ASSERT(exp->u_arr.id_exp != NULL);

    id_exp = exp->u_arr.id_exp;
    id_meta = &id_exp->meta;

    CHECK(exp_trans(trans, id_exp));

    exp->id = id_exp->id;
    ASSERT(exp->id != NULL);

    idx_exp = exp->u_arr.idx_exp;
    idx_meta = &idx_exp->meta;

    CHECK(exp_trans(trans, idx_exp));

    if (is_array_type(id_meta)) {
        ASSERT(id_meta->arr_dim > 0);
        ASSERT(id_meta->arr_size != NULL);

        if (!is_integer_type(idx_meta))
            RETURN(ERROR_INVALID_SIZE_VAL, &idx_exp->pos, meta_to_str(idx_meta));

        meta_copy(&exp->meta, id_meta);

        if (is_lit_exp(idx_exp)) {
            ASSERT(id_meta->arr_size != NULL);

            /* arr_size[0] can be negative if array is used as a parameter */
            if (id_meta->arr_size[0] > 0 &&
                val_i64(&idx_exp->u_lit.val) >= (uint)id_meta->arr_size[0])
                RETURN(ERROR_INVALID_ARR_IDX, &idx_exp->pos);
        }

        exp->meta.arr_dim--;
        ASSERT(exp->meta.arr_dim >= 0);

        if (exp->meta.arr_dim == 0)
            exp->meta.arr_size = NULL;
        else
            exp->meta.arr_size = &exp->meta.arr_size[1];
    }
    else {
        if (!is_map_type(id_meta))
            RETURN(ERROR_INVALID_SUBSCRIPT, &id_exp->pos);

        CHECK(meta_cmp(id_meta->elems[0], idx_meta));
        meta_copy(&exp->meta, id_meta->elems[1]);
    }
}

static ast_exp_t *
exp_trans_cast(trans_t *trans, ast_exp_t *exp)
{
    exp->u_cast.val_exp = exp_trans(trans, exp->u_un.val_exp);

    if (!is_primitive_type(&exp->meta) || !is_primitive_type(&exp->u_cast.to_meta)) {
        /* int addr = fn_add_tmp_var(trans->fn);
         * ast_exp_t *call_exp = exp_new_call(..., &exp->pos);
         *
         * bb_add_stmt(trans->bb, stmt_new_exp(call_exp, &exp->pos));
         *
         * return ???; */
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
            trans->post_exp = exp;

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
    op_kind_t op = exp->u_bin.kind;

    exp->u_bin.l_exp = exp_trans(trans, exp->u_bin.l_exp);
    exp->u_bin.r_exp = exp_trans(trans, exp->u_bin.r_exp);

    if (exp->u_bin.kind == OP_ADD && is_string_type(&exp->meta)) {
        /*
         * int addr = fn_add_stack_var();
         * make call exp
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

    case EXP_REF:
        return exp_trans_ref(trans, exp);

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
