/**
 * @file    value.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "meta.h"

#include "value.h"

#define ui32_fit_int(val, max)                                                           \
    (((val)->is_neg && (val)->ui32 <= (uint32_t)(max) + 1) ||                            \
     (!(val)->is_neg && (val)->ui32 <= (uint32_t)(max)))

#define ui32_fit_uint(val, max)      (!(val)->is_neg && (val)->ui32 <= (max))

#define ui64_fit_int(val, max)                                                           \
    (((val)->is_neg && (val)->ui64 <= (uint64_t)(max) + 1) ||                            \
     (!(val)->is_neg && (val)->ui64 <= (uint64_t)(max)))

#define ui64_fit_uint(val, max)      (!(val)->is_neg && (val)->ui64 <= (max))

#define value_eval_arith(op, x, y, res)                                                  \
    do {                                                                                 \
        switch (x->type) {                                                               \
        case TYPE_UINT32:                                                                \
            if (y->type == TYPE_UINT32) {                                                \
                value_set_ui32(res, ui32_val(x) op ui32_val(y));                         \
            }                                                                            \
            else {                                                                       \
                ASSERT1((y)->type == TYPE_UINT64, (y)->type);                            \
                value_set_ui64(res, ui32_val(x) op ui64_val(y));                         \
            }                                                                            \
            break;                                                                       \
        case TYPE_UINT64:                                                                \
            if (y->type == TYPE_UINT32) {                                                \
                value_set_ui64(res, ui64_val(x) op ui32_val(y));                         \
            }                                                                            \
            else {                                                                       \
                ASSERT1((y)->type == TYPE_UINT64, (y)->type);                            \
                value_set_ui64(res, ui64_val(x) op ui64_val(y));                         \
            }                                                                            \
            break;                                                                       \
        case TYPE_FLOAT:                                                                 \
            if (y->type == TYPE_FLOAT) {                                                 \
                value_set_f32(res, f32_val(x) op f32_val(y));                            \
            }                                                                            \
            else {                                                                       \
                ASSERT1((y)->type == TYPE_DOUBLE, (y)->type);                            \
                value_set_f64(res, f32_val(x) op f64_val(y));                            \
            }                                                                            \
            break;                                                                       \
        case TYPE_DOUBLE:                                                                \
            if (y->type == TYPE_FLOAT) {                                                 \
                value_set_f64(res, f64_val(x) op f32_val(y));                            \
            }                                                                            \
            else {                                                                       \
                ASSERT1((y)->type == TYPE_DOUBLE, (y)->type);                            \
                value_set_f64(res, f64_val(x) op f64_val(y));                            \
            }                                                                            \
            break;                                                                       \
        case TYPE_STRING:                                                                \
            ASSERT1((y)->type == TYPE_STRING, (y)->type);                                \
            value_set_str((res), xstrcat(str_val(x), str_val(y)));                       \
            break;                                                                       \
        default:                                                                         \
            ASSERT1(!"invalid value", (x)->type);                                        \
        }                                                                                \
    } while (0)

#define value_eval_cmp(op, x, y, res)                                                    \
    do {                                                                                 \
        bool v = false;                                                                  \
                                                                                         \
        switch (x->type) {                                                               \
        case TYPE_BOOL:                                                                  \
            ASSERT1((y)->type == TYPE_BOOL, (y)->type);                                  \
            v = bool_val(x) op bool_val(y);                                              \
            break;                                                                       \
        case TYPE_UINT32:                                                                \
            if (y->type == TYPE_UINT32) {                                                \
                v = ui32_val(x) op ui32_val(y);                                          \
            }                                                                            \
            else {                                                                       \
                ASSERT1((y)->type == TYPE_UINT64, (y)->type);                            \
                v = ui32_val(x) op ui64_val(y);                                          \
            }                                                                            \
            break;                                                                       \
        case TYPE_UINT64:                                                                \
            if (y->type == TYPE_UINT32) {                                                \
                v = ui64_val(x) op ui32_val(y);                                          \
            }                                                                            \
            else {                                                                       \
                ASSERT1((y)->type == TYPE_UINT64, (y)->type);                            \
                v = ui64_val(x) op ui64_val(y);                                          \
            }                                                                            \
            break;                                                                       \
        case TYPE_FLOAT:                                                                 \
            if (y->type == TYPE_FLOAT) {                                                 \
                v = f32_val(x) op f32_val(y);                                            \
            }                                                                            \
            else {                                                                       \
                ASSERT1((y)->type == TYPE_DOUBLE, (y)->type);                            \
                v = f32_val(x) op f64_val(y);                                            \
            }                                                                            \
            break;                                                                       \
        case TYPE_DOUBLE:                                                                \
            if (y->type == TYPE_FLOAT) {                                                 \
                v = f64_val(x) op f32_val(y);                                            \
            }                                                                            \
            else {                                                                       \
                ASSERT1((y)->type == TYPE_DOUBLE, (y)->type);                            \
                v = f64_val(x) op f64_val(y);                                            \
            }                                                                            \
            break;                                                                       \
        case TYPE_STRING:                                                                \
            if (is_null_val(x) || is_null_val(y)) {                                      \
                v = is_null_val(x) && is_null_val(y);                                    \
            }                                                                            \
            else {                                                                       \
                ASSERT1((y)->type == TYPE_STRING, (y)->type);                            \
                v = strcmp(str_val(x), str_val(y)) op 0;                                 \
            }                                                                            \
            break;                                                                       \
        case TYPE_OBJECT:                                                                \
            ASSERT1((y)->type == TYPE_OBJECT, (y)->type);                                \
            if (is_null_val(x) || is_null_val(y)) {                                      \
                v = is_null_val(x) && is_null_val(y);                                    \
            }                                                                            \
            else {                                                                       \
                v = obj_val(x) op obj_val(y);                                            \
            }                                                                            \
            break;                                                                       \
        default:                                                                         \
            ASSERT1(!"invalid value", (x)->type);                                        \
        }                                                                                \
                                                                                         \
        value_set_bool((res), v);                                                        \
    } while (0)

#define value_eval_bit(op, x, y, res)                                                    \
    do {                                                                                 \
        switch (x->type) {                                                               \
        case TYPE_UINT32:                                                                \
            if (y->type == TYPE_UINT32) {                                                \
                value_set_ui32(res, ui32_val(x) op ui32_val(y));                         \
            }                                                                            \
            else {                                                                       \
                ASSERT1((y)->type == TYPE_UINT64, (y)->type);                            \
                value_set_ui64(res, ui32_val(x) op ui64_val(y));                         \
            }                                                                            \
            break;                                                                       \
        case TYPE_UINT64:                                                                \
            if (y->type == TYPE_UINT32) {                                                \
                value_set_ui64(res, ui64_val(x) op ui32_val(y));                         \
            }                                                                            \
            else {                                                                       \
                ASSERT1((y)->type == TYPE_UINT64, (y)->type);                            \
                value_set_ui64(res, ui64_val(x) op ui64_val(y));                         \
            }                                                                            \
            break;                                                                       \
        default:                                                                         \
            ASSERT1(!"invalid value", (x)->type);                                        \
        }                                                                                \
    } while (0)

bool
value_fit(value_t *val, meta_t *meta)
{
    ASSERT1(!is_undef_type(meta), meta->type);

    switch (val->type) {
    case TYPE_BOOL:
        ASSERT1(is_bool_type(meta), meta->type);
        break;

    case TYPE_UINT32:
        ASSERT1(is_dec_family(meta), meta->type);
        if ((meta->type == TYPE_BYTE && ui32_fit_uint(val, UINT8_MAX)) ||
            (meta->type == TYPE_INT8 && ui32_fit_int(val, INT8_MAX)) ||
            (meta->type == TYPE_UINT8 && ui32_fit_uint(val, UINT8_MAX)) ||
            (meta->type == TYPE_INT16 && ui32_fit_int(val, INT16_MAX)) ||
            (meta->type == TYPE_UINT16 && ui32_fit_uint(val, UINT16_MAX)) ||
            (meta->type == TYPE_INT32 && ui32_fit_int(val, INT32_MAX)) ||
            (meta->type == TYPE_UINT32 && !val->is_neg))
            return true;

        if (meta->type == TYPE_INT64 || 
            (meta->type == TYPE_UINT64 && !val->is_neg)) {
            value_set_ui64(val, ui32_val(val));
            return true;
        }

        return false;

    case TYPE_UINT64:
        ASSERT1(is_dec_family(meta), meta->type);

        if ((meta->type == TYPE_BYTE && ui64_fit_uint(val, UINT8_MAX)) ||
            (meta->type == TYPE_INT8 && ui64_fit_int(val, INT8_MAX)) ||
            (meta->type == TYPE_UINT8 && ui64_fit_uint(val, UINT8_MAX)) ||
            (meta->type == TYPE_INT16 && ui64_fit_int(val, INT16_MAX)) ||
            (meta->type == TYPE_UINT16 && ui64_fit_uint(val, UINT16_MAX)) ||
            (meta->type == TYPE_INT32 && ui64_fit_int(val, INT32_MAX)) ||
            (meta->type == TYPE_UINT32 && ui64_fit_uint(val, UINT32_MAX))) {
            value_set_ui32(val, (uint32_t)ui64_val(val));
            return true;
        }

        if ((meta->type == TYPE_INT64 && ui64_fit_int(val, INT64_MAX)) ||
            (meta->type == TYPE_UINT64 && !val->is_neg))
            return true;

        return false;

    case TYPE_FLOAT:
        ASSERT1(is_fp_family(meta), meta->type);
        if (meta->type == TYPE_DOUBLE)
            value_set_f64(val, f32_val(val));
        break;

    case TYPE_DOUBLE:
        ASSERT1(is_fp_family(meta), meta->type);
        if (meta->type == TYPE_FLOAT) {
            if ((float)val->d > FLT_MAX || (float)val->d < -FLT_MAX)
                return false;

            value_set_f32(val, (float)f64_val(val));
        }
        break;

    case TYPE_STRING:
        ASSERT1(is_string_type(meta), meta->type);
        break;

    case TYPE_OBJECT:
        ASSERT1(is_obj_family(meta), meta->type);
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
        return bool_val(x) == bool_val(y) ? 0 : (bool_val(x) > bool_val(y) ? 1 : -1);

    case TYPE_UINT64:
        return ui64_val(x) == ui64_val(y) ? 0 : (ui64_val(x) > ui64_val(y) ? 1 : -1);

    case TYPE_DOUBLE:
        return f64_val(x) == f64_val(y) ? 0 : (f64_val(x) > f64_val(y) ? 1 : -1);

    case TYPE_STRING:
        return strcmp(str_val(x), str_val(y));

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
    if (is_ui64_val(x))
        ASSERT(y->ui64 != 0);
    else if (is_f64_val(x))
        ASSERT(y->d != 0.0f);

    value_eval_arith(/, x, y, res);
}

static void
value_mod(value_t *x, value_t *y, value_t *res)
{
    if (is_ui64_val(x)) {
        ASSERT(y->ui64 != 0);
        value_set_ui64(res, x->ui64 % y->ui64);
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

    if (is_ui32_val(x)) {
        value_set_ui32(res, ui32_val(x));
        value_set_neg(res, !x->is_neg);
    }
    else if (is_ui64_val(x)) {
        value_set_ui64(res, ui64_val(x));
        value_set_neg(res, !x->is_neg);
    }
    else if (is_f32_val(x)) {
        value_set_f32(res, -f32_val(x));
    }
    else if (is_f64_val(x)) {
        value_set_f64(res, -f64_val(x));
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
value_eval(op_kind_t op, value_t *x, value_t *y, value_t *res)
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

    case TYPE_UINT32:
        value_set_bool(val, val->ui32 != 0);
        value_set_neg(val, false);
        break;

    case TYPE_UINT64:
        value_set_bool(val, val->ui64 != 0);
        value_set_neg(val, false);
        break;

    case TYPE_FLOAT:
        value_set_bool(val, val->f != 0.0f);
        value_set_neg(val, false);
        break;

    case TYPE_DOUBLE:
        value_set_bool(val, val->d != 0.0f);
        value_set_neg(val, false);
        break;

    case TYPE_STRING:
        value_set_bool(val, str_val(val) != NULL && strcmp(str_val(val), "false") != 0);
        break;

    default:
        ASSERT1(!"invalid value", val->type);
    }
}

static void
value_cast_to_ui32(value_t *val)
{
    uint32_t ui32 = 0;

    switch (val->type) {
    case TYPE_BOOL:
        value_set_ui32(val, val->b ? 1 : 0);
        break;

    case TYPE_UINT32:
        break;

    case TYPE_UINT64:
        value_set_ui32(val, (uint32_t)val->ui64);
        break;

    case TYPE_FLOAT:
        value_set_ui32(val, (uint32_t)val->f);
        break;

    case TYPE_DOUBLE:
        value_set_ui32(val, (uint32_t)val->d);
        break;

    case TYPE_STRING:
        if (val->s != NULL) {
            if (val->s[0] == '-') {
                sscanf(val->s + 1, "%u", &ui32);
                value_set_neg(val, true);
            }
            else {
                sscanf(val->s, "%u", &ui32);
            }
        }
        value_set_ui32(val, ui32);
        break;

    default:
        ASSERT1(!"invalid value", val->type);
    }
}

static void
value_cast_to_ui64(value_t *val)
{
    uint64_t ui64 = 0;

    switch (val->type) {
    case TYPE_BOOL:
        value_set_ui64(val, val->b ? 1 : 0);
        break;

    case TYPE_UINT32:
        value_set_ui64(val, val->ui32);
        break;

    case TYPE_UINT64:
        break;

    case TYPE_FLOAT:
        value_set_ui64(val, (uint64_t)val->f);
        break;

    case TYPE_DOUBLE:
        value_set_ui64(val, (uint64_t)val->d);
        break;

    case TYPE_STRING:
        if (val->s != NULL) {
            if (val->s[0] == '-') {
                sscanf(val->s + 1, "%"SCNu64, &ui64);
                value_set_neg(val, true);
            }
            else {
                sscanf(val->s, "%"SCNu64, &ui64);
            }
        }
        value_set_ui64(val, ui64);
        break;

    default:
        ASSERT1(!"invalid value", val->type);
    }
}

static void
value_cast_to_f32(value_t *val)
{
    float f;

    switch (val->type) {
    case TYPE_BOOL:
        value_set_f32(val, bool_val(val) ? 1.0 : 0.0);
        break;

    case TYPE_UINT32:
        value_set_f32(val, (float)val->ui32);
        break;

    case TYPE_UINT64:
        value_set_f32(val, (float)val->ui64);
        break;

    case TYPE_FLOAT:
        break;

    case TYPE_DOUBLE:
        value_set_f32(val, val->d);
        break;

    case TYPE_STRING:
        sscanf(val->s, "%f", &f);
        value_set_f32(val, f);
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
        value_set_f64(val, bool_val(val) ? 1.0 : 0.0);
        break;

    case TYPE_UINT32:
        value_set_f64(val, (double)val->ui32);
        break;

    case TYPE_UINT64:
        value_set_f64(val, (double)val->ui64);
        break;

    case TYPE_FLOAT:
        value_set_f64(val, val->f);
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

    case TYPE_UINT32:
        snprintf(buf, sizeof(buf), "%u", ui32_val(val));
        value_set_str(val, xstrdup(buf));
        break;

    case TYPE_UINT64:
        snprintf(buf, sizeof(buf), "%"PRIu64, ui64_val(val));
        value_set_str(val, xstrdup(buf));
        break;

    case TYPE_FLOAT:
        snprintf(buf, sizeof(buf), "%f", f32_val(val));
        value_set_str(val, xstrdup(buf));
        break;

    case TYPE_DOUBLE:
        snprintf(buf, sizeof(buf), "%lf", f64_val(val));
        value_set_str(val, xstrdup(buf));
        break;

    case TYPE_STRING:
        break;

    default:
        ASSERT1(!"invalid value", val->type);
    }
}

void
value_cast(value_t *val, meta_t *meta)
{
    switch (meta->type) {
    case TYPE_BOOL:
        value_cast_to_bool(val);
        break;

    case TYPE_UINT32:
        value_cast_to_ui32(val);
        break;

    case TYPE_UINT64:
        value_cast_to_ui64(val);
        break;

    case TYPE_FLOAT:
        value_cast_to_f32(val);
        break;

    case TYPE_DOUBLE:
        value_cast_to_f64(val);
        break;

    case TYPE_STRING:
        value_cast_to_str(val);
        break;

    default:
        ASSERT1(!"invalid type", meta->type);
    }
}

/* end of value.c */
