/**
 * @file    gen.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _GEN_H
#define _GEN_H

#include "common.h"

#include "ir.h"
#include "binaryen-c.h"

typedef struct gen_s {
    flag_t flag;

    ir_t *ir;

    BinaryenModuleRef module;
    RelooperRef relooper;

    int local_cnt;
    BinaryenType *locals;

    int instr_cnt;
    BinaryenExpressionRef *instrs;

    bool is_lval;
} gen_t;

void gen(ir_t *ir, flag_t flag, char *path);

#endif /* ! _GEN_H */
