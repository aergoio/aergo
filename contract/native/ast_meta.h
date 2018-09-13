/**
 * @file    ast_meta.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_META_H
#define _AST_META_H

#include "common.h"

typedef enum type_e {
    TYPE_NONE       = 0,
    TYPE_ACCOUNT,
    TYPE_BOOL,
    TYPE_BYTE,
    TYPE_FLOAT,
    TYPE_DOUBLE,
    TYPE_INT8,
    TYPE_INT16,
    TYPE_INT32,
    TYPE_INT64,
    TYPE_UINT8,
    TYPE_UINT16,
    TYPE_UINT32,
    TYPE_UINT64,
    TYPE_STRING,
    TYPE_MAP,
    TYPE_STRUCT,
    TYPE_MAX
} type_t; 

typedef struct meta_map_s {
    type_t key;
    type_t value;
} meta_map_t;

typedef struct meta_struct_s {
    char *name;
} meta_struct_t;

typedef struct ast_meta_s {
    type_t type;
    bool is_const;
    int arr_size;

    union {
        meta_map_t map;
        meta_struct_t st;
    } u;
} ast_meta_t;

#endif /* _AST_META_H */
