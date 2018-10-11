/**
 * @file    value.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _VALUE_H
#define _VALUE_H

#include "common.h"

#define is_null_val(val)            ((val)->kind == VAL_NULL)
#define is_bool_val(val)            ((val)->kind == VAL_BOOL)
#define is_int_val(val)             ((val)->kind == VAL_INT)
#define is_fp_val(val)              ((val)->kind == VAL_FP)
#define is_string_val(val)          ((val)->kind == VAL_STR)

#define is_zero_val(val)                                                       \
    (is_int_val(val) ? (val)->iv == 0 :                                        \
     (is_fp_val(val) ? (val)->dv == 0.0f : false))

#define value_eval(op, val, x, y)                                              \
    do {                                                                       \
        ASSERT1((op) >= OP_ADD && (op) < OP_CF_MAX, (op));                     \
        eval_fntab_[(op)]((val), (x), (y));                                    \
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
    val_kind_t kind;

    union {
        bool bv;
        int64_t iv;
        double dv;
        char *sv;
    };
};

extern eval_fn_t eval_fntab_[OP_CF_MAX];

static inline void
value_init(value_t *val)
{
    ASSERT(val != NULL);
    memset(val, 0x00, sizeof(value_t));
}

static inline void
value_set_null(value_t *val)
{
    val->kind = VAL_NULL;
}

static inline void
value_set_bool(value_t *val, bool bv)
{
    val->kind = VAL_BOOL;
    val->bv = bv;
}

static inline void
value_set_int(value_t *val, int64_t iv)
{
    val->kind = VAL_INT;
    val->iv = iv;
}

static inline void
value_set_double(value_t *val, double dv)
{
    val->kind = VAL_FP;
    val->dv = dv;
}

static inline void
value_set_str(value_t *val, char *str)
{
    val->kind = VAL_STR;
    val->sv = str;
}

static inline bool
value_check_range(value_t *val, type_t type)
{
    if (type == TYPE_INT8 && (val->iv < INT8_MIN || val->iv > INT8_MAX))
        return false;

    if (type == TYPE_UINT8 && val->iv > UINT8_MAX)
        return false;

    if (type == TYPE_INT16 && (val->iv < INT16_MIN || val->iv > INT16_MAX))
        return false;

    if (type == TYPE_UINT16 && val->iv > UINT16_MAX)
        return false;

    if (type == TYPE_INT32 && (val->iv < INT32_MIN || val->iv > INT32_MAX))
        return false;

    if (type == TYPE_UINT32 && val->iv > UINT32_MAX)
        return false;

    if (type == TYPE_FLOAT && (val->dv < FLT_MIN || val->dv > FLT_MAX))
        return false;

    return true;
}

#endif /* ! _VALUE_H */
