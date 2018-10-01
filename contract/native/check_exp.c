/**
 * @file    check_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "check_exp.h"

static int exp_type_check(check_t *check, ast_exp_t *exp);

static int
exp_id_ref_check(check_t *check, ast_exp_t *exp)
{
    ast_id_t *id;

    ASSERT1(exp->kind == EXP_ID, exp->kind);

    id = ast_id_search(check->blk, exp->num, exp->u_id.name);
    if (id == NULL) {
        ERROR(ERROR_UNDEFINED_ID, &exp->pos, exp->u_id.name);
        return RC_ERROR;
    }

    ASSERT(id->meta != NULL);

    exp->id = id;
    exp->meta = *id->meta;

    return RC_OK;
}

static int
exp_lit_check(check_t *check, ast_exp_t *exp)
{
    ASSERT1(exp->kind == EXP_LIT, exp->kind);

    switch (exp->u_lit.val.kind) {
    case VAL_NULL:
        ast_meta_set_prim(&exp->meta, TYPE_NONE);
        break;
    case VAL_BOOL:
        ast_meta_set_prim(&exp->meta, TYPE_BOOL);
        break;
    case VAL_INT:
        ast_meta_set_prim(&exp->meta, TYPE_INT64);
        break;
    case VAL_FP:
        ast_meta_set_prim(&exp->meta, TYPE_DOUBLE);
        break;
    case VAL_STR:
        ast_meta_set_prim(&exp->meta, TYPE_STRING);
        break;
    default:
        ASSERT1(!"invalid value", exp->u_lit.val.kind);
    }

    return RC_OK;
}

static int
exp_type_check_struct(check_t *check, ast_exp_t *exp)
{
    ast_id_t *id;

    ASSERT(exp->u_type.name != NULL);
    ASSERT(exp->u_type.k_exp == NULL);
    ASSERT(exp->u_type.v_exp == NULL);

    id = ast_id_search(check->blk, exp->num, exp->u_type.name);
    if (id == NULL) {
        ERROR(ERROR_UNDEFINED_TYPE, &exp->pos, exp->u_type.name);
        return RC_ERROR;
    }

    exp->id = id;
    ast_meta_set_prim(&exp->meta, TYPE_STRUCT);

    return RC_OK;
}

static int
exp_type_check_map(check_t *check, ast_exp_t *exp)
{
    ast_meta_t *k_meta, *v_meta;

    ASSERT1(exp->u_type.name == NULL, exp->u_type.name);
    ASSERT(exp->u_type.k_exp != NULL);
    ASSERT(exp->u_type.v_exp != NULL);

    TRY(exp_type_check(check, exp->u_type.k_exp));

    k_meta = &exp->u_type.k_exp->meta;
    if (k_meta->type > TYPE_STRING) {
        ERROR(ERROR_NOT_ALLOWED_TYPE, &exp->pos, TYPENAME(k_meta->type));
        return RC_ERROR;
    }

    TRY(exp_type_check(check, exp->u_type.v_exp));

    v_meta = &exp->u_type.v_exp->meta;
    ASSERT(!type_is_tuple(v_meta->type));

    meta_set_map(&exp->meta, k_meta->type, v_meta);

    return RC_OK;
}

static int
exp_type_check(check_t *check, ast_exp_t *exp)
{
    ASSERT1(exp->kind == EXP_TYPE, exp->kind);
    ASSERT1(type_is_valid(exp->u_type.type), exp->u_type.type);

    if (type_is_struct(exp->u_type.type)) {
        TRY(exp_type_check_struct(check, exp));
    }
    else if (type_is_map(exp->u_type.type)) {
        TRY(exp_type_check_map(check, exp));
    }
    else {
        ASSERT1(exp->u_type.name == NULL, exp->u_type.name);
        ASSERT(exp->u_type.k_exp == NULL);
        ASSERT(exp->u_type.v_exp == NULL);

        ast_meta_set_prim(&exp->meta, exp->u_type.type);
    }

    return RC_OK;
}

static int
exp_array_check(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *id_exp;
    ast_meta_t *id_meta;

    ASSERT1(exp->kind == EXP_ARRAY, exp->kind);

    id_exp = exp->u_arr.id_exp;
    ASSERT(id_exp != NULL);

    TRY(exp_id_ref_check(check, id_exp));
    id_meta = &id_exp->meta;

    if (exp->u_arr.param_exp != NULL) {
        ast_exp_t *param_exp = exp->u_arr.param_exp;
        ast_meta_t *param_meta;

        if (!type_is_primitive(id_meta->type))
            ERROR(ERROR_NOT_ALLOWED_TYPE, &id_exp->pos,
                  TYPENAME(id_meta->type));

        TRY(check_exp(check, param_exp));
        param_meta = &param_exp->meta;

        // TODO: restriction of array size
        if (!type_is_integer(param_meta->type)) {
            ERROR(ERROR_NOT_ALLOWED_TYPE, &param_exp->pos,
                  TYPENAME(param_meta->type));
            return RC_ERROR;
        }
    }

    exp->meta = *id_meta;

    return RC_OK;
}

static int
exp_op_check_assign(check_t *check, ast_exp_t *exp)
{
    int i;
    ast_exp_t *l_exp, *r_exp;
    ast_meta_t *l_meta, *r_meta;

    ASSERT(exp->u_op.l_exp != NULL);
    ASSERT(exp->u_op.r_exp != NULL);

    l_exp = exp->u_op.l_exp;
    l_meta = &l_exp->meta;

    TRY(check_exp(check, l_exp));

    if (type_is_tuple(l_meta->type)) {
        int i;

        ASSERT1(l_exp->kind == EXP_TUPLE, l_exp->kind);

        for (i = 0; i < array_size(l_exp->u_tup.exps); i++) {
            ast_exp_t *item = array_item(l_exp->u_tup.exps, i, ast_exp_t);

            if (!exp_can_be_lval(item)) {
                ERROR(ERROR_INVALID_LVALUE, &item->pos);
                return RC_ERROR;
            }
        }
    }
    else if (!exp_can_be_lval(l_exp)) {
        ERROR(ERROR_INVALID_LVALUE, &l_exp->pos);
        return RC_ERROR;
    }

    r_exp = exp->u_op.r_exp;
    r_meta = &r_exp->meta;

    TRY(check_exp(check, r_exp));

    if (r_exp->kind == EXP_LIT && type_is_integer(l_meta->type)) {
        ast_val_t *val = &r_exp->u_lit.val;

        if (!type_is_integer(r_meta->type)) {
            ERROR(ERROR_MISMATCHED_TYPE, &exp->pos,
                  TYPENAME(l_meta->type), TYPENAME(r_meta->type));
            return RC_ERROR;
        }

        if (!type_check_range(l_meta->type, val->iv)) {
            ERROR(ERROR_INT_OVERFLOW, &r_exp->pos, TYPENAME(l_meta->type));
            return RC_ERROR;
        }
    }
    else if (!ast_meta_equals(l_meta, r_meta)) {
        ERROR(ERROR_MISMATCHED_TYPE, &exp->pos, TYPENAME(l_meta->type),
              TYPENAME(r_meta->type));
        return RC_ERROR;
    }

    exp->meta = *l_meta;

    return RC_OK;
}

static int
exp_op_check_arith(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *l_exp, *r_exp;
    ast_meta_t *l_meta, *r_meta;

    ASSERT(exp->u_op.l_exp != NULL);
    ASSERT(exp->u_op.r_exp != NULL);

    l_exp = exp->u_op.l_exp;
    l_meta = &l_exp->meta;

    r_exp = exp->u_op.r_exp;
    r_meta = &r_exp->meta;

    TRY(check_exp(check, l_exp));
    TRY(check_exp(check, r_exp));

    if (!ast_meta_equals(l_meta, r_meta)) {
        ERROR(ERROR_MISMATCHED_TYPE, &r_exp->pos, TYPENAME(l_meta->type),
              TYPENAME(r_meta->type));
        return RC_ERROR;
    }

    if (exp->u_op.kind == OP_ADD) {
        if (!type_is_integer(l_meta->type) && !type_is_float(l_meta->type) &&
            !type_is_string(l_meta->type)) {
            ERROR(ERROR_NOT_ALLOWED_TYPE, &exp->pos, TYPENAME(l_meta->type));
            return RC_ERROR;
        }
    }
    else if (!type_is_integer(l_meta->type) && !type_is_float(l_meta->type)) {
        ERROR(ERROR_NOT_ALLOWED_TYPE, &exp->pos, TYPENAME(l_meta->type));
        return RC_ERROR;
    }

    exp->meta = *l_meta;

    return RC_OK;
}

static int
exp_op_check_log_cmp(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *l_exp, *r_exp;
    ast_meta_t *l_meta, *r_meta;

    ASSERT(exp->u_op.l_exp != NULL);
    ASSERT(exp->u_op.r_exp != NULL);

    l_exp = exp->u_op.l_exp;
    l_meta = &l_exp->meta;

    r_exp = exp->u_op.r_exp;
    r_meta = &r_exp->meta;

    TRY(check_exp(check, l_exp));

    if (!type_is_bool(l_meta->type)) {
        ERROR(ERROR_NOT_ALLOWED_TYPE, &l_exp->pos, TYPENAME(l_meta->type));
        return RC_ERROR;
    }

    TRY(check_exp(check, r_exp));

    if (!type_is_bool(r_meta->type)) {
        ERROR(ERROR_NOT_ALLOWED_TYPE, &r_exp->pos, TYPENAME(r_meta->type));
        return RC_ERROR;
    }

    ast_meta_set_prim(&exp->meta, TYPE_BOOL);

    return RC_OK;
}

static int
exp_op_check_bit(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *l_exp, *r_exp;
    ast_meta_t *l_meta, *r_meta;

    ASSERT(exp->u_op.l_exp != NULL);
    ASSERT(exp->u_op.r_exp != NULL);

    l_exp = exp->u_op.l_exp;
    r_exp = exp->u_op.r_exp;

    TRY(check_exp(check, l_exp));

    if (!type_is_integer(l_meta->type)) {
        ERROR(ERROR_NOT_ALLOWED_TYPE, &l_exp->pos, TYPENAME(l_meta->type));
        return RC_ERROR;
    }

    l_meta = &l_exp->meta;
    r_meta = &r_exp->meta;

    TRY(check_exp(check, r_exp));

    if (!type_is_integer(r_meta->type)) {
        ERROR(ERROR_NOT_ALLOWED_TYPE, &r_exp->pos, TYPENAME(r_meta->type));
        return RC_ERROR;
    }

    if (!ast_meta_equals(l_meta, r_meta)) {
        ERROR(ERROR_MISMATCHED_TYPE, &exp->pos, TYPENAME(l_meta->type),
              TYPENAME(r_meta->type));
        return RC_ERROR;
    }

    exp->meta = l_exp->meta;

    return RC_OK;
}

static int
exp_op_check_cmp(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *l_exp, *r_exp;
    ast_meta_t *l_meta, *r_meta;

    ASSERT(exp->u_op.l_exp != NULL);
    ASSERT(exp->u_op.r_exp != NULL);

    l_exp = exp->u_op.l_exp;
    l_meta = &l_exp->meta;

    r_exp = exp->u_op.r_exp;
    r_meta = &r_exp->meta;

    TRY(check_exp(check, l_exp));
    TRY(check_exp(check, r_exp));

    if (type_is_float(l_meta->type) && type_is_integer(r_meta->type)) {
        WARN(ERROR_TRUNCATED_TYPE, &l_exp->pos, TYPENAME(l_meta->type),
             TYPENAME(r_meta->type));

        exp->meta = *r_meta;
    }
    else if (type_is_integer(l_meta->type) && type_is_float(r_meta->type)) {
        WARN(ERROR_TRUNCATED_TYPE, &r_exp->pos, TYPENAME(r_meta->type),
             TYPENAME(l_meta->type));

        exp->meta = *l_meta;
    }
    else if (!ast_meta_equals(l_meta, r_meta)) {
        ERROR(ERROR_MISMATCHED_TYPE, &r_exp->pos, TYPENAME(l_meta->type),
              TYPENAME(r_meta->type));
        return RC_ERROR;
    }
    else {
        exp->meta = *l_meta;
    }

    return RC_OK;
}

static int
exp_op_check_unary(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *l_exp;
    ast_meta_t *l_meta;

    ASSERT(exp->u_op.l_exp != NULL);
    ASSERT(exp->u_op.r_exp == NULL);

    l_exp = exp->u_op.l_exp;
    l_meta = &l_exp->meta;

    TRY(check_exp(check, l_exp));

    if (!type_is_integer(l_meta->type)) {
        ERROR(ERROR_NOT_ALLOWED_TYPE, &l_exp->pos, TYPENAME(l_meta->type));
        return RC_ERROR;
    }

    exp->meta = *l_meta;

    return RC_OK;
}

static int
exp_op_check_not(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *l_exp;
    ast_meta_t *l_meta;

    ASSERT(exp->u_op.l_exp != NULL);
    ASSERT(exp->u_op.r_exp == NULL);

    l_exp = exp->u_op.l_exp;
    l_meta = &l_exp->meta;

    TRY(check_exp(check, l_exp));

    if (!type_is_bool(l_meta->type)) {
        ERROR(ERROR_NOT_ALLOWED_TYPE, &l_exp->pos, TYPENAME(l_meta->type));
        return RC_ERROR;
    }

    ast_meta_set_prim(&exp->meta, TYPE_BOOL);

    return RC_OK;
}

static int
exp_op_check(check_t *check, ast_exp_t *exp)
{
    ASSERT1(exp->kind == EXP_OP, exp->kind);

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
        return exp_op_check_log_cmp(check, exp);

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
        return exp_op_check_unary(check, exp);

    case OP_NOT:
        return exp_op_check_not(check, exp);

    default:
        ASSERT1(!"invalid operator", exp->u_op.kind);
    }

    return RC_OK;
}

static int
exp_access_check(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *id_exp, *fld_exp;
    ast_meta_t *id_meta, *fld_meta;
    ast_id_t *id, *fld_id;
    array_t *mem_ids = NULL;

    ASSERT1(exp->kind == EXP_ACCESS, exp->kind);

    id_exp = exp->u_acc.id_exp;
    id_meta = &id_exp->meta;

    TRY(check_exp(check, id_exp));

    id = id_exp->id;
    if (id == NULL || (id->kind != ID_STRUCT && id->kind != ID_CONTRACT)) {
        ERROR(ERROR_NOT_ACCESSIBLE_EXP, &id_exp->pos);
        return RC_ERROR;
    }

    fld_exp = exp->u_acc.fld_exp;
    fld_meta = &fld_exp->meta;

    TRY(check_exp(check, fld_exp));

    fld_id = fld_exp->id;
    if (fld_id == NULL) {
        ERROR(ERROR_NOT_ACCESSIBLE_EXP, &fld_exp->pos);
        return RC_ERROR;
    }

    ASSERT(fld_id->name != NULL);

    if (id->kind == ID_STRUCT)
        mem_ids = id->u_st.fld_ids;
    else if (id->u_cont.blk != NULL)
        mem_ids = &id->u_cont.blk->ids;

    if (mem_ids != NULL) {
        int i;

        for (i = 0; i < array_size(mem_ids); i++) {
            ast_id_t *mem_id = array_item(mem_ids, i, ast_id_t);

            ASSERT1(mem_id->kind == ID_VAR, mem_id->kind);
            ASSERT(mem_id->name != NULL);

            if (strcmp(mem_id->name, fld_id->name) == 0) {
                exp->id = fld_id;
                break;
            }
        }
    }

    if (exp->id == NULL) {
        ERROR(ERROR_UNDEFINED_ID, &fld_exp->pos, fld_id->name);
        return RC_ERROR;
    }

    exp->meta = *fld_id->meta;

    return RC_OK;
}

static int
exp_call_check(check_t *check, ast_exp_t *exp)
{
    int i;
    ast_exp_t *id_exp;
    ast_meta_t *id_meta;
    ast_id_t *func_id;
    array_t *param_ids;
    array_t *param_exps;

    ASSERT1(exp->kind == EXP_CALL, exp->kind);

    id_exp = exp->u_call.id_exp;
    id_meta = &id_exp->meta;

    TRY(check_exp(check, id_exp));

    func_id = id_exp->id;
    if (func_id == NULL || func_id->kind != ID_FUNC) {
        ERROR(ERROR_NOT_CALLABLE_EXP, &id_exp->pos);
        return RC_ERROR;
    }

    param_ids = func_id->u_func.param_ids;
    param_exps = exp->u_call.param_exps;

    if (array_size(param_ids) != array_size(param_exps)) {
        ERROR(ERROR_MISMATCHED_PARAM, &id_exp->pos, func_id->name);
        return RC_ERROR;
    }

    for (i = 0; i < array_size(param_exps); i++) {
        ast_id_t *param_id = array_item(param_ids, i, ast_id_t);
        ast_exp_t *param_exp = array_item(param_exps, i, ast_exp_t);

        TRY(check_exp(check, param_exp));

        if (!ast_meta_equals(param_id->meta, &param_exp->meta)) {
            ERROR(ERROR_MISMATCHED_TYPE, &param_exp->pos,
                  TYPENAME(param_id->meta->type),
                  TYPENAME(param_exp->meta.type));
            return RC_ERROR;
        }
    }

    return RC_OK;
}

static int
exp_sql_check(check_t *check, ast_exp_t *exp)
{
    ASSERT1(exp->kind == EXP_SQL, exp->kind);

    return RC_OK;
}

static int
exp_cond_check(check_t *check, ast_exp_t *exp)
{
    ASSERT1(exp->kind == EXP_COND, exp->kind);

    return RC_OK;
}

static int
exp_tuple_check(check_t *check, ast_exp_t *exp)
{
    ASSERT1(exp->kind == EXP_TUPLE, exp->kind);

    return RC_OK;
}

int
check_exp(check_t *check, ast_exp_t *exp)
{
    switch (exp->kind) {
    case EXP_ID:
        return exp_id_ref_check(check, exp);

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

    case EXP_COND:
        return exp_cond_check(check, exp);

    case EXP_TUPLE:
        return exp_tuple_check(check, exp);

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }

    return RC_OK;
}

/* end of check_exp.c */
