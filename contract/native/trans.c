/**
 * @file    trans.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_blk.h"
#include "ir_bb.h"
#include "trans_id.h"

#include "trans.h"

static void
trans_init(trans_t *trans, flag_t flag)
{
    trans->flag = flag;

    trans->ir = ir_new();

    trans->fn = NULL;
    trans->bb = NULL;

    trans->exit_bb = bb_new();
    trans->cont_bb = NULL;
    trans->break_bb = NULL;
}

void
trans(ast_t *ast, flag_t flag, ir_t **ir)
{
    int i;
    trans_t trans;

    trans_init(&trans, flag);

    for (i = 0; i < array_size(&ast->root->ids); i++) {
        id_trans(&trans, array_get(&ast->root->ids, i, ast_id_t));
    }
}

/* end of trans.c */
