/**
 * @file    ast_type.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_TYPE_H
#define _AST_TYPE_H

#include "common.h"

#define TYPENAME(type)              type_strs_[type]

#define type_is_valid(type)         ((type) > TYPE_NONE && (type) < TYPE_MAX)

#define type_is_bool(type)          ((type) == TYPE_BOOL)
#define type_is_integer(type)                                                  \
    ((type) >= TYPE_INT16 && (type) <= TYPE_UINT64)
#define type_is_float(type)                                                    \
    ((type) == TYPE_FLOAT || (type) == TYPE_DOUBLE)

#define type_is_primitive(type)     ((type) <= TYPE_STRING)
#define type_is_struct(type)        ((type) == TYPE_STRUCT)
#define type_is_map(type)           ((type) == TYPE_MAP)
#define type_is_contract(type)      ((type) == TYPE_CONTRACT)

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
    TYPE_ACCOUNT,
    TYPE_STRUCT,
    TYPE_MAP,
    TYPE_CONTRACT,
    TYPE_MAX
} type_t;

extern char *type_strs_[TYPE_MAX];

#endif /* ! _AST_TYPE_H */
