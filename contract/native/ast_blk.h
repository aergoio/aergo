/**
 * @file    ast_blk.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_BLK_H
#define _AST_BLK_H

#include "common.h"

#include "list.h"
#include "location.h"

#ifndef _AST_BLK_T
#define _AST_BLK_T
typedef struct ast_blk_s ast_blk_t;
#endif  /* _AST_BLK_T */

struct ast_blk_s {
    list_t var_l;
    list_t struct_l;
    list_t stmt_l;

    ast_blk_t *up;

    yypos_t pos;
};

#endif /* _AST_BLK_H */
