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

typedef enum blk_kind_e {
    BLK_ANON        = 0,
    BLK_ROOT,
    BLK_LOOP,
    BLK_MAX
} blk_kind_t;

struct ast_blk_s {
    AST_NODE_DECL;

    blk_kind_t kind;

    array_t ids;
    array_t stmts;

    /* results of semantic checker */
    ast_blk_t *up;
};

ast_blk_t *blk_anon_new(src_pos_t *pos);
ast_blk_t *blk_root_new(src_pos_t *pos);
ast_blk_t *blk_loop_new(src_pos_t *pos);

ast_blk_t *blk_search_loop(ast_blk_t *blk);

void ast_blk_dump(ast_blk_t *blk, int indent);

#endif /* ! _AST_BLK_H */
