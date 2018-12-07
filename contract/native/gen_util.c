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

/* end of gen_util.c */
