/**
 * @file    gen.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _GEN_H
#define _GEN_H

#include "common.h"

#include "ir.h"
#include "dsgmt.h"
#include "binaryen-c.h"

typedef struct gen_s {
    flag_t flag;
    char path[PATH_MAX_LEN + 5];

    BinaryenModuleRef module;
    RelooperRef relooper;

    int local_cnt;
    BinaryenType *locals;

    int instr_cnt;
    BinaryenExpressionRef *instrs;

    dsgmt_t *dsgmt;
} gen_t;

void gen(ir_t *ir, flag_t flag, char *path);

#endif /* ! _GEN_H */
