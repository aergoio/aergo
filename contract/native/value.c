/**
 * @file    value.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "meta.h"

#include "value.h"

mpz_t int128_min_ = { { 0, 0, NULL } };
mpz_t int128_max_ = { { 0, 0, NULL } };
mpz_t uint128_max_ = { { 0, 0, NULL } };

#define mpz_fits_schar_p(v)     (mpz_get_si(v) >= INT8_MIN && mpz_get_si(v) <= INT8_MAX)
#define mpz_fits_uchar_p(v)     (mpz_get_si(v) >= 0 && mpz_get_si(v) <= UINT8_MAX)

#define mpz_fits_int128_p(v)    (mpz_cmp(v, int128_min_) >= 0 && mpz_cmp(v, int128_max_) <= 0)
#define mpz_fits_uint128_p(v)   (mpz_sgn(v) >= 0 && mpz_cmp(v, uint128_max_) <= 0)

#define value_eval_cmp(op, x, y, res)                                                              \
    do {                                                                                           \
        bool v = false;                                                                            \
        ASSERT2((x)->type == (y)->type, (x)->type, (y)->type);                                     \
        switch (x->type) {                                                                         \
        case TYPE_BOOL:                                                                            \
            v = val_bool(x) op val_bool(y);                                                        \
            break;                                                                                 \
        case TYPE_UINT128:                                                                         \
            v = mpz_cmp(val_mpz(x), val_mpz(y)) op 0;                                              \
            break;                                                                                 \
        case TYPE_DOUBLE:                                                                          \
            v = val_f64(x) op val_f64(y);                                                          \
            break;                                                                                 \
        case TYPE_STRING:                                                                          \
            if (is_null_val(x) || is_null_val(y))                                                  \
                v = is_null_val(x) && is_null_val(y);                                              \
            else                                                                                   \
                v = strcmp(val_str(x), val_str(y)) op 0;                                           \
            break;                                                                                 \
        case TYPE_OBJECT:                                                                          \
            ASSERT1((y)->type == TYPE_OBJECT, (y)->type);                                          \
            if (is_null_val(x) || is_null_val(y))                                                  \
                v = is_null_val(x) && is_null_val(y);                                              \
            else                                                                                   \
                v = val_ptr(x) op val_ptr(y);                                                      \
            break;                                                                                 \
        default:                                                                                   \
            ASSERT1(!"invalid value", (x)->type);                                                  \
        }                                                                                          \
        value_set_bool((res), v);                                                                  \
    } while (0)

bool
value_fit(value_t *val, meta_t *meta)
{
    ASSERT1(!is_undef_meta(meta), meta->type);

    switch (val->type) {
    case TYPE_BOOL:
        ASSERT1(is_bool_meta(meta), meta->type);
        break;

    case TYPE_UINT128:
        ASSERT1(is_integer_meta(meta), meta->type);
        if (mpz_size(int128_min_) == 0) {
            mpz_init_set_str(int128_min_, "170141183460469231731687303715884105728", 10);
            mpz_neg(int128_min_, int128_min_);
        }

        if (mpz_size(int128_max_) == 0)
            mpz_init_set_str(int128_max_, "170141183460469231731687303715884105727", 10);

        if (mpz_size(uint128_max_) == 0)
            mpz_init_set_str(uint128_max_, "340282366920938463463374607431768211455", 10);

        if ((meta->type == TYPE_INT8 && !mpz_fits_schar_p(val_mpz(val))) ||
            (meta->type == TYPE_UINT8 && !mpz_fits_uchar_p(val_mpz(val))) ||
            (meta->type == TYPE_INT16 && !mpz_fits_sshort_p(val_mpz(val))) ||
            (meta->type == TYPE_UINT16 && !mpz_fits_ushort_p(val_mpz(val))) ||
            (meta->type == TYPE_INT32 && !mpz_fits_sint_p(val_mpz(val))) ||
            (meta->type == TYPE_UINT32 && !mpz_fits_uint_p(val_mpz(val))) ||
            (meta->type == TYPE_INT64 && !mpz_fits_slong_p(val_mpz(val))) ||
            (meta->type == TYPE_UINT64 && !mpz_fits_ulong_p(val_mpz(val))) ||
            (meta->type == TYPE_INT128 && !mpz_fits_int128_p(val_mpz(val))) ||
            (meta->type == TYPE_UINT128 && !mpz_fits_uint128_p(val_mpz(val))))
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

    case TYPE_UINT128:
        return mpz_cmp(val_mpz(x), val_mpz(y));

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
    ASSERT2(x->type == y->type, x->type, y->type);

    switch (x->type) {
    case TYPE_UINT128:
        value_init_int(res);
        mpz_add(val_mpz(res), val_mpz(x), val_mpz(y));
        break;

    case TYPE_DOUBLE:
        value_set_f64(res, val_f64(x) + val_f64(y));
        break;

    case TYPE_STRING:
        if (val_str(x) != NULL && val_str(y) != NULL)
            value_set_str(res, xstrcat(val_str(x), val_str(y)));
        break;

    default:
        ASSERT1(!"invalid value", x->type);
    }
}

static void
value_sub(value_t *x, value_t *y, value_t *res)
{
    ASSERT2(x->type == y->type, x->type, y->type);

    switch (x->type) {
    case TYPE_UINT128:
        value_init_int(res);
        mpz_sub(val_mpz(res), val_mpz(x), val_mpz(y));
        break;

    case TYPE_DOUBLE:
        value_set_f64(res, val_f64(x) - val_f64(y));
        break;

    default:
        ASSERT1(!"invalid value", x->type);
    }
}

static void
value_mul(value_t *x, value_t *y, value_t *res)
{
    ASSERT2(x->type == y->type, x->type, y->type);

    switch (x->type) {
    case TYPE_UINT128:
        value_init_int(res);
        mpz_mul(val_mpz(res), val_mpz(x), val_mpz(y));
        break;

    case TYPE_DOUBLE:
        value_set_f64(res, val_f64(x) * val_f64(y));
        break;

    default:
        ASSERT1(!"invalid value", x->type);
    }
}

static void
value_div(value_t *x, value_t *y, value_t *res)
{
    if (is_int_val(x))
        ASSERT(mpz_sgn(val_mpz(y)) != 0);
    else if (is_f64_val(x))
        ASSERT(y->d != 0.0f);

    ASSERT2(x->type == y->type, x->type, y->type);

    switch (x->type) {
    case TYPE_UINT128:
        value_init_int(res);
        mpz_tdiv_q(val_mpz(res), val_mpz(x), val_mpz(y));
        break;

    case TYPE_DOUBLE:
        value_set_f64(res, val_f64(x) / val_f64(y));
        break;

    default:
        ASSERT1(!"invalid value", x->type);
    }
}

static void
value_mod(value_t *x, value_t *y, value_t *res)
{
    if (is_int_val(x)) {
        ASSERT(mpz_sgn(val_mpz(y)) != 0);
        value_init_int(res);
        mpz_mod(val_mpz(res), val_mpz(x), val_mpz(y));
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
    ASSERT2(x->type == y->type, x->type, y->type);

    switch (x->type) {
    case TYPE_UINT128:
        value_init_int(res);
        mpz_and(val_mpz(res), val_mpz(x), val_mpz(y));
        break;
    default:
        ASSERT1(!"invalid value", x->type);
    }
}

static void
value_bit_or(value_t *x, value_t *y, value_t *res)
{
    ASSERT2(x->type == y->type, x->type, y->type);

    switch (x->type) {
    case TYPE_UINT128:
        value_init_int(res);
        mpz_ior(val_mpz(res), val_mpz(x), val_mpz(y));
        break;
    default:
        ASSERT1(!"invalid value", x->type);
    }
}

static void
value_bit_xor(value_t *x, value_t *y, value_t *res)
{
    ASSERT2(x->type == y->type, x->type, y->type);

    switch (x->type) {
    case TYPE_UINT128:
        value_init_int(res);
        mpz_xor(val_mpz(res), val_mpz(x), val_mpz(y));
        break;
    default:
        ASSERT1(!"invalid value", x->type);
    }
}

static void
value_shift_l(value_t *x, value_t *y, value_t *res)
{
    ASSERT2(x->type == y->type, x->type, y->type);

    if (is_int_val(x)) {
        value_init_int(res);
        mpz_mul_2exp(val_mpz(res), val_mpz(x), mpz_get_ui(val_mpz(y)));
    }
    else {
        ASSERT1(!"invalid value", x->type);
    }
}

static void
value_shift_r(value_t *x, value_t *y, value_t *res)
{
    ASSERT2(x->type == y->type, x->type, y->type);

    if (is_int_val(x)) {
        value_init_int(res);
        mpz_tdiv_q_2exp(val_mpz(res), val_mpz(x), mpz_get_ui(val_mpz(y)));
    }
    else {
        ASSERT1(!"invalid value", x->type);
    }
}

static void
value_neg(value_t *x, value_t *y, value_t *res)
{
    ASSERT(y == NULL);

    if (is_int_val(x)) {
        value_init_int(res);
        mpz_neg(val_mpz(res), val_mpz(x));
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
    ASSERT2(x->type == y->type, x->type, y->type);

    if (is_bool_val(x))
        value_set_bool(res, val_bool(x) && val_bool(y));
    else
        ASSERT1(!"invalid value", x->type);
}

static void
value_or(value_t *x, value_t *y, value_t *res)
{
    ASSERT2(x->type == y->type, x->type, y->type);

    if (is_bool_val(x))
        value_set_bool(res, val_bool(x) || val_bool(y));
    else
        ASSERT1(!"invalid value", x->type);
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
    value_shift_r,
    value_shift_l,
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

    case TYPE_UINT128:
        value_set_bool(val, mpz_sgn(val->z) != 0);
        break;

    case TYPE_DOUBLE:
        value_set_bool(val, val->d != 0.0f);
        break;

    case TYPE_STRING:
        value_set_bool(val, val_str(val) != NULL && strcmp(val_str(val), "false") != 0);
        break;

    default:
        ASSERT1(!"invalid value", val->type);
    }
}

static void
value_cast_to_i128(value_t *val)
{
    bool b;
    double d;
    char *s;

    switch (val->type) {
    case TYPE_BOOL:
        b = val_bool(val);
        value_set_int(val, b ? 1 : 0);
        break;

    case TYPE_UINT128:
        break;

    case TYPE_DOUBLE:
        d = val_f64(val);
        value_init_int(val);
        mpz_set_d(val_mpz(val), d);
        break;

    case TYPE_STRING:
        s = val_ptr(val);
        value_init_int(val);
        if (s != NULL)
            mpz_set_str(val_mpz(val), s, 0);
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

    case TYPE_UINT128:
        value_set_f64(val, mpz_get_d(val->z));
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

    case TYPE_UINT128:
        value_set_str(val, mpz_get_str(NULL, 10, val_mpz(val)));
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
    value_cast_to_i128,
    value_cast_to_i128,
    value_cast_to_i128,
    value_cast_to_i128,
    value_cast_to_i128,
    value_cast_to_i128,
    NULL,
    NULL,
    NULL,
    NULL,
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
