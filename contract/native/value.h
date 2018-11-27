/**
 * @file    value.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _VALUE_H
#define _VALUE_H

#include "common.h"

#define is_none_val(val)            ((val)->kind == VAL_NONE)
#define is_bool_val(val)            ((val)->kind == VAL_BOOL)
#define is_int_val(val)             ((val)->kind == VAL_INT)
#define is_fp_val(val)              ((val)->kind == VAL_FP)
#define is_str_val(val)             ((val)->kind == VAL_STR)
#define is_obj_val(val)             ((val)->kind == VAL_OBJ)
#define is_addr_val(val)            ((val)->kind == VAL_ADDR)

#define bool_val(val)               ((val)->b)
#define int_val(val)                ((val)->is_neg ? -(val)->i : (val)->i)
#define fp_val(val)                 ((val)->is_neg ? -(val)->d : (val)->d)
#define str_val(val)                ((val)->s)
#define obj_val(val)                ((val)->p)
#define addr_val(val)               ((val)->a)

#define is_zero_val(val)                                                                 \
    (is_int_val(val) ? (val)->i == 0 : (is_fp_val(val) ? (val)->d == 0.0f : false))

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
        bool b;         /* boolean */
        uint64_t i;     /* integer */
        double d;       /* floating point */
        char *s;        /* string */
        void *p;        /* object pointer */
        uint32_t a;     /* memory address */
    };
};

bool value_check(value_t *val, meta_t *meta);

void value_eval(op_kind_t op, value_t *x, value_t *y, value_t *res);

void value_cast(value_t *val, meta_t *to);

int value_cmp(value_t *x, value_t *y);

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

static inline void
value_set_bool(value_t *val, bool b)
{
    val->kind = VAL_BOOL;
    val->b = b;
}

static inline void
value_set_int(value_t *val, uint64_t i)
{
    val->kind = VAL_INT;
    val->i = i;
}

static inline void
value_set_fp(value_t *val, double d)
{
    val->kind = VAL_FP;
    val->d = d;
}

static inline void
value_set_str(value_t *val, char *s)
{
    val->kind = VAL_STR;
    val->s = s;
}

static inline void
value_set_obj(value_t *val, void *p)
{
    val->kind = VAL_OBJ;
    val->p = p;
}

static inline void
value_set_addr(value_t *val, uint32_t a)
{
    val->kind = VAL_ADDR;
    val->a = a;
}

#endif /* ! _VALUE_H */
