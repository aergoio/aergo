/**
 * @file    value.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _VALUE_H
#define _VALUE_H

#include "common.h"

#define is_null_val(val)            ((val)->len == 0)
#define is_bool_val(val)            ((val)->type == TYPE_BOOL)
#define is_ui32_val(val)            ((val)->type == TYPE_UINT32)
#define is_ui64_val(val)            ((val)->type == TYPE_UINT64)
#define is_f32_val(val)             ((val)->type == TYPE_FLOAT)
#define is_f64_val(val)             ((val)->type == TYPE_DOUBLE)
#define is_str_val(val)             ((val)->type == TYPE_STRING)
#define is_obj_val(val)             ((val)->type == TYPE_OBJECT)

#define bool_val(val)               ((val)->b)
#define ui32_val(val)               ((val)->is_neg ? -(val)->ui32 : (val)->ui32)
#define ui64_val(val)               ((val)->is_neg ? -(val)->ui64 : (val)->ui64)
#define f32_val(val)                ((val)->f)
#define f64_val(val)                ((val)->d)
#define str_val(val)                ((val)->s)
#define obj_val(val)                ((val)->p)

#define is_zero_val(val)                                                                 \
    (is_ui64_val(val) ? (val)->ui64 == 0 : (is_f64_val(val) ? (val)->d == 0.0f : false))

#define value_set_null(val)         (val)->len = 0

#define value_set_bool(val, v)                                                           \
    do {                                                                                 \
        (val)->type = TYPE_BOOL;                                                         \
        (val)->len = sizeof(bool);                                                       \
        (val)->b = (v);                                                                  \
    } while (0)

#define value_set_ui32(val, v)                                                           \
    do {                                                                                 \
        (val)->type = TYPE_UINT32;                                                       \
        (val)->len = sizeof(uint32_t);                                                   \
        (val)->ui32 = (v);                                                               \
    } while (0)

#define value_set_ui64(val, v)                                                           \
    do {                                                                                 \
        (val)->type = TYPE_UINT64;                                                       \
        (val)->len = sizeof(uint64_t);                                                   \
        (val)->ui64 = (v);                                                               \
    } while (0)

#define value_set_f32(val, v)                                                            \
    do {                                                                                 \
        (val)->type = TYPE_FLOAT;                                                        \
        (val)->len = sizeof(float);                                                      \
        (val)->f = (v);                                                                  \
    } while (0)

#define value_set_f64(val, v)                                                            \
    do {                                                                                 \
        (val)->type = TYPE_DOUBLE;                                                       \
        (val)->len = sizeof(double);                                                     \
        (val)->d = (v);                                                                  \
    } while (0)

#define value_set_str(val, v)                                                            \
    do {                                                                                 \
        (val)->type = TYPE_STRING;                                                       \
        (val)->len = strlen(v);                                                          \
        (val)->s = (v);                                                                  \
    } while (0)

#define value_set_obj(val, v)                                                            \
    do {                                                                                 \
        (val)->type = TYPE_OBJECT;                                                       \
        (val)->len = sizeof(void *);                                                     \
        (val)->p = (v);                                                                  \
    } while (0)

#ifndef _VALUE_T
#define _VALUE_T
typedef struct value_s value_t;
#endif /* ! _VALUE_T */

#ifndef _META_T
#define _META_T
typedef struct meta_s meta_t;
#endif /* ! _META_T */

typedef void (*eval_fn_t)(value_t *, value_t *, value_t *) ;

struct value_s {
    type_t type;
    int len;
    bool is_neg;

    union {
        bool b;
        uint32_t ui32;
        uint64_t ui64;
        float f;
        double d;
        char *s;
        void *p;
    };
};

bool value_fit(value_t *val, meta_t *meta);

int value_cmp(value_t *x, value_t *y);

void value_eval(op_kind_t op, value_t *x, value_t *y, value_t *res);
void value_cast(value_t *val, meta_t *meta);

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
