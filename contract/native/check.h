/**
 * @file    check.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _CHECK_H
#define _CHECK_H

#include "common.h"

#include "ast.h"
#include "ast_blk.h"
#include "ast_exp.h"
#include "ast_stmt.h"
#include "ast_var.h"
#include "ast_struct.h"
#include "ast_meta.h"

typedef struct check_s {
    ast_t *ast;
    ast_blk_t *blk;     // current block
} check_t;

void check(ast_t *ast, flag_t flag);

ast_var_t *check_search_var(check_t *ctx, int num, char *name);
ast_struct_t *check_search_struct(check_t *ctx, int num, char *name);

#endif /* ! _CHECK_H */
