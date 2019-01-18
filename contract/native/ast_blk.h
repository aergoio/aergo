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
#define is_cont_blk(blk)            ((blk)->kind == BLK_CONT)
#define is_itf_blk(blk)             ((blk)->kind == BLK_ITF)
#define is_fn_blk(blk)              ((blk)->kind == BLK_FN)
#define is_loop_blk(blk)            ((blk)->kind == BLK_LOOP)

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

    array_t ids;
    array_t nodes;

    ast_blk_t *up;
};

ast_blk_t *blk_new_normal(void);
ast_blk_t *blk_new_root(void);
ast_blk_t *blk_new_contract(void);
ast_blk_t *blk_new_interface(void);
ast_blk_t *blk_new_fn(void);
ast_blk_t *blk_new_loop(void);
ast_blk_t *blk_new_switch(void);

ast_blk_t *blk_search(ast_blk_t *blk, blk_kind_t kind);

ast_id_t *blk_search_id(ast_blk_t *blk, char *name, int num, bool is_type);
ast_id_t *blk_search_label(ast_blk_t *blk, char *name);

#endif /* ! _AST_BLK_H */
