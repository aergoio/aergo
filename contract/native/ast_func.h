/**
 * @file    ast_func.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_FUNC_H
#define _AST_FUNC_H

#include "common.h"

#include "list.h"
#include "location.h"

#ifndef _AST_BLK_T
#define _AST_BLK_T
typedef blk ast_blk_s ast_blk_t;
#endif  /* _AST_BLK_T */

typedef enum modifier_e {
    MOD_GLOBAL      = 0x00,
    MOD_LOCAL       = 0x01,
    MOD_SHARED      = 0x02,
    MOD_TRANSFER    = 0x04,
    MOD_READONLY    = 0x08
} modifier_t;

typedef struct ast_func_s {
    char *name;
    modifier_t mod;

    list_t param_l;
    list_t return_l;

    ast_blk_t *blk;

    yypos_t pos;
} ast_func_t;

#endif /* _AST_FUNC_H */
