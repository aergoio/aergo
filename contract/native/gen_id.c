/**
 * @file    gen_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_blk.h"
#include "gen_exp.h"
#include "gen_meta.h"
#include "gen_util.h"

#include "gen_id.h"

static BinaryenExpressionRef
id_gen_var(gen_t *gen, ast_id_t *id)
{
    int i;
    uint32_t size;
    meta_t *meta = &id->meta;
    ast_exp_t *dflt_exp = id->u_var.dflt_exp;

    size = meta_size(meta);

    if (is_array_type(meta)) {
        for (i = 0; i < id->meta.arr_dim; i++) {
            ASSERT(id->meta.arr_size[i] > 0);
            size *= id->meta.arr_size[i];
        }
    }

    id->idx = gen_add_local(gen, meta);

    if (dflt_exp != NULL) {
        BinaryenExpressionRef val_exp;

        if (is_lit_exp(dflt_exp)) {
            val_exp = exp_gen(gen, dflt_exp, &dflt_exp->meta, false);

            ASSERT2(BinaryenExpressionGetId(val_exp) == BinaryenConstId(),
                    BinaryenExpressionGetId(val_exp), BinaryenConstId());

            meta->addr = BinaryenConstGetValueI32(val_exp);

            return BinaryenSetLocal(gen->module, id->idx, val_exp);
        }
        else {
            ASSERT1(is_init_exp(dflt_exp), dflt_exp->kind);

            meta->addr = dsgmt_occupy(gen->dsgmt, size);

            gen_add_instr(gen, exp_gen(gen, dflt_exp, meta, false));

            return BinaryenSetLocal(gen->module, id->idx, gen_i32(gen, meta->addr));
        }
    }
    else {
        if (!is_primitive_type(meta) || is_array_type(meta))
            meta->addr = dsgmt_occupy(gen->dsgmt, size);
    }

    /*
    if (is_primitive_type(meta) && !is_array_type(meta)) {
        id->idx = gen_add_local(gen, meta);

        if (dflt_exp != NULL)
            return BinaryenSetLocal(gen->module, id->idx,
                                    exp_gen(gen, dflt_exp, meta, false));
    }
    else {
        if (dflt_exp == NULL) {
            meta->addr = dsgmt_occupy(gen->dsgmt, size);
        }
        else if (is_lit_exp(dflt_exp)) {
            BinaryenExpressionRef value = exp_gen(gen, dflt_exp, &dflt_exp->meta, false);

            ASSERT2(BinaryenExpressionGetId(value) == BinaryenConstId(),
                    BinaryenExpressionGetId(value), BinaryenConstId());

            meta->addr = BinaryenConstGetValueI32(value);
        }
        else {
            ASSERT1(is_init_exp(dflt_exp), dflt_exp->kind);

            meta->addr = dsgmt_occupy(gen->dsgmt, size);

            return exp_gen(gen, dflt_exp, meta, false);
        }
    }
    */

    return NULL;
}

static BinaryenExpressionRef
id_gen_func(gen_t *gen, ast_id_t *id)
{
    int i;
    int param_cnt;
    BinaryenType *params;
    BinaryenFunctionTypeRef spec;
    BinaryenFunctionRef func;
    BinaryenExpressionRef body = NULL;
    array_t *param_ids = id->u_func.param_ids;
    array_t *ret_ids = id->u_func.ret_ids;

    param_cnt = array_size(param_ids) + array_size(ret_ids);
    params = xmalloc(sizeof(BinaryenType) * param_cnt);

    for (i = 0; i < array_size(param_ids); i++) {
        ast_id_t *param_id = array_get(param_ids, i, ast_id_t);

        param_id->idx = gen->id_idx++;
        params[param_id->idx] = meta_gen(gen, &param_id->meta);
    }

    for (i = 0; i < array_size(ret_ids); i++) {
        ast_id_t *ret_id = array_get(ret_ids, i, ast_id_t);

        ret_id->idx = gen->id_idx++;
        params[ret_id->idx] = meta_gen(gen, &ret_id->meta);

        ret_id->meta.addr = dsgmt_occupy(gen->dsgmt, meta_size(&ret_id->meta));
    }

    spec = BinaryenAddFunctionType(gen->module, id->name, meta_gen(gen, &id->meta),
                                   params, param_cnt);

    if (id->u_func.blk != NULL) {
        blk_gen(gen, id->u_func.blk);

        body = BinaryenBlock(gen->module, id->name, gen->instrs, gen->instr_cnt,
                             BinaryenTypeNone());
    }
    else {
        body = BinaryenNop(gen->module);
    }

    func = BinaryenAddFunction(gen->module, id->name, spec, gen->locals, gen->local_cnt,
                               body);

    gen->id_idx = 0;
    gen->local_cnt = 0;
    gen->locals = NULL;
    gen->instr_cnt = 0;
    gen->instrs = NULL;

    return func;
}

static BinaryenExpressionRef
id_gen_contract(gen_t *gen, ast_id_t *id)
{
    if (id->u_cont.blk != NULL)
        return blk_gen(gen, id->u_cont.blk);

    return NULL;
}

BinaryenExpressionRef
id_gen(gen_t *gen, ast_id_t *id)
{
    switch (id->kind) {
    case ID_VAR:
        return id_gen_var(gen, id);

    case ID_STRUCT:
    case ID_ENUM:
        break;

    case ID_FUNC:
        return id_gen_func(gen, id);

    case ID_CONTRACT:
        return id_gen_contract(gen, id);

    default:
        ASSERT1(!"invalid identifier", id->kind);
    }

    return NULL;
}

/* end of gen_id.c */
