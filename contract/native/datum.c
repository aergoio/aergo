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
datum_type(datum_t *dat)
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
        strbuf_cat(&buf, datum_type(dat->map.k));
        strbuf_cat(&buf, ", ");
        strbuf_cat(&buf, datum_type(dat->map.v));
        strbuf_cat(&buf, ")");
    }
    else if (is_tuple_type(dat)) {
        int i;

        strbuf_cat(&buf, "{");
        for (i = 0; i < dat->tup.cnt; i++) {
            if (i > 0)
                strbuf_cat(&buf, ", ");

            strbuf_cat(&buf, datum_type(dat->tup.evs[i]));
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
    ASSERT(arr_dim >= 0);

    dat->arr_dim = arr_dim;
    dat->arr_size = xcalloc(sizeof(int) * arr_dim);
}

void
datum_set_map(datum_t *dat, datum_t *k, datum_t *v)
{
    datum_set_type(dat, TYPE_MAP);

    dat->map.k = k;
    dat->map.v = v;
}

void
datum_set_struct(datum_t *dat, char *name, array_t *ids)
{
    int i;

    datum_set_type(dat, TYPE_STRUCT);

    dat->name = name;

    dat->tup.cnt = array_size(ids);
    dat->tup.evs = xmalloc(sizeof(datum_t *) * dat->tup.cnt);

    for (i = 0; i < array_size(ids); i++) {
        dat->tup.evs[i] = &array_item(ids, i, ast_id_t)->dat;
    }
}

void
datum_set_tuple(datum_t *dat, array_t *exps)
{
    int i;

    datum_set_type(dat, TYPE_TUPLE);

    dat->tup.cnt = array_size(exps);
    dat->tup.evs = xmalloc(sizeof(datum_t *) * dat->tup.cnt);

    for (i = 0; i < array_size(exps); i++) {
        dat->tup.evs[i] = &array_item(exps, i, ast_exp_t)->dat;
    }
}

#define val_check_int(dat, max)                                                          \
    (((dat)->is_neg && (dat)->cv.iv > (uint64_t)(max) + 1) ||                            \
     (!(dat)->is_neg && (dat)->cv.iv > (uint64_t)(max)))

#define val_check_uint(dat, max)                                                         \
    ((dat)->is_neg || (dat)->cv.iv > (max))

static void
value_check(datum_t *dat)
{
    if ((is_byte_type(dat) && val_check_uint(dat, UINT8_MAX)) ||
        (is_int8_type(dat) && val_check_int(dat, INT8_MAX)) ||
        (is_uint8_type(dat) && val_check_uint(dat, UINT8_MAX)) ||
        (is_int16_type(dat) && val_check_int(dat, INT16_MAX)) ||
        (is_uint16_type(dat) && val_check_uint(dat, UINT16_MAX)) ||
        (is_int32_type(dat) && val_check_int(dat, INT32_MAX)) ||
        (is_uint32_type(dat) && val_check_uint(dat, UINT32_MAX)) ||
        (is_int64_type(dat) && val_check_int(dat, INT64_MAX)) ||
        (is_uint64_type(dat) && is_neg_val(dat)) ||
        (is_float_type(dat) && dat->cv.dv > FLT_MAX))
        ERROR(ERROR_NUMERIC_OVERFLOW, dat->pos, datum_type(dat));
}

static void
datum_cast_bool(datum_t *dat, type_t type)
{
    switch (type) {
    case TYPE_BOOL:
        break;

    case TYPE_STRING:
        dat->cv.sv = dat->bv ? xstrdup("true") : xstrdup("false");
        break;

    default:
        ERROR(ERROR_INCOMPATIBLE_TYPE, dat->pos, datum_type(dat), TYPE_NAME(type));
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
        break;

    case TYPE_FLOAT:
    case TYPE_DOUBLE:
        dat->cv.fv = (double)dat->cv.iv;
        break;

    case TYPE_STRING:
        if (dat->is_neg)
            snprintf(buf, sizeof(buf), "%"PRId64, (int64_t)dat->cv.iv);
        else
            snprintf(buf, sizeof(buf), "%"PRIu64, dat->cv.iv);
        dat->cv.sv = xstrdup(buf);
        break;

    default:
        ERROR(ERROR_INCOMPATIBLE_TYPE, dat->pos, datum_type(dat), TYPE_NAME(type));
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
        dat->cv.iv = (uint64_t)dat->cv.fv;
        break;

    case TYPE_FLOAT:
    case TYPE_DOUBLE:
        break;

    case TYPE_STRING:
        snprintf(buf, sizeof(buf), "%lf", dat->cv.fv);
        break;

    default:
        ERROR(ERROR_INCOMPATIBLE_TYPE, dat->pos, datum_type(dat), TYPE_NAME(type));
    }
}

static void
datum_cast_str(datum_t *dat, type_t type)
{
    switch (type) {
    case TYPE_BOOL:
        dat->cv.bv = strcmp(dat->cv.sv, "true") == 0;
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
        sscanf(dat->cv.sv, "%"SCNu64, &dat->cv.iv);
        break;

    case TYPE_FLOAT:
    case TYPE_DOUBLE:
        sscanf(dat->cv.sv, "%lf", &dat->cv.fv);
        break;

    case TYPE_STRING:
        break;

    default:
        ERROR(ERROR_INCOMPATIBLE_TYPE, dat->pos, datum_type(dat), TYPE_NAME(type));
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
        ERROR(ERROR_INCOMPATIBLE_TYPE, dat->pos, datum_type(dat), TYPE_NAME(type));

    datum_set_type(dat, type);
}

/* end of datum.c */
