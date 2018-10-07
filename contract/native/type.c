/**
 * @file    type.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "type.h"

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
    "reference",
    "account",
    "map",
    "void",
    "tuple"
};

/* end of type.c */
