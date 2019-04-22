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
    int i;
    strbuf_t buf;

    strbuf_init(&buf);

    if (is_struct_meta(meta)) {
        ASSERT(meta->type_id != NULL);

        strbuf_cat(&buf, "struct ");
        strbuf_cat(&buf, meta->type_id->name);
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
        int i;

        strbuf_cat(&buf, "{");
        for (i = 0; i < meta->elem_cnt; i++) {
            if (i > 0)
                strbuf_cat(&buf, ", ");

            strbuf_cat(&buf, meta_to_str(meta->elems[i]));
        }
        strbuf_cat(&buf, "}");
    }
    else if (is_object_meta(meta) && meta->type_id != NULL) {
        strbuf_cat(&buf, meta->type_id->name);
    }
    else {
        strbuf_cat(&buf, TYPE_NAME(meta->type));
    }

    for (i = 0; i < meta->arr_dim; i++) {
        strbuf_cat(&buf, "[]");
    }

    return strbuf_str(&buf);
}

void
meta_set_map(meta_t *meta, meta_t *k, meta_t *v)
{
    meta_set(meta, TYPE_MAP);

    meta->elem_cnt = 2;
    meta->elems = xmalloc(sizeof(meta_t *) * meta->elem_cnt);

    meta->elems[0] = k;
    meta->elems[1] = v;
}

void
meta_set_struct(meta_t *meta, ast_id_t *id)
{
    int i;

    meta_set(meta, TYPE_STRUCT);

    meta->size = 0;

    meta->elem_cnt = vector_size(id->u_struc.fld_ids);
    meta->elems = xmalloc(sizeof(meta_t *) * meta->elem_cnt);

    for (i = 0; i < meta->elem_cnt; i++) {
        meta->elems[i] = &vector_get_id(id->u_struc.fld_ids, i)->meta;

        ASSERT(meta_align(meta->elems[i]) > 0);
        ASSERT(meta_bytes(meta->elems[i]) > 0);

        meta->size = ALIGN(meta->size, meta_align(meta->elems[i]));
        meta->size += meta_bytes(meta->elems[i]);
    }

    meta->align = meta_align(meta->elems[0]);
    meta->size = ALIGN(meta->size, meta->align);
    meta->type_id = id;
}

void
meta_set_tuple(meta_t *meta, vector_t *elem_exps)
{
    int i;

    meta_set(meta, TYPE_TUPLE);

    meta->size = 0;

    meta->elem_cnt = vector_size(elem_exps);
    meta->elems = xmalloc(sizeof(meta_t *) * meta->elem_cnt);

    for (i = 0; i < meta->elem_cnt; i++) {
        meta->elems[i] = &vector_get_exp(elem_exps, i)->meta;

        ASSERT(meta_align(meta->elems[i]) > 0);
        ASSERT(meta_bytes(meta->elems[i]) > 0);

        meta->size = ALIGN(meta->size, meta_align(meta->elems[i]));
        meta->size += meta_bytes(meta->elems[i]);
    }

    meta->align = meta_align(meta->elems[0]);
    meta->size = ALIGN(meta->size, meta->align);
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
            (is_fpoint_meta(x) && is_fpoint_meta(y)) ||
            (is_pointer_meta(x) && is_pointer_meta(y)))
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
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));

    return true;
}

static bool
meta_cmp_array(meta_t *x, int dim, meta_t *y)
{
    int i;

    if (is_array_meta(y)) {
        CHECK(meta_cmp_type(x, y));

        if (x->arr_dim != y->arr_dim)
            RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));

        for (i = 0; i < x->arr_dim; i++) {
            if (x->dim_sizes[i] == -1)
                meta_set_dim_size(x, i, y->dim_sizes[i]);
            else if (x->dim_sizes[i] != y->dim_sizes[i])
                RETURN(ERROR_MISMATCHED_COUNT, y->pos, "element", x->dim_sizes[i], y->dim_sizes[i]);
        }
    }
    else if (is_tuple_meta(y)) {
        if (x->dim_sizes[dim] == -1)
            meta_set_dim_size(x, dim, y->elem_cnt);
        else if (x->dim_sizes[dim] != y->elem_cnt)
            RETURN(ERROR_MISMATCHED_COUNT, y->pos, "element", x->dim_sizes[dim], y->elem_cnt);

        for (i = 0; i < y->elem_cnt; i++) {
            if (dim < x->arr_dim - 1)
                CHECK(meta_cmp_array(x, dim + 1, y->elems[i]));
            else
                CHECK(meta_cmp_type(x, y->elems[i]));
        }
    }
    else if (!is_undef_meta(y) || !is_object_meta(y)) {
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

static inline void
meta_set_size(meta_t *meta)
{
    int i;

    ASSERT1(is_tuple_meta(meta) || is_struct_meta(meta), meta->type);

    meta->size = 0;

    for (i = 0; i < meta->elem_cnt; i++) {
        meta_t *elem_meta = meta->elems[i];

        meta->size = ALIGN(meta->size, meta_align(elem_meta));
        meta->size += meta_size(elem_meta);
    }

    meta->align = meta_align(meta->elems[0]);
    meta->size = ALIGN(meta->size, meta->align);
}

static void
meta_eval_type(meta_t *x, meta_t *y)
{
    int i, j;

    if (is_undef_meta(x)) {
        x->type = y->type;
        x->size = TYPE_SIZE(x->type);
        x->align = TYPE_ALIGN(x->type);
        x->is_undef = y->is_undef;
        return;
    }
    else if (is_undef_meta(y)) {
        y->type = x->type;
        y->size = TYPE_SIZE(y->type);
        y->align = TYPE_ALIGN(y->type);
        y->is_undef = x->is_undef;
        return;
    }

    if (!is_tuple_meta(y))
        return;

    if (is_map_meta(x)) {
        for (i = 0; i < y->elem_cnt; i++) {
            for (j = 0; j < x->elem_cnt; j++) {
                meta_eval(x->elems[j], y->elems[i]->elems[j]);
            }

            meta_set_size(y->elems[i]);
        }
    }
    else if (is_tuple_meta(x)) {
        ASSERT2(x->elem_cnt == y->elem_cnt, x->elem_cnt, y->elem_cnt);

        for (i = 0; i < x->elem_cnt; i++) {
            meta_eval(x->elems[i], y->elems[i]);
        }
    }
    else {
        ASSERT1(is_struct_meta(x), x->type);
        ASSERT2(x->elem_cnt == y->elem_cnt, x->elem_cnt, y->elem_cnt);

        y->type = x->type;
        y->type_id = x->type_id;

        for (i = 0; i < y->elem_cnt; i++) {
            meta_eval(x->elems[i], y->elems[i]);
        }
    }

    meta_set_size(y);
}

static void
meta_eval_array(meta_t *x, int dim, meta_t *y)
{
    int i;

    ASSERT2(dim <= x->arr_dim, dim, x->arr_dim);

    if (dim == x->arr_dim || !is_tuple_meta(y)) {
        meta_eval_type(x, y);
    }
    else {
        for (i = 0; i < y->elem_cnt; i++) {
            meta_eval_array(x, dim + 1, y->elems[i]);
        }

        meta_set_size(y);

        y->max_dim = x->max_dim;
        y->arr_dim = x->arr_dim - dim;
        y->dim_sizes = &x->dim_sizes[dim];
        y->size += sizeof(uint64_t);
    }
}

void
meta_eval(meta_t *x, meta_t *y)
{
    if (is_array_meta(x))
        meta_eval_array(x, 0, y);
    else
        meta_eval_type(x, y);

    /* TODO: map vs tuple
    if (is_map_meta(x) || !is_tuple_meta(y))
        return;

    ASSERT2(meta_bytes(x) == y->size, meta_bytes(x), y->size);
    */
}

/* end of meta.c */
