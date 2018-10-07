/**
 * @file    types.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _TYPES_H
#define _TYPES_H

#include "common.h"

#define TYPE_NAME(type)             type_strs_[type]

#define type_is_valid(type)         ((type) > TYPE_NONE && (type) < TYPE_MAX)

typedef enum type_e {
    TYPE_NONE       = 0,
    TYPE_BOOL,
    TYPE_BYTE,
    TYPE_FLOAT,
    TYPE_DOUBLE,
    TYPE_INT8,
    TYPE_UINT8,
    TYPE_INT16,
    TYPE_UINT16,
    TYPE_INT32,
    TYPE_UINT32,
    TYPE_INT64,
    TYPE_UINT64,
    TYPE_STRING,
    TYPE_STRUCT,
    TYPE_REF,
    TYPE_ACCOUNT,
    TYPE_COMPARABLE = TYPE_ACCOUNT,

    TYPE_MAP,
    TYPE_PRIMITIVE  = TYPE_MAP,

    TYPE_VOID,                      /* for return type of function */
    TYPE_TUPLE,                     /* for tuple expression */
    TYPE_MAX
} type_t;

extern char *type_strs_[TYPE_MAX];

static inline bool
type_check_range(type_t type, int64_t val)
{
    if (type == TYPE_INT8 && (val < INT8_MIN || val > INT8_MAX))
        return false;

    if (type == TYPE_UINT8 && val > UINT8_MAX)
        return false;

    if (type == TYPE_INT16 && (val < INT16_MIN || val > INT16_MAX))
        return false;

    if (type == TYPE_UINT16 && val > UINT16_MAX)
        return false;

    if (type == TYPE_INT32 && (val < INT32_MIN || val > INT32_MAX))
        return false;

    if (type == TYPE_UINT32 && val > UINT32_MAX)
        return false;

    return true;
}

#endif /* ! _TYPES_H */
