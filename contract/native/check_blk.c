/**
 * @file    check_blk.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "check_id.h"

#include "check_blk.h"

void
check_blk(check_t *ctx, ast_blk_t *blk)
{
    int i;

    ASSERT(blk != NULL);

    blk->up = ctx->blk;
    ctx->blk = blk;

    for (i = 0; i < array_size(&blk->ids); i++) {
        check_id(ctx, array_item(&blk->ids, i, ast_id_t));
    }

    ctx->blk = blk->up;
}

/* end of check_blk.c */
