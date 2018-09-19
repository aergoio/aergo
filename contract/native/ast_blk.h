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

struct ast_blk_s {
    AST_NODE_DECL;

    list_t var_l;
    list_t struct_l;
    list_t stmt_l;

    ast_blk_t *up;
};

ast_blk_t *ast_blk_new(errpos_t *pos);

#endif /* ! _AST_BLK_H */
