/**
 * @file    ast_type.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_TYPE_H
#define _AST_TYPE_H

#include "common.h"

#define TYPENAME(type)              type_strs_[type]

#define type_is_valid(type)         ((type) > TYPE_NONE && (type) < TYPE_MAX)

typedef enum type_e {
    TYPE_NONE       = 0,
    TYPE_BOOL,
    TYPE_BYTE,
    TYPE_FLOAT,
    TYPE_DOUBLE,
    TYPE_INT16,
    TYPE_UINT16,
    TYPE_INT32,
    TYPE_UINT32,
    TYPE_INT64,
    TYPE_UINT64,
    TYPE_STRING,
    TYPE_STRUCT,
    TYPE_ACCOUNT,
    TYPE_COMPARABLE = TYPE_ACCOUNT,

    TYPE_MAP,
    TYPE_PRIMITIVE  = TYPE_MAP,

    TYPE_VOID,
    TYPE_REF,                       /* only for NULL */
    TYPE_TUPLE,
    TYPE_MAX
} type_t;

extern char *type_strs_[TYPE_MAX];

static inline bool
type_check_range(type_t type, int64_t val)
{
    switch (type) {
    case TYPE_INT16:
        if (val < INT16_MIN || val > INT16_MAX)
            return false;
        break;

    case TYPE_UINT16:
        if (val > UINT16_MAX)
            return false;
        break;

    case TYPE_INT32:
        if (val < INT32_MIN || val > INT32_MAX)
            return false;
        break;

    case TYPE_UINT32:
        if (val > UINT32_MAX)
            return false;
        break;

    default:
        break;
    }

    return true;
}

#endif /* ! _AST_TYPE_H */
