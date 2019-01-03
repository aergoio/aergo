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
    if (is_local_id(id))
        gen_add_local(gen, id->meta.type);
}

static void
id_gen_tuple(gen_t *gen, ast_id_t *id)
{
    int i;

    array_foreach(&id->u_tup.var_ids, i) {
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
