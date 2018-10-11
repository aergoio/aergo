/**
 * @file    check.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _CHECK_H
#define _CHECK_H

#include "common.h"

#include "ast.h"

#ifndef _AST_BLK_T
#define _AST_BLK_T
typedef struct ast_blk_s ast_blk_t;
#endif /* ! _AST_BLK_T */

#ifndef _AST_ID_T
#define _AST_ID_T
typedef struct ast_id_s ast_id_t;
#endif /* ! _AST_ID_T */

typedef struct check_s {
    flag_t flag;

    ast_blk_t *root;

    /* temporary context */
    ast_blk_t *blk;         /* current block */
    ast_id_t *aq_id;        /* current access qualifier */
    ast_id_t *fn_id;        /* current function */
} check_t;

void check(ast_t *ast, flag_t flag);

#endif /* ! _CHECK_H */
