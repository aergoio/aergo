/**
 * @file    value.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _VALUE_H
#define _VALUE_H

#include "common.h"

#define is_null_val(val)            ((val)->size == 0)
#define is_signed_val(val)          ((val)->is_neg)

#define is_bool_val(val)            ((val)->type == TYPE_BOOL)
#define is_i64_val(val)             ((val)->type == TYPE_UINT64)
#define is_f64_val(val)             ((val)->type == TYPE_DOUBLE)
#define is_str_val(val)             ((val)->type == TYPE_STRING)
#define is_ptr_val(val)             ((val)->type == TYPE_OBJECT)

#define val_ptr(val)                ((val)->ptr)
#define val_size(val)               ((val)->size)

#define val_bool(val)               ((val)->b)
#define val_i64(val)                ((val)->is_neg ? -(val)->i64 : (val)->i64)
#define val_f64(val)                ((val)->d)
#define val_str(val)                ((val)->ptr)

#define is_zero_val(val)                                                                 \
    (is_i64_val(val) ? (val)->i64 == 0 : (is_f64_val(val) ? (val)->d == 0.0f : false))

#define value_set_bool(val, v)                                                           \
    do {                                                                                 \
        (val)->type = TYPE_BOOL;                                                         \
        (val)->size = sizeof(bool);                                                      \
        (val)->ptr = &(val)->b;                                                          \
        (val)->b = (v);                                                                  \
    } while (0)

#define value_set_i64(val, v)                                                            \
    do {                                                                                 \
        (val)->type = TYPE_UINT64;                                                       \
        (val)->size = sizeof(uint64_t);                                                  \
        (val)->ptr = &(val)->i64;                                                        \
        (val)->i64 = (v);                                                                \
    } while (0)

#define value_set_f64(val, v)                                                            \
    do {                                                                                 \
        (val)->type = TYPE_DOUBLE;                                                       \
        (val)->size = sizeof(double);                                                    \
        (val)->ptr = &(val)->d;                                                          \
        (val)->d = (v);                                                                  \
    } while (0)

#define value_set_str(val, v)                                                            \
    do {                                                                                 \
        (val)->type = TYPE_STRING;                                                       \
        (val)->size = strlen(v);                                                         \
        (val)->ptr = (v);                                                                \
    } while (0)

#define value_set_ptr(val, v, l)                                                         \
    do {                                                                                 \
        (val)->type = TYPE_OBJECT;                                                       \
        (val)->size = l;                                                                 \
        (val)->ptr = (v);                                                                \
    } while (0)

#define value_set_null(val)         value_set_ptr(val, NULL, 0)

#ifndef _VALUE_T
#define _VALUE_T
typedef struct value_s value_t;
#endif /* ! _VALUE_T */

#ifndef _META_T
#define _META_T
typedef struct meta_s meta_t;
#endif /* ! _META_T */

typedef void (*eval_fn_t)(value_t *, value_t *, value_t *) ;
typedef void (*cast_fn_t)(value_t *) ;

struct value_s {
    type_t type;
    int size;
    bool is_neg;

    void *ptr;
    union {
        bool b;
        uint64_t i64;
        double d;
        char *s;
    };
};

bool value_fit(value_t *val, meta_t *meta);

int value_cmp(value_t *x, value_t *y);

void value_eval(value_t *x, op_kind_t op, value_t *y, value_t *res);
void value_cast(value_t *val, type_t type);

static inline void
value_init(value_t *val)
{
    ASSERT(val != NULL);
    memset(val, 0x00, sizeof(value_t));
}

static inline void
value_set_neg(value_t *val, bool is_neg)
{
    val->is_neg = is_neg;
}

#endif /* ! _VALUE_H */
