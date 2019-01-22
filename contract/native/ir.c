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
    ASSERT(idx >= 0);

    if (is_array_meta(&id->meta))
        /* The array is always accessed as a reference */
        ir->offset = ALIGN32(ir->offset);
    else
        ir->offset = ALIGN(ir->offset, meta_align(&id->meta));

    /* Global variables are always accessed with "base_idx(== heap$offset) + rel_addr",
     * and offset is used only when accessing an array or struct element */

    id->meta.base_idx = idx;
    id->meta.rel_addr = ir->offset;
    id->meta.rel_offset = 0;

    if (is_array_meta(&id->meta))
        ir->offset += sizeof(uint32_t);
    else
        ir->offset += TYPE_BYTE(id->meta.type);
}

void
ir_add_fn(ir_t *ir, ir_fn_t *fn)
{
    array_add_last(&ir->fns, fn);
}

/* end of ir.c */
