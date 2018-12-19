/**
 * @file    gen_fn.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_id.h"
#include "gen_meta.h"
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
    ast_id_t *id = fn->id;
    array_t *param_ids = id->u_fn.param_ids;
    array_t *ret_ids = id->u_fn.ret_ids;

    param_cnt = array_size(param_ids) + array_size(ret_ids);
    params = xmalloc(sizeof(BinaryenType) * param_cnt);

    for (i = 0; i < array_size(param_ids); i++) {
        ast_id_t *param_id = array_get(param_ids, i, ast_id_t);

        param_id->idx = gen->id_idx++;
        params[param_id->idx] = meta_gen(gen, &param_id->meta);
    }

    gen->ret_idx = gen->id_idx;

    for (i = 0; i < array_size(ret_ids); i++) {
        ast_id_t *ret_id = array_get(ret_ids, i, ast_id_t);

        ret_id->idx = gen->id_idx++;
        params[ret_id->idx] = meta_gen(gen, &ret_id->meta);

        ret_id->meta.addr = 
            dsgmt_occupy(gen->dsgmt, gen->module, meta_size(&ret_id->meta));
    }

    spec = BinaryenAddFunctionType(gen->module, id->name, BinaryenTypeNone(),
                                   params, param_cnt);

    for (i = 0; i < array_size(&fn->locals); i++) {
        gen_add_instr(gen, id_gen(gen, array_get(&fn->locals, i, ast_id_t)));
    }

    gen->relooper = RelooperCreate();

    for (i = 0; i < array_size(&fn->bbs); i++) {
        bb_gen(gen, array_get(&fn->bbs, i, ir_bb_t));
    }

    for (i = 0; i < array_size(&fn->bbs); i++) {
        br_gen(gen, array_get(&fn->bbs, i, ir_bb_t));
    }

    gen_add_instr(gen, RelooperRenderAndDispose(gen->relooper, fn->entry_bb, 0, gen->module));

    BinaryenAddFunction(gen->module, id->name, spec, gen->locals, gen->local_cnt,
                        BinaryenBlock(gen->module, NULL, gen->instrs, gen->instr_cnt, 
                                      BinaryenTypeNone()));

    gen->id_idx = 0;
    gen->ret_idx = 0;
    gen->local_cnt = 0;
    gen->locals = NULL;
}

/* end of gen_fn.c */
