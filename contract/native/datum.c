/**
 * @file    datum.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "strbuf.h"
#include "ast_id.h"
#include "ast_exp.h"

#include "datum.h"

char *
datum_to_str(datum_t *dat)
{
    int i;
    strbuf_t buf;

    strbuf_init(&buf);

    if (is_struct_type(dat)) {
        ASSERT(dat->name != NULL);

        strbuf_cat(&buf, "struct ");
        strbuf_cat(&buf, dat->name);
    }
    else if (is_map_type(dat)) {
        strbuf_cat(&buf, "map(");
        strbuf_cat(&buf, datum_to_str(dat->map.k));
        strbuf_cat(&buf, ", ");
        strbuf_cat(&buf, datum_to_str(dat->map.v));
        strbuf_cat(&buf, ")");
    }
    else if (is_tuple_type(dat)) {
        int i;

        strbuf_cat(&buf, "{");
        for (i = 0; i < dat->tup.cnt; i++) {
            if (i > 0)
                strbuf_cat(&buf, ", ");

            strbuf_cat(&buf, datum_to_str(dat->tup.evs[i]));
        }
        strbuf_cat(&buf, "}");
    }
    else {
        strbuf_cat(&buf, TYPE_NAME(dat->type));
    }

    for (i = 0; i < dat->arr_dim; i++) {
        strbuf_cat(&buf, "[]");
    }

    return strbuf_str(&buf);
}

void
datum_set_array(datum_t *dat, int arr_dim)
{
    ASSERT1(arr_dim >= 0, arr_dim);

    dat->arr_dim = arr_dim;
    dat->arr_size = xcalloc(sizeof(int) * arr_dim);
}

void
datum_set_map(datum_t *dat, datum_t *k, datum_t *v)
{
    datum_set_type(dat, TYPE_MAP);

    dat->elem_cnt = 2;
    dat->elems = xmalloc(sizeof(datum_t *) * 2);

    dat->elems[0] = k;
    dat->elems[1] = v;
}

void
datum_set_struct(datum_t *dat, char *name, array_t *ids)
{
    int i;

    datum_set_type(dat, TYPE_STRUCT);

    dat->name = name;

    dat->elem_cnt = array_size(ids);
    dat->elems = xmalloc(sizeof(datum_t *) * dat->elem_cnt);

    for (i = 0; i < dat->elem_cnt; i++) {
        dat->elems[i] = &array_item(ids, i, ast_id_t)->dat;
    }
}

void
datum_set_tuple(datum_t *dat, array_t *exps)
{
    int i;

    datum_set_type(dat, TYPE_TUPLE);

    dat->elem_cnt = array_size(exps);
    dat->elems = xmalloc(sizeof(datum_t *) * dat->elem_cnt);

    for (i = 0; i < dat->elem_cnt; i++) {
        dat->elems[i] = &array_item(exps, i, ast_exp_t)->dat;
    }
}

static int
datum_assign_map(datum_t *x, datum_t *y)
{
    datum_t *k = x->elems[0];
    datum_t *v = x->elems[1];

    if (is_object_datum(y))
        return NO_ERROR;

    if (is_map_datum(y)) {
        CHECK(datum_assign(k_datum, y->u_map.k_datum));
        CHECK(datum_assign(v_datum, y->u_map.v_datum));
    }
    else if (is_tuple_datum(y)) {
        int i;
        array_t *kv_elems = y->u_tup.datums;

        for (i = 0; i < array_size(kv_elems); i++) {
            datum_t *kv_elem = array_item(kv_elems, i, datum_t);
            array_t *kv_datums = kv_elem->u_tup.datums;

            if (!is_tuple_datum(kv_elem))
                RETURN(ERROR_MISMATCHED_TYPE, y->pos, datum_to_str(x),
                       datum_to_str(kv_elem));

            if (array_size(kv_datums) != 2)
                RETURN(ERROR_MISMATCHED_COUNT, y->pos, "key-value", 2,
                       array_size(kv_datums));

            CHECK(datum_assign(k_datum, array_item(kv_datums, 0, datum_t)));
            CHECK(datum_assign(v_datum, array_item(kv_datums, 1, datum_t)));
        }
    }
    else {
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, datum_to_str(x), datum_to_str(y));
    }

    return NO_ERROR;
}

static int
datum_assign_tuple(datum_t *x, datum_t *y)
{
    int i;
    array_t *x_elems = x->u_tup.datums;
    array_t *y_elems = y->u_tup.datums;

    if (!is_tuple_datum(y))
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, datum_to_str(x), datum_to_str(y));

    if (array_size(x_elems) != array_size(y_elems))
        RETURN(ERROR_MISMATCHED_COUNT, y->pos, "element", array_size(x_elems),
               array_size(y_elems));

    for (i = 0; i < array_size(x_elems); i++) {
        datum_t *x_elem = array_item(x_elems, i, datum_t);
        datum_t *y_elem = array_item(y_elems, i, datum_t);

        CHECK(datum_assign(x_elem, y_elem));
    }

    return NO_ERROR;
}

static int
datum_assign_struct(datum_t *x, datum_t *y)
{
    if (is_struct_datum(y)) {
        ASSERT(x->u_tup.name != NULL);
        ASSERT(y->u_tup.name != NULL);

        if (strcmp(x->u_tup.name, y->u_tup.name) != 0)
            RETURN(ERROR_MISMATCHED_TYPE, y->pos, datum_to_str(x), datum_to_str(y));
    }
    else if (is_tuple_datum(y)) {
        return datum_assign_tuple(x, y);
    }
    else {
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, datum_to_str(x), datum_to_str(y));
    }

    return NO_ERROR;
}

static int
datum_assign_var(datum_t *x, datum_t *y)
{
    if (is_const_datum(x) || is_const_datum(y)) {
        if (x->type == y->type ||
            (is_integer_datum(x) && is_integer_datum(y)) ||
            (is_float_datum(x) && is_float_datum(y)))
            return NO_ERROR;

        RETURN(ERROR_MISMATCHED_TYPE, y->pos, datum_to_str(x), datum_to_str(y));
    }

    if (is_map_datum(x))
        return datum_assign_map(x, y);

    if (is_tuple_datum(x))
        return datum_assign_tuple(x, y);

    if (is_struct_datum(x))
        return datum_assign_struct(x, y);

    if (x->type != y->type)
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, datum_to_str(x), datum_to_str(y));

    return NO_ERROR;
}

static int
datum_assign_array(datum_t *x, int idx, datum_t *y)
{
    int i;

    if (is_array_datum(y)) {
        CHECK(datum_assign_var(x, y));

        if (x->arr_dim != y->arr_dim)
            RETURN(ERROR_MISMATCHED_TYPE, y->pos, datum_to_str(x), datum_to_str(y));

        for (i = 0; i < x->arr_dim; i++) {
            if (x->arr_size[i] != y->arr_size[i])
                RETURN(ERROR_MISMATCHED_COUNT, y->pos, "element", x->arr_size[i],
                       y->arr_size[i]);
        }
    }
    else if (is_tuple_datum(y)) {
        array_t *arr_elems = y->u_tup.datums;

        if (x->arr_size[idx] == -1)
            x->arr_size[idx] = array_size(arr_elems);
        else if (x->arr_size[idx] != array_size(arr_elems))
            RETURN(ERROR_MISMATCHED_COUNT, y->pos, "element", x->arr_size[idx],
                   array_size(arr_elems));

        for (i = 0; i < array_size(arr_elems); i++) {
            datum_t *val_datum = array_item(arr_elems, i, datum_t);

            if (idx < x->arr_dim - 1)
                CHECK(datum_assign_array(x, idx + 1, val_datum));
            else
                CHECK(datum_assign_var(x, val_datum));
        }
    }
    else {
        RETURN(ERROR_MISMATCHED_TYPE, y->pos, datum_to_str(x), datum_to_str(y));
    }

    return NO_ERROR;
}

int
datum_assign(datum_t *x, datum_t *y)
{
    if (is_array_datum(x))
        return datum_assign_array(x, 0, y);

    return datum_assign_var(x, y);
}

#define val_check_int(dat, max)                                                          \
    (((dat)->is_neg && (dat)->iv > (uint64_t)(max) + 1) ||                               \
     (!(dat)->is_neg && (dat)->iv > (uint64_t)(max)))

#define val_check_uint(dat, max)        ((dat)->is_neg || (dat)->iv > (max))

static void
value_check(datum_t *dat, type_t type)
{
    if ((type == TYPE_BYTE && val_check_uint(dat, UINT8_MAX)) ||
        (type == TYPE_INT8 && val_check_int(dat, INT8_MAX)) ||
        (type == TYPE_UINT8 && val_check_uint(dat, UINT8_MAX)) ||
        (type == TYPE_INT16 && val_check_int(dat, INT16_MAX)) ||
        (type == TYPE_UINT16 && val_check_uint(dat, UINT16_MAX)) ||
        (type == TYPE_INT32 && val_check_int(dat, INT32_MAX)) ||
        (type == TYPE_UINT32 && val_check_uint(dat, UINT32_MAX)) ||
        (type == TYPE_INT64 && val_check_int(dat, INT64_MAX)) ||
        (type == TYPE_UINT64 && is_neg_datum(dat)) ||
        (type == TYPE_FLOAT && dat->dv > FLT_MAX))
        ERROR(ERROR_NUMERIC_OVERFLOW, dat->pos, datum_to_str(dat));
}

static void
datum_cast_bool(datum_t *dat, type_t type)
{
    switch (type) {
    case TYPE_BOOL:
        break;

    case TYPE_STRING:
        dat->sv = dat->bv ? xstrdup("true") : xstrdup("false");
        break;

    default:
        ERROR(ERROR_INCOMPATIBLE_TYPE, dat->pos, datum_to_str(dat), TYPE_NAME(type));
    }
}

static void
datum_cast_int(datum_t *dat, type_t type)
{
    char buf[256];

    switch (type) {
    case TYPE_BYTE:
    case TYPE_INT8:
    case TYPE_UINT8:
    case TYPE_INT16:
    case TYPE_UINT16:
    case TYPE_INT32:
    case TYPE_UINT32:
    case TYPE_INT64:
    case TYPE_UINT64:
        value_check(dat, type);
        break;

    case TYPE_FLOAT:
    case TYPE_DOUBLE:
        dat->fv = (double)dat->iv;
        break;

    case TYPE_STRING:
        if (dat->is_neg)
            snprintf(buf, sizeof(buf), "%"PRId64, (int64_t)dat->iv);
        else
            snprintf(buf, sizeof(buf), "%"PRIu64, dat->iv);
        dat->sv = xstrdup(buf);
        break;

    default:
        ERROR(ERROR_INCOMPATIBLE_TYPE, dat->pos, datum_to_str(dat), TYPE_NAME(type));
    }
}

static void
datum_cast_fp(datum_t *dat, type_t type)
{
    switch (type) {
    case TYPE_BYTE:
    case TYPE_INT8:
    case TYPE_UINT8:
    case TYPE_INT16:
    case TYPE_UINT16:
    case TYPE_INT32:
    case TYPE_UINT32:
    case TYPE_INT64:
    case TYPE_UINT64:
        dat->iv = (uint64_t)dat->fv;
        value_check(dat, type);
        break;

    case TYPE_FLOAT:
    case TYPE_DOUBLE:
        value_check(dat, type);
        break;

    case TYPE_STRING:
        snprintf(buf, sizeof(buf), "%lf", dat->fv);
        break;

    default:
        ERROR(ERROR_INCOMPATIBLE_TYPE, dat->pos, datum_to_str(dat), TYPE_NAME(type));
    }
}

static void
datum_cast_str(datum_t *dat, type_t type)
{
    switch (type) {
    case TYPE_BOOL:
        dat->bv = strcmp(dat->sv, "true") == 0;
        break;

    case TYPE_BYTE:
    case TYPE_INT8:
    case TYPE_UINT8:
    case TYPE_INT16:
    case TYPE_UINT16:
    case TYPE_INT32:
    case TYPE_UINT32:
    case TYPE_INT64:
    case TYPE_UINT64:
        sscanf(dat->sv, "%"SCNu64, &dat->iv);
        value_check(dat, type);
        break;

    case TYPE_FLOAT:
    case TYPE_DOUBLE:
        sscanf(dat->sv, "%lf", &dat->fv);
        value_check(dat, type);
        break;

    case TYPE_STRING:
        break;

    default:
        ERROR(ERROR_INCOMPATIBLE_TYPE, dat->pos, datum_to_str(dat), TYPE_NAME(type));
    }
}

void
datum_cast(datum_t *dat, type_t type)
{
    if (is_bool_type(dat))
        datum_cast_bool(dat, type);
    else if (is_int_type(dat))
        datum_cast_int(dat, type);
    else if (is_fp_type(dat))
        datum_cast_fp(dat, type);
    else if (is_string_type(dat))
        datum_cast_str(dat, type);
    else
        ERROR(ERROR_INCOMPATIBLE_TYPE, dat->pos, datum_to_str(dat), TYPE_NAME(type));

    datum_set_type(dat, type);
}

#define dval_set_bool(dat, v)                                                            \
    do {                                                                                 \
        (dat)->bv = (v);                                                                 \
        (dat)->is_const = true;                                                          \
    } while (0)

#define dval_set_int(dat, v)                                                             \
    do {                                                                                 \
        (dat)->iv = (v);                                                                 \
        (dat)->is_neg = (v) < 0;                                                         \
        (dat)->is_const = true;                                                          \
    } while (0)

#define dval_set_fp(dat, v)                                                              \
    do {                                                                                 \
        (dat)->fv = (v);                                                                 \
        (dat)->is_neg = (v) < 0;                                                         \
        (dat)->is_const = true;                                                          \
    } while (0)

#define dval_set_str(dat, v)                                                             \
    do {                                                                                 \
        (dat)->sv = (v);                                                                 \
        (dat)->is_const = true;                                                          \
    } while (0)

#define datum_eval_arith(op, dat, x, y)                                                  \
    do {                                                                                 \
        ASSERT2(is_same_family((x), (y)), (x)->family, (y)->family);                     \
        if (is_int_family(x))                                                            \
            dval_set_int(dval_int(x) op dval_int(y));                                    \
        else if (is_fp_family(x))                                                        \
            dval_set_fp(dat, dval_fp(x) op dval_fp(y));                                  \
        else if (is_str_family(x))                                                       \
            dval_set_str((dat), xstrcat(dval_str(x), dval_str(y)));                      \
        else                                                                             \
            ASSERT2(!"invalid datatype", (x)->type, (y)->type);                          \
        datum_set_type((dat), MAX((x)->type, (y)->type));                                \
    } while (0)

#define datum_eval_cmp(op, dat, x, y)                                                    \
    do {                                                                                 \
        bool res = false;                                                                \
        ASSERT2(is_same_family((x), (y)), (x)->family, (y)->family);                     \
        if (is_bool_family(x))                                                           \
            res = dval_bool(x) op dval_bool(y);                                          \
        else if (is_int_family(x))                                                       \
            res = dval_int(x) op dval_int(y);                                            \
        else if (is_fp_family(x))                                                        \
            res = dval_fp(x) op dval_fp(y);                                              \
        else if (is_str_family(x))                                                       \
            res = strcmp(dval_str(x), dval_str(y)) op 0;                                 \
        else                                                                             \
            ASSERT2(!"invalid datatype", (x)->type, (y)->type);                          \
        dval_set_bool((dat), res);                                                      \
    } while (0)

#define datum_eval_bit(op, dat, x, y)                                                    \
    do {                                                                                 \
        ASSERT2(is_same_family((x), (y)), (x)->family, (y)->family);                     \
        if (is_int_family(x))                                                            \
            dval_set_int(dval_int(x) op dval_int(y));                                    \
        else                                                                             \
            ASSERT2(!"invalid datatype", (x)->type, (y)->type);                          \
        datum_set_type((dat), MAX((x)->type, (y)->type));                                \
    } while (0)

static void
datum_add(datum_t *dat, datum_t *x, datum_t *y)
{
    if (!is_num_family(x) && !is_str_family(x)) {
        ERROR(ERROR_INVALID_OP_TYPE, x->pos, datum_to_str(x));
    }
    else if (!is_num_family(y) && !is_str_family(y)) {
        ERROR(ERROR_INVALID_OP_TYPE, y->pos, datum_to_str(y));
    }
    else if (!is_same_family(x, y)) {
        ERROR(ERROR_INCOMPATIBLE_TYPE, y->pos, datum_to_str(x), datum_to_str(y));
    }
    else if (is_const_datum(x) && is_const_datum(y)) {
        switch (x->family) {
        case FAM_INT:
            dval_set_int(dat, dval_int(x) + dval_int(y));
            break;

        case FAM_FP:
            dval_set_fp(dat, dval_fp(x) + dval_fp(y));
            break;

        case FAM_STR:
            dval_set_str(dat, xstrcat(dval_str(x), dval_str(y)));
            break;
        }
    }

    datum_set_type(dat, MAX(x->type, y->type));
}

static void
datum_sub(datum_t *dat, datum_t *x, datum_t *y)
{
    if (!is_num_family(x)) {
        ERROR(ERROR_INVALID_OP_TYPE, x->pos, datum_to_str(x));
    }
    else if (!is_num_family(y)) {
        ERROR(ERROR_INVALID_OP_TYPE, y->pos, datum_to_str(y));
    }
    else if (!is_same_family(x, y)) {
        ERROR(ERROR_INCOMPATIBLE_TYPE, y->pos, datum_to_str(x), datum_to_str(y));
    }
    else if (is_const_datum(x) && is_const_datum(y)) {
        switch (x->family) {
        case FAM_INT:
            dval_set_int(dat, dval_int(x) - dval_int(y));
            break;

        case FAM_FP:
            dval_set_fp(dat, dval_fp(x) - dval_fp(y));
            break;
        }
    }

    datum_set_type(dat, MAX(x->type, y->type));
}

static void
datum_mul(datum_t *dat, datum_t *x, datum_t *y)
{
    if (!is_num_family(x)) {
        ERROR(ERROR_INVALID_OP_TYPE, x->pos, datum_to_str(x));
    }
    else if (!is_num_family(y)) {
        ERROR(ERROR_INVALID_OP_TYPE, y->pos, datum_to_str(y));
    }
    else if (!is_same_family(x, y)) {
        ERROR(ERROR_INCOMPATIBLE_TYPE, y->pos, datum_to_str(x), datum_to_str(y));
    }
    else if (is_const_datum(x) && is_const_datum(y)) {
        switch (x->family) {
        case FAM_INT:
            dval_set_int(dat, dval_int(x) * dval_int(y));
            break;

        case FAM_FP:
            dval_set_fp(dat, dval_fp(x) * dval_fp(y));
            break;
        }
    }

    datum_set_type(dat, MAX(x->type, y->type));
}

static void
datum_div(datum_t *dat, datum_t *x, datum_t *y)
{
    if (!is_num_family(x)) {
        ERROR(ERROR_INVALID_OP_TYPE, x->pos, datum_to_str(x));
    }
    else if (!is_num_family(y)) {
        ERROR(ERROR_INVALID_OP_TYPE, y->pos, datum_to_str(y));
    }
    else if (!is_same_family(x, y)) {
        ERROR(ERROR_INCOMPATIBLE_TYPE, y->pos, datum_to_str(x), datum_to_str(y));
    }
    else if (is_const_datum(x) && is_const_datum(y)) {
        switch (x->family) {
        case FAM_INT:
            if (dval_int(y) == 0)
                ERROR(ERROR_DIVIDE_BY_ZERO, y->pos);
            else
                dval_set_int(dat, dval_int(x) / dval_int(y));
            break;

        case FAM_FP:
            if (dval_fp(y) == 0.0)
                ERROR(ERROR_DIVIDE_BY_ZERO, y->pos);
            else
                dval_set_fp(dat, dval_fp(x) / dval_fp(y));
            break;
        }
    }

    datum_set_type(dat, MAX(x->type, y->type));
}

static void
datum_mod(datum_t *dat, datum_t *x, datum_t *y)
{
    if (!is_int_family(x)) {
        ERROR(ERROR_INVALID_OP_TYPE, x->pos, datum_to_str(x));
    }
    else if (!is_int_family(y)) {
        ERROR(ERROR_INVALID_OP_TYPE, y->pos, datum_to_str(y));
    }
    else if (is_const_datum(x) && is_const_datum(y)) {
        if (dval_int(y) == 0)
            ERROR(ERROR_DIVIDE_BY_ZERO, y->pos);
        else
            dval_set_int(dat, dval_int(x) % dval_int(y));
    }

    datum_set_type(dat, MAX(x->type, y->type));
}

#define datum_cmp(dat, op, x, y)                                                         \
    do {                                                                                 \
        if (is_const_datum(x) || is_const_datum(y)) {                                    \
            if (!is_same_family(x, y)) {                                                 \
                ERROR(ERROR_INCOMPATIBLE_TYPE, (y)->pos, datum_to_str(x), datum_to_str(y));  \
            }                                                                            \
            else if (is_const_datum(x) && is_const_datum(y)) {                           \
                switch ((x)->family) {                                                   \
                case FAM_BOOL:                                                           \
                    dval_set_bool(dat, dval_bool(x) op dval_bool(y));                    \
                    break;                                                               \
                case FAM_INT:                                                            \
                    dval_set_bool(dat, dval_int(x) op dval_int(y));                      \
                    break;                                                               \
                case FAM_FP:                                                             \
                    dval_set_bool(dat, dval_fp(x) op dval_fp(y));                        \
                    break;                                                               \
                case FAM_STR:                                                            \
                    dval_set_bool(dat, strcmp(dval_str(x), dval_str(y)) op 0);           \
                    break;                                                               \
                case FAM_OBJ:                                                            \
                    dval_set_bool(dat, dval_ptr(x) op dval_ptr(y));                      \
                    break;                                                               \
                }                                                                        \
            }                                                                            \
        }                                                                                \
        else if ((x)->type != (y)->type) {                                               \
            ERROR(ERROR_INCOMPATIBLE_TYPE, (y)->pos, datum_to_str(x), datum_to_str(y));      \
        }                                                                                \
        datum_set_bool(dat);                                                             \
    } while (0)

static void
datum_cmp_eq(datum_t *dat, datum_t *x, datum_t *y)
{
    datum_cmp(dat, ==, x, y);
}

static void
datum_cmp_ne(datum_t *dat, datum_t *x, datum_t *y)
{
    datum_cmp(!=, dat, x, y);
}

static void
datum_cmp_lt(datum_t *dat, datum_t *x, datum_t *y)
{
    datum_cmp(<, dat, x, y);
}

static void
datum_cmp_gt(datum_t *dat, datum_t *x, datum_t *y)
{
    datum_cmp(>, dat, x, y);
}

static void
datum_cmp_le(datum_t *dat, datum_t *x, datum_t *y)
{
    datum_cmp(<=, dat, x, y);
}

static void
datum_cmp_ge(datum_t *dat, datum_t *x, datum_t *y)
{
    datum_cmp(>=, dat, x, y);
}

static void
datum_bit_and(datum_t *dat, datum_t *x, datum_t *y)
{
    datum_eval_bit(&, dat, x, y);
}

static void
datum_bit_or(datum_t *dat, datum_t *x, datum_t *y)
{
    datum_eval_bit(|, dat, x, y);
}

static void
datum_bit_xor(datum_t *dat, datum_t *x, datum_t *y)
{
    datum_eval_bit(^, dat, x, y);
}

static void
datum_shift_l(datum_t *dat, datum_t *x, datum_t *y)
{
    datum_eval_bit(<<, dat, x, y);
}

static void
datum_shift_r(datum_t *dat, datum_t *x, datum_t *y)
{
    ASSERT2((x)->kind == (y)->kind, (x)->kind, (y)->kind);             \
    datum_eval_bit(>>, dat, x, y);
}

static void
datum_neg(datum_t *dat, datum_t *x, datum_t *y)
{
    if (is_int_type(x))
        datum_set_int(dat, dval_int(x));
    else if (is_fp_type(x))
        datum_set_double(dat, dval_fp(x));
    else
        ASSERT1(!"invalid datum", dat->kind);

    datum_set_neg(dat, !x->is_neg);
}

static void
datum_not(datum_t *dat, datum_t *x, datum_t *y)
{
    if (is_bool_type(x))
        datum_set_bool(dat, !x->bv);
    else
        ASSERT1(!"invalid datum", dat->kind);
}

eval_fn_t eval_fntab_[OP_CF_MAX] = {
    datum_add,
    datum_sub,
    datum_mul,
    datum_div,
    datum_mod,
    datum_cmp_eq,
    datum_cmp_ne,
    datum_cmp_lt,
    datum_cmp_gt,
    datum_cmp_le,
    datum_cmp_ge,
    datum_bit_and,
    datum_bit_or,
    datum_bit_xor,
    datum_shift_l,
    datum_shift_r,
    datum_neg,
    datum_not
};

void
datum_eval(datum_t *dat, op_t op, datum_t *x, datum_t *y)
{
    ASSERT2(is_same_family(x, y), x->type, y->type);

    eval_fntab_[op](dat, x, y);
}

/* end of datum.c */
