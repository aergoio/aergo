/**
 * @file    ast.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_H
#define _AST_H

#include "common.h"

#include "src_pos.h"
#include "error.h"

#define AST_NODE_DECL               src_pos_t pos

#define ast_node_init(node, npos)   (node)->pos = (npos)

#ifndef _AST_BLK_T
#define _AST_BLK_T
typedef struct ast_blk_s ast_blk_t;
#endif /* ! _AST_BLK_T */

typedef struct ast_s {
    /* If there are multiple contracts in a file, the root block is used to follow the up-pointer
     * of the block to make the name resolution more natural for the variable. */
    ast_blk_t *root;
} ast_t;

ast_t *ast_new(void);

#endif /* ! _AST_H */
