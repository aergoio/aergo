/**
 * @file    check_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_blk.h"
#include "check_id.h"
#include "syslib.h"

#include "check_exp.h"

static bool
exp_check_lit(check_t *check, ast_exp_t *exp)
{
    value_t *val = &exp->u_lit.val;

    ASSERT1(is_lit_exp(exp), exp->kind);

    switch (val->type) {
    case TYPE_BOOL:
        meta_set_bool(&exp->meta);
        break;

    case TYPE_BYTE:
        meta_set_byte(&exp->meta);
        break;

    case TYPE_INT256:
        meta_set_int256(&exp->meta);
        meta_set_undef(&exp->meta);
        break;

    case TYPE_STRING:
        meta_set_string(&exp->meta);
        break;

    case TYPE_OBJECT:
        ASSERT1(is_null_val(val), val_size(val));
        meta_set_object(&exp->meta, NULL);
        meta_set_undef(&exp->meta);
        break;

    default:
        ASSERT1(!"invalid value", val->type);
    }

    exp->usable_lval = false;

    return true;
}

static bool
exp_check_id(check_t *check, ast_exp_t *exp)
{
    ast_id_t *id = NULL;
    char *name = exp->u_id.name;

    ASSERT1(is_id_exp(exp), exp->kind);
    ASSERT(name != NULL);

    if (check->qual_id != NULL) {
        id = id_search_fld(check->qual_id, name, check->cont_id == check->qual_id);
    }
    else {
        if (check->fn_id != NULL)
            id = id_search_param(check->fn_id, name);

        if (id == NULL) {
            id = blk_search_id(check->blk, name);

            if (id == NULL && strcmp(name, "this") == 0)
                id = check->cont_id;
        }
    }

    if (id == NULL)
        RETURN(ERROR_UNDEFINED_ID, &exp->pos, name);

    id_trycheck(check, id);

    id->is_used = true;

    if (is_const_id(id) && id->val != NULL)
        exp_set_lit(exp, id->val);
    else
        exp->id = id;

    if (!is_var_id(id) || is_const_id(id))
        exp->usable_lval = false;

    meta_copy(&exp->meta, &id->meta);

    return true;
}

static bool
exp_check_type(check_t *check, ast_exp_t *exp)
{
    ASSERT1(is_type_exp(exp), exp->kind);

    if (exp->u_type.type == TYPE_NONE) {
        ast_id_t *id;
        char *name = exp->u_type.name;

        ASSERT(name != NULL);
        ASSERT(exp->u_type.k_exp == NULL);
        ASSERT(exp->u_type.v_exp == NULL);

        id = blk_search_type(check->blk, name);
        if (id == NULL)
            RETURN(ERROR_UNDEFINED_TYPE, &exp->pos, name);

        ASSERT1(is_type_id(id), id->kind);

        id_trycheck(check, id);

        id->is_used = true;

		exp->id = id;
        meta_copy(&exp->meta, &id->meta);
    }
    else if (exp->u_type.type == TYPE_MAP) {
        ast_exp_t *k_exp = exp->u_type.k_exp;
        ast_exp_t *v_exp = exp->u_type.v_exp;
        meta_t *k_meta, *v_meta;

        ASSERT(exp->u_type.name == NULL);
        ASSERT(exp->u_type.k_exp != NULL);
        ASSERT(exp->u_type.v_exp != NULL);

        k_meta = &k_exp->meta;
        v_meta = &v_exp->meta;

        CHECK(exp_check_type(check, k_exp));

        if (!is_comparable_meta(k_meta))
            RETURN(ERROR_NOT_COMPARABLE_TYPE, k_meta->pos, meta_to_str(k_meta));

        CHECK(exp_check_type(check, v_exp));

        ASSERT(!is_tuple_meta(v_meta));
        meta_set_map(&exp->meta, k_meta, v_meta);
    }
    else {
        ASSERT(exp->u_type.name == NULL);
        ASSERT(exp->u_type.k_exp == NULL);
        ASSERT(exp->u_type.v_exp == NULL);

        meta_set(&exp->meta, exp->u_type.type);
    }

    exp->usable_lval = false;

    return true;
}

static bool
exp_check_array(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *id_exp;
    meta_t *id_meta;
    ast_exp_t *idx_exp;
    meta_t *idx_meta;

    ASSERT1(is_array_exp(exp), exp->kind);
    ASSERT(exp->u_arr.id_exp != NULL);

    id_exp = exp->u_arr.id_exp;
    id_meta = &id_exp->meta;

    CHECK(exp_check(check, id_exp));

    exp->id = id_exp->id;
    ASSERT(exp->id != NULL);

    idx_exp = exp->u_arr.idx_exp;
    idx_meta = &idx_exp->meta;

    CHECK(exp_check(check, idx_exp));

    if (is_array_meta(id_meta)) {
        ASSERT1(id_meta->elem_cnt == 1, id_meta->elem_cnt);
        ASSERT1(id_meta->arr_dim > 0, id_meta->arr_dim);
        ASSERT1(id_meta->dim_sizes != NULL, id_meta->arr_dim);

        if (!is_integer_meta(idx_meta))
            RETURN(ERROR_INVALID_SIZE_VAL, &idx_exp->pos, meta_to_str(idx_meta));

        meta_copy(&exp->meta, id_meta);

        if (is_lit_exp(idx_exp)) {
            value_t *idx_val = &idx_exp->u_lit.val;

            /* The "dim_sizes[0]" can be negative if array is used as a parameter */
            if (is_neg_val(idx_val) || val_i64(idx_val) > INT32_MAX ||
                (id_meta->dim_sizes[0] > 0 && val_i64(idx_val) >= (uint)id_meta->dim_sizes[0]))
                RETURN(ERROR_INVALID_ARR_IDX, &idx_exp->pos);

            meta_set_int32(&idx_exp->meta);
        }

        /* Whenever an array element is accessed, strip it by one dimension */
        meta_strip_arr_dim(&exp->meta);

        if (is_array_meta(&exp->meta))
            exp->usable_lval = false;
    }
    else if (is_map_meta(id_meta)) {
        CHECK(meta_cmp(id_meta->elems[0], idx_meta));

        meta_eval(id_meta->elems[0], idx_meta);
        meta_copy(&exp->meta, id_meta->elems[1]);
    }
    else if (is_string_meta(id_meta)) {
        if (!is_integer_meta(idx_meta))
            RETURN(ERROR_INVALID_SIZE_VAL, &idx_exp->pos, meta_to_str(idx_meta));

        if (is_lit_exp(idx_exp)) {
            value_t *idx_val = &idx_exp->u_lit.val;

            if (is_neg_val(idx_val) || val_i64(idx_val) > INT32_MAX)
                RETURN(ERROR_INVALID_ARR_IDX, &idx_exp->pos);

            meta_set_int32(&idx_exp->meta);
        }

        meta_set_byte(&exp->meta);
    }
    else {
        RETURN(ERROR_INVALID_SUBSCRIPT, &id_exp->pos);
    }

    return true;
}

static bool
exp_check_cast(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *val_exp;
    meta_t *val_meta;

    ASSERT1(is_cast_exp(exp), exp->kind);
    ASSERT1(is_valid_type(exp->u_cast.to_meta.type), exp->u_cast.to_meta.type);
    ASSERT(exp->u_cast.val_exp != NULL);

    val_exp = exp->u_cast.val_exp;
    val_meta = &val_exp->meta;

    CHECK(exp_check(check, val_exp));

    meta_copy(&exp->meta, &exp->u_cast.to_meta);

    if (is_array_meta(val_meta) || !is_compatible_meta(val_meta, &exp->meta))
        RETURN(ERROR_INCOMPATIBLE_TYPE, &val_exp->pos, meta_to_str(val_meta),
               meta_to_str(&exp->meta));

    if (exp->meta.type == val_meta->type) {
        *exp = *val_exp;
    }
    else if (is_lit_exp(val_exp)) {
        exp_set_lit(exp, &val_exp->u_lit.val);
        meta_set_undef(&exp->meta);

        value_cast(&exp->u_lit.val, exp->meta.type);
    }

    exp->usable_lval = false;

    return true;
}

static bool
exp_check_unary(check_t *check, ast_exp_t *exp)
{
    op_kind_t op = exp->u_un.kind;
    ast_exp_t *val_exp;
    meta_t *val_meta;

    ASSERT1(is_unary_exp(exp), exp->kind);
    ASSERT(exp->u_un.val_exp != NULL);

    val_exp = exp->u_un.val_exp;
    val_meta = &val_exp->meta;

    CHECK(exp_check(check, val_exp));

    switch (op) {
    case OP_INC:
    case OP_DEC:
        if (!is_usable_lval(val_exp))
            RETURN(ERROR_INVALID_LVALUE, &val_exp->pos);

        if (!is_integer_meta(val_meta))
            RETURN(ERROR_INVALID_OP_TYPE, &val_exp->pos, meta_to_str(val_meta));

        meta_copy(&exp->meta, val_meta);
        break;

    case OP_NEG:
        if (is_byte_meta(val_meta) || !is_numeric_meta(val_meta))
            RETURN(ERROR_INVALID_OP_TYPE, &val_exp->pos, meta_to_str(val_meta));

        meta_copy(&exp->meta, val_meta);
        break;

    case OP_NOT:
        if (!is_bool_meta(val_meta))
            RETURN(ERROR_INVALID_OP_TYPE, &val_exp->pos, meta_to_str(val_meta));

        meta_copy(&exp->meta, val_meta);
        break;

    case OP_BIT_NOT:
        if (!is_integer_meta(val_meta))
            RETURN(ERROR_INVALID_OP_TYPE, &val_exp->pos, meta_to_str(val_meta));

        meta_copy(&exp->meta, val_meta);
        break;

    default:
        ASSERT1(!"invalid operator", exp->u_un.kind);
    }

    if (is_lit_exp(val_exp)) {
        exp_set_lit(exp, NULL);
        meta_set_undef(&exp->meta);

        value_eval(&val_exp->u_lit.val, op, NULL, &exp->u_lit.val);
    }

    exp->usable_lval = false;

    return true;
}

static bool
exp_check_op_arith(check_t *check, ast_exp_t *exp)
{
    op_kind_t op = exp->u_bin.kind;
    ast_exp_t *l_exp, *r_exp;
    meta_t *l_meta, *r_meta;

    ASSERT(exp->u_bin.l_exp != NULL);
    ASSERT(exp->u_bin.r_exp != NULL);

    l_exp = exp->u_bin.l_exp;
    l_meta = &l_exp->meta;

    CHECK(exp_check(check, l_exp));

    if (op == OP_ADD) {
        if (!is_numeric_meta(l_meta) && !is_string_meta(l_meta))
            RETURN(ERROR_INVALID_OP_TYPE, &l_exp->pos, meta_to_str(l_meta));
    }
    else if (op == OP_MOD) {
        if (!is_integer_meta(l_meta))
            RETURN(ERROR_INVALID_OP_TYPE, &l_exp->pos, meta_to_str(l_meta));
    }
    else if (!is_numeric_meta(l_meta)) {
        RETURN(ERROR_INVALID_OP_TYPE, &l_exp->pos, meta_to_str(l_meta));
    }

    r_exp = exp->u_bin.r_exp;
    r_meta = &r_exp->meta;

    CHECK(exp_check(check, r_exp));
    CHECK(meta_cmp(l_meta, r_meta));

    meta_eval(l_meta, r_meta);
    meta_copy(&exp->meta, l_meta);

    return true;
}

static bool
exp_check_op_bit(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *l_exp, *r_exp;
    meta_t *l_meta, *r_meta;

    ASSERT(exp->u_bin.l_exp != NULL);
    ASSERT(exp->u_bin.r_exp != NULL);

    l_exp = exp->u_bin.l_exp;
    l_meta = &l_exp->meta;

    CHECK(exp_check(check, l_exp));

    if (!is_integer_meta(l_meta))
        RETURN(ERROR_INVALID_OP_TYPE, &l_exp->pos, meta_to_str(l_meta));

    r_exp = exp->u_bin.r_exp;
    r_meta = &r_exp->meta;

    CHECK(exp_check(check, r_exp));

    switch (exp->u_bin.kind) {
    case OP_BIT_AND:
    case OP_BIT_OR:
    case OP_BIT_XOR:
        if (!is_integer_meta(r_meta))
            RETURN(ERROR_INVALID_OP_TYPE, &r_exp->pos, meta_to_str(r_meta));
        break;

    case OP_BIT_SHR:
    case OP_BIT_SHL:
        if (is_lit_exp(r_exp) && is_neg_val(&r_exp->u_lit.val))
            RETURN(ERROR_INVALID_OP_TYPE, &r_exp->pos, meta_to_str(r_meta));
        break;

    default:
        ASSERT1(!"invalid operator", exp->u_bin.kind);
    }

    meta_eval(l_meta, r_meta);
    meta_copy(&exp->meta, l_meta);

    return true;
}

static bool
exp_check_op_cmp(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *l_exp, *r_exp;
    meta_t *l_meta, *r_meta;

    ASSERT(exp->u_bin.l_exp != NULL);
    ASSERT(exp->u_bin.r_exp != NULL);

    l_exp = exp->u_bin.l_exp;
    l_meta = &l_exp->meta;

    CHECK(exp_check(check, l_exp));

    r_exp = exp->u_bin.r_exp;
    r_meta = &r_exp->meta;

    CHECK(exp_check(check, r_exp));

    /* Comparisons are possible between variables of type object, struct or array */
    if (is_tuple_meta(l_meta))
        RETURN(ERROR_INVALID_OP_TYPE, &l_exp->pos, meta_to_str(l_meta));
    else if (is_tuple_meta(r_meta))
        RETURN(ERROR_INVALID_OP_TYPE, &r_exp->pos, meta_to_str(r_meta));

    CHECK(meta_cmp(l_meta, r_meta));

    meta_eval(l_meta, r_meta);
    meta_set_bool(&exp->meta);

    return true;
}

static bool
exp_check_op_andor(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *l_exp, *r_exp;
    meta_t *l_meta, *r_meta;

    ASSERT(exp->u_bin.l_exp != NULL);
    ASSERT(exp->u_bin.r_exp != NULL);

    l_exp = exp->u_bin.l_exp;
    l_meta = &l_exp->meta;

    CHECK(exp_check(check, l_exp));

    if (!is_bool_meta(l_meta))
        RETURN(ERROR_INVALID_COND_TYPE, &l_exp->pos, meta_to_str(l_meta));

    r_exp = exp->u_bin.r_exp;
    r_meta = &r_exp->meta;

    CHECK(exp_check(check, r_exp));

    if (!is_bool_meta(r_meta))
        RETURN(ERROR_INVALID_COND_TYPE, &r_exp->pos, meta_to_str(r_meta));

    CHECK(meta_cmp(l_meta, r_meta));

    meta_eval(l_meta, r_meta);
    meta_set_bool(&exp->meta);

    return true;
}

static bool
exp_check_binary(check_t *check, ast_exp_t *exp)
{
    op_kind_t op = exp->u_bin.kind;

    ASSERT1(is_binary_exp(exp), exp->kind);

    switch (op) {
    case OP_ADD:
    case OP_SUB:
    case OP_MUL:
    case OP_DIV:
    case OP_MOD:
        CHECK(exp_check_op_arith(check, exp));
        break;

    case OP_BIT_AND:
    case OP_BIT_OR:
    case OP_BIT_XOR:
    case OP_BIT_SHR:
    case OP_BIT_SHL:
        CHECK(exp_check_op_bit(check, exp));
        break;

    case OP_EQ:
    case OP_NE:
    case OP_LT:
    case OP_GT:
    case OP_LE:
    case OP_GE:
        CHECK(exp_check_op_cmp(check, exp));
        break;

    case OP_AND:
    case OP_OR:
        CHECK(exp_check_op_andor(check, exp));
        break;

    default:
        ASSERT1(!"invalid operator", exp->u_bin.kind);
    }

    if (is_lit_exp(exp->u_bin.l_exp)) {
        ast_exp_t *l_exp = exp->u_bin.l_exp;
        ast_exp_t *r_exp = exp->u_bin.r_exp;

        if (!is_lit_exp(r_exp))
            return true;

        if ((op == OP_DIV || op == OP_MOD) && is_zero_val(&r_exp->u_lit.val))
            RETURN(ERROR_DIVIDE_BY_ZERO, &r_exp->pos);

        exp_set_lit(exp, NULL);
        meta_set_undef(&exp->meta);

        value_eval(&l_exp->u_lit.val, op, &r_exp->u_lit.val, &exp->u_lit.val);
    }

    exp->usable_lval = false;

    return true;
}

static bool
exp_check_ternary(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *pre_exp, *in_exp, *post_exp;
    meta_t *pre_meta, *in_meta, *post_meta;

    ASSERT1(is_ternary_exp(exp), exp->kind);
    ASSERT(exp->u_tern.pre_exp != NULL);
    ASSERT(exp->u_tern.in_exp != NULL);
    ASSERT(exp->u_tern.post_exp != NULL);

    pre_exp = exp->u_tern.pre_exp;
    pre_meta = &pre_exp->meta;

    CHECK(exp_check(check, pre_exp));

    if (!is_bool_meta(pre_meta))
        RETURN(ERROR_INVALID_COND_TYPE, &pre_exp->pos, meta_to_str(pre_meta));

    in_exp = exp->u_tern.in_exp;
    in_meta = &in_exp->meta;

    CHECK(exp_check(check, in_exp));

    post_exp = exp->u_tern.post_exp;
    post_meta = &post_exp->meta;

    CHECK(exp_check(check, post_exp));
    CHECK(meta_cmp(in_meta, post_meta));

    meta_eval(in_meta, post_meta);
    meta_copy(&exp->meta, in_meta);

    exp->usable_lval = false;

    return true;
}

static bool
exp_check_access(check_t *check, ast_exp_t *exp)
{
    ast_exp_t *qual_exp, *fld_exp;
    meta_t *qual_meta;
    ast_id_t *qual_id;

    ASSERT1(is_access_exp(exp), exp->kind);

    qual_exp = exp->u_acc.qual_exp;
    qual_meta = &qual_exp->meta;

    CHECK(exp_check(check, qual_exp));

    qual_id = qual_exp->id;
    if (qual_id == NULL)
        RETURN(ERROR_INACCESSIBLE_TYPE, &qual_exp->pos, meta_to_str(qual_meta));

    fld_exp = exp->u_acc.fld_exp;

    if (is_array_meta(qual_meta) || is_map_meta(qual_meta)) {
        if (!is_id_exp(fld_exp) || strcmp(fld_exp->u_id.name, "size") != 0)
            RETURN(ERROR_INACCESSIBLE_TYPE, &qual_exp->pos, meta_to_str(qual_meta));

        meta_set_int32(&exp->meta);
        exp->id = qual_exp->id;
        exp->usable_lval = false;

        return true;
    }

    if (!is_enum_id(qual_id) && !is_struct_meta(qual_meta) && !is_object_meta(qual_meta))
        RETURN(ERROR_INACCESSIBLE_TYPE, &qual_exp->pos, meta_to_str(qual_meta));

    /* Get the actual struct, contract or interface identifier */
    if (is_struct_meta(qual_meta) || is_object_meta(qual_meta)) {
        /* if "this.x" is used, "qual_id" is the contract identifier */
        ASSERT1(is_var_id(qual_id) || is_fn_id(qual_id) || is_cont_id(qual_id), qual_id->kind);
        ASSERT(qual_meta->type_id != NULL);

        if (is_cont_id(qual_id))
            qual_exp->usable_lval = true;

        qual_id = qual_meta->type_id;
    }

    ASSERT(qual_id != NULL);
    ASSERT1(is_type_id(qual_id) || is_enum_id(qual_id), qual_id->kind);

    check->qual_id = qual_id;

    if (!exp_check(check, fld_exp)) {
        check->qual_id = NULL;
        return false;
    }

    check->qual_id = NULL;

    if (is_lit_exp(fld_exp))
        /* enum or contract constant */
        exp_set_lit(exp, &fld_exp->u_lit.val);
    else
        exp->id = fld_exp->id;

    meta_copy(&exp->meta, &fld_exp->meta);

    exp->usable_lval = qual_exp->usable_lval && fld_exp->usable_lval;

    return true;
}

static bool
exp_check_syslib(check_t *check, ast_exp_t *exp)
{
    int i;
    sys_fn_t *sys_fn;
    vector_t *arg_exps = exp->u_call.arg_exps;

    ASSERT1(is_call_exp(exp), exp->kind);
    ASSERT1(exp->u_call.kind < FN_MAX, exp->u_call.kind);
    ASSERT(exp->u_call.id_exp == NULL);
    ASSERT(arg_exps != NULL);

    sys_fn = SYS_FN(exp->u_call.kind);

    ASSERT2(sys_fn->param_cnt == vector_size(arg_exps), sys_fn->param_cnt, vector_size(arg_exps));

    vector_foreach(arg_exps, i) {
        ast_exp_t *arg_exp = vector_get_exp(arg_exps, i);

        CHECK(exp_check(check, arg_exp));

        ASSERT2(sys_fn->params[i] == arg_exp->meta.type, sys_fn->params[i], arg_exp->meta.type);
    }

    meta_set(&exp->meta, sys_fn->result);

    exp->usable_lval = false;
    exp->u_call.qname = sys_fn->qname;

    return true;
}

static bool
exp_check_call(check_t *check, ast_exp_t *exp)
{
    int i;
    ast_exp_t *id_exp;
    vector_t *arg_exps;
    ast_id_t *id;
    vector_t *param_ids;

    ASSERT1(is_call_exp(exp), exp->kind);

    if (exp->u_call.kind != FN_UDF && exp->u_call.kind != FN_NEW)
        return exp_check_syslib(check, exp);

    id_exp = exp->u_call.id_exp;
    arg_exps = exp->u_call.arg_exps;

    CHECK(exp_check(check, id_exp));

    id = id_exp->id;
    if (id == NULL)
        RETURN(ERROR_NOT_CALLABLE_EXP, &id_exp->pos);

    if (exp->u_call.kind == FN_NEW) {
        if (is_cont_id(id)) {
            /* In case of the contract identifier, the constructor is searched again */
            id = blk_search_id(id->u_cont.blk, id->name);
            ASSERT(id != NULL);

            id_trycheck(check, id);
            ASSERT(id->up != NULL);
        }

        if (!is_ctor_id(id))
            RETURN(ERROR_UNDEFINED_ID, &id_exp->pos, id->name);
    }

    if (!is_fn_id(id))
        RETURN(ERROR_NOT_CALLABLE_EXP, &id_exp->pos);

    if (id == check->fn_id)
        WARN(ERROR_RECURSIVE_CALL, &id_exp->pos);

    ASSERT1(is_id_exp(id_exp) || is_access_exp(id_exp), id_exp->kind);

    param_ids = id->u_fn.param_ids;

    if (vector_size(param_ids) != vector_size(arg_exps))
        RETURN(ERROR_MISMATCHED_COUNT, &id_exp->pos, "parameter", vector_size(param_ids),
               vector_size(arg_exps));

    vector_foreach(arg_exps, i) {
        ast_id_t *param_id = vector_get_id(param_ids, i);
        ast_exp_t *arg_exp = vector_get_exp(arg_exps, i);

        CHECK(exp_check(check, arg_exp));

        if (!is_call_exp(arg_exp) && arg_exp->id != NULL && !is_var_id(arg_exp->id) &&
            (!is_cont_id(arg_exp->id) || !is_id_exp(arg_exp) ||
             strcmp(arg_exp->u_id.name, "this") != 0))
            ERROR(ERROR_NOT_ALLOWED_PARAM, &arg_exp->pos);

        meta_eval(&param_id->meta, &arg_exp->meta);

        exp_check_overflow(arg_exp, &param_id->meta);
    }

    exp->id = id;
    meta_copy(&exp->meta, &id->meta);

    exp->usable_lval = false;
    exp->u_call.qname = id->u_fn.qname;

    return true;
}

static bool
exp_check_sql(check_t *check, ast_exp_t *exp)
{
    ASSERT1(is_sql_exp(exp), exp->kind);
    ASSERT(exp->u_sql.sql != NULL);

    switch (exp->u_sql.kind) {
    case SQL_QUERY:
        // TODO: need column meta
        break;

    case SQL_INSERT:
    case SQL_UPDATE:
    case SQL_DELETE:
    case SQL_REPLACE:
        meta_set_int32(&exp->meta);
        break;

    default:
        ASSERT1(!"invalid sql", exp->u_sql.kind);
    }

    exp->usable_lval = false;

    return true;
}

static bool
exp_check_tuple(check_t *check, ast_exp_t *exp)
{
    int i;
    vector_t *elem_exps = exp->u_tup.elem_exps;

    ASSERT1(is_tuple_exp(exp), exp->kind);
    ASSERT(elem_exps != NULL);

    vector_foreach(elem_exps, i) {
        ast_exp_t *elem_exp = vector_get_exp(elem_exps, i);

        CHECK(exp_check(check, elem_exp));

        exp->usable_lval &= elem_exp->usable_lval;
    }

    meta_set_tuple(&exp->meta, elem_exps);

    return true;
}

static bool
exp_check_init(check_t *check, ast_exp_t *exp)
{
    int i;
    vector_t *elem_exps = exp->u_init.elem_exps;

    ASSERT1(is_init_exp(exp), exp->kind);
    ASSERT(elem_exps != NULL);

    exp->u_init.is_static = true;

    vector_foreach(elem_exps, i) {
        ast_exp_t *elem_exp = vector_get_exp(elem_exps, i);

        ASSERT1(!is_tuple_exp(elem_exp), elem_exp->kind);

        CHECK(exp_check(check, elem_exp));

        if ((is_lit_exp(elem_exp) && is_int_val(&elem_exp->u_lit.val) &&
             !mpz_fits_slong_p(val_mpz(&elem_exp->u_lit.val)) &&
             !mpz_fits_ulong_p(val_mpz(&elem_exp->u_lit.val))) ||
            (!is_lit_exp(elem_exp) && (!is_init_exp(elem_exp) || !elem_exp->u_init.is_static)))
            exp->u_init.is_static = false;
    }

    meta_set_tuple(&exp->meta, elem_exps);

    exp->usable_lval = false;

    return true;
}

static bool
exp_check_alloc(check_t *check, ast_exp_t *exp)
{
    meta_t *meta = &exp->meta;
    ast_exp_t *type_exp = exp->u_alloc.type_exp;
    vector_t *size_exps = exp->u_alloc.size_exps;

    ASSERT1(is_alloc_exp(exp), exp->kind);
    ASSERT(type_exp != NULL);

    CHECK(exp_check(check, type_exp));

    if (size_exps == NULL) {
        meta_copy(meta, &type_exp->meta);

        if (is_primitive_meta(meta) || is_object_meta(meta))
            RETURN(ERROR_NOT_ALLOWED_ALLOC, &exp->pos);
    }
    else {
        int i, dim_sz;

        meta_set_array(meta, &type_exp->meta, vector_size(size_exps));

        vector_foreach(size_exps, i) {
            value_t *size_val = NULL;
            ast_exp_t *size_exp = vector_get_exp(size_exps, i);

            CHECK(exp_check(check, size_exp));

            if (size_exp->id != NULL && is_const_id(size_exp->id))
                /* constant variable */
                size_val = size_exp->id->val;
            else if (is_lit_exp(size_exp) && is_integer_meta(&size_exp->meta))
                /* integer literal */
                size_val = &size_exp->u_lit.val;
            else
                RETURN(ERROR_INVALID_SIZE_VAL, &size_exp->pos);

            ASSERT(size_val != NULL);
            ASSERT1(is_int_val(size_val), size_val->type);

            dim_sz = val_i64(size_val);
            if (dim_sz <= 0)
                RETURN(ERROR_INVALID_SIZE_VAL, &size_exp->pos);

            meta_set_arr_dim(meta, i, dim_sz);
        }
    }

    exp->usable_lval = false;

    return true;
}

bool
exp_check(check_t *check, ast_exp_t *exp)
{
    ASSERT(exp != NULL);

    switch (exp->kind) {
    case EXP_NULL:
        return true;

    case EXP_LIT:
        return exp_check_lit(check, exp);

    case EXP_ID:
        return exp_check_id(check, exp);

    case EXP_TYPE:
        return exp_check_type(check, exp);

    case EXP_ARRAY:
        return exp_check_array(check, exp);

    case EXP_CAST:
        return exp_check_cast(check, exp);

    case EXP_UNARY:
        return exp_check_unary(check, exp);

    case EXP_BINARY:
        return exp_check_binary(check, exp);

    case EXP_TERNARY:
        return exp_check_ternary(check, exp);

    case EXP_ACCESS:
        return exp_check_access(check, exp);

    case EXP_CALL:
        return exp_check_call(check, exp);

    case EXP_SQL:
        return exp_check_sql(check, exp);

    case EXP_TUPLE:
        return exp_check_tuple(check, exp);

    case EXP_INIT:
        return exp_check_init(check, exp);

    case EXP_ALLOC:
        return exp_check_alloc(check, exp);

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }

    return true;
}

/* end of check_exp.c */
