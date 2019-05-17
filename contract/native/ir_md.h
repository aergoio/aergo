/**
 * @file    ir_md.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _IR_MD_H
#define _IR_MD_H

#include "common.h"

#include "vector.h"
#include "ir_sgmt.h"

#ifndef _IR_MD_T
#define _IR_MD_T
typedef struct ir_md_s ir_md_t;
#endif /* ! _IR_MD_T */

#ifndef _IR_ABI_T
#define _IR_ABI_T
typedef struct ir_abi_s ir_abi_t;
#endif /* ! _IR_ABI_T */

#ifndef _IR_FN_T
#define _IR_FN_T
typedef struct ir_fn_s ir_fn_t;
#endif /* ! _IR_FN_T */

typedef struct ir_di_s {
    BinaryenExpressionRef instr;
    int line;
    int col;
} ir_di_t;

struct ir_md_s {
    char *name;

    vector_t abis;
    vector_t fns;

    ir_sgmt_t sgmt;

    /* debug information */
    int fno;
    vector_t *dis;
};

ir_md_t *md_new(char *name);

void md_add_abi(ir_md_t *md, ir_abi_t *abi);
void md_add_fn(ir_md_t *md, ir_fn_t *fn);

void md_add_di(ir_md_t *md, BinaryenExpressionRef instr, src_pos_t *pos);

#endif /* no _IR_MD_H */
