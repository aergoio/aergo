/**
 * @file    gen_blk.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_id.h"
#include "gen_stmt.h"

#include "gen_blk.h"

BinaryenExpressionRef
blk_gen(gen_t *gen, ast_blk_t *blk)
{
    int i, j = 0;
    BinaryenExpressionRef entry;
    BinaryenExpressionRef *entries;

    ASSERT(blk != NULL);

    entries = xmalloc(sizeof(BinaryenExpressionRef *) *
                      (array_size(&blk->ids) + array_size(&blk->stmts)));

    for (i = 0; i < array_size(&blk->ids); i++) {
        entry = id_gen(gen, array_get(&blk->ids, i, ast_id_t));

        if (entry != NULL)
            entries[j++] = entry;
    }

    for (i = 0; i < array_size(&blk->stmts); i++) {
        entry = stmt_gen(gen, array_get(&blk->stmts, i, ast_stmt_t));

        if (entry != NULL)
            entries[j++] = entry;
    }

    return BinaryenBlock(gen->module, NULL, entries, j, BinaryenTypeNone());
}

/* end of gen_blk.c */
