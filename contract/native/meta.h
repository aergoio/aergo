/**
 * @file    meta.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _META_H
#define _META_H

#include "common.h"

#include "enum.h"

#define is_none_type(type)          ((type) == TYPE_NONE)
#define is_bool_type(type)          ((type) == TYPE_BOOL)
#define is_byte_type(type)          ((type) == TYPE_BYTE)
#define is_int8_type(type)          ((type) == TYPE_INT8)
#define is_int16_type(type)         ((type) == TYPE_INT16)
#define is_int32_type(type)         ((type) == TYPE_INT32)
#define is_int64_type(type)         ((type) == TYPE_INT64)
#define is_int256_type(type)        ((type) == TYPE_INT256)
#define is_string_type(type)        ((type) == TYPE_STRING)
#define is_account_type(type)       ((type) == TYPE_ACCOUNT)
#define is_struct_type(type)        ((type) == TYPE_STRUCT)
#define is_map_type(type)           ((type) == TYPE_MAP)
#define is_object_type(type)        ((type) == TYPE_OBJECT)
#define is_void_type(type)          ((type) == TYPE_VOID)
#define is_tuple_type(type)         ((type) == TYPE_TUPLE)

#define is_none_meta(meta)          is_none_type((meta)->type)
#define is_bool_meta(meta)          is_bool_type((meta)->type)
#define is_byte_meta(meta)          is_byte_type((meta)->type)
#define is_int8_meta(meta)          is_int8_type((meta)->type)
#define is_int16_meta(meta)         is_int16_type((meta)->type)
#define is_int32_meta(meta)         is_int32_type((meta)->type)
#define is_int64_meta(meta)         is_int64_type((meta)->type)
#define is_int256_meta(meta)        is_int256_type((meta)->type)
#define is_string_meta(meta)        is_string_type((meta)->type)
#define is_account_meta(meta)       is_account_type((meta)->type)
#define is_struct_meta(meta)        is_struct_type((meta)->type)
#define is_map_meta(meta)           is_map_type((meta)->type)
#define is_object_meta(meta)        is_object_type((meta)->type)
#define is_void_meta(meta)          is_void_type((meta)->type)
#define is_tuple_meta(meta)         is_tuple_type((meta)->type)

#define is_unsigned_meta(meta)      (is_bool_meta(meta) || is_byte_meta(meta))
#define is_signed_meta(meta)        (!is_unsigned_meta(meta))

#define is_integer_meta(meta)                                                                      \
    (is_byte_meta(meta) || is_int8_meta(meta) || is_int16_meta(meta) || is_int32_meta(meta) ||     \
     is_int64_meta(meta) || is_int256_meta(meta))
#define is_numeric_meta(meta)       (is_integer_meta(meta))

#define is_nullable_meta(meta)                                                                     \
    ((is_array_meta(meta) && !is_fixed_array(meta)) || is_string_meta(meta) ||                     \
     is_struct_meta(meta) || is_map_meta(meta) || is_object_meta(meta))

#define is_address_meta(meta)                                                                      \
    (is_string_meta(meta) || is_array_meta(meta) || is_struct_meta(meta) || is_map_meta(meta))

#define is_primitive_meta(meta)     ((meta)->type <= TYPE_STRING)
#define is_comparable_meta(meta)    ((meta)->type > TYPE_NONE && (meta)->type <= TYPE_COMPARABLE)
#define is_compatible_meta(x, y)                                                                   \
    ((x)->type > TYPE_NONE && (x)->type <= TYPE_COMPATIBLE &&                                      \
     (y)->type > TYPE_NONE && (y)->type <= TYPE_COMPATIBLE &&                                      \
     (((x)->type == (y)->type) || is_string_meta(y) ||                                             \
      (is_string_meta(x) && !is_bool_meta(y)) ||                                                   \
      (is_numeric_meta(x) && is_numeric_meta(y))))

#define is_undef_meta(meta)         (meta)->is_undef

#define is_array_meta(meta)         ((meta)->arr_dim > 0)
#define is_fixed_array(meta)        ((meta)->is_fixed)

#define meta_set_bool(meta)         meta_set((meta), TYPE_BOOL)
#define meta_set_byte(meta)         meta_set((meta), TYPE_BYTE)
#define meta_set_int8(meta)         meta_set((meta), TYPE_INT8)
#define meta_set_int16(meta)        meta_set((meta), TYPE_INT16)
#define meta_set_int32(meta)        meta_set((meta), TYPE_INT32)
#define meta_set_int64(meta)        meta_set((meta), TYPE_INT64)
#define meta_set_int256(meta)       meta_set((meta), TYPE_INT256)
#define meta_set_string(meta)       meta_set((meta), TYPE_STRING)
#define meta_set_account(meta)      meta_set((meta), TYPE_ACCOUNT)
#define meta_set_void(meta)         meta_set((meta), TYPE_VOID)

#define meta_align(meta)            (meta)->align

#define meta_cnt(meta)                                                                             \
    ((is_tuple_meta(meta) || is_struct_meta(meta)) ? (meta)->elem_cnt : 1)

#define meta_iosz(meta)                                                                            \
    ((is_array_meta(meta) && !is_fixed_array(meta)) ? ADDR_SIZE : TYPE_C_SIZE((meta)->type))

#ifndef _META_T
#define _META_T
typedef struct meta_s meta_t;
#endif /* ! _META_T */

#ifndef _VECTOR_T
#define _VECTOR_T
typedef struct vector_s vector_t;
#endif /* ! _VECTOR_T */

#ifndef _AST_ID_T
#define _AST_ID_T
typedef struct ast_id_s ast_id_t;
#endif /* ! _AST_ID_T */

struct meta_s {
    type_t type;

    bool is_undef;          /* whether it is literal */

    /* array specifications */
    bool is_fixed;          /* whether it is fixed-length array */
    int max_dim;            /* maximum dimension */
    int arr_dim;            /* current dimension */
    int *dim_sizes;         /* size of each dimension */

    /* structured elements (e.g, struct, tuple, map) */
    int elem_cnt;
    meta_t **elems;

    ast_id_t *type_id;      /* identifier of struct, contract, interface */

    /* memory location to be stored */
    uint8_t align;          /* memory alignment (if struct, alignment of first element) */
    uint32_t base_idx;      /* register index having base address */
    uint32_t rel_addr;      /* relative address from "base_idx" */
    uint32_t rel_offset;    /* relative offset from "rel_addr" */

    src_pos_t *pos;
};

char *meta_to_str(meta_t *x);

void meta_set_map(meta_t *meta, meta_t *k, meta_t *v);
void meta_set_tuple(meta_t *meta, vector_t *elem_exps);

void meta_set_struct(meta_t *meta, ast_id_t *id);
void meta_set_object(meta_t *meta, ast_id_t *id);

bool meta_cmp(meta_t *x, meta_t *y);
void meta_eval(meta_t *x, meta_t *y);

static inline uint32_t meta_memsz(meta_t *meta);

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
    meta->align = TYPE_ALIGN(type);
}

static inline void
meta_set_undef(meta_t *meta)
{
    meta->is_undef = true;
}

static inline void
meta_set_arr_dim(meta_t *meta, int arr_dim)
{
    int i;

    ASSERT1(arr_dim > 0, arr_dim);

    meta->is_fixed = true;
    meta->max_dim = arr_dim;
    meta->arr_dim = arr_dim;

    meta->dim_sizes = xmalloc(sizeof(int) * arr_dim);
    for (i = 0; i < arr_dim; i++) {
        meta->dim_sizes[i] = -1;
    }
}

static inline void
meta_set_dim_sz(meta_t *meta, int dim, int size)
{
    ASSERT1(dim >= 0, dim);
    ASSERT1(size > 0, size);
    ASSERT1(meta->arr_dim > 0, meta->arr_dim);

    meta->is_fixed = true;
    meta->dim_sizes[dim] = size;
}

static inline void
meta_strip_arr_dim(meta_t *meta)
{
    ASSERT1(meta->arr_dim > 0, meta->arr_dim);

    meta->arr_dim--;

    if (meta->arr_dim == 0)
        meta->dim_sizes = NULL;
    else
        meta->dim_sizes = &meta->dim_sizes[1];
}

static inline uint32_t
meta_typsz(meta_t *meta)
{
    uint32_t size = 0;

    if (is_struct_meta(meta)) {
        int i;

        for (i = 0; i < meta->elem_cnt; i++) {
            size = ALIGN(size, meta_align(meta->elems[i])) + meta_memsz(meta->elems[i]);
        }
    }
    else {
        size = TYPE_SIZE(meta->type);
    }

    return size;
}

static inline uint32_t
meta_memsz(meta_t *meta)
{
    uint32_t size = meta_typsz(meta);

    ASSERT(!is_tuple_meta(meta));

    if (is_array_meta(meta)) {
        int i;
        uint32_t dim_sz = 1;

        ASSERT2(meta_align(meta) > 0, meta->type, meta_align(meta));

        if (!is_fixed_array(meta))
            return ADDR_SIZE;

        /* element size */
        for (i = 0; i < meta->arr_dim; i++) {
            ASSERT1(meta->dim_sizes[i] > 0, meta->dim_sizes[i]);
            size *= meta->dim_sizes[i];
        }

        /* Each dimension has the count of elements (4 or 8 bytes) as a header. */
        size += meta_align(meta);
        for (i = 0; i < meta->arr_dim - 1; i++) {
            dim_sz *= meta->dim_sizes[i];
            size += dim_sz * meta_align(meta);
        }
    }

    return size;
}

static inline void
meta_copy(meta_t *dest, meta_t *src)
{
    /* deliberately excluded base_idx, rel_addr, rel_offset, pos */
    memcpy(dest, src, offsetof(meta_t, base_idx));
}

#endif /* ! _META_H */
