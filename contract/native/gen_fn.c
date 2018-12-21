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

    param_cnt = array_size(&fn->params);
    params = xmalloc(sizeof(BinaryenType) * param_cnt);

    for (i = 0; i < param_cnt; i++) {
        params[i] = meta_gen(gen, &array_get_id(&fn->params, i)->meta);
    }

    spec = BinaryenAddFunctionType(gen->module, fn->name, BinaryenTypeNone(),
                                   params, param_cnt);

    /* local-0 for base stack address */
    gen_add_local(gen, TYPE_INT32);

    /* local-1 for relooper */
    gen_add_local(gen, TYPE_INT32);

    for (i = 0; i < array_size(&fn->locals); i++) {
        gen_add_instr(gen, id_gen(gen, array_get_id(&fn->locals, i)));
    }

    gen->relooper = RelooperCreate();

    for (i = 0; i < array_size(&fn->bbs); i++) {
        bb_gen(gen, array_get_bb(&fn->bbs, i));
    }

    for (i = 0; i < array_size(&fn->bbs); i++) {
        br_gen(gen, array_get_bb(&fn->bbs, i));
    }

    gen_add_instr(gen,
                  RelooperRenderAndDispose(gen->relooper, fn->entry_bb, 1, gen->module));

    BinaryenAddFunction(gen->module, fn->name, spec, gen->locals, gen->local_cnt,
                        BinaryenBlock(gen->module, NULL, gen->instrs, gen->instr_cnt,
                                      BinaryenTypeNone()));

    gen->id_idx = 0;
    gen->local_cnt = 0;
    gen->locals = NULL;
}

/* end of gen_fn.c */
