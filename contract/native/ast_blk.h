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

    array_t ids;
    array_t stmts;

    ast_blk_t *up;
};

ast_blk_t *ast_blk_new(errpos_t *pos);

void ast_blk_dump(ast_blk_t *blk, int indent);

#endif /* ! _AST_BLK_H */
