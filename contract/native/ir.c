/**
 * @file    ir.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ir_sgmt.h"

#include "ir.h"

ir_t *
ir_new(void)
{
    ir_t *ir = xmalloc(sizeof(ir_t));

    array_init(&ir->globals);
    array_init(&ir->fns);

    ir->sgmt = sgmt_new();

    return ir;
}

void
ir_add_global(ir_t *ir, ast_id_t *id)
{
    array_add_last(&ir->globals, id);

    id->addr = sgmt_add_global(ir->sgmt, id->meta.type);
}

void
ir_add_fn(ir_t *ir, ir_fn_t *fn)
{
    array_add_last(&ir->fns, fn);
}

/* end of ir.c */