/**
 * @file    ast.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_H
#define _AST_H

#include "common.h"

#include "list.h"
#include "location.h"

#ifndef _AST_BLK_T
#define _AST_BLK_T
typedef struct ast_blk_s ast_blk_t;
#endif  /* _AST_BLK_T */

typedef struct ast_ctr_s {
    char *name;

    list_t var_l;
    list_t struct_l;
    list_t func_l;

    ast_blk_t *blk;     // constructor

    yypos_t pos;
} ast_ctr_t;

typedef struct ast_s {
    list_t ctr_l;
} ast_t; 

#endif /* _AST_H */
