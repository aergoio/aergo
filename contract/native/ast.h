/**
 * @file    ast.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_H
#define _AST_H

#include "common.h"

#define AST_NODE_NAME_SIZE          128

#define AST_NODE_DECL                                                                    \
    int num;                                                                             \
    src_pos_t pos

#define ast_node_init(node, pos)                                                         \
    do {                                                                                 \
        (node)->num = node_num_++;                                                       \
        (node)->pos = *(pos);                                                            \
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

#endif /* ! _AST_H */
