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

    vector_init(&ir->mds);

    return ir;
}

void
ir_add_md(ir_t *ir, ir_md_t *md)
{
    vector_add_last(&ir->mds, md);
}

/* end of ir.c */
