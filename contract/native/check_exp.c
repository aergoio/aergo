/**
 * @file    check_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"

#include "check_exp.h"

static int
exp_id_check(check_t *check, ast_exp_t *exp)
{
    ast_id_t *id;

    ASSERT1(is_id_exp(exp), exp->kind);

    if (check->aq_id != NULL) {
        id = ast_id_search_fld(check->aq_id, exp->num, exp->u_id.name);
    }
    else {
        id = ast_id_search_blk(check->blk, exp->num, exp->u_id.name);

        if (id != NULL && is_contract_id(id))
            id = ast_id_search_fld(id, exp->num, exp->u_id.name);
    }

    if (id == NULL)
        THROW(ERROR_UNDEFINED_ID, &exp->trc, exp->u_id.name);

    id->is_used = true;

    exp->id = id;
    exp->meta = id->meta;

    return NO_ERROR;
}

static int
exp_lit_check(check_t *check, ast_exp_t *exp)
{
    ASSERT1(is_lit_exp(exp), exp->kind);

    switch (exp->u_lit.val.kind) {
    case VAL_NULL:
        meta_set_prim(&exp->meta, TYPE_REF);
        break;
    case VAL_BOOL:
        meta_set_prim(&exp->meta, TYPE_BOOL);
        break;
    case VAL_INT:
        meta_set_dynamic(&exp->meta, TYPE_INT64);
        break;
    case VAL_FP:
        meta_set_dynamic(&exp->meta, TYPE_DOUBLE);
        break;
    case VAL_STR:
        meta_set_prim(&exp->meta, TYPE_STRING);
        break;
    default:
        ASSERT1(!"invalid value", exp->u_lit.val.kind);
    }

    return NO_ERROR;
}

static int
exp_type_check(check_t *check, ast_exp_t *exp)
{
    ASSERT1(is_type_exp(exp), exp->kind);
    ASSERT1(is_valid_type(exp->u_type.type), exp->u_type.type);

    if (exp->u_type.type == TYPE_STRUCT) {
        ast_id_t *id;

        ASSERT(exp->u_type.name != NULL);
        ASSERT(exp->u_type.k_exp == NULL);
        ASSERT(exp->u_type.v_exp == NULL);

        if (check->aq_id != NULL)
            id = ast_id_search_fld(check->aq_id, exp->num, exp->u_type.name);
        else
            id = ast_id_search_blk(check->blk, exp->num, exp->u_type.name);

        if (id == NULL || (!is_struct_id(id) && !is_contract_id(id)))
            THROW(ERROR_UNDEFINED_TYPE, &exp->trc, exp->u_type.name);

        exp->id = id;
        meta_set_prim(&exp->meta, is_struct_id(id) ? TYPE_STRUCT : TYPE_REF);
    }
    else if (exp->u_type.type == TYPE_MAP) {
        ast_exp_t *k_exp, *v_exp;
        meta_t *k_meta, *v_meta;

        ASSERT1(exp->u_type.name == NULL, exp->u_type.name);
        ASSERT(exp->u_type.k_exp != NULL);
        ASSERT(exp->u_type.v_exp != NULL);

        k_exp = exp->u_type.k_exp;
        k_meta = &k_exp->meta;

        TRY(exp_type_check(check, k_exp));

        if (!is_comparable_meta(k_meta))
            THROW(ERROR_INVALID_KEY_TYPE, &k_exp->trc, TYPE_NAME(k_meta->type));

        v_exp = exp->u_type.v_exp;
        v_meta = &v_exp->meta;

        TRY(exp_type_check(check, v_exp));

        ASSERT(!is_tuple_meta(v_meta));
        meta_set_map(&exp->meta, k_meta->type, v_meta);
    }
    else {
        ASSERT1(exp->u_type.name == NULL, exp->u_type.name);
        ASSERT(exp->u_type.k_exp == NULL);
        ASSERT(exp->u_type.v_exp == NULL);

        meta_set_prim(&exp->meta, exp->u_type.type);
    }

    return NO_ERROR;
}

static int
exp_array_check(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *id_exp;
    meta_t *id_meta;

    ASSERT1(is_array_exp(exp), exp->kind);
    ASSERT(exp->u_arr.id_exp != NULL);

    id_exp = exp->u_arr.id_exp;
    id_meta = &id_exp->meta;

    TRY(check_exp(check, id_exp));

    if (id_exp->id == NULL || !is_array_meta(id_meta) || !is_map_meta(id_meta))
        THROW(ERROR_INVALID_SUBSCRIPT, &id_exp->trc);

    exp->id = id_exp->id;

    if (exp->u_arr.idx_exp != NULL) {
        ast_exp_t *idx_exp = exp->u_arr.idx_exp;
        meta_t *param_meta = &idx_exp->meta;

        TRY(check_exp(check, idx_exp));

        // TODO: restriction of array size
        if (!is_integer_meta(param_meta))
            THROW(ERROR_INVALID_IDX_TYPE, &idx_exp->trc,
                  TYPE_NAME(param_meta->type));
    }

    exp->meta = *id_meta;
    exp->meta.is_array = true;

    return NO_ERROR;
}

static int
exp_op_check_assign(check_t *check, ast_exp_t *exp)
{
    int i;
    ast_exp_t *l_exp, *r_exp;
    meta_t *l_meta, *r_meta;

    ASSERT(exp->u_op.l_exp != NULL);
    ASSERT(exp->u_op.r_exp != NULL);

    l_exp = exp->u_op.l_exp;
    l_meta = &l_exp->meta;

    TRY(check_exp(check, l_exp));

    if (is_tuple_exp(l_exp)) {
        int i;
        array_t *tuple = l_exp->u_tup.exps;;

        for (i = 0; i < array_size(tuple); i++) {
            ast_exp_t *item = array_item(tuple, i, ast_exp_t);

            if (!is_usable_lval(item))
                THROW(ERROR_INVALID_LVALUE, &item->trc);
        }
    }
    else if (!is_usable_lval(l_exp)) {
        THROW(ERROR_INVALID_LVALUE, &l_exp->trc);
    }

    r_exp = exp->u_op.r_exp;
    r_meta = &r_exp->meta;

    TRY(check_exp(check, r_exp));

    if (!meta_equals(l_meta, r_meta))
        THROW(ERROR_MISMATCHED_TYPE, &r_exp->trc, TYPE_NAME(l_meta->type),
              TYPE_NAME(r_meta->type));

    if (is_lit_exp(r_exp) && 
        !value_check_range(&r_exp->u_lit.val, l_meta->type))
        THROW(ERROR_NUMERIC_OVERFLOW, &r_exp->trc, TYPE_NAME(l_meta->type));

    meta_set_from(&exp->meta, l_meta, r_meta);

    return NO_ERROR;
}

static int
exp_op_check_arith(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *l_exp, *r_exp;
    meta_t *l_meta, *r_meta;

    ASSERT(exp->u_op.l_exp != NULL);
    ASSERT(exp->u_op.r_exp != NULL);

    l_exp = exp->u_op.l_exp;
    l_meta = &l_exp->meta;

    TRY(check_exp(check, l_exp));

    r_exp = exp->u_op.r_exp;
    r_meta = &r_exp->meta;

    TRY(check_exp(check, r_exp));

    if (!meta_equals(l_meta, r_meta))
        THROW(ERROR_MISMATCHED_TYPE, &exp->trc, TYPE_NAME(l_meta->type),
              TYPE_NAME(r_meta->type));

    if (exp->u_op.kind == OP_ADD) {
        if (!is_integer_meta(l_meta) && !is_float_meta(l_meta) &&
            !is_string_meta(l_meta))
            THROW(ERROR_INVALID_OP_TYPE, &exp->trc, TYPE_NAME(l_meta->type));
    }
    else if (!is_integer_meta(l_meta) && !is_float_meta(l_meta)) {
        THROW(ERROR_INVALID_OP_TYPE, &exp->trc, TYPE_NAME(l_meta->type));
    }

    meta_set_from(&exp->meta, l_meta, r_meta);

    return NO_ERROR;
}

static int
exp_op_check_bool_cmp(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *l_exp, *r_exp;
    meta_t *l_meta, *r_meta;

    ASSERT(exp->u_op.l_exp != NULL);
    ASSERT(exp->u_op.r_exp != NULL);

    l_exp = exp->u_op.l_exp;
    l_meta = &l_exp->meta;

    TRY(check_exp(check, l_exp));

    if (!is_bool_meta(l_meta))
        THROW(ERROR_INVALID_OP_TYPE, &l_exp->trc, TYPE_NAME(l_meta->type));

    r_exp = exp->u_op.r_exp;
    r_meta = &r_exp->meta;

    TRY(check_exp(check, r_exp));

    if (!is_bool_meta(r_meta))
        THROW(ERROR_INVALID_OP_TYPE, &r_exp->trc, TYPE_NAME(r_meta->type));

    meta_set_prim(&exp->meta, TYPE_BOOL);

    return NO_ERROR;
}

static int
exp_op_check_bit(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *l_exp, *r_exp;
    meta_t *l_meta, *r_meta;

    ASSERT(exp->u_op.l_exp != NULL);
    ASSERT(exp->u_op.r_exp != NULL);

    l_exp = exp->u_op.l_exp;
    l_meta = &l_exp->meta;

    TRY(check_exp(check, l_exp));

    if (!is_integer_meta(l_meta))
        THROW(ERROR_INVALID_OP_TYPE, &l_exp->trc, TYPE_NAME(l_meta->type));

    r_exp = exp->u_op.r_exp;
    r_meta = &r_exp->meta;

    TRY(check_exp(check, r_exp));

    if (!is_integer_meta(r_meta))
        THROW(ERROR_INVALID_OP_TYPE, &r_exp->trc, TYPE_NAME(r_meta->type));

    exp->meta = l_exp->meta;

    return NO_ERROR;
}

static int
exp_op_check_cmp(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *l_exp, *r_exp;
    meta_t *l_meta, *r_meta;

    ASSERT(exp->u_op.l_exp != NULL);
    ASSERT(exp->u_op.r_exp != NULL);

    l_exp = exp->u_op.l_exp;
    l_meta = &l_exp->meta;

    TRY(check_exp(check, l_exp));

    r_exp = exp->u_op.r_exp;
    r_meta = &r_exp->meta;

    TRY(check_exp(check, r_exp));

    if (is_float_meta(l_meta) && is_integer_meta(r_meta))
        WARN(ERROR_TRUNCATED_TYPE, &l_exp->trc, TYPE_NAME(l_meta->type),
             TYPE_NAME(r_meta->type));
    else if (is_integer_meta(l_meta) && is_float_meta(r_meta))
        WARN(ERROR_TRUNCATED_TYPE, &r_exp->trc, TYPE_NAME(r_meta->type),
             TYPE_NAME(l_meta->type));
    else if (!meta_equals(l_meta, r_meta))
        THROW(ERROR_MISMATCHED_TYPE, &exp->trc, TYPE_NAME(l_meta->type),
              TYPE_NAME(r_meta->type));

    meta_set_prim(&exp->meta, TYPE_BOOL);

    return NO_ERROR;
}

static int
exp_op_check_unary(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *l_exp;
    meta_t *l_meta;

    ASSERT(exp->u_op.l_exp != NULL);
    ASSERT(exp->u_op.r_exp == NULL);

    l_exp = exp->u_op.l_exp;
    l_meta = &l_exp->meta;

    TRY(check_exp(check, l_exp));

    switch (exp->u_op.kind) {
    case OP_INC:
    case OP_DEC:
        if (!is_integer_meta(l_meta))
            THROW(ERROR_INVALID_OP_TYPE, &l_exp->trc, TYPE_NAME(l_meta->type));
        break;

    case OP_NOT:
        if (!is_bool_meta(l_meta))
            THROW(ERROR_INVALID_OP_TYPE, &l_exp->trc, TYPE_NAME(l_meta->type));
        break;

    default:
        ASSERT1(!"invalid operator", exp->u_op.kind);
    }

    exp->meta = *l_meta;

    return NO_ERROR;
}

static int
exp_op_check(check_t *check, ast_exp_t *exp)
{
    ASSERT1(is_op_exp(exp), exp->kind);

    switch (exp->u_op.kind) {
    case OP_ASSIGN:
        return exp_op_check_assign(check, exp);

    case OP_ADD:
    case OP_SUB:
    case OP_MUL:
    case OP_DIV:
    case OP_MOD:
        return exp_op_check_arith(check, exp);

    case OP_AND:
    case OP_OR:
        return exp_op_check_bool_cmp(check, exp);

    case OP_BIT_AND:
    case OP_BIT_OR:
    case OP_BIT_XOR:
    case OP_RSHIFT:
    case OP_LSHIFT:
        return exp_op_check_bit(check, exp);

    case OP_EQ:
    case OP_NE:
    case OP_LT:
    case OP_GT:
    case OP_LE:
    case OP_GE:
        return exp_op_check_cmp(check, exp);

    case OP_INC:
    case OP_DEC:
    case OP_NOT:
        return exp_op_check_unary(check, exp);

    default:
        ASSERT1(!"invalid operator", exp->u_op.kind);
    }

    return NO_ERROR;
}

static int
exp_access_check(check_t *check, ast_exp_t *exp)
{
    ec_t ec;
    ast_exp_t *id_exp, *fld_exp;
    meta_t *id_meta, *fld_meta;
    ast_id_t *id;

    ASSERT1(is_access_exp(exp), exp->kind);

    id_exp = exp->u_acc.id_exp;
    id_meta = &id_exp->meta;

    TRY(check_exp(check, id_exp));

    id = id_exp->id;
    if (id == NULL || is_var_id(id) || is_tuple_meta(id_meta) ||
        (is_func_id(id) && !is_struct_meta(id_meta) && !is_ref_meta(id_meta)))
        THROW(ERROR_NOT_ACCESSIBLE_EXP, &id_exp->trc);

    fld_exp = exp->u_acc.fld_exp;
    fld_meta = &fld_exp->meta;

    check->aq_id = id;
    ec = check_exp(check, fld_exp);
    check->aq_id = NULL;

    if (ec != NO_ERROR)
        return ec;

    exp->id = fld_exp->id;
    exp->meta = *fld_meta;

    return NO_ERROR;
}

static int
exp_call_check(check_t *check, ast_exp_t *exp)
{
    int i;
    ast_exp_t *id_exp;
    meta_t *id_meta;
    ast_id_t *func_id;
    array_t *param_ids;
    array_t *param_exps;

    ASSERT1(is_call_exp(exp), exp->kind);

    id_exp = exp->u_call.id_exp;
    id_meta = &id_exp->meta;

    TRY(check_exp(check, id_exp));

    func_id = id_exp->id;
    if (func_id == NULL || !is_func_id(func_id))
        THROW(ERROR_NOT_CALLABLE_EXP, &id_exp->trc);

    param_ids = func_id->u_func.param_ids;
    param_exps = exp->u_call.param_exps;

    if (array_size(param_ids) != array_size(param_exps))
        THROW(ERROR_MISMATCHED_PARAM, &id_exp->trc, func_id->name);

    for (i = 0; i < array_size(param_exps); i++) {
        ast_id_t *param_id = array_item(param_ids, i, ast_id_t);
        ast_exp_t *param_exp = array_item(param_exps, i, ast_exp_t);

        TRY(check_exp(check, param_exp));

        if (!meta_equals(&param_id->meta, &param_exp->meta))
            THROW(ERROR_MISMATCHED_TYPE, &param_exp->trc,
                  TYPE_NAME(param_id->meta.type),
                  TYPE_NAME(param_exp->meta.type));
    }

    exp->id = func_id;

    return NO_ERROR;
}

static int
exp_sql_check(check_t *check, ast_exp_t *exp)
{
    ASSERT1(is_sql_exp(exp), exp->kind);
    ASSERT(exp->u_sql.sql != NULL);

    // TODO: need column meta

    return NO_ERROR;
}

static int
exp_ternary_check(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *pre_exp, *in_exp, *post_exp;
    meta_t *pre_meta, *in_meta, *post_meta;

    ASSERT1(is_ternary_exp(exp), exp->kind);
    ASSERT(exp->u_tern.pre_exp != NULL);
    ASSERT(exp->u_tern.in_exp != NULL);
    ASSERT(exp->u_tern.post_exp != NULL);

    pre_exp = exp->u_tern.pre_exp;
    pre_meta = &pre_exp->meta;

    TRY(check_exp(check, pre_exp));

    if (!is_bool_meta(pre_meta))
        THROW(ERROR_INVALID_OP_TYPE, &pre_exp->trc,
              TYPE_NAME(pre_meta->type));

    in_exp = exp->u_tern.in_exp;
    in_meta = &in_exp->meta;

    TRY(check_exp(check, in_exp));

    post_exp = exp->u_tern.post_exp;
    post_meta = &post_exp->meta;

    TRY(check_exp(check, post_exp));

    if (!meta_equals(in_meta, post_meta))
        THROW(ERROR_MISMATCHED_TYPE, &post_exp->trc, TYPE_NAME(in_meta->type),
              TYPE_NAME(post_meta->type));

    exp->meta = *in_meta;

    return NO_ERROR;
}

static int
exp_tuple_check(check_t *check, ast_exp_t *exp)
{
    int i;
    array_t *exps = exp->u_tup.exps;

    ASSERT1(is_tuple_exp(exp), exp->kind);
    ASSERT(exps != NULL);

    for (i = 0; i < array_size(exps); i++) {
        ast_exp_t *item = array_item(exps, i, ast_exp_t);

        TRY(check_exp(check, item));
    }

    meta_set_tuple(&exp->meta);

    return NO_ERROR;
}

int
check_exp(check_t *check, ast_exp_t *exp)
{
    switch (exp->kind) {
    case EXP_NULL:
        return NO_ERROR;

    case EXP_ID:
        return exp_id_check(check, exp);

    case EXP_LIT:
        return exp_lit_check(check, exp);

    case EXP_TYPE:
        return exp_type_check(check, exp);

    case EXP_ARRAY:
        return exp_array_check(check, exp);

    case EXP_OP:
        return exp_op_check(check, exp);

    case EXP_ACCESS:
        return exp_access_check(check, exp);

    case EXP_CALL:
        return exp_call_check(check, exp);

    case EXP_SQL:
        return exp_sql_check(check, exp);

    case EXP_TERNARY:
        return exp_ternary_check(check, exp);

    case EXP_TUPLE:
        return exp_tuple_check(check, exp);

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }

    return NO_ERROR;
}

/* end of check_exp.c */
