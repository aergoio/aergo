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
    int instr_cnt = gen->instr_cnt;
    BinaryenExpressionRef *instrs = gen->instrs;
    BinaryenExpressionRef code;

    gen->instr_cnt = 0;
    gen->instrs = NULL;

    for (i = 0; i < array_size(&bb->stmts); i++) {
        gen_add_instr(gen, stmt_gen(gen, array_get(&bb->stmts, i, ast_stmt_t)));
    }

    code = BinaryenBlock(gen->module, NULL, gen->instrs, gen->instr_cnt,
                         BinaryenTypeNone());

    bb->rb = RelooperAddBlock(gen->relooper, code);

    gen->instr_cnt = instr_cnt;
    gen->instrs = instrs;
}

void
br_gen(gen_t *gen, ir_bb_t *bb)
{
    int i;

    ASSERT(bb->rb != NULL);

    for (i = 0; i < array_size(&bb->brs); i++) {
        ir_br_t *br = array_get(&bb->brs, i, ir_br_t);
        BinaryenExpressionRef cond = NULL;

        ASSERT(br->bb->rb != NULL);

        if (br->cond != NULL)
            cond = exp_gen(gen, br->cond, &br->cond->meta, false);

        RelooperAddBranch(bb->rb, br->bb->rb, cond, NULL);
    }
}

/* end of gen_bb.c */
