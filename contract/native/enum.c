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
    "int8",
    "uint8",
    "int16",
    "uint16",
    "int32",
    "uint32",
    "int64",
    "uint64",
    "int128",
    "uint128",
    "float",
    "double",
    "string",
    "account",
    "struct",
    "map",
    "object",
    "cursor",
    "void",
    "tuple"
};

#define I32             4
#define I64             8
#define F32             4
#define F64             8
#define ADDR            4

int type_sizes_[TYPE_MAX] = {
    0,                  /* TYPE_NONE */
    I32,                /* TYPE_BOOL */
    I32,                /* TYPE_INT8 */
    I32,                /* TYPE_UINT8 */
    I32,                /* TYPE_INT16 */
    I32,                /* TYPE_UINT16 */
    I32,                /* TYPE_INT32 */
    I32,                /* TYPE_UINT32 */
    I64,                /* TYPE_INT64 */
    I64,                /* TYPE_UINT64 */
    ADDR,               /* TYPE_INT128 */
    ADDR,               /* TYPE_UINT128 */
    F32,                /* TYPE_FLOAT */
    F64,                /* TYPE_DOUBLE */
    ADDR,               /* TYPE_STRING */
    ADDR,               /* TYPE_ACCOUNT */
    ADDR,               /* TYPE_STRUCT */
    ADDR,               /* TYPE_MAP */
    ADDR,               /* TYPE_OBJECT */
    ADDR,               /* TYPE_CURSOR */
    0,                  /* TYPE_VOID */
    ADDR                /* TYPE_TUPLE */
};

int type_bytes_[TYPE_MAX] = {
    0,                  /* TYPE_NONE */
    sizeof(bool),
    sizeof(int8_t),
    sizeof(uint8_t),
    sizeof(int16_t),
    sizeof(uint16_t),
    sizeof(int32_t),
    sizeof(uint32_t),
    sizeof(int64_t),
    sizeof(uint64_t),
    sizeof(int32_t),
    sizeof(int32_t),
    sizeof(float),
    sizeof(double),
    sizeof(int32_t),    /* TYPE_STRING */
    sizeof(int32_t),    /* TYPE_ACCOUNT */
    sizeof(int32_t),    /* TYPE_STRUCT */
    sizeof(int32_t),    /* TYPE_MAP */
    sizeof(int32_t),    /* TYPE_OBJECT */
    sizeof(int32_t),    /* TYPE_CURSOR */
    0,                  /* TYPE_VOID */
    0                   /* TYPE_TUPLE */
};

/* end of enum.c */
