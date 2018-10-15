/**
 * @file    check_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "array.h"
#include "ast_exp.h"

#include "check_exp.h"
#include "check_blk.h"

#include "check_id.h"

static int
id_var_check_array(check_t *check, ast_id_t *id, bool is_param)
{
    int i;
    array_t *size_exps = id->u_var.size_exps;

    meta_set_array(&id->meta, array_size(size_exps));

    for (i = 0; i < array_size(size_exps); i++) {
        ast_exp_t *size_exp = array_item(size_exps, i, ast_exp_t);

        CHECK(exp_check(check, size_exp));

        if (is_null_exp(size_exp)) {
            if (!is_param && id->u_var.init_exp == NULL)
                RETURN(ERROR_MISSING_ARR_SIZE, &size_exp->pos);

            id->meta.arr_size[i] = -1;
        }
        else {
            value_t *size_val = &size_exp->u_val.val;
            meta_t *size_meta = &size_exp->meta;

            if (!is_integer_meta(size_meta))
                RETURN(ERROR_INVALID_SIZE_TYPE, &size_exp->pos,
                       META_NAME(size_meta));

            if (!is_untyped_meta(size_meta))
                RETURN(ERROR_INVALID_SIZE_VAL, &size_exp->pos);

            ASSERT1(is_val_exp(size_exp), size_exp->kind);
            ASSERT1(is_int_val(size_val), size_val->kind);

            id->meta.arr_size[i] = size_val->iv;
        }
    }

    return NO_ERROR;
}

static int
id_var_check(check_t *check, ast_id_t *id)
{
    ast_exp_t *type_exp;
    meta_t *type_meta;

    ASSERT1(is_var_id(id), id->kind);
    ASSERT(id->u_var.type_exp != NULL);

    type_exp = id->u_var.type_exp;
    type_meta = &type_exp->meta;

    ASSERT1(is_type_exp(type_exp), type_exp->kind);

    CHECK(exp_check(check, type_exp));

    if (type_exp->u_type.is_public)
        flag_set(id->mod, MOD_PUBLIC);

    id->meta = *type_meta;

    if (id->u_var.size_exps != NULL)
        CHECK(id_var_check_array(check, id, false));

    if (id->u_var.init_exp != NULL) {
        /* TODO: named initializer */
        ast_exp_t *init_exp = id->u_var.init_exp;
        meta_t *init_meta = &init_exp->meta;

        CHECK(exp_check(check, init_exp));

        if (id->u_var.size_exps == NULL) {
            if (is_tuple_meta(init_meta) &&
                !is_map_meta(type_meta) && !is_struct_meta(type_meta))
                RETURN(ERROR_NOT_ALLOWED_INIT, &init_exp->pos);

            /* TODO: need to check value overflow */
            //return meta_check(&id->meta, init_meta);
        }
        else if (is_tuple_meta(init_meta)) {
            /* in case of contract type */
            if (type_exp->id != NULL &&
                !is_struct_meta(type_meta) && !is_map_meta(type_meta))
                RETURN(ERROR_NOT_ALLOWED_INIT, &init_exp->pos);

            //return meta_check_array(&id->meta, 0, init_meta);
        }
        else {
            RETURN(ERROR_NOT_SUPPORTED, &init_exp->pos);
        }

        return meta_check(&id->meta, init_meta);
    }

    return NO_ERROR;
}

static int
id_struct_check(check_t *check, ast_id_t *id)
{
    int i;
    array_t *fld_ids;

    ASSERT1(is_struct_id(id), id->kind);

    fld_ids = id->u_st.fld_ids;
    ASSERT(fld_ids != NULL);

    for (i = 0; i < array_size(fld_ids); i++) {
        ast_id_t *fld_id = array_item(fld_ids, i, ast_id_t);

        CHECK(id_var_check(check, fld_id));
    }

    meta_set_struct(&id->meta, fld_ids);

    return NO_ERROR;
}

static int
id_param_check(check_t *check, ast_id_t *id)
{
    ast_exp_t *type_exp;

    ASSERT1(is_var_id(id), id->kind);
    ASSERT(id->u_var.type_exp != NULL);
    ASSERT(id->u_var.init_exp == NULL);

    type_exp = id->u_var.type_exp;
    ASSERT1(is_type_exp(type_exp), type_exp->kind);

    CHECK(exp_check(check, type_exp));

    id->meta = type_exp->meta;

    if (id->u_var.size_exps != NULL)
        CHECK(id_var_check_array(check, id, true));

    return NO_ERROR;
}

static int
id_func_check(check_t *check, ast_id_t *id)
{
    int i;
    array_t *param_ids;
    array_t *ret_exps;

    ASSERT1(is_func_id(id), id->kind);

    param_ids = id->u_func.param_ids;

    for (i = 0; i < array_size(param_ids); i++) {
        ast_id_t *param_id = array_item(param_ids, i, ast_id_t);

        id_param_check(check, param_id);
    }

    ret_exps = id->u_func.ret_exps;

    if (ret_exps != NULL) {
        for (i = 0; i < array_size(ret_exps); i++) {
            ast_exp_t *type_exp = array_item(ret_exps, i, ast_exp_t);

            ASSERT1(is_type_exp(type_exp), type_exp->kind);

            exp_check(check, type_exp);
        }

        meta_set_tuple(&id->meta, ret_exps);
    }
    else if (id->mod == MOD_CTOR) {
        meta_set_ref(&id->meta);
    }
    else {
        meta_set_void(&id->meta);
    }

    check->fn_id = id;

    if (id->u_func.blk != NULL)
        blk_check(check, id->u_func.blk);

    check->fn_id = NULL;

    return NO_ERROR;
}

static int
id_contract_check(check_t *check, ast_id_t *id)
{
    ASSERT1(is_contract_id(id), id->kind);

    if (id->u_cont.blk != NULL)
        blk_check(check, id->u_cont.blk);

    meta_set_ref(&id->meta);

    return NO_ERROR;
}

void
id_check(check_t *check, ast_id_t *id)
{
    ASSERT(id->name != NULL);

    switch (id->kind) {
    case ID_VAR:
        id_var_check(check, id);
        break;

    case ID_STRUCT:
        id_struct_check(check, id);
        break;

    case ID_FUNC:
        id_func_check(check, id);
        break;

    case ID_CONTRACT:
        id_contract_check(check, id);
        break;

    default:
        ASSERT1(!"invalid identifier", id->kind);
    }
}

/* end of check_id.c */
