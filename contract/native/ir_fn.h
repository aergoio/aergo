/**
 * @file    ir_fn.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _IR_FN_H
#define _IR_FN_H

#include "common.h"

#include "array.h"
#include "ast_id.h"
#include "binaryen-c.h"

#ifndef _IR_FN_T
#define _IR_FN_T
typedef struct ir_fn_s ir_fn_t;
#endif /* ! _IR_FN_T */

#ifndef _IR_BB_T
#define _IR_BB_T
typedef struct ir_bb_s ir_bb_t;
#endif /* ! _IR_BB_T */

struct ir_fn_s {
    char *name;

    int param_cnt;
    BinaryenType *params;   /* parameter types (including return) */

    array_t locals;         /* entire local variables */
    array_t bbs;            /* basic blocks */

    ir_bb_t *entry_bb;
    ir_bb_t *exit_bb;

    uint32_t usage;         /* stack usage */
};

ir_fn_t *fn_new(ast_id_t *id);

void fn_add_local(ir_fn_t *fn, ast_id_t *id);
void fn_add_stack(ir_fn_t *fn, ast_id_t *id);
void fn_add_basic_blk(ir_fn_t *fn, ir_bb_t *bb);

#endif /* no _IR_FN_H */
