/**
 * @file    trans.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _TRANS_H
#define _TRANS_H

#include "common.h"

#include "ast.h"
#include "ir.h"

typedef struct trans_s {
    flag_t flag;

    ir_t *ir;

    ir_fn_t *fn;
    ir_bb_t *bb;

    ir_bb_t *exit_bb;
    ir_bb_t *cont_bb;
    ir_bb_t *break_bb;
} trans_t;

void trans(ast_t *ast, flag_t flag, ir_t **ir);

#endif /* ! _TRANS_H */
