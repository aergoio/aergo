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

    trans->blk = NULL;

    trans->md = NULL;
    trans->fn = NULL;
    trans->bb = NULL;

    trans->cont_bb = NULL;
    trans->break_bb = NULL;

    trans->is_global = false;
}

void
trans(ast_t *ast, flag_t flag, ir_t **ir)
{
    int i;
    trans_t trans;

    if (has_error())
        return;

    trans_init(&trans, flag);

    vector_foreach(&ast->root->ids, i) {
        id_trans(&trans, vector_get_id(&ast->root->ids, i));
    }

    *ir = trans.ir;
}

/* end of trans.c */
