/**
 * @file    check_blk.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "check_id.h"
#include "check_stmt.h"

#include "check_blk.h"

void
check_blk(check_t *check, ast_blk_t *blk)
{
    int i;

    ASSERT(blk != NULL);

    blk->up = check->blk;
    check->blk = blk;

    for (i = 0; i < array_size(&blk->ids); i++) {
        check_id(check, array_item(&blk->ids, i, ast_id_t));
    }

    for (i = 0; i < array_size(&blk->stmts); i++) {
        check_stmt(check, array_item(&blk->stmts, i, ast_stmt_t));
    }

    check->blk = blk->up;
}

/* end of check_blk.c */
