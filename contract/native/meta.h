/**
 * @file    meta.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _META_H
#define _META_H

#include "common.h"

#include "array.h"
#include "value.h"

#define is_bool_type(meta)          ((meta)->type == TYPE_BOOL)
#define is_byte_type(meta)          ((meta)->type == TYPE_BYTE)
#define is_int8_type(meta)          ((meta)->type == TYPE_INT8)
#define is_uint8_type(meta)         ((meta)->type == TYPE_UINT8)
#define is_int16_type(meta)         ((meta)->type == TYPE_INT16)
#define is_uint16_type(meta)        ((meta)->type == TYPE_UINT16)
#define is_int32_type(meta)         ((meta)->type == TYPE_INT32)
#define is_uint32_type(meta)        ((meta)->type == TYPE_UINT32)
#define is_int64_type(meta)         ((meta)->type == TYPE_INT64)
#define is_uint64_type(meta)        ((meta)->type == TYPE_UINT64)
#define is_float_type(meta)         ((meta)->type == TYPE_FLOAT)
#define is_double_type(meta)        ((meta)->type == TYPE_DOUBLE)
#define is_string_type(meta)        ((meta)->type == TYPE_STRING)
#define is_account_type(meta)       ((meta)->type == TYPE_ACCOUNT)
#define is_struct_type(meta)        ((meta)->type == TYPE_STRUCT)
#define is_map_type(meta)           ((meta)->type == TYPE_MAP)
#define is_object_type(meta)        ((meta)->type == TYPE_OBJECT)
#define is_void_type(meta)          ((meta)->type == TYPE_VOID)
#define is_tuple_type(meta)         ((meta)->type == TYPE_TUPLE)

#define is_int_family(meta)                                                              \
    (is_int8_type(meta) || is_int16_type(meta) || is_int32_type(meta) ||                 \
     is_int64_type(meta))
#define is_uint_family(meta)                                                             \
    (is_byte_type(meta) ||   is_uint8_type(meta) ||  is_uint16_type(meta) ||             \
     is_uint32_type(meta) || is_uint64_type(meta))

#define is_dec_family(meta)         (is_int_family(meta) || is_uint_family(meta))
#define is_fp_family(meta)          (is_float_type(meta) || is_double_type(meta))
#define is_num_family(meta)         (is_dec_family(meta) || is_fp_family(meta))

#define is_primitive_type(meta)                                                          \
    ((meta)->type > TYPE_NONE && (meta)->type <= TYPE_PRIMITIVE)
#define is_builtin_type(meta)                                                            \
    ((meta)->type > TYPE_NONE && (meta)->type <= TYPE_BUILTIN)
#define is_comparable_type(meta)                                                         \
    ((meta)->type > TYPE_NONE && (meta)->type <= TYPE_COMPARABLE)

#define is_compatible_type(x, y)    (is_primitive_type(x) && is_primitive_type(y))

#define is_undef_type(meta)         (meta)->is_undef
#define is_array_type(meta)         ((meta)->arr_dim > 0)

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

#define meta_cnt(meta)                                                                   \
    (is_void_type(meta) ? 0 :                                                            \
     ((is_tuple_type(meta) || is_struct_type(meta)) ?  (meta)->elem_cnt : 1))

#ifndef _META_T
#define _META_T
typedef struct meta_s meta_t;
#endif /* ! _META_T */

struct meta_s {
    type_t type;

    char *name;             /* name of struct */

    int arr_dim;            /* dimension of array */
    int *arr_size;          /* size of each dimension */

    bool is_undef;          /* whether it is numeric literal */

    /* structured meta (map, struct or tuple) */
    int elem_cnt;           /* count of elements */
    meta_t **elems;         /* metas of elements */

    src_pos_t *pos;
};

char *meta_to_str(meta_t *x);

void meta_set_map(meta_t *meta, meta_t *k, meta_t *v);
void meta_set_struct(meta_t *meta, char *name, array_t *ids);
void meta_set_tuple(meta_t *meta, array_t *exps);

int meta_cmp(meta_t *x, meta_t *y);

int meta_size(meta_t *meta);

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
meta_set_undef(meta_t *meta)
{
    meta->is_undef = true;
}

static inline void
meta_set_array(meta_t *meta, int arr_dim)
{
    ASSERT(arr_dim >= 0);

    meta->arr_dim = arr_dim;
    meta->arr_size = xcalloc(sizeof(int) * arr_dim);
}

static inline void
meta_eval(meta_t *meta, meta_t *x, meta_t *y)
{
    ASSERT(meta != NULL);

    if (is_undef_type(x) && is_undef_type(y)) {
        ASSERT1(is_builtin_type(x), x->type);
        ASSERT1(is_builtin_type(y), y->type);

        meta_set(meta, MAX(x->type, y->type));
        meta_set_undef(meta);
    }
    else if (is_undef_type(x)) {
        meta_copy(meta, y);
    }
    else {
        meta_copy(meta, x);
    }
}

#endif /* ! _META_H */
