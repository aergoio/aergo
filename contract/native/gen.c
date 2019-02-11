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
    gen->flag = flag;
    gen->sgmt = NULL;

    gen->module = NULL;
    gen->relooper = NULL;

    array_init(&gen->instrs, BinaryenExpressionRef);

    gen->is_lval = false;
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
