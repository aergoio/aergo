/**
 * @file    trans.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _TRANS_H
#define _TRANS_H

#include "common.h"

#include "flag.h"
#include "ast.h"
#include "ir.h"

#ifndef _IR_FN_T
#define _IR_FN_T
typedef struct ir_fn_s ir_fn_t;
#endif /* ! _IR_FN_T */

#ifndef _IR_BB_T
#define _IR_BB_T
typedef struct ir_bb_s ir_bb_t;
#endif /* ! _IR_BB_T */

typedef struct trans_s {
    flag_t flag;

    ir_t *ir;

    ast_blk_t *blk;         /* current block */

    ir_md_t *md;            /* current module */
    ir_fn_t *fn;            /* current function */
    ir_bb_t *bb;            /* current basic block */

    ir_bb_t *cont_bb;       /* for continue */
    ir_bb_t *break_bb;      /* for break */

    bool is_heap;           /* whether allocate to heap */
} trans_t;

void trans(ast_t *ast, flag_t flag, ir_t **ir);

#endif /* ! _TRANS_H */
