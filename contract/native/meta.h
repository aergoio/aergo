/**
 * @file    meta.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _META_H
#define _META_H

#include "common.h"

#include "array.h"
#include "value.h"

#define is_bool_meta(meta)          ((meta)->type == TYPE_BOOL)
#define is_byte_meta(meta)          ((meta)->type == TYPE_BYTE)
#define is_int8_meta(meta)          ((meta)->type == TYPE_INT8)
#define is_uint8_meta(meta)         ((meta)->type == TYPE_UINT8)
#define is_int16_meta(meta)         ((meta)->type == TYPE_INT16)
#define is_uint16_meta(meta)        ((meta)->type == TYPE_UINT16)
#define is_int32_meta(meta)         ((meta)->type == TYPE_INT32)
#define is_uint32_meta(meta)        ((meta)->type == TYPE_UINT32)
#define is_int64_meta(meta)         ((meta)->type == TYPE_INT64)
#define is_uint64_meta(meta)        ((meta)->type == TYPE_UINT64)
#define is_float_meta(meta)         ((meta)->type == TYPE_FLOAT)
#define is_double_meta(meta)        ((meta)->type == TYPE_DOUBLE)
#define is_string_meta(meta)        ((meta)->type == TYPE_STRING)
#define is_account_meta(meta)       ((meta)->type == TYPE_ACCOUNT)
#define is_struct_meta(meta)        ((meta)->type == TYPE_STRUCT)
#define is_map_meta(meta)           ((meta)->type == TYPE_MAP)
#define is_object_meta(meta)        ((meta)->type == TYPE_OBJECT)
#define is_void_meta(meta)          ((meta)->type == TYPE_VOID)
#define is_tuple_meta(meta)         ((meta)->type == TYPE_TUPLE)

#define is_int_meta(meta)                                                                \
    (is_byte_meta(meta) ||                                                               \
     is_int8_meta(meta) || is_uint8_meta(meta) ||                                        \
     is_int16_meta(meta) || is_uint16_meta(meta) ||                                      \
     is_int32_meta(meta) || is_uint32_meta(meta) ||                                      \
     is_int64_meta(meta) || is_uint64_meta(meta))

#define is_fp_meta(meta)            (is_float_meta(meta) || is_double_meta(meta))
#define is_num_meta(meta)           (is_int_meta(meta) || is_fp_meta(meta))

#define is_const_meta(meta)         (meta)->is_const
#define is_array_meta(meta)         ((meta)->arr_dim > 0)

#define is_primitive_meta(meta)                                                          \
    ((meta)->type > TYPE_NONE && (meta)->type <= TYPE_PRIMITIVE)
#define is_builtin_meta(meta)                                                            \
    ((meta)->type > TYPE_NONE && (meta)->type <= TYPE_BUILTIN)
#define is_comparable_meta(meta)                                                         \
    ((meta)->type > TYPE_NONE && (meta)->type <= TYPE_COMPARABLE)

#define is_compatible_meta(x, y)    (is_primitive_meta(x) && is_primitive_meta(y))

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
#define meta_set_object(meta)       meta_set((meta), TYPE_OBJECT)
#define meta_set_account(meta)      meta_set((meta), TYPE_ACCOUNT)
#define meta_set_void(meta)         meta_set((meta), TYPE_VOID)

#define meta_copy(dest, src)        *(dest) = *(src)

#define meta_elem_size(meta)                                                             \
    (is_void_meta(meta) ? 0 :                                                            \
     ((is_tuple_meta(meta) || is_struct_meta(meta)) ?  (meta)->elem_cnt : 1))

#ifndef _META_T
#define _META_T
typedef struct meta_s meta_t;
#endif /* ! _META_T */

struct meta_s {
    type_t type;

    char *name;             /* name of struct */

    int arr_dim;            /* dimension of array */
    int *arr_size;          /* size of each dimension */

    bool is_const;          /* whether it is constant */

    /* structured meta (array, map, struct initializer or tuple) */
    int elem_cnt;           /* count of elements */
    meta_t **elems;         /* metas of elements */

    src_pos_t *pos;
};

char *meta_to_str(meta_t *x);

void meta_set_map(meta_t *meta, meta_t *k, meta_t *v);
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
    ASSERT1(is_valid_type(type), type);

    meta->type = type;
}

static inline void
meta_set_const(meta_t *meta)
{
    meta->is_const = true;
}

static inline void
meta_set_array(meta_t *meta, int arr_dim)
{
    ASSERT(arr_dim >= 0);

    meta->arr_dim = arr_dim;
    meta->arr_size = xcalloc(sizeof(int) * arr_dim);
}

static inline void
meta_merge(meta_t *meta, meta_t *x, meta_t *y)
{
    ASSERT(meta != NULL);

    if (is_const_meta(x) && is_const_meta(y)) {
        ASSERT1(is_builtin_meta(x), x->type);
        ASSERT1(is_builtin_meta(y), y->type);

        meta_set(meta, MAX(x->type, y->type));
        meta_set_const(meta);
    }
    else if (is_const_meta(x)) {
        meta_copy(meta, y);
    }
    else {
        meta_copy(meta, x);
    }
}

#endif /* ! _META_H */
