/**
 * @file    gen_blk.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_id.h"
#include "gen_stmt.h"
#include "gen_util.h"

#include "gen_blk.h"

BinaryenExpressionRef
blk_gen(gen_t *gen, ast_blk_t *blk)
{
    int i;
    int instr_cnt;
    BinaryenExpressionRef *instrs;
    BinaryenExpressionRef block;

    if (blk == NULL)
        return BinaryenNop(gen->module);

    instr_cnt = gen->instr_cnt;
    instrs = gen->instrs;

    gen->instr_cnt = 0;
    gen->instrs = NULL;

    for (i = 0; i < array_size(&blk->ids); i++) {
        gen_add_instr(gen, id_gen(gen, array_get_id(&blk->ids, i)));
    }

    for (i = 0; i < array_size(&blk->stmts); i++) {
        gen_add_instr(gen, stmt_gen(gen, array_get_stmt(&blk->stmts, i)));
    }

    block = BinaryenBlock(gen->module, NULL, gen->instrs, gen->instr_cnt,
                          BinaryenTypeNone());

    gen->instr_cnt = instr_cnt;
    gen->instrs = instrs;

    return block;
}

/* end of gen_blk.c */
