/**
 * @file    gen_bb.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_stmt.h"
#include "gen_exp.h"
#include "gen_util.h"

#include "gen_bb.h"

void
bb_gen(gen_t *gen, ir_bb_t *bb)
{
    int i;
    BinaryenExpressionRef block;

    vector_foreach(&bb->stmts, i) {
        bb_add_instr(bb, stmt_gen(gen, vector_get_stmt(&bb->stmts, i)));
    }

    block = BinaryenBlock(gen->module, NULL, (BinaryenExpressionRef *)array_items(&bb->instrs),
                          array_size(&bb->instrs), BinaryenTypeNone());

    bb->rb = RelooperAddBlock(gen->relooper, block);
}

void
br_gen(gen_t *gen, ir_bb_t *bb)
{
    int i;

    ASSERT(bb->rb != NULL);

    vector_foreach(&bb->brs, i) {
        ir_br_t *br = vector_get_br(&bb->brs, i);
        BinaryenExpressionRef cond = NULL;

        ASSERT(br->bb->rb != NULL);

        if (br->cond_exp != NULL)
            cond = exp_gen(gen, br->cond_exp);

        RelooperAddBranch(bb->rb, br->bb->rb, cond, NULL);
    }
}

/* end of gen_bb.c */
