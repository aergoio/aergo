/**
 * @file    gen.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _GEN_H
#define _GEN_H

#include "common.h"

#include "ast.h"
#include "binaryen-c.h"

typedef struct gen_s {
    flag_t flag;

    ast_blk_t *root;

    BinaryenModuleRef module;
} gen_t;

void gen(ast_t *ast, flag_t flag);

#endif /* ! _GEN_H */
