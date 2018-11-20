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
id_check_var_array(check_t *check, ast_id_t *id, bool is_param)
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
            ast_id_t *size_id = size_exp->id;
            meta_t *size_meta = &size_exp->meta;
            value_t *size_val;

            if (size_id != NULL && size_id->val != NULL) {
                /* constant variable */
                size_val = size_id->val;
            }
            else if (is_dec_family(size_meta) && is_const_type(size_meta)) {
                /* integer literal */
                ASSERT1(is_val_exp(size_exp), size_exp->kind);
                size_val = &size_exp->u_val.val;
            }
            else {
                RETURN(ERROR_INVALID_SIZE_VAL, &size_exp->pos);
            }

            ASSERT1(is_int_val(size_val), size_val->kind);
            id->meta.arr_size[i] = int_val(size_val);
        }
    }

    return NO_ERROR;
}

static int
id_check_var(check_t *check, ast_id_t *id)
{
    ast_exp_t *type_exp;
    meta_t *type_meta;

    ASSERT1(is_var_id(id), id->kind);
    ASSERT(id->u_var.type_exp != NULL);

    type_exp = id->u_var.type_exp;
    type_meta = &type_exp->meta;

    ASSERT1(is_type_exp(type_exp), type_exp->kind);

    CHECK(exp_check(check, type_exp));

    meta_copy(&id->meta, type_meta);

    if (id->u_var.size_exps != NULL)
        CHECK(id_check_var_array(check, id, false));

    if (id->u_var.init_exp != NULL) {
        /* TODO: named initializer */
        ast_exp_t *init_exp = id->u_var.init_exp;
        meta_t *init_meta = &init_exp->meta;

        CHECK(exp_check(check, init_exp));

        /* This might be a redundant check, but is done to show 
         * more specific error not just mismatch error. */
        if (is_tuple_type(init_meta)) {
			if (!is_array_type(&id->meta) &&
                !is_map_type(type_meta) && !is_struct_type(type_meta))
                /* not allowed static initializer except map or struct */
                RETURN(ERROR_NOT_ALLOWED_INIT, &init_exp->pos);

            if (type_exp->id != NULL && is_contract_id(type_exp->id))
                /* contract object */
                RETURN(ERROR_NOT_ALLOWED_INIT, &init_exp->pos);
        }

		CHECK(meta_cmp(&id->meta, &init_exp->meta));

        if (is_val_exp(init_exp)) {
            if (!value_check(&init_exp->u_val.val, &id->meta))
                RETURN(ERROR_NUMERIC_OVERFLOW, &init_exp->pos, meta_to_str(&id->meta));

            id->val = &init_exp->u_val.val;
        }
    }
    else if (is_const_id(id)) {
        RETURN(ERROR_MISSING_CONST_VAL, &id->pos);
    }

    id->idx = check->var_idx++;

    return NO_ERROR;
}

static int
id_check_struct(check_t *check, ast_id_t *id)
{
    int i;
    array_t *fld_ids;

    ASSERT1(is_struct_id(id), id->kind);

    fld_ids = id->u_struc.fld_ids;
    ASSERT(fld_ids != NULL);

    for (i = 0; i < array_size(fld_ids); i++) {
        ast_id_t *fld_id = array_item(fld_ids, i, ast_id_t);

        CHECK(id_check_var(check, fld_id));

        flag_set(fld_id->mod, MOD_PUBLIC);
    }

    meta_set_struct(&id->meta, id->name, fld_ids);

    return NO_ERROR;
}

static int
id_check_enum(check_t *check, ast_id_t *id)
{
    int i, j;
    int enum_val = 0;
    array_t *elem_ids;

    ASSERT1(is_enum_id(id), id->kind);

    elem_ids = id->u_enum.elem_ids;
    ASSERT(elem_ids != NULL);

    for (i = 0; i < array_size(elem_ids); i++) {
        ast_id_t *elem_id = array_item(elem_ids, i, ast_id_t);
        ast_exp_t *init_exp = elem_id->u_var.init_exp;

        if (elem_id->u_var.init_exp == NULL) {
            ast_exp_t *val_exp = exp_new_val(&elem_id->pos);

            value_set_int(&val_exp->u_val.val, enum_val);

            elem_id->u_var.init_exp = val_exp;
        }
        else {
            meta_t *init_meta = &init_exp->meta;
            value_t *init_val;

            CHECK(exp_check(check, init_exp));

            if (!is_const_type(init_meta) || !is_dec_family(init_meta))
                RETURN(ERROR_INVALID_ENUM_VAL, &init_exp->pos);

            ASSERT1(is_val_exp(init_exp), init_exp->kind);

            init_val = &init_exp->u_val.val;
            ASSERT1(is_int_val(init_val), init_val->kind);

            for (j = 0; j < i; j++) {
                ast_id_t *prev_id = array_item(elem_ids, j, ast_id_t);

                if (prev_id->u_var.init_exp != NULL) {
                    value_t *prev_val = &prev_id->u_var.init_exp->u_val.val;

                    if (value_cmp(init_val, prev_val) == 0)
                        RETURN(ERROR_DUPLICATED_ENUM_VAL, &init_exp->pos);
                }
            }

            enum_val = int_val(init_val);
        }

        meta_set_int32(&elem_id->meta);

        flag_set(elem_id->mod, MOD_PUBLIC);
        enum_val++;
    }

    meta_set_int32(&id->meta);

    return NO_ERROR;
}

static int
id_check_param(check_t *check, ast_id_t *id)
{
    ast_exp_t *type_exp;

    ASSERT1(is_var_id(id), id->kind);
    ASSERT(id->u_var.type_exp != NULL);
    ASSERT(id->u_var.init_exp == NULL);

    type_exp = id->u_var.type_exp;
    ASSERT1(is_type_exp(type_exp), type_exp->kind);

    CHECK(exp_check(check, type_exp));

    meta_copy(&id->meta, &type_exp->meta);

    if (id->u_var.size_exps != NULL)
        CHECK(id_check_var_array(check, id, true));

    id->idx = check->var_idx++;

    return NO_ERROR;
}

static int
id_check_func(check_t *check, ast_id_t *id)
{
    int i;
    array_t *param_ids;
    array_t *ret_exps;

    ASSERT1(is_func_id(id), id->kind);

    param_ids = id->u_func.param_ids;

    for (i = 0; i < array_size(param_ids); i++) {
        ast_id_t *param_id = array_item(param_ids, i, ast_id_t);

        id_check_param(check, param_id);
    }

    ret_exps = id->u_func.ret_exps;

    if (ret_exps != NULL) {
        ast_exp_t *type_exp;

        if (array_size(ret_exps) == 1) {
            type_exp = array_item(ret_exps, 0, ast_exp_t);
            ASSERT1(is_type_exp(type_exp), type_exp->kind);

            exp_check(check, type_exp);
            meta_copy(&id->meta, &type_exp->meta);
        }
        else {
            for (i = 0; i < array_size(ret_exps); i++) {
                type_exp = array_item(ret_exps, i, ast_exp_t);
                ASSERT1(is_type_exp(type_exp), type_exp->kind);

                exp_check(check, type_exp);
            }
            meta_set_tuple(&id->meta, ret_exps);
        }
    }
    else if (is_ctor_id(id)) {
        meta_set_object(&id->meta);
    }
    else {
        meta_set_void(&id->meta);
    }

    if (id->u_func.blk != NULL) {
        check->func_id = id;

        blk_check(check, id->u_func.blk);

        check->func_id = NULL;
    }

    return NO_ERROR;
}

static int
id_check_contract(check_t *check, ast_id_t *id)
{
    ASSERT1(is_contract_id(id), id->kind);

    check->cont_id = id;

    if (id->u_cont.blk != NULL)
        blk_check(check, id->u_cont.blk);

    meta_set_object(&id->meta);

    check->cont_id = NULL;

    return NO_ERROR;
}

void
id_check(check_t *check, ast_id_t *id)
{
    ASSERT(id->name != NULL);

    switch (id->kind) {
    case ID_VAR:
        id_check_var(check, id);
        break;

    case ID_STRUCT:
        id_check_struct(check, id);
        break;

    case ID_ENUM:
        id_check_enum(check, id);
        break;

    case ID_FUNC:
        id_check_func(check, id);
        break;

    case ID_CONTRACT:
        id_check_contract(check, id);
        break;

    default:
        ASSERT1(!"invalid identifier", id->kind);
    }
}

/* end of check_id.c */
