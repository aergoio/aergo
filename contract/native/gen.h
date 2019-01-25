/**
 * @file    gen.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _GEN_H
#define _GEN_H

#include "common.h"

#include "ir.h"
#include "array.h"
#include "binaryen-c.h"

#define instr_add(gen, instr)                                                            \
    do {                                                                                 \
        if ((instr) != NULL)                                                             \
            array_add(&(gen)->instrs, instr, BinaryenExpressionRef);                     \
    } while (0)

typedef struct gen_s {
    flag_t flag;

    ir_t *ir;

    BinaryenModuleRef module;
    RelooperRef relooper;

    array_t instrs;

    bool is_lval;
} gen_t;

void gen(ir_t *ir, flag_t flag, char *path);

#endif /* ! _GEN_H */
