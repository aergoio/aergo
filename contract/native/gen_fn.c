/**
 * @file    gen_fn.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_id.h"
#include "gen_bb.h"
#include "gen_util.h"

#include "gen_fn.h"

void
fn_gen(gen_t *gen, ir_fn_t *fn)
{
    int i;
    int param_cnt;
    BinaryenType *params;
    BinaryenFunctionTypeRef spec;
    BinaryenExpressionRef body;

    param_cnt = array_size(&fn->params);
    params = xmalloc(sizeof(BinaryenType) * param_cnt);

    for (i = 0; i < param_cnt; i++) {
        params[i] = meta_gen(&array_get_id(&fn->params, i)->meta);
    }

    spec = BinaryenAddFunctionType(gen->module, fn->name, BinaryenTypeNone(), params,
                                   param_cnt);

    /* 1st local for base stack address */
    gen_add_local(gen, TYPE_INT32);

    /* 2nd local for relooper */
    gen_add_local(gen, TYPE_INT32);

    array_foreach(&fn->locals, i) {
        id_gen(gen, array_get_id(&fn->locals, i));
    }

    gen->relooper = RelooperCreate();

    array_foreach(&fn->bbs, i) {
        bb_gen(gen, array_get_bb(&fn->bbs, i));
    }

    array_foreach(&fn->bbs, i) {
        br_gen(gen, array_get_bb(&fn->bbs, i));
    }

    body = RelooperRenderAndDispose(gen->relooper, fn->entry_bb->rb, param_cnt + 1,
                                    gen->module);

    BinaryenAddFunction(gen->module, fn->name, spec, gen->locals, gen->local_cnt,
                        BinaryenBlock(gen->module, NULL, &body, 1, BinaryenTypeNone()));

    gen->local_cnt = 0;
    gen->locals = NULL;
}

/* end of gen_fn.c */
