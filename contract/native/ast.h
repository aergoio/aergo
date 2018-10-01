/**
 * @file    ast.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_H
#define _AST_H

#include "common.h"

#include "array.h"

#define AST_NODE_DECL                                                          \
    int num;                                                                   \
    errpos_t pos

#define ast_node_init(node, epos)                                              \
    do {                                                                       \
        (node)->num = node_num_++;                                             \
        (node)->pos = *(epos);                                                 \
    } while (0)

#ifndef _AST_BLK_T
#define _AST_BLK_T
typedef struct ast_blk_s ast_blk_t;
#endif /* ! _AST_BLK_T */

typedef struct ast_s {
    ast_blk_t *root;
} ast_t;

extern int node_num_;

ast_t *ast_new(void);

void ast_dump(ast_t *ast);

#endif /* ! _AST_H */
