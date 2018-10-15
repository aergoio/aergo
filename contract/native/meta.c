/**
 * @file    meta.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "strbuf.h"
#include "ast_id.h"
#include "ast_exp.h"

#include "meta.h"

char *
meta_to_str(meta_t *x)
{
    strbuf_t buf;

    if (is_struct_meta(x)) {
        ASSERT(x->u_tup.name != NULL);

        strbuf_init(&buf);
        strbuf_cat(&buf, "struct ");
        strbuf_cat(&buf, x->u_tup.name);

        return strbuf_str(&buf);
    }
    else if (is_map_meta(x)) {
        strbuf_init(&buf);
        strbuf_cat(&buf, "map(");
        strbuf_cat(&buf, meta_to_str(x->u_map.k_meta));
        strbuf_cat(&buf, ", ");
        strbuf_cat(&buf, meta_to_str(x->u_map.v_meta));
        strbuf_cat(&buf, ")");

        return strbuf_str(&buf);
    }
    else if (is_tuple_meta(x)) {
        int i;
        array_t *elems = x->u_tup.metas;

        strbuf_init(&buf);
        strbuf_cat(&buf, "{");
        for (i = 0; i < array_size(elems); i++) {
            if (i > 0)
                strbuf_cat(&buf, ", ");

            strbuf_cat(&buf, meta_to_str(array_item(elems, i, meta_t)));
        }
        strbuf_cat(&buf, "}");

        return strbuf_str(&buf);
    }

    return TYPE_NAME(x->type);
}

void
meta_set_struct(meta_t *meta, char *name, array_t *ids)
{
    int i;
    array_t *metas = array_new();

    meta_set(meta, TYPE_STRUCT);

    for (i = 0; i < array_size(ids); i++) {
        array_add_last(metas, &array_item(ids, i, ast_id_t)->meta);
    }

    meta->u_tup.name = name;
    meta->u_tup.metas = metas;
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

    meta->u_tup.name = NULL;
    meta->u_tup.metas = metas;
}

bool
meta_equals(meta_t *x, meta_t *y)
{
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

    if (is_struct_meta(x) || is_struct_meta(y) ||
        is_tuple_meta(x) || is_tuple_meta(y)) {
        int i;
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
meta_check_map(meta_t *x, meta_t *y)
{
    meta_t *k_meta = x->u_map.k_meta;
    meta_t *v_meta = x->u_map.v_meta;

    if (is_ref_meta(y))
        return NO_ERROR;

    if (is_map_meta(y)) {
        CHECK(meta_check(k_meta, y->u_map.k_meta));
        CHECK(meta_check(v_meta, y->u_map.v_meta));
    }
    else if (is_tuple_meta(y)) {
        int i;
        array_t *kv_elems = y->u_tup.metas;

        for (i = 0; i < array_size(kv_elems); i++) {
            meta_t *kv_elem = array_item(kv_elems, i, meta_t);
            array_t *kv_metas = kv_elem->u_tup.metas;

            if (!is_tuple_meta(kv_elem))
                RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), 
                       meta_to_str(kv_elem));

            if (array_size(kv_metas) != 2)
                RETURN(ERROR_MISMATCHED_ELEM_CNT, y->pos, 2, array_size(kv_metas));

            CHECK(meta_check(k_meta, array_item(kv_metas, 0, meta_t)));
            CHECK(meta_check(v_meta, array_item(kv_metas, 1, meta_t)));
        }
    }
    else {
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));
    }

    return NO_ERROR;
}

static int
meta_check_struct(meta_t *x, meta_t *y)
{
    int i;
    array_t *x_elems = x->u_tup.metas;
    array_t *y_elems = y->u_tup.metas;

    if (!is_struct_meta(y) && !is_tuple_meta(y))
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));

    if (array_size(x_elems) != array_size(y_elems))
        RETURN(ERROR_MISMATCHED_ELEM_CNT, y->pos, array_size(x_elems),
               array_size(y_elems));

    for (i = 0; i < array_size(x_elems); i++) {
        meta_t *x_elem = array_item(x_elems, i, meta_t);
        meta_t *y_elem = array_item(y_elems, i, meta_t);

        CHECK(meta_check(x_elem, y_elem));
    }

    return NO_ERROR;
}

static int
meta_check_type(meta_t *x, meta_t *y)
{
    if (is_untyped_meta(y)) {
        ASSERT1(!is_untyped_meta(x), x->type);

        if (x->type == y->type ||
            (is_integer_meta(x) && is_integer_meta(y)) ||
            (is_float_meta(x) && is_float_meta(y)))
            return NO_ERROR;

        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));
    }

    if (is_map_meta(x))
        return meta_check_map(x, y);

    if (is_struct_meta(x))
        return meta_check_struct(x, y);

    if (x->type != y->type)
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));

    return NO_ERROR;
}

static int
meta_check_array(meta_t *x, int idx, meta_t *y)
{
    int i;
    array_t *arr_elems;

    if (!is_tuple_meta(y))
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));

    arr_elems = y->u_tup.metas;

    if (x->arr_size[idx] == -1)
        x->arr_size[idx] = array_size(arr_elems);
    else if (x->arr_size[idx] != array_size(arr_elems))
        RETURN(ERROR_MISMATCHED_ELEM_CNT, y->pos, x->arr_size[idx],
               array_size(arr_elems));

    for (i = 0; i < array_size(arr_elems); i++) {
        meta_t *val_meta = array_item(arr_elems, i, meta_t);

        if (idx < x->arr_dim - 1)
            CHECK(meta_check_array(x, idx + 1, val_meta));
        else
            CHECK(meta_check_type(x, val_meta));
    }

    return NO_ERROR;
}

int
meta_check(meta_t *x, meta_t *y)
{
    if (is_array_meta(x))
        return meta_check_array(x, 0, y);

    return meta_check_type(x, y);
}

void
meta_dump(meta_t *meta, int indent)
{
}

/* end of meta.c */
