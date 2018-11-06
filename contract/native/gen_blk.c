/**
 * @file    gen_blk.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_id.h"
#include "gen_stmt.h"

#include "gen_blk.h"

void
blk_gen(gen_t *gen, ast_blk_t *blk)
{
    int i;

    ASSERT(blk != NULL);

    for (i = 0; i < array_size(&blk->ids); i++) {
        id_gen(gen, array_item(&blk->ids, i, ast_id_t));
    }

    for (i = 0; i < array_size(&blk->stmts); i++) {
        stmt_gen(gen, array_item(&blk->stmts, i, ast_stmt_t));
    }

} 

/* end of gen_blk.c */
