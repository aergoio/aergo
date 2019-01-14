/**
 * @file    ir.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ir_abi.h"
#include "ir_fn.h"
#include "ir_sgmt.h"

#include "ir.h"

ir_t *
ir_new(void)
{
    ir_t *ir = xmalloc(sizeof(ir_t));

    array_init(&ir->abis);
    array_init(&ir->fns);

    sgmt_init(&ir->sgmt);

    ir->offset = 0;

    return ir;
}

void
ir_add_global(ir_t *ir, ast_id_t *id, int idx)
{
    /*
    int addr;
    ast_id_t *cont_id = id->up;

    addr = sgmt_add_global(&ir->sgmt, id->meta.type);

    if (cont_id->addr < 0)
        cont_id->addr = addr;

    id->addr = cont_id->addr;
    id->offset = addr - cont_id->addr;
    */
    ASSERT(idx >= 0);

    ir->offset = ALIGN(ir->offset, meta_align(&id->meta));

    /* The global variable does not use "addr",
     * but uses the local variable set to "base" as the address */
    id->base = idx;
    id->addr = 0;
    id->offset = ir->offset;

    ir->offset += meta_size(&id->meta);
}

void
ir_add_fn(ir_t *ir, ir_fn_t *fn)
{
    array_add_last(&ir->fns, fn);
}

/* end of ir.c */
