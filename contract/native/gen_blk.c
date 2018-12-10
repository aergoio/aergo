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
    BinaryenExpressionRef instr;

    ASSERT(blk != NULL);

    for (i = 0; i < array_size(&blk->ids); i++) {
        instr = id_gen(gen, array_get(&blk->ids, i, ast_id_t));

        if (!is_contract_blk(blk))
            gen_add_instr(gen, instr);
    }

    for (i = 0; i < array_size(&blk->stmts); i++) {
        instr = stmt_gen(gen, array_get(&blk->stmts, i, ast_stmt_t));

        if (!is_contract_blk(blk))
            gen_add_instr(gen, instr);
    }

    return NULL;
}

/* end of gen_blk.c */
