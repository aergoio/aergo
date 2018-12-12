/**
 * @file    check_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "array.h"
#include "ast_exp.h"

#include "check_exp.h"
#include "check_blk.h"
#include "check_meta.h"

#include "check_id.h"

static int
id_check_var_array(check_t *check, ast_id_t *id, bool is_param)
{
    int i;
    int arr_size;
    array_t *size_exps = id->u_var.size_exps;

    meta_set_array(&id->meta, array_size(size_exps));

    for (i = 0; i < array_size(size_exps); i++) {
        ast_exp_t *size_exp = array_get(size_exps, i, ast_exp_t);

        CHECK(exp_check(check, size_exp));

        if (is_null_exp(size_exp)) {
            if (!is_param && id->u_var.dflt_exp == NULL)
                RETURN(ERROR_MISSING_ARR_SIZE, &size_exp->pos);

            id->meta.arr_size[i] = -1;
        }
        else {
            ast_id_t *size_id = size_exp->id;
            meta_t *size_meta = &size_exp->meta;
            value_t *size_val;

            if (size_id != NULL && size_id->val != NULL)
                /* constant variable */
                size_val = size_id->val;
            else if (is_integer_type(size_meta) && is_lit_exp(size_exp))
                /* integer literal */
                size_val = &size_exp->u_lit.val;
            else
                RETURN(ERROR_INVALID_SIZE_VAL, &size_exp->pos);

            ASSERT1(is_i64_val(size_val), size_val->type);
            arr_size = val_i64(size_val);

            if (arr_size <= 0)
                RETURN(ERROR_INVALID_SIZE_VAL, &size_exp->pos);

            id->meta.arr_size[i] = arr_size;
        }
    }

    return NO_ERROR;
}

static int
id_check_var(check_t *check, ast_id_t *id)
{
    meta_t *type_meta;

    ASSERT1(is_var_id(id), id->kind);
    ASSERT(id->u_var.type_meta != NULL);

    id->is_checked = true;

    type_meta = id->u_var.type_meta;

    CHECK(meta_check(check, type_meta));

    meta_copy(&id->meta, type_meta);

    if (is_const_id(id) && id->u_var.dflt_exp == NULL)
        RETURN(ERROR_MISSING_CONST_VAL, &id->pos);

    if (id->u_var.size_exps != NULL)
        CHECK(id_check_var_array(check, id, false));

    if (id->u_var.dflt_exp != NULL) {
        /* TODO: named initializer */
        ast_exp_t *dflt_exp = id->u_var.dflt_exp;
        meta_t *init_meta = &dflt_exp->meta;

        CHECK(exp_check(check, dflt_exp));

        /* This might be a duplicate check because it will be checked by meta_cmp(),
         * but is done to show more specific error not just mismatch error. */
        if (is_tuple_type(init_meta) && !is_array_type(&id->meta) &&
            !is_map_type(type_meta) && !is_struct_type(type_meta))
            /* not allowed static initializer except map or struct */
            RETURN(ERROR_NOT_ALLOWED_INIT, &dflt_exp->pos);

        CHECK(meta_cmp(&id->meta, &dflt_exp->meta));

        if (is_lit_exp(dflt_exp)) {
            if (!value_fit(&dflt_exp->u_lit.val, &id->meta))
                RETURN(ERROR_NUMERIC_OVERFLOW, &dflt_exp->pos, meta_to_str(&id->meta));

            id->val = &dflt_exp->u_lit.val;
        }
    }

    return NO_ERROR;
}

static int
id_check_struct(check_t *check, ast_id_t *id)
{
    int i;
    int offset = 0;
    array_t *fld_ids;

    ASSERT1(is_struct_id(id), id->kind);

    id->is_checked = true;

    fld_ids = id->u_struc.fld_ids;
    ASSERT(fld_ids != NULL);

    for (i = 0; i < array_size(fld_ids); i++) {
        ast_id_t *fld_id = array_get(fld_ids, i, ast_id_t);

        CHECK(id_check_var(check, fld_id));

        flag_set(fld_id->mod, MOD_PUBLIC);

        fld_id->offset = offset;
        offset += ALIGN64(fld_id->meta.size);
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

    id->is_checked = true;

    elem_ids = id->u_enum.elem_ids;
    ASSERT(elem_ids != NULL);

    for (i = 0; i < array_size(elem_ids); i++) {
        ast_id_t *elem_id = array_get(elem_ids, i, ast_id_t);
        ast_exp_t *dflt_exp = elem_id->u_var.dflt_exp;

        elem_id->is_checked = true;

        if (dflt_exp == NULL) {
            dflt_exp = exp_new_lit(&elem_id->pos);

            value_set_i64(&dflt_exp->u_lit.val, enum_val);

            CHECK(exp_check(check, dflt_exp));

            elem_id->u_var.dflt_exp = dflt_exp;
        }
        else {
            meta_t *init_meta = &dflt_exp->meta;
            value_t *init_val;

            CHECK(exp_check(check, dflt_exp));

            if (!is_lit_exp(dflt_exp) || !is_integer_type(init_meta))
                RETURN(ERROR_INVALID_ENUM_VAL, &dflt_exp->pos);

            init_val = &dflt_exp->u_lit.val;
            ASSERT1(is_i64_val(init_val), init_val->type);

            for (j = 0; j < i; j++) {
                ast_id_t *prev_id = array_get(elem_ids, j, ast_id_t);

                if (prev_id->u_var.dflt_exp != NULL) {
                    value_t *val_prev = &prev_id->u_var.dflt_exp->u_lit.val;

                    if (value_cmp(init_val, val_prev) == 0)
                        RETURN(ERROR_DUPLICATED_VALUE, &dflt_exp->pos, "enumerator");
                }
            }

            enum_val = val_i64(init_val);
        }

        elem_id->val = &dflt_exp->u_lit.val;
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
    meta_t *type_meta;

    ASSERT1(is_var_id(id), id->kind);
    ASSERT(id->u_var.type_meta != NULL);
    ASSERT(id->u_var.dflt_exp == NULL);

    id->is_checked = true;

    type_meta = id->u_var.type_meta;

    CHECK(meta_check(check, type_meta));

    meta_copy(&id->meta, type_meta);

    if (id->u_var.size_exps != NULL)
        CHECK(id_check_var_array(check, id, true));

    return NO_ERROR;
}

static int
id_check_func(check_t *check, ast_id_t *id)
{
    int i;
    array_t *param_ids;
    array_t *ret_ids;
    meta_t *id_meta;

    ASSERT1(is_func_id(id), id->kind);

    id->is_checked = true;

    param_ids = id->u_func.param_ids;

    for (i = 0; i < array_size(param_ids); i++) {
        ast_id_t *param_id = array_get(param_ids, i, ast_id_t);

        id_check_param(check, param_id);
    }

    id_meta = &id->meta;
    ret_ids = id->u_func.ret_ids;

    if (ret_ids != NULL) {
        ast_id_t *ret_id;

        if (array_size(ret_ids) == 1) {
            ret_id = array_get(ret_ids, 0, ast_id_t);

            id_check_var(check, ret_id);
            meta_copy(id_meta, &ret_id->meta);
        }
        else {
            /* TODO: add API for meta_set_tuple() from ids */
            meta_set(id_meta, TYPE_TUPLE);

            id_meta->elem_cnt = array_size(ret_ids);
            id_meta->elems = xmalloc(sizeof(meta_t *) * id_meta->elem_cnt);

            for (i = 0; i < array_size(ret_ids); i++) {
                ret_id = array_get(ret_ids, i, ast_id_t);

                id_check_var(check, ret_id);
                id_meta->elems[i] = &ret_id->meta;
            }
        }
    }
    else if (is_ctor_id(id)) {
        meta_set_object(id_meta);
    }
    else {
        meta_set_void(id_meta);
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

    id->is_checked = true;

    check->cont_id = id;

    if (id->u_cont.blk != NULL)
        blk_check(check, id->u_cont.blk);

    meta_set_object(&id->meta);
    id->meta.name = id->name;

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
