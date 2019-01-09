/**
 * @file    gen_bb.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_stmt.h"
#include "gen_exp.h"

#include "gen_bb.h"

void
bb_gen(gen_t *gen, ir_bb_t *bb)
{
    int i;
    BinaryenExpressionRef block;

    ASSERT1(gen->instr_cnt == 0, gen->instr_cnt);
    ASSERT(gen->instrs == NULL);

    array_foreach(&bb->stmts, i) {
        instr_add(gen, stmt_gen(gen, array_get_stmt(&bb->stmts, i)));
    }

    block = BinaryenBlock(gen->module, NULL, gen->instrs, gen->instr_cnt,
                          BinaryenTypeNone());

    bb->rb = RelooperAddBlock(gen->relooper, block);

    gen->instr_cnt = 0;
    gen->instrs = NULL;
}

void
br_gen(gen_t *gen, ir_bb_t *bb)
{
    int i;

    ASSERT(bb->rb != NULL);

    array_foreach(&bb->brs, i) {
        ir_br_t *br = array_get_br(&bb->brs, i);
        BinaryenExpressionRef cond = NULL;

        ASSERT(br->bb->rb != NULL);

        if (br->cond_exp != NULL)
            cond = exp_gen(gen, br->cond_exp);

        RelooperAddBranch(bb->rb, br->bb->rb, cond, NULL);
    }
}

/* end of gen_bb.c */
