/**
 * @file    gen_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_blk.h"

#include "gen_id.h"

static void
id_gen_var(gen_t *gen, ast_id_t *id)
{
}

static void
id_gen_contract(gen_t *gen, ast_id_t *id)
{
    if (id->u_cont.blk != NULL)
        blk_gen(gen, id->u_cont.blk);
}

void
id_gen(gen_t *gen, ast_id_t *id)
{
    switch (id->kind) {
    case ID_VAR:
        id_gen_var(gen, id);
        break;

    case ID_STRUCT:
        //id_gen_struct(gen, id);
        break;

    case ID_ENUM:
        //id_gen_enum(gen, id);
        break;

    case ID_FUNC:
        //id_gen_func(gen, id);
        break;

    case ID_CONTRACT:
        id_gen_contract(gen, id);
        break;

    default:
        ASSERT1(!"invalid identifier", id->kind);
    }
}

/* end of gen_id.c */
