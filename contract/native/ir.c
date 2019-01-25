/**
 * @file    ir.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ir.h"

ir_t *
ir_new(void)
{
    ir_t *ir = xmalloc(sizeof(ir_t));

    vector_init(&ir->abis);
    vector_init(&ir->fns);

    sgmt_init(&ir->sgmt);

    ir->offset = 0;

    return ir;
}

void
ir_add_fn(ir_t *ir, ir_fn_t *fn)
{
    vector_add_last(&ir->fns, fn);
}

/* end of ir.c */
