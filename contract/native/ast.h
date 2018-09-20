/**
 * @file    ast.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_H
#define _AST_H

#include "common.h"

#include "list.h"

#define AST_NODE_DECL                                                          \
    list_link_t link;                                                          \
    errpos_t pos

typedef struct ast_cont_s {
    AST_NODE_DECL;

    char *name;

    list_t var_l;
    list_t struct_l;
    list_t func_l;
} ast_cont_t;

typedef struct ast_s {
    list_t cont_l;
} ast_t;

#endif /* ! _AST_H */
