/**
 * @file    value.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "meta.h"

#include "value.h"

#define value_eval_arith(op, val, x, y)                                        \
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

#define value_eval_cmp(op, val, x, y)                                          \
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
        else if (is_str_val(x))                                                \
            res = strcmp((x)->sv, (y)->sv) op 0;                               \
        else                                                                   \
            ASSERT1(!"invalid value", (val)->kind);                            \
                                                                               \
        value_set_bool((val), res);                                            \
    } while (0)

#define value_eval_bit(op, val, x, y)                                          \
    do {                                                                       \
        ASSERT2((x)->kind == (y)->kind, (x)->kind, (y)->kind);                 \
                                                                               \
        if (is_int_val(x))                                                     \
            value_set_int((val), (x)->iv op (y)->iv);                          \
        else                                                                   \
            ASSERT1(!"invalid value", (val)->kind);                            \
    } while (0)

static void
value_add(value_t *val, value_t *x, value_t *y)
{
    if (is_str_val(x))
        value_set_str(val, xstrcat(x->sv, y->sv));
    else
        value_eval_arith(+, val, x, y);
}

static void
value_sub(value_t *val, value_t *x, value_t *y)
{
    value_eval_arith(-, val, x, y);
}

static void
value_mul(value_t *val, value_t *x, value_t *y)
{
    value_eval_arith(*, val, x, y);
}

static void
value_div(value_t *val, value_t *x, value_t *y)
{
    if (is_int_val(x))
        ASSERT(y->iv != 0);
    else if (is_fp_val(x))
        ASSERT(y->dv != 0.0f);

    value_eval_arith(/, val, x, y);
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

static void
value_cmp_eq(value_t *val, value_t *x, value_t *y)
{
    value_eval_cmp(==, val, x, y);
}

static void
value_cmp_ne(value_t *val, value_t *x, value_t *y)
{
    value_eval_cmp(!=, val, x, y);
}

static void
value_cmp_lt(value_t *val, value_t *x, value_t *y)
{
    value_eval_cmp(<, val, x, y);
}

static void
value_cmp_gt(value_t *val, value_t *x, value_t *y)
{
    value_eval_cmp(>, val, x, y);
}

static void
value_cmp_le(value_t *val, value_t *x, value_t *y)
{
    value_eval_cmp(<=, val, x, y);
}

static void
value_cmp_ge(value_t *val, value_t *x, value_t *y)
{
    value_eval_cmp(>=, val, x, y);
}

static void
value_bit_and(value_t *val, value_t *x, value_t *y)
{
    value_eval_bit(&, val, x, y);
}

static void
value_bit_or(value_t *val, value_t *x, value_t *y)
{
    value_eval_bit(|, val, x, y);
}

static void
value_bit_xor(value_t *val, value_t *x, value_t *y)
{
    value_eval_bit(^, val, x, y);
}

static void
value_shift_l(value_t *val, value_t *x, value_t *y)
{
    value_eval_bit(<<, val, x, y);
}

static void
value_shift_r(value_t *val, value_t *x, value_t *y)
{
    ASSERT2((x)->kind == (y)->kind, (x)->kind, (y)->kind);             \
    value_eval_bit(>>, val, x, y);
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

int
value_cmp(value_t *x, value_t *y)
{
    ASSERT2(x->kind == y->kind, x->kind, y->kind);

    switch (x->kind) {
    case VAL_BOOL:
        return x->bv == y->bv ? 0 : (x->bv > y->bv ? 1 : -1);

    case VAL_INT:
        return x->iv == y->iv ? 0 : (x->iv > y->iv ? 1 : -1);

    case VAL_FP:
        return x->dv == y->dv ? 0 : (x->dv > y->dv ? 1 : -1);

    case VAL_STR:
        return strcmp(x->sv, y->sv);

    default:
        ASSERT1(!"invalid value", x->kind);
    }

    return 0;
}

/* end of value.c */
