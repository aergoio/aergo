/**
 * @file    gen_util.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_meta.h"

#include "gen_util.h"

uint32_t
gen_add_local(gen_t *gen, meta_t *meta)
{
    if (gen->locals == NULL)
        gen->locals = xmalloc(sizeof(BinaryenType));
    else
        gen->locals = xrealloc(gen->locals, sizeof(BinaryenType) * (gen->local_cnt + 1));

    gen->locals[gen->local_cnt++] = meta_gen(gen, meta);

    return gen->id_idx++;
}

void
gen_add_instr(gen_t *gen, BinaryenExpressionRef instr)
{
    if (instr == NULL)
        return;

    if (gen->instrs == NULL)
        gen->instrs = xmalloc(sizeof(BinaryenExpressionRef));
    else
        gen->instrs = xrealloc(gen->instrs,
                               sizeof(BinaryenExpressionRef) * (gen->instr_cnt + 1));

    gen->instrs[gen->instr_cnt++] = instr;
}

/* end of gen_util.c */
