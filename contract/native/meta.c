/**
 * @file    meta.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "meta.h"

char *type_strs_[TYPE_MAX] = {
    "undefined",
    "bool",
    "byte",
    "int8",
    "uint8",
    "int16",
    "uint16",
    "int32",
    "uint32",
    "float",
    "int64",
    "uint64",
    "double",
    "string",
    "struct",
    "reference",
    "account",
    "map",
    "void",
    "tuple"
};

void 
meta_dump(meta_t *meta, int indent)
{
}

/* end of meta.c */
