/**
 * @file    enum.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gmp.h"

#include "enum.h"

char *type_names_[TYPE_MAX] = {
    "undefined",
    "bool",
    "byte",
    "int8",
    "int16",
    "int32",
    "int64",
    "int256",
    "string",
    "account",
    "struct",
    "array",
    "map",
    "object",
    "cursor",
    "void",
    "tuple"
};

#define I32             4
#define I64             8

int type_sizes_[TYPE_MAX] = {
    0,                  /* TYPE_NONE */
    I32,                /* TYPE_BOOL */
    I32,                /* TYPE_BYTE */
    I32,                /* TYPE_INT8 */
    I32,                /* TYPE_INT16 */
    I32,                /* TYPE_INT32 */
    I64,                /* TYPE_INT64 */
    I32,                /* TYPE_INT256 */
    I32,                /* TYPE_STRING */
    I32,                /* TYPE_ACCOUNT */
    I32,                /* TYPE_STRUCT */
    I32,                /* TYPE_ARRAY */
    I32,                /* TYPE_MAP */
    I32,                /* TYPE_OBJECT */
    I32,                /* TYPE_CURSOR */
    0,                  /* TYPE_VOID */
    0                   /* TYPE_TUPLE */
};

int type_c_sizes_[TYPE_MAX] = {
    0,                  /* TYPE_NONE */
    sizeof(bool),       /* TYPE_BOOL */
    sizeof(uint8_t),    /* TYPE_BYTE */
    sizeof(int8_t),     /* TYPE_INT8 */
    sizeof(int16_t),    /* TYPE_INT16 */
    sizeof(int32_t),    /* TYPE_INT32 */
    sizeof(int64_t),    /* TYPE_INT64 */
    sizeof(uint32_t),   /* TYPE_INT256 */
    sizeof(int32_t),    /* TYPE_STRING */
    sizeof(int32_t),    /* TYPE_ACCOUNT */
    sizeof(int32_t),    /* TYPE_STRUCT */
    sizeof(int32_t),    /* TYPE_ARRAY */
    sizeof(int32_t),    /* TYPE_MAP */
    sizeof(int32_t),    /* TYPE_OBJECT */
    sizeof(int32_t),    /* TYPE_CURSOR */
    0,                  /* TYPE_VOID */
    0                   /* TYPE_TUPLE */
};

/* end of enum.c */
