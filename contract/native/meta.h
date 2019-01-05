/**
 * @file    meta.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _META_H
#define _META_H

#include "common.h"

#include "ast.h"
#include "array.h"
#include "value.h"

#define is_none_type(meta)          ((meta)->type == TYPE_NONE)
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

#define is_signed_type(meta)                                                             \
    (is_int8_type(meta) || is_int16_type(meta) || is_int32_type(meta) ||                 \
     is_int64_type(meta))
#define is_unsigned_type(meta)                                                           \
    (is_byte_type(meta) ||   is_uint8_type(meta) ||  is_uint16_type(meta) ||             \
     is_uint32_type(meta) || is_uint64_type(meta))

#define is_integer_type(meta)       (is_signed_type(meta) || is_unsigned_type(meta))
#define is_fpoint_type(meta)        (is_float_type(meta) || is_double_type(meta))
#define is_numeric_type(meta)       (is_integer_type(meta) || is_fpoint_type(meta))

#define is_pointer_type(meta)                                                            \
    (is_string_type(meta) || is_map_type(meta) || is_object_type(meta))

#define is_primitive_type(meta)                                                          \
    ((meta)->type > TYPE_NONE && (meta)->type <= TYPE_PRIMITIVE)
#define is_builtin_type(meta)                                                            \
    ((meta)->type > TYPE_NONE && (meta)->type <= TYPE_BUILTIN)
#define is_comparable_type(meta)                                                         \
    ((meta)->type > TYPE_NONE && (meta)->type <= TYPE_COMPARABLE)
#define is_compatible_type(x, y)                                                         \
    ((x)->type > TYPE_NONE && (x)->type <= TYPE_COMPATIBLE &&                            \
     (y)->type > TYPE_NONE && (y)->type <= TYPE_COMPATIBLE)

#define is_array_type(meta)         ((meta)->arr_dim > 0)
#define is_undef_type(meta)         (meta)->is_undef

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
#define meta_set_account(meta)      meta_set((meta), TYPE_ACCOUNT)
#define meta_set_void(meta)         meta_set((meta), TYPE_VOID)

#define meta_size(meta)             ((meta)->size)
#define meta_align(meta)            TYPE_ALIGN((meta)->type)

#define meta_cnt(meta)                                                                   \
    (is_void_type(meta) ? 0 :                                                            \
     ((is_tuple_type(meta) || is_struct_type(meta)) ? (meta)->u_tup.elem_cnt : 1))

#ifndef _META_T
#define _META_T
typedef struct meta_s meta_t;
#endif /* ! _META_T */

#ifndef _AST_ID_T
#define _AST_ID_T
typedef struct ast_id_s ast_id_t;
#endif /* ! _AST_ID_T */

typedef struct meta_tuple_s {
    int elem_cnt;
    meta_t **elems;
} meta_tuple_t;

typedef struct meta_object_s {
    ast_id_t *id;
} meta_object_t;

struct meta_s {
    type_t type;
    int size;               /* unit size of type */

    char *name;             /* name of struct or contract */

    int arr_dim;            /* dimension of array */
    int arr_size;           /* total size of array */
    int *dim_sizes;         /* size of each dimension */

    bool is_undef;          /* whether it is literal */

    union {
        meta_tuple_t u_tup;
        meta_object_t u_obj;
    };

    int num;
    src_pos_t *pos;
};

char *meta_to_str(meta_t *x);

void meta_set_map(meta_t *meta, meta_t *k, meta_t *v);
void meta_set_struct(meta_t *meta, char *name, array_t *ids);
void meta_set_tuple(meta_t *meta, array_t *elem_exps);
void meta_set_object(meta_t *meta, ast_id_t *id);

int meta_cmp(meta_t *x, meta_t *y);

void meta_eval(meta_t *x, meta_t *y);

void meta_dump(meta_t *meta, int indent);

static inline void
meta_init(meta_t *meta, src_pos_t *pos)
{
    ASSERT(meta != NULL);
    ASSERT(pos != NULL);

    memset(meta, 0x00, sizeof(meta_t));

    meta->num = node_num_++;
    meta->pos = pos;
}

static inline meta_t *
meta_new(type_t type, src_pos_t *pos)
{
    meta_t *meta = xmalloc(sizeof(meta_t));

    meta_init(meta, pos);

    meta->type = type;
    meta->size = TYPE_SIZE(type);

    meta->num = node_num_++;

    meta->pos = xmalloc(sizeof(src_pos_t));
    memcpy(meta->pos, pos, sizeof(src_pos_t));

    return meta;
}

static inline void
meta_set(meta_t *meta, type_t type)
{
    ASSERT1(is_valid_type(type), type);

    meta->type = type;
    meta->size = TYPE_SIZE(type);
}

static inline void
meta_set_undef(meta_t *meta)
{
    meta->is_undef = true;
}

static inline void
meta_set_arr_dim(meta_t *meta, int arr_dim)
{
    ASSERT(arr_dim > 0);

    meta->arr_dim = arr_dim;
    meta->arr_size = ALIGN(meta->size, TYPE_ALIGN(meta->type));
    meta->dim_sizes = xcalloc(sizeof(int) * arr_dim);
}

static inline void
meta_set_dim_size(meta_t *meta, int dim, int size)
{
    ASSERT(dim >= 0);
    ASSERT(meta->arr_dim > 0);

    meta->dim_sizes[dim] = size;

    if (size > 0)
        meta->arr_size *= size;
}

static inline void
meta_strip_arr_dim(meta_t *meta)
{
    ASSERT1(meta->arr_dim > 0, meta->arr_dim);

    meta->arr_dim--;

    if (meta->dim_sizes[0] > 0)
        /* In case of a parameter, dim_size can be negative */
        meta->arr_size /= meta->dim_sizes[0];

    if (meta->arr_dim == 0)
        meta->dim_sizes = NULL;
    else
        meta->dim_sizes = &meta->dim_sizes[1];
}

static inline void
meta_copy(meta_t *dest, meta_t *src)
{
    dest->type = src->type;
    dest->size = src->size;
    dest->name = src->name;
    dest->arr_dim = src->arr_dim;
    dest->arr_size = src->arr_size;
    dest->dim_sizes = src->dim_sizes;
    dest->is_undef = src->is_undef;

    if (is_tuple_type(src) || is_map_type(src) || is_struct_type(src)) {
        dest->u_tup.elem_cnt = src->u_tup.elem_cnt;
        dest->u_tup.elems = src->u_tup.elems;
    }
    else if (is_object_type(src)) {
        dest->u_obj.id = src->u_obj.id;
    }

    /* deliberately excluded num and src_pos */
}

#endif /* ! _META_H */
