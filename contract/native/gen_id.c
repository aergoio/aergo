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
    if (id->u_var.init_exp != NULL) {
        BinaryenExpressionRef value = exp_gen(gen, id->u_var.init_exp);

        return BinaryenSetLocal(gen->module, 0, value);
    }

    return NULL;
}

static BinaryenExpressionRef
id_gen_func(gen_t *gen, ast_id_t *id)
{
    int i, j;
    int param_cnt;
    BinaryenType *params;
    int local_cnt = 0;
    BinaryenType *locals = NULL;
    BinaryenFunctionRef spec;
    BinaryenExpressionRef body = NULL;
    ast_blk_t *blk = id->u_func.blk;
    array_t *param_ids = id->u_func.param_ids;

    param_cnt = array_size(param_ids);
    params = xmalloc(sizeof(BinaryenType) * param_cnt);

    for (i = 0, j = 0; param_cnt; i++) {
        params[j++] = meta_gen(gen, &array_item(param_ids, i, ast_id_t)->meta);
    }

    spec = BinaryenAddFunctionType(gen->module, id->name, meta_gen(gen, &id->meta), 
                                   params, param_cnt);

    if (blk != NULL) {
        local_cnt = array_size(&blk->ids);
        locals = xmalloc(sizeof(BinaryenType) * local_cnt);

        for (i = 0, j = 0; i < local_cnt; i++) {
            locals[j++] = meta_gen(gen, &array_item(&blk->ids, i, ast_id_t)->meta);
        }

        body = blk_gen(gen, blk);
    }

    return BinaryenAddFunction(gen->module, id->name, spec, locals, local_cnt, body);
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
        //id_gen_struct(gen, id);
        break;

    case ID_ENUM:
        //id_gen_enum(gen, id);
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
