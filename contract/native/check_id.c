/**
 * @file    check_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_exp.h"
#include "ast_stmt.h"
#include "check_exp.h"
#include "check_blk.h"
#include "check_stmt.h"

#include "check_id.h"

static bool
id_check_array(check_t *check, ast_id_t *id)
{
    int i;
    int dim_sz;
    ast_exp_t *dflt_exp = id->u_var.dflt_exp;
    vector_t *size_exps = id->u_var.size_exps;

    ASSERT1(is_var_id(id), id->kind);

    meta_set_arr_dim(&id->meta, vector_size(size_exps));

    vector_foreach(size_exps, i) {
        ast_exp_t *size_exp = vector_get_exp(size_exps, i);

        CHECK(exp_check(check, size_exp));

        if (is_null_exp(size_exp)) {
            if (!is_param_id(id) && dflt_exp == NULL)
                RETURN(ERROR_MISSING_ARR_SIZE, &size_exp->pos);

            id->meta.is_fixed = false;
        }
        else {
            value_t *size_val = NULL;

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

            meta_set_dim_sz(&id->meta, i, dim_sz);
        }
    }

    if (id->meta.is_fixed && dflt_exp != NULL && is_alloc_exp(dflt_exp))
        RETURN(ERROR_NOT_ALLOWED_ALLOC, &dflt_exp->pos);

    return true;
}

static bool
id_check_var(check_t *check, ast_id_t *id)
{
    ast_exp_t *dflt_exp = id->u_var.dflt_exp;

    ASSERT1(is_var_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->up != NULL);
    ASSERT(id->u_var.type_exp != NULL);

    CHECK(exp_check(check, id->u_var.type_exp));

    meta_copy(&id->meta, &id->u_var.type_exp->meta);

    if (id->u_var.size_exps != NULL)
        CHECK(id_check_array(check, id));

    if (dflt_exp != NULL) {
        /* TODO: named initializer */
        CHECK(exp_check(check, dflt_exp));
        CHECK(meta_cmp(&id->meta, &dflt_exp->meta));

        meta_eval(&id->meta, &dflt_exp->meta);

        exp_check_overflow(dflt_exp, &id->meta);

        if (is_const_id(id) && is_lit_exp(dflt_exp))
            id->val = &dflt_exp->u_lit.val;
    }
    else if (is_const_id(id)) {
        RETURN(ERROR_MISSING_CONST_VAL, &id->pos);
    }

    return true;
}

static bool
id_check_struct(check_t *check, ast_id_t *id)
{
    int i;
    uint32_t offset = 0;
    vector_t *fld_ids = id->u_struc.fld_ids;

    ASSERT1(is_struct_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->up != NULL);
    ASSERT(fld_ids != NULL);

    vector_foreach(fld_ids, i) {
        ast_id_t *fld_id = vector_get_id(fld_ids, i);
        meta_t *fld_meta = &fld_id->meta;

        fld_id->up = id;
        fld_id->is_checked = true;

        if (!id_check_var(check, fld_id))
            continue;

        BIT_SET(fld_id->mod, MOD_PUBLIC);

        ASSERT1(fld_meta->rel_addr == 0, fld_meta->rel_addr);

        fld_meta->rel_offset = ALIGN(offset, meta_align(fld_meta));
        offset += meta_memsz(fld_meta);
    }

    meta_set_struct(&id->meta, id);

    return true;
}

static bool
id_check_enum(check_t *check, ast_id_t *id)
{
    int i, j;
    int enum_val = 0;
    vector_t *elem_ids;

    ASSERT1(is_enum_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->up != NULL);

    elem_ids = id->u_enum.elem_ids;
    ASSERT(elem_ids != NULL);

    vector_foreach(elem_ids, i) {
        ast_id_t *elem_id = vector_get_id(elem_ids, i);
        ast_exp_t *dflt_exp = elem_id->u_var.dflt_exp;

        /* Check directly for processing value in the enumerator */
        elem_id->up = id;
        elem_id->is_checked = true;

        meta_set_int32(&elem_id->meta);
        meta_set_undef(&elem_id->meta);

        if (dflt_exp == NULL) {
            elem_id->val = xmalloc(sizeof(value_t));

            value_set_int(elem_id->val, enum_val);
        }
        else {
            CHECK(exp_check(check, dflt_exp));

            if (!is_lit_exp(dflt_exp) || !is_integer_meta(&dflt_exp->meta))
                RETURN(ERROR_INVALID_ENUM_VAL, &dflt_exp->pos);

            elem_id->val = &dflt_exp->u_lit.val;

            for (j = 0; j < i; j++) {
                ast_id_t *prev_id = vector_get_id(elem_ids, j);

                if (value_cmp(elem_id->val, prev_id->val) == 0)
                    RETURN(ERROR_DUPLICATED_ENUM, &dflt_exp->pos);
            }

            enum_val = val_i64(elem_id->val);
        }

        enum_val++;
    }

    meta_set_int32(&id->meta);

    return true;
}

static bool
id_check_param(check_t *check, ast_id_t *id)
{
    ASSERT1(is_param_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->up != NULL);
    ASSERT(id->u_var.type_exp != NULL);
    ASSERT(id->u_var.dflt_exp == NULL);

    CHECK(exp_check(check, id->u_var.type_exp));

    meta_copy(&id->meta, &id->u_var.type_exp->meta);

    if (id->u_var.size_exps != NULL)
        CHECK(id_check_array(check, id));

    return true;
}

static bool
id_check_fn(check_t *check, ast_id_t *id)
{
    int i;
    vector_t *param_ids = id->u_fn.param_ids;

    ASSERT1(is_fn_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->up != NULL);
    ASSERT1(is_root_id(id->up), id->up->kind);

    if (is_ctor_id(id) && strcmp(id->name, id->up->name) != 0)
        ERROR(ERROR_MISMATCHED_NAME, &id->pos, id->up->name, id->name);

    vector_foreach(param_ids, i) {
        ast_id_t *param_id = vector_get_id(param_ids, i);

        param_id->up = id;
        param_id->is_checked = true;

        CHECK(id_check_param(check, param_id));
    }

    if (id->u_fn.ret_id != NULL) {
        ast_id_t *ret_id = id->u_fn.ret_id;

        ret_id->up = id;
        ret_id->is_checked = true;

        CHECK(id_check_param(check, ret_id));

        meta_copy(&id->meta, &ret_id->meta);
    }
    else {
        meta_set_void(&id->meta);
    }

    if (check->impl_id != NULL) {
        /* Mark "is_used" flag */
        ast_blk_t *blk = check->impl_id->u_itf.blk;

        vector_foreach(&blk->ids, i) {
            ast_id_t *spec_id = vector_get_id(&blk->ids, i);

            if (strcmp(spec_id->name, id->name) == 0 && id_cmp(spec_id, id)) {
                spec_id->is_used = true;
                break;
            }
        }
    }

    snprintf(id->u_fn.qname, sizeof(id->u_fn.qname), "%s.%s", id->up->name,
             id->u_fn.alias != NULL ? id->u_fn.alias : id->name);

    /* The body of the function is checked after the resolution of all identifiers.
     * (See blk_check()) */

    return true;
}

static bool
id_check_contract(check_t *check, ast_id_t *id)
{
    int i;
    ast_exp_t *impl_exp = id->u_cont.impl_exp;

    ASSERT1(is_cont_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->up == NULL);

    if (impl_exp != NULL) {
        if (exp_check(check, impl_exp)) {
            ast_id_t *impl_id = impl_exp->id;

            /* Unmark "is_used" flag */
            ASSERT(impl_id != NULL);
            ASSERT1(is_itf_id(impl_id), impl_id->kind);

            vector_foreach(&impl_id->u_itf.blk->ids, i) {
                vector_get_id(&impl_id->u_itf.blk->ids, i)->is_used = false;
            }

            check->impl_id = impl_id;
        }
    }

    /* It can be used the contract variable in the block, so the meta is set before blk_check() */
    meta_set_object(&id->meta, id);

    check->cont_id = id;

    if (id->u_cont.blk != NULL)
        blk_check(check, id->u_cont.blk);

    check->cont_id = NULL;

    if (check->impl_id != NULL) {
        ast_blk_t *blk = check->impl_id->u_itf.blk;

        vector_foreach(&blk->ids, i) {
            ast_id_t *spec_id = vector_get_id(&blk->ids, i);

            if (!spec_id->is_used)
                ERROR(ERROR_NOT_IMPLEMENTED, &id->pos, spec_id->name);
        }

        check->impl_id = NULL;
    }

    return true;
}

static bool
id_check_interface(check_t *check, ast_id_t *id)
{
    ASSERT1(is_itf_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->up == NULL);
    ASSERT(id->u_itf.blk != NULL);

    meta_set_object(&id->meta, id);

    blk_check(check, id->u_itf.blk);

    return true;
}

static bool
id_check_library(check_t *check, ast_id_t *id)
{
    ASSERT1(is_lib_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->up == NULL);
    ASSERT(id->u_lib.blk != NULL);

    meta_set_void(&id->meta);

    blk_check(check, id->u_lib.blk);

    return true;
}

static bool
id_check_label(check_t *check, ast_id_t *id)
{
    ASSERT1(is_label_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->up != NULL);
    ASSERT(id->u_lab.stmt != NULL);

    return true;
}

static bool
id_check_tuple(check_t *check, ast_id_t *id)
{
    int i;
    vector_t *elem_ids = id->u_tup.elem_ids;
    ast_exp_t *dflt_exp = id->u_tup.dflt_exp;

    ASSERT1(is_tuple_id(id), id->kind);

    if (dflt_exp != NULL) {
        if (!is_tuple_exp(dflt_exp))
            ERROR(ERROR_MISMATCHED_COUNT, &dflt_exp->pos, "assignment", vector_size(elem_ids), 1);
        else if (vector_size(elem_ids) != vector_size(dflt_exp->u_tup.elem_exps))
            ERROR(ERROR_MISMATCHED_COUNT, &dflt_exp->pos, "assignment", vector_size(elem_ids),
                  vector_size(dflt_exp->u_tup.elem_exps));
    }

    id->meta.type = TYPE_TUPLE;

    id->meta.elem_cnt = vector_size(elem_ids);
    id->meta.elems = xmalloc(sizeof(meta_t *) * id->meta.elem_cnt);

    vector_foreach(elem_ids, i) {
        ast_id_t *elem_id = vector_get_id(elem_ids, i);

        ASSERT1(is_var_id(elem_id), elem_id->kind);
        ASSERT(elem_id->u_var.type_exp == NULL);

        elem_id->mod = id->mod;
        elem_id->u_var.type_exp = id->u_tup.type_exp;

        if (dflt_exp != NULL && is_tuple_exp(dflt_exp) &&
            i < vector_size(dflt_exp->u_tup.elem_exps))
            elem_id->u_var.dflt_exp = vector_get_exp(dflt_exp->u_tup.elem_exps, i);

        elem_id->up = id->up;
        elem_id->is_checked = true;

        id_check_var(check, elem_id);

        id->meta.elems[i] = &elem_id->meta;
    }

    return true;
}

void
id_check(check_t *check, ast_id_t *id)
{
    ASSERT(id != NULL);

    if (id->is_checked)
        return;

    id->up = check->id;
    check->id = id;

    id->is_checked = true;

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

    case ID_FN:
        id_check_fn(check, id);
        break;

    case ID_CONT:
        id_check_contract(check, id);
        break;

    case ID_ITF:
        id_check_interface(check, id);
        break;

    case ID_LIB:
        id_check_library(check, id);
        break;

    case ID_LABEL:
        id_check_label(check, id);
        break;

    case ID_TUPLE:
        id_check_tuple(check, id);
        break;

    default:
        ASSERT1(!"invalid identifier", id->kind);
    }

    check->id = id->up;
}

void
id_trycheck(check_t *check, ast_id_t *id)
{
    ast_id_t *org_id = check->id;
    ast_id_t *org_impl_id = check->impl_id;
    ast_id_t *org_qual_id = check->qual_id;

    if (is_root_id(id))
        check->id = NULL;
    else if (is_fn_id(id))
        check->id = check->qual_id != NULL ? check->qual_id : check->cont_id;

    id_check(check, id);

    check->id = org_id;
    check->impl_id = org_impl_id;
    check->qual_id = org_qual_id;
}

/* end of check_id.c */
