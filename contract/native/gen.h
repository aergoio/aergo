/**
 * @file    gen.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _GEN_H
#define _GEN_H

#include "common.h"

#include "flag.h"
#include "error.h"
#include "ir.h"
#include "binaryen-c.h"

#ifndef _IR_SGMT_T
#define _IR_SGMT_T
typedef struct ir_sgmt_s ir_sgmt_t;
#endif /* ! _IR_SGMT_T */

typedef struct gen_s {
    flag_t flag;

    ir_md_t *md;

    BinaryenModuleRef module;
    RelooperRef relooper;

    bool is_lval;
    BinaryenExpressionRef rvalue;
} gen_t;

void gen(ir_t *ir, flag_t flag, char *path);

#endif /* ! _GEN_H */
