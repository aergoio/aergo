/**
 * @file    gen.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _GEN_H
#define _GEN_H

#include "common.h"

#include "ir.h"
#include "meta.h"
#include "binaryen-c.h"

#define i32_gen(gen, v)     BinaryenConst((gen)->module, BinaryenLiteralInt32(v))
#define i64_gen(gen, v)     BinaryenConst((gen)->module, BinaryenLiteralInt64(v))
#define f32_gen(gen, v)     BinaryenConst((gen)->module, BinaryenLiteralFloat32(v))
#define f64_gen(gen, v)     BinaryenConst((gen)->module, BinaryenLiteralFloat64(v))

typedef struct gen_s {
    flag_t flag;
    char path[PATH_MAX_LEN + 5];

    BinaryenModuleRef module;
    RelooperRef relooper;

    int local_cnt;
    BinaryenType *locals;

    int instr_cnt;
    BinaryenExpressionRef *instrs;

    bool is_lval;
} gen_t;

void gen(ir_t *ir, flag_t flag, char *path);

void local_add(gen_t *gen, type_t type);
void instr_add(gen_t *gen, BinaryenExpressionRef instr);

BinaryenType meta_gen(meta_t *meta);

#endif /* ! _GEN_H */
