/**
 * @file    check.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _CHECK_H
#define _CHECK_H

#include "common.h"

#include "ast.h"

typedef struct check_s {
    ast_t *ast;
    ast_blk_t *blk;     // current block
} check_t;

void check(ast_t *ast, flag_t flag);

#endif /* ! _CHECK_H */
