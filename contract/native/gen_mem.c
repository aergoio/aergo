/**
 * @file    gen_mem.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "binaryen-c.h"

#include "gen_mem.h"

void 
mem_gen(gen_t *gen, dsgmt_t *dsgmt)
{
    int i;
    BinaryenExpressionRef *offsets;

    offsets = xmalloc(sizeof(BinaryenExpressionRef) * dsgmt->size);

    for (i = 0; i < dsgmt->size; i++) {
        offsets[i] = BinaryenConst(gen->module, BinaryenLiteralInt32(dsgmt->addrs[i]));
    }

    BinaryenSetMemory(gen->module, 0, dsgmt->offset, "memory", 
                      (const char **)dsgmt->datas, offsets, dsgmt->lens, dsgmt->size, 1);
}

/* end of gen_mem.c */
