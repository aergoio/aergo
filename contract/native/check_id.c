/**
 * @file    check_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "array.h"
#include "ast_exp.h"
#include "ast_stmt.h"

#include "check_exp.h"
#include "check_blk.h"
#include "check_meta.h"
#include "check_stmt.h"

#include "check_id.h"

static int
id_check_var_array(check_t *check, ast_id_t *id, bool is_param)
{
    int i;
    int dim_sizes;
    array_t *size_exps = id->u_var.size_exps;

    meta_set_arr_dim(&id->meta, array_size(size_exps));

    for (i = 0; i < array_size(size_exps); i++) {
        ast_exp_t *size_exp = array_get_exp(size_exps, i);

        CHECK(exp_check(check, size_exp));

        if (is_null_exp(size_exp)) {
            if (!is_param && id->u_var.dflt_exp == NULL)
                RETURN(ERROR_MISSING_ARR_SIZE, &size_exp->pos);

            /* -1 means that the size is determined by the initializer */
            meta_set_dim_size(&id->meta, i, -1);
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

            dim_sizes = val_i64(size_val);
            if (dim_sizes <= 0)
                RETURN(ERROR_INVALID_SIZE_VAL, &size_exp->pos);

            meta_set_dim_size(&id->meta, i, dim_sizes);
        }
    }

    return NO_ERROR;
}

static int
id_check_var(check_t *check, ast_id_t *id, bool is_tuple)
{
    meta_t *type_meta;
    ast_exp_t *dflt_exp;

    ASSERT1(is_var_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->u_var.type_meta != NULL);

    id->is_checked = true;

    type_meta = id->u_var.type_meta;
    dflt_exp = id->u_var.dflt_exp;

    CHECK(meta_check(check, type_meta));

    meta_copy(&id->meta, type_meta);

    /* The default expression for constant variables of type tuple is set
     * in id_trans_tuple() */
    if (!is_tuple && is_const_id(id) && dflt_exp == NULL)
        RETURN(ERROR_MISSING_CONST_VAL, &id->pos);

    if (id->u_var.size_exps != NULL)
        CHECK(id_check_var_array(check, id, false));

    if (dflt_exp != NULL) {
        /* TODO: named initializer */
        CHECK(exp_check(check, dflt_exp));
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
id_check_tuple(check_t *check, ast_id_t *id)
{
    int i;
    array_t *var_ids = &id->u_tup.var_ids;
    ast_exp_t *dflt_exp = id->u_tup.dflt_exp;

    ASSERT1(is_tuple_id(id), id->kind);
    ASSERT(id->u_tup.type_meta != NULL);
    ASSERT(!is_empty_array(var_ids));

    id->meta.type = TYPE_TUPLE;

    id->meta.elem_cnt = array_size(var_ids);
    id->meta.elems = xmalloc(sizeof(meta_t *) * id->meta.elem_cnt);

    /* The size of the tuple identifier is not used,
     * so we do not need to set size here */
    for (i = 0; i < array_size(var_ids); i++) {
        ast_id_t *var_id = array_get_id(var_ids, i);

        var_id->mod = id->mod;
        var_id->scope = id->scope;
        var_id->u_var.type_meta = id->u_tup.type_meta;

        id_check_var(check, var_id, true);

        id->meta.elems[i] = &var_id->meta;

        if (is_const_id(var_id) && dflt_exp == NULL)
            ERROR(ERROR_MISSING_CONST_VAL, &var_id->pos);
    }

    if (dflt_exp != NULL) {
        CHECK(exp_check(check, dflt_exp));
        CHECK(meta_cmp(&id->meta, &dflt_exp->meta));
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
    ASSERT(id->name != NULL);

    id->is_checked = true;

    fld_ids = id->u_struc.fld_ids;
    ASSERT(fld_ids != NULL);

    for (i = 0; i < array_size(fld_ids); i++) {
        ast_id_t *fld_id = array_get_id(fld_ids, i);
        meta_t *fld_meta = &fld_id->meta;

        CHECK(id_check_var(check, fld_id, false));

        flag_set(fld_id->mod, MOD_PUBLIC);

        offset = ALIGN(offset, TYPE_ALIGN(fld_meta->type));

        fld_id->offset = offset;
        offset += meta_size(fld_meta);
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
    ASSERT(id->name != NULL);

    id->is_checked = true;

    elem_ids = id->u_enum.elem_ids;
    ASSERT(elem_ids != NULL);

    for (i = 0; i < array_size(elem_ids); i++) {
        ast_id_t *elem_id = array_get_id(elem_ids, i);
        ast_exp_t *dflt_exp = elem_id->u_var.dflt_exp;

        elem_id->is_checked = true;

        meta_set_int32(&elem_id->meta);

        if (dflt_exp == NULL) {
            elem_id->val = xmalloc(sizeof(value_t));

            value_init(elem_id->val);
            value_set_i64(elem_id->val, enum_val);
        }
        else {
            CHECK(exp_check(check, dflt_exp));

            if (!is_lit_exp(dflt_exp) || !is_i64_val(&dflt_exp->u_lit.val))
                RETURN(ERROR_INVALID_ENUM_VAL, &dflt_exp->pos);

            elem_id->val = &dflt_exp->u_lit.val;

            for (j = 0; j < i; j++) {
                ast_id_t *prev_id = array_get_id(elem_ids, j);

                if (value_cmp(elem_id->val, prev_id->val) == 0)
                    RETURN(ERROR_DUPLICATED_VALUE, &dflt_exp->pos, "enumerator");
            }

            enum_val = val_i64(elem_id->val);
        }

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
    ASSERT(id->name != NULL);
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
id_check_fn(check_t *check, ast_id_t *id)
{
    int i;
    array_t *param_ids;
    ast_id_t *ret_id;
    meta_t *id_meta;

    ASSERT1(is_fn_id(id), id->kind);
    ASSERT(id->name != NULL);

    id->is_checked = true;

    param_ids = id->u_fn.param_ids;

    for (i = 0; i < array_size(param_ids); i++) {
        ast_id_t *param_id = array_get_id(param_ids, i);

        id_check_param(check, param_id);
    }

    id_meta = &id->meta;
    ret_id = id->u_fn.ret_id;

    if (ret_id != NULL) {
        /* TODO: do we really need identifier not expression in here??? */
        if (is_tuple_id(ret_id))
            id_check_tuple(check, ret_id);
        else
            id_check_var(check, ret_id, false);

        meta_copy(id_meta, &ret_id->meta);
    }
    else if (is_ctor_id(id)) {
        meta_set_object(id_meta);
    }
    else {
        meta_set_void(id_meta);
    }

    if (id->u_fn.blk != NULL) {
        check->fn_id = id;

        blk_check(check, id->u_fn.blk);

        check->fn_id = NULL;
    }

    ASSERT(check->cont_id != NULL);

    id->u_fn.cont_id = check->cont_id;

    return NO_ERROR;
}

static int
id_check_contract(check_t *check, ast_id_t *id)
{
    ASSERT1(is_cont_id(id), id->kind);
    ASSERT(id->name != NULL);

    id->is_checked = true;

    check->cont_id = id;

    if (id->u_cont.blk != NULL)
        blk_check(check, id->u_cont.blk);

    meta_set_object(&id->meta);
    id->meta.name = id->name;

    check->cont_id = NULL;

    return NO_ERROR;
}

static int
id_check_label(check_t *check, ast_id_t *id)
{
    ASSERT1(is_label_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->u_lab.stmt != NULL);

    return NO_ERROR;
}

void
id_check(check_t *check, ast_id_t *id)
{
    switch (id->kind) {
    case ID_VAR:
        id_check_var(check, id, false);
        break;

    case ID_TUPLE:
        id_check_tuple(check, id);
        break;

    case ID_STRUCT:
        id_check_struct(check, id);
        break;

    case ID_ENUM:
        id_check_enum(check, id);
        break;

    case ID_FN:
        id_check_fn(check, id);
        break;

    case ID_CONTRACT:
        id_check_contract(check, id);
        break;

    case ID_LABEL:
        id_check_label(check, id);
        break;

    default:
        ASSERT1(!"invalid identifier", id->kind);
    }
}

/* end of check_id.c */
