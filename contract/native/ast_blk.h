/**
 * @file    ast_blk.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_BLK_H
#define _AST_BLK_H

#include "common.h"

#include "ast.h"

#ifndef _AST_BLK_T
#define _AST_BLK_T
typedef struct ast_blk_s ast_blk_t;
#endif /* ! _AST_BLK_T */

#ifndef _AST_STRUCT_T
#define _AST_STRUCT_T
typedef struct ast_struct_s ast_struct_t;
#endif /* ! _AST_STRUCT_T */

struct ast_blk_s {
    AST_NODE_DECL;

    char *name;

    list_t var_l;
    list_t struct_l;
    list_t stmt_l;
    list_t func_l;

    ast_blk_t *up;
};

ast_blk_t *ast_blk_new(errpos_t *pos);

void ast_blk_dump(ast_blk_t *blk, int indent);

#endif /* ! _AST_BLK_H */
