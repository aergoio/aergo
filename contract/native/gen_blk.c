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
    int entry_cnt;
    BinaryenExpressionRef *entries;

    ASSERT(blk != NULL);

    entry_cnt = array_size(&blk->ids) + array_size(&blk->stmts);
    entries = xmalloc(sizeof(BinaryenExpressionRef *) * entry_cnt);

    for (i = 0; i < array_size(&blk->ids); i++) {
        entries[j++] = id_gen(gen, array_item(&blk->ids, i, ast_id_t));
    }

    for (i = 0; i < array_size(&blk->stmts); i++) {
        entries[j++] = stmt_gen(gen, array_item(&blk->stmts, i, ast_stmt_t));
    }

    ASSERT2(j == entry_cnt, j, entry_cnt);

    return BinaryenBlock(gen->module, NULL, entries, entry_cnt, BinaryenTypeNone());
} 

/* end of gen_blk.c */
