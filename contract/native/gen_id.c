/**
 * @file    gen_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_blk.h"
#include "gen_exp.h"
#include "gen_meta.h"

#include "gen_id.h"

static BinaryenExpressionRef
id_gen_var(gen_t *gen, ast_id_t *id)
{
    /*
    int i = 1;
        BinaryenExpressionRef value = BinaryenConst(module, BinaryenLiteralInt32(1));
        BinaryenSetLocal(module, 0, value);

    int x = y;
        if (y is local)
            BinaryenExpressionRef value = BinaryenGetLocal(module, idx);
        else
            BinaryenExpressionRef value = BinaryenLoad(module, size, signed, 0, align, type, address of y);

        BinaryenSetLocal(module, 0, value);

    string x = "abc";
    int x[2] = { 1, 2 };
    struct x = { 1, "abc" };
        // nothing to do
        // x should has address of value in data segment

    string y = x;
        // nothing to do
        // x should has address of x in data segment

    int x[2] = { 1, y };
    struct x = { 1, y };
        BinaryenExpressionRef value = BinaryenConst(module, BinaryenLiteralInt32(1));
        BinaryenStore(module, size, offset, align, addr, value, type);

        if (y is local)
            BinaryenExpressionRef value = BinaryenGetLocal(module, index of y);
        else
            BinaryenExpressionRef value = BinaryenLoad(module, size, signed, offset, align, type, address of y);

        BinaryenStore(module, size, offset + x, align, address of x, value, type);
    */

    int i;
    uint32_t size = meta_size(&id->meta);
    ast_exp_t *dflt_exp = id->u_var.dflt_exp;

    if (is_array_type(&id->meta)) {
        for (i = 0; i < id->meta.arr_dim; i++) {
            ASSERT(id->meta.arr_size[i] > 0);
            size *= id->meta.arr_size[i];
        }
    }

    if (is_primitive_type(&id->meta) && !is_array_type(&id->meta)) {
        id->idx = gen_add_local(gen, &id->meta);

        if (dflt_exp != NULL)
            return BinaryenSetLocal(gen->module, id->idx, exp_gen(gen, dflt_exp, false));
    }
    else {
        if (dflt_exp == NULL) {
            id->addr = dsgmt_occupy(gen->dsgmt, size);
        }
        else if (is_lit_exp(dflt_exp)) {
            BinaryenExpressionRef value = exp_gen(gen, dflt_exp, false);

            ASSERT2(BinaryenExpressionGetId(value) == BinaryenConstId(),
                    BinaryenExpressionGetId(value), BinaryenConstId());

            id->addr = BinaryenConstGetValueI32(value);
        }
        else {
            id->addr = dsgmt_occupy(gen->dsgmt, size);

            return exp_gen(gen, dflt_exp, false);
        }
    }

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
        params[gen->id_idx++] = meta_gen(gen, &array_get(param_ids, i, ast_id_t)->meta);
    }

    for (i = 0; i < array_size(ret_ids); i++) {
        params[gen->id_idx++] = meta_gen(gen, &array_get(ret_ids, i, ast_id_t)->meta);
    }

    spec = BinaryenAddFunctionType(gen->module, id->name, meta_gen(gen, &id->meta),
                                   params, param_cnt);

    if (id->u_func.blk != NULL)
        body = blk_gen(gen, id->u_func.blk);

    func = BinaryenAddFunction(gen->module, id->name, spec, gen->locals, gen->local_cnt,
                               body);

    gen->id_idx = 0;
    gen->local_cnt = 0;
    gen->locals = NULL;

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
