/**
 * @file    gen_fn.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_fn.h"

void
fn_gen(gen_t *gen, ir_fn_t *fn)
{
    /*
    int i;
    int param_cnt;
    BinaryenType *params;
    BinaryenFunctionTypeRef spec;
    BinaryenFunctionRef func;
    ast_id_t *fn_id = fn->id;
    array_t *param_ids = fn_id->u_fn.param_ids;
    array_t *ret_ids = fn_id->u_fn.ret_ids;

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

    spec = BinaryenAddFunctionType(gen->module, fn_id->name, BinaryenTypeNone(),
                                   params, param_cnt);

    func = BinaryenAddFunction(gen->module, fn_id->name, spec, gen->locals, gen->local_cnt,
                               blk_gen(gen, fn_id->u_fn.blk));

    gen->id_idx = 0;
    gen->ret_idx = 0;
    gen->local_cnt = 0;
    gen->locals = NULL;
    */
}

/* end of gen_fn.c */
