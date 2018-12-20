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
meta_set_struct(meta_t *meta, char *name, array_t *ids)
{
    int i;

    meta_set(meta, TYPE_STRUCT);

    meta->name = name;

    meta->elem_cnt = array_size(ids);
    meta->elems = xmalloc(sizeof(meta_t *) * meta->elem_cnt);

    meta->size = 0;

    for (i = 0; i < meta->elem_cnt; i++) {
        meta_t *elem_meta = &array_get(ids, i, ast_id_t)->meta;

        ASSERT(elem_meta->size > 0);

        meta->elems[i] = elem_meta;

        elem_meta->offset = meta->size;
        meta->size += ALIGN64(elem_meta->size);
    }
}

void
meta_set_tuple(meta_t *meta, array_t *exps)
{
    int i;

    meta_set(meta, TYPE_TUPLE);

    meta->elem_cnt = array_size(exps);
    meta->elems = xmalloc(sizeof(meta_t *) * meta->elem_cnt);

    meta->size = 0;

    for (i = 0; i < meta->elem_cnt; i++) {
        meta_t *elem_meta = &array_get(exps, i, ast_exp_t)->meta;

        ASSERT(elem_meta->size > 0);

        meta->elems[i] = elem_meta;

        elem_meta->offset = meta->size;
        meta->size += ALIGN64(elem_meta->size);
    }
}

static int
meta_cmp_tuple(meta_t *x, meta_t *y, char *kind)
{
    int i;

    if (!is_tuple_type(y))
        RETURN(ERROR_MISMATCHED_TYPE, &y->pos, meta_to_str(x), meta_to_str(y));

    if (x->elem_cnt != y->elem_cnt) {
        int y_elem_cnt = 0;

        if (x->elem_cnt < y->elem_cnt)
            RETURN(ERROR_MISMATCHED_COUNT, &y->pos, kind, x->elem_cnt, y->elem_cnt);

        for (i = 0; i < y->elem_cnt; i++) {
            if (is_tuple_type(y->elems[i]))
                y_elem_cnt += y->elems[i]->elem_cnt;
            else
                y_elem_cnt++;
        }

        if (x->elem_cnt != y_elem_cnt)
            RETURN(ERROR_MISMATCHED_COUNT, &y->pos, kind, x->elem_cnt, y_elem_cnt);
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
        int i;

        /* y is a tuple of key-value pairs */
        for (i = 0; i < y->elem_cnt; i++) {
            CHECK(meta_cmp_tuple(x, y->elems[i], "key-value"));
        }
    }
    else {
        RETURN(ERROR_MISMATCHED_TYPE, &y->pos, meta_to_str(x), meta_to_str(y));
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
            RETURN(ERROR_MISMATCHED_TYPE, &y->pos, meta_to_str(x), meta_to_str(y));
    }
    else if (is_tuple_type(y)) {
        return meta_cmp_tuple(x, y, "field");
    }
    else {
        RETURN(ERROR_MISMATCHED_TYPE, &y->pos, meta_to_str(x), meta_to_str(y));
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
            (is_pointer_type(x) && is_pointer_type(y))) {
            if (is_undef_type(x)) {
                *x = *y;
                meta_set_undef(x);
            }
            else {
                *y = *x;
                meta_set_undef(y);
            }
            return NO_ERROR;
        }

        RETURN(ERROR_MISMATCHED_TYPE, &y->pos, meta_to_str(x), meta_to_str(y));
    }

    if (is_map_type(x))
        return meta_cmp_map(x, y);

    if (is_tuple_type(x))
        return meta_cmp_tuple(x, y, "element");

    if (is_struct_type(x))
        return meta_cmp_struct(x, y);

    if (x->type != y->type)
        RETURN(ERROR_MISMATCHED_TYPE, &y->pos, meta_to_str(x), meta_to_str(y));

    return NO_ERROR;
}

static int
meta_cmp_array(meta_t *x, int idx, meta_t *y)
{
    int i;

    if (is_array_type(y)) {
        CHECK(meta_cmp_type(x, y));

        if (x->arr_dim != y->arr_dim)
            RETURN(ERROR_MISMATCHED_TYPE, &y->pos, meta_to_str(x), meta_to_str(y));

        for (i = 0; i < x->arr_dim; i++) {
            if (x->arr_size[i] != y->arr_size[i])
                RETURN(ERROR_MISMATCHED_COUNT, &y->pos, "element", x->arr_size[i],
                       y->arr_size[i]);
        }
    }
    else if (is_tuple_type(y)) {
        if (x->arr_size[idx] == -1)
            x->arr_size[idx] = y->elem_cnt;
        else if (x->arr_size[idx] != y->elem_cnt)
            RETURN(ERROR_MISMATCHED_COUNT, &y->pos, "element", x->arr_size[idx],
                   y->elem_cnt);

        for (i = 0; i < y->elem_cnt; i++) {
            if (idx < x->arr_dim - 1)
                CHECK(meta_cmp_array(x, idx + 1, y->elems[i]));
            else
                CHECK(meta_cmp_type(x, y->elems[i]));
        }
    }
    else {
        RETURN(ERROR_MISMATCHED_TYPE, &y->pos, meta_to_str(x), meta_to_str(y));
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

void
meta_dump(meta_t *meta, int indent)
{
}

/* end of meta.c */
