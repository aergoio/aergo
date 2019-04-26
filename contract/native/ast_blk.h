/**
 * @file    ast_blk.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_BLK_H
#define _AST_BLK_H

#include "common.h"

#include "ast.h"
#include "vector.h"

#define is_normal_blk(blk)          ((blk)->kind == BLK_NORMAL)
#define is_root_blk(blk)            ((blk)->kind == BLK_ROOT)
#define is_cont_blk(blk)            ((blk)->kind == BLK_CONT)
#define is_itf_blk(blk)             ((blk)->kind == BLK_ITF)
#define is_lib_blk(blk)             ((blk)->kind == BLK_LIB)
#define is_loop_blk(blk)            ((blk)->kind == BLK_LOOP)
#define is_switch_blk(blk)          ((blk)->kind == BLK_SWITCH)

#ifndef _AST_BLK_T
#define _AST_BLK_T
typedef struct ast_blk_s ast_blk_t;
#endif /* ! _AST_BLK_T */

#ifndef _AST_ID_T
#define _AST_ID_T
typedef struct ast_id_s ast_id_t;
#endif /* ! _AST_ID_T */

struct ast_blk_s {
    blk_kind_t kind;

    vector_t ids;
    vector_t stmts;

    ast_blk_t *up;

    AST_NODE_DECL;
};

ast_blk_t *blk_new_normal(src_pos_t *pos);
ast_blk_t *blk_new_root(src_pos_t *pos);
ast_blk_t *blk_new_contract(src_pos_t *pos);
ast_blk_t *blk_new_interface(src_pos_t *pos);
ast_blk_t *blk_new_library(src_pos_t *pos);
ast_blk_t *blk_new_loop(src_pos_t *pos);
ast_blk_t *blk_new_switch(src_pos_t *pos);

ast_blk_t *blk_search(ast_blk_t *blk, blk_kind_t kind);

ast_id_t *blk_search_id(ast_blk_t *blk, char *name);
ast_id_t *blk_search_type(ast_blk_t *blk, char *name);
ast_id_t *blk_search_label(ast_blk_t *blk, char *name);

#endif /* ! _AST_BLK_H */
