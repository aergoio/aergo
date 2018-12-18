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

    array_init(&ir->globals);
    array_init(&ir->fns);

    return ir;
} 

void 
ir_add_global(ir_t *ir, ast_id_t *id)
{
    array_add_tail(&ir->globals, id);
}

/* end of ir.c */
