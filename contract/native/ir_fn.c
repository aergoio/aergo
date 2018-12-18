/**
 * @file    ir_fn.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ir_fn.h"

ir_fn_t *
fn_new(ast_id_t *id)
{
    ir_fn_t *fn = xmalloc(sizeof(ir_fn_t));

    fn->id = id;

    array_init(&fn->params);
    array_init(&fn->locals);
    array_init(&fn->bbs);

    return fn;
}

void
fn_add_local(ir_fn_t *fn, ast_id_t *id)
{
    array_add_last(&fn->locals, id);
}

void 
fn_add_basic_blk(ir_fn_t *fn, ir_bb_t *bb)
{
    array_add_last(&fn->bbs, bb);
}

/* end of ir_fn.c */
