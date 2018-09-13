/**
 * @file    ast_struct.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_STRUCT_H
#define _AST_STRUCT_H

#include "common.h"

#include "list.h"
#include "location.h"

typedef struct ast_struct_s {
    char *name;
    list_t field_l;

    yypos_t pos;
} ast_struct_t;

#endif /* _AST_STRUCT_H */
