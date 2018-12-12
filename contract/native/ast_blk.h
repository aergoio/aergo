/**
 * @file    ast_blk.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_BLK_H
#define _AST_BLK_H

#include "common.h"

#include "ast.h"
#include "array.h"

#define is_normal_blk(blk)          ((blk)->kind == BLK_NORMAL)
#define is_root_blk(blk)            ((blk)->kind == BLK_ROOT)
#define is_contract_blk(blk)        ((blk)->kind == BLK_CONTRACT)
#define is_func_blk(blk)            ((blk)->kind == BLK_FUNC)
#define is_loop_blk(blk)            ((blk)->kind == BLK_LOOP)

#ifndef _AST_BLK_T
#define _AST_BLK_T
typedef struct ast_blk_s ast_blk_t;
#endif /* ! _AST_BLK_T */

struct ast_blk_s {
    AST_NODE_DECL;

    blk_kind_t kind;

    char name[AST_NODE_NAME_SIZE + 1];

    array_t ids;
    array_t stmts;

    /* results of semantic checker */
    ast_blk_t *up;
};

ast_blk_t *blk_new_normal(src_pos_t *pos);
ast_blk_t *blk_new_root(src_pos_t *pos);
ast_blk_t *blk_new_loop(src_pos_t *pos);
ast_blk_t *blk_new_switch(src_pos_t *pos);

void blk_set_loop(ast_blk_t *blk);

ast_blk_t *blk_search(ast_blk_t *blk, blk_kind_t kind);

void ast_blk_dump(ast_blk_t *blk, int indent);

#endif /* ! _AST_BLK_H */
