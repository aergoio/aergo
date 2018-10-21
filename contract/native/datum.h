/**
 * @file    datum.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _DATUM_H
#define _DATUM_H

#include "common.h"

#include "array.h"

#define is_bool_type(dat)           ((dat)->type == TYPE_BOOL)
#define is_byte_type(dat)           ((dat)->type == TYPE_BYTE)

#define is_int8_type(dat)           ((dat)->type == TYPE_INT8)
#define is_uint8_type(dat)          ((dat)->type == TYPE_UINT8)
#define is_int16_type(dat)          ((dat)->type == TYPE_INT16)
#define is_uint16_type(dat)         ((dat)->type == TYPE_UINT16)
#define is_int32_type(dat)          ((dat)->type == TYPE_INT32)
#define is_uint32_type(dat)         ((dat)->type == TYPE_UINT32)
#define is_int64_type(dat)          ((dat)->type == TYPE_INT64)
#define is_uint64_type(dat)         ((dat)->type == TYPE_UINT64)
#define is_float_type(dat)          ((dat)->type == TYPE_FLOAT)
#define is_double_type(dat)         ((dat)->type == TYPE_DOUBLE)

#define is_int_type(dat)                                                                \
    (is_byte_type(dat) || ((dat)->type >= TYPE_INT8 && (dat)->type <= TYPE_UINT64))
#define is_fp_type(dat)                                                                 \
    ((dat)->type == TYPE_FLOAT || (dat)->type == TYPE_DOUBLE)
#define is_numeric_type(dat)        (is_int_type(dat) || is_fp_type(dat))

#define is_string_type(dat)         ((dat)->type == TYPE_STRING)

#define is_struct_type(dat)         ((dat)->type == TYPE_STRUCT)
#define is_map_type(dat)            ((dat)->type == TYPE_MAP)
#define is_object_type(dat)         ((dat)->type == TYPE_OBJECT)
#define is_void_type(dat)           ((dat)->type == TYPE_VOID)
#define is_tuple_type(dat)          ((dat)->type == TYPE_TUPLE)

#define is_primitive_type(dat)                                                          \
    ((dat)->type > TYPE_NONE && (dat)->type <= TYPE_PRIMITIVE)
#define is_builtin_type(dat)                                                            \
    ((dat)->type > TYPE_NONE && (dat)->type <= TYPE_BUILTIN)
#define is_comparable_type(dat)                                                         \
    ((dat)->type > TYPE_NONE && (dat)->type <= TYPE_COMPARABLE)

#define is_compatible_type(x, y)    (is_primitive_type(x) && is_primitive_type(y))

#define is_array_type(dat)          ((dat)->arr_dim > 0)

#define datum_set_bool(dat)         datum_set_type((dat), TYPE_BOOL)
#define datum_set_byte(dat)         datum_set_type((dat), TYPE_BYTE)
#define datum_set_int8(dat)         datum_set_type((dat), TYPE_INT8)
#define datum_set_uint8(dat)        datum_set_type((dat), TYPE_UINT8)
#define datum_set_int16(dat)        datum_set_type((dat), TYPE_INT16)
#define datum_set_uint16(dat)       datum_set_type((dat), TYPE_UINT16)
#define datum_set_int32(dat)        datum_set_type((dat), TYPE_INT32)
#define datum_set_uint32(dat)       datum_set_type((dat), TYPE_UINT32)
#define datum_set_int64(dat)        datum_set_type((dat), TYPE_INT64)
#define datum_set_uint64(dat)       datum_set_type((dat), TYPE_UINT64)
#define datum_set_float(dat)        datum_set_type((dat), TYPE_FLOAT)
#define datum_set_double(dat)       datum_set_type((dat), TYPE_DOUBLE)
#define datum_set_string(dat)       datum_set_type((dat), TYPE_STRING)
#define datum_set_object(dat)       datum_set_type((dat), TYPE_OBJECT)
#define datum_set_account(dat)      datum_set_type((dat), TYPE_ACCOUNT)
#define datum_set_void(dat)         datum_set_type((dat), TYPE_VOID)

#define is_neg_val(dat)             ((dat)->is_neg)

#ifndef _DATUM_T
#define _DATUM_T
typedef struct datum_s datum_t;
#endif  /* ! _DATUM_T */

typedef struct dat_map_s dat_map_t {
    datum_t *k;             /* key of map */
    datum_t *v;             /* value of map */
};

typedef struct dat_tup_s dat_tup_t {
    int cnt;                /* count of elements */
    datum_t **evs;          /* datums of elements */
};

typedef union value_u {
    bool bv;
    uint64_t iv;
    double fv;
    char *sv;
    void *pv;
} value_t;

struct datum_s {
    type_t type;

    char *name;             /* name of struct */

    int arr_dim;            /* dimension of array */
    int *arr_size;          /* size of each dimension */

    /* structured data */
    union {
        dat_map_t map;
        dat_tup_t tup;
    };

    bool is_const;          /* whether it is constant */
    bool is_neg;            /* whether it is negative */

    value_t cv;             /* constant value */

    src_pos_t *pos;
};

char *datum_type(datum_t *dat);

void datum_set_array(datum_t *dat, int arr_dim);
void datum_set_map(datum_t *dat, datum_t *k, datum_t *v);
void datum_set_struct(datum_t *dat, char *name, array_t *ids);
void datum_set_tuple(datum_t *dat, array_t *exps);

void datum_cast(datum_t *dat, type_t type);

static inline void
datum_init(datum_t *dat, src_pos_t *pos)
{
    ASSERT(pos != NULL);

    memset(dat, 0x00, sizeof(datum_t));
    dat->pos = pos;
}

static inline void
datum_set_type(datum_t *dat, type_t type)
{
    ASSERT1(is_valid_type(type), type);

    dat->type = type;
}

#endif /* no _DATUM_H */
