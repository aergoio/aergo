/**
 * @file    meta.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ast_exp.h"

#include "meta.h"

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
             meta_equals(x->u_map.k_meta, y->u_map.k_meta) &&
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

static int
meta_cmp_map(meta_t *x, meta_t *y)
{
    meta_t *k_meta;
    meta_t *v_meta;
    array_t *elems;

    if ((!is_map_meta(x) && !is_tuple_meta(x)) ||
        (!is_map_meta(y) && !is_tuple_meta(y)))
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, META_NAME(x), META_NAME(y));

    if (is_map_meta(x) && is_map_meta(y)) {
        CHECK(meta_cmp(x->u_map.k_meta, y->u_map.k_meta));

        return meta_cmp(x->u_map.v_meta, y->u_map.v_meta);
    }

    if (is_map_meta(x)) {
        k_meta = x->u_map.k_meta;
        v_meta = x->u_map.v_meta;
        elems = y->u_tup.metas;
    }
    else {
        k_meta = y->u_map.k_meta;
        v_meta = y->u_map.v_meta;
        elems = x->u_tup.metas;
    }

    if (array_size(elems) != 2)
        RETURN(ERROR_MISMATCHED_ELEM_CNT, y->pos, 2, array_size(elems));

    /* If one side is a tuple, it may be insufficient to check only by type comparison. 
     * So, we have to check untyped flag. */
    CHECK(meta_cmp(k_meta, array_item(elems, 0, meta_t)));

    return meta_cmp(v_meta, array_item(elems, 1, meta_t));
}

static int
meta_cmp_struct(meta_t *x, meta_t *y)
{
    int i;
    array_t *x_elems, *y_elems;

    if ((!is_struct_meta(x) && !is_tuple_meta(x)) ||
        (!is_struct_meta(y) && !is_tuple_meta(y)))
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, META_NAME(x), META_NAME(y));

    x_elems = is_struct_meta(x) ? x->u_st.metas : x->u_tup.metas;
    y_elems = is_struct_meta(y) ? y->u_st.metas : y->u_tup.metas;

    if (array_size(x_elems) != array_size(y_elems))
        RETURN(ERROR_MISMATCHED_ELEM_CNT, y->pos,
               array_size(x_elems), array_size(y_elems));

    for (i = 0; i < array_size(x_elems); i++) {
        meta_t *x_elem = array_item(x_elems, i, meta_t);
        meta_t *y_elem = array_item(y_elems, i, meta_t);

        CHECK(meta_cmp(x_elem, y_elem));
    }

    return NO_ERROR;
}

int
meta_cmp(meta_t *x, meta_t *y)
{
    if (is_untyped_meta(x) || is_untyped_meta(y)) {
        if (x->type == y->type ||
            (is_integer_meta(x) && is_integer_meta(y)) ||
            (is_float_meta(x) && is_float_meta(y)))
            return NO_ERROR;

        RETURN(ERROR_MISMATCHED_TYPE, y->pos, META_NAME(x), META_NAME(y));
    }

    if (is_map_meta(x) || is_map_meta(y)) {
        if (is_ref_meta(x) || is_ref_meta(y))
            return NO_ERROR;

        return meta_cmp_map(x, y);
    }

    if (is_struct_meta(x) || is_struct_meta(y))
        return meta_cmp_struct(x, y);

    if (x->type != y->type)
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, META_NAME(x), META_NAME(y));

    return NO_ERROR;
}

int
meta_cmp_array(meta_t *x, array_t *y)
{
    int i;

    for (i = 0; i < array_size(y); i++) {
        CHECK(meta_cmp(x, array_item(y, i, meta_t)));
    }

    return NO_ERROR;
}

void
meta_dump(meta_t *meta, int indent)
{
}

/* end of meta.c */
