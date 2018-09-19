/**
 * @file    ast_func.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_FUNC_H
#define _AST_FUNC_H

#include "common.h"

#include "ast.h"

#ifndef _AST_STMT_T
#define _AST_STMT_T
typedef struct ast_stmt_s ast_stmt_t;
#endif /* ! _AST_STMT_T */

typedef enum modifier_e {
    MOD_GLOBAL      = 0x00,
    MOD_LOCAL       = 0x01,
    MOD_SHARED      = 0x02,
    MOD_TRANSFER    = 0x04,
    MOD_READONLY    = 0x08
} modifier_t;

typedef struct ast_func_s {
    AST_NODE_DECL;

    char *name;
    modifier_t mod;

    list_t *param_l;
    list_t *return_l;

    ast_stmt_t *blk;
} ast_func_t;

ast_func_t *ast_func_new(char *name, modifier_t mod, list_t *param_l,
                         list_t *return_l, ast_stmt_t *blk, errpos_t *pos);

#endif /* ! _AST_FUNC_H */
