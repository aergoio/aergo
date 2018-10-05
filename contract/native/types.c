/**
 * @file    types.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "types.h"

char *type_strs_[TYPE_MAX] = {
    "undefined",
    "bool",
    "byte",
    "float",
    "double",
    "int8",
    "uint8",
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
    "void",
    "reference",
    "tuple"
}; 

/* end of types.c */
