/**
 * @file    ast.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_H
#define _AST_H

#include "common.h"

#include "list.h"

#define AST_NODE_DECL                                                          \
    int num;                                                                   \
    errpos_t pos

#define ast_node_init(node, epos)                                              \
    do {                                                                       \
        (node)->num = node_num_++;                                             \
        (node)->pos = *(epos);                                                 \
    } while (0)

typedef struct ast_s {
    list_t blk_l;
} ast_t;

extern int node_num_;

ast_t *ast_new(void);

void ast_dump(ast_t *ast);

#endif /* ! _AST_H */
