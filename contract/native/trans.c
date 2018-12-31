/**
 * @file    trans.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_blk.h"
#include "trans_id.h"

#include "trans.h"

static void
trans_init(trans_t *trans, flag_t flag)
{
    trans->flag = flag;

    trans->ir = ir_new();

    trans->fn = NULL;
    trans->bb = NULL;

    trans->cont_bb = NULL;
    trans->break_bb = NULL;

    trans->is_lval = false;
}

void
trans(ast_t *ast, flag_t flag, ir_t **ir)
{
    int i;
    trans_t trans;

    if (has_error())
        return;

    trans_init(&trans, flag);

    for (i = 0; i < array_size(&ast->root->ids); i++) {
        id_trans(&trans, array_get_id(&ast->root->ids, i));
    }

    *ir = trans.ir;
}

/* end of trans.c */
