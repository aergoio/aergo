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
id_var_check(check_t *check, ast_id_t *id)
{
    ast_exp_t *type_exp;
    meta_t *type_meta;
    ast_exp_t *init_exp;

    ASSERT1(is_var_id(id), id->kind);
    ASSERT(id->u_var.type_exp != NULL);

    type_exp = id->u_var.type_exp;
    type_meta = &type_exp->meta;

    ASSERT1(is_type_exp(type_exp), type_exp->kind);

    TRY(check_exp(check, type_exp));

    if (type_exp->u_type.is_local)
        flag_set(id->mod, MOD_LOCAL);

    id->meta = *type_meta;

    init_exp = id->u_var.init_exp;

    if (id->u_var.arr_exp != NULL) {
        ast_exp_t *arr_exp = id->u_var.arr_exp;
        meta_t *arr_meta = &arr_exp->meta;

        id->meta.is_array = true;

        TRY(check_exp(check, arr_exp));

        if (is_null_exp(arr_exp)) {
            if (init_exp == NULL)
                THROW(ERROR_MISSING_INITIALIZER, id_pos(id));
        }
        else if (!is_integer_meta(arr_meta)) {
            THROW(ERROR_INVALID_SIZE_TYPE, id_pos(id), TYPE_NAME(arr_meta->type));
        }
    }

    if (init_exp != NULL) {
        /* TODO: check size of initializer */
        meta_t *init_meta = &init_exp->meta;

        TRY(check_exp(check, init_exp));

        if (is_tuple_meta(init_meta)) {
            int i;
            ast_id_t *type_id = type_exp->id;
            array_t *val_metas = init_meta->u_tup.metas;

            if (type_id == NULL) {
                /* array of primitive type */
                for (i = 0; i < array_size(val_metas); i++) {
                    meta_t *val_meta = array_item(val_metas, i, meta_t);

                    if (!meta_equals(type_meta, val_meta))
                        THROW(ERROR_MISMATCHED_TYPE, meta_pos(val_meta),
                              TYPE_NAME(type_meta->type),
                              TYPE_NAME(val_meta->type));
                }
            }
            else {
                array_t *fld_metas;

                ASSERT1(is_struct_id(type_id), type_id->kind);
                ASSERT1(is_struct_meta(type_meta), type_meta->type);

                fld_metas = type_meta->u_st.metas;

                if (array_size(fld_metas) != array_size(val_metas))
                    THROW(ERROR_MISMATCHED_COUNT, exp_pos(init_exp),
                          array_size(fld_metas), array_size(val_metas));

                for (i = 0; i < array_size(fld_metas); i++) {
                    meta_t *fld_meta = array_item(fld_metas, i, meta_t);
                    meta_t *val_meta = array_item(val_metas, i, meta_t);

                    if (!meta_equals(fld_meta, val_meta))
                        THROW(ERROR_MISMATCHED_TYPE, meta_pos(val_meta),
                              TYPE_NAME(fld_meta->type),
                              TYPE_NAME(val_meta->type));
                }
            }
        }
        else if (!meta_equals(type_meta, &init_exp->meta)) {
            THROW(ERROR_MISMATCHED_TYPE, exp_pos(init_exp),
                  TYPE_NAME(type_meta->type), TYPE_NAME(init_exp->meta.type));
        }
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

        TRY(id_var_check(check, fld_id));
    }

    meta_set_struct(&id->meta, fld_ids);

    return NO_ERROR;
}

static int
id_param_check(check_t *check, ast_id_t *id)
{
    ast_exp_t *type_exp;
    meta_t *type_meta;

    ASSERT1(is_var_id(id), id->kind);
    ASSERT(id->u_var.type_exp != NULL);
    ASSERT(id->u_var.init_exp == NULL);

    type_exp = id->u_var.type_exp;
    type_meta = &type_exp->meta;

    ASSERT1(is_type_exp(type_exp), type_exp->kind);

    TRY(check_exp(check, type_exp));

    id->meta = *type_meta;

    if (id->u_var.arr_exp != NULL) {
        ast_exp_t *arr_exp = id->u_var.arr_exp;
        meta_t *arr_meta = &arr_exp->meta;

        id->meta.is_array = true;

        TRY(check_exp(check, arr_exp));

        if (!is_null_exp(arr_exp) && !is_integer_meta(arr_meta))
            THROW(ERROR_INVALID_SIZE_TYPE, id_pos(id), TYPE_NAME(arr_meta->type));
    }

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

            check_exp(check, type_exp);
        }

        meta_set_tuple(&id->meta, ret_exps);
    }
    else if (id->mod == MOD_INITIAL) {
        meta_set_ref(&id->meta);
    }
    else {
        meta_set_void(&id->meta);
    }

    check->fn_id = id;

    if (id->u_func.blk != NULL)
        check_blk(check, id->u_func.blk);

    check->fn_id = NULL;

    return NO_ERROR;
}

static int
id_contract_check(check_t *check, ast_id_t *id)
{
    ASSERT1(is_contract_id(id), id->kind);

    if (id->u_cont.blk != NULL)
        check_blk(check, id->u_cont.blk);

    meta_set_ref(&id->meta);

    return NO_ERROR;
}

void
check_id(check_t *check, ast_id_t *id)
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
