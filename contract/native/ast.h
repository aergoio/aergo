/**
 * @file    ast.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_H
#define _AST_H

#include "common.h"

#define AST_NODE_DECL                                                                    \
    node_kind_t kind;                                                                    \
    int num

#define ast_node_init(node, kind_)                                                       \
    do {                                                                                 \
        (node)->kind = (kind_);                                                          \
        (node)->num = node_num_;                                                         \
    } while (0)

#define is_id_node(node)            ((node)->kind >= ID_START && (node)->kind <= ID_MAX)
#define is_exp_node(node)           ((node)->kind >= EXP_START && (node)->kind <= EXP_MAX)
#define is_stmt_node(node)                                                               \
    ((node)->kind >= STMT_START && (node)->kind <= STMT_MAX)

#define node_add                    array_add_last

#ifndef _AST_BLK_T
#define _AST_BLK_T
typedef struct ast_blk_s ast_blk_t;
#endif /* ! _AST_BLK_T */

typedef struct ast_node_s {
    AST_NODE_DECL;
} ast_node_t;

typedef struct ast_s {
    /* If there are multiple contracts in a file, the root block is used
     * to follow the up-pointer of the block to make the name resolution
     * more natural for the variable. */
    ast_blk_t *root;
} ast_t;

extern int node_num_;

ast_t *ast_new(void);

#endif /* ! _AST_H */
