/**
 * @file    gen_fn.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "gen_bb.h"
#include "gen_util.h"

#include "gen_fn.h"

void
fn_gen(gen_t *gen, ir_fn_t *fn)
{
    int i;
    ir_abi_t *abi = fn->abi;
    BinaryenExpressionRef body;

    ASSERT(abi != NULL);

    /* local variables */
    vector_foreach(&fn->locals, i) {
        local_add(gen, vector_get_id(&fn->locals, i)->meta.type);
    }

    gen->relooper = RelooperCreate();

    /* basic blocks */
    vector_foreach(&fn->bbs, i) {
        bb_gen(gen, vector_get_bb(&fn->bbs, i));
    }

    /* branches */
    vector_foreach(&fn->bbs, i) {
        br_gen(gen, vector_get_bb(&fn->bbs, i));
    }

    body = RelooperRenderAndDispose(gen->relooper, fn->entry_bb->rb, fn->reloop_idx,
                                    gen->module);

    BinaryenAddFunction(gen->module, fn->name, abi->spec, gen->locals, gen->local_cnt,
                        BinaryenBlock(gen->module, NULL, &body, 1, abi->result));

    if (fn->exp_name != NULL)
        BinaryenAddFunctionExport(gen->module, fn->name, fn->exp_name);

    gen->local_cnt = 0;
    gen->locals = NULL;
}

void
abi_gen(gen_t *gen, ir_abi_t *abi)
{
    abi->spec = BinaryenAddFunctionType(gen->module, abi->name, abi->result, abi->params,
                                        abi->param_cnt);
}

/* end of gen_fn.c */
