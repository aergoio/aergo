/**
 * @file    ast_blk.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_BLK_H
#define _AST_BLK_H

#include "common.h"

#include "ast.h"

#define LABEL_MAX_SIZE              128

#ifndef _AST_BLK_T
#define _AST_BLK_T
typedef struct ast_blk_s ast_blk_t;
#endif /* ! _AST_BLK_T */

struct ast_blk_s {
    AST_NODE_DECL;

    blk_kind_t kind;

    array_t ids;
    array_t stmts;

    /* results of semantic checker */
    ast_blk_t *up;

    char loop_label[LABEL_MAX_SIZE + 1];
    char cont_label[LABEL_MAX_SIZE + 1];
    char exit_label[LABEL_MAX_SIZE + 1];
};

ast_blk_t *blk_new_anon(src_pos_t *pos);
ast_blk_t *blk_new_root(src_pos_t *pos);
ast_blk_t *blk_new_loop(src_pos_t *pos);

ast_blk_t *blk_search_loop(ast_blk_t *blk);

void ast_blk_dump(ast_blk_t *blk, int indent);

#endif /* ! _AST_BLK_H */
