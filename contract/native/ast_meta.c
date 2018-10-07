/**
 * @file    ast_meta.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_meta.h"

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

void 
ast_meta_dump(ast_meta_t *meta, int indent)
{
}

/* end of ast_meta.c */
