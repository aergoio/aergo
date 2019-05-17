/**
 * @file    gen_fn.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ir_md.h"
#include "gen_bb.h"
#include "gen_util.h"

#include "gen_fn.h"

BinaryenFunctionTypeRef
abi_gen(gen_t *gen, ir_abi_t *abi)
{
    return BinaryenAddFunctionType(gen->module, NULL, abi->result, abi->params, abi->param_cnt);
}

void
fn_gen(gen_t *gen, ir_fn_t *fn)
{
    int i;
    ir_abi_t *abi = fn->abi;
    BinaryenFunctionRef func;
    BinaryenExpressionRef body;

    ASSERT(abi != NULL);

    gen->relooper = RelooperCreate(gen->module);

    /* basic blocks */
    vector_foreach(&fn->bbs, i) {
        bb_gen(gen, vector_get_bb(&fn->bbs, i));
    }

    /* branches */
    vector_foreach(&fn->bbs, i) {
        br_gen(gen, vector_get_bb(&fn->bbs, i));
    }

    body = RelooperRenderAndDispose(gen->relooper, fn->entry_bb->rb, fn->reloop_idx);

    /* The ABI of the built-in function can be added during the generation of the basic block. */
    func = BinaryenAddFunction(gen->module, fn->name, abi_gen(gen, abi),
        (BinaryenType *)array_items(&fn->types), array_size(&fn->types), body);

    if (fn->apiname != NULL)
        BinaryenAddFunctionExport(gen->module, fn->name, fn->apiname);

    if (is_flag_on(gen->flag, FLAG_DEBUG)) {
        ir_md_t *md = gen->md;

        ASSERT1(md->fno >= 0, md->fno);
        ASSERT(md->dis != NULL);

        vector_foreach(md->dis, i) {
            ir_di_t *di = vector_get(md->dis, i, ir_di_t);

            BinaryenFunctionSetDebugLocation(func, di->instr, md->fno, di->line, di->col);
        }

        vector_reset(md->dis);
    }
}

/* end of gen_fn.c */
