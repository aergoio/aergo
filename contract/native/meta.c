/**
 * @file    meta.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "strbuf.h"
#include "ast_id.h"
#include "ast_exp.h"
#include "ast_blk.h"

#include "meta.h"

char *
meta_to_str(meta_t *meta)
{
    strbuf_t buf;

    strbuf_init(&buf);

    if (is_struct_meta(meta)) {
        ASSERT(meta->type_id != NULL);

        strbuf_cat(&buf, "struct ");
        strbuf_cat(&buf, meta->type_id->name);
    }
    else if (is_array_meta(meta)) {
        int i;

        ASSERT1(meta->elem_cnt == 1, meta->elem_cnt);

        strbuf_cat(&buf, meta_to_str(meta->elems[0]));
        for (i = 0; i < meta->arr_dim; i++) {
            strbuf_cat(&buf, "[]");
        }
    }
    else if (is_map_meta(meta)) {
        ASSERT1(meta->elem_cnt == 2, meta->elem_cnt);

        strbuf_cat(&buf, "map(");
        strbuf_cat(&buf, meta_to_str(meta->elems[0]));
        strbuf_cat(&buf, ", ");
        strbuf_cat(&buf, meta_to_str(meta->elems[1]));
        strbuf_cat(&buf, ")");
    }
    else if (is_tuple_meta(meta)) {
        strbuf_cat(&buf, "tuple");
    }
    else if (is_object_meta(meta) && meta->type_id != NULL) {
        strbuf_cat(&buf, meta->type_id->name);
    }
    else if (is_undef_meta(meta) && is_numeric_meta(meta)) {
        strbuf_cat(&buf, "number");
    }
    else {
        strbuf_cat(&buf, TYPE_NAME(meta->type));
    }

    return strbuf_str(&buf);
}

void
meta_set_array(meta_t *meta, meta_t *elem_meta, int arr_dim)
{
    int i;

    ASSERT1(arr_dim > 0, arr_dim);

    meta_set(meta, TYPE_ARRAY);

    meta->is_fixed = true;
    meta->max_dim = arr_dim;
    meta->arr_dim = arr_dim;

    meta->dim_sizes = xmalloc(sizeof(int) * arr_dim);
    for (i = 0; i < arr_dim; i++) {
        meta->dim_sizes[i] = -1;
    }

    meta->elem_cnt = 1;
    meta->elems = xmalloc(sizeof(meta_t *) * meta->elem_cnt);
    meta->elems[0] = elem_meta;

    meta->align = TYPE_ALIGN(elem_meta->type);
}

void
meta_set_map(meta_t *meta, meta_t *key_meta, meta_t *val_meta)
{
    meta_set(meta, TYPE_MAP);

    meta->elem_cnt = 2;
    meta->elems = xmalloc(sizeof(meta_t *) * meta->elem_cnt);

    meta->elems[0] = key_meta;
    meta->elems[1] = val_meta;
}

void
meta_set_struct(meta_t *meta, ast_id_t *id)
{
    int i;

    meta_set(meta, TYPE_STRUCT);

    meta->elem_cnt = vector_size(id->u_struc.fld_ids);
    meta->elems = xmalloc(sizeof(meta_t *) * meta->elem_cnt);

    for (i = 0; i < meta->elem_cnt; i++) {
        meta->elems[i] = &vector_get_id(id->u_struc.fld_ids, i)->meta;
    }

    meta->align = meta_align(meta->elems[0]);
    meta->type_id = id;
}

void
meta_set_tuple(meta_t *meta, vector_t *elem_exps)
{
    int i;

    meta_set(meta, TYPE_TUPLE);

    meta->elem_cnt = vector_size(elem_exps);
    meta->elems = xmalloc(sizeof(meta_t *) * meta->elem_cnt);

    for (i = 0; i < meta->elem_cnt; i++) {
        meta->elems[i] = &vector_get_exp(elem_exps, i)->meta;
    }

    meta->align = meta_align(meta->elems[0]);
}

void
meta_set_object(meta_t *meta, ast_id_t *id)
{
    meta_set(meta, TYPE_OBJECT);

    if (id != NULL)
        ASSERT1(is_cont_id(id) || is_itf_id(id), id->kind);

    meta->type_id = id;
}

static bool
meta_cmp_map(meta_t *x, meta_t *y)
{
    ASSERT1(x->elem_cnt == 2, x->elem_cnt);

    if (is_object_meta(y))
        /* TODO: null value check */
        return true;

    if (is_map_meta(y)) {
        CHECK(meta_cmp(x->elems[0], y->elems[0]));
        CHECK(meta_cmp(x->elems[1], y->elems[1]));
    }
    else if (is_tuple_meta(y)) {
        int i, j;

        /* y is a tuple of key-value pairs */
        for (i = 0; i < y->elem_cnt; i++) {
            meta_t *y_elem = y->elems[i];

            if (!is_tuple_meta(y_elem))
                RETURN(ERROR_MISMATCHED_TYPE, y_elem->pos, meta_to_str(x), meta_to_str(y_elem));

            if (x->elem_cnt != y_elem->elem_cnt)
                RETURN(ERROR_MISMATCHED_COUNT, y_elem->pos, "key-value", x->elem_cnt,
                       y_elem->elem_cnt);

            for (j = 0; j < x->elem_cnt; j++) {
                CHECK(meta_cmp(x->elems[j], y_elem->elems[j]));
            }
        }
    }
    else {
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));
    }

    return true;
}

static bool
meta_cmp_struct(meta_t *x, meta_t *y)
{
    if (is_struct_meta(y)) {
        ASSERT(x->type_id != NULL);
        ASSERT(y->type_id != NULL);

        if (strcmp(x->type_id->name, y->type_id->name) != 0)
            RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));
    }
    else if (is_tuple_meta(y)) {
        int i;

        if (x->elem_cnt != y->elem_cnt)
            RETURN(ERROR_MISMATCHED_COUNT, y->pos, "field", x->elem_cnt, y->elem_cnt);

        for (i = 0; i < x->elem_cnt; i++) {
            CHECK(meta_cmp(x->elems[i], y->elems[i]));
        }
    }
    else {
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));
    }

    return true;
}

static bool
meta_cmp_tuple(meta_t *x, meta_t *y)
{
    int i;

    if (!is_tuple_meta(y))
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));

    if (x->elem_cnt != y->elem_cnt) {
        int y_elem_cnt = 0;

        if (x->elem_cnt < y->elem_cnt)
            RETURN(ERROR_MISMATCHED_COUNT, y->pos, "element", x->elem_cnt, y->elem_cnt);

        for (i = 0; i < y->elem_cnt; i++) {
            if (is_tuple_meta(y->elems[i]))
                y_elem_cnt += y->elems[i]->elem_cnt;
            else
                y_elem_cnt++;
        }

        if (x->elem_cnt != y_elem_cnt)
            RETURN(ERROR_MISMATCHED_COUNT, y->pos, "element", x->elem_cnt, y_elem_cnt);
    }

    if (x->elem_cnt == y->elem_cnt) {
        for (i = 0; i < x->elem_cnt; i++) {
            CHECK(meta_cmp(x->elems[i], y->elems[i]));
        }
    }
    else {
        int j;
        int x_idx = 0;

        for (i = 0; i < y->elem_cnt; i++) {
            meta_t *y_elem = y->elems[i];

            if (is_tuple_meta(y_elem)) {
                for (j = 0; j < y_elem->elem_cnt; j++) {
                    CHECK(meta_cmp(x->elems[x_idx + j], y_elem->elems[j]));
                }
                x_idx += y_elem->elem_cnt;
            }
            else {
                CHECK(meta_cmp(x->elems[x_idx++], y_elem));
            }
        }
    }

    return true;
}

static bool
meta_cmp_object(meta_t *x, meta_t *y)
{
    if (!is_object_meta(y))
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));

    ASSERT(x->type_id != NULL);
    ASSERT(y->type_id != NULL);
    ASSERT1(is_cont_id(x->type_id) || is_itf_id(x->type_id), x->type_id->kind);
    ASSERT1(is_cont_id(y->type_id) || is_itf_id(y->type_id), y->type_id->kind);

    if (is_cont_id(x->type_id)) {
        if (is_itf_id(y->type_id))
            RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));

        if (x->type_id != y->type_id)
            RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));
    }

    return true;
}

static bool
meta_cmp_type(meta_t *x, meta_t *y)
{
    if (is_undef_meta(x) || is_undef_meta(y)) {
        if (x->type == y->type ||
            (is_integer_meta(x) && is_integer_meta(y)) ||
            (is_nullable_meta(x) && is_nullable_meta(y)))
            return true;

        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));
    }

    if (is_map_meta(x))
        return meta_cmp_map(x, y);

    if (is_tuple_meta(x))
        return meta_cmp_tuple(x, y);

    if (is_struct_meta(x))
        return meta_cmp_struct(x, y);

    if (is_object_meta(x))
        return meta_cmp_object(x, y);

    if (x->type != y->type)
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, TYPE_NAME(x->type), TYPE_NAME(y->type));

    return true;
}

static bool
meta_cmp_array(meta_t *x, int dim, meta_t *y)
{
    int i;

    ASSERT1(x->elem_cnt == 1, x->elem_cnt);

    if (is_array_meta(y)) {
        ASSERT1(y->elem_cnt == 1, y->elem_cnt);

        if (x->arr_dim != y->arr_dim)
            RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));

        if (is_fixed_meta(x)) {
            for (i = 0; i < x->arr_dim; i++) {
                if (x->dim_sizes[i] != y->dim_sizes[i])
                    RETURN(ERROR_MISMATCHED_COUNT, y->pos, "element", x->dim_sizes[i],
                           y->dim_sizes[i]);
            }
        }

        CHECK(meta_cmp_type(x->elems[0], y->elems[0]));
    }
    else if (is_tuple_meta(y)) {
        if (is_fixed_meta(x) && x->dim_sizes[dim] != y->elem_cnt)
            RETURN(ERROR_MISMATCHED_COUNT, y->pos, "element", x->dim_sizes[dim], y->elem_cnt);

        for (i = 0; i < y->elem_cnt; i++) {
            if (dim < x->arr_dim - 1)
                CHECK(meta_cmp_array(x, dim + 1, y->elems[i]));
            else
                CHECK(meta_cmp_type(x->elems[0], y->elems[i]));
        }
    }
    else if (!is_undef_meta(y) || !is_object_meta(y)) {
        /* not allowed except "null" */
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));
    }

    return true;
}

bool
meta_cmp(meta_t *x, meta_t *y)
{
    if (is_array_meta(x))
        return meta_cmp_array(x, 0, y);

    return meta_cmp_type(x, y);
}

static void
meta_eval_type(meta_t *x, meta_t *y)
{
    int i, j;

    if (is_undef_meta(x)) {
        meta_set(x, y->type);
        x->is_undef = y->is_undef;
        return;
    }
    else if (is_undef_meta(y)) {
        meta_set(y, x->type);
        y->is_undef = x->is_undef;
        return;
    }

    if (!is_tuple_meta(y))
        return;

    ASSERT1(y->elem_cnt > 0, y->elem_cnt);

    if (is_map_meta(x)) {
        ASSERT1(x->elem_cnt == 2, x->elem_cnt);

        for (i = 0; i < y->elem_cnt; i++) {
            ASSERT1(y->elems[i]->elem_cnt == 2, y->elems[i]->elem_cnt);

            for (j = 0; j < x->elem_cnt; j++) {
                meta_eval(x->elems[j], y->elems[i]->elems[j]);
            }
        }

        meta_copy(y, x);
    }
    else if (is_struct_meta(x)) {
        ASSERT2(x->elem_cnt == y->elem_cnt, x->elem_cnt, y->elem_cnt);

        for (i = 0; i < y->elem_cnt; i++) {
            meta_eval(x->elems[i], y->elems[i]);
        }

        meta_set(y, TYPE_STRUCT);
    }
    else {
        ASSERT1(is_tuple_meta(x), x->type);
        ASSERT2(x->elem_cnt == y->elem_cnt, x->elem_cnt, y->elem_cnt);

        for (i = 0; i < x->elem_cnt; i++) {
            meta_eval(x->elems[i], y->elems[i]);
        }
    }

    y->align = y->elems[0]->align;
}

static void
meta_eval_array(meta_t *x, int dim, meta_t *y)
{
    int i;

    ASSERT1(x->elem_cnt == 1, x->elem_cnt);
    ASSERT2(x->max_dim == x->arr_dim, x->max_dim, x->arr_dim);
    ASSERT2(dim <= x->max_dim, dim, x->max_dim);

    if (is_tuple_meta(y)) {
        if (dim == x->max_dim) {
            meta_eval_type(x->elems[0], y);
        }
        else {
            ASSERT2(x->dim_sizes[dim] == -1 || x->dim_sizes[dim] == y->elem_cnt,
                    x->dim_sizes[dim], y->elem_cnt);

            for (i = 0; i < y->elem_cnt; i++) {
                meta_eval_array(x, dim + 1, y->elems[i]);
            }

            y->type = TYPE_ARRAY;
            y->is_fixed = true;
            y->max_dim = x->max_dim;
            y->arr_dim = x->max_dim - dim;
            y->dim_sizes = &x->dim_sizes[dim];
            y->dim_sizes[0] = y->elem_cnt;
            y->align = y->elems[0]->align;
            y->elem_cnt = 1;
            y->elems[0] = x->elems[0];
        }
    }
    else {
        meta_eval_type(x->elems[0], y);
    }
}

void
meta_eval(meta_t *x, meta_t *y)
{
    if (is_array_meta(x))
        meta_eval_array(x, 0, y);
    else
        meta_eval_type(x, y);
}

/* end of meta.c */
