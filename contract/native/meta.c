/**
 * @file    meta.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ast_exp.h"

#include "meta.h"

char *type_strs_[TYPE_MAX] = {
    "undefined",
    "bool",
    "byte",
    "int8",
    "uint8",
    "int16",
    "uint16",
    "int32",
    "uint32",
    "float",
    "int64",
    "uint64",
    "double",
    "string",
    "struct",
    "reference",
    "account",
    "map",
    "void",
    "tuple"
};

void
meta_set_struct(meta_t *meta, array_t *ids)
{
    int i;
    array_t *metas = array_new();

    meta_set(meta, TYPE_STRUCT);

    for (i = 0; i < array_size(ids); i++) {
        array_add_last(metas, &array_item(ids, i, ast_id_t)->meta);
    }

    meta->u_st.metas = metas;
}

void
meta_set_tuple(meta_t *meta, array_t *exps)
{
    int i;
    array_t *metas = array_new();

    meta_set(meta, TYPE_TUPLE);

    for (i = 0; i < array_size(exps); i++) {
        array_add_last(metas, &array_item(exps, i, ast_exp_t)->meta);
    }

    meta->u_st.metas = metas;
}

bool
meta_equals(meta_t *x, meta_t *y)
{
    int i;

    if (is_untyped_meta(x) || is_untyped_meta(y)) {
        if (x->type == y->type ||
            (is_integer_meta(x) && is_integer_meta(y)) ||
            (is_float_meta(x) && is_float_meta(y)))
            return true;

        return false;
    }

    if (is_map_meta(x) || is_map_meta(y)) {
        if (is_ref_meta(x) || is_ref_meta(y) ||
            (x->type == y->type &&
             x->u_map.k_type == y->u_map.k_type &&
             meta_equals(x->u_map.v_meta, y->u_map.v_meta)))
            return true;

        return false;
    }

    if (is_struct_meta(x) || is_struct_meta(y)) {
        array_t *x_metas = x->u_st.metas;
        array_t *y_metas = y->u_st.metas;

        if (x->type != y->type || array_size(x_metas) != array_size(y_metas))
            return false;

        for (i = 0; i < array_size(x_metas); i++) {
            if (!meta_equals(array_item(x_metas, i, meta_t),
                             array_item(y_metas, i, meta_t)))
                return false;
        }
        return true;
    }

    if (is_tuple_meta(x) || is_tuple_meta(y)) {
        array_t *x_metas = x->u_tup.metas;
        array_t *y_metas = y->u_tup.metas;

        if (x->type != y->type || array_size(x_metas) != array_size(y_metas))
            return false;

        for (i = 0; i < array_size(x_metas); i++) {
            if (!meta_equals(array_item(x_metas, i, meta_t),
                             array_item(y_metas, i, meta_t)))
                return false;
        }
        return true;
    }

    return x->type == y->type;
}


void
meta_dump(meta_t *meta, int indent)
{
}

/* end of meta.c */
