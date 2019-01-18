/**
 * @file    value.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "meta.h"

#include "value.h"

#define i64_fit_signed(val, max)                                                         \
    (((val)->is_neg && (val)->i64 <= (uint64_t)(max) + 1) ||                             \
     (!(val)->is_neg && (val)->i64 <= (uint64_t)(max)))

#define i64_fit_unsigned(val, max)      (!(val)->is_neg && (val)->i64 <= (max))

#define value_eval_arith(op, x, y, res)                                                  \
    do {                                                                                 \
        ASSERT2((x)->type == (y)->type, (x)->type, (y)->type);                           \
        switch (x->type) {                                                               \
        case TYPE_UINT64:                                                                \
            value_set_i64((res), val_i64(x) op val_i64(y));                              \
            break;                                                                       \
        case TYPE_DOUBLE:                                                                \
            value_set_f64((res), val_f64(x) op val_f64(y));                              \
            break;                                                                       \
        case TYPE_STRING:                                                                \
            value_set_str((res), xstrcat(val_str(x), val_str(y)));                       \
            break;                                                                       \
        default:                                                                         \
            ASSERT1(!"invalid value", (x)->type);                                        \
        }                                                                                \
    } while (0)

#define value_eval_cmp(op, x, y, res)                                                    \
    do {                                                                                 \
        bool v = false;                                                                  \
        ASSERT2((x)->type == (y)->type, (x)->type, (y)->type);                           \
        switch (x->type) {                                                               \
        case TYPE_BOOL:                                                                  \
            v = val_bool(x) op val_bool(y);                                              \
            break;                                                                       \
        case TYPE_UINT64:                                                                \
            v = val_i64(x) op val_i64(y);                                                \
            break;                                                                       \
        case TYPE_DOUBLE:                                                                \
            v = val_f64(x) op val_f64(y);                                                \
            break;                                                                       \
        case TYPE_STRING:                                                                \
            if (is_null_val(x) || is_null_val(y))                                        \
                v = is_null_val(x) && is_null_val(y);                                    \
            else                                                                         \
                v = strcmp(val_str(x), val_str(y)) op 0;                                 \
            break;                                                                       \
        case TYPE_OBJECT:                                                                \
            ASSERT1((y)->type == TYPE_OBJECT, (y)->type);                                \
            if (is_null_val(x) || is_null_val(y))                                        \
                v = is_null_val(x) && is_null_val(y);                                    \
            else                                                                         \
                v = val_ptr(x) op val_ptr(y);                                            \
            break;                                                                       \
        default:                                                                         \
            ASSERT1(!"invalid value", (x)->type);                                        \
        }                                                                                \
        value_set_bool((res), v);                                                        \
    } while (0)

#define value_eval_bit(op, x, y, res)                                                    \
    do {                                                                                 \
        ASSERT2((x)->type == (y)->type, (x)->type, (y)->type);                           \
        switch (x->type) {                                                               \
        case TYPE_UINT64:                                                                \
            value_set_i64(res, val_i64(x) op val_i64(y));                                \
            break;                                                                       \
        default:                                                                         \
            ASSERT1(!"invalid value", (x)->type);                                        \
        }                                                                                \
    } while (0)

bool
value_fit(value_t *val, meta_t *meta)
{
    ASSERT1(!is_undef_meta(meta), meta->type);

    switch (val->type) {
    case TYPE_BOOL:
        ASSERT1(is_bool_meta(meta), meta->type);
        break;

    case TYPE_UINT64:
        ASSERT1(is_integer_meta(meta), meta->type);
        if ((meta->type == TYPE_BYTE && !i64_fit_unsigned(val, UINT8_MAX)) ||
            (meta->type == TYPE_INT8 && !i64_fit_signed(val, INT8_MAX)) ||
            (meta->type == TYPE_UINT8 && !i64_fit_unsigned(val, UINT8_MAX)) ||
            (meta->type == TYPE_INT16 && !i64_fit_signed(val, INT16_MAX)) ||
            (meta->type == TYPE_UINT16 && !i64_fit_unsigned(val, UINT16_MAX)) ||
            (meta->type == TYPE_INT32 && !i64_fit_signed(val, INT32_MAX)) ||
            (meta->type == TYPE_UINT32 && !i64_fit_unsigned(val, UINT32_MAX)) ||
            (meta->type == TYPE_INT64 && !i64_fit_signed(val, INT64_MAX)) ||
            (meta->type == TYPE_UINT64 && val->is_neg))
            return false;
        break;

    case TYPE_DOUBLE:
        ASSERT1(is_fpoint_meta(meta), meta->type);
        if (meta->type == TYPE_FLOAT && 
            ((float)val->d > FLT_MAX || (float)val->d < -FLT_MAX))
            return false;
        break;

    case TYPE_STRING:
        ASSERT1(is_string_meta(meta), meta->type);
        break;

    case TYPE_OBJECT:
        if (is_null_val(val))
            ASSERT1(is_pointer_meta(meta), meta->type);
        break;

    default:
        ASSERT1(!"invalid value", val->type);
    }

    return true;
}

int
value_cmp(value_t *x, value_t *y)
{
    ASSERT2(x->type == y->type, x->type, y->type);

    switch (x->type) {
    case TYPE_BOOL:
        return val_bool(x) == val_bool(y) ? 0 : (val_bool(x) > val_bool(y) ? 1 : -1);

    case TYPE_UINT64:
        return val_i64(x) == val_i64(y) ? 0 : (val_i64(x) > val_i64(y) ? 1 : -1);

    case TYPE_DOUBLE:
        return val_f64(x) == val_f64(y) ? 0 : (val_f64(x) > val_f64(y) ? 1 : -1);

    case TYPE_STRING:
        return strcmp(val_str(x), val_str(y));

    default:
        ASSERT1(!"invalid value", x->type);
    }

    return 0;
}

static void
value_add(value_t *x, value_t *y, value_t *res)
{
    value_eval_arith(+, x, y, res);
}

static void
value_sub(value_t *x, value_t *y, value_t *res)
{
    value_eval_arith(-, x, y, res);
}

static void
value_mul(value_t *x, value_t *y, value_t *res)
{
    value_eval_arith(*, x, y, res);
}

static void
value_div(value_t *x, value_t *y, value_t *res)
{
    if (is_i64_val(x))
        ASSERT(y->i64 != 0);
    else if (is_f64_val(x))
        ASSERT(y->d != 0.0f);

    value_eval_arith(/, x, y, res);
}

static void
value_mod(value_t *x, value_t *y, value_t *res)
{
    if (is_i64_val(x)) {
        ASSERT(y->i64 != 0);
        value_set_i64(res, x->i64 % y->i64);
    }
    else {
        ASSERT1(!"invalid value", res->type);
    }
}

static void
value_cmp_eq(value_t *x, value_t *y, value_t *res)
{
    value_eval_cmp(==, x, y, res);
}

static void
value_cmp_ne(value_t *x, value_t *y, value_t *res)
{
    value_eval_cmp(!=, x, y, res);
}

static void
value_cmp_lt(value_t *x, value_t *y, value_t *res)
{
    value_eval_cmp(<, x, y, res);
}

static void
value_cmp_gt(value_t *x, value_t *y, value_t *res)
{
    value_eval_cmp(>, x, y, res);
}

static void
value_cmp_le(value_t *x, value_t *y, value_t *res)
{
    value_eval_cmp(<=, x, y, res);
}

static void
value_cmp_ge(value_t *x, value_t *y, value_t *res)
{
    value_eval_cmp(>=, x, y, res);
}

static void
value_bit_and(value_t *x, value_t *y, value_t *res)
{
    value_eval_bit(&, x, y, res);
}

static void
value_bit_or(value_t *x, value_t *y, value_t *res)
{
    value_eval_bit(|, x, y, res);
}

static void
value_bit_xor(value_t *x, value_t *y, value_t *res)
{
    value_eval_bit(^, x, y, res);
}

static void
value_shift_l(value_t *x, value_t *y, value_t *res)
{
    value_eval_bit(<<, x, y, res);
}

static void
value_shift_r(value_t *x, value_t *y, value_t *res)
{
    value_eval_bit(>>, x, y, res);
}

static void
value_neg(value_t *x, value_t *y, value_t *res)
{
    ASSERT(y == NULL);

    if (is_i64_val(x)) {
        value_set_i64(res, val_i64(x));
        value_set_neg(res, !x->is_neg);
    }
    else if (is_f64_val(x)) {
        value_set_f64(res, -val_f64(x));
    }
    else {
        ASSERT1(!"invalid value", x->type);
    }
}

static void
value_not(value_t *x, value_t *y, value_t *res)
{
    ASSERT(y == NULL);

    if (is_bool_val(x))
        value_set_bool(res, !x->b);
    else
        ASSERT1(!"invalid value", x->type);
}

static void
value_and(value_t *x, value_t *y, value_t *res)
{
    value_eval_cmp(&&, x, y, res);
}

static void
value_or(value_t *x, value_t *y, value_t *res)
{
    value_eval_cmp(||, x, y, res);
}

eval_fn_t eval_fntab_[OP_CF_MAX + 1] = {
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
    value_not,
    value_and,
    value_or
};

void
value_eval(value_t *x, op_kind_t op, value_t *y, value_t *res)
{
    ASSERT1(op >= OP_ADD && op <= OP_CF_MAX, op);

    value_init(res);

    eval_fntab_[op](x, y, res);
}

static void
value_cast_to_bool(value_t *val)
{
    switch (val->type) {
    case TYPE_BOOL:
        break;

    case TYPE_UINT64:
        value_set_bool(val, val->i64 != 0);
        value_set_neg(val, false);
        break;

    case TYPE_DOUBLE:
        value_set_bool(val, val->d != 0.0f);
        value_set_neg(val, false);
        break;

    case TYPE_STRING:
        value_set_bool(val, val_str(val) != NULL && strcmp(val_str(val), "false") != 0);
        break;

    default:
        ASSERT1(!"invalid value", val->type);
    }
}

static void
value_cast_to_i64(value_t *val)
{
    uint64_t i64 = 0;

    switch (val->type) {
    case TYPE_BOOL:
        value_set_i64(val, val->b ? 1 : 0);
        break;

    case TYPE_UINT64:
        break;

    case TYPE_DOUBLE:
        value_set_i64(val, (uint64_t)val->d);
        break;

    case TYPE_STRING:
        if (val->s != NULL) {
            if (val->s[0] == '-') {
                sscanf(val->s + 1, "%"SCNu64, &i64);
                value_set_neg(val, true);
            }
            else {
                sscanf(val->s, "%"SCNu64, &i64);
            }
        }
        value_set_i64(val, i64);
        break;

    default:
        ASSERT1(!"invalid value", val->type);
    }
}

static void
value_cast_to_f64(value_t *val)
{
    double d;

    switch (val->type) {
    case TYPE_BOOL:
        value_set_f64(val, val_bool(val) ? 1.0 : 0.0);
        break;

    case TYPE_UINT64:
        value_set_f64(val, (double)val->i64);
        break;

    case TYPE_DOUBLE:
        break;

    case TYPE_STRING:
        sscanf(val->s, "%lf", &d);
        value_set_f64(val, d);
        break;

    default:
        ASSERT1(!"invalid value", val->type);
    }
}

static void
value_cast_to_str(value_t *val)
{
    char buf[256];

    switch (val->type) {
    case TYPE_BOOL:
        value_set_str(val, val->b ? xstrdup("true") : xstrdup("false"));
        break;

    case TYPE_UINT64:
        snprintf(buf, sizeof(buf), "%"PRIu64, val_i64(val));
        value_set_str(val, xstrdup(buf));
        break;

    case TYPE_DOUBLE:
        snprintf(buf, sizeof(buf), "%lf", val_f64(val));
        value_set_str(val, xstrdup(buf));
        break;

    case TYPE_STRING:
        break;

    default:
        ASSERT1(!"invalid value", val->type);
    }
}

cast_fn_t cast_fntab_[TYPE_COMPATIBLE + 1] = {
    NULL,
    value_cast_to_bool,
    value_cast_to_i64,
    value_cast_to_i64,
    value_cast_to_i64,
    value_cast_to_i64,
    value_cast_to_i64,
    value_cast_to_i64,
    value_cast_to_i64,
    value_cast_to_i64,
    value_cast_to_i64,
    value_cast_to_f64,
    value_cast_to_f64,
    value_cast_to_str
};

void
value_cast(value_t *val, type_t type)
{
    ASSERT1(type > TYPE_NONE && type <= TYPE_STRING, type);

    cast_fntab_[type](val);
}

/* end of value.c */
