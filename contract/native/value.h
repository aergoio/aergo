/**
 * @file    value.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _VALUE_H
#define _VALUE_H

#include "common.h"

#define is_none_val(val)            ((val)->kind == VAL_NONE)
#define is_null_val(val)            ((val)->kind == VAL_NULL)
#define is_bool_val(val)            ((val)->kind == VAL_BOOL)
#define is_int_val(val)             ((val)->kind == VAL_INT)
#define is_fp_val(val)              ((val)->kind == VAL_FP)
#define is_str_val(val)             ((val)->kind == VAL_STR)

#define bool_val(val)               ((val)->bv)
#define int_val(val)                ((val)->is_neg ? -(val)->iv : (val)->iv)
#define fp_val(val)                 ((val)->is_neg ? -(val)->dv : (val)->dv)
#define str_val(val)                ((val)->sv)

#define is_zero_val(val)                                                                 \
    (is_int_val(val) ? (val)->iv == 0 : (is_fp_val(val) ? (val)->dv == 0.0f : false))

#define value_eval(op, val, x, y)                                                        \
    do {                                                                                 \
        ASSERT1((op) >= OP_ADD && (op) < OP_CF_MAX, (op));                               \
        eval_fntab_[(op)]((val), (x), (y));                                              \
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

    bool is_neg;

    union {
        bool bv;
        uint64_t iv;
        double dv;
        char *sv;
    };
};

extern eval_fn_t eval_fntab_[OP_CF_MAX];

bool value_check(value_t *val, meta_t *meta);

int value_cmp(value_t *x, value_t *y);

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
value_set_int(value_t *val, uint64_t iv)
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

static inline void
value_set_neg(value_t *val, bool is_neg)
{
    val->is_neg = is_neg;
}

#endif /* ! _VALUE_H */
