/**
 * @file    ast_type.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_type.h"

char *type_strs_[TYPE_MAX] = {
    "undefined",
    "bool",
    "byte",
    "float",
    "double",
    "int16",
    "uint16",
    "int32",
    "uint32",
    "int64",
    "uint64",
    "string",
    "struct",
    "account",
    "map",
    "reference",
    "tuple"
}; 

/* end of ast_type.c */
