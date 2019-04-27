/**
 * @file    ir_bb.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _IR_BB_H
#define _IR_BB_H

#include "common.h"

#include "ast_exp.h"
#include "ast_stmt.h"
#include "ir.h"
#include "binaryen-c.h"

#ifndef _IR_BB_T
#define _IR_BB_T
typedef struct ir_bb_s ir_bb_t;
#endif /* ! _IR_BB_T */

typedef struct ir_br_s {
    ast_exp_t *cond_exp;
    ir_bb_t *bb;
} ir_br_t;

struct ir_bb_s {
    int ref_cnt;

    vector_t stmts;
    vector_t brs;

    RelooperBlockRef rb;
};

ir_bb_t *bb_new(void);

void bb_add_stmt(ir_bb_t *bb, ast_stmt_t *stmt);
void bb_add_branch(ir_bb_t *bb, ast_exp_t *cond_exp, ir_bb_t *br_bb);

#endif /* no _IR_BB_H */
