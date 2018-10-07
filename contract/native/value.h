/**
 * @file    value.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _VALUE_H
#define _VALUE_H

#include "common.h"

#define is_null_val(val)            ((val)->kind == VAL_NULL)
#define is_bool_val(val)            ((val)->kind == VAL_BOOL)
#define is_int_val(val)             ((val)->kind == VAL_INT)
#define is_float_val(val)           ((val)->kind == VAL_FP)
#define is_string_val(val)          ((val)->kind == VAL_STR)

#ifndef _VALUE_T
#define _VALUE_T
typedef struct value_s value_t;
#endif /* ! _VALUE_T */

typedef enum val_kind_e {
    VAL_NULL        = 0,
    VAL_BOOL,
    VAL_INT,
    VAL_FP,
    VAL_STR,
    VAL_MAX
} val_kind_t;

struct value_s {
    val_kind_t kind;

    union {
        bool bv;
        int64_t iv;
        double dv;
        char *sv;
    };
};

static inline void
value_init(value_t *val)
{
    ASSERT(val != NULL);
    memset(val, 0x00, sizeof(value_t));
}

static inline void
val_set_null(value_t *val)
{
    val->kind = VAL_NULL;
}

static inline void 
val_set_bool(value_t *val, bool bv)
{
    val->kind = VAL_BOOL;
    val->bv = bv;
}

static inline void 
val_set_int(value_t *val, char *str)
{
    val->kind = VAL_INT;
    sscanf(str, "%"SCNd64, &val->iv);
}

static inline void 
val_set_hexa(value_t *val, char *str)
{
    val->kind = VAL_INT;
    sscanf(str, "%"SCNx64, &val->iv);
}

static inline void 
val_set_fp(value_t *val, char *str)
{
    val->kind = VAL_FP;
    sscanf(str, "%lf", &val->dv);
}

static inline void 
val_set_str(value_t *val, char *str)
{
    val->kind = VAL_STR;
    val->sv = str;
}

static inline bool
value_check_range(value_t *val, type_t type)
{
    if (type == TYPE_INT8 && (val->iv < INT8_MIN || val->iv > INT8_MAX))
        return false;

    if (type == TYPE_UINT8 && val->iv > UINT8_MAX)
        return false;

    if (type == TYPE_INT16 && (val->iv < INT16_MIN || val->iv > INT16_MAX))
        return false;

    if (type == TYPE_UINT16 && val->iv > UINT16_MAX)
        return false;

    if (type == TYPE_INT32 && (val->iv < INT32_MIN || val->iv > INT32_MAX))
        return false;

    if (type == TYPE_UINT32 && val->iv > UINT32_MAX)
        return false;

    if (type == TYPE_FLOAT && (val->dv < FLT_MIN || val->dv > FLT_MAX))
        return false;

    return true;
}

#endif /* ! _VALUE_H */
