/**
 * @file    ast.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_H
#define _AST_H

#include "common.h"

#include "list.h"

#define AST_NODE_DECL                                                          \
    errpos_t pos

typedef struct ast_s {
    list_t blk_l;
} ast_t;

ast_t *ast_new(void);

void ast_dump(ast_t *ast);

#endif /* ! _AST_H */
