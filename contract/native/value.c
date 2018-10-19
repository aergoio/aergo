/**
 * @file    value.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "meta.h"

#include "value.h"

#define value_check_int(val, max)                                                        \
    (((val)->is_neg && (val)->iv > (uint64_t)(max) + 1) ||                               \
     (!(val)->is_neg && (val)->iv > (uint64_t)(max)))
#define value_check_uint(val, max)                                                       \
    ((val)->is_neg || (val)->iv > (max))

#define value_eval_arith(op, val, x, y)                                                  \
    do {                                                                                 \
        ASSERT2((x)->kind == (y)->kind, (x)->kind, (y)->kind);                           \
                                                                                         \
        if (is_int_val(x))                                                               \
            value_set_int(val, int_val(x) op int_val(y));                                \
        else if (is_fp_val(x))                                                           \
            value_set_double(val, fp_val(x) op fp_val(y));                               \
        else if (is_str_val(x))                                                          \
            value_set_str((val), xstrcat(str_val(x), str_val(y)));                       \
        else                                                                             \
            ASSERT1(!"invalid value", (val)->kind);                                      \
    } while (0)

#define value_eval_cmp(op, val, x, y)                                                    \
    do {                                                                                 \
        bool res = false;                                                                \
                                                                                         \
        ASSERT2((x)->kind == (y)->kind, (x)->kind, (y)->kind);                           \
                                                                                         \
        if (is_null_val(x))                                                              \
            res = NULL op NULL;                                                          \
        else if (is_bool_val(x))                                                         \
            res = bool_val(x) op bool_val(y);                                            \
        else if (is_int_val(x))                                                          \
            res = int_val(x) op int_val(y);                                              \
        else if (is_fp_val(x))                                                           \
            res = fp_val(x) op fp_val(y);                                                \
        else if (is_str_val(x))                                                          \
            res = strcmp(str_val(x), str_val(y)) op 0;                                   \
        else                                                                             \
            ASSERT1(!"invalid value", (val)->kind);                                      \
                                                                                         \
        value_set_bool((val), res);                                                      \
    } while (0)

#define value_eval_bit(op, val, x, y)                                                    \
    do {                                                                                 \
        ASSERT2((x)->kind == (y)->kind, (x)->kind, (y)->kind);                           \
                                                                                         \
        if (is_int_val(x))                                                               \
            value_set_int((val), int_val(x) op int_val(y));                              \
        else                                                                             \
            ASSERT1(!"invalid value", (val)->kind);                                      \
    } while (0)

bool
value_check(value_t *val, meta_t *meta)
{
    switch (val->kind) {
    case VAL_NULL:
        ASSERT1(is_map_meta(meta) || is_object_meta(meta), meta->type);
        break;

    case VAL_BOOL:
        ASSERT1(is_bool_meta(meta), meta->type);
        break;

    case VAL_INT:
        ASSERT1(is_integer_meta(meta), meta->type);
        if ((meta->type == TYPE_BYTE && value_check_uint(val, UINT8_MAX)) ||
            (meta->type == TYPE_INT8 && value_check_int(val, INT8_MAX)) ||
            (meta->type == TYPE_UINT8 && value_check_uint(val, UINT8_MAX)) ||
            (meta->type == TYPE_INT16 && value_check_int(val, INT16_MAX)) ||
            (meta->type == TYPE_UINT16 && value_check_uint(val, UINT16_MAX)) ||
            (meta->type == TYPE_INT32 && value_check_int(val, INT32_MAX)) ||
            (meta->type == TYPE_UINT32 && value_check_uint(val, UINT32_MAX)) ||
            (meta->type == TYPE_INT64 && value_check_int(val, INT64_MAX)) ||
            (meta->type == TYPE_UINT64 && val->is_neg))
            return false;
        break;
    
    case VAL_FP:
        ASSERT1(is_float_meta(meta), meta->type);
        if (meta->type == TYPE_FLOAT && val->dv > FLT_MAX)
            return false;
        break;

    case VAL_STR:
        ASSERT1(is_string_meta(meta), meta->type);
        break;

    default:
        ASSERT1(!"invalid value", val->kind);
    }

    return true;
}

static void
value_add(value_t *val, value_t *x, value_t *y)
{
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
        value_set_int(val, int_val(x));
    else if (is_fp_val(x))
        value_set_double(val, fp_val(x));
    else
        ASSERT1(!"invalid value", val->kind);

    value_set_neg(val, !x->is_neg);
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
        return bool_val(x) == bool_val(y) ? 0 : (bool_val(x) > bool_val(y) ? 1 : -1);

    case VAL_INT:
        return int_val(x) == int_val(y) ? 0 : (int_val(x) > int_val(y) ? 1 : -1);

    case VAL_FP:
        return fp_val(x) == fp_val(y) ? 0 : (fp_val(x) > fp_val(y) ? 1 : -1);

    case VAL_STR:
        return strcmp(str_val(x), str_val(y));

    default:
        ASSERT1(!"invalid value", x->kind);
    }

    return 0;
}

/* end of value.c */
