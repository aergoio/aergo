/**
 * @file    gen_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_exp.h"
#include "gen_stmt.h"
#include "gen_util.h"

#include "gen_id.h"

static void
id_gen_var(gen_t *gen, ast_id_t *id)
{
    uint32_t size;
    meta_t *meta = &id->meta;

    size = meta_size(meta);

    if (is_array_type(meta)) {
        int i;

        for (i = 0; i < id->meta.arr_dim; i++) {
            ASSERT(id->meta.arr_size[i] > 0);
            size *= id->meta.arr_size[i];
        }
    }

    if (is_global_id(id)) {
        if (!is_primitive_type(meta) || is_array_type(meta))
            meta->addr = dsgmt_occupy(gen->dsgmt, gen->module, size);
        else
            BinaryenAddGlobal(gen->module, id->name, meta_gen(gen, meta), 1, NULL);
    }
    else {
        if (!is_primitive_type(meta) || is_array_type(meta))
            meta->addr = dsgmt_occupy(gen->dsgmt, gen->module, size);

        gen_add_local(gen, meta->type);
    }

    /*
    id->idx = gen_add_local(gen, meta->type);

    if (dflt_exp != NULL) {
        BinaryenExpressionRef val_exp;

        if (is_lit_exp(dflt_exp)) {
            val_exp = exp_gen(gen, dflt_exp, &dflt_exp->meta, false);

            ASSERT2(BinaryenExpressionGetId(val_exp) == BinaryenConstId(),
                    BinaryenExpressionGetId(val_exp), BinaryenConstId());

            if (BinaryenExpressionGetType(val_exp) == BinaryenTypeInt32())
                meta->addr = BinaryenConstGetValueI32(val_exp);
            else
                meta->addr = BinaryenConstGetValueI64(val_exp);

            return BinaryenSetLocal(gen->module, id->idx, val_exp);
        }
        else {
            ASSERT1(is_init_exp(dflt_exp), dflt_exp->kind);

            meta->addr = dsgmt_occupy(gen->dsgmt, gen->module, size);

            gen_add_instr(gen, exp_gen(gen, dflt_exp, meta, false));

            return BinaryenSetLocal(gen->module, id->idx, gen_i32(gen, meta->addr));
        }
    }
    else {
        if (!is_primitive_type(meta) || is_array_type(meta))
            meta->addr = dsgmt_occupy(gen->dsgmt, gen->module, size);
    }
    */
}

static void
id_gen_tuple(gen_t *gen, ast_id_t *id)
{
    int i;

    for (i = 0; i < array_size(&id->u_tup.var_ids); i++) {
        id_gen_var(gen, array_get_id(&id->u_tup.var_ids, i));
    }
}

void
id_gen(gen_t *gen, ast_id_t *id)
{
    switch (id->kind) {
    case ID_VAR:
        id_gen_var(gen, id);

    case ID_TUPLE:
        id_gen_tuple(gen, id);

    default:
        ASSERT1(!"invalid identifier", id->kind);
    }
}

/* end of gen_id.c */
