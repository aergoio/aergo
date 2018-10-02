/**
 * @file    check_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "check_exp.h"
#include "check_blk.h"

#include "check_id.h"

static int
id_var_check(check_t *check, ast_id_t *id)
{
    ast_exp_t *type_exp;
    ast_meta_t *type_meta;

    ASSERT1(id_is_var(id), id->kind);
    ASSERT(id->u_var.type_exp != NULL);

    type_exp = id->u_var.type_exp;
    type_meta = &type_exp->meta;

    TRY(exp_type_check(check, type_exp));

    id->meta = *type_meta;

    if (id->u_var.arr_exp != NULL) {
        ast_exp_t *arr_exp = id->u_var.arr_exp;
        ast_meta_t *arr_meta = &arr_exp->meta;

        TRY(check_exp(check, arr_exp));

        if (arr_exp->kind == EXP_NULL && id->u_var.init_exp == NULL) 
            THROW(ERROR_MISSING_INITIALIZER, &id->pos);
        else if (!meta_is_integer(arr_meta))
            THROW(ERROR_INVALID_SIZE_TYPE, &id->pos, TYPENAME(arr_meta->type));

        id->meta.is_array = true;
    }

    if (id->u_var.init_exp != NULL) {
        array_t *val_exps;
        ast_exp_t *init_exp = id->u_var.init_exp;
        ast_id_t *type_id = type_exp->id;

        TRY(exp_tuple_check(check, init_exp));

        val_exps = init_exp->u_tup.exps;

        if (type_id != NULL) {
            int i;
            array_t *fld_ids;

            ASSERT1(id_is_struct(type_id), type_id->kind);
            fld_ids = type_id->u_st.fld_ids;

            if (array_size(fld_ids) != array_size(val_exps))
                THROW(ERROR_MISMATCHED_ELEM, &init_exp->pos);

            for (i = 0; i < array_size(fld_ids); i++) {
                ast_id_t *fld_id = array_item(fld_ids, i, ast_id_t);
                ast_exp_t *val_exp = array_item(val_exps, i, ast_exp_t);

                if (!meta_equals(&fld_id->meta, &val_exp->meta))
                    THROW(ERROR_MISMATCHED_TYPE, &val_exp->pos,
                          TYPENAME(fld_id->meta.type),
                          TYPENAME(val_exp->meta.type));
            }
        }
        else {
            ast_meta_t *val_meta;

            ASSERT1(array_size(val_exps) == 1, array_size(val_exps));
            val_meta = &array_item(val_exps, 0, ast_exp_t)->meta;

            if (!meta_equals(type_meta, val_meta)) 
                THROW(ERROR_MISMATCHED_TYPE, &init_exp->pos, 
                      TYPENAME(type_meta->type), TYPENAME(val_meta->type));
        }
    }

    return NO_ERROR;
}

static int
id_struct_check(check_t *check, ast_id_t *id)
{
    int i;
    array_t *fld_ids;

    ASSERT1(id_is_struct(id), id->kind);

    fld_ids = id->u_st.fld_ids;
    ASSERT(fld_ids != NULL);

    for (i = 0; i < array_size(fld_ids); i++) {
        ast_id_t *fld_id = array_item(fld_ids, i, ast_id_t);

        TRY(id_var_check(check, fld_id));
    }

    meta_set_prim(&id->meta, TYPE_STRUCT);

    return NO_ERROR;
}

static int
id_func_check(check_t *check, ast_id_t *id)
{
    int i;
    array_t *param_ids;
    array_t *ret_exps;

    ASSERT1(id_is_func(id), id->kind);

    param_ids = id->u_func.param_ids;
    if (param_ids != NULL) {
        for (i = 0; i < array_size(param_ids); i++) {
            ast_id_t *param_id = array_item(param_ids, i, ast_id_t);

            TRY(id_var_check(check, param_id));
        }
    }

    ret_exps = id->u_func.ret_exps;
    if (ret_exps != NULL) {
        for (i = 0; i < array_size(ret_exps); i++) {
            ast_exp_t *type_exp = array_item(ret_exps, i, ast_exp_t);

            TRY(exp_type_check(check, type_exp));
        }

        meta_set_tuple(&id->meta, ret_exps);
    }

    return NO_ERROR;
}

static int
id_contract_check(check_t *check, ast_id_t *id)
{
    ASSERT1(id_is_contract(id), id->kind);

    if (id->u_cont.blk != NULL)
        check_blk(check, id->u_cont.blk);

    return NO_ERROR;
}

int
check_id(check_t *check, ast_id_t *id)
{
    ASSERT(id->name != NULL);

    switch (id->kind) {
    case ID_VAR:
        return id_var_check(check, id);

    case ID_STRUCT:
        return id_struct_check(check, id);

    case ID_FUNC:
        return id_func_check(check, id);

    case ID_CONTRACT:
        return id_contract_check(check, id);

    default:
        ASSERT1(!"invalid identifier", id->kind);
    }

    return NO_ERROR;
}

/* end of check_id.c */
