/**
 * @file    ir_fn.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _IR_FN_H
#define _IR_FN_H

#include "common.h"

#include "array.h"
#include "ast_id.h"

#ifndef _IR_FN_T
#define _IR_FN_T
typedef struct ir_fn_s ir_fn_t;
#endif /* ! _IR_FN_T */

struct ir_fn_s {
    ast_id_t *id;

    array_t params;
    array_t locals;
    array_t bbs;
};

ir_fn_t *fn_new(ast_id_t *id);

void fn_add_local(ir_fn_t *fn, ast_id_t *id);
void fn_add_basic_blk(ir_fn_t *fn, ir_bb_t *bb);

#endif /* no _IR_FN_H */
