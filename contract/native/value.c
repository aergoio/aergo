/**
 * @file    value.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "meta.h"

#include "value.h"

#define value_arith(op, val, x, y)                                             \
    do {                                                                       \
        ASSERT2((x)->kind == (y)->kind, (x)->kind, (y)->kind);                 \
                                                                               \
        if (is_int_val(x))                                                     \
            value_set_int(val, (x)->iv op (y)->iv);                            \
        else if (is_fp_val(x))                                                 \
            value_set_double(val, (x)->dv op (y)->dv);                         \
        else                                                                   \
            ASSERT1(!"invalid value", (val)->kind);                            \
    } while (0)

static void
value_add(value_t *val, value_t *x, value_t *y)
{
    if (is_str_val(x))
        value_set_str(val, xstrcat(x->sv, y->sv));
    else
        value_arith(+, val, x, y);
}

static void
value_sub(value_t *val, value_t *x, value_t *y)
{
    value_arith(-, val, x, y);
}

static void
value_mul(value_t *val, value_t *x, value_t *y)
{
    value_arith(*, val, x, y);
}

static void
value_div(value_t *val, value_t *x, value_t *y)
{
    if (is_int_val(x))
        ASSERT(y->iv != 0);
    else if (is_fp_val(x))
        ASSERT(y->dv != 0.0f);

    value_arith(/, val, x, y);
}

static void
value_mod(value_t *val, value_t *x, value_t *y)
{
    if (is_int_val(x)) {
        ASSERT(y->iv != 0);
        value_set_int(val, x->iv % y->iv);
    }
    else {
        ASSERT1(!"invalid value", val->kind);
    }
}

#define value_cmp(op, val, x, y)                                               \
    do {                                                                       \
        bool res = false;                                                      \
                                                                               \
        ASSERT2((x)->kind == (y)->kind, (x)->kind, (y)->kind);                 \
                                                                               \
        if (is_null_val(x))                                                    \
            res = NULL op NULL;                                                \
        else if (is_bool_val(x))                                               \
            res = (x)->bv op (y)->bv;                                          \
        else if (is_int_val(x))                                                \
            res = (x)->iv op (y)->iv;                                          \
        else if (is_fp_val(x))                                                 \
            res = (x)->iv op (y)->iv;                                          \
        else if (is_str_val(x))                                             \
            res = strcmp((x)->sv, (y)->sv) op 0;                               \
        else                                                                   \
            ASSERT1(!"invalid value", (val)->kind);                            \
                                                                               \
        value_set_bool((val), res);                                            \
    } while (0)

static void
value_cmp_eq(value_t *val, value_t *x, value_t *y)
{
    value_cmp(==, val, x, y);
}

static void
value_cmp_ne(value_t *val, value_t *x, value_t *y)
{
    value_cmp(!=, val, x, y);
}

static void
value_cmp_lt(value_t *val, value_t *x, value_t *y)
{
    value_cmp(<, val, x, y);
}

static void
value_cmp_gt(value_t *val, value_t *x, value_t *y)
{
    value_cmp(>, val, x, y);
}

static void
value_cmp_le(value_t *val, value_t *x, value_t *y)
{
    value_cmp(<=, val, x, y);
}

static void
value_cmp_ge(value_t *val, value_t *x, value_t *y)
{
    value_cmp(>=, val, x, y);
}

#define value_bit_op(op, val, x, y)                                            \
    do {                                                                       \
        ASSERT2((x)->kind == (y)->kind, (x)->kind, (y)->kind);                 \
                                                                               \
        if (is_int_val(x))                                                     \
            value_set_int((val), (x)->iv op (y)->iv);                          \
        else                                                                   \
            ASSERT1(!"invalid value", (val)->kind);                            \
    } while (0)

static void
value_bit_and(value_t *val, value_t *x, value_t *y)
{
    value_bit_op(&, val, x, y);
}

static void
value_bit_or(value_t *val, value_t *x, value_t *y)
{
    value_bit_op(|, val, x, y);
}

static void
value_bit_xor(value_t *val, value_t *x, value_t *y)
{
    value_bit_op(^, val, x, y);
}

static void
value_shift_l(value_t *val, value_t *x, value_t *y)
{
    value_bit_op(<<, val, x, y);
}

static void
value_shift_r(value_t *val, value_t *x, value_t *y)
{
    ASSERT2((x)->kind == (y)->kind, (x)->kind, (y)->kind);             \
    value_bit_op(>>, val, x, y);
}

static void
value_neg(value_t *val, value_t *x, value_t *y)
{
    if (is_int_val(x))
        value_set_int(val, -x->iv);
    else if (is_fp_val(x))
        value_set_double(val, -x->dv);
    else
        ASSERT1(!"invalid value", val->kind);
}

static void
value_not(value_t *val, value_t *x, value_t *y)
{
    if (is_bool_val(x))
        value_set_bool(val, !x->bv);
    else
        ASSERT1(!"invalid value", val->kind);
}

eval_fn_t eval_fntab_[OP_CF_MAX] = {
    value_add,
    value_sub,
    value_mul,
    value_div,
    value_mod,
    value_cmp_eq,
    value_cmp_ne,
    value_cmp_lt,
    value_cmp_gt,
    value_cmp_le,
    value_cmp_ge,
    value_bit_and,
    value_bit_or,
    value_bit_xor,
    value_shift_l,
    value_shift_r,
    value_neg,
    value_not
};

/* end of value.c */
