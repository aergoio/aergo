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
meta_to_str(meta_t *meta)
{
    int i;
    strbuf_t buf;

    strbuf_init(&buf);

    if (is_struct_type(meta)) {
        ASSERT(meta->name != NULL);

        strbuf_cat(&buf, "struct ");
        strbuf_cat(&buf, meta->name);
    }
    else if (is_map_type(meta)) {
        ASSERT1(meta->elem_cnt == 2, meta->elem_cnt);

        strbuf_cat(&buf, "map(");
        strbuf_cat(&buf, meta_to_str(meta->elems[0]));
        strbuf_cat(&buf, ", ");
        strbuf_cat(&buf, meta_to_str(meta->elems[1]));
        strbuf_cat(&buf, ")");
    }
    else if (is_tuple_type(meta)) {
        int i;

        strbuf_cat(&buf, "{");
        for (i = 0; i < meta->elem_cnt; i++) {
            if (i > 0)
                strbuf_cat(&buf, ", ");

            strbuf_cat(&buf, meta_to_str(meta->elems[i]));
        }
        strbuf_cat(&buf, "}");
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

    meta->name = id->name;

    meta->elem_cnt = array_size(id->u_struc.fld_ids);
    meta->elems = xmalloc(sizeof(meta_t *) * meta->elem_cnt);

    meta->size = 0;

    for (i = 0; i < meta->elem_cnt; i++) {
        meta_t *elem_meta = &array_get_id(id->u_struc.fld_ids, i)->meta;

        ASSERT(elem_meta->size > 0);

        meta->elems[i] = elem_meta;
        meta->size = ALIGN(meta->size, meta_align(elem_meta));

        if (is_array_type(elem_meta)) {
            ASSERT2(elem_meta->arr_size >= meta->size, elem_meta->arr_size, meta->size);
            meta->size += elem_meta->arr_size;
        }
        else {
            meta->size += elem_meta->size;
        }
    }

    meta->size = ALIGN64(meta->size);

    meta->type_id = id;
}

void
meta_set_tuple(meta_t *meta, array_t *elem_exps)
{
    int i;

    meta_set(meta, TYPE_TUPLE);

    meta->elem_cnt = array_size(elem_exps);
    meta->elems = xmalloc(sizeof(meta_t *) * meta->elem_cnt);

    meta->size = 0;

    for (i = 0; i < meta->elem_cnt; i++) {
        meta_t *elem_meta = &array_get_exp(elem_exps, i)->meta;

        ASSERT(elem_meta->size > 0);

        meta->elems[i] = elem_meta;
        meta->size = ALIGN(meta->size, meta_align(elem_meta));

        if (is_array_type(elem_meta)) {
            ASSERT2(elem_meta->arr_size >= meta->size, elem_meta->arr_size, meta->size);
            meta->size += elem_meta->arr_size;
        }
        else {
            meta->size += elem_meta->size;
        }
    }

    meta->size = ALIGN64(meta->size);
}

void
meta_set_object(meta_t *meta, ast_id_t *id)
{
    meta_set(meta, TYPE_OBJECT);

    if (id != NULL) {
        ASSERT1(is_cont_id(id) || is_itf_id(id), id->kind);
        meta->name = id->name;
    }

    meta->type_id = id;
}

static int
meta_cmp_map(meta_t *x, meta_t *y)
{
    ASSERT1(x->elem_cnt == 2, x->elem_cnt);

    if (is_object_type(y))
        /* TODO: null value check */
        return NO_ERROR;

    if (is_map_type(y)) {
        CHECK(meta_cmp(x->elems[0], y->elems[0]));
        CHECK(meta_cmp(x->elems[1], y->elems[1]));
    }
    else if (is_tuple_type(y)) {
        int i, j;

        /* y is a tuple of key-value pairs */
        for (i = 0; i < y->elem_cnt; i++) {
            meta_t *y_elem = y->elems[i];

            if (!is_tuple_type(y_elem))
                RETURN(ERROR_MISMATCHED_TYPE, y_elem->pos, meta_to_str(x),
                       meta_to_str(y_elem));

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

    return NO_ERROR;
}

static int
meta_cmp_struct(meta_t *x, meta_t *y)
{
    if (is_struct_type(y)) {
        ASSERT(x->name != NULL);
        ASSERT(y->name != NULL);

        if (strcmp(x->name, y->name) != 0)
            RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));
    }
    else if (is_tuple_type(y)) {
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

    return NO_ERROR;
}

static int
meta_cmp_tuple(meta_t *x, meta_t *y)
{
    int i;

    if (!is_tuple_type(y))
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));

    if (x->elem_cnt != y->elem_cnt) {
        int y_elem_cnt = 0;

        if (x->elem_cnt < y->elem_cnt)
            RETURN(ERROR_MISMATCHED_COUNT, y->pos, "element", x->elem_cnt, y->elem_cnt);

        for (i = 0; i < y->elem_cnt; i++) {
            if (is_tuple_type(y->elems[i]))
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

            if (is_tuple_type(y_elem)) {
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

    return NO_ERROR;
}

static int
meta_cmp_type(meta_t *x, meta_t *y)
{
    if (is_undef_type(x) || is_undef_type(y)) {
        if (x->type == y->type ||
            (is_integer_type(x) && is_integer_type(y)) ||
            (is_fpoint_type(x) && is_fpoint_type(y)) ||
            (is_pointer_type(x) && is_pointer_type(y)))
            return NO_ERROR;

        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));
    }

    if (is_map_type(x))
        return meta_cmp_map(x, y);

    if (is_tuple_type(x))
        return meta_cmp_tuple(x, y);

    if (is_struct_type(x))
        return meta_cmp_struct(x, y);

    if (x->type != y->type)
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));

    return NO_ERROR;
}

static int
meta_cmp_array(meta_t *x, int dim, meta_t *y)
{
    int i;

    if (is_array_type(y)) {
        CHECK(meta_cmp_type(x, y));

        if (x->arr_dim != y->arr_dim)
            RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));

        for (i = 0; i < x->arr_dim; i++) {
            if (x->dim_sizes[i] != y->dim_sizes[i])
                RETURN(ERROR_MISMATCHED_COUNT, y->pos, "element", x->dim_sizes[i],
                       y->dim_sizes[i]);
        }
    }
    else if (is_tuple_type(y)) {
        if (x->dim_sizes[dim] == -1)
            meta_set_dim_size(x, dim, y->elem_cnt);
        else if (x->dim_sizes[dim] != y->elem_cnt)
            RETURN(ERROR_MISMATCHED_COUNT, y->pos, "element", x->dim_sizes[dim],
                   y->elem_cnt);

        for (i = 0; i < y->elem_cnt; i++) {
            if (dim < x->arr_dim - 1)
                CHECK(meta_cmp_array(x, dim + 1, y->elems[i]));
            else
                CHECK(meta_cmp_type(x, y->elems[i]));
        }
    }
    else {
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, meta_to_str(x), meta_to_str(y));
    }

    return NO_ERROR;
}

int
meta_cmp(meta_t *x, meta_t *y)
{
    if (is_array_type(x))
        return meta_cmp_array(x, 0, y);

    return meta_cmp_type(x, y);
}

static void
meta_eval_type(meta_t *x, meta_t *y)
{
    int i, j;

    if (is_undef_type(x)) {
        meta_copy(x, y);
    }
    else if (is_undef_type(y)) {
        meta_copy(y, x);
    }
    else if (is_map_type(x)) {
        if (!is_tuple_type(y))
            return;

        for (i = 0; i < y->elem_cnt; i++) {
            for (j = 0; j < x->elem_cnt; j++) {
                meta_eval(x->elems[j], y->elems[i]->elems[j]);
            }
        }
    }
    else if (is_tuple_type(x)) {
        ASSERT1(is_tuple_type(y), y->type);

        if (x->elem_cnt == y->elem_cnt) {
            for (i = 0; i < x->elem_cnt; i++) {
                meta_eval(x->elems[i], y->elems[i]);
            }
        }
    }
    else if (is_struct_type(x)) {
        if (!is_tuple_type(y))
            return;

        for (i = 0; i < y->elem_cnt; i++) {
            meta_eval(x->elems[i], y->elems[i]);
        }
    }
}

static void
meta_eval_array(meta_t *x, int dim, meta_t *y)
{
    int i;

    if (!is_tuple_type(y))
        return;

    for (i = 0; i < y->elem_cnt; i++) {
        if (dim < x->arr_dim - 1)
            meta_eval_array(x, dim + 1, y->elems[i]);
        else
            meta_eval_type(x, y->elems[i]);
    }
}

void
meta_eval(meta_t *x, meta_t *y)
{
    if (is_array_type(x))
        meta_eval_array(x, 0, y);
    else
        meta_eval_type(x, y);
}

void
meta_dump(meta_t *meta, int indent)
{
}

/* end of meta.c */
