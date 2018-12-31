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
    "account",
    "struct",
    "map",
    "object",
    "void",
    "tuple"
};

int type_sizes_[TYPE_MAX] = {
    0,                  /* TYPE_NONE */
    sizeof(bool),
    sizeof(uint8_t),
    sizeof(int8_t),
    sizeof(uint8_t),
    sizeof(int16_t),
    sizeof(uint16_t),
    sizeof(int32_t),
    sizeof(uint32_t),
    sizeof(int64_t),
    sizeof(uint64_t),
    sizeof(float),
    sizeof(double),
    sizeof(int32_t),    /* TYPE_STRING */
    sizeof(int32_t),    /* TYPE_ACCOUNT */
    sizeof(int32_t),    /* TYPE_STRUCT */
    sizeof(int32_t),    /* TYPE_MAP */
    sizeof(int32_t),    /* TYPE_OBJECT */
    -1,                 /* TYPE_VOID */
    -1                  /* TYPE_TUPLE */
};

int type_aligns_[TYPE_MAX] = {
    -1,                 /* TYPE_NONE */
    4,                  /* TYPE_BOOL */
    4,                  /* TYPE_BYTE */
    4,                  /* TYPE_INT8 */
    4,                  /* TYPE_UINT8 */
    4,                  /* TYPE_INT16 */
    4,                  /* TYPE_UINT16 */
    4,                  /* TYPE_INT32 */
    4,                  /* TYPE_UINT32 */
    8,                  /* TYPE_INT64 */
    8,                  /* TYPE_UINT64 */
    4,                  /* TYPE_FLOAT */
    8,                  /* TYPE_DOUBLE */
    4,                  /* TYPE_STRING */
    4,                  /* TYPE_ACCOUNT */
    4,                  /* TYPE_STRUCT */
    4,                  /* TYPE_MAP */
    4,                  /* TYPE_OBJECT */
    -1,                 /* TYPE_VOID */
    4                   /* TYPE_TUPLE */
};

char *id_kinds_[ID_MAX] = {
    "variable",
    "struct",
    "enumeration",
    "function",
    "contract",
    "label",
    "tuple"
};

char *stmt_kinds_[STMT_MAX] = {
    "null",
    "exp",
    "assign",
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
