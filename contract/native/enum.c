/**
 * @file    enum.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "enum.h"

char *type_names_[TYPE_MAX] = {
    "undefined",
    "bool",
    "byte",
    "int8",
    "uint8",
    "int16",
    "uint16",
    "int32",
    "uint32",
    "int64",
    "uint64",
    "float",
    "double",
    "string",
    "struct",
    "reference",
    "account",
    "map",
    "void",
    "tuple"
};

char *stmt_names_[STMT_MAX] = {
    "null",
    "exp",
    "if",
    "loop",
    "switch",
    "case",
    "continue",
    "break",
    "return",
    "goto",
    "sql",
    "block"
};

/* end of enum.c */
