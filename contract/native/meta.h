/**
 * @file    meta.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _META_H
#define _META_H

#include "common.h"

#include "array.h"

#define is_bool_meta(meta)          ((meta)->type == TYPE_BOOL)

#define is_integer_meta(meta)                                                            \
    ((meta)->type >= TYPE_INT8 && (meta)->type <= TYPE_UINT64)
#define is_float_meta(meta)                                                              \
    ((meta)->type == TYPE_FLOAT || (meta)->type == TYPE_DOUBLE)
#define is_numeric_meta(meta)                                                            \
    ((meta)->type >= TYPE_INT8 && (meta)->type <= TYPE_DOUBLE)

#define is_string_meta(meta)        ((meta)->type == TYPE_STRING)

#define is_struct_meta(meta)        ((meta)->type == TYPE_STRUCT)
#define is_map_meta(meta)           ((meta)->type == TYPE_MAP)
#define is_object_meta(meta)        ((meta)->type == TYPE_OBJECT)
#define is_void_meta(meta)          ((meta)->type == TYPE_VOID)
#define is_tuple_meta(meta)         ((meta)->type == TYPE_TUPLE)

#define is_untyped_meta(meta)       (meta)->is_untyped
#define is_array_meta(meta)         ((meta)->arr_dim > 0)

#define is_builtin_meta(meta)       ((meta)->type <= TYPE_BUILTIN)
#define is_primitive_meta(meta)     ((meta)->type <= TYPE_PRIMITIVE)
#define is_comparable_meta(meta)    ((meta)->type <= TYPE_COMPARABLE)

#define meta_set_bool(meta)         meta_set((meta), TYPE_BOOL)
#define meta_set_byte(meta)         meta_set((meta), TYPE_BYTE)
#define meta_set_int8(meta)         meta_set((meta), TYPE_INT8)
#define meta_set_uint8(meta)        meta_set((meta), TYPE_UINT8)
#define meta_set_int16(meta)        meta_set((meta), TYPE_INT16)
#define meta_set_uint16(meta)       meta_set((meta), TYPE_UINT16)
#define meta_set_int32(meta)        meta_set((meta), TYPE_INT32)
#define meta_set_uint32(meta)       meta_set((meta), TYPE_UINT32)
#define meta_set_int64(meta)        meta_set((meta), TYPE_INT64)
#define meta_set_uint64(meta)       meta_set((meta), TYPE_UINT64)
#define meta_set_float(meta)        meta_set((meta), TYPE_FLOAT)
#define meta_set_double(meta)       meta_set((meta), TYPE_DOUBLE)
#define meta_set_string(meta)       meta_set((meta), TYPE_STRING)
#define meta_set_object(meta)          meta_set((meta), TYPE_OBJECT)
#define meta_set_account(meta)      meta_set((meta), TYPE_ACCOUNT)
#define meta_set_void(meta)         meta_set((meta), TYPE_VOID)

#define meta_size(meta)                                                                  \
    (is_void_meta(meta) ? 0 :                                                            \
     ((is_tuple_meta(meta) || is_struct_meta(meta)) ?                                    \
      array_size((meta)->u_tup.metas) : 1))

#ifndef _META_T
#define _META_T
typedef struct meta_s meta_t;
#endif /* ! _META_T */

typedef struct meta_tuple_s {
    char *name;             /* name of struct */
    array_t *metas;
} meta_tuple_t;

typedef struct meta_map_s {
    meta_t *k_meta;         /* key */
    meta_t *v_meta;         /* value */
} meta_map_t;

struct meta_s {
    type_t type;

    bool is_untyped;        /* integer or float literal, new map() */
    int arr_dim;            /* dimension of array */
    int *arr_size;          /* array size of each dimension */

    union {
        meta_map_t u_map;
        meta_tuple_t u_tup;
    };

    src_pos_t *pos;
};

char *meta_to_str(meta_t *x);

void meta_set_struct(meta_t *meta, char *name, array_t *ids);
void meta_set_tuple(meta_t *meta, array_t *exps);

int meta_check(meta_t *x, meta_t *y);

void meta_dump(meta_t *meta, int indent);

static inline void
meta_init(meta_t *meta, src_pos_t *pos)
{
    ASSERT(meta != NULL);
    ASSERT(pos != NULL);

    memset(meta, 0x00, sizeof(meta_t));

    meta->pos = pos;
}

static inline void
meta_set(meta_t *meta, type_t type)
{
    ASSERT(meta != NULL);
    ASSERT1(type > TYPE_NONE && type < TYPE_MAX, type);

    meta->type = type;
}

static inline void
meta_set_array(meta_t *meta, int arr_dim)
{
    ASSERT(meta != NULL);
    ASSERT(arr_dim >= 0);

    meta->arr_dim = arr_dim;
    meta->arr_size = xcalloc(sizeof(int) * arr_dim);
}

static inline void
meta_set_untyped(meta_t *meta, type_t type)
{
    meta_set(meta, type);

    meta->is_untyped = true;
}

static inline void
meta_set_map(meta_t *meta, meta_t *k_meta, meta_t *v_meta)
{
    meta_set(meta, TYPE_MAP);

    meta->u_map.k_meta = k_meta;
    meta->u_map.v_meta = v_meta;
}

static inline void
meta_copy(meta_t *dest, meta_t *src)
{
    ASSERT(dest != NULL);
    ASSERT(src != NULL);

    dest->type = src->type;
    dest->is_untyped = src->is_untyped;
    dest->arr_dim = src->arr_dim;
    dest->arr_size = src->arr_size;

    if (is_struct_meta(src) || is_tuple_meta(src)) {
        dest->u_tup.name = src->u_tup.name;
        dest->u_tup.metas = src->u_tup.metas;
    }
    else if (is_map_meta(src)) {
        dest->u_map.k_meta = src->u_map.k_meta;
        dest->u_map.v_meta = src->u_map.v_meta;
    }
}

static inline void
meta_merge(meta_t *meta, meta_t *x, meta_t *y)
{
    ASSERT(meta != NULL);

    if (is_untyped_meta(x) && is_untyped_meta(y)) {
        ASSERT1(is_primitive_meta(x), x->type);
        ASSERT1(is_primitive_meta(y), y->type);

        meta_set_untyped(meta, MAX(x->type, y->type));
    }
    else if (is_untyped_meta(x)) {
        meta_copy(meta, y);
    }
    else {
        meta_copy(meta, x);
    }
}

#endif /* ! _META_H */
