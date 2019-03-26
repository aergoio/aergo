/**
 * @file    gen.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_md.h"

#include "gen.h"

static void
gen_init(gen_t *gen, flag_t flag)
{
    memset(gen, 0x00, sizeof(gen_t));

    gen->flag = flag;

    array_init(&gen->instrs, BinaryenExpressionRef);
}

void
gen(ir_t *ir, flag_t flag, char *infile)
{
    int i;
    gen_t gen;

    if (has_error())
        return;

    gen_init(&gen, flag);

    vector_foreach(&ir->mds, i) {
        ir_md_t *md = vector_get_md(&ir->mds, i);

        gen.sgmt = &md->sgmt;

        md_gen(&gen, md);
    }
}

/* end of gen.c */
