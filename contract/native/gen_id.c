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
    meta_t *meta = &id->meta;

#if 0
    if (is_global_id(id)) {
        ASSERT(id->u_var.dflt_exp != NULL);

        if (is_stack_id(id))
            /* 0 is a temporary address. We will set an actual address in constructor */
            BinaryenAddGlobal(gen->module, id->name, meta_gen(meta), 1, gen_i32(gen, 0));
        else
            BinaryenAddGlobal(gen->module, id->name, meta_gen(meta), 1,
                              exp_gen(gen, id->u_var.dflt_exp));
    }
    else if (!is_stack_id(id)) {
        gen_add_local(gen, meta->type);
    }
#endif
    if (is_local_id(id))
        gen_add_local(gen, meta->type);
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
        break;

    case ID_TUPLE:
        id_gen_tuple(gen, id);
        break;

    default:
        ASSERT1(!"invalid identifier", id->kind);
    }
}

/* end of gen_id.c */
