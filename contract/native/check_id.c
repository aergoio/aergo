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
#include "check_stmt.h"

#include "check_id.h"

#if 0
static bool
id_check_type(check_t *check, meta_t *meta)
{
    if (is_none_meta(meta)) {
        ast_id_t *id;

        ASSERT(meta->name != NULL);

        id = blk_search_id(check->blk, meta->name, meta->num, true);
        if (id == NULL || !is_type_id(id))
            RETURN(ERROR_UNDEFINED_TYPE, meta->pos, meta->name);

        id_trycheck(check, id);

        id->is_used = true;

        meta_copy(meta, &id->meta);
    }
    else if (is_map_meta(meta)) {
        meta_t *k_meta, *v_meta;

        ASSERT1(meta->elem_cnt == 2, meta->elem_cnt);

        k_meta = meta->elems[0];
        v_meta = meta->elems[1];

        CHECK(id_check_type(check, k_meta));
        CHECK(id_check_type(check, v_meta));

        if (!is_comparable_meta(k_meta))
            RETURN(ERROR_NOT_COMPARABLE_TYPE, k_meta->pos, meta_to_str(k_meta));

        ASSERT(!is_tuple_meta(v_meta));
    }

    return true;
}
#endif

static bool
id_check_array(check_t *check, ast_id_t *id)
{
    int i;
    int dim_size;
    array_t *size_exps;

    if (is_var_id(id))
        size_exps = id->u_var.size_exps;
    else
        size_exps = id->u_ret.size_exps;

    meta_set_arr_dim(&id->meta, array_size(size_exps));

    array_foreach(size_exps, i) {
        ast_exp_t *size_exp = array_get_exp(size_exps, i);

        CHECK(exp_check(check, size_exp));

        if (is_null_exp(size_exp)) {
            if (!is_param_id(id) && is_var_id(id) && id->u_var.dflt_exp == NULL)
                RETURN(ERROR_MISSING_ARR_SIZE, &size_exp->pos);

            /* -1 means that the size is determined by the initializer */
            meta_set_dim_size(&id->meta, i, -1);
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
            ASSERT1(is_i64_val(size_val), size_val->type);

            dim_size = val_i64(size_val);
            if (dim_size <= 0)
                RETURN(ERROR_INVALID_SIZE_VAL, &size_exp->pos);

            meta_set_dim_size(&id->meta, i, dim_size);
        }
    }

    return true;
}

static bool
id_check_var(check_t *check, ast_id_t *id)
{
    ASSERT1(is_var_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->up != NULL);
    ASSERT(id->u_var.type_exp != NULL);

    CHECK(exp_check(check, id->u_var.type_exp));

    meta_copy(&id->meta, &id->u_var.type_exp->meta);

    if (id->u_var.size_exps != NULL)
        CHECK(id_check_array(check, id));

    if (id->u_var.dflt_exp != NULL) {
        ast_exp_t *dflt_exp = id->u_var.dflt_exp;

        /* TODO: named initializer */
        CHECK(exp_check(check, dflt_exp));
        CHECK(meta_cmp(&id->meta, &dflt_exp->meta));

        meta_eval(&id->meta, &dflt_exp->meta);

        exp_check_overflow(dflt_exp, &id->meta);

        if (is_const_id(id) && is_lit_exp(dflt_exp))
            id->val = &dflt_exp->u_lit.val;
    }
    else if (!is_tuple_id(id->up) && is_const_id(id)) {
        /* The constant value of the tuple element is checked by id_check_tuple() */
        RETURN(ERROR_MISSING_CONST_VAL, &id->pos);
    }

    return true;
}

static bool
id_check_struct(check_t *check, ast_id_t *id)
{
    int i;
    uint32_t offset = 0;
    array_t *fld_ids;

    ASSERT1(is_struct_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->up != NULL);

    fld_ids = id->u_struc.fld_ids;
    ASSERT(fld_ids != NULL);

    array_foreach(fld_ids, i) {
        ast_id_t *fld_id = array_get_id(fld_ids, i);

        ASSERT1(is_var_id(fld_id), fld_id->kind);

        id_check(check, fld_id);

        flag_set(fld_id->mod, MOD_PUBLIC);

        meta_set_rel_offset(&fld_id->meta, &offset);
    }

    meta_set_struct(&id->meta, id);

    return true;
}

static bool
id_check_enum(check_t *check, ast_id_t *id)
{
    int i, j;
    int enum_val = 0;
    array_t *elem_ids;

    ASSERT1(is_enum_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->up != NULL);

    elem_ids = id->u_enum.elem_ids;
    ASSERT(elem_ids != NULL);

    array_foreach(elem_ids, i) {
        ast_id_t *elem_id = array_get_id(elem_ids, i);
        ast_exp_t *dflt_exp = elem_id->u_var.dflt_exp;

        /* check directly for value processing in the enumerator */
        elem_id->up = id;
        elem_id->is_checked = true;

        meta_set_int32(&elem_id->meta);

        if (dflt_exp == NULL) {
            elem_id->val = xmalloc(sizeof(value_t));

            value_init(elem_id->val);
            value_set_i64(elem_id->val, enum_val);
        }
        else {
            CHECK(exp_check(check, dflt_exp));

            if (!is_lit_exp(dflt_exp) || !is_integer_meta(&dflt_exp->meta))
                RETURN(ERROR_INVALID_ENUM_VAL, &dflt_exp->pos);

            elem_id->val = &dflt_exp->u_lit.val;

            for (j = 0; j < i; j++) {
                ast_id_t *prev_id = array_get_id(elem_ids, j);

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
id_check_return(check_t *check, ast_id_t *id)
{
    ASSERT1(is_return_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->up != NULL);
    ASSERT(is_param_id(id));
    ASSERT(id->u_ret.type_exp != NULL);

    CHECK(exp_check(check, id->u_ret.type_exp));

    meta_copy(&id->meta, &id->u_ret.type_exp->meta);

    if (id->u_ret.size_exps != NULL)
        CHECK(id_check_array(check, id));

    return true;
}

static bool
id_check_fn(check_t *check, ast_id_t *id)
{
    int i;
    array_t *param_ids = id->u_fn.param_ids;

    ASSERT1(is_fn_id(id), id->kind);
    ASSERT(id->name != NULL);
    ASSERT(id->up != NULL);
    ASSERT1(is_cont_id(id->up) || is_itf_id(id->up), id->up->kind);

    array_foreach(param_ids, i) {
        ast_id_t *param_id = array_get_id(param_ids, i);

        ASSERT1(is_var_id(param_id), param_id->kind);
        ASSERT(is_param_id(param_id));
        ASSERT(param_id->u_var.dflt_exp == NULL);

        id_check(check, param_id);
    }

    if (id->u_fn.ret_id != NULL) {
        /* The return identifier may be a tuple */
        id_check(check, id->u_fn.ret_id);

        meta_copy(&id->meta, &id->u_fn.ret_id->meta);
    }
    else {
        meta_set_void(&id->meta);
    }

    if (check->impl_id != NULL) {
        /* mark is_used flag */
        ast_blk_t *blk = check->impl_id->u_itf.blk;

        array_foreach(&blk->ids, i) {
            ast_id_t *spec_id = array_get_id(&blk->ids, i);

            if (strcmp(spec_id->name, id->name) == 0 && id_cmp(spec_id, id)) {
                spec_id->is_used = true;
                break;
            }
        }
    }

    check->fn_id = id;

    if (id->u_fn.blk != NULL)
        blk_check(check, id->u_fn.blk);

    check->fn_id = NULL;

    if (id->u_fn.ret_id != NULL && !is_ctor_id(id) && !is_itf_id(id->up) &&
        (id->u_fn.blk == NULL || is_empty_array(&id->u_fn.blk->stmts) ||
         !is_return_stmt(array_get_last(&id->u_fn.blk->stmts, ast_stmt_t))))
        RETURN(ERROR_MISSING_RETURN, &id->pos);

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

            /* unmark is_used flag */
            ASSERT(impl_id != NULL);
            ASSERT1(is_itf_id(impl_id), impl_id->kind);

            array_foreach(&impl_id->u_itf.blk->ids, i) {
                array_get_id(&impl_id->u_itf.blk->ids, i)->is_used = false;
            }

            check->impl_id = impl_id;
        }
    }

    /* It can be used the contract variable in the contract block,
     * so the meta is set before blk_check() */
    meta_set_object(&id->meta, id);

    check->cont_id = id;

    if (id->u_cont.blk != NULL)
        blk_check(check, id->u_cont.blk);

    check->cont_id = NULL;

    if (check->impl_id != NULL) {
        ast_blk_t *blk = check->impl_id->u_itf.blk;

        array_foreach(&blk->ids, i) {
            ast_id_t *spec_id = array_get_id(&blk->ids, i);

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
    array_t *elem_ids = id->u_tup.elem_ids;
    ast_exp_t *dflt_exp = id->u_tup.dflt_exp;

    ASSERT1(is_tuple_id(id), id->kind);

    id->meta.type = TYPE_TUPLE;

    id->meta.elem_cnt = array_size(elem_ids);
    id->meta.elems = xmalloc(sizeof(meta_t *) * id->meta.elem_cnt);

    id->meta.size = 0;

    /* The meta size of the tuple identifier is never used, so we do not need to set
     * size here */
    array_foreach(elem_ids, i) {
        ast_id_t *elem_id = array_get_id(elem_ids, i);

        elem_id->mod = id->mod;

        if (is_var_id(elem_id)) {
            /* The default expression for the tuple identifier is set in id_trans_var() */
            ASSERT(elem_id->u_var.dflt_exp == NULL);

            if (elem_id->u_var.type_exp == NULL)
                elem_id->u_var.type_exp = id->u_tup.type_exp;

            id_check(check, elem_id);
        }
        else {
            ASSERT1(is_return_id(elem_id), elem_id->kind);
            id_check(check, elem_id);
        }

        elem_id->up = id->up;

        id->meta.elems[i] = &elem_id->meta;

        id->meta.size = ALIGN(id->meta.size, meta_align(&elem_id->meta));
        id->meta.size += meta_size(&elem_id->meta);

        if (id->meta.align == 0)
            id->meta.align = meta_align(&elem_id->meta);

        if (is_const_id(elem_id) && dflt_exp == NULL)
            ERROR(ERROR_MISSING_CONST_VAL, &elem_id->pos);
    }

    id->meta.size = ALIGN(id->meta.size, id->meta.align);

    if (dflt_exp != NULL) {
        CHECK(exp_check(check, dflt_exp));
        CHECK(meta_cmp(&id->meta, &dflt_exp->meta));

        meta_eval(&id->meta, &dflt_exp->meta);
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

    case ID_RETURN:
        id_check_return(check, id);
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

    /* TODO: Currently it can be a variable identifier due to self-reference */

    if (is_cont_id(id) || is_itf_id(id))
        check->id = NULL;
    else if (is_fn_id(id))
        check->id = check->cont_id;

    id_check(check, id);

    check->id = org_id;
}

/* end of check_id.c */
